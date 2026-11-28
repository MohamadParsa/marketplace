import { CatalogChoice } from "./catalog-choice";
import { BrandCatalogGrid } from "./brand-catalog-grid";
import { CategorySalesDialog } from "../commerce/sales-methods-page";
import { CategoryCatalogGrid } from "./category-catalog-grid";
import { ManagementPage } from "./management-page";
import { useEffect, useRef, useState, type FormEvent } from "react";
import {
  Alert,
  Button,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
} from "@mui/material";
import { Link, useSearchParams } from "react-router-dom";
import { useWorkspace } from "../../shared/workspace";
import type {
  CatalogInput,
  HomeSettings,
  Product,
  CategorySEO,
} from "../../shared/types";
import { CategoryTree } from "./category-tree";
import {
  ProductContentFields,
  emptyProductContent,
} from "./product-content-fields";
const empty: CatalogInput = {
  name: "",
  english_name: "",
  description: "",
  category_id: "",
  brand_id: "",
  parent_id: "",
  published: false,
  seo: { meta_title: "", meta_keywords: "", meta_description: "" },
  attributes: [],
};
const slugify = (s: string) =>
  s
    .trim()
    .replace(/([a-z0-9])([A-Z])/g, "$1-$2")
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-+|-+$/g, "");
const titles = ["موبایل", "هندزفری", "پاوربانک", "ساعت هوشمند"];
export function CatalogPage({
  section,
}: {
  section: "products" | "categories" | "brands" | "home";
}) {
  const { state, execute, busy, can, surface } = useWorkspace();
  const [params, setParams] = useSearchParams();
  const opened = useRef("");
  const tab = section === "home" ? "rails" : section;
  const [salesCategory, setSalesCategory] = useState<string>();
  const [error, setError] = useState("");
  const [notice, setNotice] = useState("");
  const [editing, setEditing] = useState<{
    kind: "product" | "category" | "brand";
    id?: string;
    version?: number;
    key: string;
    input: CatalogInput;
  }>();
  const [settings, setSettings] = useState<HomeSettings>(state.home_settings);
  useEffect(() => setSettings(state.home_settings), [state.home_settings]);
  async function run(fn: () => Promise<void>, message = "تغییرات ذخیره شد.") {
    setError("");
    setNotice("");
    try {
      await fn();
      setNotice(message);
    } catch (e) {
      setError(e instanceof Error ? e.message : "ذخیره انجام نشد.");
      throw e;
    }
  }
  function open(
    kind: "product" | "category" | "brand",
    row?: {
      id: string;
      name: string;
      version: number;
      display_priority?: number;
      english_name?: string;
      slug?: string;
      public_code?: string;
      description?: string;
      category_id?: string;
      brand_id?: string;
      parent_id?: string;
      published?: boolean;
      active?: boolean;
      status?: string;
      seo?: CategorySEO;
      content?: Product["content"];
      attributes?: { attribute_id: string; value: string }[];
    },
  ) {
    setError("");
    setEditing({
      key: crypto.randomUUID(),
      kind,
      id: row?.id,
      version: row?.version,
      input: row
        ? {
            active: kind === "product" ? row.active !== false : undefined,
            name: row.name,
            display_priority:
              kind === "category" ? row.display_priority : undefined,
            english_name: row.english_name ?? "",
            slug: kind === "category" ? row.slug : undefined,
            description: row.description ?? "",
            seo: row.seo ?? {
              meta_title: "",
              meta_keywords: "",
              meta_description: "",
            },
            content: row.content ?? {
              ...emptyProductContent(),
              full_description: row.description
                ? "<p>" +
                  row.description
                    .replace(/&/g, "&amp;")
                    .replace(/</g, "&lt;")
                    .replace(/>/g, "&gt;")
                    .replace(/\n/g, "<br>") +
                  "</p>"
                : "",
            },
            attributes: row.attributes ?? [],
            category_id: row.category_id ?? "",
            brand_id: row.brand_id ?? "",
            parent_id: row.parent_id ?? "",
            published:
              row.published ?? row.active ?? row.status === "published",
          }
        : {
            ...empty,
            active: kind === "product" ? true : undefined,
            content: emptyProductContent(),
          },
    });
  }
  const closeEditing = () => {
    setEditing(undefined);
    const next = new URLSearchParams(params);
    for (const key of ["create", "category", "brand", "product"])
      next.delete(key);
    setParams(next, { replace: true });
  };
  useEffect(() => {
    if (section === "home" || surface !== "backoffice") return;
    const kind =
      section === "categories"
        ? "category"
        : section === "brands"
          ? "brand"
          : "product";
    const create = params.get("create") === kind;
    const id = params.get(kind);
    const marker = create ? `new-${kind}` : id;
    if (!marker) {
      opened.current = "";
      return;
    }
    if (opened.current === marker) return;
    if (create) {
      opened.current = marker;
      open(kind);
      return;
    }
    const row = (
      kind === "category"
        ? state.categories
        : kind === "brand"
          ? state.brands
          : state.products
    )?.find((r) => r.id === id);
    if (row) {
      opened.current = marker;
      open(kind, row);
    }
  }, [
    params,
    section,
    surface,
    state.categories,
    state.products,
    state.brands,
  ]);
  async function save(e: FormEvent) {
    e.preventDefault();
    if (
      surface !== "backoffice" ||
      !editing ||
      !can(
        `catalog.${editing.kind}.${editing.id ? "update" : "create"}`,
        "management",
      )
    )
      return;
    if (
      editing.kind === "product" &&
      (!editing.input.category_id ||
        (editing.input.model === "independent" &&
          !editing.input.owner_store_id))
    ) {
      setError("دسته و در محصول مستقل، فروشگاه مالک را انتخاب کنید.");
      return;
    }
    const original = state.categories?.find((c) => c.id === editing.id);
    if (
      editing.kind === "category" &&
      original?.published &&
      editing.input.slug !== original.slug &&
      !window.confirm(
        "تغییر آدرس دسته منتشرشده؟ آدرس قبلی به آدرس جدید هدایت می‌شود.",
      )
    )
      return;
    try {
      await run(() =>
        execute({
          action: `catalog.${editing.kind}.${editing.id ? "update" : "create"}`,
          target: editing.id,
          version: editing.version,
          idempotency_key: editing.key,
          catalog: {
            ...editing.input,
            confirm_slug_change:
              editing.kind === "category" &&
              !!original?.published &&
              editing.input.slug !== original.slug,
            content:
              editing.kind === "product" && editing.input.content
                ? {
                    ...editing.input.content,
                    advantages: editing.input.content.advantages
                      .map((s) => s.trim())
                      .filter(Boolean),
                    disadvantages: editing.input.content.disadvantages
                      .map((s) => s.trim())
                      .filter(Boolean),
                  }
                : undefined,
          },
        }),
      );
      closeEditing();
    } catch {
      /* form stays open */
    }
  }
  return (
    <>
      {section !== "products" && (
        <div className="page-header">
          <div>
            <h1>
              {
                {
                  categories: "دسته‌بندی‌ها",
                  brands: "برندها",
                  home: "مدیریت صفحه اصلی",
                }[section]
              }
            </h1>
            <p>
              {section === "home"
                ? "چهار کروسل ثابت و پیوند مدیریت بنرها"
                : section === "categories"
                  ? "ساختار درختی، انتشار و خصوصیات مستقل هر دسته"
                  : "ایجاد، ویرایش و آیکن برندهای کاتالوگ"}
            </p>
          </div>
        </div>
      )}
      {error && (
        <Alert severity="error" onClose={() => setError("")}>
          {error}
        </Alert>
      )}
      {notice && (
        <Alert severity="success" onClose={() => setNotice("")}>
          {notice}
        </Alert>
      )}
      <section
        className={
          section === "categories" || section === "products"
            ? "catalog-page-content"
            : "panel home-admin-panel"
        }
      >
        {tab === "products" && (
          <ManagementPage
            section="products"
            onCatalogEdit={(row) => open("product", row)}
          />
        )}
        {tab === "categories" && (
          <CategoryCatalogGrid onEdit={(row) => open("category", row)} />
        )}
        {tab === "brands" && (
          <BrandCatalogGrid onEdit={(row) => open("brand", row)} />
        )}
        {tab === "rails" && (
          <>
            <h2>چهار کروسل صفحه اصلی</h2>
            <p>
              ترتیب و نوع ریل‌ها ثابت است. هر ریل فقط محصولات منتشرشده و موجود
              دسته انتخاب‌شده را نشان می‌دهد؛ کالای تکراری اضافه نمی‌شود.
            </p>
            <form
              onSubmit={(e) => {
                e.preventDefault();
                void run(() =>
                  execute({
                    action: "homepage.save",
                    target: "home",
                    version: settings.version,
                    home: settings,
                  }),
                ).catch(() => {});
              }}
              className="home-admin-form"
            >
              {settings.rails.map((r, i) => (
                <div className="home-admin-rail" key={r.key}>
                  <b>
                    {i + 1}. {titles[i]}
                  </b>
                  <div>
                    دسته مرتبط
                    <CategoryTree
                      categories={(state.categories ?? []).filter(
                        (c) => c.published && !c.archived,
                      )}
                      value={r.category_id}
                      allowRoot={false}
                      disabled={busy || !can("homepage.save", "management")}
                      onChange={(id) =>
                        setSettings({
                          ...settings,
                          rails: settings.rails.map((x, k) =>
                            k === i ? { ...x, category_id: id } : x,
                          ),
                        })
                      }
                    />
                  </div>
                  <label>
                    حداکثر کالا
                    <input
                      type="number"
                      min="1"
                      max="10"
                      value={r.limit}
                      onChange={(e) =>
                        setSettings({
                          ...settings,
                          rails: settings.rails.map((x, k) =>
                            k === i
                              ? { ...x, limit: Number(e.target.value) }
                              : x,
                          ),
                        })
                      }
                    />
                  </label>
                </div>
              ))}
              <Button
                variant="contained"
                type="submit"
                disabled={busy || !can("homepage.save", "management")}
              >
                ذخیره چهار کروسل
              </Button>
            </form>
            <Link to="/banners">مدیریت بنر دسکتاپ و موبایل</Link>
          </>
        )}
      </section>
      <Dialog
        open={!!editing}
        onClose={() => !busy && closeEditing()}
        fullWidth
        maxWidth="xl"
        dir="rtl"
      >
        <form onSubmit={save}>
          <DialogTitle>
            {editing?.id ? "ویرایش" : "ایجاد"}{" "}
            {editing?.kind === "product"
              ? "محصول"
              : editing?.kind === "category"
                ? "دسته"
                : "برند"}
          </DialogTitle>
          <DialogContent>
            {error && <Alert severity="error">{error}</Alert>}
            <fieldset
              className="home-admin-form"
              disabled={
                busy ||
                !editing ||
                !can(
                  `catalog.${editing.kind}.${editing.id ? "update" : "create"}`,
                  "management",
                )
              }
            >
              {editing && (
                <>
                  <label>
                    نام فارسی *
                    <input
                      required
                      maxLength={160}
                      value={editing.input.name}
                      onChange={(e) =>
                        setEditing({
                          ...editing,
                          input: { ...editing.input, name: e.target.value },
                        })
                      }
                    />
                  </label>
                  <>
                    <label>
                      نام انگلیسی
                      <input
                        required={editing.kind === "category"}
                        dir="ltr"
                        maxLength={160}
                        value={editing.input.english_name}
                        onChange={(e) =>
                          setEditing({
                            ...editing,
                            input: {
                              ...editing.input,
                              english_name: e.target.value,
                            },
                          })
                        }
                      />
                    </label>
                    {editing.kind !== "product" && (
                      <label>
                        توضیحات
                        <textarea
                          aria-label="توضیحات"
                          maxLength={4000}
                          rows={4}
                          value={editing.input.description}
                          onChange={(e) =>
                            setEditing({
                              ...editing,
                              input: {
                                ...editing.input,
                                description: e.target.value,
                              },
                            })
                          }
                        />
                      </label>
                    )}
                  </>
                  {editing.kind === "brand" && (
                    <>
                      <Alert
                        severity={editing.input.published ? "info" : "warning"}
                      >
                        {editing.input.published
                          ? "برند فعال در نمایش عمومی استفاده می‌شود."
                          : "محصولات این برند تا فعال‌شدن آن از نمایش عمومی خارج می‌شوند؛ اتصال محصولات و سفارش‌های گذشته محفوظ می‌مانند."}
                        {editing.id &&
                          ` تعداد محصولات متصل: ${(state.products ?? []).filter((p) => p.brand_id === editing.id).length.toLocaleString("fa-IR")}`}
                      </Alert>
                      <small>
                        تصویر از دکمه «تصویر برند» روی ردیف مدیریت می‌شود؛ ابتدا
                        اطلاعات برند را ذخیره کنید.
                      </small>
                    </>
                  )}
                  {editing.kind === "category" && editing.id && (
                    <Button
                      type="button"
                      onClick={() => setSalesCategory(editing.id)}
                    >
                      روش‌های فروش و مارکاپ دسته
                    </Button>
                  )}
                  {editing.kind === "category" && (
                    <label>
                      آدرس انگلیسی دسته (slug)
                      <input
                        required
                        dir="ltr"
                        maxLength={100}
                        pattern="[a-z0-9]+(-[a-z0-9]+)*"
                        value={
                          editing.input.slug ??
                          slugify(editing.input.english_name)
                        }
                        onChange={(e) =>
                          setEditing({
                            ...editing,
                            input: { ...editing.input, slug: e.target.value },
                          })
                        }
                      />
                      <small dir="ltr">
                        /category/
                        {editing.input.slug ??
                          slugify(editing.input.english_name)}
                      </small>
                      <small>
                        تغییر نام انگلیسی، آدرس ذخیره‌شده را عوض نمی‌کند.
                        آدرس‌های قبلی محفوظ می‌مانند.
                      </small>
                    </label>
                  )}
                  {editing.kind === "product" && editing.id && (
                    <label>
                      کد عمومی محصول (فقط‌خواندنی)
                      <input
                        aria-label="کد عمومی محصول (فقط‌خواندنی)"
                        readOnly
                        dir="ltr"
                        value={
                          state.products?.find((p) => p.id === editing.id)
                            ?.public_code ?? "پس از ذخیره تخصیص می‌یابد"
                        }
                      />
                      <small dir="ltr">
                        /product/
                        {
                          state.products?.find((p) => p.id === editing.id)
                            ?.public_code
                        }
                      </small>
                    </label>
                  )}
                  {editing.kind === "category" && (
                    <label>
                      اولویت نمایش دسته
                      <input
                        type="number"
                        min={0}
                        max={1000000}
                        step={1}
                        placeholder="بدون اولویت"
                        value={editing.input.display_priority ?? ""}
                        onChange={(event) =>
                          setEditing({
                            ...editing,
                            input: {
                              ...editing.input,
                              display_priority:
                                event.target.value === ""
                                  ? undefined
                                  : Number(event.target.value),
                              clear_display_priority: event.target.value === "",
                            },
                          })
                        }
                      />
                      <small>
                        عدد کوچک‌تر مقدم است؛ خالی یعنی پس از دسته‌های
                        اولویت‌دار در همین سطح.
                      </small>
                    </label>
                  )}
                  {editing.kind === "category" && (
                    <fieldset className="home-admin-form">
                      <legend>سئوی فهرست این دسته</legend>
                      {(
                        [
                          ["meta_title", "عنوان متا", 160],
                          ["meta_keywords", "کلمات کلیدی متا", 500],
                          ["meta_description", "توضیحات متا", 400],
                        ] as const
                      ).map(([key, label, max]) => (
                        <label key={key}>
                          {label}
                          <input
                            maxLength={max}
                            value={editing.input.seo?.[key] ?? ""}
                            onChange={(e) =>
                              setEditing({
                                ...editing,
                                input: {
                                  ...editing.input,
                                  seo: {
                                    meta_title: "",
                                    meta_keywords: "",
                                    meta_description: "",
                                    ...editing.input.seo,
                                    [key]: e.target.value,
                                  },
                                },
                              })
                            }
                          />
                        </label>
                      ))}
                      <small>
                        این مقادیر فقط در PLP همین دسته استفاده می‌شوند.
                      </small>
                    </fieldset>
                  )}
                  {editing.kind === "product" && (
                    <>
                      <div>
                        <CategoryTree
                          label="دسته"
                          required
                          categories={state.categories ?? []}
                          value={editing.input.category_id}
                          disabled={!!editing.id}
                          allowRoot={false}
                          onChange={(id) =>
                            setEditing({
                              ...editing,
                              input: { ...editing.input, category_id: id },
                            })
                          }
                        />
                      </div>
                      {!editing.id && (
                        <>
                          <CatalogChoice
                            label="نوع محصول"
                            value={editing.input.model ?? "shared"}
                            options={[
                              { id: "shared", name: "مشترک" },
                              { id: "independent", name: "مستقل فروشگاه" },
                            ]}
                            onChange={(v) =>
                              setEditing({
                                ...editing,
                                input: {
                                  ...editing.input,
                                  model: v as "shared" | "independent",
                                  owner_store_id: "",
                                },
                              })
                            }
                          />
                          {editing.input.model === "independent" && (
                            <CatalogChoice
                              label="فروشگاه مالک"
                              value={editing.input.owner_store_id ?? ""}
                              options={(state.stores ?? [])
                                .filter((st) => st.active)
                                .map((st) => ({ id: st.id, name: st.name }))}
                              onChange={(v) =>
                                setEditing({
                                  ...editing,
                                  input: {
                                    ...editing.input,
                                    owner_store_id: v,
                                  },
                                })
                              }
                            />
                          )}
                        </>
                      )}
                      <CatalogChoice
                        label="برند"
                        value={editing.input.brand_id}
                        options={(state.brands ?? [])
                          .filter(
                            (b) => b.active || b.id === editing.input.brand_id,
                          )
                          .map((b) => ({
                            id: b.id,
                            name: b.name + (b.active ? "" : " (غیرفعال)"),
                          }))}
                        clearLabel="بدون برند"
                        onChange={(v) =>
                          setEditing({
                            ...editing,
                            input: { ...editing.input, brand_id: v },
                          })
                        }
                      />
                      <label className="home-admin-check">
                        <input
                          type="checkbox"
                          checked={editing.input.active !== false}
                          onChange={(e) =>
                            setEditing({
                              ...editing,
                              input: {
                                ...editing.input,
                                active: e.target.checked,
                              },
                            })
                          }
                        />
                        محصول فعال و مجاز به خرید باشد
                      </label>
                      {editing.input.active === false && (
                        <Alert severity="warning">
                          محصول از خانه حذف می‌شود؛ در PLP و PDP با برچسب
                          «غیرقابل خرید» و بدون قیمت و خرید باقی می‌ماند. انتشار
                          محتوا و سفارش گذشته محفوظ‌اند.
                        </Alert>
                      )}
                      {!editing.id && (
                        <div className="catalog-code-guide">
                          <b>کد عمومی خودکار</b>
                          <code dir="ltr">chr-[id]</code>
                          <small>
                            پس از ذخیره تخصیص می‌یابد؛ شناسه مبدأ یا کد دلخواه
                            وارد نکنید. تصویر، تنوع و قیمت از اکشن‌های ردیف
                            تکمیل می‌شوند.
                          </small>
                        </div>
                      )}
                    </>
                  )}
                  {editing.kind === "product" && (
                    <ProductContentFields
                      value={editing.input.content ?? emptyProductContent()}
                      onChange={(content) =>
                        setEditing({
                          ...editing,
                          input: { ...editing.input, content },
                        })
                      }
                    />
                  )}
                  {editing.kind === "product" && (
                    <fieldset className="home-admin-form">
                      <legend>مقادیر خصوصیات همین دسته</legend>
                      <p>
                        این فهرست فقط از خصوصیات غیرمحورِ منتسب به دسته
                        انتخاب‌شده می‌آید. نمایش روی کارت و فیلتر در تنظیم دسته
                        تعیین می‌شود؛ رنگ/گارانتی و سایر محورهای تنوع در فرم
                        تنوع ثبت می‌شوند.
                      </p>
                      {editing.input.category_id && (
                        <Link
                          to={
                            "/categories?category=" +
                            encodeURIComponent(editing.input.category_id)
                          }
                        >
                          مشاهده تنظیم دسته و خصوصیات
                        </Link>
                      )}
                      {(
                        state.categories?.find(
                          (c) => c.id === editing.input.category_id,
                        )?.bindings ?? []
                      )
                        .filter((b) => !b.axis)
                        .map((b) => {
                          const a = state.attributes?.find(
                            (x) => x.id === b.attribute_id,
                          );
                          if (!a) return null;
                          const value =
                            editing.input.attributes?.find(
                              (x) => x.attribute_id === a.id,
                            )?.value ?? "";
                          const setValue = (v: string) =>
                            setEditing({
                              ...editing,
                              input: {
                                ...editing.input,
                                attributes: [
                                  ...(editing.input.attributes ?? []).filter(
                                    (x) => x.attribute_id !== a.id,
                                  ),
                                  ...(v
                                    ? [{ attribute_id: a.id, value: v }]
                                    : []),
                                ],
                              },
                            });
                          return (
                            <div className="home-admin-form" key={a.id}>
                              <b>
                                {a.name}
                                {a.unit ? " (" + a.unit + ")" : ""}
                              </b>
                              {a.kind === "enum" ? (
                                <CatalogChoice
                                  label={a.name}
                                  value={value}
                                  options={a.options.map((o) => ({
                                    id: o,
                                    name: o,
                                  }))}
                                  clearLabel="ثبت نشده"
                                  onChange={setValue}
                                />
                              ) : (
                                <input
                                  aria-label={a.name}
                                  type={a.kind === "number" ? "number" : "text"}
                                  step="any"
                                  maxLength={160}
                                  value={value}
                                  onChange={(e) => setValue(e.target.value)}
                                />
                              )}
                            </div>
                          );
                        })}
                      <small>
                        ویژگی‌ها فقط از انتساب صریح همین دسته می‌آیند. محورهای
                        تنوع از فرم تنوع مدیریت می‌شوند.
                      </small>
                    </fieldset>
                  )}
                  {editing.kind === "category" && (
                    <div>
                      <b>دسته والد و ساختار درختی</b>
                      <CategoryTree
                        categories={state.categories ?? []}
                        value={editing.input.parent_id}
                        currentId={editing.id}
                        disabled={!!editing.id}
                        onChange={(id) =>
                          setEditing({
                            ...editing,
                            input: { ...editing.input, parent_id: id },
                          })
                        }
                      />
                    </div>
                  )}
                  {(editing.kind !== "product" || editing.id) && (
                    <label className="home-admin-check">
                      <input
                        type="checkbox"
                        disabled={
                          (editing.kind === "product" &&
                            !can("product.publish", "management")) ||
                          (editing.kind === "category" &&
                            !can("category.publish", "management"))
                        }
                        checked={editing.input.published}
                        onChange={(e) =>
                          setEditing({
                            ...editing,
                            input: {
                              ...editing.input,
                              published: e.target.checked,
                            },
                          })
                        }
                      />
                      {editing.kind === "brand"
                        ? "برند فعال باشد"
                        : "منتشر شود"}
                    </label>
                  )}
                  {editing.kind === "product" && (
                    <small>
                      انتشار به نام، دسته منتشرشده، تصویر معتبر و عرضه موجود
                      نیاز دارد. تغییر دسته دارای تنوع در این مرحله فعال نیست.
                    </small>
                  )}
                </>
              )}
            </fieldset>
          </DialogContent>
          <DialogActions>
            <Button disabled={busy} onClick={closeEditing}>
              انصراف
            </Button>
            <Button
              variant="contained"
              type="submit"
              disabled={
                busy ||
                !editing ||
                !can(
                  `catalog.${editing.kind}.${editing.id ? "update" : "create"}`,
                  "management",
                )
              }
            >
              ذخیره
            </Button>
          </DialogActions>
        </form>
      </Dialog>
      {salesCategory && (
        <CategorySalesDialog
          id={salesCategory}
          close={() => setSalesCategory(undefined)}
        />
      )}
    </>
  );
}
