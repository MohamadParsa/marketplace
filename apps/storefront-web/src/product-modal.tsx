import { money, stores, type Product } from "./catalog";
import { Modal } from "./components";
import { PhotoGallery } from "./photo-gallery";
import {
  PurchaseButton,
  VariantChips,
  type ProductPurchase,
} from "./product-purchase";

export function ProductModal({
  product,
  purchase,
  image,
  onImageChange,
  onStoreChange,
  onClose,
}: {
  product: Product;
  purchase: ProductPurchase;
  image: number;
  onImageChange: (index: number) => void;
  onStoreChange: (storeId: string) => void;
  onClose: () => void;
}) {
  return (
    <Modal
      title={product.name}
      onClose={onClose}
      wide
      className="product-modal"
    >
      <div className="product-detail">
        <PhotoGallery
          product={product}
          index={image}
          onChange={onImageChange}
          expanded
        />
        <div className="detail-controls">
          <p className="description">{product.description}</p>
          <VariantChips
            product={product}
            color={purchase.color}
            onChange={purchase.onColorChange}
          />
          {product.stock > 0 && (
            <fieldset className="store-choices">
              <legend>مقایسه و انتخاب فروشنده</legend>
              {stores.map((store) => (
                <label key={store.id} className="store-choice">
                  <input
                    type="radio"
                    name="store"
                    checked={purchase.store.id === store.id}
                    onChange={() => onStoreChange(store.id)}
                  />
                  <span>
                    <strong>{store.name}</strong>
                    <small>آماده‌سازی {store.preparation}</small>
                  </span>
                  <b>{money(product.price + store.extra)} تومان</b>
                </label>
              ))}
            </fieldset>
          )}
          <div className="detail-price">
            <strong>{money(product.price + purchase.store.extra)} تومان</strong>
            <small>
              {product.stock ? `موجودی ${money(product.stock)} عدد` : "ناموجود"}
            </small>
          </div>
          <PurchaseButton product={product} purchase={purchase} />
        </div>
      </div>
    </Modal>
  );
}
