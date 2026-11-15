export const number = (value: number) =>
  new Intl.NumberFormat("fa-IR").format(value);
export const money = (rials: number) => `${number(rials / 10)} تومان`;
export const date = (value: string) =>
  new Intl.DateTimeFormat("fa-IR", {
    dateStyle: "medium",
    timeZone: "Asia/Tehran",
  }).format(new Date(value));
export const labels: Record<string, string> = {
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
