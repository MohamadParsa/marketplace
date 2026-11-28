import { useRef, useLayoutEffect, type ReactNode } from "react";
import "./busy-region.css";

export function BusyOverlay({ label = "در حال پردازش…" }: { label?: string }) {
  return (
    <div className="sf-progress-overlay">
      <div className="sf-progress-bar" role="progressbar" aria-label={label} />
      <span className="sf-progress-announcement" role="status">
        {label}
      </span>
    </div>
  );
}

/** Blocks interaction in the affected area; keeps its content and height mounted. */
export function BusyRegion({
  busy,
  label,
  children,
  className = "",
}: {
  busy: boolean;
  label?: string;
  children: ReactNode;
  className?: string;
}) {
  const content = useRef<HTMLDivElement>(null);
  const previousFocus = useRef<HTMLElement | null>(null);
  useLayoutEffect(() => {
    // Attribute form supports inert on React 18 without adding a polyfill/dependency.
    if (busy && content.current?.contains(document.activeElement))
      previousFocus.current = document.activeElement as HTMLElement;
    content.current?.toggleAttribute("inert", busy);
    if (!busy && previousFocus.current) {
      const previous = previousFocus.current;
      previousFocus.current = null;
      if (previous.isConnected && document.activeElement === document.body)
        previous.focus({ preventScroll: true });
    }
  }, [busy]);
  return (
    <div className={`sf-busy-region ${className}`} aria-busy={busy}>
      <div className="sf-busy-content" ref={content}>
        {children}
      </div>
      {busy && <BusyOverlay label={label} />}
    </div>
  );
}
export function LoadingPlaceholder({
  label = "در حال دریافت اطلاعات…",
  children,
  className = "",
}: {
  children?: ReactNode;
  className?: string;
  label?: string;
}) {
  return (
    <BusyRegion
      busy
      label={label}
      className={`sf-loading-placeholder ${className}`}
    >
      {children ?? (
        <div className="sf-skeleton-lines" aria-hidden="true">
          <i />
          <i />
          <i />
        </div>
      )}
    </BusyRegion>
  );
}
