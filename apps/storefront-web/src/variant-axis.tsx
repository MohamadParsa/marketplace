import { useId, useRef, useState } from "react";
import "./variant-axis.css";
export type Axis = {
  id: string;
  name: string;
  options: string[];
  groups?: { id: string; name: string; values: string[]; active: boolean }[];
  labels?: Record<string, string>;
  unit?: string;
};

// Exact decimal ordering for explicit units; labels sent to the API never change.
function orderedOptions(axis: Axis): string[] {
  const unit = axis.unit?.trim() ?? "";
  const normalize = (s: string) =>
    s
      .trim()
      .replace(/[۰-۹]/g, (c) => String(c.charCodeAt(0) - 1776))
      .replace(/[٠-٩]/g, (c) => String(c.charCodeAt(0) - 1632))
      .replace(/٫/g, ".");
  const gramUnits = ["گرم", "g"],
    milliUnits = ["میلی‌گرم", "میلی گرم", "mg"];
  const weightUnit = (u: string) =>
    gramUnits.includes(u) ? 3 : milliUnits.includes(u) ? 0 : null;
  const parse = (raw: string) => {
    let value = normalize(raw),
      shift = weightUnit(unit) ?? 0;
    const weight = value.match(
      /^(\d+(?:\.\d+)?)\s+(گرم|میلی‌گرم|میلی گرم|g|mg)$/,
    );
    if (weight && (!unit || weightUnit(unit) !== null)) {
      value = weight[1];
      shift = weightUnit(weight[2])!;
    } else {
      if (!unit) return null;
      if (value.endsWith(" " + unit))
        value = value.slice(0, -unit.length).trim();
    }
    if (!/^\d+(?:\.\d+)?$/.test(value)) return null;
    let [whole, fraction = ""] = value.split(".");
    // Convert explicit g to mg by a decimal shift, without floating-point rounding.
    if (shift) {
      fraction = fraction.padEnd(shift, "0");
      whole += fraction.slice(0, shift);
      fraction = fraction.slice(shift);
    }
    return {
      whole: whole.replace(/^0+(?=\d)/, ""),
      fraction: fraction.replace(/0+$/, ""),
    };
  };
  const values = axis.options.map((raw, index) => ({
    raw,
    index,
    number: parse(raw),
  }));
  if (values.some((v) => !v.number)) return axis.options;
  return values
    .sort((a, b) => {
      const x = a.number!,
        y = b.number!;
      if (x.whole.length !== y.whole.length)
        return x.whole.length - y.whole.length;
      if (x.whole !== y.whole) return x.whole < y.whole ? -1 : 1;
      const length = Math.max(x.fraction.length, y.fraction.length),
        xf = x.fraction.padEnd(length, "0"),
        yf = y.fraction.padEnd(length, "0");
      return xf === yf ? a.index - b.index : xf < yf ? -1 : 1;
    })
    .map((v) => v.raw);
}

export function VariantAxis({
  axis,
  selected,
  disabled,
  isDefined,
  onSelect,
}: {
  axis: Axis;
  selected?: string;
  disabled: boolean;
  isDefined: (value: string) => boolean;
  onSelect: (value: string) => void;
}) {
  const [expanded, setExpanded] = useState(false),
    id = useId();
  const root = useRef<HTMLFieldSetElement>(null);
  const options = orderedOptions(axis),
    limit = 6;
  let visible = expanded ? options : options.slice(0, limit);
  if (
    !expanded &&
    selected &&
    options.includes(selected) &&
    !visible.includes(selected)
  )
    visible = [...options.slice(0, limit - 1), selected];
  return (
    <fieldset ref={root} className="pdp-variant-axis" data-axis-id={axis.id}>
      <legend>
        {axis.name}:{" "}
        <b dir="auto">
          {selected ? (axis.labels?.[selected] ?? selected) : "انتخاب نشده"}
        </b>
      </legend>
      <div className="pdp-axis-options" id={id}>
        {visible.map((value) => (
          <button
            type="button"
            key={value}
            className="pdp-axis-chip"
            title={axis.labels?.[value] ?? value}
            aria-pressed={value === selected}
            disabled={disabled || !isDefined(value)}
            onClick={() => onSelect(value)}
          >
            <bdi>{axis.labels?.[value] ?? value}</bdi>
          </button>
        ))}
      </div>
      {options.length > limit && (
        <button
          type="button"
          className="pdp-axis-toggle"
          aria-expanded={expanded}
          aria-controls={id}
          aria-label={`${expanded ? "مشاهده کمتر" : "مشاهده همه"} گزینه‌های ${axis.name}`}
          disabled={disabled}
          onClick={() => {
            if (
              expanded &&
              (root.current?.getBoundingClientRect().top ?? 0) < 0
            )
              root.current?.scrollIntoView({
                block: "start",
                behavior: "instant",
              });
            setExpanded(!expanded);
          }}
        >
          {expanded
            ? "مشاهده کمتر"
            : `مشاهده همه (${options.length.toLocaleString("fa-IR")})`}
          <span aria-hidden="true">{expanded ? " −" : " +"}</span>
        </button>
      )}
    </fieldset>
  );
}
