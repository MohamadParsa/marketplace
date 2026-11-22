import { useId } from "react";
import "./choice-group.css";
export function ChoiceGroup({
  label,
  value,
  options,
  onChange,
  disabled = false,
}: {
  label: string;
  value: string;
  options: {
    value: string;
    label: string;
    detail?: string;
    disabled?: boolean;
  }[];
  onChange: (value: string) => void;
  disabled?: boolean;
}) {
  const name = useId();
  return (
    <fieldset className="choice-group" disabled={disabled}>
      <legend>{label}</legend>
      <div className="choice-group-options">
        {options.map((o) => (
          <label key={o.value} className={o.disabled ? "disabled" : ""}>
            <input
              type="radio"
              name={name}
              value={o.value}
              checked={value === o.value}
              disabled={o.disabled}
              onChange={() => onChange(o.value)}
            />
            <span>
              {o.label}
              {o.detail && <small>{o.detail}</small>}
            </span>
          </label>
        ))}
      </div>
    </fieldset>
  );
}
