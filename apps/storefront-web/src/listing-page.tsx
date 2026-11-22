import { ChoiceGroup } from "./choice-group";
import { useEffect, useState } from "react";
import {
  Icon,
  ProductCard,
  SiteHeader,
  SiteFooter,
  type Card,
} from "./home-page";
import { navigate } from "./navigation";
const api = import.meta.env.VITE_API_BASE_URL ?? "";
type Nav = { id: string; name: string; url: string; parent_id?: string };
type Filter = {
  id: string;
  name: string;
  kind: string;
  unit?: string;
  options: string[];
};
type Listing = {
  category: (Nav & { description: string }) | null;
  breadcrumbs: Nav[];
  children: Nav[];
  navigation: Nav[];
  products: Card[];
  filters: Filter[];
  seo: { meta_title: string; meta_description: string; meta_keywords: string };
  title: string;
  query: string;
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
  sort: string;
  sort_options: string[];
};
const sortLabels: Record<string, string> = {
  default: "پیش‌فرض",
  newest: "جدیدترین",
  "price-asc": "ارزان‌ترین",
  "price-desc": "گران‌ترین",
};
function useDesktop() {
  const [wide, setWide] = useState(
    () => matchMedia("(min-width:900px)").matches,
  );
  useEffect(() => {
    const q = matchMedia("(min-width:900px)");
    const sync = () => setWide(q.matches);
    q.addEventListener("change", sync);
    return () => q.removeEventListener("change", sync);
  }, []);
  return wide;
}
function FiltersPanel({
  data,
  params,
  change,
}: {
  data: Listing;
  params: URLSearchParams;
  change: (edit: (p: URLSearchParams) => void) => void;
}) {
  const [min, setMin] = useState(
    params.has("min_price") ? String(Number(params.get("min_price")) / 10) : "",
  );
  const [max, setMax] = useState(
    params.has("max_price") ? String(Number(params.get("max_price")) / 10) : "",
  );
  const [priceError, setPriceError] = useState("");
  useEffect(() => {
    setMin(
      params.has("min_price")
        ? String(Number(params.get("min_price")) / 10)
        : "",
    );
    setMax(
      params.has("max_price")
        ? String(Number(params.get("max_price")) / 10)
        : "",
    );
    setPriceError("");
  }, [params.toString()]);
  return (
    <>
      <div className="plp-filter-heading">
        <h2>فیلترها</h2>
        <button
          onClick={() =>
            change((p) => {
              for (const k of [...p.keys()])
                if (
                  k.startsWith("attr.") ||
                  ["min_price", "max_price", "in_stock"].includes(k)
                )
                  p.delete(k);
            })
          }
        >
          پاک کردن
        </button>
      </div>
      <label className="plp-stock">
        <span>فقط کالاهای موجود</span>
        <input
          type="checkbox"
          checked={params.get("in_stock") === "1"}
          onChange={(e) =>
            change((p) =>
              e.target.checked ? p.set("in_stock", "1") : p.delete("in_stock"),
            )
          }
        />
      </label>
      <form
        className="plp-price-filter"
        onSubmit={(e) => {
          e.preventDefault();
          if (
            (min !== "" && (!Number.isFinite(+min) || +min < 0)) ||
            (max !== "" && (!Number.isFinite(+max) || +max < 0)) ||
            (min !== "" && max !== "" && +min > +max)
          ) {
            setPriceError("بازه قیمت معتبر وارد کنید.");
            return;
          }
          setPriceError("");
          change((p) => {
            for (const [k, v] of [
              ["min_price", min],
              ["max_price", max],
            ])
              v === ""
                ? p.delete(k)
                : p.set(k, String(Math.round(Number(v) * 10)));
          });
        }}
      >
        <h3>
          محدوده قیمت <small>تومان</small>
        </h3>
        <label>
          از
          <input
            type="number"
            min="0"
            step="0.1"
            aria-label="حداقل قیمت (تومان)"
            value={min}
            onChange={(e) => setMin(e.target.value)}
          />
        </label>
        <label>
          تا
          <input
            type="number"
            min="0"
            step="0.1"
            aria-label="حداکثر قیمت (تومان)"
            value={max}
            onChange={(e) => setMax(e.target.value)}
          />
        </label>
        {priceError && <p role="alert">{priceError}</p>}
        <button type="submit">اعمال محدوده قیمت</button>
      </form>
      {data.filters.map((f) => (
        <fieldset className="plp-attribute" key={f.id}>
          <legend>
            {f.name}
            {f.unit ? " (" + f.unit + ")" : ""}
          </legend>
          {!f.options.length ? (
            <p>گزینه‌ای ثبت نشده است.</p>
          ) : (
            f.options.map((value) => (
              <label key={value}>
                <input
                  type="checkbox"
                  checked={params.getAll("attr." + f.id).includes(value)}
                  onChange={(e) =>
                    change((p) => {
                      const key = "attr." + f.id,
                        values = p.getAll(key).filter((v) => v !== value);
                      p.delete(key);
                      for (const v of [
                        ...values,
                        ...(e.target.checked ? [value] : []),
                      ])
                        p.append(key, v);
                    })
                  }
                />
                <span>{value}</span>
              </label>
            ))
          )}
        </fieldset>
      ))}
    </>
  );
}
export function ListingPage({ url }: { url: string }) {
  const current = new URL(url, location.origin),
    params = new URLSearchParams(current.search),
    categoryPath = current.pathname.startsWith("/category/");
  let categoryId = "";
  try {
    if (categoryPath)
      categoryId = decodeURIComponent(
        current.pathname.slice("/category/".length),
      );
  } catch {
    categoryId = current.pathname.slice("/category/".length);
  }
  if (categoryPath && !categoryId) categoryId = "__missing__";
  const apiParams = new URLSearchParams(current.search);
  if (categoryPath) apiParams.set("category", categoryId);
  if (current.pathname === "/offers") apiParams.set("special", "1");
  const query = apiParams.toString();
  const [data, setData] = useState<Listing>();
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(0);
  const [attempt, setAttempt] = useState(0);
  const [filtersOpen, setFiltersOpen] = useState(false);
  const desktop = useDesktop();
  useEffect(() => {
    if (!filtersOpen || desktop) return;
    const previous = document.activeElement as HTMLElement | null;
    const overflow = document.body.style.overflow;
    document.body.style.overflow = "hidden";
    const trap = (e: KeyboardEvent) => {
      if (e.key === "Escape") {
        setFiltersOpen(false);
        return;
      }
      if (e.key !== "Tab") return;
      const dialog = document.querySelector<HTMLElement>(".plp-filter-sheet");
      const items = dialog?.querySelectorAll<HTMLElement>(
        "button:not([disabled]),input:not([disabled]),select:not([disabled]),a[href]",
      );
      if (!items?.length) return;
      const first = items[0],
        last = items[items.length - 1];
      if (e.shiftKey && document.activeElement === first) {
        e.preventDefault();
        last.focus();
      } else if (!e.shiftKey && document.activeElement === last) {
        e.preventDefault();
        first.focus();
      }
    };
    document.addEventListener("keydown", trap);
    return () => {
      document.body.style.overflow = overflow;
      document.removeEventListener("keydown", trap);
      previous?.focus();
    };
  }, [filtersOpen, desktop]);
  useEffect(() => {
    const c = new AbortController();
    let active = true;
    setLoading(true);
    setError(0);
    const restore = history.state?.cherrivaScroll ?? 0;
    void fetch(api + "/storefront/catalog?" + query, { signal: c.signal })
      .then(async (r) => {
        if (!r.ok) {
          if (active && !c.signal.aborted) setError(r.status);
          throw Error();
        }
        const body = (await r.json()) as Listing;
        if (active && !c.signal.aborted) {
          setData(body);
          requestAnimationFrame(() =>
            scrollTo({ top: restore, behavior: "instant" }),
          );
        }
      })
      .catch(() => {
        if (active && !c.signal.aborted) setError((e) => e || 503);
      })
      .finally(() => {
        if (active && !c.signal.aborted) setLoading(false);
      });
    return () => {
      active = false;
      c.abort();
    };
  }, [query, attempt]);
  useEffect(() => {
    if (error) {
      document.title = "صفحه در دسترس نیست | Cherriva";
      return;
    }
    if (!data) return;
    document.title = data.seo.meta_title;
    for (const [name, content] of [
      ["description", data.seo.meta_description],
      ["keywords", data.seo.meta_keywords],
    ]) {
      let el = document.head.querySelector<HTMLMetaElement>(
        'meta[name="' + name + '"]',
      );
      if (!el) {
        el = document.createElement("meta");
        el.name = name;
        document.head.append(el);
      }
      el.content = content;
    }
  }, [data, error]);
  const change = (edit: (p: URLSearchParams) => void) => {
    const next = new URLSearchParams(current.search);
    edit(next);
    next.delete("page");
    navigate(current.pathname + (next.size ? "?" + next.toString() : ""));
  };
  const categories = (data?.navigation ?? [])
    .filter((c) => !c.parent_id)
    .slice(0, 7)
    .map((c) => ({ key: c.id, title: c.name, url: c.url }));
  const filters = data ? (
    <FiltersPanel data={data} params={params} change={change} />
  ) : null;
  const heading =
    current.pathname === "/offers"
      ? "تخفیف‌های ویژه"
      : (data?.title ?? "فهرست کالاها");
  return (
    <div className="launch-home" dir="rtl">
      <a className="lh-skip" href="#main">
        رفتن به محتوا
      </a>
      <SiteHeader categories={categories} />
      <main className="lh-wrap plp-main" id="main">
        <nav className="plp-breadcrumb" aria-label="مسیر صفحه">
          <ol>
            <li>
              <a href="/">خانه</a>
            </li>
            {!error &&
              data?.breadcrumbs.map((c, i) => (
                <li key={c.id}>
                  {categoryPath && i === data.breadcrumbs.length - 1 ? (
                    <span aria-current="page">{c.name}</span>
                  ) : (
                    <a href={c.url}>{c.name}</a>
                  )}
                </li>
              ))}
            {!categoryPath && (
              <li>
                <span aria-current="page">
                  {current.pathname === "/offers"
                    ? "تخفیف‌های ویژه"
                    : "نتایج جست‌وجو"}
                </span>
              </li>
            )}
          </ol>
        </nav>
        {error ? (
          <section className="lh-error" role="alert">
            <h1>
              {error === 404
                ? "این دسته در دسترس نیست."
                : error === 400
                  ? "فیلترهای درخواست‌شده معتبر نیستند."
                  : "دریافت کالاها انجام نشد."}
            </h1>
            <p>
              {error === 400
                ? "فیلترها را پاک کنید و دوباره انتخاب کنید."
                : "ارتباط را بررسی کنید یا به دسته‌های دیگر بروید."}
            </p>
            {error === 400 ? (
              <a className="lh-primary" href={current.pathname}>
                پاک کردن فیلترها
              </a>
            ) : (
              <button
                className="lh-primary"
                onClick={() => setAttempt((n) => n + 1)}
              >
                تلاش دوباره
              </button>
            )}
            <a href="/">صفحه اصلی</a>
          </section>
        ) : (
          <>
            <div className="plp-title">
              <div>
                <span className="lh-eyebrow">CHERRIVA / همراه انتخاب تو</span>
                <h1>{heading}</h1>
                {data?.category && (
                  <small>انتخاب‌ها در دسته {data.category.name}</small>
                )}
              </div>
              <span>
                {data
                  ? new Intl.NumberFormat("fa-IR").format(data.total) + " کالا"
                  : "در حال دریافت…"}
              </span>
            </div>
            {!!data?.children.length && (
              <nav className="plp-children" aria-label="زیرشاخه‌های دسته">
                {data.children.map((c) => (
                  <a href={c.url} key={c.id}>
                    {c.name}
                    <Icon name="arrow" />
                  </a>
                ))}
              </nav>
            )}
            <div className="plp-layout">
              {desktop && data && (
                <aside className="plp-filters" aria-label="فیلتر کالاها">
                  {filters}
                </aside>
              )}
              <section className="plp-results" aria-label="فهرست محصولات">
                <div className="plp-toolbar">
                  {!desktop && (
                    <button
                      className="plp-filter-toggle"
                      onClick={() => setFiltersOpen(true)}
                      disabled={!data}
                      aria-expanded={filtersOpen}
                    >
                      فیلترها
                    </button>
                  )}
                  <ChoiceGroup
                    label="مرتب‌سازی کالاها"
                    value={params.get("sort") ?? "default"}
                    options={(
                      data?.sort_options ?? Object.keys(sortLabels)
                    ).map((v) => ({ value: v, label: sortLabels[v] }))}
                    onChange={(v) => change((p) => p.set("sort", v))}
                  />
                  {current.pathname === "/search" && data && (
                    <ChoiceGroup
                      label="دسته نتایج جست‌وجو"
                      value={params.get("category") ?? ""}
                      options={[
                        { value: "", label: "همه دسته‌ها" },
                        ...data.navigation.map((c) => ({
                          value: c.id,
                          label: c.name,
                        })),
                      ]}
                      onChange={(v) =>
                        change((p) => {
                          v ? p.set("category", v) : p.delete("category");
                          for (const k of [...p.keys()])
                            if (k.startsWith("attr.")) p.delete(k);
                        })
                      }
                    />
                  )}
                </div>
                {loading && (
                  <p role="status" className="plp-progress">
                    در حال دریافت کالاها…
                  </p>
                )}
                {!data && loading ? (
                  <div className="plp-grid" aria-hidden="true">
                    {Array.from({ length: 8 }, (_, i) => (
                      <div className="plp-skeleton" key={i} />
                    ))}
                  </div>
                ) : data?.products.length ? (
                  <div className="plp-grid" aria-busy={loading}>
                    {data.products.map((card) => (
                      <ProductCard card={card} key={card.id} />
                    ))}
                  </div>
                ) : !loading ? (
                  <div className="plp-empty">
                    <Icon name="search" />
                    <h2>کالایی با این شرایط پیدا نشد.</h2>
                    <p>
                      {data?.children.length
                        ? "می‌توانید زیرشاخه‌های بالا را بررسی کنید."
                        : "فیلترها یا عبارت جست‌وجو را تغییر دهید."}
                    </p>
                    <button onClick={() => navigate(current.pathname)}>
                      پاک کردن جست‌وجو و فیلترها
                    </button>
                  </div>
                ) : null}
                {data && data.total_pages > 1 && (
                  <nav className="plp-pagination" aria-label="صفحه‌بندی">
                    <button
                      disabled={data.page <= 1}
                      onClick={() => {
                        const p = new URLSearchParams(current.search);
                        p.set("page", String(data.page - 1));
                        navigate(current.pathname + "?" + p);
                      }}
                    >
                      قبلی
                    </button>
                    <span>
                      صفحه {new Intl.NumberFormat("fa-IR").format(data.page)} از{" "}
                      {new Intl.NumberFormat("fa-IR").format(data.total_pages)}
                    </span>
                    <button
                      disabled={data.page >= data.total_pages}
                      onClick={() => {
                        const p = new URLSearchParams(current.search);
                        p.set("page", String(data.page + 1));
                        navigate(current.pathname + "?" + p);
                      }}
                    >
                      بعدی
                    </button>
                  </nav>
                )}
              </section>
            </div>
            {data?.category?.description && (
              <section className="plp-description">
                <h2>درباره {data.category.name}</h2>
                <p>{data.category.description}</p>
              </section>
            )}
          </>
        )}
        {!desktop && filtersOpen && (
          <div
            className="plp-filter-overlay"
            onClick={() => setFiltersOpen(false)}
          >
            <section
              role="dialog"
              aria-modal="true"
              aria-label="فیلتر کالاها"
              className="plp-filter-sheet"
              onClick={(e) => e.stopPropagation()}
              onKeyDown={(e) => {
                if (e.key === "Escape") setFiltersOpen(false);
              }}
            >
              <button
                className="plp-filter-close"
                autoFocus
                onClick={() => setFiltersOpen(false)}
              >
                بستن فیلترها
              </button>
              {filters}
              <button
                className="lh-primary"
                onClick={() => setFiltersOpen(false)}
              >
                مشاهده نتایج
              </button>
            </section>
          </div>
        )}
      </main>
      <SiteFooter categories={categories} />
    </div>
  );
}
