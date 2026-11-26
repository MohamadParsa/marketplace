import { useState } from "react";
import { Button, TextField } from "@mui/material";
import { DataTable } from "../../shared/components";
import { useWorkspace } from "../../shared/workspace";
import type { Brand } from "../../shared/types";
import {
  CatalogIcon,
  CatalogIconDialog,
  type IconTarget,
} from "./catalog-icon";

const norm = (v: string) =>
  v.trim().replace(/ي/g, "ی").replace(/ك/g, "ک").toLocaleLowerCase();
export function BrandCatalogGrid({
  onEdit,
}: {
  onEdit: (brand?: Brand) => void;
}) {
  const { state, busy, can } = useWorkspace();
  const [name, setName] = useState(""),
    [english, setEnglish] = useState(""),
    [active, setActive] = useState("all"),
    [image, setImage] = useState("all");
  const [icon, setIcon] = useState<IconTarget>();
  const rows = (state.brands ?? []).filter(
    (b) =>
      norm(b.name).includes(norm(name)) &&
      norm(b.english_name ?? "").includes(norm(english)) &&
      (active === "all" || b.active === (active === "yes")) &&
      (image === "all" || !!b.icon_id === (image === "yes")),
  );
  const choice = (
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
      <section className="catalog-filters" aria-label="فیلترهای برند">
        <div className="catalog-filter-text">
          <TextField
            size="small"
            label="فیلتر نام فارسی برند"
            value={name}
            onChange={(e) => setName(e.target.value)}
          />
          <TextField
            size="small"
            label="فیلتر نام انگلیسی برند"
            value={english}
            onChange={(e) => setEnglish(e.target.value)}
          />
        </div>
        {choice("وضعیت برند", active, setActive, [
          ["all", "همه وضعیت‌ها"],
          ["yes", "فعال"],
          ["no", "غیرفعال"],
        ])}
        {choice("تصویر برند", image, setImage, [
          ["all", "همه تصاویر"],
          ["yes", "دارای تصویر"],
          ["no", "بدون تصویر"],
        ])}
        <div className="toolbar">
          <span aria-live="polite">
            {rows.length.toLocaleString("fa-IR")} برند
          </span>
          <Button
            onClick={() => {
              setName("");
              setEnglish("");
              setActive("all");
              setImage("all");
            }}
          >
            پاک‌کردن فیلترهای برند
          </Button>
          <Button
            variant="contained"
            disabled={busy || !can("catalog.brand.create", "management")}
            onClick={() => onEdit()}
          >
            برند جدید
          </Button>
        </div>
      </section>
      <section aria-label="فهرست برندها" className="category-catalog-grid">
        <DataTable
          key={JSON.stringify([name, english, active, image])}
          rows={rows}
          columns={[
            {
              title: "نام فارسی",
              render: (b) => (
                <div className="catalog-name">
                  <CatalogIcon
                    kind="brand"
                    name={b.name}
                    iconId={b.active ? b.icon_id : undefined}
                  />
                  <b>{b.name}</b>
                </div>
              ),
            },
            {
              title: "نام انگلیسی",
              render: (b) => <span dir="ltr">{b.english_name || "—"}</span>,
            },
            {
              title: "توضیحات",
              render: (b) => (
                <span className="brand-description">
                  {b.description || "—"}
                </span>
              ),
            },
            { title: "وضعیت", render: (b) => (b.active ? "فعال" : "غیرفعال") },
            {
              title: "تصویر",
              render: (b) => (b.icon_id ? "دارای تصویر" : "پیش‌فرض"),
            },
            {
              title: "محصولات متصل",
              render: (b) =>
                (state.products ?? [])
                  .filter((p) => p.brand_id === b.id)
                  .length.toLocaleString("fa-IR"),
            },
            {
              title: "عملیات",
              render: (b) => (
                <div className="catalog-row-actions">
                  <Button
                    disabled={
                      busy || !can("catalog.brand.update", "management")
                    }
                    onClick={() => onEdit(b)}
                  >
                    ویرایش برند
                  </Button>
                  <Button
                    disabled={
                      busy ||
                      !(
                        can("brand.icon.set", "management") ||
                        can("brand.icon.remove", "management")
                      )
                    }
                    onClick={() => setIcon({ ...b, kind: "brand" })}
                  >
                    تصویر برند
                  </Button>
                </div>
              ),
            },
          ]}
        />
      </section>
      {icon && (
        <CatalogIconDialog target={icon} close={() => setIcon(undefined)} />
      )}
    </>
  );
}
