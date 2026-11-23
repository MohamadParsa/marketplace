import { BusyOverlay } from "./busy-region";
import { useEffect, useRef } from "react";
import "./purchase-control.css";

/** One purchase control: add initially, then the selected design's diagonal counter. */
export function PurchaseControl({
  name,
  quantity,
  busy = false,
  canIncrease = true,
  onChange,
}: {
  name: string;
  quantity: number;
  busy?: boolean;
  canIncrease?: boolean;
  onChange: (quantity: number) => void;
}) {
  const root = useRef<HTMLDivElement>(null);
  const restore = useRef(false);
  useEffect(() => {
    if (restore.current && !busy) {
      restore.current = false;
      root.current
        ?.querySelector<HTMLButtonElement>("button:not(:disabled)")
        ?.focus({ preventScroll: true });
    }
  }, [quantity, busy]);
  const change = (next: number) => {
    restore.current = true;
    onChange(next);
  };
  return (
    <div
      ref={root}
      className="purchase-control"
      role="group"
      aria-label={`خرید ${name}`}
      aria-busy={busy}
    >
      {quantity === 0 ? (
        <button
          type="button"
          className="purchase-control-add"
          disabled={busy || !canIncrease}
          onClick={() => change(1)}
        >
          {canIncrease ? "افزودن به سبد خرید" : "ناموجود"}
        </button>
      ) : (
        <>
          <button
            type="button"
            className="purchase-control-more"
            disabled={busy || !canIncrease || quantity >= 99}
            aria-label={`افزایش تعداد ${name}`}
            onClick={() => change(quantity + 1)}
          >
            <output aria-live="polite">
              {quantity.toLocaleString("fa-IR")}
            </output>
          </button>
          <button
            type="button"
            className="purchase-control-less"
            disabled={busy}
            aria-label={
              quantity === 1 ? `حذف از سبد ${name}` : `کاهش تعداد ${name}`
            }
            onClick={() => change(quantity - 1)}
          >
            −
          </button>
        </>
      )}
      {busy && <BusyOverlay label="به‌روزرسانی سبد خرید" />}
    </div>
  );
}
