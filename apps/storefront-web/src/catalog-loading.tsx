import { LoadingPlaceholder } from "./busy-region";
import "./catalog-loading.css";
export function CardSkeleton() {
  return (
    <div className="lh-product sf-card-skeleton" aria-hidden="true">
      <div className="lh-product-image sf-skeleton" />
      <i className="sf-skeleton sf-skeleton-short" />
      <i className="sf-skeleton" />
      <i className="sf-skeleton" />
      <i className="sf-skeleton sf-skeleton-price" />
    </div>
  );
}
export function HomeCatalogSkeleton() {
  return (
    <LoadingPlaceholder label="دریافت محصولات" className="sf-home-skeleton">
      <section className="lh-category-block" aria-hidden="true">
        <div className="lh-category-heading">
          <i className="sf-skeleton sf-skeleton-heading" />
        </div>
        <div className="sf-category-skeletons" aria-hidden="true">
          {Array.from({ length: 8 }, (_, i) => (
            <span key={i}>
              <i className="sf-skeleton" />
              <b className="sf-skeleton" />
            </span>
          ))}
        </div>
        <i className="sf-skeleton sf-category-toggle-skeleton" />
      </section>
      {Array.from({ length: 5 }, (_, i) => (
        <section className="lh-rail" aria-hidden="true" key={i}>
          <div className="lh-section-heading">
            <i className="sf-skeleton sf-skeleton-heading" />
          </div>
          <div className="lh-products">
            {Array.from({ length: 6 }, (_, n) => (
              <CardSkeleton key={n} />
            ))}
          </div>
        </section>
      ))}
    </LoadingPlaceholder>
  );
}
export function ProductSkeleton() {
  return (
    <LoadingPlaceholder label="دریافت محصول" className="sf-product-skeleton">
      <div className="sf-skeleton sf-skeleton-heading" aria-hidden="true" />
      <div className="sf-product-skeleton-layout" aria-hidden="true">
        <div className="pdp-overview">
          <div className="pdp-gallery">
            <div className="pdp-large-photo sf-skeleton" />
            <div className="pdp-thumbnails">
              {[0, 1, 2].map((i) => (
                <i className="sf-skeleton" key={i} />
              ))}
            </div>
          </div>
          <div className="pdp-information">
            <i className="sf-skeleton" />
            <i className="sf-skeleton" />
            <i className="sf-skeleton sf-skeleton-short" />
            <div className="sf-axis-skeleton">
              {[0, 1, 2, 3].map((i) => (
                <i className="sf-skeleton" key={i} />
              ))}
            </div>
            <div className="sf-skeleton sf-spec-skeleton" />
          </div>
        </div>
        <div className="pdp-buybox">
          <i className="sf-skeleton" />
          <i className="sf-skeleton" />
          <i className="sf-skeleton sf-spec-skeleton" />
          <i className="sf-skeleton" />
        </div>
      </div>
    </LoadingPlaceholder>
  );
}
