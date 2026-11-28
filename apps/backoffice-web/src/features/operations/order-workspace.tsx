import { useRef, useState } from "react";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { Alert, Button, Stack, TextField } from "@mui/material";
import { Link, useParams, useSearchParams } from "react-router-dom";
import { apiRequest } from "../../shared/api";
import { useWorkspace } from "../../shared/workspace";
import { DataTable, PageHeader } from "../../shared/components";
import { FilterPanel } from "../../shared/filter-panel";
import { ChoiceField } from "../../shared/option-select";
import { EditorTabs } from "../../shared/editor-tabs";
import { dateTime } from "../../shared/format";
import { label, money, type Purchase } from "../../../../../packages/commerce-client/types";
import { OrderOperationsPanel } from "./order-operations-panel";
import { PurchaseRequests, PurchaseHistory } from "./order-requests-panel";
import { PurchaseReview } from "../commerce/commerce-page";
import { OperationsPage } from "./operations-page";
import { SupplyOverduePanel } from "./supply-overdue-panel";

interface Row {id:string;number:string;customer:string;stores:string[];total_toman:string;order_status:string;payment_status:string;fulfillment_status:string;groups:number;legacy:boolean}
interface Page {items:Row[];stores:{id:string;name:string}[];total:number;page:number;size:number}
interface Detail {purchase?:Purchase & {customer_name?:string;customer_phone?:string};legacy:boolean;permissions:Record<string,boolean>}
const choices=(values:string[])=>values.map(id=>({id,name:label(id)}));
const fa=(value:string)=>value.replace(/\d/g,d=>"۰۱۲۳۴۵۶۷۸۹"[Number(d)]);

