import { useState } from "react";
import { Button, TextField } from "@mui/material";

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
  const [search, setSearch] = useState(""),
    [limit, setLimit] = useState(20);
  const rows = options.filter((o) =>
    o.name.toLocaleLowerCase().includes(search.trim().toLocaleLowerCase()),
  );
  return (
    <fieldset className="catalog-choice" disabled={disabled}>
      <legend>{label}</legend>
      {options.length > 5 && (
        <TextField
          size="small"
          label={`جست‌وجوی ${label}`}
          value={search}
          onChange={(e) => {
            setSearch(e.target.value);
            setLimit(20);
          }}
        />
      )}
      <small>
        انتخاب فعلی:{" "}
        {options.find((o) => o.id === value)?.name ||
          clearLabel ||
          "انتخاب نشده"}
      </small>
      <div className="catalog-choice-options">
        {clearLabel && (
          <Button
            size="small"
            variant={!value ? "contained" : "outlined"}
            aria-pressed={!value}
            onClick={() => onChange("")}
          >
            {clearLabel}
          </Button>
        )}
        {rows.slice(0, limit).map((o) => (
          <Button
            key={o.id}
            size="small"
            variant={o.id === value ? "contained" : "outlined"}
            aria-pressed={o.id === value}
            onClick={() => onChange(o.id)}
          >
            {o.name}
          </Button>
        ))}
      </div>
      {!rows.length && <small>گزینه‌ای پیدا نشد.</small>}
      {rows.length > limit && (
        <Button onClick={() => setLimit((n) => n + 20)}>
          نمایش گزینه‌های بیشتر
        </Button>
      )}
    </fieldset>
  );
}
