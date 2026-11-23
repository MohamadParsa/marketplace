import "./attribute-filter.css";
import { useId, useState } from "react";
export type AttributeFilterDefinition = {
  id: string;
  name: string;
  kind: string;
  unit?: string;
  options: string[];
};
export function AttributeFilter({
  filter: f,
  selected,
  change,
}: {
  filter: AttributeFilterDefinition;
  selected: string[];
  change: (values: string[]) => void;
}) {
  const [expanded, setExpanded] = useState(false);
  const id = useId();
  const choices = expanded ? f.options : f.options.slice(0, 5);
  const choose = (value: string, checked: boolean) =>
    change([
      ...selected.filter((v) => v !== value),
      ...(checked ? [value] : []),
    ]);
  return (
    <fieldset className="plp-attribute" data-filter-id={f.id}>
      <legend>
        {f.name}
        {f.unit ? " (" + f.unit + ")" : ""}
      </legend>
      {selected.length > 0 && (
        <div
          className="plp-filter-selected"
          aria-label={`فیلترهای انتخاب‌شده ${f.name}`}
        >
          {selected.map((v) => (
            <button
              type="button"
              key={v}
              aria-label={`حذف فیلتر ${f.name}: ${v}`}
              onClick={() => choose(v, false)}
            >
              {v} ×
            </button>
          ))}
        </div>
      )}
      {!f.options.length ? (
        <p>گزینه‌ای ثبت نشده است.</p>
      ) : (
        <div id={id} className="plp-attribute-options">
          {choices.map((value) => (
            <label key={value}>
              <input
                type="checkbox"
                checked={selected.includes(value)}
                onChange={(event) => choose(value, event.target.checked)}
              />
              <span className="plp-attribute-check" aria-hidden="true" />
              <span>{value}</span>
            </label>
          ))}
        </div>
      )}
      {f.options.length > 5 && (
        <button
          type="button"
          className="plp-filter-more"
          aria-expanded={expanded}
          aria-controls={id}
          onClick={() => setExpanded(!expanded)}
        >
          {expanded
            ? "موارد کمتر"
            : `موارد بیشتر (${(f.options.length - 5).toLocaleString("fa-IR")})`}
        </button>
      )}
    </fieldset>
  );
}
