import { useCallback, useEffect, useRef, useState, type FormEvent } from "react";
import type { OrderRequest, OrderRequests } from "./types";

type Transport = <T>(path:string,body?:unknown,key?:string) => Promise<T>;
const names:Record<string,string> = {cancel:"کنسلی پیش از ارسال",return:"مرجوعی",requested:"در انتظار بررسی",approved:"تأیید شده",rejected:"رد شده",approve:"تأیید درخواست",reject:"رد درخواست",receive:"ثبت دریافت فیزیکی کالا",inspect:"ثبت نتیجه بازرسی"};
const fa = (value:string|number) => String(value).replace(/\d/g,d=>"۰۱۲۳۴۵۶۷۸۹"[Number(d)]);

export function OrderRequestsPanel({id,prefix,request,onChanged}:{id:string;prefix:string;request:Transport;onChanged?:()=>void|Promise<void>}) {
  const [data,setData]=useState<OrderRequests>();
  const [error,setError]=useState("");
  const [busy,setBusy]=useState(false);
  const [kind,setKind]=useState("cancel");
  const [reason,setReason]=useState("");
  const [quantities,setQuantities]=useState<Record<string,string>>({});
  const running=useRef(false), keys=useRef(new Map<string,string>());
  const load=useCallback(async()=>setData(await request<OrderRequests>(`${prefix}/${encodeURIComponent(id)}/requests`)),[request,prefix,id]);
  useEffect(()=>{void load().catch(e=>setError(e.message));},[load]);
  async function mutate(path:string,body:unknown,done?:()=>void) {
    if(running.current)return;
    const signature=path+JSON.stringify(body);
    if(!keys.current.has(signature))keys.current.set(signature,crypto.randomUUID());
    running.current=true;setBusy(true);setError("");
    try{await request(path,body,keys.current.get(signature));keys.current.delete(signature);done?.();await load();await onChanged?.();}
    catch(e){setError(e instanceof Error?e.message:"ثبت انجام نشد؛ دوباره تلاش کنید.");await load().catch(()=>{});}
    finally{running.current=false;setBusy(false);}
  }
  function submit(e:FormEvent) {
    e.preventDefault();if(!data)return;
    const items=data.items.map(item=>({order_id:item.order_id,item_id:item.item_id,quantity:Number(quantities[item.item_id]??0)})).filter(item=>item.quantity>0);
    if(!items.length){setError("حداقل یک قلم و تعداد را انتخاب کنید.");return;}
    void mutate(`${prefix}/${encodeURIComponent(id)}/requests`,{version:data.version,kind,reason,items},()=>{setReason("");setQuantities({});});
  }
  return <section className="order-requests-panel" aria-label="درخواست‌های کنسلی و مرجوعی">
    <h3>درخواست‌های کنسلی و مرجوعی</h3>
    <p>ثبت درخواست وجه یا موجودی را برنمی‌گرداند. رسیدگی، دریافت کالا و استرداد مالی جداگانه انجام می‌شوند.</p>
    {error&&<p role="alert">{error}</p>}
    {!data?<button type="button" onClick={()=>void load().catch(e=>setError(e.message))}>دریافت درخواست‌ها</button>:<>
      {data.legacy_requested&&<p role="status">درخواست قدیمی بدون جزئیات اقلام: {data.legacy_reason}؛ نیازمند بررسی پشتیبانی.</p>}
      {data.can_create&&<details><summary>ثبت درخواست جدید</summary><form onSubmit={submit}>
        <fieldset disabled={busy}>
          <label>نوع درخواست<select value={kind} onChange={e=>{setKind(e.target.value);setQuantities({});}}><option value="cancel">کنسلی پیش از ارسال</option><option value="return">مرجوعی حین حمل یا پس از تحویل</option></select></label>
          {data.items.filter(item=>(kind==="cancel"?item.available_cancel:item.available_return)>0).map(item=><label key={item.item_id}>{item.name} · {item.store_name} · حداکثر {fa(kind==="cancel"?item.available_cancel:item.available_return)} عدد<input type="number" inputMode="numeric" min={0} max={kind==="cancel"?item.available_cancel:item.available_return} step={1} value={quantities[item.item_id]??"0"} onChange={e=>setQuantities({...quantities,[item.item_id]:e.target.value})}/></label>)}
          <label>دلیل درخواست<textarea required maxLength={1000} value={reason} onChange={e=>setReason(e.target.value)}/></label>
          <button className="buy-primary" type="submit">ثبت درخواست برای بررسی</button>
        </fieldset>
      </form></details>}
      {!data.requests.length&&<p>پرونده‌ای ثبت نشده است.</p>}
      {data.requests.map(row=><article key={row.id} className="order-request-card">
        <h4>درخواست {fa(row.number)} · {names[row.kind]} · {names[row.status]}</h4>
        <p>{new Date(row.created_at).toLocaleString("fa-IR",{timeZone:"Asia/Tehran"})}</p>
        {row.items.map(item=><p key={item.item_id}>{item.name} · {item.store_name} · {fa(item.quantity)} عدد</p>)}
        {row.reason&&<p>دلیل: {row.reason}</p>}{row.review_reason&&<p>نتیجه رسیدگی: {row.review_reason}</p>}
        {row.kind==="return"&&<p>دریافت کالا: {row.goods_received?"ثبت شده":"ثبت نشده"} · بازرسی: {row.inspected?(row.healthy?"سالم":"ناسالم"):"انجام نشده"} · برگشت موجودی: {row.stock_applied?"انجام شده":"انجام نشده"}</p>}
        {!!row.allowed_actions.length&&<Review key={row.id+row.version} row={row} busy={busy} submit={(body)=>void mutate(`/backoffice/order-requests/${encodeURIComponent(row.id)}/review`,body)}/>}
      </article>)}
    </>}
  </section>;
}

function Review({row,busy,submit}:{row:OrderRequest;busy:boolean;submit:(body:unknown)=>void}) {
  const [action,setAction]=useState(row.allowed_actions[0]);
  const [reason,setReason]=useState("");const [health,setHealth]=useState("");
  return <form onSubmit={e=>{e.preventDefault();submit({version:row.version,action,reason,...(health?{healthy:health==="healthy"}:{})});}}><fieldset disabled={busy}>
    <label>رسیدگی<select value={action} onChange={e=>setAction(e.target.value)}>{row.allowed_actions.map(a=><option value={a} key={a}>{names[a]}</option>)}</select></label>
    {["receive","inspect"].includes(action)&&<label>نتیجه سلامت<select required={action==="inspect"} value={health} onChange={e=>setHealth(e.target.value)}><option value="">بازرسی بعداً انجام می‌شود</option><option value="healthy">سالم؛ قابل برگشت به موجودی</option><option value="damaged">ناسالم؛ بدون برگشت به موجودی</option></select></label>}
    <label>دلیل / نتیجه رسیدگی<textarea required maxLength={1000} value={reason} onChange={e=>setReason(e.target.value)}/></label><button type="submit">{names[action]}</button>
  </fieldset></form>;
}
