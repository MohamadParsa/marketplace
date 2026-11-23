/** Customer wording only: lead time is not a delivery promise. */
export function dispatchTime(days: number): string {
  return Number.isFinite(days) && days >= 0
    ? `ارسال از ${new Intl.NumberFormat("fa-IR").format(days)} روز کاری`
    : "زمان ارسال مشخص نشده است";
}
