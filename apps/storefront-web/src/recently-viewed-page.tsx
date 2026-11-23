import { useEffect, useState } from "react";
import { call, useCommerce } from "./commerce";
import { CommerceShell, useAction } from "./commerce-pages";
import { ProductCard, type Card } from "./home-page";
import { BusyRegion, LoadingPlaceholder } from "./busy-region";
import { Modal } from "./components";
import { HistoryIcon } from "./recently-viewed";
import "./recently-viewed.css";
type Result = {
  items: { product_id: string; product: Card | null; viewed_at: string }[];
  total: number;
  page: number;
  page_size: number;
  max_items: number;
  retention_days: number;
};
const endpoint = "/storefront/account/recently-viewed";
const viewedDate = (value: string) =>
  new Intl.DateTimeFormat("fa-IR", {
    dateStyle: "medium",
    timeZone: "Asia/Tehran",
  }).format(new Date(value));
export function RecentlyViewedPage() {
  const { token } = useCommerce();
  const action = useAction();
  const [page, setPage] = useState(1),
    [reload, setReload] = useState(0),
    [confirm, setConfirm] = useState(false);
  const [result, setResult] = useState<{
      session: string;
      data: Result;
    } | null>(null),
    [loading, setLoading] = useState(true),
    [error, setError] = useState("");
  useEffect(() => {
    setPage(1);
    setResult(null);
    setError("");
    setConfirm(false);
  }, [token]);
  useEffect(() => {
    let cancelled = false;
    if (!token) {
      setLoading(false);
      return;
    }
    setLoading(true);
    setError("");
    void call<Result>(endpoint + "?page=" + page, token)
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
  }, [token, page, reload]);
  const data = result?.session === token ? result.data : null,
    pages = data ? Math.ceil(data.total / data.page_size) : 0;
  const remove = (id?: string) =>
    action.run(async () => {
      await call(
        endpoint + (id ? "/" + encodeURIComponent(id) : ""),
        token,
        undefined,
        "DELETE",
      );
      setConfirm(false);
      setReload((n) => n + 1);
    }, id ?? "clear-history");
  return (
    <CommerceShell title="بازدیدهای اخیر">
      {!token ? (
        <section className="wishlist-empty">
          <HistoryIcon />
          <h2>بازدیدهای اخیر، ویژه حساب شما</h2>
          <p>برای ثبت و دیدن کالاهایی که بررسی کرده‌اید، وارد حساب خود شوید.</p>
          <a
            className="lh-primary"
            href="/login?return=%2Faccount%2Frecently-viewed"
          >
            ورود به حساب
          </a>
        </section>
      ) : (
        <>
          <div className="recent-toolbar">
            <p className="wishlist-caption">
              {data
                ? `${data.max_items.toLocaleString("fa-IR")} کالای آخر از ${data.retention_days.toLocaleString("fa-IR")} روز اخیر؛ بازدید دوباره، کالا را به ابتدای فهرست می‌آورد.`
                : "کالاهایی که اخیراً دیده‌اید، به ترتیب آخرین بازدید."}
            </p>
            {!!data?.total && (
              <button
                className="buy-secondary"
                disabled={loading || action.busy}
                onClick={() => {
                  action.setError("");
                  setConfirm(true);
                }}
              >
                پاک کردن همه بازدیدها
              </button>
            )}
          </div>
          {error && (
            <div className="buy-alert" role="alert">
              {error}{" "}
              <button onClick={() => setReload((n) => n + 1)}>تلاش مجدد</button>
            </div>
          )}
          {action.error && !confirm && (
            <p className="buy-alert" role="alert">
              {action.error}
            </p>
          )}
          {loading && !data ? (
            <LoadingPlaceholder label="دریافت بازدیدهای اخیر" />
          ) : (
            <BusyRegion
              busy={loading || action.busy}
              label="به‌روزرسانی بازدیدهای اخیر"
            >
              {data && data.total > 0 ? (
                <>
                  <p className="wishlist-count">
                    {data.total.toLocaleString("fa-IR")} کالای دیده‌شده
                  </p>
                  <div className="wishlist-grid">
                    {data.items.map((item) => (
                      <article className="wishlist-item" key={item.product_id}>
                        {item.product ? (
                          <ProductCard card={item.product} />
                        ) : (
                          <div className="wishlist-unavailable">
                            <HistoryIcon />
                            <h2>این کالا دیگر در دسترس نیست</h2>
                            <p>می‌توانید آن را از سابقه خود حذف کنید.</p>
                          </div>
                        )}
                        <p className="recent-viewed-date">
                          آخرین مشاهده:{" "}
                          <time dateTime={item.viewed_at}>
                            {viewedDate(item.viewed_at)}
                          </time>
                        </p>
                        <button
                          type="button"
                          className="wishlist-remove"
                          disabled={action.busy || loading}
                          aria-label={`حذف از سابقه: ${item.product?.name ?? "کالای غیرقابل‌دسترسی"}`}
                          onClick={() => void remove(item.product_id)}
                        >
                          <span aria-hidden="true">×</span> حذف از سابقه
                        </button>
                      </article>
                    ))}
                  </div>
                </>
              ) : (
                data && (
                  <section className="wishlist-empty">
                    <HistoryIcon />
                    <h2>سابقه بازدید شما خالی است</h2>
                    <p>
                      بعد از باز کردن صفحه یک محصول، آن را اینجا پیدا می‌کنید.
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
                  aria-label="صفحه‌بندی بازدیدهای اخیر"
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
          {confirm && (
            <Modal
              title="پاک کردن سابقه بازدید"
              busy={action.busy}
              onClose={() => {
                if (!action.busy) setConfirm(false);
              }}
            >
              <p>
                همه بازدیدهای ثبت‌شده این حساب پاک شوند؟ بازدیدهای بعدی دوباره
                ثبت خواهند شد.
              </p>
              {action.error && (
                <p className="buy-alert" role="alert">
                  {action.error}
                </p>
              )}
              <div className="recent-clear-actions">
                <button
                  className="buy-secondary"
                  autoFocus
                  onClick={() => setConfirm(false)}
                  disabled={action.busy}
                >
                  انصراف
                </button>
                <button
                  className="buy-primary"
                  disabled={action.busy}
                  onClick={() => void remove()}
                >
                  بله، پاک شود
                </button>
              </div>
            </Modal>
          )}
        </>
      )}
    </CommerceShell>
  );
}
