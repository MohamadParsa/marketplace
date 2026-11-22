import { useEffect, useId, useRef, useState } from "react";
import { navigate } from "./navigation";
const api = import.meta.env.VITE_API_BASE_URL ?? "";
type Suggestion = {
  id: string;
  name: string;
  url: string;
  kind: "product" | "category";
};
export function SearchBox() {
  const [query, setQuery] = useState(
    new URLSearchParams(location.search).get("q") ?? "",
  );
  const [items, setItems] = useState<Suggestion[]>([]);
  const [open, setOpen] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(false);
  const [active, setActive] = useState(-1);
  const id = useId();
  const sequence = useRef(0);
  const wrap = useRef<HTMLFormElement>(null);
  useEffect(() => {
    const sync = () => {
      setQuery(new URLSearchParams(location.search).get("q") ?? "");
      setOpen(false);
    };
    addEventListener("popstate", sync);
    addEventListener("cherriva:navigate", sync);
    return () => {
      removeEventListener("popstate", sync);
      removeEventListener("cherriva:navigate", sync);
    };
  }, []);
  useEffect(() => {
    const seq = ++sequence.current;
    const controller = new AbortController();
    setItems([]);
    setError(false);
    setActive(-1);
    if (!query.trim()) {
      setLoading(false);
      return () => controller.abort();
    }
    setLoading(true);
    const timer = setTimeout(() => {
      void fetch(
        api + "/storefront/suggestions?q=" + encodeURIComponent(query.trim()),
        { signal: controller.signal },
      )
        .then(async (r) => {
          if (!r.ok) throw Error();
          const data = (await r.json()) as Suggestion[];
          if (sequence.current === seq && !controller.signal.aborted)
            setItems(data);
        })
        .catch(() => {
          if (sequence.current === seq && !controller.signal.aborted)
            setError(true);
        })
        .finally(() => {
          if (sequence.current === seq && !controller.signal.aborted)
            setLoading(false);
        });
    }, 250);
    return () => {
      clearTimeout(timer);
      controller.abort();
    };
  }, [query]);
  const submit = () => {
    if (!query.trim()) return;
    setOpen(false);
    navigate("/search?q=" + encodeURIComponent(query.trim()));
  };
  return (
    <form
      className="lh-search search-combobox"
      role="search"
      ref={wrap}
      onBlur={(e) => {
        if (!wrap.current?.contains(e.relatedTarget)) setOpen(false);
      }}
      onSubmit={(e) => {
        e.preventDefault();
        submit();
      }}
    >
      <svg
        viewBox="0 0 32 32"
        fill="none"
        stroke="currentColor"
        aria-hidden="true"
      >
        <circle cx="14" cy="14" r="8" />
        <path d="m20 20 7 7" />
      </svg>
      <input
        role="combobox"
        type="search"
        aria-label="جست‌وجو در چریوا"
        placeholder="جست‌وجو در چریوا"
        autoComplete="off"
        maxLength={80}
        aria-autocomplete="list"
        aria-expanded={open && !!query.trim()}
        aria-controls={id}
        aria-activedescendant={active >= 0 ? id + "-" + active : undefined}
        value={query}
        onChange={(e) => {
          sequence.current++;
          setQuery(e.target.value);
          setItems([]);
          setActive(-1);
          setOpen(true);
        }}
        onFocus={() => setOpen(true)}
        onKeyDown={(e) => {
          if (e.key === "Escape") {
            e.preventDefault();
            setOpen(false);
            setActive(-1);
          }
          if (e.key === "ArrowDown" || e.key === "ArrowUp") {
            e.preventDefault();
            setOpen(true);
            if (items.length)
              setActive(
                (a) =>
                  (a + (e.key === "ArrowDown" ? 1 : -1) + items.length) %
                  items.length,
              );
          }
          if (e.key === "Enter" && open && active >= 0 && items[active]) {
            e.preventDefault();
            setOpen(false);
            navigate(items[active].url);
          }
        }}
      />
      {open && query.trim() && (
        <div className="search-popup">
          <div id={id} role="listbox" aria-label="پیشنهادهای جست‌وجو">
            {items.map((item, i) => (
              <a
                id={id + "-" + i}
                key={item.kind + item.id}
                role="option"
                aria-selected={i === active}
                href={item.url}
                onMouseEnter={() => setActive(i)}
                onClick={() => setOpen(false)}
              >
                <span>{item.name}</span>
                <small>{item.kind === "category" ? "دسته‌بندی" : "کالا"}</small>
              </a>
            ))}
          </div>
          {loading ? (
            <p role="status">در حال جست‌وجو…</p>
          ) : error ? (
            <p role="status">پیشنهادها دریافت نشد؛ همه نتایج را بررسی کنید.</p>
          ) : !items.length ? (
            <p role="status">پیشنهادی پیدا نشد.</p>
          ) : null}
          <button type="submit" className="search-all">
            مشاهده همه نتایج برای «{query}»
          </button>
        </div>
      )}
    </form>
  );
}
