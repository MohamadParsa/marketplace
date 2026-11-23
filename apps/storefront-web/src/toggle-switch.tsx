import "./toggle-switch.css";
/** Native boolean state/label semantics, with a switch presentation. Never submits on its own. */
export function ToggleSwitch({
  checked,
  onChange,
  disabled = false,
  invalid = false,
  label,
  describedBy,
  name,
}: {
  checked: boolean;
  onChange: (value: boolean) => void;
  disabled?: boolean;
  invalid?: boolean;
  label?: string;
  describedBy?: string;
  name?: string;
}) {
  return (
    <span
      className={`sf-switch${disabled ? " is-disabled" : ""}${invalid ? " is-invalid" : ""}`}
      dir="ltr"
    >
      <input
        className="sf-switch-input"
        type="checkbox"
        role="switch"
        checked={checked}
        disabled={disabled}
        name={name}
        aria-label={label}
        aria-describedby={describedBy}
        aria-invalid={invalid || undefined}
        onChange={(e) => onChange(e.currentTarget.checked)}
      />
      <span className="sf-switch-track" aria-hidden="true">
        <span className="sf-switch-thumb" />
      </span>
    </span>
  );
}
