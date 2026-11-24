import { useEffect, useState } from "react";
import { call, useCommerce } from "./commerce";
import { useWishlist, HeartIcon } from "./wishlist";
import { CommerceShell } from "./commerce-pages";
import { ProductCard, type Card } from "./home-page";
import { BusyRegion, LoadingPlaceholder } from "./busy-region";
type Result = {
  items: { product_id: string; product: Card | null }[];
  total: number;
  page: number;
  page_size: number;
};
export function WishlistPage() {
  const { token } = useCommerce();
  const { setSaved, revision } = useWishlist();
  const [page, setPage] = useState(1),
    [retry, setRetry] = useState(0);
  const [result, setResult] = useState<{
    session: string;
    data: Result;
  } | null>(null);
  const [loading, setLoading] = useState(true),
    [error, setError] = useState(""),
    [removing, setRemoving] = useState("");
  useEffect(() => {
    setPage(1);
    setResult(null);
    setError("");
  }, [token]);
  useEffect(() => {
    let cancelled = false;
    if (!token) {
      setLoading(false);
      return;
    }
    setLoading(true);
    setError("");
    void call<Result>("/storefront/account/wishlist?page=" + page, token)
      .then((data) => {
        if (cancelled) return;
        const last = Math.max(1, Math.ceil(data.total / data.page_size));
        if (page > last) {
          setPage(last);
          return;
        }
        setResult({ session: token, data });
      })
      .catch((e) => {
        if (!cancelled) setError((e as Error).message);
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, [token, page, revision, retry]);
  const data = result?.session === token ? result.data : null;
  const pages = data ? Math.ceil(data.total / data.page_size) : 0;
  return (
    <CommerceShell title="علاقه‌مندی‌های من">
      {!token ? (
        <section className="wishlist-empty">
          <HeartIcon />
          <h2>انتخاب‌های دوست‌داشتنی شما، یک‌جا</h2>
          <p>برای دیدن و ذخیره کالاهای مورد علاقه، وارد حساب خود شوید.</p>
          <a className="lh-primary" href="/login?return=%2Faccount%2Fwishlist">
            ورود به حساب
          </a>
        </section>
      ) : (
        <>
          <p className="wishlist-caption">
            کالاهایی که برای بعد ذخیره کرده‌اید؛ قیمت و موجودی هنگام خرید دوباره
            بررسی می‌شود.
          </p>
          {error && (
            <div className="buy-alert" role="alert">
              {error}{" "}
              <button type="button" onClick={() => setRetry((n) => n + 1)}>
                تلاش مجدد
              </button>
            </div>
          )}
          {loading && !data ? (
            <LoadingPlaceholder label="دریافت علاقه‌مندی‌ها" />
          ) : (
            <BusyRegion
              busy={loading || !!removing}
              label="به‌روزرسانی علاقه‌مندی‌ها"
            >
              {data && data.total > 0 ? (
                <>
                  <p className="wishlist-count">
                    {data.total.toLocaleString("fa-IR")} کالای ذخیره‌شده
                  </p>
                  <div className="wishlist-grid">
                    {data.items.map((item) => (
                      <article className="wishlist-item" key={item.product_id}>
                        {item.product ? (
                          <ProductCard card={item.product} />
                        ) : (
                          <div className="wishlist-unavailable">
                            <HeartIcon />
                            <h2>این کالا دیگر در دسترس نیست</h2>
                            <p>می‌توانید آن را از فهرست خود حذف کنید.</p>
                          </div>
                        )}
                        <button
                          type="button"
                          className="wishlist-remove"
                          disabled={!!removing || loading}
                          aria-label={`حذف از علاقه‌مندی‌ها: ${item.product?.name ?? "کالای غیرقابل‌دسترسی"}`}
                          onClick={() => {
                            setRemoving(item.product_id);
                            setError("");
                            void setSaved(item.product_id, false)
                              .catch((e) => setError((e as Error).message))
                              .finally(() => setRemoving(""));
                          }}
                        >
                          <HeartIcon filled /> حذف از علاقه‌مندی‌ها
                        </button>
                      </article>
                    ))}
                  </div>
                </>
              ) : (
                data && (
                  <section className="wishlist-empty">
                    <HeartIcon />
                    <h2>هنوز کالایی ذخیره نکرده‌اید</h2>
                    <p>
                      با زدن قلب کنار تصاویر محصول، آن را به این فهرست اضافه
                      کنید.
                    </p>
                    <a className="lh-primary" href="/">
                      دیدن کالاها
                    </a>
                  </section>
                )
              )}
              {pages > 1 && (
                <nav
                  className="wishlist-pagination"
                  aria-label="صفحه‌بندی علاقه‌مندی‌ها"
                >
                  <button
                    disabled={page <= 1}
                    onClick={() => setPage((n) => n - 1)}
                  >
                    قبلی
                  </button>
                  <span>
                    صفحه {page.toLocaleString("fa-IR")} از{" "}
                    {pages.toLocaleString("fa-IR")}
                  </span>
                  <button
                    disabled={page >= pages}
                    onClick={() => setPage((n) => n + 1)}
                  >
                    بعدی
                  </button>
                </nav>
              )}
            </BusyRegion>
          )}
        </>
      )}
    </CommerceShell>
  );
}