export function OrderWorkspaceList(){
 const {authToken,can}=useWorkspace();const [params,setParams]=useSearchParams();
 const [draft,setDraft]=useState({q:params.get("q")??"",mobile:params.get("mobile")??"",store:params.get("store")??"",order:params.get("order")??"",payment:params.get("payment")??"",fulfillment:params.get("fulfillment")??""});
 const page=Math.max(0,Number(params.get("page"))||0),size=[10,20,50,100].includes(Number(params.get("size")))?Number(params.get("size")):20;
 const query=new URLSearchParams(params);query.set("page",String(page));query.set("size",String(size));
 const url="/backoffice/order-workspace?"+query;
 const data=useQuery({queryKey:["order-workspace-list",authToken,url],queryFn:()=>apiRequest<Page>(url,authToken)});
 const update=(changes:Record<string,string>)=>{const next=new URLSearchParams(params);Object.entries(changes).forEach(([key,value])=>value?next.set(key,value):next.delete(key));setParams(next);};
 return <><PageHeader title="سفارش‌ها" description="هر خرید، مرسوله‌های مستقل فروشگاه‌ها و سه وضعیت سفارش، پرداخت و عملیات دارد."/>
  <Stack direction="row" spacing={1} mb={2}><Button variant="outlined" onClick={()=>void data.refetch()}>به‌روزرسانی فهرست</Button>{can("commerce.deposits.read")&&<Button component={Link} to="/wallet-operations">واریز و برداشت کیف پول</Button>}</Stack>
  <details className="panel"><summary>صف مهلت‌های تأمین</summary><SupplyOverduePanel/></details>
  <FilterPanel onSearch={()=>update({...draft,page:"0"})}><div className="admin-filter-grid">
   <TextField label="شماره سفارش یا نام مشتری" value={draft.q} onChange={e=>setDraft({...draft,q:e.target.value})}/>
   <TextField label="موبایل مشتری" placeholder="شماره کامل موبایل" value={draft.mobile} inputProps={{dir:"ltr",inputMode:"tel",maxLength:30}} onChange={e=>setDraft({...draft,mobile:e.target.value})}/>
   <ChoiceField label="فروشگاه" value={draft.store} clearLabel="همه فروشگاه‌ها" options={data.data?.stores??[]} onChange={store=>setDraft({...draft,store})}/>
   <ChoiceField label="وضعیت سفارش" value={draft.order} clearLabel="همه وضعیت‌ها" options={choices(["placed","processing","completed","cancelled"])} onChange={order=>setDraft({...draft,order})}/>
   <ChoiceField label="وضعیت پرداخت" value={draft.payment} clearLabel="همه وضعیت‌ها" options={choices(["unpaid","paid","returned-to-wallet","authorized","failed"])} onChange={payment=>setDraft({...draft,payment})}/>
   <ChoiceField label="وضعیت عملیات" value={draft.fulfillment} clearLabel="همه مراحل" options={choices(["awaiting-payment","awaiting-supply","supplied","packing","shipped","delivered","cancelled","mixed","on-hold"])} onChange={fulfillment=>setDraft({...draft,fulfillment})}/>
  </div></FilterPanel>
  {data.error&&<Alert severity="error" action={<Button onClick={()=>void data.refetch()}>تلاش مجدد</Button>}>{data.error.message}</Alert>}
  {data.isFetching&&<p role="status">دریافت فهرست سفارش‌ها…</p>}
  <DataTable rows={data.data?.items??[]} searchable={false} columns={[
   {title:"شماره سفارش",render:r=><Link to={`/orders/${encodeURIComponent(r.id)}`}>{fa(r.number)}{r.legacy?" · سابقه قدیمی":""}</Link>},
   {title:"مشتری",render:r=>r.customer},{title:"فروشگاه / مرسوله",render:r=>r.stores.join("، ")+" · "+r.groups.toLocaleString("fa-IR")},
   {title:"مبلغ",render:r=>money(r.total_toman)},{title:"سفارش",render:r=>label(r.order_status)},{title:"پرداخت",render:r=>label(r.payment_status)},{title:"عملیات",render:r=>label(r.fulfillment_status)},
   {title:"مشاهده",render:r=><Button component={Link} variant="outlined" size="small" to={`/orders/${encodeURIComponent(r.id)}`}>مشاهده</Button>},
  ]} remote={{filterKey:url,controls:{page,pages:Math.max(1,Math.ceil((data.data?.total??0)/size)),size,total:data.data?.total??0,setPage:n=>update({page:String(n)}),setSize:n=>update({size:String(n),page:"0"})},exportRows:async(ids,signal)=>{let all:Row[]=[];let index=0;while(true){if(signal?.aborted)throw new Error("دریافت خروجی لغو شد.");const next=new URLSearchParams(query);next.set("page",String(index++));next.set("size","100");const result=await apiRequest<Page>("/backoffice/order-workspace?"+next,authToken);all.push(...result.items);if(all.length>=result.total||!result.items.length)break;}return ids?all.filter(row=>ids.includes(row.id)):all;}}}/>
 </>;
}

