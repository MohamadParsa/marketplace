export const number = (value: number) =>
  new Intl.NumberFormat("fa-IR").format(value);
export const money = (rials: number) => `${number(rials / 10)} تومان`;
export const recordNumber = (state: { record_numbers?: Record<string, number> }, kind: string, id: string) => state.record_numbers?.[kind + "|" + id]?.toLocaleString("fa-IR", { useGrouping: false }) ?? "—";
export const variantCode = (state: { record_numbers?: Record<string, number> }, id: string) => {
  const value = state.record_numbers?.["variant|" + id];
  return value ? "chr-" + value : "کد در حال تخصیص";
};
export const displayRecord = (state: { record_numbers?: Record<string, number> }, id: string) => {
  const key = Object.keys(state.record_numbers ?? {}).find(key => key.endsWith("|" + id));
  return key ? state.record_numbers![key].toLocaleString("fa-IR", { useGrouping: false }) : "—";
};
export const date = (value: string) =>
  new Intl.DateTimeFormat("fa-IR", {
    dateStyle: "medium",
    timeZone: "Asia/Tehran",
  }).format(new Date(value));
export const dateTime = (value: string) =>
  new Intl.DateTimeFormat("fa-IR", {dateStyle:"medium",timeStyle:"short",timeZone:"Asia/Tehran"}).format(new Date(value));
export const labels: Record<string, string> = {
  placed: "ثبت سفارش",
  processing: "در حال پردازش",
  completed: "تکمیل شده",
  cancelled: "لغو شده",
  supplied: "تأمین شده",
  packing: "در حال آماده‌سازی",
  "awaiting-supply": "در انتظار تأمین",
  paid: "پرداخت شده",
  unpaid: "پرداخت نشده",
  "returned-to-wallet": "برگشت به کیف پول",
  success: "موفق",
  failure: "ناموفق",
  unknown: "نیازمند بررسی",
  pending: "در انتظار",
  published: "منتشرشده",
  rejected: "ردشده",
  open: "باز",
  closed: "بسته‌شده",
  "in-progress": "در حال رسیدگی",
  assigned: "واگذارشده",
  "refund-pending": "در انتظار بازپرداخت",
  "on-hold": "توقف تأمین",
  preparing: "در حال آماده‌سازی",
  shipped: "ارسال‌شده",
  delivered: "تحویل‌شده",
  "awaiting-payment": "در انتظار پرداخت",
  "awaiting-confirmation": "در انتظار تأیید",
  "supply-exception": "عدم تأمین",
  confirmed: "تأییدشده",
  payment: "پرداخت",
  refund: "بازپرداخت",
  withdrawal: "برداشت",
  settlement: "تسویه",
  debt: "بدهی",
  outstanding: "تسویه‌نشده",
  shared: "مشترک",
  independent: "مستقل",
  category: "دسته",
  store: "فروشگاه",
  product: "محصول",
  percent: "درصدی",
  fixed: "مبلغ ثابت",
};
