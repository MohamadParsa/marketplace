import { useMemo, useState } from "react";
import { Alert, Button, TextField } from "@mui/material";
import { useWorkspace } from "../../shared/workspace";
import { ActionDialog, DataTable } from "../../shared/components";
import type { Category } from "../../shared/types";
import { CategoryTree, categoryPath } from "./category-tree";
import { CategoryBindingsDialog } from "./shared-attributes";
import {
  CatalogIcon,
  CatalogIconDialog,
  type IconTarget,
} from "./catalog-icon";
import { CategorySalesDialog } from "../commerce/sales-methods-page";

const norm = (s: string) =>
  s.trim().replace(/ي/g, "ی").replace(/ك/g, "ک").toLocaleLowerCase();
export function CategoryCatalogGrid({
  onEdit,
}: {
  onEdit: (category?: Category) => void;
}) {
  const { state, can, busy, execute } = useWorkspace();
  const [name, setName] = useState(""),
    [english, setEnglish] = useState(""),
    [published, setPublished] = useState("all"),
    [archived, setArchived] = useState("active"),
    [attribute, setAttribute] = useState("all"),
    [parent, setParent] = useState(""),
    [branch, setBranch] = useState("direct");
  const [bindings, setBindings] = useState<Category>(),
    [icon, setIcon] = useState<IconTarget>(),
    [sales, setSales] = useState<string>(),
    [archive, setArchive] = useState<Category>();
  const [archiveKey, setArchiveKey] = useState(crypto.randomUUID());
  const categories = state.categories ?? [];
  const rows = useMemo(
    () =>
      categories.filter((c) => {
        if (
          !norm(c.name).includes(norm(name)) ||
          !norm(c.english_name ?? "").includes(norm(english))
        )
          return false;
        if (published !== "all" && c.published !== (published === "yes"))
          return false;
        if (archived !== "all" && c.archived !== (archived === "archived"))
          return false;
        if (
          (attribute === "bound" && !c.bindings?.length) ||
          (attribute === "none" && !!c.bindings?.length) ||
          (attribute === "axis" && !c.bindings?.some((b) => b.axis))
        )
          return false;
        if (branch === "root") return !c.parent_id;
        if (!parent) return true;
        if (branch === "direct") return c.parent_id === parent;
        let current: Category | undefined = c;
        const seen = new Set<string>();
        while (current && !seen.has(current.id)) {
          if (current.id === parent) return true;
          seen.add(current.id);
          current = categories.find((x) => x.id === current?.parent_id);
        }
        return false;
      }),
    [categories, name, english, published, archived, attribute, parent, branch],
  );
  const chips = (
    label: string,
    value: string,
    change: (v: string) => void,
    options: string[][],
  ) => (
    <div className="catalog-filter-choice" role="group" aria-label={label}>
      <strong>{label}</strong>
      <div>
        {options.map(([key, text]) => (
          <Button
            key={key}
            size="small"
            variant={key === value ? "contained" : "outlined"}
            aria-pressed={key === value}
            onClick={() => change(key)}
          >
            {text}
          </Button>
        ))}
      </div>
    </div>
  );
  return (
    <>
      <section
        className="panel catalog-filters"
        aria-label="فیلترهای دسته‌بندی"
      >
        <h2 className="filter-panel-title">فیلترهای دسته‌بندی</h2>
        <div className="catalog-filter-text">
          <TextField
            size="small"
            label="فیلتر نام فارسی"
            value={name}
            onChange={(e) => setName(e.target.value)}
          />
          <TextField
            size="small"
            label="فیلتر نام انگلیسی"
            value={english}
            onChange={(e) => setEnglish(e.target.value)}
          />
        </div>
        {chips("انتشار", published, setPublished, [
          ["all", "همه انتشارها"],
          ["yes", "منتشرشده"],
          ["no", "پیش‌نویس"],
        ])}
        {chips("بایگانی", archived, setArchived, [
          ["active", "غیربایگانی"],
          ["archived", "بایگانی‌شده"],
          ["all", "همه وضعیت‌ها"],
        ])}
        {chips("خصوصیت", attribute, setAttribute, [
          ["all", "همه دسته‌ها"],
          ["bound", "دارای خصوصیت"],
          ["none", "بدون خصوصیت"],
          ["axis", "دارای محور تنوع"],
        ])}
        <details className="catalog-branch-filter">
          <summary>
            والد و شاخه:{" "}
            {branch === "root"
              ? "دسته‌های اصلی"
              : parent
                ? categoryPath(categories, parent)
                : "همه دسته‌ها"}
          </summary>
          {chips("محدوده شاخه", branch, setBranch, [
            ["direct", "فرزندان مستقیم"],
            ["branch", "والد و همه زیرشاخه‌ها"],
            ["root", "فقط دسته‌های اصلی"],
          ])}
          {branch !== "root" && (
            <CategoryTree
              categories={categories}
              value={parent}
              includeArchived
              allowRoot={false}
              onChange={setParent}
            />
          )}
          <Button
            onClick={() => {
              setParent("");
              setBranch("direct");
            }}
          >
            حذف فیلتر شاخه
          </Button>
        </details>
        <div className="toolbar">
          <span aria-live="polite">
            {rows.length.toLocaleString("fa-IR")} دسته از{" "}
            {categories.length.toLocaleString("fa-IR")}
          </span>
          <Button
            onClick={() => {
              setName("");
              setEnglish("");
              setPublished("all");
              setArchived("active");
              setAttribute("all");
              setParent("");
              setBranch("direct");
            }}
          >
            پاک‌کردن فیلترهای دسته
          </Button>
          <Button
            variant="contained"
            disabled={busy || !can("catalog.category.create", "management")}
            onClick={() => onEdit()}
          >
            دسته جدید
          </Button>
        </div>
      </section>
      <section
        className="panel category-catalog-grid"
        aria-label="فهرست دسته‌بندی‌ها"
      >
        <DataTable
        searchable={false}
          key={JSON.stringify([
            name,
            english,
            published,
            archived,
            attribute,
            parent,
            branch,
          ])}
          rows={rows}
          columns={[
            {
              title: "دسته",
              render: (c) => (
                <div className="catalog-name">
                  <CatalogIcon
                    kind="category"
                    name={c.name}
                    iconId={c.icon_id}
                  />
                  <div>
                    <b>{c.name}</b>
                    <small>{c.english_name || "نام انگلیسی ثبت نشده"}</small>
                    <small>{categoryPath(categories, c.id)}</small>
                    <small>
                      {c.display_priority == null
                        ? "بدون اولویت"
                        : `اولویت: ${c.display_priority.toLocaleString("fa-IR")}`}
                    </small>
                  </div>
                </div>
              ),
            },
            {
              title: "انتشار و بایگانی",
              render: (c) => (
                <>
                  <span>{c.published ? "منتشرشده" : "پیش‌نویس"}</span>
                  <small>{c.archived ? "بایگانی‌شده" : "غیربایگانی"}</small>
                </>
              ),
            },
            {
              title: "خصوصیات",
              render: (c) => (
                <>
                  {(c.bindings ?? [])
                    .map(
                      (b) =>
                        state.attributes?.find((a) => a.id === b.attribute_id)
                          ?.name,
                    )
                    .filter(Boolean)
                    .join("، ") || "بدون انتساب صریح"}
                  <small>
                    {(c.bindings ?? [])
                      .filter((b) => b.axis)
                      .length.toLocaleString("fa-IR")}{" "}
                    محور تنوع
                  </small>
                </>
              ),
            },
            {
              title: "عملیات",
              render: (c) => (
                <div className="catalog-row-actions">
                  <Button
                    disabled={
                      busy ||
                      c.archived ||
                      !can("catalog.category.update", "management")
                    }
                    onClick={() => onEdit(c)}
                  >
                    ویرایش
                  </Button>
                  <Button onClick={() => setBindings(c)}>تنظیم خصوصیات</Button>
                  <Button
                    disabled={
                      busy ||
                      c.archived ||
                      !can("category.icon.set", "management")
                    }
                    onClick={() => setIcon({ ...c, kind: "category" })}
                  >
                    آیکن
                  </Button>
                  <Button onClick={() => setSales(c.id)}>
                    مارکاپ روش فروش
                  </Button>
                  <Button
                    disabled={
                      busy ||
                      c.archived ||
                      !can("category.archive", "management")
                    }
                    onClick={() => {
                      setArchiveKey(crypto.randomUUID());
                      setArchive(c);
                    }}
                  >
                    بایگانی
                  </Button>
                </div>
              ),
            },
          ]}
        />
      </section>
      {bindings && (
        <CategoryBindingsDialog
          category={bindings}
          close={() => setBindings(undefined)}
        />
      )}
      {icon && (
        <CatalogIconDialog target={icon} close={() => setIcon(undefined)} />
      )}
      {sales && (
        <CategorySalesDialog id={sales} close={() => setSales(undefined)} />
      )}
      {archive && (
        <ActionDialog
          title="بایگانی دسته"
          fields={[{ name: "reason", label: "دلیل بایگانی" }]}
          open
          onClose={() => setArchive(undefined)}
          onSubmit={async (values) => {
            await execute({
              action: "category.archive",
              target: archive.id,
              version: archive.version,
              reason: values.reason,
              idempotency_key: archiveKey,
            });
            setArchive(undefined);
          }}
        />
      )}
      {categories.length === 0 && (
        <Alert severity="info">
          دسته‌ای ثبت نشده است؛ از «دسته جدید» شروع کنید.
        </Alert>
      )}
    </>
  );
}
