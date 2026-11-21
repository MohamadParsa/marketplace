import { useCommerce, apiBase } from "./commerce";
import { Modal } from "./components";
import { money as tomanMoney } from "../../../packages/commerce-client/types";
import { useEffect, useRef, useState } from "react";
import { SiteHeader, SiteFooter, Icon } from "./home-page";
import { navigate } from "./navigation";
const api = import.meta.env.VITE_API_BASE_URL ?? "";
type Nav = { id: string; name: string; url: string };
type Feature = {
  attribute_id: string;
  name: string;
  value: string;
  unit?: string;
};
type Content = {
  short_description: string;
  full_description: string;
  advantages: string[];
  disadvantages: string[];
  reviews_enabled: boolean;
  qa_enabled: boolean;
  shipping: {
    weight?: number;
    length?: number;
    width?: number;
    height?: number;
    weight_unit: string;
    dimension_unit: string;
  };
  seo: { meta_title: string; meta_description: string; meta_keywords: string };
  noindex: boolean;
  nofollow: boolean;
};
type Offer = {
  id: string;
  variant_id: string;
  store_id: string;
  store_name: string;
  lead_days: number;
  available: boolean;
  price: number | null;
  regular_price?: number;
  version: number;
};
type Detail = {
  id: string;
  name: string;
  english_name: string;
  brand: { id: string; name: string } | null;
  category: Nav;
  breadcrumbs: Nav[];
  navigation: Nav[];
  content: Content;
  gallery: string[];
  axes: { id: string; name: string; options: string[] }[];
  variants: {
    id: string;
    values: { attribute_id: string; value: string }[];
    available: boolean;
    image_url?: string;
  }[];
  offers: Offer[];
  selected_variant_id: string;
  selected_offer_id: string;
  features: Feature[];
  specifications: Feature[];
  interactions: { reviews: boolean; qa: boolean; notify: boolean };
  checkout_enabled: boolean;
  robots: string;
};
const money = (n: number) =>
  new Intl.NumberFormat("fa-IR", { maximumFractionDigits: 1 }).format(n / 10);
