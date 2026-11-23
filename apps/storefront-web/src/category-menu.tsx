import { useEffect, useRef, useState } from "react";
import "./category-menu.css";
const api = import.meta.env.VITE_API_BASE_URL ?? "";
export type CategoryNode = {
  id: string;
  name: string;
  url: string;
  parent_id?: string;
};
function Branch({
  node,
  children,
  onNavigate,
}: {
  node: CategoryNode;
  children: Map<string, CategoryNode[]>;
  onNavigate: () => void;
}) {
  const [open, setOpen] = useState(false);
  const nested = children.get(node.id) ?? [];
  return (
    <li className="category-branch">
      <div>
        <a data-category-link href={node.url} onClick={onNavigate}>
          {node.name}
        </a>
        {nested.length > 0 && (
          <button
            type="button"
            aria-label={`زیرشاخه‌های ${node.name}`}
            aria-expanded={open}
            onClick={() => setOpen(!open)}
          >
            <Chevron open={open} />
          </button>
        )}
      </div>
      {open && (
        <ul>
          {nested.map((n) => (
            <Branch
              key={n.id}
              node={n}
              children={children}
              onNavigate={onNavigate}
            />
          ))}
        </ul>
      )}
    </li>
  );
}
function Chevron({ open = false }: { open?: boolean }) {
  return (
    <svg
      className={open ? "is-expanded" : ""}
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.7"
      aria-hidden="true"
    >
      <path d="m14 7-5 5 5 5" />
    </svg>
  );
}
export function CategoryMenu() {
  const [open, setOpen] = useState(false),
    [nodes, setNodes] = useState<CategoryNode[]>(),
    [error, setError] = useState(false),
    [attempt, setAttempt] = useState(0),
    [active, setActive] = useState("");
  const [mobile, setMobile] = useState(
    () => matchMedia("(max-width: 600px)").matches,
  );
  const trigger = useRef<HTMLButtonElement>(null),
    panel = useRef<HTMLDivElement>(null);
  const close = (restore = false) => {
    setOpen(false);
    if (restore) trigger.current?.focus();
  };
  useEffect(() => {
    const mq = matchMedia("(max-width: 600px)");
    const update = () => setMobile(mq.matches);
    mq.addEventListener("change", update);
    return () => mq.removeEventListener("change", update);
  }, []);
  useEffect(() => {
    const controller = new AbortController();
    setError(false);
    fetch(api + "/storefront/navigation", { signal: controller.signal })
      .then(async (r) => {
        if (!r.ok) throw Error();
        const rows = (await r.json()) as CategoryNode[];
        if (!Array.isArray(rows)) throw Error();
        setNodes(rows);
      })
      .catch(() => {
        if (!controller.signal.aborted) setError(true);
      });
    return () => controller.abort();
  }, [attempt]);
  useEffect(() => {
    if (!open) return;
    const previous = document.body.style.overflow;
    if (mobile) document.body.style.overflow = "hidden";
    const outside = (e: PointerEvent) => {
      if (
        !panel.current?.contains(e.target as Node) &&
        !trigger.current?.contains(e.target as Node)
      )
        close();
    };
    const key = (e: KeyboardEvent) => {
      if (e.key === "Escape") {
        e.preventDefault();
        close(true);
      }
      if (e.key === "Tab" && mobile) {
        const els = panel.current?.querySelectorAll<HTMLElement>(
          "a[href],button:not(:disabled)",
        );
        if (!els?.length) return;
        const first = els[0],
          last = els[els.length - 1];
        if (e.shiftKey && document.activeElement === first) {
          e.preventDefault();
          last.focus();
        } else if (!e.shiftKey && document.activeElement === last) {
          e.preventDefault();
          first.focus();
        }
      }
    };
    const focusOutside = (e: FocusEvent) => {
      if (
        !mobile &&
        !panel.current?.contains(e.target as Node) &&
        !trigger.current?.contains(e.target as Node)
      )
        close();
    };
    document.addEventListener("focusin", focusOutside);
    const route = () => close();
    document.addEventListener("pointerdown", outside);
    document.addEventListener("keydown", key);
    addEventListener("cherriva:navigate", route);
    addEventListener("popstate", route);
    const frame = requestAnimationFrame(() =>
      panel.current
        ?.querySelector<HTMLElement>("[data-category-link],button")
        ?.focus(),
    );
    return () => {
      cancelAnimationFrame(frame);
      if (mobile) document.body.style.overflow = previous;
      document.removeEventListener("focusin", focusOutside);
      document.removeEventListener("pointerdown", outside);
      document.removeEventListener("keydown", key);
      removeEventListener("cherriva:navigate", route);
      removeEventListener("popstate", route);
    };
  }, [open, mobile]);
  useEffect(() => {
    if (open && mobile)
      panel.current
        ?.querySelector<HTMLElement>(
          active ? ".category-menu-back" : ".category-menu-roots a",
        )
        ?.focus();
  }, [active, mobile, open]);
  const children = new Map<string, CategoryNode[]>();
  for (const n of nodes ?? []) {
    const p = n.parent_id ?? "";
    children.set(p, [...(children.get(p) ?? []), n]);
  }
  const roots = children.get("") ?? [];
  const selected = roots.find((r) => r.id === active);
  return (
    <div className="category-menu">
      <button
        ref={trigger}
        className="lh-menu-label category-menu-trigger"
        type="button"
        aria-haspopup="dialog"
        aria-expanded={open}
        aria-controls="catalog-category-menu"
        onClick={() => {
          setActive("");
          setOpen(!open);
        }}
        onKeyDown={(e) => {
          if (e.key === "ArrowDown") {
            e.preventDefault();
            setOpen(true);
          }
        }}
      >
        <svg
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          strokeWidth="1.7"
          aria-hidden="true"
        >
          <path d="M4 6h16M4 12h16M4 18h16" />
        </svg>
        دسته‌بندی کالاها
        <Chevron open={open} />
      </button>
      {open && (
        <>
          <div className="category-menu-shade" aria-hidden="true" />
          <div
            ref={panel}
            id="catalog-category-menu"
            role="dialog"
            aria-label="فهرست دسته‌بندی کالاها"
            aria-modal={mobile || undefined}
            className={`category-menu-panel ${selected ? "has-active" : ""}`}
          >
            <div className="category-menu-heading">
              <strong>دسته‌های چریوا</strong>
              <button
                type="button"
                aria-label="بستن دسته‌بندی‌ها"
                onClick={() => close(true)}
              >
                ×
              </button>
            </div>
            {!nodes && !error && (
              <p role="status" className="category-menu-message">
                در حال دریافت دسته‌ها…
              </p>
            )}
            {error && (
              <div className="category-menu-message" role="alert">
                <p>دسته‌ها دریافت نشدند.</p>
                <button type="button" onClick={() => setAttempt(attempt + 1)}>
                  تلاش دوباره
                </button>
              </div>
            )}
            {nodes &&
              !error &&
              (roots.length ? (
                <div className="category-menu-columns">
                  <nav
                    className="category-menu-roots"
                    aria-label="دسته‌های والد"
                  >
                    <ul>
                      {roots.map((n) => (
                        <li
                          key={n.id}
                          className={n.id === active ? "is-active" : ""}
                        >
                          <a
                            data-category-link
                            href={n.url}
                            onClick={() => close()}
                          >
                            {n.name}
                          </a>
                          {children.has(n.id) && (
                            <button
                              type="button"
                              aria-label={`زیرشاخه‌های ${n.name}`}
                              aria-expanded={n.id === active}
                              onClick={() => setActive(n.id)}
                            >
                              <Chevron open={n.id === active} />
                            </button>
                          )}
                        </li>
                      ))}
                    </ul>
                  </nav>
                  <nav
                    className="category-menu-children"
                    aria-label="زیرشاخه‌های دسته منتخب"
                  >
                    {selected ? (
                      <>
                        <button
                          type="button"
                          className="category-menu-back"
                          onClick={() => setActive("")}
                        >
                          بازگشت به همه دسته‌ها
                        </button>
                        <a
                          className="category-menu-parent-link"
                          href={selected.url}
                          onClick={() => close()}
                        >
                          همه کالاهای {selected.name}
                          <Chevron />
                        </a>
                        <ul key={selected.id}>
                          {(children.get(selected.id) ?? []).map((n) => (
                            <Branch
                              key={n.id}
                              node={n}
                              children={children}
                              onNavigate={() => close()}
                            />
                          ))}
                        </ul>
                      </>
                    ) : (
                      <p className="category-menu-placeholder">
                        برای دیدن زیرشاخه‌ها، پیکان کنار دسته را انتخاب کنید.
                      </p>
                    )}
                  </nav>
                </div>
              ) : (
                <p className="category-menu-message">
                  دسته‌ای برای نمایش وجود ندارد.
                </p>
              ))}
          </div>
        </>
      )}
    </div>
  );
}
