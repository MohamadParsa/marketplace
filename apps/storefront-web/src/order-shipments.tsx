import { useCallback, useEffect, useRef, useState } from "react";
import { call, useCommerce } from "./commerce";
import { label, type OrderOperations, type Quote, type Shipment as ShipmentSnapshot } from "../../../packages/commerce-client/types";
import { OrderItems } from "./order-items";
import { dispatchTime } from "./dispatch-time";
import { deliveryPromise, shippingCharge } from "../../../packages/commerce-client/shipping-presentation";
import { OrderHistoryPanel } from "../../../packages/commerce-client/order-history-panel";
import { shipmentProgress } from "../../../packages/commerce-client/order-progress";
import { OrderProgress } from "../../../packages/commerce-client/order-progress-view";
import { orderStageHelp } from "../../../packages/commerce-client/order-stage-help";
import "../../../packages/commerce-client/order-requests.css";

export function CustomerOrderHistory({id,version}:{id:string;version:number}) {
 const {token}=useCommerce();const request=useCallback(<T,>(path:string)=>call<T>(path,token),[token]);
 return <OrderHistoryPanel path={`/storefront/account/orders/${encodeURIComponent(id)}/history`} request={request} version={version}/>;
}
export function CustomerShipments({id,orders,quote,version,onChanged}:{id:string;orders:string[];quote:Quote;version:number;onChanged:()=>Promise<void>}) {
 const {token}=useCommerce();const [rows,setRows]=useState<OrderOperations[]|null>(null),[error,setError]=useState("");
 const generation=useRef(0);const ids=JSON.stringify(orders);
 const load=useCallback(async()=>{
  const current=++generation.current;
  try {
   const result=await Promise.all((JSON.parse(ids) as string[]).map(order=>call<OrderOperations>(`/storefront/account/orders/${encodeURIComponent(id)}/shipments/${encodeURIComponent(order)}`,token)));
   if(current===generation.current){setRows(result);setError("")}
  } catch(e){if(current===generation.current)setError(e instanceof Error?e.message:"دریافت مرسوله‌ها انجام نشد.")}
 },[id,ids,token]);
 useEffect(()=>{setRows(null);setError("");void load();return()=>{generation.current++}},[load,version]);
 return <section className="order-requests-panel" aria-label="مرسوله‌های سفارش"><h3>مرسوله‌ها و وضعیت تحویل</h3>
  {error&&<p role="alert">{error}<button type="button" onClick={()=>void load()}>تلاش مجدد</button></p>}
  {rows===null&&!error&&<p role="status">دریافت مرسوله‌ها…</p>}
  {rows?.length===0&&<p>مرسوله‌ای برای این سفارش ثبت نشده است.</p>}
  {rows?.map(row=><Shipment key={row.id} row={row} lines={quote.lines.filter(line=>line.store_id===row.store_id)} snapshot={quote.shipments.find(shipment=>shipment.store_id===row.store_id)} token={token} parent={id} onChanged={async()=>{await load();await onChanged()}}/>)}
 </section>;
}
function Shipment({row,lines,snapshot,token,parent,onChanged}:{row:OrderOperations;lines:Quote["lines"];snapshot?:ShipmentSnapshot;token:string;parent:string;onChanged:()=>Promise<void>}) {
 const [checked,setChecked]=useState(false),[busy,setBusy]=useState(false),[error,setError]=useState("");
 const running=useRef(false),pending=useRef<{version:number;key:string}>();
 async function receive(){if(running.current||!checked)return;if(pending.current?.version!==row.version)pending.current={version:row.version,key:crypto.randomUUID()};running.current=true;setBusy(true);setError("");try{await call(`/storefront/account/orders/${encodeURIComponent(parent)}/shipments/${encodeURIComponent(row.id)}/receive`,token,{version:row.version,all_items:true},undefined,pending.current.key);pending.current=undefined;setChecked(false);await onChanged()}catch(e){setError(e instanceof Error?e.message:"ثبت دریافت انجام نشد.")}finally{running.current=false;setBusy(false)}}
 const fa=(value:number)=>value.toLocaleString("fa-IR");
 return <article className="order-request-card customer-shipment"><h4>{row.store_name} · {row.shipping_name}</h4>
  {snapshot&&<section className="buy-shipment customer-shipment-promise" aria-label="وعده ارسال این مرسوله">
   <p>{shippingCharge(snapshot.collection_mode,snapshot.fee_toman)}</p>
   <p>{deliveryPromise(snapshot.delivery,snapshot.delivery_from,snapshot.delivery_to)}</p>
   <small>{dispatchTime(snapshot.lead_days,snapshot.lead_day_basis)}</small>
  </section>}
  <p>سفارش: {label(row.order_status)} · پرداخت: {label(row.payment_status)} · عملیات: <strong>{label(row.status)}</strong></p>
  <OrderProgress steps={shipmentProgress(row.status,row.events)} label={`مراحل مرسوله ${row.store_name}`}/>
  {orderStageHelp[row.status]&&<p className={['cancelled','on-hold','supply-exception','refund-pending'].includes(row.status)?'order-progress-exception':'buy-note'}>{orderStageHelp[row.status]}</p>}
  <h5>اقلام این مرسوله</h5>
  {lines.length?<OrderItems lines={lines}/>:<ul>{row.items.map(item=><li key={item.id}>{item.name} · {fa(item.quantity)} عدد</li>)}</ul>}
  {row.items.some(item=>item.cancelled_quantity||item.returned_quantity)&&<ul>{row.items.filter(item=>item.cancelled_quantity||item.returned_quantity).map(item=><li key={item.id}>{item.name}{item.cancelled_quantity?" · لغو شده: "+fa(item.cancelled_quantity):""}{item.returned_quantity?" · برگشتی سالم: "+fa(item.returned_quantity):""}</li>)}</ul>}
  {!!row.identifiers?.length&&<p>{row.dispatch_identifier==="delivery"?"کد تحویل":"کد رهگیری"}: <span dir="ltr">{row.identifiers.join("، ")}</span></p>}
  <details className="customer-shipment-history"><summary>تاریخچه این مرسوله</summary>
   {row.events.length?<ol className="order-event-timeline">{row.events.map((event,index)=><li key={event.id??index}>
    <span className="order-event-marker" aria-hidden="true"/><div><strong>{label(event.action)}</strong>
    <time dateTime={event.at}>{new Date(event.at).toLocaleString("fa-IR",{timeZone:"Asia/Tehran"})}</time>
    {event.customer_description&&<p>{event.customer_description}</p>}</div>
   </li>)}</ol>:<p>تاریخچه مرحله‌ای ثبت نشده است؛ وضعیت فعلی در بالا نمایش داده می‌شود.</p>}
  </details>
  {row.allowed_actions.includes("deliver")&&<form onSubmit={e=>{e.preventDefault();void receive()}}><fieldset disabled={busy}>
   <label className="shipment-receive-choice"><input type="checkbox" checked={checked} onChange={e=>setChecked(e.target.checked)}/>تمام اقلام این مرسوله را دریافت کرده‌ام.</label>
   <button type="submit" disabled={!checked}>تأیید دریافت این مرسوله</button></fieldset></form>}
  {error&&<p role="alert">{error}</p>}
 </article>;
}
