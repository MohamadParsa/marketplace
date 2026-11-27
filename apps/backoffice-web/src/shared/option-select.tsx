import {
  Autocomplete,
  Checkbox,
  TextField,
  type TextFieldProps,
} from "@mui/material";

export type SelectOption = { id: string; name: string; disabled?: boolean };
export function ChoiceField({
  label,
  value,
  options,
  onChange,
  disabled,
  clearLabel,
  ...field
}: {
  label: string;
  value: string;
  options: SelectOption[];
  onChange: (value: string) => void;
  disabled?: boolean;
  clearLabel?: string;
} & Pick<TextFieldProps, "required" | "helperText" | "error">) {
  const choices = clearLabel
    ? [{ id: "", name: clearLabel }, ...options.filter((o) => o.id !== "")]
    : options;
  return (
    <Autocomplete
      fullWidth
      size="small"
      options={choices}
      disabled={disabled}
      value={
        choices.find((o) => o.id === value) ??
        (value ? { id: value, name: value } : null)
      }
      getOptionLabel={(o) => o.name}
      isOptionEqualToValue={(a, b) => a.id === b.id}
      getOptionDisabled={(o) => !!o.disabled}
      disableClearable={!clearLabel}
      onChange={(_, next) => onChange(next?.id ?? "")}
      noOptionsText="گزینه‌ای پیدا نشد"
      openText="نمایش گزینه‌ها"
      closeText="بستن گزینه‌ها"
      clearText="پاک‌کردن انتخاب"
      renderInput={(params) => (
        <TextField {...params} {...field} label={label} />
      )}
    />
  );
}
export function MultiChoiceField({
  label,
  value,
  options,
  onChange,
  disabled,
  disableClearable = false,
}: {
  label: string;
  value: string[];
  options: SelectOption[];
  onChange: (value: string[]) => void;
  disabled?: boolean;
  disableClearable?: boolean;
}) {
  const choices = [
    ...options,
    ...value
      .filter((id) => !options.some((o) => o.id === id))
      .map((id) => ({ id, name: id, disabled: true })),
  ];
  return (
    <Autocomplete
      multiple
      fullWidth
      size="small"
      disableCloseOnSelect
      disableClearable={disableClearable}
      options={choices}
      disabled={disabled}
      value={choices.filter((o) => value.includes(o.id))}
      getOptionLabel={(o) => o.name}
      isOptionEqualToValue={(a, b) => a.id === b.id}
      getOptionDisabled={(o) => !!o.disabled}
      onChange={(_, next) => onChange(next.map((o) => o.id))}
      renderTags={(selected) => (
        <span className="admin-select-summary">
          {selected.map((o) => o.name).join("، ")}
        </span>
      )}
      renderOption={(props, option, { selected }) => {
        const { key, ...rest } = props;
        return (
          <li key={key} {...rest}>
            <Checkbox checked={selected} tabIndex={-1} disableRipple />
            {option.name}
          </li>
        );
      }}
      noOptionsText="گزینه‌ای پیدا نشد"
      openText="نمایش گزینه‌ها"
      closeText="بستن گزینه‌ها"
      clearText="پاک‌کردن انتخاب‌ها"
      renderInput={(params) => <TextField {...params} label={label} />}
    />
  );
}

// Each option owns its existing state transition, including permission and reset rules.
export function OptionSelect({
  label,
  options = [],
  multiple = false,
}: {
  label: string;
  multiple?: boolean;
  options?: {
    id: string;
    name: string;
    selected: boolean;
    disabled?: boolean;
    onSelect: () => void;
  }[];
}) {
  const selected = options.filter((o) => o.selected).map((o) => o.id);
  return multiple ? (
    <MultiChoiceField
      label={label}
      value={selected}
      options={options}
      disableClearable
      disabled={options.every((o) => o.disabled)}
      onChange={(ids) => {
        const changed = options.filter(
          (o) => o.selected !== ids.includes(o.id),
        );
        if (changed.length === 1) changed[0].onSelect();
      }}
    />
  ) : (
    <ChoiceField
      label={label}
      value={selected[0] ?? ""}
      options={options}
      disabled={options.every((o) => o.disabled)}
      onChange={(id) => options.find((o) => o.id === id)?.onSelect()}
    />
  );
}
