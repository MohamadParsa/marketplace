const calendar = new Intl.DateTimeFormat("en-u-ca-persian-nu-latn", {
  timeZone: "Asia/Tehran",
  year: "numeric",
  month: "2-digit",
  day: "2-digit",
});
export const asciiDigits = (value: string) =>
  value.replace(/[۰-۹٠-٩]/g, (char) =>
    String(char.charCodeAt(0) - (char >= "۰" ? 1776 : 1632)),
  );
function persianParts(date: Date): number[] {
  const parts = calendar.formatToParts(date);
  return ["year", "month", "day"].map((type) =>
    Number(parts.find((p) => p.type === type)?.value),
  );
}
export function persianCapacityDate(iso: string): string {
  if (!/^\d{4}-\d{2}-\d{2}$/.test(iso)) return "";
  const date = new Date(iso + "T12:00:00+03:30");
  if (!Number.isFinite(date.valueOf())) return "";
  return persianParts(date)
    .map((part, i) => String(part).padStart(i === 0 ? 4 : 2, "0"))
    .join("/");
}
export function parsePersianCapacityDate(value: string): string {
  const match = asciiDigits(value)
    .trim()
    .match(/^(\d{4})[/-](\d{1,2})[/-](\d{1,2})$/);
  if (!match) return "";
  const target = match.slice(1).map(Number);
  if (
    target[0] < 1000 ||
    target[0] > 9000 ||
    target[1] < 1 ||
    target[1] > 12 ||
    target[2] < 1 ||
    target[2] > 31
  )
    return "";
  const targetNumber = target[0] * 10000 + target[1] * 100 + target[2];
  let low = Math.floor(Date.UTC(target[0] + 621, 0, 1, 12) / 86400000),
    high = Math.floor(Date.UTC(target[0] + 622, 11, 31, 12) / 86400000);
  while (low <= high) {
    const middle = Math.floor((low + high) / 2);
    const date = new Date(middle * 86400000 + 12 * 3600000);
    const parts = persianParts(date);
    const number = parts[0] * 10000 + parts[1] * 100 + parts[2];
    if (number === targetNumber) return date.toISOString().slice(0, 10);
    if (number < targetNumber) low = middle + 1;
    else high = middle - 1;
  }
  return "";
}
export function tehranCapacityDate(now = new Date()): string {
  const parts = new Intl.DateTimeFormat("en", {
    timeZone: "Asia/Tehran",
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
  }).formatToParts(now);
  return ["year", "month", "day"]
    .map((type) => parts.find((p) => p.type === type)?.value)
    .join("-");
}
