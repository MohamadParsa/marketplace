import { useEffect, useRef } from "react";
import { money, stores, type Product } from "./catalog";

export type ProductPurchase = {
  color: string;
  store: (typeof stores)[number];
  quantity: number;
  canAdd: boolean;
  onColorChange: (color: string) => void;
  onQuantityChange: (change: number) => void;
};

const swatches: Record<string, string> = {
  سفید: "#fffdf7",
  کرم: "#cda980",
  مشکی: "#25232a",
  نقره‌ای: "#b6b4be",
};

export function VariantChips({
  product,
  color,
  onChange,
}: {
  product: Product;
  color: string;
  onChange: (color: string) => void;
}) {
  return (
    <div
      className="chips variant-chips"
      role="group"
      aria-label={`رنگ یا نوع ${product.name}`}
    >
      {product.colors.map((choice) => (
        <button
          key={choice}
          type="button"
          aria-pressed={choice === color}
          disabled={!product.stock}
          onClick={() => onChange(choice)}
        >
          {swatches[choice] && (
            <span
              className="color-swatch"
              style={{ background: swatches[choice] }}
              aria-hidden="true"
            />
          )}
          {choice}
        </button>
      ))}
    </div>
  );
}

export function PurchaseButton({
  product,
  purchase,
}: {
  product: Product;
  purchase: ProductPurchase;
}) {
  const { quantity, canAdd, onQuantityChange } = purchase;
  const add = useRef<HTMLButtonElement>(null);
  const more = useRef<HTMLButtonElement>(null);
  const less = useRef<HTMLButtonElement>(null);
  const moveFocus = useRef(false);

  useEffect(() => {
    if (!moveFocus.current) return;
    moveFocus.current = false;
    (quantity ? (canAdd ? more.current : less.current) : add.current)?.focus({
      preventScroll: true,
    });
  }, [quantity, canAdd]);

  if (!quantity)
    return (
      <button
        ref={add}
        type="button"
        className="primary purchase-add"
        aria-label={product.stock ? `افزودن ${product.name} به سبد` : undefined}
        disabled={!canAdd}
        onClick={() => {
          moveFocus.current = true;
          onQuantityChange(1);
        }}
      >
        {product.stock ? "افزودن به سبد" : "ناموجود"}
      </button>
    );

  return (
    <div
      className="order-stepper"
      role="group"
      aria-label={`تعداد ${product.name} در سبد`}
    >
      <button
        ref={more}
        className="order-more"
        type="button"
        disabled={!canAdd}
        aria-label={`افزایش تعداد، ${money(quantity)} در سبد`}
        onClick={() => {
          moveFocus.current = true;
          onQuantityChange(1);
        }}
      >
        <span aria-live="polite">{money(quantity)}</span>
      </button>
      <button
        ref={less}
        className="order-less"
        type="button"
        aria-label={quantity === 1 ? "حذف از سبد" : "کاهش تعداد"}
        onClick={() => {
          if (quantity === 1) moveFocus.current = true;
          onQuantityChange(-1);
        }}
      >
        −
      </button>
    </div>
  );
}
