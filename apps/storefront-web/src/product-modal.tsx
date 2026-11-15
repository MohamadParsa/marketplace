import { useState } from "react";
import { categories, stores, money, type Product } from "./catalog";
import { Icon, Modal } from "./components";
export function ProductModal({
  product,
  color,
  setColor,
  onClose,
  onAdd,
}: {
  product: Product;
  color: string;
  setColor: (color: string) => void;
  onClose: () => void;
  onAdd: (p: Product, store: string, quantity: number) => boolean | undefined;
}) {
  const [storeId, setStoreId] = useState(stores[0].id);
  const [quantity, setQuantity] = useState(1);
  const store = stores.find((s) => s.id === storeId)!;
  return (
    <Modal title={product.name} onClose={onClose} wide>
      <div className="product-detail">
        <img
          className="detail-picture"
          src={`/images/${product.image}.svg`}
          alt={product.name}
        />
        <div>
          <span className="overline">
            {categories.find((c) => c.id === product.category)!.name}
          </span>
          <p className="description">{product.description}</p>
          <label>
            رنگ / نوع
            <select
              aria-label="رنگ / نوع"
              value={color}
              onChange={(e) => setColor(e.target.value)}
            >
              <option disabled value="">
                یک گزینه انتخاب کنید
              </option>
              {product.colors.map((c) => (
                <option key={c}>{c}</option>
              ))}
            </select>
          </label>
          <fieldset className="store-choices">
            <legend>مقایسه فروشگاه‌های نمونه</legend>
            {stores.map((s) => (
              <label
                key={s.id}
                className={
                  s.id === storeId ? "store-choice selected" : "store-choice"
                }
              >
                <input
                  type="radio"
                  name="store"
                  checked={s.id === storeId}
                  onChange={() => setStoreId(s.id)}
                />
                <span>
                  <strong>{s.name}</strong>
                  <small>آماده‌سازی: {s.preparation}</small>
                </span>
                <b>
                  {money(product.price + s.extra)}
                  <small> تومان</small>
                </b>
              </label>
            ))}
          </fieldset>
          <label>
            تعداد
            <input
              type="number"
              min="1"
              max={Math.max(1, product.stock)}
              value={quantity}
              onChange={(e) =>
                setQuantity(
                  Math.max(
                    1,
                    Math.min(
                      product.stock || 1,
                      Math.floor(Number(e.target.value)) || 1,
                    ),
                  ),
                )
              }
            />
          </label>
          <div className="detail-price">
            <span>
              {product.stock
                ? `موجودی نمونه: ${money(product.stock)}`
                : "ناموجود"}
            </span>
            <strong>
              {money((product.price + store.extra) * quantity)}{" "}
              <small>تومان</small>
            </strong>
          </div>
          <button
            className="primary full"
            disabled={!product.stock || !color}
            onClick={() => {
              if (onAdd(product, storeId, quantity)) onClose();
            }}
          >
            <Icon name="bag" /> افزودن به سبد خرید
          </button>
        </div>
      </div>
    </Modal>
  );
}
