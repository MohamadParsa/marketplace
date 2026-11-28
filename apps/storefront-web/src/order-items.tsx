import { apiBase } from "./commerce";
import { CatalogImage } from "./catalog-image";
import { money, type Quote } from "../../../packages/commerce-client/types";
export function OrderItems({ lines }: { lines: Quote["lines"] }) {
  return (
    <div className="buy-order-items">
      {lines.map((l) => (
        <article className="buy-order-item" key={l.offer_id}>
          <div className="buy-order-item-heading">
            <CatalogImage
              className="buy-order-thumbnail"
              src={l.image_url ? apiBase + l.image_url : undefined}
              alt=""
              loading="lazy"
            />
            <div className="buy-order-product-name"><b>{l.name}</b>{l.variant_label&&<small>{l.variant_label}</small>}</div>
          </div>
          <dl className="buy-order-item-details">
            <div>
              <dt>تعداد</dt>
              <dd>{l.quantity.toLocaleString("fa-IR")}</dd>
            </div>
            <div>
              <dt>هر عدد</dt>
              <dd><bdi>{money(l.unit_price_toman)}</bdi></dd>
            </div>
            <div>
              <dt>جمع</dt>
              <dd><bdi>{money(l.total_toman)}</bdi></dd>
            </div>
          </dl>
        </article>
      ))}
    </div>
  );
}
