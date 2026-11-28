import { useEffect, useRef, useState } from "react";
import { Alert, Button, MenuItem, TextField } from "@mui/material";
import { useWorkspace } from "../../shared/workspace";
import { FilterPanel } from "../../shared/filter-panel";
import { number } from "../../shared/format";
import { PersianDateField } from "../../../../../packages/date-controls/persian-date-field";

type Query={from:string;to:string;device:string;category_id:string;environment:string};
type Funnel={environment:string;matched_events:number;unlinked_events:number;unique_paid_purchases:number;stages:{key:string;label:string;events:number;known_identities:number}[]};
const empty:Query={from:"",to:"",device:"",category_id:"",environment:""};
const environments:Record<string,string>={test:"آزمایشی",local:"محلی",staging:"پیش‌انتشار",production:"عملیاتی"};
export function ActivityFunnelPanel(){
  const {authToken,state}=useWorkspace();const [draft,setDraft]=useState(empty),[query,setQuery]=useState(empty),[data,setData]=useState<Funnel|null>(null),[error,setError]=useState(""),[busy,setBusy]=useState(false),[refresh,setRefresh]=useState(0);const generation=useRef(0);
  useEffect(()=>{const current=++generation.current;const controller=new AbortController();setBusy(true);setError("");
    const input={...query,from:query.from?query.from+"T00:00:00+03:30":undefined,to:query.to?query.to+"T23:59:59.999+03:30":undefined};
    fetch((import.meta.env.VITE_API_BASE_URL??"")+"/backoffice/observability/funnel/query",{method:"POST",headers:{Authorization:"Bearer "+authToken,"Content-Type":"application/json"},body:JSON.stringify(input),cache:"no-store",signal:controller.signal})
      .then(async response=>{if(!response.ok)throw Error(response.status===403?"مشاهده گزارش فقط برای ادمین مجاز است.":"دریافت آمار انجام نشد؛ بازه یا فیلتر را بررسی و دوباره تلاش کنید.");return response.json() as Promise<Funnel>;})
      .then(value=>{if(current===generation.current)setData(value);}).catch(e=>{if(!controller.signal.aborted&&current===generation.current)setError(e instanceof Error?e.message:"خطای دریافت گزارش");}).finally(()=>{if(current===generation.current)setBusy(false);});
    return()=>{generation.current++;controller.abort();};
  },[authToken,query,refresh]);
  const max=Math.max(1,...(data?.stages.map(s=>s.events)??[]));
  return <section aria-label="گزارش مسیر رفتار مشتری" aria-busy={busy}>
    <p>شمار رخدادهای موفق سروری در مرحله‌های مشاهده، سبد و خرید؛ هر پرداخت قطعی برای یک سفارش فقط یک‌بار شمرده می‌شود.</p>
    <FilterPanel onSearch={()=>setQuery({...draft})}><div className="operation-report-filters">
      <PersianDateField label="از تاریخ گزارش" value={draft.from} onChange={from=>setDraft({...draft,from})}/><PersianDateField label="تا تاریخ گزارش" value={draft.to} onChange={to=>setDraft({...draft,to})}/>
      <TextField select size="small" label="نوع دستگاه" value={draft.device} onChange={e=>setDraft({...draft,device:e.target.value})}><MenuItem value="">همه</MenuItem>{Object.entries({desktop:"دسکتاپ",mobile:"موبایل",tablet:"تبلت",unknown:"نامشخص"}).map(([key,label])=><MenuItem key={key} value={key}>{label}</MenuItem>)}</TextField>
      <TextField select size="small" label="دسته‌بندی" value={draft.category_id} onChange={e=>setDraft({...draft,category_id:e.target.value})}><MenuItem value="">همه</MenuItem>{state.categories?.map(c=><MenuItem key={c.id} value={c.id}>{c.name}</MenuItem>)}</TextField>
      <TextField select size="small" label="محیط داده" value={draft.environment} onChange={e=>setDraft({...draft,environment:e.target.value})}><MenuItem value="">محیط فعلی</MenuItem>{Object.entries(environments).map(([key,label])=><MenuItem key={key} value={key}>{label}</MenuItem>)}</TextField>
    </div></FilterPanel>
    {error&&<Alert severity="error">{error}</Alert>}
    {data&&<><p>محیط گزارش: {environments[data.environment]??data.environment} · سفارش‌های پرداخت‌شده: <strong>{number(data.unique_paid_purchases)}</strong></p>
      <div className="activity-funnel">{data.stages.map(stage=><article className="activity-funnel-stage" key={stage.key}><div><strong>{stage.label}</strong><span>{number(stage.events)} رخداد</span></div><div className="activity-funnel-track" aria-hidden="true"><span style={{width:(stage.events/max*100)+"%"}}/></div><small>{number(stage.known_identities)} کاربر یا سبد شناخته‌شده</small></article>)}</div>
      {!data.matched_events&&<Alert severity="info">هنوز رویدادی در این بازه و فیلتر ثبت نشده است.</Alert>}
      <p className="gateway-field-help">آمار مرحله‌ها شمار رخداد است، نه نرخ تبدیل یک گروه ثابت. مهمان فقط با سبد موجود شناخته می‌شود؛ برای بازدید بدون سبد یا ورود، هویت تخمین زده نمی‌شود. شروع خرید یعنی ایجاد موفق پیش‌فاکتور؛ شارژ کیف پول به‌تنهایی پرداخت سفارش نیست.</p>
    </>}
    <Button disabled={busy} onClick={()=>setRefresh(v=>v+1)}>به‌روزرسانی آمار</Button>
  </section>;
}
