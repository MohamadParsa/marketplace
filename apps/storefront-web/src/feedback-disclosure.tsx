import { useId, useState, type ReactNode } from "react";
import "./expandable-content.css";
import "./product-feedback.css";

export function FeedbackDisclosure({
  label,
  children,
}: {
  label: string;
  children: ReactNode;
}) {
  const [open, setOpen] = useState(false);
  const id = useId();
  return (
    <div className="pdp-feedback-disclosure">
      <button
        className="pdp-expandable-toggle"
        type="button"
        aria-expanded={open}
        aria-controls={id}
        onClick={() => setOpen(!open)}
      >
        {label}
        <span aria-hidden="true">{open ? " −" : " +"}</span>
      </button>
      <div id={id} hidden={!open}>
        {children}
      </div>
    </div>
  );
}
