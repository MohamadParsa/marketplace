import { categories, money, type Product } from "./catalog";
import { PhotoGallery } from "./photo-gallery";
import {
  PurchaseButton,
  VariantChips,
  type ProductPurchase,
} from "./product-purchase";

export function ProductCard({
  product,
  purchase,
  image,
  onImageChange,
  onDetails,
}: {
  product: Product;
  purchase: ProductPurchase;
  image: number;
  onImageChange: (index: number) => void;
  onDetails: () => void;
}) {
  return (
    <article className="product-card" data-product={product.id}>
      <PhotoGallery product={product} index={image} onChange={onImageChange} />
      <div className="product-info">
        <div className="product-title">
          <small>
            {
              categories.find((category) => category.id === product.category)
                ?.name
            }
          </small>
          <h3>{product.name}</h3>
        </div>
        <VariantChips
          product={product}
          color={purchase.color}
          onChange={purchase.onColorChange}
        />
        <div className="product-offer">
          <div>
            <strong>
              {product.stock ? purchase.store.name : "پیشنهاد فعال ندارد"}
            </strong>
            {product.stock > 0 && (
              <small>آماده‌سازی {purchase.store.preparation}</small>
            )}
          </div>
          <button
            type="button"
            className="details-button"
            aria-haspopup="dialog"
            aria-label={`مشاهده ${product.name}`}
            onClick={onDetails}
          >
            جزییات
          </button>
        </div>
        <div className="product-bottom">
          <small>
            {product.stock ? `موجودی ${money(product.stock)} عدد` : "ناموجود"}
          </small>
          <strong>{money(product.price + purchase.store.extra)} تومان</strong>
        </div>
        <PurchaseButton product={product} purchase={purchase} />
      </div>
    </article>
  );
}
