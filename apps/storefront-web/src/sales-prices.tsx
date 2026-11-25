import { useEffect, useState } from "react";
import { apiBase } from "./commerce";
type Price = {
  method_id: string;
  name: string;
  price_toman: string;
  markup_toman: string;
  markup_bps: number;
};
export function SalesPrices({
  product,
  offer,
  variant,
}: {
  product: string;
  offer: string;
  variant: string;
}) {
  const [rows, setRows] = useState<Price[]>([]);
  useEffect(() => {
    setRows([]);
    const abort = new AbortController();
    void fetch(
      apiBase +
        `/storefront/products/${encodeURIComponent(product)}/sales?` +
        new URLSearchParams({ offer, variant }),
      { signal: abort.signal, cache: "no-store" },
    )
      .then((r) => (r.ok ? r.json() : []))
      .then((next: Price[]) => { if (!abort.signal.aborted) setRows(next); })
      .catch(() => {});
    return () => abort.abort();
  }, [product, offer, variant]);
  const priced = rows.filter((p) => p.markup_bps > 0);
  if (!priced.length) return null;
  return (
    <div className="pdp-sales-prices" aria-label="قیمت بر اساس روش فروش">
      {priced.map((p) => (
        <p key={p.method_id}>
          <b>
            {p.name}: {Number(p.price_toman).toLocaleString("fa-IR")} تومان
          </b>
          <small>
            مارکاپ {p.markup_bps / 100}٪؛ مبلغ نهایی با روش پرداخت و تخفیف در
            خرید محاسبه می‌شود.
          </small>
        </p>
      ))}
    </div>
  );
}