function Photo({
  src,
  name,
  className,
}: {
  src?: string;
  name: string;
  className?: string;
}) {
  const [failed, setFailed] = useState(false);
  useEffect(() => setFailed(false), [src]);
  return src && !failed ? (
    <img
      className={className}
      src={api + src}
      alt={name}
      onError={() => setFailed(true)}
    />
  ) : (
    <span className="pdp-photo-fallback">تصویر در دسترس نیست</span>
  );
}
function Lightbox({
  src,
  name,
  close,
}: {
  src: string;
  name: string;
  close: () => void;
}) {
  const ref = useRef<HTMLDialogElement>(null);
  useEffect(() => {
    const previous = document.activeElement as HTMLElement;
    ref.current?.showModal();
    const overflow = document.body.style.overflow;
    document.body.style.overflow = "hidden";
    return () => {
      document.body.style.overflow = overflow;
      previous?.focus();
    };
  }, []);
  return (
    <dialog
      className="pdp-lightbox"
      ref={ref}
      onClose={close}
      onClick={(e) => {
        if (e.target === e.currentTarget) close();
      }}
    >
      <button autoFocus aria-label="بستن تصویر بزرگ" onClick={close}>
        بستن ×
      </button>
      <Photo src={src} name={name} />
    </dialog>
  );
}
export function ProductPage({ url }: { url: string }) {
  const commerce = useCommerce();
  const [adding, setAdding] = useState(false),
    [addError, setAddError] = useState(""),
    [added, setAdded] = useState(false);
  async function addToCart() {
    if (!data || !commerce.cart || adding) return;
    setAdding(true);
    setAddError("");
    try {
      await commerce.mutate({
        action: "add",
        offer_id: data.selected_offer_id,
        quantity,
        version: commerce.cart.version,
      });
      setAdded(true);
    } catch (e) {
      setAddError((e as Error).message);
      await commerce.refreshCart().catch(() => {});
    } finally {
      setAdding(false);
    }
  }
  const [data, setData] = useState<Detail>(),
    [loading, setLoading] = useState(true),
    [status, setStatus] = useState(0),
    [attempt, setAttempt] = useState(0),
    [photo, setPhoto] = useState(0),
    [zoom, setZoom] = useState(false),
    [quantity, setQuantity] = useState(1);
  const initialPosition = useRef(true);
  useEffect(() => {
    const abort = new AbortController();
    setLoading(true);
    setStatus(0);
    document.title = "محصول | Cherriva";
    for (const name of ["description", "keywords"]) {
      const m = document.head.querySelector<HTMLMetaElement>(
        'meta[name="' + name + '"]',
      );
      if (m) m.content = "";
    }

    const current = new URL(url, location.origin);
    let id = "";
    try {
      id = decodeURIComponent(current.pathname.slice("/product/".length));
    } catch {
      id = "__invalid__";
    }
    void fetch(
      api + "/storefront/products/" + encodeURIComponent(id) + current.search,
      { signal: abort.signal },
    )
      .then(async (r) => {
        if (!r.ok) {
          setStatus(r.status);
          throw Error();
        }
        const d = (await r.json()) as Detail;
        if (abort.signal.aborted) return;
        setData(d);
        if (initialPosition.current) {
          initialPosition.current = false;
          const top = history.state?.cherrivaScroll ?? 0;
          requestAnimationFrame(() => scrollTo({ top, behavior: "instant" }));
        }

        setPhoto(0);
        setQuantity(1);
        document.title = d.content.seo.meta_title;
        for (const [name, content] of Object.entries({
          ...d.content.seo,
          robots: d.robots,
        })) {
          const key = name.replace("meta_", "");
          let m = document.head.querySelector<HTMLMetaElement>(
            `meta[name="${key === "title" ? "__skip" : key}"]`,
          );
          if (key === "title") continue;
          if (!m) {
            m = document.createElement("meta");
            m.name = key;
            document.head.append(m);
          }
          m.content = content;
        }
      })
      .catch(() => {
        if (!abort.signal.aborted) setStatus((s) => s || 503);
      })
      .finally(() => {
        if (!abort.signal.aborted) setLoading(false);
      });
    return () => abort.abort();
  }, [url, attempt]);
  const choose = (variant: string, offer?: string) => {
    const u = new URL(url, location.origin);
    u.search = "";
    if (variant) u.searchParams.set("variant", variant);
    if (offer) u.searchParams.set("offer", offer);
    navigate(u.pathname + u.search);
  };
  const selected = data?.variants.find(
      (v) => v.id === data.selected_variant_id,
    ),
    offer = data?.offers.find((o) => o.id === data.selected_offer_id);
  const available = !!offer?.available;
  const canPurchase = available && !loading && status === 0;
  const [selectionError, setSelectionError] = useState(""),
    [checking, setChecking] = useState(false);
  useEffect(() => {
    setSelectionError("");
    if (!data || !offer?.available) {
      setChecking(false);
      return;
    }
    const abort = new AbortController();
    setChecking(true);
    void fetch(
      api +
        "/storefront/products/" +
        encodeURIComponent(data.id) +
        "/selection",
      {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          variant_id: data.selected_variant_id,
          offer_id: data.selected_offer_id,
          quantity,
        }),
        signal: abort.signal,
      },
    )
      .then((r) => {
        if (!r.ok)
          throw Error(
            r.status === 409
              ? "تعداد انتخابی موجود نیست؛ تعداد را کاهش دهید یا اطلاعات محصول را تازه کنید."
              : "بررسی انتخاب انجام نشد؛ دوباره تلاش کنید.",
          );
      })
      .catch((e) => {
        if (!abort.signal.aborted) setSelectionError(e.message);
      })
      .finally(() => {
        if (!abort.signal.aborted) setChecking(false);
      });
    return () => abort.abort();
  }, [
    data?.id,
    data?.selected_variant_id,
    data?.selected_offer_id,
    offer?.available,
    quantity,
    attempt,
  ]);
  const content = data?.content;
  const selectAxis = (axis: string, value: string) => {
    if (!data || !selected) return;
    const next = {
      ...Object.fromEntries(
        selected.values.map((v) => [v.attribute_id, v.value]),
      ),
      [axis]: value,
    };
    const exact = data.variants.find((v) =>
      data.axes.every((a) =>
        v.values.some((x) => x.attribute_id === a.id && x.value === next[a.id]),
      ),
    );
    const candidates = data.variants.filter((v) =>
      v.values.some((x) => x.attribute_id === axis && x.value === value),
    );
    const target =
      exact ?? candidates.find((v) => v.available) ?? candidates[0];
    if (target) choose(target.id);
  };
  return (
    <div className="launch-home product-page" dir="rtl">
      <a className="lh-skip" href="#main">
        رفتن به محتوا
      </a>
      <SiteHeader
        categories={(data?.navigation ?? []).map((c) => ({
          key: c.id,
          title: c.name,
          url: c.url,
        }))}
      />
      <main className="lh-wrap" id="main">
        {loading && !data ? (
          <section className="pdp-state" role="status">
            در حال دریافت محصول…
          </section>
        ) : status ? (
          <section className="pdp-state" role="alert">
            <h1>
              {status === 404
                ? "این محصول در دسترس نیست."
                : status === 400
                  ? "انتخاب محصول معتبر نیست."
                  : "دریافت اطلاعات محصول انجام نشد."}
            </h1>
            {status !== 404 && (
              <button
                className="lh-primary"
                onClick={() =>
                  status === 400
                    ? navigate(new URL(url, location.origin).pathname)
                    : setAttempt((n) => n + 1)
                }
              >
                {status === 400 ? "بازنشانی انتخاب" : "تلاش مجدد"}
              </button>
            )}
            <a href="/">بازگشت به خانه</a>
          </section>
        ) : data && content ? (
          <>
            <nav className="plp-breadcrumb" aria-label="مسیر صفحه">
              <ol>
                <li>
                  <a href="/">خانه</a>
                </li>
                {data.breadcrumbs.map((c) => (
                  <li key={c.id}>
                    <a href={c.url}>{c.name}</a>
                  </li>
                ))}
                {data.brand && <li>{data.brand.name}</li>}
                <li aria-current="page">{data.name}</li>
              </ol>
            </nav>
            {loading && (
              <div className="pdp-progress" role="status">
                در حال به‌روزرسانی انتخاب…
              </div>
            )}
            <section className="pdp-overview" aria-busy={loading}>
              <div className="pdp-gallery">
                <button
                  className="pdp-large-photo"
                  disabled={!data.gallery.length}
                  onClick={() => setZoom(true)}
                  aria-label="نمایش تصویر بزرگ"
                >
                  <Photo src={data.gallery[photo]} name={data.name} />
                </button>
                <div className="pdp-thumbnails" aria-label="تصاویر محصول">
                  {data.gallery.map((src, i) => (
                    <button
                      key={src}
                      aria-label={`تصویر ${i + 1}`}
                      aria-pressed={photo === i}
                      onClick={() => setPhoto(i)}
                    >
                      <Photo src={src} name={`تصویر ${i + 1} ${data.name}`} />
                    </button>
                  ))}
                </div>
              </div>
              <div className="pdp-information">
                <span className="pdp-brand">
                  {data.brand?.name ?? data.category.name}
                </span>
                <h1>{data.name}</h1>
                {data.english_name && (
                  <p className="pdp-english" dir="auto">
                    {data.english_name}
                  </p>
                )}
                {content.short_description && (
                  <p className="pdp-short">{content.short_description}</p>
                )}
                <div className="pdp-axes">
                  {data.axes.map((a) => (
                    <fieldset key={a.id}>
                      <legend>
                        {a.name}:{" "}
                        <b>
                          {selected?.values.find((v) => v.attribute_id === a.id)
                            ?.value ?? "انتخاب نشده"}
                        </b>
                      </legend>
                      {a.options.map((value) => (
                        <button
                          key={value}
                          disabled={
                            loading ||
                            !data.variants.some((v) =>
                              v.values.some(
                                (x) =>
                                  x.attribute_id === a.id && x.value === value,
                              ),
                            )
                          }
                          aria-pressed={
                            selected?.values.some(
                              (v) =>
                                v.attribute_id === a.id && v.value === value,
                            ) ?? false
                          }
                          onClick={() => selectAxis(a.id, value)}
                        >
                          {value}
                        </button>
                      ))}
                    </fieldset>
                  ))}
                </div>
                {!!data.features.length && (
                  <div className="pdp-key-features">
                    <h2>ویژگی‌های کلیدی</h2>
                    <dl>
                      {data.features.map((f) => (
                        <div key={f.attribute_id}>
                          <dt>{f.name}</dt>
                          <dd dir="auto">
                            {f.value} {f.unit}
                          </dd>
                        </div>
                      ))}
                    </dl>
                    <a href="#specifications">مشاهده همه مشخصات</a>
                  </div>
                )}
              </div>
              <aside className="pdp-buybox" aria-label="انتخاب فروشنده و خرید">
                <span className="pdp-buybox-tag">
                  {available ? "قیمت نقدی این انتخاب" : "این انتخاب موجود نیست"}
                </span>
                <h2>{offer?.store_name ?? "فروشنده فعالی ندارد"}</h2>
                {offer && (
                  <p>
                    <Icon name="bag" />
                    وقفه تأمین فروشنده:{" "}
                    {new Intl.NumberFormat("fa-IR").format(offer.lead_days)} روز
                    کاری
                  </p>
                )}
                <small>
                  هزینه و بازه تحویل پس از انتخاب نشانی و روش ارسال مشخص می‌شود.
                </small>
                {available && offer?.price != null ? (
                  <div className="pdp-price">
                    {offer.regular_price &&
                    offer.regular_price > offer.price ? (
                      <div>
                        <del>{money(offer.regular_price)}</del>
                        <b className="pdp-discount">
                          {Math.floor(
                            ((offer.regular_price - offer.price) * 100) /
                              offer.regular_price,
                          )}
                          ٪
                        </b>
                      </div>
                    ) : null}
                    <strong>
                      {money(offer.price)} <small>تومان</small>
                    </strong>
                    <span>قیمت نهایی نقدی، بدون هزینه ارسال</span>
                  </div>
                ) : (
                  <p className="pdp-unavailable">ناموجود</p>
                )}
                {canPurchase && (
                  <div className="pdp-quantity" aria-label="تعداد انتخابی">
                    <button
                      aria-label="افزایش تعداد"
                      disabled={quantity >= 99}
                      onClick={() => setQuantity((q) => q + 1)}
                    >
                      +
                    </button>
                    <output aria-live="polite">
                      {new Intl.NumberFormat("fa-IR").format(quantity)}
                    </output>
                    <button
                      className="pdp-decrease"
                      aria-label="کاهش تعداد"
                      disabled={quantity <= 1}
                      onClick={() => setQuantity((q) => q - 1)}
                    >
                      −
                    </button>
                  </div>
                )}
                <div className="pdp-selection-status" aria-live="polite">
                  {checking ? (
                    "در حال بررسی تعداد…"
                  ) : selectionError ? (
                    <p role="alert">{selectionError}</p>
                  ) : null}
                </div>
                <button
                  className="pdp-add"
                  data-product-id={data.id}
                  data-variant-id={data.selected_variant_id}
                  data-offer-id={data.selected_offer_id}
                  data-quantity={quantity}
                  disabled={
                    !canPurchase ||
                    checking ||
                    !!selectionError ||
                    adding ||
                    !commerce.cart
                  }
                  onClick={() => void addToCart()}
                  aria-describedby="pdp-purchase-note"
                >
                  افزودن به سبد خرید
                </button>
                {addError && <p role="alert">{addError}</p>}
                <p id="pdp-purchase-note" className="pdp-purchase-note">
                  {available
                    ? "قیمت نقدی بدون هزینه ارسال؛ روش پرداخت در ادامه مشخص می‌شود."
                    : "در حال حاضر امکان خرید این انتخاب وجود ندارد."}
                </p>
              </aside>
            </section>
            <section
              className="pdp-sellers"
              aria-labelledby="pdp-sellers-title"
            >
              <h2 id="pdp-sellers-title">
                فروشنده‌های این انتخاب{" "}
                <small>({data.offers.length.toLocaleString("fa-IR")})</small>
              </h2>
              {!data.offers.length ? (
                <p>برای این تنوع عرضه فعالی ثبت نشده است.</p>
              ) : (
                data.offers.map((o) => (
                  <article
                    className={o.id === offer?.id ? "selected" : ""}
                    key={o.id}
                  >
                    <div>
                      <h3>{o.store_name}</h3>
                      <span>
                        {o.id === offer?.id
                          ? "فروشنده انتخاب‌شده"
                          : o.available
                            ? "عرضه موجود"
                            : "ناموجود"}
                      </span>
                    </div>
                    <p>
                      تأمین در {o.lead_days.toLocaleString("fa-IR")} روز کاری
                    </p>
                    <b>
                      {o.available && o.price != null
                        ? money(o.price) + " تومان"
                        : "ناموجود"}
                    </b>
                    <button
                      disabled={loading || o.id === offer?.id || !o.available}
                      onClick={() => choose(data.selected_variant_id, o.id)}
                    >
                      {o.id === offer?.id ? "انتخاب‌شده" : "انتخاب فروشنده"}
                    </button>
                  </article>
                ))
              )}
            </section>
            <nav className="pdp-tabs" aria-label="بخش‌های محصول">
              <a href="#introduction">معرفی محصول</a>
              <a href="#specifications">مشخصات فنی</a>
              <a href="#reviews">نظرات کاربران</a>
              <a href="#questions">پرسش و پاسخ</a>
            </nav>
            <section className="pdp-section" id="introduction">
              <h2>معرفی محصول</h2>
              {content.full_description ? (
                <div
                  className="pdp-full-description"
                  dangerouslySetInnerHTML={{ __html: content.full_description }}
                />
              ) : (
                <p>توضیح تکمیلی هنوز ثبت نشده است.</p>
              )}
              <div className="pdp-pros-cons">
                {content.advantages.length > 0 && (
                  <div>
                    <h3>مزایا</h3>
                    <ul>
                      {content.advantages.map((s, i) => (
                        <li key={i}>{s}</li>
                      ))}
                    </ul>
                  </div>
                )}
                {content.disadvantages.length > 0 && (
                  <div>
                    <h3>معایب</h3>
                    <ul>
                      {content.disadvantages.map((s, i) => (
                        <li key={i}>{s}</li>
                      ))}
                    </ul>
                  </div>
                )}
              </div>
            </section>
            <section className="pdp-section" id="specifications">
              <h2>مشخصات فنی</h2>
              <dl className="pdp-specifications">
                {data.brand && (
                  <div>
                    <dt>برند</dt>
                    <dd>{data.brand.name}</dd>
                  </div>
                )}
                {data.specifications.map((f) => (
                  <div key={f.attribute_id}>
                    <dt>{f.name}</dt>
                    <dd dir="auto">
                      {f.value} {f.unit}
                    </dd>
                  </div>
                ))}
                {(["weight", "length", "width", "height"] as const)
                  .filter((k) => content.shipping[k] != null)
                  .map((k) => (
                    <div key={k}>
                      <dt>
                        {
                          {
                            weight: "وزن حمل",
                            length: "طول",
                            width: "عرض",
                            height: "ارتفاع",
                          }[k]
                        }
                      </dt>
                      <dd>
                        {content.shipping[k]?.toLocaleString("fa-IR")}{" "}
                        {k === "weight"
                          ? "گرم"
                          : content.shipping.dimension_unit === "mm"
                            ? "میلی‌متر"
                            : "سانتی‌متر"}
                      </dd>
                    </div>
                  ))}
              </dl>
            </section>
            <section className="pdp-section" id="reviews">
              <h2>نظرات کاربران</h2>
              <p>
                {content.reviews_enabled
                  ? "امکان ثبت و نمایش نظر هنوز فعال نشده است."
                  : "ثبت نظر برای این محصول غیرفعال است."}
              </p>
            </section>
            <section className="pdp-section" id="questions">
              <h2>پرسش و پاسخ</h2>
              <p>
                {content.qa_enabled
                  ? "امکان ثبت و نمایش پرسش هنوز فعال نشده است."
                  : "ثبت پرسش برای این محصول غیرفعال است."}
              </p>
            </section>
            {zoom && data.gallery[photo] && (
              <Lightbox
                src={data.gallery[photo]}
                name={data.name}
                close={() => setZoom(false)}
              />
            )}
          </>
        ) : null}
      </main>
      <SiteFooter
        categories={(data?.navigation ?? []).map((c) => ({
          key: c.id,
          title: c.name,
          url: c.url,
        }))}
      />
      {added && (
        <Modal
          title="به سبد خرید اضافه شد"
          className="buy-add-popup"
          onClose={() => setAdded(false)}
        >
          <ul>
            {commerce.cart?.items.map((i) => (
              <li key={i.offer_id}>
                <img src={apiBase + i.image_url} alt="" />
                <div>
                  <b>{i.name}</b>
                  <small>
                    {i.variant_label} · {i.store_name} · تعداد {i.quantity}
                  </small>
                  <strong>{tomanMoney(i.total_toman)}</strong>
                </div>
              </li>
            ))}
          </ul>
          <div className="buy-row">
            <a className="buy-primary" href="/cart">
              مشاهده سبد خرید
            </a>
            <button className="buy-secondary" onClick={() => setAdded(false)}>
              ادامه خرید
            </button>
          </div>
        </Modal>
      )}
    </div>
  );
}
