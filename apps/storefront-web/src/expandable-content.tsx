import {
  useId,
  useLayoutEffect,
  useRef,
  useState,
  type ReactNode,
} from "react";
import "./expandable-content.css";

/** Preserve complete sanitized markup in DOM while limiting its visible height. */
export function ExpandableContent({
  label,
  children,
  resetKey,
}: {
  label: string;
  children: ReactNode;
  resetKey: string;
}) {
  const id = useId(),
    body = useRef<HTMLDivElement>(null),
    button = useRef<HTMLButtonElement>(null);
  const [expanded, setExpanded] = useState(false),
    [overflow, setOverflow] = useState(false);
  useLayoutEffect(() => {
    setExpanded(false);
  }, [resetKey]);
  useLayoutEffect(() => {
    const element = body.current;
    if (!element) return;
    const saved = new Map<HTMLElement, string | null>();
    const restore = () => {
      saved.forEach((value, item) => {
        if (value === null) item.removeAttribute("tabindex");
        else item.setAttribute("tabindex", value);
      });
      saved.clear();
    };
    const measure = () => {
      restore();
      const limit =
        parseFloat(
          getComputedStyle(element).getPropertyValue("--pdp-content-limit"),
        ) || 320;
      const hasOverflow = element.scrollHeight > limit + 1;
      setOverflow(hasOverflow);
      if (!expanded && hasOverflow) {
        const edge = element.getBoundingClientRect().top + limit - 26;
        element
          .querySelectorAll<HTMLElement>(
            "a[href],button,input,select,textarea,[tabindex]",
          )
          .forEach((item) => {
            if (item.getBoundingClientRect().bottom > edge) {
              saved.set(item, item.getAttribute("tabindex"));
              item.setAttribute("tabindex", "-1");
              if (item === document.activeElement)
                button.current?.focus({ preventScroll: true });
            }
          });
      }
    };
    measure();
    const observer = new ResizeObserver(measure);
    observer.observe(element);
    if (element.firstElementChild) observer.observe(element.firstElementChild);
    element.addEventListener("load", measure, true);
    window.addEventListener("resize", measure);
    document.fonts.addEventListener("loadingdone", measure);
    return () => {
      observer.disconnect();
      element.removeEventListener("load", measure, true);
      window.removeEventListener("resize", measure);
      document.fonts.removeEventListener("loadingdone", measure);
      restore();
    };
  }, [children, expanded, resetKey]);
  return (
    <div
      className={`pdp-expandable ${!expanded && overflow ? "is-collapsed" : ""}`}
    >
      <div id={id} ref={body} className="pdp-expandable-body">
        {children}
      </div>
      {overflow && (
        <button
          ref={button}
          type="button"
          className="pdp-expandable-toggle"
          aria-expanded={expanded}
          aria-controls={id}
          aria-label={`${expanded ? "مشاهده کمتر" : "مشاهده بیشتر"} ${label}`}
          onClick={() => {
            if (expanded) {
              const top =
                body.current?.parentElement?.getBoundingClientRect().top ?? 0;
              // Collapse may move the button above the viewport; keep the same section in sight.
              if (top < 0)
                body.current?.parentElement?.scrollIntoView({
                  block: "start",
                  behavior: "instant",
                });
            }
            setExpanded(!expanded);
          }}
        >
          {expanded ? "مشاهده کمتر" : "مشاهده بیشتر"}
          <span aria-hidden="true">{expanded ? " −" : " +"}</span>
        </button>
      )}
    </div>
  );
}
