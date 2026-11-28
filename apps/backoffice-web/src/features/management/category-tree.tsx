import { ChoiceField } from "../../shared/option-select";
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
  includeArchived = false,
  label,
  required = false,
}: {
  categories: Category[];
  value: string;
  onChange: (id: string) => void;
  disabled?: boolean;
  currentId?: string;
  allowRoot?: boolean;
  includeArchived?: boolean;
  label?: string;
  required?: boolean;
}) {
  const blocked = (id: string) => {
    const seen = new Set<string>();
    let current = categories.find((c) => c.id === id);
    while (current && !seen.has(current.id)) {
      if (current.id === currentId) return true;
      seen.add(current.id);
      current = categories.find((c) => c.id === current?.parent_id);
    }
    return false;
  };
  return (
    <ChoiceField
      label={label ?? (allowRoot ? "دسته والد" : "دسته‌بندی")}
      required={required}
      value={value}
      disabled={disabled}
      clearLabel={
        required
          ? undefined
          : allowRoot
            ? "دسته اصلی (بدون والد)"
            : "همه دسته‌ها"
      }
      options={categories
        .filter((c) => includeArchived || !c.archived || c.id === value)
        .map((c) => ({
          id: c.id,
          name: categoryPath(categories, c.id),
          disabled: blocked(c.id),
        }))
        .sort((a, b) => a.name.localeCompare(b.name, "fa"))}
      onChange={onChange}
    />
  );
}
