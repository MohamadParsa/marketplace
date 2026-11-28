import { useCallback, useEffect, useRef, useState } from "react";
import { call, useCommerce } from "./commerce";
import { label, money, type Entry, type Wallet } from "../../../packages/commerce-client/types";

const groups:Record<string,string>={"deposit-credit":"funding","admin-credit":"funding","order-payment":"payment","order-cancel-credit":"return","manual-refund":"withdrawal"};
const amount=(value:string)=>money(value).replace(/\d/g,d=>"۰۱۲۳۴۵۶۷۸۹"[Number(d)]);
const date=(value:string)=>new Date(value).toLocaleString("fa-IR",{timeZone:"Asia/Tehran"});
export function WalletTransactions(){
 const {token}=useCommerce();const [wallet,setWallet]=useState<Wallet|null>(null),[error,setError]=useState(""),[busy,setBusy]=useState(false),[filter,setFilter]=useState(""),[limit,setLimit]=useState(20);
 const generation=useRef(0);
 const load=useCallback(async()=>{const current=++generation.current;setBusy(true);setError("");try{const result=await call<Wallet>("/storefront/wallet",token);if(current===generation.current)setWallet(result)}catch(e){if(current===generation.current)setError(e instanceof Error?e.message:"دریافت تراکنش‌ها انجام نشد.")}finally{if(current===generation.current)setBusy(false)}},[token]);
 useEffect(()=>{setWallet(null);void load();return()=>{generation.current++}},[load]);
 const entries=(wallet?.entries??[]).slice().reverse().filter(entry=>!filter||groups[entry.kind]===filter);
 const pendingDeposits=wallet?.deposits.filter(item=>item.status!=="approved")??[];
 const pendingWithdrawals=wallet?.refunds.filter(item=>item.status!=="completed")??[];
 return <section className="buy-wallet-transactions" aria-label="تراکنش‌های کیف پول" aria-busy={busy}>
  <div className="buy-row"><a className="buy-secondary" href="/wallet">بازگشت به کیف پول</a><button className="buy-secondary" disabled={busy} onClick={()=>void load()}>به‌روزرسانی تراکنش‌ها</button></div>
  {error&&<p className="buy-alert" role="alert">{error}</p>}{busy&&!wallet&&<p role="status">دریافت تراکنش‌ها…</p>}
  {wallet&&<><section className="buy-panel"><h2>تراکنش‌های قطعی</h2><p>واریز و شارژ، پرداخت سفارش، بازگشت وجه و برداشت بانکی ثبت‌شده در کیف پول شما.</p>
   <label className="buy-field">نوع تراکنش<select value={filter} onChange={e=>{setFilter(e.target.value);setLimit(20)}}><option value="">همه تراکنش‌ها</option><option value="funding">واریز و شارژ</option><option value="payment">پرداخت سفارش</option><option value="return">بازگشت وجه سفارش</option><option value="withdrawal">برداشت بانکی</option></select></label>
   {!entries.length?<p>تراکنشی با این شرایط ثبت نشده است.</p>:<ol className="wallet-transaction-list">{entries.slice(0,limit).map(entry=><WalletTransaction key={entry.id} entry={entry}/>)}</ol>}
   <p className="buy-note">{Math.min(limit,entries.length).toLocaleString("fa-IR")} از {entries.length.toLocaleString("fa-IR")} تراکنش</p>
   {entries.length>limit&&<button className="buy-secondary" onClick={()=>setLimit(value=>value+20)}>مشاهده تراکنش‌های قدیمی‌تر</button>}
  </section><section className="buy-panel"><h2>واریزها و برداشت‌های نهایی‌نشده</h2><p>این موارد هنوز تراکنش قطعی کیف پول نیستند؛ درخواست ردشده نیز موجودی را کم یا زیاد نمی‌کند.</p>
   {!pendingDeposits.length&&!pendingWithdrawals.length?<p>درخواست نهایی‌نشده‌ای ندارید.</p>:<ul className="wallet-request-history">
    {pendingDeposits.map(item=><li key={"deposit:"+item.id}><strong>اعلام واریز · {label(item.status)}</strong><span>{amount(item.amount_toman)}</span><time dateTime={item.created_at}>{date(item.created_at)}</time></li>)}
    {pendingWithdrawals.map(item=><li key={"withdrawal:"+item.id}><strong>درخواست برداشت بانکی · {label(item.status)}</strong><span>{amount(item.amount_toman)}</span><time dateTime={item.created_at}>{date(item.created_at)}</time></li>)}
   </ul>}
  </section></>}
 </section>;
}
function WalletTransaction({entry}:{entry:Entry}){
 const negative=entry.amount_toman.startsWith("-");
 const title=entry.kind==="manual-refund"?"برداشت / بازپرداخت بانکی":label(entry.kind);
 const order=entry.kind==="order-payment"&&entry.source.startsWith("order:")?entry.source.slice(6):"";
 return <li className="wallet-transaction" data-direction={negative?"debit":"credit"}>
  <div><strong>{title}</strong><time dateTime={entry.at}>{date(entry.at)}</time><small>{negative?"کاهش موجودی":"افزایش موجودی"} · ثبت قطعی</small></div>
  <bdi className="wallet-transaction-amount">{amount(entry.amount_toman)}</bdi>
  {order&&<a className="buy-secondary" href={"/account/orders/"+encodeURIComponent(order)}>مشاهده سفارش</a>}
 </li>;
}
