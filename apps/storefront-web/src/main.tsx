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
import { ProductCard } from "./product-card";
import { productPhoto, productPhotos } from "./product-photos";
import type { ProductPurchase } from "./product-purchase";
import { CheckoutPage } from "./checkout/checkout-page";
import { AuthPanel, useDemoAuth } from "./checkout/demo-auth";
import { initialDraft } from "./checkout/model";
import "./styles.css";
import "./checkout/styles.css";

const featuredProduct = products.find((product) => product.id === "cup")!;

function quantityForProduct(cart: CartLine[], productId: string) {
  return cart.reduce(
    (total, line) => total + (line.productId === productId ? line.quantity : 0),
    0,
  );
}

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
  const [onlyAvailable, setOnlyAvailable] = useState(false);
  const [storeFilter, setStoreFilter] = useState("all");
  const [maxPrice, setMaxPrice] = useState(4000000);
  const [colorFilter, setColorFilter] = useState("all");
  const [selection, setSelection] = useState<Record<string, string>>({});
  const [selectedStores, setSelectedStores] = useState<Record<string, string>>(
    {},
  );
  const [modalStoreId, setModalStoreId] = useState("");
  const [pictures, setPictures] = useState<Record<string, number>>({});
  const [cart, setCart] = useState<CartLine[]>([]);
  const [activeProduct, setActiveProduct] = useState(
    products.find((p) => p.id === initial.get("product")) ?? null,
  );
  const [cartOpen, setCartOpen] = useState(false);
  const [accountOpen, setAccountOpen] = useState(false);
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
  function selectedColor(product: Product) {
    if (!product.stock) return "";
    return selection[product.id] ?? product.colors[0] ?? "";
  }

  function selectedStore(product: Product) {
    const storeId =
      storeFilter === "all" ? selectedStores[product.id] : storeFilter;
    return (
      stores.find(
        (store) =>
          store.id === storeId && product.price + store.extra <= maxPrice,
      ) ?? stores[0]
    );
  }

  function pictureIndex(product: Product) {
    return (
      pictures[product.id] ??
      Math.max(
        0,
        productPhotos(product).findIndex(
          (photo) => photo.color === selectedColor(product),
        ),
      )
    );
  }

  function chooseColor(product: Product, color: string) {
    setSelection((current) => ({ ...current, [product.id]: color }));
    setPictures((current) => {
      const next = { ...current };
      delete next[product.id];
      return next;
    });
  }

  function chooseFilterColor(color: string) {
    setColorFilter(color);
    if (color === "all") return;
    const matching = products.filter((product) =>
      product.colors.includes(color),
    );
    setSelection((current) => ({
      ...current,
      ...Object.fromEntries(matching.map((product) => [product.id, color])),
    }));
    setPictures((current) =>
      Object.fromEntries(
        Object.entries(current).filter(
          ([id]) => !matching.some((product) => product.id === id),
        ),
      ),
    );
  }

  function purchase(
    product: Product,
    store = selectedStore(product),
  ): ProductPurchase {
    const color = selectedColor(product);
    const item = { productId: product.id, color, storeId: store.id };
    const key = lineKey(item);
    const quantity = cart.find((line) => lineKey(line) === key)?.quantity ?? 0;
    return {
      color,
      store,
      quantity,
      canAdd:
        Boolean(color) && quantityForProduct(cart, product.id) < product.stock,
      onColorChange: (value) => chooseColor(product, value),
      onQuantityChange: (change) => {
        setCart((current) => {
          const match = current.find((line) => lineKey(line) === key);
          if (
            !color ||
            (change > 0 &&
              quantityForProduct(current, product.id) >= product.stock)
          )
            return current;
          const nextQuantity = (match?.quantity ?? 0) + change;
          if (!match)
            return nextQuantity > 0
              ? [...current, { ...item, quantity: nextQuantity }]
              : current;
          return nextQuantity > 0
            ? current.map((line) =>
                line === match ? { ...line, quantity: nextQuantity } : line,
              )
            : current.filter((line) => line !== match);
        });
      },
    };
  }

  function openProduct(product: Product) {
    setModalStoreId("");
    setActiveProduct(product);
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
      const productId = current.find(
        (line) => lineKey(line) === key,
      )?.productId;
      const product = products.find((entry) => entry.id === productId);
      return product && quantityForProduct(merged, product.id) > product.stock
        ? current
        : merged;
    });
  }
  const filtered = products
    .filter(
      (p) =>
        (!onlyAvailable || p.stock > 0) &&
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
    setOnlyAvailable(false);
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
      <header className="header">
        <div className="header-inner wrap">
          <Brand />
          <label className="search">
            <Icon name="search" />
            <input
              type="search"
              aria-label="جست‌وجوی کالا"
              placeholder="جست‌وجوی کالا"
              value={query}
              onChange={(event) => setQuery(event.target.value)}
            />
          </label>
          <div className="header-actions">
            <button
              type="button"
              className="account icon-button"
              onClick={() => setAccountOpen(true)}
              aria-label={auth.customer ? "حساب نمونه من" : "ورود / ثبت‌نام"}
            >
              <Icon name="user" />
            </button>
            <button
              type="button"
              className="bag"
              aria-label={`سبد خرید، ${cartCount} کالا`}
              onClick={() => setCartOpen(true)}
            >
              <Icon name="bag" /> <span className="bag-label">سبد خرید</span>{" "}
              <b>{money(cartCount)}</b>
            </button>
          </div>
        </div>
      </header>
      <main className="wrap storefront-main">
        <section className="hero" aria-labelledby="hero-title">
          <div className="hero-copy">
            <p>خانه و زندگی</p>
            <h1 id="hero-title">
              ماگ سرامیکی
              <br />
              مینیمال
            </h1>
            <button
              className="primary"
              type="button"
              onClick={() => {
                reset();
                openProduct(featuredProduct);
              }}
            >
              انتخاب رنگ <Icon name="arrow" />
            </button>
          </div>
          <img
            src={productPhoto(featuredProduct, "سفید")}
            alt="ماگ سرامیکی سفید"
            fetchPriority="high"
          />
        </section>
        <nav className="chips categories" aria-label="دسته‌بندی کالا">
          {categories.map((entry) => (
            <button
              key={entry.id}
              type="button"
              data-category={entry.id}
              aria-pressed={category === entry.id}
              onClick={() => browse(entry.id)}
            >
              {entry.name}
            </button>
          ))}
        </nav>
        <section
          className="catalog-section"
          id="catalog"
          aria-labelledby="catalog-title"
        >
          <div className="section-heading">
            <h2 id="catalog-title">کالاها</h2>
            <small className="result-count" aria-live="polite">
              {money(filtered.length)} کالا
            </small>
          </div>
          <div
            className="catalog-controls"
            role="group"
            aria-label="فیلتر و مرتب‌سازی"
          >
            <div
              className="sort-row chips"
              role="group"
              aria-label="مرتب‌سازی و موجودی"
            >
              {[
                ["popular", "پرفروش‌ترین"],
                ["cheap", "کمترین قیمت"],
                ["expensive", "بیشترین قیمت"],
                ["new", "جدیدترین"],
              ].map(([value, label]) => (
                <button
                  key={value}
                  type="button"
                  data-sort={value}
                  aria-pressed={sort === value}
                  onClick={() => setSort(value)}
                >
                  {label}
                </button>
              ))}
              <button
                type="button"
                data-available
                aria-pressed={onlyAvailable}
                onClick={() => setOnlyAvailable(!onlyAvailable)}
              >
                فقط کالاهای موجود
              </button>
            </div>
            <div className="filter-row">
              <label>
                فروشگاه
                <select
                  aria-label="فروشگاه"
                  value={storeFilter}
                  onChange={(event) => setStoreFilter(event.target.value)}
                >
                  <option value="all">همه فروشگاه‌ها</option>
                  {stores.map((store) => (
                    <option key={store.id} value={store.id}>
                      {store.name}
                    </option>
                  ))}
                </select>
              </label>
              <label>
                رنگ / نوع
                <select
                  value={colorFilter}
                  onChange={(event) => chooseFilterColor(event.target.value)}
                >
                  <option value="all">همه گزینه‌ها</option>
                  {colors.map((color) => (
                    <option key={color}>{color}</option>
                  ))}
                </select>
              </label>
              <label>
                تا {money(maxPrice)} تومان
                <input
                  type="range"
                  aria-label="حداکثر قیمت"
                  min="0"
                  max="4000000"
                  step="50000"
                  value={maxPrice}
                  onChange={(event) => setMaxPrice(Number(event.target.value))}
                />
              </label>
              <button type="button" className="filter-reset" onClick={reset}>
                حذف فیلترها
              </button>
            </div>
          </div>
          {filtered.length ? (
            <div className="product-grid">
              {filtered.map((product) => (
                <ProductCard
                  key={product.id}
                  product={product}
                  purchase={purchase(product)}
                  image={pictureIndex(product)}
                  onImageChange={(index) =>
                    setPictures((current) => ({
                      ...current,
                      [product.id]: index,
                    }))
                  }
                  onDetails={() => openProduct(product)}
                />
              ))}
            </div>
          ) : (
            <div className="empty">
              <h3>هنوز چیزی پیدا نکردیم.</h3>
              <p>عبارت جست‌وجو یا فیلترها را تغییر بده.</p>
              <button className="primary" onClick={reset}>
                نمایش همه کالاها
              </button>
            </div>
          )}
        </section>
      </main>
      <footer className="storefront-footer">
        <div className="wrap footer-inner">
          <Brand />
          <p>نسخه بررسی با کالاها، قیمت‌ها و فرایند خرید آزمایشی</p>
        </div>
      </footer>
      {activeProduct && (
        <ProductModal
          key={activeProduct.id}
          product={activeProduct}
          purchase={purchase(
            activeProduct,
            stores.find((store) => store.id === modalStoreId) ??
              selectedStore(activeProduct),
          )}
          image={pictureIndex(activeProduct)}
          onImageChange={(index) =>
            setPictures((current) => ({
              ...current,
              [activeProduct.id]: index,
            }))
          }
          onStoreChange={(storeId) => {
            setModalStoreId(storeId);
            if (storeFilter === "all")
              setSelectedStores((current) => ({
                ...current,
                [activeProduct.id]: storeId,
              }));
          }}
          onClose={() => setActiveProduct(null)}
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
                        const totalQuantity = quantityForProduct(
                          cart,
                          product.id,
                        );
                        return (
                          <div className="cart-line" key={key}>
                            <img
                              src={productPhoto(product, line.color)}
                              alt=""
                            />
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
    </>
  );
}
createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <App />
  </StrictMode>,
);
