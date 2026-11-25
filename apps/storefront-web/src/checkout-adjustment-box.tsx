import { useId, useState, type ReactNode } from "react";

export type AdjustmentFeedback = {
  id: number;
  kind: "success" | "error";
  text: string;
};
export function CheckoutAdjustmentBox({
  kind,
  title,
  subtitle,
  applied,
  busy,
  summary,
  children,
  feedback,
  onRemove,
}: {
  kind: "coupon" | "wallet";
  title: string;
  subtitle?: ReactNode;
  applied: boolean;
  busy: boolean;
  summary: ReactNode;
  children: ReactNode;
  feedback?: AdjustmentFeedback;
  onRemove: () => void;
}) {
  const [expanded, setExpanded] = useState(false),
    id = useId();
  const open = applied || expanded;
  return (
    <section
      className={`buy-panel buy-${kind}-box buy-adjustment-box`}
      aria-label={title}
      data-applied={applied}
    >
      <h2>
        <button
          className="buy-adjustment-header"
          type="button"
          aria-expanded={open}
          aria-controls={id}
          aria-disabled={applied || busy}
          onClick={() => {
            if (!applied && !busy) setExpanded((v) => !v);
          }}
        >
          <span>
            {title}
            {subtitle && <small>{subtitle}</small>}
          </span>
          <span aria-hidden="true">{open ? "−" : "+"}</span>
        </button>
      </h2>
      <div className="buy-adjustment-collapse" data-open={open}>
        <div id={id} className="buy-adjustment-body" aria-hidden={!open}>
          {applied ? (
            <div className="buy-adjustment-applied">
              <p>{summary}</p>
              <button
                type="button"
                className="buy-link"
                disabled={busy}
                aria-label={
                  kind === "coupon" ? "حذف کد تخفیف" : "حذف اعتبار کیف پول"
                }
                onClick={() => {
                  setExpanded(true);
                  onRemove();
                }}
              >
                انصراف
              </button>
            </div>
          ) : (
            children
          )}
        </div>
      </div>
      {feedback && (
        <p
          key={feedback.id}
          className={feedback.kind === "error" ? "buy-alert" : "buy-notice"}
          role={feedback.kind === "error" ? "alert" : "status"}
        >
          {feedback.text}
        </p>
      )}
    </section>
  );
}
