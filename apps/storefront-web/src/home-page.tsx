import { LoadingPlaceholder } from "./busy-region";
import "./header-actions.css";
import { CategoryMenu } from "./category-menu";
import { useEffect, useRef, useState } from "react";
import { useCommerce } from "./commerce";
import { SearchBox } from "./search-box";
const api = import.meta.env.VITE_API_BASE_URL ?? "";
interface HomeCategory {
  id: string;
  name: string;
  url: string;
  image_url: string;
  count: number;
}
export interface Card {
  available?: boolean;
  features?: {
    attribute_id: string;
    name: string;
    value: string;
    unit?: string;
  }[];
  regular_price?: number;
  discount_percent?: number;
  id: string;
  name: string;
  category_name: string;
  url: string;
  image_url: string;
  price: number | null;
  offer_id?: string;
  variant_id?: string;
}
interface Rail {
  key: string;
  title: string;
  category_id: string;
  url: string;
  products: Card[];
}
interface Banner {
  id: string;
  config: {
    desktop_id: string;
    mobile_id: string;
    alt: string;
    link: string;
    placement?: string;
    catalog_category_id?: string;
  };
}
const keys = ["mobile", "earbuds", "power-bank", "smartwatch"];
const paths: Record<string, React.ReactNode> = {
  mobile: (
    <>
      <rect x="8" y="3" width="16" height="26" rx="3" />
      <path d="M13 7h6M14 25h4" />
    </>
  ),
  earbuds: (
    <>
      <path d="M5 15V11a11 11 0 0 1 22 0v4" />
      <rect x="3" y="13" width="6" height="12" rx="3" />
      <rect x="23" y="13" width="6" height="12" rx="3" />
    </>
  ),
  "power-bank": (
    <>
      <rect x="6" y="7" width="20" height="20" rx="4" />
      <path d="M12 3h8M17 12l-4 7h6l-4 5" />
    </>
  ),
  smartwatch: (
    <>
      <rect x="8" y="9" width="16" height="15" rx="4" />
      <path d="M12 9V3h8v6m-8 15v5h8v-5m-4-12v5l3 2" />
    </>
  ),
  search: (
    <>
      <circle cx="14" cy="14" r="8" />
      <path d="m20 20 7 7" />
    </>
  ),
  user: (
    <>
      <circle cx="16" cy="10" r="5" />
      <path d="M6 29v-3a10 10 0 0 1 20 0v3" />
    </>
  ),
  bag: (
    <>
      <path d="M6 10h20l2 19H4L6 10Z" />
      <path d="M11 11V7a5 5 0 0 1 10 0v4" />
    </>
  ),
  arrow: <path d="m20 6-10 10 10 10" />,
};
export function Icon({ name }: { name: string }) {
  return (
    <svg
      viewBox="0 0 32 32"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.6"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
    >
      {paths[name] ?? paths.mobile}
    </svg>
  );
}
function Price({ value }: { value: number }) {
  return (
    <span className="lh-price">
      <strong>
        {new Intl.NumberFormat("fa-IR", { maximumFractionDigits: 1 }).format(
          value / 10,
        )}
      </strong>
      <small>تومان</small>
    </span>
  );
}
export function ProductCard({ card }: { card: Card }) {
  const [failed, setFailed] = useState(false);
  useEffect(() => setFailed(false), [card.image_url]);
  return (
    <a
      className={`lh-product ${card.available === false ? "is-unavailable" : ""}`}
      href={card.url}
      data-product-id={card.id}
    >
      <div className="lh-product-image">
        {failed ? (
          <div className="lh-image-fallback">
            <Icon name="bag" />
            <span>تصویر در دسترس نیست</span>
          </div>
        ) : (
          <img
            src={api + card.image_url}
            alt={card.name}
            loading="lazy"
            decoding="async"
            onError={() => setFailed(true)}
          />
        )}
      </div>
      <span className="lh-category-name">{card.category_name}</span>
      <h3>{card.name}</h3>
      {card.features?.length ? (
        <ul className="lh-card-features">
          {card.features.map((f) => (
            <li key={f.attribute_id}>
              <span>{f.name}</span>
              <b dir="auto">
                {f.value}
                {f.unit ? " " + f.unit : ""}
              </b>
            </li>
          ))}
        </ul>
      ) : null}
      <div className="lh-product-price">
        {card.price !== null &&
        card.available !== false &&
        card.regular_price &&
        card.regular_price > card.price ? (
          <span className="lh-before-price">
            <del>
              {new Intl.NumberFormat("fa-IR", {
                maximumFractionDigits: 1,
              }).format(card.regular_price / 10)}
            </del>
            <b>
              {card.discount_percent
                ? `${new Intl.NumberFormat("fa-IR").format(card.discount_percent)}٪`
                : "ویژه"}
            </b>
          </span>
        ) : (
          <span>
            {card.available === false ? "فعلاً قابل خرید نیست" : "قیمت نقدی از"}
          </span>
        )}
        {card.price !== null && card.available !== false ? (
          <Price value={card.price} />
        ) : (
          <strong className="lh-unavailable">ناموجود</strong>
        )}
      </div>
    </a>
  );
}
function ProductRail({
  rail,
  index,
  special = false,
}: {
  rail: Rail;
  index: number;
  special?: boolean;
}) {
  const ref = useRef<HTMLDivElement>(null);
  return (
    <section
      className={`lh-rail ${special ? "lh-special" : ""}`}
      aria-labelledby={`rail-${rail.key}`}
      data-rail={rail.key}
    >
      <div className="lh-section-heading">
        <div>
          <span className="lh-section-index">
            {special ? "٪" : `۰${index + 1}`}
          </span>
          <h2 id={`rail-${rail.key}`}>{rail.title}</h2>
        </div>
        <div className="lh-rail-tools">
          {rail.url && (
            <a href={rail.url}>
              {special ? "مشاهده همه تخفیف‌ها" : "مشاهده همه"}{" "}
              <Icon name="arrow" />
            </a>
          )}
          <button
            aria-label={`پیمایش ${rail.title}`}
            onClick={() =>
              ref.current?.scrollBy({
                left: -280,
                behavior: matchMedia("(prefers-reduced-motion: reduce)").matches
                  ? "auto"
                  : "smooth",
              })
            }
            disabled={!rail.products.length}
          >
            <Icon name="arrow" />
          </button>
        </div>
      </div>
      {rail.products.length ? (
        <div
          className="lh-products"
          ref={ref}
          tabIndex={0}
          aria-label={`محصولات ${rail.title}`}
        >
          {rail.products.map((p) => (
            <ProductCard key={p.id} card={p} />
          ))}
        </div>
      ) : (
        <div className="lh-empty">
          <Icon name={rail.key} />
          <p>
            {special
              ? "در حال حاضر تخفیف فعالی وجود ندارد."
              : "در حال حاضر کالایی برای این دسته موجود نیست."}
          </p>
        </div>
      )}
    </section>
  );
}
function Hero({ banners, photos }: { banners: Banner[]; photos: Card[] }) {
  const [index, setIndex] = useState(0);
  const [broken, setBroken] = useState<string[]>([]);
  const valid = banners.filter(
    (b) => b.config.placement !== "promo" && !broken.includes(b.id),
  );
  const b = valid[index % Math.max(1, valid.length)];
  const visual = b ? (
    <picture>
      <source
        media="(max-width: 600px)"
        srcSet={
          api +
          `/content/banner-media/${encodeURIComponent(b.config.mobile_id)}`
        }
      />
      <img
        src={
          api +
          `/content/banner-media/${encodeURIComponent(b.config.desktop_id)}`
        }
        alt={b.config.alt}
        fetchPriority="high"
        onError={() => setBroken((x) => [...x, b.id])}
      />
    </picture>
  ) : null;
  return (
    <section
      className={`lh-hero ${b ? "with-media" : ""}`}
      aria-label="پیشنهادهای چریوا"
    >
      {b ? (
        b.config.link ? (
          <a className="lh-hero-media" href={b.config.link}>
            {visual}
          </a>
        ) : (
          <div className="lh-hero-media">{visual}</div>
        )
      ) : (
        <>
          <div className="lh-hero-copy">
            <span className="lh-eyebrow">CHERRIVA / انتخابی به سلیقه تو</span>
            <h1>
              دنیایی برای
              <br />
              انتخاب تو<span>.</span>
            </h1>
            <p>
              تازه‌های دیجیتال، همراه انتخاب‌های خوب تو.
              <br />
              اطلاعات روشن، قیمت نقدی مشخص.
            </p>
            <a href="#home-categories" className="lh-primary">
              دسته‌ها را ببین <Icon name="arrow" />
            </a>
          </div>
          <div
            className={`lh-hero-art ${photos.length ? "has-photos" : ""}`}
            aria-hidden="true"
          >
            {photos.length > 0 && (
              <div className="lh-hero-photos">
                {photos.slice(0, 2).map((p) => (
                  <img src={api + p.image_url} alt="" key={p.id} />
                ))}
              </div>
            )}
            <span className="lh-orbit one" />
            <span className="lh-orbit two" />
            <div className="lh-art-phone">
              <span />
              <i />
              <b>C</b>
            </div>
            <div className="lh-art-watch">
              <Icon name="smartwatch" />
            </div>
            <span className="lh-art-label">cherriva</span>
          </div>
        </>
      )}
      {valid.length > 1 && (
        <div className="lh-hero-dots">
          {valid.map((x, i) => (
            <button
              key={x.id}
              aria-label={`بنر ${i + 1}`}
              aria-pressed={index % valid.length === i}
              onClick={() => setIndex(i)}
            />
          ))}
        </div>
      )}
    </section>
  );
}
function PromoBanners({
  banners,
  rails,
}: {
  banners: Banner[];
  rails: Rail[];
}) {
  const [broken, setBroken] = useState<string[]>([]);
  const promos = banners
    .filter(
      (b) =>
        b.config.placement === "promo" &&
        !broken.includes(b.id) &&
        (!b.config.catalog_category_id ||
          rails.some(
            (r) =>
              r.category_id === b.config.catalog_category_id &&
              r.products.length > 0,
          )),
    )
    .slice(0, 2);
  return promos.length ? (
    <section className="lh-promos" aria-label="بنرهای منتخب">
      {promos.map((b) => {
        const rail = rails.find(
          (r) => r.category_id === b.config.catalog_category_id,
        );
        const product = rail?.products[0];
        if (b.config.catalog_category_id && rail && product)
          return (
            <a
              key={b.id}
              className={`lh-catalog-promo promo-${rail.key}`}
              data-category-id={rail.category_id}
              href={rail.url}
            >
              <div>
                <span className="lh-eyebrow">منتخب‌های چریوا</span>
                <h2>{b.config.alt || rail.title}</h2>
                <p>از میان مدل‌های موجود، همراه دلخواهت را انتخاب کن.</p>
                <b className="lh-promo-cta">
                  مدل‌های موجود را ببین <Icon name="arrow" />
                </b>
              </div>
              <img
                src={api + product.image_url}
                alt={product.name}
                loading="lazy"
                onError={() => setBroken((x) => [...x, b.id])}
              />
            </a>
          );
        const visual = (
          <picture>
            <source
              media="(max-width:600px)"
              srcSet={api + `/content/banner-media/${b.config.mobile_id}`}
            />
            <img
              src={api + `/content/banner-media/${b.config.desktop_id}`}
              alt={b.config.alt}
              loading="lazy"
              onError={() => setBroken((x) => [...x, b.id])}
            />
          </picture>
        );
        return b.config.link ? (
          <a href={b.config.link} key={b.id}>
            {visual}
          </a>
        ) : (
          <div key={b.id}>{visual}</div>
        );
      })}
    </section>
  ) : null;
}
export function HomePage() {
  useEffect(() => {
    document.title = "Cherriva | دنیایی برای انتخاب تو";
    for (const [name, content] of [
      [
        "description",
        "چریوا؛ انتخاب‌های دیجیتال با اطلاعات روشن و قیمت نقدی مشخص.",
      ],
      ["keywords", ""],
    ]) {
      const m = document.head.querySelector<HTMLMetaElement>(
        'meta[name="' + name + '"]',
      );
      if (m) m.content = content;
    }
  }, []);
  const [availableCategories, setAvailableCategories] = useState<
    HomeCategory[]
  >([]);
  const [special, setSpecial] = useState<Card[]>([]);
  const [specialUrl, setSpecialUrl] = useState("/offers");
  const [rails, setRails] = useState<Rail[]>([]);
  const [banners, setBanners] = useState<Banner[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(false);
  const [attempt, setAttempt] = useState(0);
  useEffect(() => {
    const controller = new AbortController();
    setLoading(true);
    setError(false);
    void fetch(api + "/storefront/home", { signal: controller.signal })
      .then(async (r) => {
        if (!r.ok) throw Error();
        const data = (await r.json()) as {
          rails: Rail[];
          categories: HomeCategory[];
          special_offers: Card[];
          special_offers_url: string;
        };
        if (
          data.rails.length !== 4 ||
          data.rails.some((r, i) => r.key !== keys[i])
        )
          throw Error();
        setRails(data.rails);
        setAvailableCategories(
          data.categories.filter((c: { count: number }) => c.count > 0),
        );
        setSpecial(data.special_offers);
        setSpecialUrl(data.special_offers_url);
      })
      .catch(() => {
        if (!controller.signal.aborted) setError(true);
      })
      .finally(() => {
        if (!controller.signal.aborted) setLoading(false);
      });
    void fetch(api + "/content/banners", { signal: controller.signal })
      .then((r) => (r.ok ? r.json() : []))
      .then((data: Banner[]) => setBanners(data))
      .catch(() => {});
    return () => controller.abort();
  }, [attempt]);
  useEffect(() => {
    if (loading || location.pathname !== "/" || !location.hash) return;
    const target = document.getElementById(location.hash.slice(1));
    target?.scrollIntoView();
  }, [loading]);
  const categories = rails.filter((r) => r.url);
  const isHome = location.pathname === "/";
  return (
    <div className="launch-home" dir="rtl">
      <a className="lh-skip" href="#main">
        رفتن به محتوا
      </a>
      <SiteHeader />
      <main id="main" className="lh-wrap">
        {!isHome ? (
          <section className="lh-route-unavailable">
            <h1>این صفحه در دسترس نیست.</h1>
            <a href="/" className="lh-primary">
              بازگشت به صفحه اصلی
            </a>
          </section>
        ) : (
          <>
            <Hero
              banners={banners}
              photos={rails.slice(0, 2).flatMap((r) => r.products.slice(0, 1))}
            />
            <section
              className="lh-category-block"
              id="home-categories"
              aria-label="دسته‌های دارای کالای موجود"
            >
              <div className="lh-category-heading">
                <h2>دسته‌های اصلی</h2>
                <span>انتخاب از دسته‌های موجود چریوا</span>
              </div>
              <div className="lh-categories lh-category-strip">
                {availableCategories.slice(0, 8).map((r) => (
                  <a key={r.id} href={r.url}>
                    <span>
                      {r.image_url ? (
                        <img src={api + r.image_url} alt="" />
                      ) : (
                        <Icon name="bag" />
                      )}
                    </span>
                    <b>{r.name}</b>
                    <small>
                      {new Intl.NumberFormat("fa-IR").format(r.count)} کالا
                    </small>
                  </a>
                ))}
              </div>
              {availableCategories.length > 8 && (
                <details className="lh-all-categories">
                  <summary>
                    مشاهده همه دسته‌ها (
                    {new Intl.NumberFormat("fa-IR").format(
                      availableCategories.length,
                    )}
                    )
                  </summary>
                  <div className="lh-categories lh-category-remainder">
                    {availableCategories.slice(8).map((r) => (
                      <a key={r.id} href={r.url}>
                        <span>
                          {r.image_url ? (
                            <img
                              src={api + r.image_url}
                              alt=""
                              loading="lazy"
                            />
                          ) : (
                            <Icon name="bag" />
                          )}
                        </span>
                        <b>{r.name}</b>
                        <small>
                          {new Intl.NumberFormat("fa-IR").format(r.count)} کالا
                        </small>
                      </a>
                    ))}
                  </div>
                </details>
              )}
            </section>
            {loading ? (
              <LoadingPlaceholder label="دریافت محصولات" />
            ) : error ? (
              <div className="lh-error" role="alert">
                <h2>دریافت محصولات انجام نشد.</h2>
                <p>ارتباط را بررسی کنید و دوباره تلاش کنید.</p>
                <button
                  className="lh-primary"
                  onClick={() => setAttempt((x) => x + 1)}
                >
                  تلاش دوباره
                </button>
              </div>
            ) : (
              <>
                <ProductRail
                  special
                  rail={{
                    key: "special",
                    title: "تخفیف‌های ویژه",
                    category_id: "",
                    url: specialUrl,
                    products: special,
                  }}
                  index={0}
                />
                {rails.map((rail, i) => (
                  <div key={rail.key}>
                    <ProductRail rail={rail} index={i} />
                    {i === 0 && (
                      <PromoBanners banners={banners} rails={rails} />
                    )}
                  </div>
                ))}
              </>
            )}
          </>
        )}
        <section className="lh-about" id="about">
          <span className="lh-eyebrow">یک انتخاب، هزار همراه</span>
          <h2>چریوا؛ همراه انتخاب‌های شما</h2>
          <p>
            در چریوا، موبایل، هندزفری، پاوربانک و ساعت هوشمند را کنار هم ببینید.
            مشخصات و قیمت‌ها را بررسی کنید و همراه دیجیتال مناسب خود را انتخاب
            کنید.
          </p>
        </section>
        <div className="lh-values">
          <div>
            <Icon name="search" />
            <b>اطلاعات روشن کالا</b>
          </div>
          <div>
            <Icon name="mobile" />
            <b>انتخاب‌های دیجیتال</b>
          </div>
          <div>
            <Icon name="bag" />
            <b>قیمت نقدی مشخص</b>
          </div>
        </div>
      </main>
      <SiteFooter categories={categories} />
    </div>
  );
}

type HeaderCategory = { key: string; title: string; url: string };
function HeaderActionIcon({ kind }: { kind: "account" | "cart" }) {
  return (
    <svg
      className="header-action-icon"
      viewBox="0 0 32 32"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.65"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
    >
      {kind === "account" ? (
        <>
          <circle cx="16" cy="10.5" r="4.5" />
          <path d="M7.5 26c.5-5.1 3.5-8 8.5-8s8 2.9 8.5 8" />
        </>
      ) : (
        <>
          <path d="M8 11h16l1.4 14.8a2 2 0 0 1-2 2.2H8.6a2 2 0 0 1-2-2.2L8 11Z" />
          <path d="M11.5 12V8a4.5 4.5 0 0 1 9 0v4" />
          <path d="M12.5 22h7" opacity=".35" />
        </>
      )}
    </svg>
  );
}
export function SiteHeader() {
  const { token, cart, logout } = useCommerce();
  const profileRef = useRef<HTMLDetailsElement>(null);
  useEffect(() => {
    const dismiss = (restore = false) => {
      if (profileRef.current?.open) {
        profileRef.current.open = false;
        if (restore)
          profileRef.current.querySelector<HTMLElement>("summary")?.focus();
      }
    };
    const outside = (e: PointerEvent | FocusEvent) => {
      if (!profileRef.current?.contains(e.target as Node)) dismiss();
    };
    const key = (e: KeyboardEvent) => {
      if (e.key === "Escape" && profileRef.current?.open) {
        e.preventDefault();
        dismiss(true);
      }
    };
    const route = () => dismiss();
    document.addEventListener("pointerdown", outside);
    document.addEventListener("focusin", outside);
    document.addEventListener("keydown", key);
    addEventListener("cherriva:navigate", route);
    addEventListener("cherriva:category-menu-open", route);
    addEventListener("popstate", route);
    return () => {
      document.removeEventListener("pointerdown", outside);
      document.removeEventListener("focusin", outside);
      document.removeEventListener("keydown", key);
      removeEventListener("cherriva:navigate", route);
      removeEventListener("cherriva:category-menu-open", route);
      removeEventListener("popstate", route);
    };
  }, []);
  const count = cart?.item_count ?? 0;
  const cartStatus = count
    ? `${count.toLocaleString("fa-IR")} کالا در سبد خرید`
    : "سبد خرید خالی است";
  return (
    <>
      {" "}
      <div className="lh-topline">چریوا؛ انتخابی ساده برای زندگی امروز</div>
      <header className="lh-header">
        <div className="lh-header-main lh-wrap">
          <a href="/" className="lh-logo" aria-label="چریوا؛ صفحه اصلی">
            <img
              src="/brand/cherriva-wordmark.svg"
              alt="Cherriva"
              width="177"
              height="48"
            />
          </a>
          <SearchBox />
          <div className="lh-header-actions">
            <details
              ref={profileRef}
              className="lh-profile-menu"
              data-authenticated={!!token}
            >
              <summary
                className="header-action"
                aria-label="حساب کاربری"
                aria-describedby="header-account-status"
                title={token ? "حساب کاربری" : "ورود / ثبت‌نام"}
              >
                <HeaderActionIcon kind="account" />
                {token && (
                  <span className="header-account-dot" aria-hidden="true" />
                )}
              </summary>
              <span
                id="header-account-status"
                className="header-action-description"
              >
                {token ? "وارد حساب شده‌اید" : "ورود یا ثبت‌نام در چریوا"}
              </span>
              <nav aria-label="منوی حساب">
                {token ? (
                  <>
                    <a href="/account">پروفایل</a>
                    <a href="/account/addresses">آدرس‌ها</a>
                    <a href="/account/orders">سفارش‌ها</a>
                    <a href="/wallet">کیف پول</a>
                    <button onClick={logout}>خروج</button>
                  </>
                ) : (
                  <a href="/login">ورود / ثبت‌نام</a>
                )}
              </nav>
            </details>
            <span className="header-action-divider" aria-hidden="true" />
            <a
              href="/cart"
              className="lh-cart-link header-action"
              aria-label="سبد خرید"
              aria-describedby="header-cart-status"
              aria-current={location.pathname === "/cart" ? "page" : undefined}
              title={cartStatus}
            >
              <HeaderActionIcon kind="cart" />
              {count > 0 && (
                <b className="lh-cart-count" aria-hidden="true">
                  {count > 99 ? "۹۹+" : count.toLocaleString("fa-IR")}
                </b>
              )}
            </a>
            <span id="header-cart-status" className="header-action-description">
              {cartStatus}
            </span>
          </div>
        </div>
        <nav className="lh-nav lh-wrap" aria-label="دسته‌های اصلی">
          <CategoryMenu />
        </nav>
      </header>
    </>
  );
}
export function SiteFooter({ categories }: { categories: HeaderCategory[] }) {
  return (
    <footer className="lh-footer">
      <div className="lh-wrap lh-footer-main">
        <div>
          <a href="/" className="lh-logo">
            <img
              src="/brand/cherriva-wordmark.svg"
              alt="Cherriva"
              width="160"
              height="44"
            />
          </a>
          <p>انتخابی ساده، برای همراه‌های هر روز.</p>
        </div>
        <div>
          <h2>دسته‌های چریوا</h2>
          {categories.map((r) => (
            <a href={r.url} key={r.key}>
              {r.title}
            </a>
          ))}
        </div>
        <div>
          <h2>درباره چریوا</h2>
          <a href="/info/about">آشنایی با چریوا</a>
          <a href="/info/contact">تماس با ما</a>
          <a href="/info/shipping">شرایط ارسال</a>
          <a href="/info/returns">شرایط بازگشت</a>
          <a href="/info/terms">قوانین خرید</a>
          <a href="/info/privacy">حریم خصوصی</a>
          <a href="/#home-categories">دسته‌بندی کالاها</a>
        </div>
      </div>
      <div className="lh-wrap lh-copyright">
        <span>تمام حقوق این وب‌سایت برای چریوا محفوظ است.</span>
        <span dir="ltr">cherriva.ir</span>
      </div>
    </footer>
  );
}
