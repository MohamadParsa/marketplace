export const persianDigits = (value: string) =>
  value.replace(
    /[0-9٠-٩]/g,
    (digit) =>
      "۰۱۲۳۴۵۶۷۸۹"[digit >= "٠" ? digit.charCodeAt(0) - 1632 : Number(digit)],
  );

export const numericInput = (value: string) =>
  value
    .replace(/[۰-۹]/g, (digit) => String(digit.charCodeAt(0) - 1776))
    .replace(/[٠-٩]/g, (digit) => String(digit.charCodeAt(0) - 1632))
    .replace(/[٬,\s]/g, "")
    .replace(/٫/g, ".");

// The API uses tenths of a toman. Parse the decimal exactly before converting to a safe number.
export function tomanInputMinor(value: string): number | null {
  const normalized = numericInput(value);
  if (!/^\d+(?:\.\d)?$/.test(normalized)) return null;
  const [whole, fraction = "0"] = normalized.split(".");
  const minor = BigInt(whole) * 10n + BigInt(fraction);
  return minor <= BigInt(Number.MAX_SAFE_INTEGER) ? Number(minor) : null;
}

export function displayPriceInput(value: string, editing = false): string {
  if (editing) return persianDigits(value.replace(/\./g, "٫"));
  const [whole, fraction] = value.split(".");
  return persianDigits(
    whole.replace(/\B(?=(\d{3})+(?!\d))/g, "٬") +
      (fraction == null ? "" : "٫" + fraction),
  );
}

export const attributeLabel = (name: string, value: string) =>
  /گارانتی|ضمانت|warranty/i.test(name) ? persianDigits(value) : value;
