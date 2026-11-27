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

export const persianDigits = (value: string | number) =>
  String(value).replace(/\d/g, (d) => "۰۱۲۳۴۵۶۷۸۹"[Number(d)]);
export const validISOInstant = (value: string) =>
  /^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d+)?(?:Z|[+-]\d{2}:\d{2})$/.test(
    value,
  ) &&
  validISODate(value.slice(0, 10)) &&
  Number.isFinite(Date.parse(value));
export function validISODate(iso: string): boolean {
  if (!/^\d{4}-\d{2}-\d{2}$/.test(iso)) return false;
  const date = new Date(iso + "T12:00:00Z");
  return (
    Number.isFinite(date.valueOf()) && date.toISOString().slice(0, 10) === iso
  );
}
export function addISODays(iso: string, count: number): string {
  if (!validISODate(iso)) return "";
  const date = new Date(iso + "T12:00:00Z");
  date.setUTCDate(date.getUTCDate() + count);
  return date.toISOString().slice(0, 10);
}
const wallClock = new Intl.DateTimeFormat("en-GB-u-ca-gregory-nu-latn", {
  timeZone: "Asia/Tehran",
  year: "numeric",
  month: "2-digit",
  day: "2-digit",
  hour: "2-digit",
  minute: "2-digit",
  second: "2-digit",
  hourCycle: "h23",
});
function wallParts(date: Date): number[] {
  const parts = wallClock.formatToParts(date);
  return ["year", "month", "day", "hour", "minute", "second"].map((type) =>
    Number(parts.find((p) => p.type === type)?.value),
  );
}
export function persianDateTime(instant: string): string {
  if (!instant) return "";
  const date = new Date(instant);
  if (!Number.isFinite(date.valueOf())) return "";
  const [y, m, d, h, min, sec] = wallParts(date);
  const day = persianCapacityDate(
    `${String(y).padStart(4, "0")}-${String(m).padStart(2, "0")}-${String(d).padStart(2, "0")}`,
  );
  return (
    day +
    " " +
    [h, min, ...(sec ? [sec] : [])]
      .map((n) => String(n).padStart(2, "0"))
      .join(":")
  );
}
export function parsePersianDateTime(text: string): string {
  const match = asciiDigits(text)
    .trim()
    .match(/^(\d{4}[/-]\d{1,2}[/-]\d{1,2})[ T](\d{1,2}):(\d{2})(?::(\d{2}))?$/);
  if (!match) return "";
  const day = parsePersianCapacityDate(match[1]),
    h = Number(match[2]),
    min = Number(match[3]),
    sec = Number(match[4] ?? 0);
  if (!day || h > 23 || min > 59 || sec > 59) return "";
  const [y, m, d] = day.split("-").map(Number),
    target = Date.UTC(y, m - 1, d, h, min, sec);
  let candidate = target - 210 * 60000;
  for (let attempt = 0; attempt < 4; attempt++) {
    const [wy, wm, wd, wh, wmin, ws] = wallParts(new Date(candidate));
    const actual = Date.UTC(wy, wm - 1, wd, wh, wmin, ws);
    if (actual === target) return new Date(candidate).toISOString();
    candidate += target - actual;
  }
  return "";
}
export function persianCapacityDate(iso: string): string {
  if (!validISODate(iso)) return "";
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
