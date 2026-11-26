/** Customer wording only: lead time is not a delivery promise. */
export function dispatchTime(days: number, basis?: string): string {
  return Number.isFinite(days) && days >= 0
    ? `ارسال از ${new Intl.NumberFormat("fa-IR").format(days)} روز ${basis === "calendar" ? "تقویمی" : "کاری"}`
    : "زمان ارسال مشخص نشده است";
}
