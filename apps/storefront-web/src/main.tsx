import { StrictMode, useEffect, useState } from "react";
import { createRoot } from "react-dom/client";
import {
  categories,
  products,
  stores,
  money,
  lineKey,
  type Category,
  type CartLine,
  type Product,
} from "./catalog";
import { Brand, Icon, Modal } from "./components";
import { ProductModal } from "./product-modal";
import { CheckoutPage } from "./checkout/checkout-page";
import { AuthPanel, useDemoAuth } from "./checkout/demo-auth";
import { initialDraft } from "./checkout/model";
import "./styles.css";
import "./checkout/styles.css";

function App() {
  const auth = useDemoAuth();
  const [checkoutDraft, setCheckoutDraft] = useState(initialDraft);
  const [checkoutOpen, setCheckoutOpen] = useState(
    location.hash === "#checkout",
  );
  useEffect(() => {
    const sync = () => setCheckoutOpen(location.hash === "#checkout");
    window.addEventListener("popstate", sync);
    window.addEventListener("hashchange", sync);
    return () => {
      window.removeEventListener("popstate", sync);
      window.removeEventListener("hashchange", sync);
    };
  }, []);
  function startCheckout() {
    setCartOpen(false);
    setAccountOpen(false);
    setActiveProduct(null);
    setNotice("");
    history.pushState(
      null,
      "",
      `${location.pathname}${location.search}#checkout`,
    );
    setCheckoutOpen(true);
  }
  function exitCheckout() {
    history.replaceState(null, "", `${location.pathname}${location.search}`);
    setCheckoutOpen(false);
    setCartOpen(true);
  }
  const initial = new URLSearchParams(location.search);
  const [query, setQuery] = useState(initial.get("q") ?? "");
  const [category, setCategory] = useState<Category>(
    categories.some((c) => c.id === initial.get("category"))
      ? (initial.get("category") as Category)
      : "all",
  );
  const [sort, setSort] = useState("popular");
  const [storeFilter, setStoreFilter] = useState("all");
  const [maxPrice, setMaxPrice] = useState(4000000);
  const [colorFilter, setColorFilter] = useState("all");
  const [selection, setSelection] = useState<Record<string, string>>({});
  const [cart, setCart] = useState<CartLine[]>([]);
  const [activeProduct, setActiveProduct] = useState(
    products.find((p) => p.id === initial.get("product")) ?? null,
  );
  const [cartOpen, setCartOpen] = useState(false);
  const [accountOpen, setAccountOpen] = useState(false);
  const [notice, setNotice] = useState("");
  useEffect(() => {
    const params = new URLSearchParams();
    if (query) params.set("q", query);
    if (category !== "all") params.set("category", category);
    if (activeProduct) params.set("product", activeProduct.id);
    history.replaceState(
      null,
      "",
      `${location.pathname}${params.size ? "?" + params : ""}${location.hash}`,
    );
  }, [query, category, activeProduct]);
  useEffect(() => {
    if (!notice) return;
    const timer = setTimeout(() => setNotice(""), 4000);
    return () => clearTimeout(timer);
  }, [notice]);
  const selectedColor = (product: Product) =>
    selection[product.id] ??
    (product.colors.length === 1 ? product.colors[0] : "");
  function add(
    product: Product,
    storeId = storeFilter === "all" ? stores[0].id : storeFilter,
    quantity = 1,
  ) {
    const color = selectedColor(product);
    if (!color) {
      setNotice("ابتدا رنگ یا نوع کالا را انتخاب کنید.");
      return;
    }
    const total = cart
      .filter((line) => line.productId === product.id)
      .reduce((sum, line) => sum + line.quantity, 0);
    if (total + quantity > product.stock) {
      setNotice("تعداد درخواستی از موجودی نمونه بیشتر است.");
      return false;
    }
    const line = { productId: product.id, color, storeId, quantity };
    setCart((current) => {
      const match = current.find(
        (existing) => lineKey(existing) === lineKey(line),
      );
      return match
        ? current.map((existing) =>
            existing === match
              ? { ...existing, quantity: existing.quantity + quantity }
              : existing,
          )
        : [...current, line];
    });
    setNotice("کالا به سبد خرید اضافه شد.");
    return true;
  }
  function updateLine(key: string, patch: Partial<CartLine>) {
    setCart((current) => {
      const changed = current
        .map((l) => (lineKey(l) === key ? { ...l, ...patch } : l))
        .filter((l) => l.quantity > 0);
      const merged: CartLine[] = [];
      for (const line of changed) {
        const existing = merged.find((l) => lineKey(l) === lineKey(line));
        if (existing) existing.quantity += line.quantity;
        else merged.push({ ...line });
      }
      return merged;
    });
  }
  const filtered = products
    .filter(
      (p) =>
        (category === "all" || p.category === category) &&
        `${p.name} ${p.subtitle}`.includes(query.trim()) &&
        p.price + (storeFilter === "pearl" ? stores[1].extra : 0) <= maxPrice &&
        (colorFilter === "all" || p.colors.includes(colorFilter)),
    )
    .sort(
      (a, b) =>
        Number(b.stock > 0) - Number(a.stock > 0) ||
        (sort === "cheap"
          ? a.price - b.price
          : sort === "expensive"
            ? b.price - a.price
            : sort === "new"
              ? products.indexOf(b) - products.indexOf(a)
              : b.sales - a.sales),
    );
  const colors = [
    ...new Set(
      products
        .filter((p) => category === "all" || p.category === category)
        .flatMap((p) => p.colors),
    ),
  ];
  const cartCount = cart.reduce((sum, line) => sum + line.quantity, 0);
  const cartTotal = cart.reduce(
    (sum, line) =>
      sum +
      (products.find((p) => p.id === line.productId)!.price +
        stores.find((s) => s.id === line.storeId)!.extra) *
        line.quantity,
    0,
  );
  function browse(id: Category) {
    setCategory(id);
    setColorFilter("all");
    document.getElementById("catalog")?.scrollIntoView({ behavior: "smooth" });
  }
  function reset() {
    setQuery("");
    setCategory("all");
    setMaxPrice(4000000);
    setColorFilter("all");
    setStoreFilter("all");
    setSort("popular");
  }
  if (checkoutOpen)
    return (
      <CheckoutPage
        cart={cart}
        auth={auth}
        draft={checkoutDraft}
        onChange={setCheckoutDraft}
        onExit={exitCheckout}
      />
    );
  return (
    <>
      <div className="announcement">
        <span>چیزهای خوب، از یک انتخاب شروع می‌شوند.</span>
        <span className="review-label">
          نسخه بررسی · کالاها و قیمت‌ها آزمایشی‌اند
        </span>
      </div>
      <header className="header wrap">
        <Brand />
        <label className="search">
          <Icon name="search" />
          <input
            type="search"
            aria-label="جست‌وجوی کالا"
            placeholder="دنبال چه چیزی می‌گردی؟"
            value={query}
            onChange={(e) => {
              setQuery(e.target.value);
              document
                .getElementById("catalog")
                ?.scrollIntoView({ behavior: "smooth" });
            }}
          />
          <kbd>جست‌وجو</kbd>
        </label>
        <div className="header-actions">
          <button className="account" onClick={() => setAccountOpen(true)}>
            <Icon name="user" />
            <span>{auth.customer ? "حساب نمونه من" : "ورود / ثبت‌نام"}</span>
          </button>
          <i />
          <button
            className="bag icon-button"
            aria-label={`سبد خرید، ${cartCount} کالا`}
            onClick={() => setCartOpen(true)}
          >
            <Icon name="bag" />
            {cartCount > 0 && <b>{money(cartCount)}</b>}
          </button>
        </div>
      </header>
      <nav className="nav wrap" aria-label="دسته‌بندی اصلی">
        <button className="category-trigger" onClick={() => browse("all")}>
          <span>☰</span> دسته‌بندی کالاها
        </button>
        <div className="nav-links">
          {categories.slice(1).map((c) => (
            <button key={c.id} onClick={() => browse(c.id)}>
              {c.name}
            </button>
          ))}
        </div>
        <span className="nav-note">یک بازار، انتخاب‌های بسیار</span>
      </nav>
      <main className="wrap">
        <section className="hero">
          <div className="hero-copy">
            <span className="overline">
              <i /> به دنیای چریوا خوش آمدی
            </span>
            <h1>
              برای سلیقه‌ات،
              <br />
              <em>چیزی اینجاست.</em>
            </h1>
            <p>
              از یک فنجان آرامش تا همراه دیجیتال روزهایت؛
              <br />
              انتخاب‌های متفاوت را یک‌جا پیدا کن.
            </p>
            <button className="primary" onClick={() => browse("all")}>
              کشف همه محصولات <Icon name="arrow" />
            </button>
            <div className="hero-foot">
              <span>با سلیقه انتخاب کن</span>
              <span className="line" />
              <span>از فروشگاه‌های مختلف</span>
            </div>
          </div>
          <div className="hero-image">
            <img
              src="/images/coffee.svg"
              alt="فنجان قهوه روی میز، همراه یک لحظه آرام"
              fetchPriority="high"
            />
            <div className="image-overlay" />
            <span className="image-caption">
              THE LITTLE THINGS
              <br />
              <b>لحظه‌های کوچک، حس‌های بزرگ</b>
            </span>
            <div className="image-sticker">
              <span>انتخاب امروز</span>
              <strong>یک مکث دل‌نشین</strong>
              <button
                onClick={() => browse("tea")}
                aria-label="دیدن چای و قهوه"
              >
                <Icon name="arrow" />
              </button>
            </div>
          </div>
        </section>
        <section className="category-strip" aria-label="انتخاب دسته">
          {categories.map((c) => (
            <button
              key={c.id}
              className={category === c.id ? "category active" : "category"}
              onClick={() => browse(c.id)}
            >
              <span className={`category-icon ${c.id}`}>{c.icon}</span>
              <span>{c.name}</span>
              <small>←</small>
            </button>
          ))}
        </section>
        <section className="catalog-section" id="catalog">
          <div className="section-heading">
            <div>
              <span className="overline">به انتخاب خودت</span>
              <h2>
                {category === "all"
                  ? "کشف کن، مقایسه کن، انتخاب کن."
                  : categories.find((c) => c.id === category)!.name}
              </h2>
            </div>
            <span className="result-count">
              {money(filtered.length)} کالای نمونه
            </span>
          </div>
          <div className="catalog-layout">
            <aside className="filters">
              <div className="filter-heading">
                <h3>فیلترها</h3>
                <button onClick={reset}>حذف فیلترها</button>
              </div>
              <label>
                فروشگاه
                <select
                  aria-label="فروشگاه"
                  value={storeFilter}
                  onChange={(e) => setStoreFilter(e.target.value)}
                >
                  <option value="all">همه فروشگاه‌ها</option>
                  {stores.map((s) => (
                    <option key={s.id} value={s.id}>
                      {s.name}
                    </option>
                  ))}
                </select>
              </label>
              <label>
                محدوده قیمت
                <span className="range-value">تا {money(maxPrice)} تومان</span>
                <input
                  aria-label="حداکثر قیمت"
                  type="range"
                  min="0"
                  max="4000000"
                  step="50000"
                  value={maxPrice}
                  onChange={(e) => setMaxPrice(Number(e.target.value))}
                />
                <span className="range-ends">
                  <span>۰</span>
                  <span>۴ میلیون تومان</span>
                </span>
              </label>
              <label>
                رنگ / نوع
                <select
                  value={colorFilter}
                  onChange={(e) => setColorFilter(e.target.value)}
                >
                  <option value="all">همه گزینه‌ها</option>
                  {colors.map((c) => (
                    <option key={c}>{c}</option>
                  ))}
                </select>
              </label>
              <div className="filter-note">
                <span>◈</span>
                <strong>انتخاب آگاهانه‌تر</strong>
                <p>
                  در جزئیات هر کالا، قیمت و زمان آماده‌سازی فروشگاه‌ها را مقایسه
                  کن.
                </p>
              </div>
            </aside>
            <div className="catalog-results">
              <div className="sort-bar">
                <span>مرتب‌سازی:</span>
                <label className="sr-only" htmlFor="sort">
                  مرتب‌سازی
                </label>
                <select
                  id="sort"
                  value={sort}
                  onChange={(e) => setSort(e.target.value)}
                >
                  <option value="popular">پرفروش‌ترین</option>
                  <option value="cheap">ارزان‌ترین</option>
                  <option value="expensive">گران‌ترین</option>
                  <option value="new">جدیدترین</option>
                </select>
                <span className="sample-note">موجودی و فروش نمونه</span>
              </div>
              {filtered.length ? (
                <div className="product-grid">
                  {filtered.map((product) => (
                    <article className="product-card" key={product.id}>
                      <button
                        className="product-picture"
                        onClick={() => setActiveProduct(product)}
                        aria-label={`مشاهده ${product.name}`}
                      >
                        <img
                          src={`/images/${product.image}.svg`}
                          alt={product.name}
                          loading="lazy"
                        />
                        {!product.stock ? (
                          <span className="badge unavailable">ناموجود</span>
                        ) : (
                          <span className="badge">انتخاب چریوا</span>
                        )}
                        <span className="quick-view">مشاهده محصول ↖</span>
                      </button>
                      <div className="product-info">
                        <small>
                          {
                            categories.find((c) => c.id === product.category)!
                              .name
                          }
                        </small>
                        <button
                          className="product-title"
                          onClick={() => setActiveProduct(product)}
                        >
                          {product.name}
                        </button>
                        <p>{product.subtitle}</p>
                        <div className="card-choice">
                          <label
                            className="sr-only"
                            htmlFor={`color-${product.id}`}
                          >
                            رنگ یا نوع {product.name}
                          </label>
                          <select
                            id={`color-${product.id}`}
                            value={selectedColor(product)}
                            onChange={(e) =>
                              setSelection({
                                ...selection,
                                [product.id]: e.target.value,
                              })
                            }
                          >
                            <option value="" disabled>
                              انتخاب رنگ / نوع
                            </option>
                            {product.colors.map((c) => (
                              <option key={c}>{c}</option>
                            ))}
                          </select>
                          <span>
                            {stores.find((s) => s.id === storeFilter)?.name ??
                              stores[0].name}
                          </span>
                        </div>
                        <div className="product-bottom">
                          <div>
                            <strong>
                              {money(
                                product.price +
                                  (storeFilter === "pearl"
                                    ? stores[1].extra
                                    : 0),
                              )}
                            </strong>{" "}
                            <small>تومان</small>
                          </div>
                          <button
                            className="add-button"
                            aria-label={`افزودن ${product.name} به سبد`}
                            disabled={!product.stock || !selectedColor(product)}
                            onClick={() => add(product)}
                          >
                            <span>+</span>
                          </button>
                        </div>
                      </div>
                    </article>
                  ))}
                </div>
              ) : (
                <div className="empty">
                  <span>⌕</span>
                  <h3>هنوز چیزی پیدا نکردیم.</h3>
                  <p>عبارت جست‌وجو یا فیلترها را تغییر بده.</p>
                  <button className="primary" onClick={reset}>
                    نمایش همه کالاها
                  </button>
                </div>
              )}
            </div>
          </div>
        </section>
        <section className="collection-banner">
          <div>
            <span className="overline">دنیای دیجیتال چریوا</span>
            <h2>همراه کوچکِ روزهای بزرگ.</h2>
            <p>انتخاب‌هایی برای کار، موسیقی و زندگی روزمره.</p>
            <button onClick={() => browse("digital")}>
              کشف کالاهای دیجیتال <Icon name="arrow" />
            </button>
          </div>
          <span className="banner-art" aria-hidden="true">
            ⌘
          </span>
          <div className="banner-caption">
            CHERRIVA
            <br />
            <strong>DIGI</strong>
          </div>
        </section>
        <section className="benefits">
          <div>
            <span>◈</span>
            <h3>چند فروشگاه، یک تجربه</h3>
            <p>مقایسه پیشنهادها در کنار هم</p>
          </div>
          <div>
            <span>⌕</span>
            <h3>انتخاب با جزئیات روشن</h3>
            <p>قیمت، موجودی و آماده‌سازی مشخص</p>
          </div>
          <div>
            <span>♡</span>
            <h3>برای سلیقه‌های متفاوت</h3>
            <p>از خانه تا دنیای دیجیتال</p>
          </div>
        </section>
      </main>
      <footer>
        <div className="wrap footer-inner">
          <div>
            <Brand />
            <p>چیزهای خوب را کنار هم پیدا کنیم.</p>
          </div>
          <div>
            <strong>دنیای چریوا</strong>
            {categories.slice(1).map((c) => (
              <button key={c.id} onClick={() => browse(c.id)}>
                {c.name}
              </button>
            ))}
          </div>
          <div className="footer-review">
            <strong>درباره این نسخه</strong>
            <p>
              پیش‌نمایش سایت مشتری با داده ساختگی.
              <br />
              ثبت سفارش، ورود و پرداخت واقعی فعال نیست.
            </p>
          </div>
        </div>
        <div className="wrap copyright">
          <span>Cherriva · نسخه بررسی رابط کاربری</span>
          <span>ساخته‌شده برای انتخاب‌های خوب</span>
        </div>
      </footer>
      {activeProduct && (
        <ProductModal
          product={activeProduct}
          color={selectedColor(activeProduct)}
          setColor={(color) =>
            setSelection({ ...selection, [activeProduct.id]: color })
          }
          onClose={() => setActiveProduct(null)}
          onAdd={add}
        />
      )}
      {cartOpen && (
        <Modal
          title={`سبد خرید (${money(cartCount)})`}
          wide
          onClose={() => setCartOpen(false)}
        >
          {cart.length ? (
            <>
              <p className="dialog-note">
                سبد آزمایشی است و با بارگذاری مجدد صفحه پاک می‌شود. قیمت‌ها شامل
                ارسال نیستند.
              </p>
              {stores
                .filter((s) => cart.some((l) => l.storeId === s.id))
                .map((store) => (
                  <section className="cart-store" key={store.id}>
                    <h3>{store.name}</h3>
                    <p className="muted">
                      آماده‌سازی نمونه: {store.preparation}
                    </p>
                    {cart
                      .filter((l) => l.storeId === store.id)
                      .map((line) => {
                        const product = products.find(
                          (p) => p.id === line.productId,
                        )!;
                        const key = lineKey(line);
                        const totalQuantity = cart
                          .filter((l) => l.productId === product.id)
                          .reduce((sum, l) => sum + l.quantity, 0);
                        return (
                          <div className="cart-line" key={key}>
                            <img src={`/images/${product.image}.svg`} alt="" />
                            <div className="cart-line-info">
                              <strong>{product.name}</strong>
                              <label>
                                نوع
                                <select
                                  aria-label={`نوع ${product.name} در سبد`}
                                  value={line.color}
                                  onChange={(e) =>
                                    updateLine(key, { color: e.target.value })
                                  }
                                >
                                  {product.colors.map((c) => (
                                    <option key={c}>{c}</option>
                                  ))}
                                </select>
                              </label>
                              <label>
                                فروشگاه
                                <select
                                  aria-label={`فروشگاه ${product.name} در سبد`}
                                  value={line.storeId}
                                  onChange={(e) =>
                                    updateLine(key, { storeId: e.target.value })
                                  }
                                >
                                  {stores.map((s) => (
                                    <option key={s.id} value={s.id}>
                                      {s.name} ·{" "}
                                      {money(product.price + s.extra)} تومان
                                    </option>
                                  ))}
                                </select>
                              </label>
                              <div className="quantity">
                                <button
                                  aria-label={`کاهش تعداد ${product.name}`}
                                  onClick={() =>
                                    updateLine(key, {
                                      quantity: line.quantity - 1,
                                    })
                                  }
                                >
                                  −
                                </button>
                                <span>{money(line.quantity)}</span>
                                <button
                                  aria-label={`افزایش تعداد ${product.name}`}
                                  disabled={totalQuantity >= product.stock}
                                  onClick={() =>
                                    updateLine(key, {
                                      quantity: line.quantity + 1,
                                    })
                                  }
                                >
                                  +
                                </button>
                                <button
                                  className="remove"
                                  onClick={() =>
                                    updateLine(key, { quantity: 0 })
                                  }
                                >
                                  حذف
                                </button>
                              </div>
                            </div>
                            <strong className="line-price">
                              {money(
                                (product.price + store.extra) * line.quantity,
                              )}{" "}
                              <small>تومان</small>
                            </strong>
                          </div>
                        );
                      })}
                  </section>
                ))}
              <div className="cart-total">
                <span>جمع اقلام نمونه</span>
                <strong>{money(cartTotal)} تومان</strong>
              </div>
              <p className="dialog-note">
                در ادامه، ورود نمونه، اطلاعات و آدرس، روش ارسال و صفحه پرداخت را
                بررسی کن. هیچ پرداخت واقعی انجام نمی‌شود.
              </p>
              <button className="primary full" onClick={startCheckout}>
                ادامه خرید و انتخاب آدرس <Icon name="arrow" />
              </button>
            </>
          ) : (
            <div className="empty">
              <Icon name="bag" />
              <h3>سبدت منتظر انتخاب توست.</h3>
              <p>از میان کالاهای نمونه، چیزی انتخاب کن.</p>
              <button className="primary" onClick={() => setCartOpen(false)}>
                مشاهده کالاها
              </button>
            </div>
          )}
        </Modal>
      )}
      {accountOpen && (
        <Modal title="حساب کاربری نمونه" onClose={() => setAccountOpen(false)}>
          <AuthPanel auth={auth} onDone={() => setAccountOpen(false)} />
        </Modal>
      )}
      <div className={`toast ${notice ? "visible" : ""}`} role="status">
        {notice}
        <button
          onClick={() => {
            setActiveProduct(null);
            setCartOpen(true);
            setNotice("");
          }}
          hidden={!notice.includes("اضافه شد")}
        >
          مشاهده سبد ←
        </button>
      </div>
    </>
  );
}
createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <App />
  </StrictMode>,
);