export function OrderWorkspaceDetail(){
 const {id=""}=useParams();const {authToken}=useWorkspace();const cache=useQueryClient();const [params,setParams]=useSearchParams();
 const [busy,setBusy]=useState(false),[error,setError]=useState("");const running=useRef(false),pending=useRef<{signature:string;key:string}>();
 const query=useQuery({queryKey:["order-workspace-detail",authToken,id],queryFn:()=>apiRequest<Detail>(`/backoffice/order-workspace/${encodeURIComponent(id)}`,authToken)});
 const p=query.data?.purchase;
 async function refresh(){await query.refetch();await cache.invalidateQueries({queryKey:["order-workspace-list"]});await cache.invalidateQueries({queryKey:["workspace"]});}
 async function submit(cancel:boolean,reason:string){if(!p||running.current)return;const input={id:p.id,version:p.version,reason};const signature=JSON.stringify({cancel,...input});if(pending.current?.signature!==signature)pending.current={signature,key:crypto.randomUUID()};running.current=true;setBusy(true);setError("");try{await apiRequest(`/backoffice/commerce/orders/${cancel?"cancel":"confirm"}`,authToken,input,pending.current.key);pending.current=undefined;await refresh();}catch(e){setError(e instanceof Error?e.message:"عملیات انجام نشد.");}finally{running.current=false;setBusy(false);}}
 if(query.isPending)return <p role="status">دریافت جزئیات سفارش…</p>;
 if(query.error)return <Alert severity="error" action={<Button onClick={()=>void query.refetch()}>تلاش مجدد</Button>}>{query.error.message}</Alert>;
 if(query.data?.legacy)return <><Alert severity="info">سابقهٔ قدیمی؛ این رکورد خرید جدید یا تلاش پرداخت ساختگی ندارد.</Alert><OperationsPage section="orders"/></>;
 if(!p)return <Alert severity="warning">سفارش پیدا نشد.</Alert>;
 const chosen=params.get("tab")??(id!==p.id&&id!==p.number?"operations":"summary");
 const tab=["summary","payment","operations","requests","history"].includes(chosen)?chosen:"summary";
 return <><PageHeader title={`سفارش ${fa(p.number)}`} description={`${p.customer_name??"مشتری"} · ${dateTime(p.created_at)}`}/><Stack direction="row" spacing={1} mb={2}><Button component={Link} to="/orders">بازگشت به سفارش‌ها</Button><Button onClick={()=>void refresh()}>به‌روزرسانی وضعیت</Button></Stack>{error&&<Alert severity="error">{error}</Alert>}
 <div className="order-stage-summary"><span>سفارش: <b>{label(p.order_status)}</b></span><span>پرداخت: <b>{label(p.payment_status)}</b></span><span>عملیات: <b>{label(p.fulfillment_status??"")}</b></span></div>
 <EditorTabs label="جزئیات سفارش" value={tab} onChange={value=>{const next=new URLSearchParams(params);next.set("tab",value);setParams(next);}} panels={[
  {id:"summary",label:"خلاصه و اقلام",content:<section className="panel"><h2>اقلام سفارش</h2>{p.quote.lines.map(line=><article key={line.offer_id} className="commerce-order-item"><img src={(import.meta.env.VITE_API_BASE_URL??"")+line.image_url} alt={line.name}/><div><b>{line.name}</b><p>{line.quantity.toLocaleString("fa-IR")} عدد · هر عدد {money(line.unit_price_toman)}</p></div><strong>{money(line.total_toman)}</strong></article>)}<p>مبلغ خرید: {money(p.quote.total_toman)}</p><h3>گیرنده</h3><p>{p.quote.address.recipient} · {p.quote.address.phone}</p><p>{p.quote.address.full_address||p.quote.address.street}</p><h3>مرسوله‌های فروشگاه‌ها</h3>{p.quote.shipments.map(shipment=><p key={shipment.store_id}>{shipment.store_name} · {shipment.method_name}</p>)}</section>},
  {id:"payment",label:"پرداخت‌ها",content:<PurchaseReview mode="payment" purchase={p} token={authToken} canReadDeposits={!!query.data?.permissions["commerce.deposits.read"]} allowed={!!query.data?.permissions["commerce.orders.confirm"]} busy={busy} submit={(cancel,reason)=>void submit(cancel,reason)}/>},
  {id:"operations",label:"عملیات و مرسوله‌ها",content:<Stack spacing={2}>{p.order_ids.map(order=><OrderOperationsPanel key={order} id={order} onChanged={refresh}/>)}</Stack>},
  {id:"requests",label:"درخواست‌ها و استرداد",content:<><PurchaseRequests id={p.id} onChanged={refresh}/>{query.data?.permissions["commerce.refunds.resolve"]&&<Button component={Link} to="/wallet-operations">رسیدگی انتقال‌های بانکی در صف مالی</Button>}</>},
  {id:"history",label:"تاریخچه",content:<PurchaseHistory id={p.id} version={p.version}/>},
 ]}/></>;
}
