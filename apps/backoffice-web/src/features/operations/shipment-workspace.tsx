import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { Alert, Button, Stack, TextField } from "@mui/material";
import { Link, useParams, useSearchParams } from "react-router-dom";
import { apiRequest } from "../../shared/api";
import { useWorkspace } from "../../shared/workspace";
import { DataTable, PageHeader } from "../../shared/components";
import { FilterPanel } from "../../shared/filter-panel";
import { ChoiceField } from "../../shared/option-select";
import { dateTime } from "../../shared/format";
import { label } from "../../../../../packages/commerce-client/types";
import { OrderOperationsPanel } from "./order-operations-panel";

interface Row {id:string;number:string;purchase_id:string;purchase_number:string;store_id:string;store_name:string;status:string;order_status:string;payment_status:string;shipping_name:string;quantity:number;started_at:string}
interface Page {items:Row[];stores:{id:string;name:string}[];total:number;page:number;size:number}
const fa=(value:string)=>value.replace(/\d/g,d=>"۰۱۲۳۴۵۶۷۸۹"[Number(d)]);
export function ShipmentWorkspace(){
 const {authToken,surface}=useWorkspace();const [params,setParams]=useSearchParams();
 const [draft,setDraft]=useState({q:params.get("q")??"",store:params.get("store")??"",status:params.get("status")??""});
 const page=Math.max(0,Number(params.get("page"))||0),size=[10,20,50,100].includes(Number(params.get("size")))?Number(params.get("size")):20;
 const query=new URLSearchParams(params);query.set("page",String(page));query.set("size",String(size));
 const url=`/${surface}/shipments?${query}`;
 const data=useQuery({queryKey:["shipments",authToken,url],queryFn:()=>apiRequest<Page>(url,authToken)});
 const update=(changes:Record<string,string>)=>{const next=new URLSearchParams(params);Object.entries(changes).forEach(([key,value])=>value?next.set(key,value):next.delete(key));setParams(next)};
 return <><PageHeader title="مرسوله‌ها" description="مرسوله‌های پرداخت‌شده‌ای که پردازش آن‌ها آغاز شده است؛ تحویل‌شده و لغوشده در فیلتر تاریخچه قابل مشاهده‌اند."/>
 <Button variant="outlined" onClick={()=>void data.refetch()}>به‌روزرسانی فهرست</Button>
 <FilterPanel onSearch={()=>update({...draft,page:"0"})}><div className="admin-filter-grid">
  <TextField label="شماره مرسوله یا سفارش" value={draft.q} onChange={e=>setDraft({...draft,q:e.target.value})}/>
  <ChoiceField label="فروشگاه" value={draft.store} clearLabel="همه فروشگاه‌های مجاز" options={data.data?.stores??[]} onChange={store=>setDraft({...draft,store})}/>
  <ChoiceField label="وضعیت مرسوله" value={draft.status} clearLabel="مرسوله‌های فعال" options={[{id:"all",name:"همه؛ همراه تاریخچه"},...["awaiting-supply","supplied","packing","shipped","delivered","cancelled","on-hold"].map(id=>({id,name:label(id)}))]} onChange={status=>setDraft({...draft,status})}/>
 </div></FilterPanel>
 {data.error&&<Alert severity="error" action={<Button onClick={()=>void data.refetch()}>تلاش مجدد</Button>}>{data.error.message}</Alert>}
 {data.isFetching&&<p role="status">دریافت مرسوله‌ها…</p>}
 <DataTable searchable={false} rows={data.data?.items??[]} columns={[
  {title:"مرسوله",render:r=>fa(r.number)},{title:"سفارش",render:r=>fa(r.purchase_number)},{title:"فروشگاه",render:r=>r.store_name},{title:"روش ارسال",render:r=>r.shipping_name},
  {title:"تعداد اقلام",render:r=>r.quantity.toLocaleString("fa-IR")},{title:"وضعیت عملیات",render:r=>label(r.status)},{title:"شروع پردازش",render:r=>dateTime(r.started_at)},
  {title:"مشاهده و عملیات",render:r=><Button component={Link} variant="outlined" to={`/shipments/${encodeURIComponent(r.id)}`}>مشاهده</Button>}
 ]} remote={{filterKey:url,controls:{page,pages:Math.max(1,Math.ceil((data.data?.total??0)/size)),size,total:data.data?.total??0,setPage:n=>update({page:String(n)}),setSize:n=>update({size:String(n),page:"0"})},exportRows:async(ids,signal)=>{let rows:Row[]=[];let index=0;while(true){if(signal?.aborted)throw new Error("خروجی لغو شد.");const next=new URLSearchParams(query);next.set("page",String(index++));next.set("size","100");const result=await apiRequest<Page>(`/${surface}/shipments?${next}`,authToken);rows.push(...result.items);if(rows.length>=result.total||!result.items.length)break}return ids?rows.filter(r=>ids.includes(r.id)):rows}}}/></>;
}
export function ShipmentWorkspaceDetail(){
 const {id=""}=useParams();const {authToken,surface}=useWorkspace();
 const query=useQuery({queryKey:["shipment-detail",authToken,surface,id],queryFn:()=>apiRequest<{number:string;operations:{purchase_id:string;purchase_number:string}}>(`/${surface}/shipments/${encodeURIComponent(id)}`,authToken)});
 if(query.isPending)return <p role="status">دریافت مرسوله…</p>;
 if(query.error)return <Alert severity="error" action={<Button onClick={()=>void query.refetch()}>تلاش مجدد</Button>}>{query.error.message}</Alert>;
 return <><PageHeader title={`مرسوله ${fa(query.data.number)}`} description={`مربوط به سفارش ${fa(query.data.operations.purchase_number)}`}/>
 <Stack direction="row" spacing={1}><Button component={Link} to="/shipments">بازگشت به مرسوله‌ها</Button><Button component={Link} to={`/orders/${encodeURIComponent(surface==="seller"?id:query.data.operations.purchase_id)}`}>مشاهده سفارش</Button></Stack>
 <OrderOperationsPanel id={id} onChanged={async()=>{await query.refetch()}}/></>;
}
