import { useEffect, useRef, useState } from "react";
import { label, type OrderEvent } from "./types";

interface HistoryEvent extends OrderEvent {sequence:number;store_name?:string;request_number?:string;actor_name?:string;reason?:string}
interface HistoryPage {events:HistoryEvent[];next_before?:number}
export function OrderHistoryPanel({path,request,version}:{path:string;request:<T>(path:string)=>Promise<T>;version:number}) {
 const [events,setEvents]=useState<HistoryEvent[]>([]),[before,setBefore]=useState<number>(),[busy,setBusy]=useState(false),[error,setError]=useState("");
 const generation=useRef(0),running=useRef(false);
 useEffect(()=>{const current=++generation.current;running.current=true;setBusy(true);setError("");setEvents([]);setBefore(undefined);
  void request<HistoryPage>(path+"?limit=10").then(data=>{if(current===generation.current){setEvents(data.events);setBefore(data.next_before);}}).catch(e=>{if(current===generation.current)setError(e.message);}).finally(()=>{if(current===generation.current){running.current=false;setBusy(false);}});
  return ()=>{generation.current++;};
 },[path,request,version]);
 async function more(){if(running.current)return;const current=generation.current;running.current=true;setBusy(true);setError("");try{const data=await request<HistoryPage>(path+"?limit=10"+(before?"&before="+before:""));if(current===generation.current){setEvents(previous=>[...previous,...data.events.filter(e=>!previous.some(old=>old.id===e.id))]);setBefore(data.next_before);}}catch(e){if(current===generation.current)setError(e instanceof Error?e.message:"دریافت تاریخچه انجام نشد.");}finally{if(current===generation.current){running.current=false;setBusy(false);}}}
 return <details className="order-requests-panel"><summary>تاریخچه سفارش، پرداخت و مرسوله‌ها</summary><section aria-label="تاریخچه مجاز سفارش" aria-busy={busy}>
  {!events.length&&!busy&&!error&&<p>رویدادی ثبت نشده است.</p>}
  <ol>{events.map(event=><li key={event.id}><b>{label(event.action)}</b> · <time>{new Date(event.at).toLocaleString("fa-IR",{timeZone:"Asia/Tehran"})}</time>{event.store_name&&<p>فروشگاه: {event.store_name}</p>}{event.request_number&&<p>درخواست {event.request_number.replace(/\d/g,d=>"۰۱۲۳۴۵۶۷۸۹"[Number(d)])}</p>}<p>سفارش: {label(event.order_status)} · پرداخت: {label(event.payment_status)}{event.to?" · عملیات: "+label(event.to):""}</p>{event.actor_name&&<p>ثبت‌کننده: {event.actor_name}</p>}{event.reason&&<p>علت ثبت: {event.reason}</p>}</li>)}</ol>
  {busy&&<p role="status">دریافت تاریخچه…</p>}{error&&<p role="alert">{error}</p>}{(before||error)&&<button type="button" disabled={busy} onClick={()=>void more()}>{error?"تلاش مجدد":"مشاهده رویدادهای قدیمی‌تر"}</button>}
 </section></details>;
}
