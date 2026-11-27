import { ChoiceField } from "../../shared/option-select";
export function CatalogChoice({
  label,
  value,
  options,
  onChange,
  clearLabel,
  disabled = false,
}: {
  label: string;
  value: string;
  options: { id: string; name: string }[];
  onChange: (value: string) => void;
  clearLabel?: string;
  disabled?: boolean;
}) {
  return (
    <ChoiceField
      label={label}
      value={value}
      options={options}
      onChange={onChange}
      clearLabel={clearLabel}
      disabled={disabled}
    />
  );
}
