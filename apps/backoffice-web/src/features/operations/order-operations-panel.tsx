import { useRef, useState, type FormEvent } from "react";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { Alert, Button, Checkbox, FormControlLabel, Stack, TextField } from "@mui/material";
import { api, ApiError } from "../../shared/api";
import { useWorkspace } from "../../shared/workspace";
import { ChoiceField } from "../../shared/option-select";
import { dateTime } from "../../shared/format";
import { label } from "../../../../../packages/commerce-client/types";

const actions: Record<string,string> = {
  start: "شروع پردازش سفارش", supply: "تأیید تأمین تمام اقلام", pack: "شروع آماده‌سازی",
  ship: "ثبت ارسال به حمل‌کننده", deliver: "تأیید دریافت مشتری",
  "configure-dispatch": "تعیین نوع شناسه ارسال", correct: "اصلاح مشخصات ارسال با دلیل",
};

export function OrderOperationsPanel({ id, onChanged }: { id: string; onChanged?: () => void | Promise<void> }) {
  const { authToken, surface } = useWorkspace();
  const cache = useQueryClient();
  const queryKey = ["order-operations",authToken,surface,id];
  const query = useQuery({ queryKey, queryFn: () => api.orderOperations(authToken,id,surface) });
  const [chosen,setChosen] = useState("");
  const [reason,setReason] = useState("");
  const [codes,setCodes] = useState<string>();
  const [policy,setPolicy] = useState("");
  const [allItems,setAllItems] = useState(false);
  const [busy,setBusy] = useState(false);
  const [error,setError] = useState("");
  const pending = useRef<{ signature:string;key:string }>();
  const running = useRef(false);
  const data = query.data;
  const action = data?.allowed_actions.includes(chosen) ? chosen : data?.allowed_actions[0] ?? "";
  const requiresReason = ["deliver","configure-dispatch","correct"].includes(action);
  const requiresCodes = ["ship","correct"].includes(action) && ["tracking","delivery"].includes(data?.dispatch_identifier ?? "");
  const identifiers = codes ?? data?.identifiers?.join("\n") ?? "";

  async function submit(event: FormEvent) {
    event.preventDefault();
    if (!data || running.current || !action) return;
    const input = { version:data.version, action, reason:reason.trim(), all_items:allItems,
      identifiers: ["ship","correct"].includes(action) ? identifiers.split(/\r?\n/).map(s => s.trim()).filter(Boolean) : [],
      dispatch_identifier: action === "configure-dispatch" ? policy : undefined };
    const signature = JSON.stringify(input);
    if (pending.current?.signature !== signature) pending.current = { signature,key:crypto.randomUUID() };
    running.current = true;setBusy(true);setError("");
    try {
      const result = await api.transitionOrder(authToken,id,surface,input,pending.current.key);
      cache.setQueryData(queryKey,result);
      pending.current = undefined;setReason("");setCodes(undefined);setAllItems(false);setChosen("");
      await cache.invalidateQueries({ queryKey:["workspace"] });
      await onChanged?.();
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "ثبت عملیات انجام نشد.");
      if (cause instanceof ApiError && cause.status === 409) { pending.current = undefined; await query.refetch(); }
    } finally { running.current = false;setBusy(false); }
  }

  return <section className="panel order-operations-panel" aria-label="عملیات مرسوله">
    <h2>عملیات مرسوله</h2>
    {query.isPending && <p role="status">دریافت وضعیت مرسوله…</p>}
    {query.error && <Alert severity="error" action={<Button onClick={() => void query.refetch()}>تلاش مجدد</Button>}>{query.error.message}</Alert>}
    {data && <Stack spacing={2}>
      <p>{data.store_name} · سفارش {data.purchase_number.replace(/\d/g,d => "۰۱۲۳۴۵۶۷۸۹"[Number(d)])} · {data.shipping_name}</p>
      <div className="order-stage-summary"><span>سفارش: <strong>{label(data.order_status)}</strong></span><span>پرداخت: <strong>{label(data.payment_status)}</strong></span><span>عملیات: <strong>{label(data.status)}</strong></span></div>
      {!!data.identifiers?.length && <p>{data.dispatch_identifier === "delivery" ? "کد تحویل" : "کد رهگیری"}: <span dir="ltr">{data.identifiers.join("، ")}</span></p>}
      {data.address ? <section><h3>اطلاعات لازم برای تحویل</h3><p>{data.address.recipient} · {data.address.phone}</p><p>{data.address.province}، {data.address.city}، {data.address.street} · پلاک {data.address.house_number} · {data.address.postal_code}</p></section> : surface === "seller" && <Alert severity="info">اطلاعات گیرنده پس از تأیید تأمین تمام اقلام این مرسوله نمایش داده می‌شود.</Alert>}
      {!data.dispatch_identifier && data.payment_status === "paid" && <Alert severity="info">مدیر باید نوع شناسه ارسال این مرسوله قدیمی را پیش از ارسال تعیین کند.</Alert>}
      {action ? <form onSubmit={submit} className="order-operation-form">
        <ChoiceField label="اقدام عملیاتی" value={action} disabled={busy} options={data.allowed_actions.map(id => ({ id,name:actions[id] ?? id }))} onChange={value => { setChosen(value);setAllItems(false); }} />
        {action === "configure-dispatch" && <ChoiceField label="نوع شناسه ارسال" required clearLabel="انتخاب کنید" value={policy} disabled={busy} options={[{id:"tracking",name:"کد رهگیری"},{id:"delivery",name:"کد تحویل"},{id:"none",name:"بدون کد"}]} onChange={setPolicy} />}
        {requiresCodes && <TextField label={data.dispatch_identifier === "delivery" ? "کدهای تحویل؛ هر کد در یک سطر" : "کدهای رهگیری؛ هر کد در یک سطر"} required value={identifiers} onChange={e => setCodes(e.target.value)} multiline minRows={2} disabled={busy} inputProps={{dir:"ltr",maxLength:3200}} />}
        <TextField label={requiresReason ? "دلیل ثبت / اصلاح" : "توضیح عملیات"} required={requiresReason} value={reason} onChange={e => setReason(e.target.value)} disabled={busy} multiline minRows={2} inputProps={{maxLength:1000}} />
        {["ship","deliver"].includes(action) && <FormControlLabel control={<Checkbox checked={allItems} disabled={busy} onChange={e => setAllItems(e.target.checked)} />} label={action === "ship" ? "تمام اقلام باقی‌مانده این مرسوله به حمل‌کننده تحویل شده‌اند." : "دریافت تمام اقلام این مرسوله توسط مشتری تأیید شده است."} />}
        <Button type="submit" variant="contained" disabled={busy || requiresReason && !reason.trim() || requiresCodes && !identifiers.trim() || ["ship","deliver"].includes(action) && !allItems || action === "configure-dispatch" && !policy}>{actions[action]}</Button>
      </form> : <p>اقدام بعدی در این مرحله یا در محدوده دسترسی شما وجود ندارد.</p>}
      {error && <Alert severity="error">{error}</Alert>}
      {!!data.events.length && <details><summary>تاریخچه عملیات مرسوله</summary>{data.events.map((event,index) => <p key={event.id ?? index}>{dateTime(event.at)} · {label(event.action)} · {label(event.from)} ← {label(event.to)}{event.reason ? " · " + event.reason : ""}</p>)}</details>}
    </Stack>}
  </section>;
}
