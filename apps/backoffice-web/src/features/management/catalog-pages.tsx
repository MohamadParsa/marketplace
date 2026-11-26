import { CategorySalesDialog } from "../commerce/sales-methods-page";
import { CategoryCatalogGrid } from "./category-catalog-grid";
import { ManagementPage } from "./management-page";
import {
  CatalogIcon,
  CatalogIconDialog,
  type IconTarget,
} from "./catalog-icon";
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
import { PaginationControls, usePagedRows } from "../../shared/pagination";
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
  const [iconTarget, setIconTarget] = useState<IconTarget>();
  const [search, setSearch] = useState("");
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
        : { ...empty, content: emptyProductContent() },
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
  const query = search.toLocaleLowerCase();
  const brandRows = usePagedRows(
    (state.brands ?? []).filter((b) =>
      b.name.toLocaleLowerCase().includes(query),
    ),
  );
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
        {tab === "brands" && (
          <input
            className="search"
            aria-label="جست‌وجو در کاتالوگ"
            placeholder="جست‌وجو در این بخش…"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
          />
        )}
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
          <>
            <div className="toolbar">
              <h2>برندها</h2>
              <Button
                variant="contained"
                disabled={busy || !can("catalog.brand.create", "management")}
                onClick={() => open("brand")}
              >
                برند جدید
              </Button>
            </div>
            <div className="home-admin-list">
              {brandRows.rows.map((b) => (
                <article key={b.id}>
                  <div className="catalog-name">
                    <CatalogIcon
                      kind="brand"
                      name={b.name}
                      iconId={b.icon_id}
                    />
                    <b>{b.name}</b>
                  </div>
                  <span>{b.active ? "فعال" : "غیرفعال"}</span>
                  <Button
                    disabled={
                      busy || !can("catalog.brand.update", "management")
                    }
                    onClick={() => open("brand", b)}
                  >
                    ویرایش برند
                  </Button>
                  <Button
                    disabled={
                      busy || !b.active || !can("brand.icon.set", "management")
                    }
                    onClick={() => setIconTarget({ ...b, kind: "brand" })}
                  >
                    آیکن
                  </Button>
                </article>
              ))}
            </div>
            <PaginationControls {...brandRows.controls} />
          </>
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
      {iconTarget && (
        <CatalogIconDialog
          target={iconTarget}
          close={() => setIconTarget(undefined)}
        />
      )}
      <Dialog
        open={!!editing}
        onClose={() => !busy && closeEditing()}
        fullWidth
        maxWidth="sm"
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
                  {editing.kind !== "brand" && (
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
                      {editing.kind === "category" && (
                        <label>
                          توضیحات
                          <textarea
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
                        <b>دسته *</b>
                        <CategoryTree
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
                          <label>
                            نوع محصول
                            <select
                              value={editing.input.model ?? "shared"}
                              onChange={(e) =>
                                setEditing({
                                  ...editing,
                                  input: {
                                    ...editing.input,
                                    model: e.target.value as
                                      | "shared"
                                      | "independent",
                                    owner_store_id: "",
                                  },
                                })
                              }
                            >
                              <option value="shared">مشترک</option>
                              <option value="independent">مستقل فروشگاه</option>
                            </select>
                          </label>
                          {editing.input.model === "independent" && (
                            <label>
                              فروشگاه مالک
                              <select
                                required
                                value={editing.input.owner_store_id ?? ""}
                                onChange={(e) =>
                                  setEditing({
                                    ...editing,
                                    input: {
                                      ...editing.input,
                                      owner_store_id: e.target.value,
                                    },
                                  })
                                }
                              >
                                <option value="">انتخاب فروشگاه</option>
                                {state.stores
                                  ?.filter((s) => s.active)
                                  .map((s) => (
                                    <option key={s.id} value={s.id}>
                                      {s.name}
                                    </option>
                                  ))}
                              </select>
                            </label>
                          )}
                        </>
                      )}
                      <label>
                        برند
                        <select
                          value={editing.input.brand_id}
                          onChange={(e) =>
                            setEditing({
                              ...editing,
                              input: {
                                ...editing.input,
                                brand_id: e.target.value,
                              },
                            })
                          }
                        >
                          <option value="">بدون برند</option>
                          {(state.brands ?? [])
                            .filter((b) => b.active)
                            .map((b) => (
                              <option key={b.id} value={b.id}>
                                {b.name}
                              </option>
                            ))}
                        </select>
                      </label>
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
                      <legend>ویژگی‌های نمایشی و فیلتر</legend>
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
                            <label key={a.id}>
                              {a.name}
                              {a.unit ? " (" + a.unit + ")" : ""}
                              {a.kind === "enum" ? (
                                <select
                                  value={value}
                                  onChange={(e) => setValue(e.target.value)}
                                >
                                  <option value="">ثبت نشده</option>
                                  {a.options.map((o) => (
                                    <option key={o} value={o}>
                                      {o}
                                    </option>
                                  ))}
                                </select>
                              ) : (
                                <input
                                  type={a.kind === "number" ? "number" : "text"}
                                  step="any"
                                  maxLength={160}
                                  value={value}
                                  onChange={(e) => setValue(e.target.value)}
                                />
                              )}
                            </label>
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
