import type { DeliveryPlan } from "./types";

export const shippingDate = (date: string) =>
  new Date(date + "T12:00:00+03:30").toLocaleDateString("fa-IR", {
    timeZone: "Asia/Tehran",
    weekday: "long",
    day: "numeric",
    month: "long",
    year: "numeric",
  });
export const shippingClock = (date: string) =>
  new Date(date).toLocaleTimeString("fa-IR", {
    timeZone: "Asia/Tehran",
    hour: "2-digit",
    minute: "2-digit",
    hour12: false,
  });
export function deliveryPromise(
  plan: DeliveryPlan | undefined,
  from: string,
  to: string,
) {
  if (plan?.slot)
    return `تحویل ${shippingDate(plan.capacity_day)}، ساعت ${shippingClock(plan.slot.start_at)} تا ${shippingClock(plan.slot.end_at)}`;
  return `تحویل حدودی از ${shippingDate(from)} تا ${shippingDate(to)}`;
}
export function shippingCharge(mode: string | undefined, fee: string) {
  if (mode === "collect")
    return "هزینه هنگام تحویل توسط حمل‌کننده دریافت می‌شود؛ جدا از مبلغ سایت";
  if (Number(fee) === 0) return "ارسال رایگان";
  return `${Number(fee).toLocaleString("fa-IR")} تومان؛ پرداخت همراه سفارش`;
}
