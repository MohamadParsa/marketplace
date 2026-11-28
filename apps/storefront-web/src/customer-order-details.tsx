import { label, money as formatMoney, type Purchase } from "../../../packages/commerce-client/types";
import { purchaseProgress } from "../../../packages/commerce-client/order-progress";
import { OrderProgress } from "../../../packages/commerce-client/order-progress-view";
import "../../../packages/commerce-client/order-requests.css";
const money=(value:string)=>formatMoney(value).replace(/\d/g,d=>"۰۱۲۳۴۵۶۷۸۹"[Number(d)]);

export function CustomerPurchaseProgress({purchase:p}:{purchase:Purchase}) {
 return <section className="order-requests-panel" aria-label="روند سفارش"><h3>روند سفارش</h3>
  <OrderProgress steps={purchaseProgress(p.order_status,p.created_at,p.events??[])} label="مراحل سفارش"/>
  {p.order_status==="cancelled"&&<p className="order-progress-exception" role="status">سفارش لغو شده است؛ وضعیت استرداد را در درخواست‌ها و بخش مالی ببینید.</p>}
  <p>وضعیت پرداخت: <strong>{label(p.payment_status)}</strong></p>
 </section>;
}
export function CustomerOrderFinance({purchase:p}:{purchase:Purchase}) {
 return <section className="order-requests-panel customer-order-finance" aria-label="پرداخت‌ها و تخفیف سفارش"><h3>پرداخت‌ها و تخفیف سفارش</h3>
 <dl>
  <div><dt>جمع کالاها</dt><dd>{money(p.quote.reference_subtotal_toman??p.quote.cash_subtotal_toman??p.quote.subtotal_toman)}</dd></div>
  {Number(p.quote.markup_toman??"0")>0&&<div><dt>افزایش روش فروش {p.quote.sales_method?.name}</dt><dd>{money(p.quote.markup_toman!)}</dd></div>}
  {Number(p.quote.direct_discount_toman??"0")>0&&<div><dt>تخفیف کالاها</dt><dd>− {money(p.quote.direct_discount_toman!)}</dd></div>}
  <div><dt>هزینه ارسال</dt><dd>{money(p.quote.shipping_toman)}</dd></div>
  <div><dt>کد تخفیف {p.quote.coupon_code&&<bdi>{p.quote.coupon_code}</bdi>}</dt><dd>{money(p.quote.discount_toman??"0")}</dd></div>
  <div><dt>مبلغ نهایی سفارش</dt><dd>{money(p.quote.total_toman)}</dd></div>
  <div><dt>پرداخت از کیف پول</dt><dd>{money(p.wallet_consumed_toman??"0")}</dd></div>
 </dl>
 {!p.confirmed_at&&<p>تا پیش از قطعی‌شدن پرداخت، اعتبار انتخاب‌شده مصرف قطعی محسوب نمی‌شود.</p>}
 {p.payment_status==="returned-to-wallet"&&<p>وجه سفارش به کیف پول بازگردانده شده است؛ جزئیات رسیدگی را در درخواست‌ها ببینید.</p>}
 <h4>سابقه پرداخت و تأمین اعتبار</h4>
 <p className="buy-note">واریزهای تأییدشده ابتدا کیف پول را شارژ می‌کنند؛ مبلغ مصرف سفارش جداگانه در بالا آمده است.</p>
 {(p.payment_attempts??[]).length?<ul className="customer-payment-history">{p.payment_attempts!.map(attempt=><li key={attempt.id}>
  <strong>{attempt.purpose==="order-payment"?"پرداخت سفارش از کیف پول":"واریز برای تأمین اعتبار"}</strong>
  <span>{attempt.method_name} · {money(attempt.amount_toman)} · {label(attempt.status)}</span>
  <time dateTime={attempt.updated_at}>{new Date(attempt.updated_at).toLocaleString("fa-IR",{timeZone:"Asia/Tehran"})}</time>
 </li>)}</ul>:<p>سابقه پرداخت ثبت‌شده‌ای برای نمایش وجود ندارد.</p>}
 </section>;
}
