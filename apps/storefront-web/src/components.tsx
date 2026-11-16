import { useEffect, useId, useRef } from "react";
export function Icon({
  name,
}: {
  name: "search" | "bag" | "arrow" | "close" | "user";
}) {
  const paths = {
    search: (
      <>
        <circle cx="10.5" cy="10.5" r="6.5" />
        <path d="m16 16 4 4" />
      </>
    ),
    bag: (
      <>
        <path d="M5 7h14l1 14H4L5 7Z" />
        <path d="M8 8V6a4 4 0 0 1 8 0v2" />
      </>
    ),
    arrow: (
      <>
        <path d="M20 12H4m6-6-6 6 6 6" />
      </>
    ),
    close: <path d="m6 6 12 12M6 18 18 6" />,
    user: (
      <>
        <circle cx="12" cy="8" r="4" />
        <path d="M4 21a8 8 0 0 1 16 0" />
      </>
    ),
  };
  return (
    <svg
      width="21"
      height="21"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.6"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
    >
      {paths[name]}
    </svg>
  );
}
export function Modal({
  title,
  children,
  onClose,
  wide = false,
  className,
}: {
  title: string;
  children: React.ReactNode;
  onClose: () => void;
  wide?: boolean;
  className?: string;
}) {
  const ref = useRef<HTMLDialogElement>(null);
  const titleId = useId();
  useEffect(() => {
    const dialog = ref.current!;
    dialog.showModal();
    const previous = document.body.style.overflow;
    document.body.style.overflow = "hidden";
    return () => {
      dialog.close();
      document.body.style.overflow = previous;
    };
  }, []);
  return (
    <dialog
      ref={ref}
      className={["modal", wide && "wide", className].filter(Boolean).join(" ")}
      onCancel={(event) => {
        if (!event.defaultPrevented) onClose();
      }}
      onClick={(e) => {
        if (e.target === e.currentTarget) onClose();
      }}
      aria-labelledby={titleId}
    >
      <div className="modal-head">
        <h2 id={titleId}>{title}</h2>
        <button className="icon-button" aria-label="بستن" onClick={onClose}>
          <Icon name="close" />
        </button>
      </div>
      {children}
    </dialog>
  );
}

export function Brand({ onClick }: { onClick?: () => void }) {
  return (
    <a
      className="wordmark brand-logo"
      href="/"
      aria-label="Cherriva، صفحه اصلی"
      onClick={(event) => {
        if (onClick) {
          event.preventDefault();
          onClick();
        }
      }}
    >
      <img
        className="brand-letter"
        src="/brand/open-c-main-symbol.svg"
        alt=""
        width="64"
        height="64"
      />
      <span className="brand-tail" aria-hidden="true">
        <img src="/brand/open-c-main.svg" alt="" width="590" height="144" />
      </span>
    </a>
  );
}
