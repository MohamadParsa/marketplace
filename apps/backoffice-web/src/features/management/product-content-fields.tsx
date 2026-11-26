import { lazy, Suspense } from "react";
const RichProductEditor = lazy(() =>
  import("./rich-product-editor").then((m) => ({
    default: m.RichProductEditor,
  })),
);
import type { ProductContent } from "../../shared/types";
export const emptyProductContent = (): ProductContent => ({
  short_description: "",
  full_description: "",
  advantages: [],
  disadvantages: [],
  reviews_enabled: false,
  qa_enabled: false,
  shipping: { weight_unit: "g", dimension_unit: "mm" },
  seo: { meta_title: "", meta_keywords: "", meta_description: "" },
  noindex: false,
  nofollow: false,
});
export function ProductContentFields({
  value,
  onChange,
}: {
  value: ProductContent;
  onChange: (c: ProductContent) => void;
}) {
  const set = (patch: Partial<ProductContent>) =>
    onChange({ ...value, ...patch });
  return (
    <>
      <fieldset className="home-admin-form">
        <legend>محتوای صفحه محصول</legend>
        <label>
          توضیح کوتاه
          <textarea
            aria-label="توضیح کوتاه"
            maxLength={2000}
            rows={3}
            value={value.short_description}
            onChange={(e) => set({ short_description: e.target.value })}
          />
        </label>
        <Suspense fallback={<p role="status">در حال آماده‌سازی ویرایشگر…</p>}>
          <RichProductEditor
            value={value.full_description}
            onChange={(full_description) => set({ full_description })}
          />
        </Suspense>
        <small>
          پاراگراف، عنوان، فهرست، تأکید و لینک؛ تصویر از گالری مدیریت می‌شود.
        </small>
        {(
          [
            ["advantages", "مزایا"],
            ["disadvantages", "معایب"],
          ] as const
        ).map(([key, label]) => (
          <label key={key}>
            {label}
            <textarea
              aria-label={label}
              rows={3}
              value={value[key].join("\n")}
              onChange={(e) => set({ [key]: e.target.value.split("\n") })}
            />
            <small>
              هر مورد در یک خط؛ حداکثر۲۰ مورد و۳۰۰ نویسه برای هر مورد.
            </small>
          </label>
        ))}
        {(
          [
            ["reviews_enabled", "اجازه ثبت نظر مشتری"],
            ["qa_enabled", "اجازه پرسش و پاسخ"],
          ] as const
        ).map(([key, label]) => (
          <label className="home-admin-check" key={key}>
            <input
              type="checkbox"
              checked={value[key]}
              onChange={(e) => set({ [key]: e.target.checked })}
            />
            {label}
          </label>
        ))}
        <small>
          تنظیم‌ها مستقل‌اند؛ انتشار سرویس ثبت نظر و پرسش به پذیرش قابلیت مربوط
          وابسته است.
        </small>
      </fieldset>
      <fieldset className="home-admin-form">
        <legend>وزن و ابعاد حمل</legend>
        <small>
          این وزن برای ارسال است و در محاسبه قیمت طلا استفاده نمی‌شود.
        </small>
        <label>
          واحد وزن
          <select
            value={value.shipping.weight_unit}
            onChange={(e) =>
              set({
                shipping: { ...value.shipping, weight_unit: e.target.value },
              })
            }
          >
            <option value="">انتخاب واحد</option>
            <option value="g">گرم</option>
          </select>
        </label>
        <label>
          واحد ابعاد
          <select
            value={value.shipping.dimension_unit}
            onChange={(e) =>
              set({
                shipping: { ...value.shipping, dimension_unit: e.target.value },
              })
            }
          >
            <option value="">انتخاب واحد</option>
            <option value="mm">میلی‌متر</option>
          </select>
        </label>
        {(
          [
            ["weight", "وزن حمل"],
            ["length", "طول"],
            ["width", "عرض"],
            ["height", "ارتفاع"],
          ] as const
        ).map(([key, label]) => (
          <label key={key}>
            {label}
            <input
              type="number"
              min="0.001"
              max="1000000000"
              step="0.001"
              value={value.shipping[key] ?? ""}
              onChange={(e) => {
                const shipping = { ...value.shipping };
                if (e.target.value === "") delete shipping[key];
                else shipping[key] = Number(e.target.value);
                set({ shipping });
              }}
            />
          </label>
        ))}
      </fieldset>
      <fieldset className="home-admin-form">
        <legend>سئوی صفحه محصول</legend>
        {(
          [
            ["meta_title", "عنوان متای محصول", 160],
            ["meta_keywords", "کلمات کلیدی متای محصول", 500],
            ["meta_description", "توضیحات متای محصول", 400],
          ] as const
        ).map(([key, label, max]) => (
          <label key={key}>
            {label}
            <input
              maxLength={max}
              value={value.seo[key]}
              onChange={(e) =>
                set({ seo: { ...value.seo, [key]: e.target.value } })
              }
            />
          </label>
        ))}
        {(
          [
            ["noindex", "عدم ایندکس (noindex)"],
            ["nofollow", "عدم دنبال‌کردن لینک‌ها (nofollow)"],
          ] as const
        ).map(([key, label]) => (
          <label className="home-admin-check" key={key}>
            <input
              type="checkbox"
              checked={value[key]}
              onChange={(e) => set({ [key]: e.target.checked })}
            />
            {label}
          </label>
        ))}
      </fieldset>
    </>
  );
}
