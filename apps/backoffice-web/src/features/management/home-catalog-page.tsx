import { useEffect, useState, type FormEvent } from "react";
import {
  Alert,
  Button,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
} from "@mui/material";
import { Link } from "react-router-dom";
import { useWorkspace } from "../../shared/workspace";
import type {
  CatalogInput,
  HomeSettings,
  Product,
  CategorySEO,
} from "../../shared/types";
import { ProductQuickDialog } from "./product-workspace";
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
const titles = ["موبایل", "هندزفری", "پاوربانک", "ساعت هوشمند"];
function ManagedPhoto({ id }: { id?: string }) {
  const { loadCatalogMedia } = useWorkspace();
  const [url, setUrl] = useState("");
  useEffect(() => {
    let active = true;
    let current = "";
    setUrl("");
    if (id)
      void loadCatalogMedia(id)
        .then((b) => {
          current = URL.createObjectURL(b);
          if (active) setUrl(current);
          else URL.revokeObjectURL(current);
        })
        .catch(() => {});
    return () => {
      active = false;
      if (current) URL.revokeObjectURL(current);
    };
  }, [id, loadCatalogMedia]);
  return url ? (
    <img className="home-admin-photo" src={url} alt="تصویر اصلی محصول" />
  ) : (
    <span className="home-admin-photo empty">بدون تصویر</span>
  );
}
export function HomeCatalogPage() {
  const { state, execute, busy, can, uploadIcon } = useWorkspace();
  const [tab, setTab] = useState("products");
  const [error, setError] = useState("");
  const [notice, setNotice] = useState("");
  const [editing, setEditing] = useState<{
    kind: "product" | "category" | "brand";
    id?: string;
    version?: number;
    key: string;
    input: CatalogInput;
  }>();
  const [quick, setQuick] = useState<string>();
  const [image, setImage] = useState<Product>();
  const [file, setFile] = useState<File>();
  const [preview, setPreview] = useState("");
  const [imageKey, setImageKey] = useState(crypto.randomUUID());
  const [settings, setSettings] = useState<HomeSettings>(state.home_settings);
  useEffect(() => setSettings(state.home_settings), [state.home_settings]);
  useEffect(() => {
    if (!file) {
      setPreview("");
      return;
    }
    const url = URL.createObjectURL(file);
    setPreview(url);
    return () => URL.revokeObjectURL(url);
  }, [file]);
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
      english_name?: string;
      description?: string;
      category_id?: string;
      brand_id?: string;
      parent_id?: string;
      published?: boolean;
      active?: boolean;
      status?: string;
      seo?: CategorySEO;
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
            english_name: row.english_name ?? "",
            description: row.description ?? "",
            seo: row.seo ?? {
              meta_title: "",
              meta_keywords: "",
              meta_description: "",
            },
            attributes: row.attributes ?? [],
            category_id: row.category_id ?? "",
            brand_id: row.brand_id ?? "",
            parent_id: row.parent_id ?? "",
            published:
              row.published ?? row.active ?? row.status === "published",
          }
        : { ...empty },
    });
  }
  async function save(e: FormEvent) {
    e.preventDefault();
    if (!editing) return;
    try {
      await run(() =>
        execute({
          action: `catalog.${editing.kind}.${editing.id ? "update" : "create"}`,
          target: editing.id,
          version: editing.version,
          idempotency_key: editing.key,
          catalog: editing.input,
        }),
      );
      setEditing(undefined);
    } catch {
      /* form stays open */
    }
  }
  const categoryLabel = (id: string) => {
    const names: string[] = [];
    const seen = new Set<string>();
    let c = state.categories?.find((c) => c.id === id);
    while (c && !seen.has(c.id)) {
      seen.add(c.id);
      names.push(c.name);
      c = state.categories?.find((x) => x.id === c?.parent_id);
    }
    return names.reverse().join(" / ");
  };
  const editable = can("catalog.product.update", "management");
  return (
    <>
      <div className="page-header">
        <div>
          <h1>کاتالوگ و صفحه اصلی</h1>
          <p>مدیریت کالا و چهار ریل ثابت صفحه اصلی</p>
        </div>
      </div>
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
      <div className="home-admin-tabs" role="tablist" aria-label="بخش کاتالوگ">
        {[
          ["products", "محصولات"],
          ["categories", "دسته‌ها"],
          ["brands", "برندها"],
          ["rails", "چهار کروسل"],
        ].map(([key, label]) => (
          <button
            key={key}
            role="tab"
            aria-selected={tab === key}
            onClick={() => setTab(key)}
          >
            {label}
          </button>
        ))}
      </div>
      <section className="panel home-admin-panel">
        {tab === "products" && (
          <>
            <div className="toolbar">
              <h2>محصولات</h2>
              <Button
                disabled={busy || !can("catalog.product.create", "management")}
                variant="contained"
                onClick={() => open("product")}
              >
                محصول جدید
              </Button>
            </div>
            <p className="muted">
              ابتدا محصول را بسازید، تصویر و عرضه را ثبت کنید و سپس منتشر کنید.
              قیمت عرضه، قیمت نهایی نقدی مشتری بدون ارسال است.
            </p>
            <div className="home-admin-list">
              {(state.products ?? []).map((p) => (
                <article key={p.id}>
                  <ManagedPhoto id={p.image_id} />
                  <div>
                    <b>{p.name}</b>
                    <small>
                      {p.status === "published"
                        ? "منتشرشده"
                        : "پیش‌نویس / در انتظار بررسی"}{" "}
                      ·{" "}
                      {
                        (state.categories ?? []).find(
                          (c) => c.id === p.category_id,
                        )?.name
                      }
                    </small>
                    <div className="home-admin-actions">
                      <Button
                        disabled={!editable || busy}
                        onClick={() => open("product", p)}
                      >
                        ویرایش محتوا
                      </Button>
                      <Button onClick={() => setQuick(p.id)}>
                        تنوع، قیمت و موجودی
                      </Button>
                      <Button
                        disabled={
                          !can("product.icon.set", "management") || busy
                        }
                        onClick={() => {
                          setImage(p);
                          setFile(undefined);
                        }}
                      >
                        تصویر اصلی
                      </Button>
                    </div>
                  </div>
                </article>
              ))}
              {!state.products?.length && <p>هنوز محصولی ثبت نشده است.</p>}
            </div>
          </>
        )}
        {tab === "categories" && (
          <>
            <div className="toolbar">
              <h2>دسته‌ها</h2>
              <Button
                variant="contained"
                disabled={busy || !can("catalog.category.create", "management")}
                onClick={() => open("category")}
              >
                دسته جدید
              </Button>
            </div>
            <div className="home-admin-list">
              {(state.categories ?? [])
                .filter((c) => !c.archived)
                .map((c) => (
                  <article key={c.id}>
                    <div>
                      <b>{c.name}</b>
                      <small>{categoryLabel(c.id)}</small>
                      <small>
                        {c.published ? "منتشرشده" : "پیش‌نویس"} ·{" "}
                        {c.english_name || "نام انگلیسی ثبت نشده"}
                      </small>
                    </div>
                    <Button
                      disabled={
                        busy || !can("catalog.category.update", "management")
                      }
                      onClick={() => open("category", c)}
                    >
                      ویرایش دسته
                    </Button>
                    <Link to="/categories">آیکن و خصوصیات</Link>
                  </article>
                ))}
            </div>
          </>
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
              {(state.brands ?? []).map((b) => (
                <article key={b.id}>
                  <b>{b.name}</b>
                  <span>{b.active ? "فعال" : "غیرفعال"}</span>
                  <Button
                    disabled={
                      busy || !can("catalog.brand.update", "management")
                    }
                    onClick={() => open("brand", b)}
                  >
                    ویرایش برند
                  </Button>
                </article>
              ))}
            </div>
            <Link to="/brands">مدیریت آیکن برندها</Link>
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
                  <label>
                    دسته مرتبط
                    <select
                      aria-label={`دسته ${titles[i]}`}
                      value={r.category_id}
                      onChange={(e) =>
                        setSettings({
                          ...settings,
                          rails: settings.rails.map((x, k) =>
                            k === i ? { ...x, category_id: e.target.value } : x,
                          ),
                        })
                      }
                    >
                      <option value="">انتخاب نشده</option>
                      {(state.categories ?? [])
                        .filter((c) => c.published && !c.archived)
                        .map((c) => (
                          <option key={c.id} value={c.id}>
                            {categoryLabel(c.id)}
                          </option>
                        ))}
                    </select>
                  </label>
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
        onClose={() => !busy && setEditing(undefined)}
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
            <div className="home-admin-form">
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
                    </>
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
                      <label>
                        دسته *
                        <select
                          required
                          disabled={!!editing.id}
                          value={editing.input.category_id}
                          onChange={(e) =>
                            setEditing({
                              ...editing,
                              input: {
                                ...editing.input,
                                category_id: e.target.value,
                              },
                            })
                          }
                        >
                          <option value="">انتخاب دسته</option>
                          {(state.categories ?? [])
                            .filter((c) => !c.archived)
                            .map((c) => (
                              <option key={c.id} value={c.id}>
                                {categoryLabel(c.id)}
                              </option>
                            ))}
                        </select>
                      </label>
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
                    <label>
                      دسته والد
                      <select
                        disabled={!!editing.id}
                        value={editing.input.parent_id}
                        onChange={(e) =>
                          setEditing({
                            ...editing,
                            input: {
                              ...editing.input,
                              parent_id: e.target.value,
                            },
                          })
                        }
                      >
                        <option value="">بدون والد</option>
                        {(state.categories ?? [])
                          .filter((c) => !c.archived)
                          .map((c) => (
                            <option key={c.id} value={c.id}>
                              {c.name}
                            </option>
                          ))}
                      </select>
                    </label>
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
            </div>
          </DialogContent>
          <DialogActions>
            <Button disabled={busy} onClick={() => setEditing(undefined)}>
              انصراف
            </Button>
            <Button variant="contained" type="submit" disabled={busy}>
              ذخیره
            </Button>
          </DialogActions>
        </form>
      </Dialog>
      <Dialog
        open={!!image}
        onClose={() => !busy && setImage(undefined)}
        fullWidth
        maxWidth="xs"
        dir="rtl"
      >
        <DialogTitle>تصویر اصلی محصول</DialogTitle>
        <DialogContent>
          {error && <Alert severity="error">{error}</Alert>}
          <p>{image?.name}</p>
          {preview ? (
            <img
              className="home-admin-upload-preview"
              src={preview}
              alt="پیش‌نمایش تصویر انتخاب‌شده"
            />
          ) : (
            <ManagedPhoto id={image?.image_id} />
          )}
          <p>
            PNG، JPEG یا WebP؛ حداکثر۵MiB و هر ضلع۴۰۹۶پیکسل؛ تصویر با حفظ نسبت
            تا۱۰۲۴پیکسل آماده می‌شود.
          </p>
          <input
            type="file"
            accept="image/png,image/jpeg,image/webp"
            aria-label="انتخاب تصویر محصول"
            onChange={(e) => {
              setFile(e.target.files?.[0]);
              setImageKey(crypto.randomUUID());
            }}
          />
        </DialogContent>
        <DialogActions>
          <Button disabled={busy} onClick={() => setImage(undefined)}>
            انصراف
          </Button>
          <Button
            variant="contained"
            disabled={!file || !image || busy}
            onClick={() => {
              if (file && image)
                void run(() =>
                  uploadIcon({
                    kind: "product",
                    target: image.id,
                    version: image.version,
                    key: imageKey,
                    file,
                  }),
                )
                  .then(() => setImage(undefined))
                  .catch(() => {});
            }}
          >
            ذخیره تصویر
          </Button>
        </DialogActions>
      </Dialog>
      {quick && (
        <ProductQuickDialog
          id={quick}
          initialTab={1}
          close={() => setQuick(undefined)}
        />
      )}
    </>
  );
}
