import { useEffect, useRef, useState, type CSSProperties } from "react";
import "./price-range-filter.css";
export type PriceRange = { min: number | null; max: number | null };
export function PriceRangeFilter({
  range,
  params,
  change,
}: {
  range?: PriceRange;
  params: URLSearchParams;
  change: (edit: (p: URLSearchParams) => void) => void;
}) {
  const query = params.toString();
  const initial = () => [
    params.has("min_price") ? String(Number(params.get("min_price")) / 10) : "",
    params.has("max_price") ? String(Number(params.get("max_price")) / 10) : "",
  ];
  const [draft, setDraft] = useState(initial);
  const current = useRef(draft);
  const timer = useRef<ReturnType<typeof setTimeout>>();
  const [error, setError] = useState("");
  const cancel = () => {
    if (timer.current) clearTimeout(timer.current);
    timer.current = undefined;
  };
  useEffect(() => {
    cancel();
    const v = initial();
    current.current = v;
    setDraft(v);
    setError("");
    return cancel;
  }, [query]);
  const update = (v: string[]) => {
    current.current = v;
    setDraft(v);
    setError("");
  };
  const apply = () => {
    cancel();
    const [min, max] = current.current;
    const low = min === "" ? null : Math.round(Number(min) * 10),
      high = max === "" ? null : Math.round(Number(max) * 10);
    if (
      (low !== null && (!Number.isSafeInteger(low) || low < 0)) ||
      (high !== null && (!Number.isSafeInteger(high) || high < 0)) ||
      (low !== null && high !== null && low > high)
    ) {
      setError("بازه قیمت معتبر وارد کنید.");
      return;
    }
    const before = params.get("min_price") ?? "",
      after = params.get("max_price") ?? "";
    if (
      (low === null ? "" : String(low)) === before &&
      (high === null ? "" : String(high)) === after
    )
      return;
    change((p) => {
      low === null ? p.delete("min_price") : p.set("min_price", String(low));
      high === null ? p.delete("max_price") : p.set("max_price", String(high));
    });
  };
  const applied = ["min_price", "max_price"].flatMap((key) => {
    const n = Number(params.get(key));
    return params.has(key) && Number.isSafeInteger(n) && n >= 0 ? [n] : [];
  });
  const floor = Math.min(range?.min ?? 0, ...applied),
    ceiling = Math.max(range?.max ?? 0, ...applied),
    available = range?.min != null && range?.max != null,
    enabled = available && ceiling > floor;
  const clamp = (n: number) => Math.min(ceiling, Math.max(floor, n));
  const low = clamp(draft[0] === "" ? floor : Number(draft[0]) * 10),
    high = clamp(draft[1] === "" ? ceiling : Number(draft[1]) * 10);
  const minValue = Number.isFinite(low) ? Math.min(low, high) : floor,
    maxValue = Number.isFinite(high) ? Math.max(low, high) : ceiling;
  const slide = (index: number, value: number) => {
    cancel();
    const bounded =
      index === 0 ? Math.min(value, maxValue) : Math.max(value, minValue);
    const next = [...current.current];
    next[index] = String(bounded / 10);
    update(next);
    timer.current = setTimeout(apply, 350);
  };
  const money = (n: number) =>
    (n / 10).toLocaleString("fa-IR", { maximumFractionDigits: 1 }) + " تومان";
  const start = enabled ? (100 * (minValue - floor)) / (ceiling - floor) : 0,
    end = enabled ? (100 * (maxValue - floor)) / (ceiling - floor) : 0;
  return (
    <form
      className="plp-price-filter"
      onSubmit={(e) => {
        e.preventDefault();
        apply();
      }}
    >
      <h3>
        محدوده قیمت <small>تومان</small>
      </h3>
      <div
        className="price-slider"
        dir="ltr"
        data-enabled={enabled}
        style={
          {
            "--range-start": start + "%",
            "--range-end": end + "%",
          } as CSSProperties
        }
      >
        <div className="price-slider-track" aria-hidden="true">
          <span />
        </div>
        {[0, 1].map((i) => (
          <input
            key={i}
            type="range"
            min={floor}
            max={ceiling}
            step={1}
            value={i === 0 ? minValue : maxValue}
            disabled={!enabled}
            aria-label={
              i === 0 ? "حداقل قیمت روی اسلایدر" : "حداکثر قیمت روی اسلایدر"
            }
            aria-valuetext={money(i === 0 ? minValue : maxValue)}
            className={i === 0 ? "range-low" : "range-high"}
            style={{ zIndex: i === 0 && minValue === maxValue ? 4 : undefined }}
            onChange={(e) => slide(i, Number(e.target.value))}
            onPointerUp={apply}
          />
        ))}
      </div>
      {!enabled && (
        <p className="price-slider-note">
          {available
            ? "قیمت‌های این نتیجه یکسان‌اند."
            : "قیمت قابل خریدی برای این نتیجه وجود ندارد."}
        </p>
      )}
      <div className="price-slider-inputs">
        {["از", "تا"].map((label, i) => (
          <label key={i}>
            {label}
            <input
              type="number"
              min="0"
              step="0.1"
              dir="ltr"
              aria-label={
                i === 0 ? "حداقل قیمت (تومان)" : "حداکثر قیمت (تومان)"
              }
              value={draft[i]}
              placeholder={
                available ? String((i === 0 ? floor : ceiling) / 10) : ""
              }
              onChange={(e) => {
                cancel();
                const next = [...current.current];
                next[i] = e.target.value;
                update(next);
              }}
            />
          </label>
        ))}
      </div>
      {error && <p role="alert">{error}</p>}
      <button type="submit">اعمال محدوده قیمت</button>
    </form>
  );
}
