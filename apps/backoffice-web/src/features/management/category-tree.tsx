import { useMemo, useState, type ReactNode } from "react";
import type { Category } from "../../shared/types";

export function categoryPath(categories: Category[], id: string) {
  const byId = new Map(categories.map((c) => [c.id, c]));
  const names: string[] = [],
    seen = new Set<string>();
  let current = byId.get(id);
  while (current && !seen.has(current.id)) {
    seen.add(current.id);
    names.unshift(current.name);
    current = byId.get(current.parent_id ?? "");
  }
  return names.join(" / ");
}
export function CategoryTree({
  categories,
  value,
  onChange,
  disabled = false,
  currentId,
  allowRoot = true,
}: {
  categories: Category[];
  value: string;
  onChange: (id: string) => void;
  disabled?: boolean;
  currentId?: string;
  allowRoot?: boolean;
}) {
  const [search, setSearch] = useState("");
  const children = useMemo(() => {
    const groups = new Map<string, Category[]>();
    for (const c of categories.filter((c) => !c.archived)) {
      const parent = c.parent_id ?? "";
      groups.set(parent, [...(groups.get(parent) ?? []), c]);
    }
    for (const items of groups.values())
      items.sort((a, b) => a.name.localeCompare(b.name, "fa"));
    return groups;
  }, [categories]);
  const matches = new Set<string>();
  const activePath = new Set<string>();
  for (const id of [value, currentId ?? ""]) {
    let current = categories.find((c) => c.id === id);
    while (current && !activePath.has(current.id)) {
      activePath.add(current.id);
      current = categories.find((c) => c.id === current?.parent_id);
    }
  }
  for (const c of categories)
    if (
      !search ||
      categoryPath(categories, c.id)
        .toLocaleLowerCase()
        .includes(search.toLocaleLowerCase())
    ) {
      let current: Category | undefined = c;
      const seen = new Set<string>();
      while (current && !seen.has(current.id)) {
        seen.add(current.id);
        matches.add(current.id);
        current = categories.find((c) => c.id === current?.parent_id);
      }
    }
  const render = (parent: string, seen: Set<string>): ReactNode => (
    <ul>
      {(children.get(parent) ?? [])
        .filter((c) => matches.has(c.id) && !seen.has(c.id))
        .map((c) => {
          const hasChildren = (children.get(c.id)?.length ?? 0) > 0;
          const button = (
            <button
              type="button"
              disabled={disabled || c.id === currentId}
              data-category-id={c.id}
          aria-pressed={value === c.id}
              onClick={() => onChange(c.id)}
            >
              <span>{c.name}</span>
              {c.id === currentId && <small>دسته جاری</small>}
            </button>
          );
          return (
            <li key={c.id}>
              {hasChildren ? (
                <details
                  key={c.id + search}
                  open={search ? true : activePath.has(c.id) || undefined}
                >
                  <summary>
                    {button}
                    <small>
                      {new Intl.NumberFormat("fa-IR").format(
                        children.get(c.id)!.length,
                      )}{" "}
                      زیرشاخه
                    </small>
                  </summary>
                  {render(c.id, new Set([...seen, c.id]))}
                </details>
              ) : (
                button
              )}
            </li>
          );
        })}
    </ul>
  );
  return (
    <section className="category-tree" aria-label="ساختار درختی دسته‌ها">
      <input
        aria-label="جست‌وجو در درخت دسته‌ها"
        placeholder="جست‌وجو در دسته‌ها…"
        value={search}
        onChange={(e) => setSearch(e.target.value)}
      />
      <p className="category-tree-selection">
        {value
          ? categoryPath(categories, value)
          : allowRoot
            ? "دسته اصلی؛ بدون والد"
            : "دسته انتخاب نشده"}
      </p>
      <div className="category-tree-nodes">
        {allowRoot && (
          <button
            type="button"
            disabled={disabled}
            aria-pressed={!value}
            onClick={() => onChange("")}
          >
            دسته اصلی (بدون والد)
          </button>
        )}
        {render("", new Set())}
        {!matches.size && <p>دسته‌ای پیدا نشد.</p>}
      </div>
      {disabled && (
        <small>
          درخت برای نمایش موقعیت دسته است؛ جابه‌جایی والد از این فرم انجام
          نمی‌شود.
        </small>
      )}
    </section>
  );
}
