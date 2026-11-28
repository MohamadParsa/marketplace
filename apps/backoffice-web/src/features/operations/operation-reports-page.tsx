import { useCallback, useEffect, useRef, useState } from "react";
import { Alert, Button, MenuItem, TextField } from "@mui/material";
import { useWorkspace } from "../../shared/workspace";
import { DataTable, PageHeader } from "../../shared/components";
import { FilterPanel } from "../../shared/filter-panel";
import { EditorTabs } from "../../shared/editor-tabs";
import { dateTime, number, displayRecord, labels } from "../../shared/format";
import { PersianDateField } from "../../../../../packages/date-controls/persian-date-field";

type LogEvent = { event_id: string; event_name: string; log_type: string; severity: string; outcome: string; recorded_at: string; operation_id: string; actor_id?: string; provider?: string; provider_operation?: string; error_code?: string; duration_ms?: number; http_status?: number; attributes?: Record<string, unknown> };
type AuditRow = { id: string; actor: string; action: string; target: string; store_id?: string; operation_id?: string; outcome: string; version?: number; before: Record<string,string>; after: Record<string,string>; reason_present: boolean; at: string };
type Stats = { status: string; recorder?: { recorded: number; rejected: number; write_failures: number }; storage?: { written: number; dropped: number; write_failures: number; corrupt_lines: number; queued: number } };
type Filters = { from: string; to: string; log_type: string; severity: string; outcome: string; operation_id: string; actor_id: string; event_name: string; store_id: string };
const emptyFilters: Filters = { from:"", to:"", log_type:"", severity:"", outcome:"", operation_id:"", actor_id:"", event_name:"", store_id:"" };
const base=import.meta.env.VITE_API_BASE_URL??"";
const types:Record<string,string>={technical:"فنی",integration:"ارتباط سرویس",business:"عملیات کسب‌وکار",security:"امنیت و ورود",behavior:"رفتار مشتری"};
const outcomes:Record<string,string>={started:"شروع‌شده",succeeded:"موفق",rejected:"ردشده",failed:"ناموفق",pending:"در انتظار",unknown:"نتیجه نامشخص",cancelled:"لغوشده",replayed:"درخواست تکراری"};
const fields:Record<string,string>={amount_rial:"مبلغ (ریال)",grants:"نقش‌های منتسب",permissions:"مجوزها",order_status:"وضعیت سفارش",payment:"وضعیت پرداخت",version:"نسخه",active:"فعال",role:"نقش",access_version:"نسخه دسترسی",password_configured:"رمز تنظیم شده",price_rial:"قیمت (ریال)",stock:"موجودی",reserved:"رزرو",status:"وضعیت",content_policy_version:"نسخه سیاست محتوا",published:"منتشرشده",value:"مقدار",kind:"نوع",allow_wallet_topup:"شارژ کیف پول",credential_count:"تعداد اطلاعات اتصال"};
function auditValue(value?:string){if(value===undefined)return "—";if(value==="true")return "بله";if(value==="false")return "خیر";if(/^-?\d+$/.test(value))return new Intl.NumberFormat("fa-IR").format(BigInt(value));return labels[value]??value;}
export function OperationReportsPage(){
  const { authToken, role, surface, state }=useWorkspace();
  const [tab,setTab]=useState("audit"),[draft,setDraft]=useState(emptyFilters),[filters,setFilters]=useState(emptyFilters);
  const [audit,setAudit]=useState<AuditRow[]>([]),[events,setEvents]=useState<LogEvent[]>([]),[next,setNext]=useState(""),[stats,setStats]=useState<Stats|null>(null),[error,setError]=useState(""),[busy,setBusy]=useState(false);
  const generation=useRef(0);
  const load=useCallback(async(cursor="",selected=filters)=>{
    const current=++generation.current;setBusy(true);setError("");
    try{
      const payload={...selected,from:selected.from?selected.from+"T00:00:00+03:30":undefined,to:selected.to?selected.to+"T23:59:59.999+03:30":undefined,limit:50,cursor};
      const response=await fetch(base+"/backoffice/observability/"+(tab==="audit"?"audit":"events")+"/query",{method:"POST",headers:{Authorization:"Bearer "+authToken,"Content-Type":"application/json"},body:JSON.stringify(payload),cache:"no-store"});
      if(!response.ok)throw Error(response.status===403?"مشاهده گزارش فقط برای ادمین مجاز است.":"دریافت گزارش انجام نشد؛ فیلترها و اتصال را بررسی کنید.");
      const data=await response.json();
      const statusResponse=await fetch(base+"/backoffice/observability/stats",{headers:{Authorization:"Bearer "+authToken},cache:"no-store"});
      if(current!==generation.current)return;
      if(tab==="audit")setAudit(rows=>cursor?[...rows,...data.items]:data.items);else setEvents(rows=>cursor?[...rows,...data.items]:data.items);
      setNext(data.next_cursor??"");if(statusResponse.ok)setStats(await statusResponse.json());
    }catch(e){if(current===generation.current)setError(e instanceof Error?e.message:"خطای دریافت گزارش");}finally{if(current===generation.current)setBusy(false);}
  },[authToken,tab,filters]);
  useEffect(()=>{if(role==="super-admin"&&surface==="backoffice")void load();return()=>{generation.current++;};},[load,role,surface]);
  if(role!=="super-admin"||surface!=="backoffice")return <Alert severity="warning">مشاهده گزارش فقط برای ادمین مجاز است.</Alert>;
  const actor=(id:string)=>state.users?.find(u=>u.id===id)?.name??(id==="system"?"سیستم":"کاربر ثبت‌شده");
  const filter=<FilterPanel onSearch={()=>{setFilters({...draft});}}><div className="operation-report-filters">
    <PersianDateField label="از تاریخ" value={draft.from} onChange={value=>setDraft({...draft,from:value})}/><PersianDateField label="تا تاریخ" value={draft.to} onChange={value=>setDraft({...draft,to:value})}/>
    {tab==="events"&&<TextField select size="small" label="نوع لاگ" value={draft.log_type} onChange={e=>setDraft({...draft,log_type:e.target.value})}><MenuItem value="">همه</MenuItem>{Object.entries(types).map(([v,n])=><MenuItem key={v} value={v}>{n}</MenuItem>)}</TextField>}
    {tab==="events"&&<TextField select size="small" label="سطح" value={draft.severity} onChange={e=>setDraft({...draft,severity:e.target.value})}><MenuItem value="">همه</MenuItem>{["debug","info","warn","error"].map(v=><MenuItem key={v} value={v}>{v}</MenuItem>)}</TextField>}
    <TextField select size="small" label="نتیجه" value={draft.outcome} onChange={e=>setDraft({...draft,outcome:e.target.value})}><MenuItem value="">همه</MenuItem>{Object.entries(outcomes).map(([v,n])=><MenuItem key={v} value={v}>{n}</MenuItem>)}</TextField>
    <TextField select size="small" label="عامل" value={draft.actor_id} onChange={e=>setDraft({...draft,actor_id:e.target.value})}><MenuItem value="">همه</MenuItem><MenuItem value="system">سیستم</MenuItem>{state.users?.map(u=><MenuItem key={u.id} value={u.id}>{u.name}</MenuItem>)}</TextField>
    <TextField size="small" label="شناسه پیگیری عملیات" value={draft.operation_id} onChange={e=>setDraft({...draft,operation_id:e.target.value})}/>
    <TextField size="small" label="کد رویداد / عملیات" value={draft.event_name} onChange={e=>setDraft({...draft,event_name:e.target.value})}/>
    {tab==="audit"&&<TextField select size="small" label="فروشگاه" value={draft.store_id} onChange={e=>setDraft({...draft,store_id:e.target.value})}><MenuItem value="">همه</MenuItem>{state.stores?.map(s=><MenuItem key={s.id} value={s.id}>{s.name}</MenuItem>)}</TextField>}
  </div></FilterPanel>;
  return <><div className="operation-report-header"><PageHeader title="گزارش عملیات و لاگ‌ها" description="تاریخچه تغییرات و پیگیری مسیر اجرای عملیات؛ ویژه ادمین"/></div>
    {error&&<Alert severity="error">{error}</Alert>}
    {stats&&<Alert severity={stats.status==="ready"&&!(stats.storage?.dropped||stats.storage?.write_failures||stats.recorder?.write_failures)?"info":"warning"}>{stats.status==="ready"?"ذخیره محلی فعال است.":stats.status==="disabled"?"ذخیره محلی در این محیط تنظیم نشده است.":"ذخیره محلی در دسترس نیست؛ عملیات اصلی ادامه دارد."} ثبت‌شده: {number(stats.storage?.written??0)} · افت صف: {number(stats.storage?.dropped??0)} · خطای نوشتن: {number((stats.storage?.write_failures??0)+(stats.recorder?.write_failures??0))}</Alert>}
    <section className="panel operation-reports" aria-busy={busy}><EditorTabs label="بخش‌های گزارش عملیات" value={tab} onChange={value=>{setTab(value);setNext("");setDraft(emptyFilters);setFilters(emptyFilters);}} panels={[
      {id:"audit",label:"سابقه عملیات",content:<>{tab==="audit"&&filter}<DataTable searchable={false} rows={audit} empty={busy?"در حال دریافت…":"عملیاتی با این فیلتر ثبت نشده است."} columns={[
        {title:"عملیات",render:a=><span className="code">{a.action}</span>},{title:"عامل",render:a=>actor(a.actor)},{title:"هدف",render:a=>a.target?displayRecord(state,a.target):"رکورد جدید"},{title:"زمان",render:a=>dateTime(a.at)},
        {title:"جزئیات",render:a=><details><summary>مشاهده تغییرات</summary><p>نتیجه: {outcomes[a.outcome]??a.outcome} · نسخه: {number(a.version??0)}</p>{a.operation_id&&<p className="log-reference">پیگیری: <bdi>{a.operation_id}</bdi></p>}{Array.from(new Set([...Object.keys(a.before),...Object.keys(a.after)])).map(key=><p key={key}>{fields[key]??key}: <bdi>{auditValue(a.before[key])}</bdi> ← <bdi>{auditValue(a.after[key])}</bdi></p>)}{!Object.keys(a.after).length&&!Object.keys(a.before).length&&<p>جزئیات تغییر برای سابقهٔ قدیمی ثبت نشده است.</p>}{a.reason_present&&<p>شرح عملیات در پرونده یا فرم مربوط نگهداری شده است.</p>}</details>}
      ]}/></>},
      {id:"events",label:"لاگ اجرا و سرویس‌ها",content:<>{tab==="events"&&filter}<DataTable searchable={false} rows={events.map(e=>({...e,id:e.event_id}))} empty={busy?"در حال دریافت…":"رویدادی با این فیلتر ثبت نشده است."} columns={[
        {title:"رویداد",render:e=><span className="code">{e.event_name}</span>},{title:"نوع / سطح",render:e=><>{types[e.log_type]??e.log_type} · {e.severity}</>},{title:"نتیجه",render:e=>outcomes[e.outcome]??e.outcome},{title:"زمان",render:e=>dateTime(e.recorded_at)},
        {title:"جزئیات",render:e=><details><summary>مشاهده پیگیری</summary><p className="log-reference">عملیات: <bdi>{e.operation_id}</bdi></p>{e.provider&&<p>سرویس: {e.provider} · {e.provider_operation}</p>}<p>مدت: {number(e.duration_ms??0)} میلی‌ثانیه · پاسخ: {e.http_status??"—"}</p>{e.error_code&&<p>کد خطا: {e.error_code}</p>}</details>}
      ]}/></>}
    ]}/><div className="page-toolbar"><Button disabled={busy} onClick={()=>void load()}>به‌روزرسانی گزارش</Button>{next&&<Button variant="outlined" disabled={busy} onClick={()=>void load(next)}>نمایش بیشتر</Button>}</div></section>
  </>;
}
