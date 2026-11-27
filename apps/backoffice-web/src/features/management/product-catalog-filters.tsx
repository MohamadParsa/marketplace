import { useState, useEffect } from "react";
import { FilterPanel } from "../../shared/filter-panel";
import { OptionSelect } from "../../shared/option-select";
import { Button, TextField } from "@mui/material";
import type { Workspace } from "../../shared/types";
import { CategoryTree } from "./category-tree";
import { CatalogChoice } from "./catalog-choice";
export type ProductFilters = {
  name: string;
  english: string;
  code: string;
  category: string;
  brand: string;
  status: string;
  active: string;
  stock: string;
};
export const emptyProductFilters: ProductFilters = {
  name: "",
  english: "",
  code: "",
  category: "",
  brand: "",
  status: "all",
  active: "all",
  stock: "all",
};
const norm = (s: string) =>
  s.trim().replace(/ي/g, "ی").replace(/ك/g, "ک").toLocaleLowerCase();
export function filterProductRows(w: Workspace, f: ProductFilters) {
  const stock = new Set(
    (w.offers ?? [])
      .filter((o) => o.stock - o.reserved > 0)
      .map((o) => o.product_id),
  );
  return (w.products ?? []).filter(
    (p) =>
      norm(p.name).includes(norm(f.name)) &&
      norm(p.english_name ?? "").includes(norm(f.english)) &&
      norm(p.public_code ?? "").includes(norm(f.code)) &&
      (!f.category || p.category_id === f.category) &&
      (!f.brand || p.brand_id === f.brand) &&
      (f.status === "all" || p.status === f.status) &&
      (f.active === "all" || (p.active !== false) === (f.active === "yes")) &&
      (f.stock === "all" || stock.has(p.id) === (f.stock === "yes")),
  );
}
export function ProductCatalogFilters({
  state,
  value,
  onChange,
}: {
  state: Workspace;
  value: ProductFilters;
  onChange: (f: ProductFilters) => void;
}) {
  const [draft, setDraft] = useState(value);
  useEffect(() => setDraft(value), [value]);
  const set = (key: keyof ProductFilters, v: string) =>
    setDraft({ ...draft, [key]: v });
  const choices = (
    key: keyof ProductFilters,
    label: string,
    options: string[][],
  ) => (
    <div className="catalog-filter-choice" role="group" aria-label={label}>
      <div>
        <OptionSelect
          label={label}
          multiple={false}
          options={options.map(([id, name]) => ({
            id: String(id),
            name: String("" + name + ""),
            selected: draft[key] === id,
            disabled: false,
            onSelect: () => set(key, id),
          }))}
        />
      </div>
    </div>
  );
  return (
    <FilterPanel onSearch={() => onChange(draft)}>
      <section className="catalog-filters" aria-label="فیلترهای محصول">
        <h2 className="filter-panel-title">فیلترهای محصول</h2>
        <div className="catalog-filter-text">
          {(
            [
              ["name", "نام فارسی محصول"],
              ["english", "نام انگلیسی محصول"],
              ["code", "کد عمومی محصول"],
            ] as const
          ).map(([key, label]) => (
            <TextField
              key={key}
              size="small"
              label={`فیلتر ${label}`}
              value={draft[key]}
              onChange={(e) => set(key, e.target.value)}
            />
          ))}
        </div>
        {choices("status", "انتشار محصول", [
          ["all", "همه انتشارها"],
          ["published", "منتشرشده"],
          ["draft", "پیش‌نویس"],
          ["pending", "در انتظار بررسی"],
          ["rejected", "ردشده"],
        ])}
        {choices("active", "فعال‌بودن محصول", [
          ["all", "همه وضعیت‌ها"],
          ["yes", "فعال"],
          ["no", "غیرفعال"],
        ])}
        {choices("stock", "موجودی آزاد فیزیکی", [
          ["all", "همه موجودی‌ها"],
          ["yes", "دارای موجودی"],
          ["no", "بدون موجودی"],
        ])}
        <details>
          <summary>فیلتر دسته و برند</summary>
          <CategoryTree
            categories={state.categories ?? []}
            value={draft.category}
            allowRoot={false}
            onChange={(id) => set("category", id)}
          />
          <Button onClick={() => set("category", "")}>همه دسته‌ها</Button>
          <CatalogChoice
            label="برند محصول"
            value={draft.brand}
            options={(state.brands ?? []).map((b) => ({
              id: b.id,
              name: b.name,
            }))}
            onChange={(id) => set("brand", id)}
            clearLabel="همه برندها"
          />
        </details>
        <Button
          onClick={() => {
            setDraft({ ...emptyProductFilters });
            onChange({ ...emptyProductFilters });
          }}
        >
          پاک‌کردن فیلترهای محصول
        </Button>
      </section>
    </FilterPanel>
  );
}
