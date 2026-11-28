import { useCallback, useEffect, useRef, useState, type FormEvent } from "react";
import { Alert, Button, Checkbox, Dialog, DialogActions, DialogContent, DialogTitle, FormControlLabel, MenuItem, TextField } from "@mui/material";
import { EditorTabs } from "../../shared/editor-tabs";
import { MultiChoiceField } from "../../shared/option-select";
import { money } from "../../../../../packages/commerce-client/types";

type ProviderField = { key: string; label: string; kind: string; default: string; required_for_activation: boolean; read_only: boolean; helper: string; options?: { value: string; label: string }[] };
type Provider = { id: string; name: string; execution_ready: boolean; fields: ProviderField[] };
type Gateway = {
  id: string; version: number; provider: string; name: string; active: boolean;
  min_toman: string; max_toman: string; all_categories: boolean; allowed_categories: string[]; excluded_categories: string[];
  instructions: string; max_automatic_retries: number; allow_wallet_topup: boolean;
  values: Record<string, string>; configured_credentials: string[]; execution_ready: boolean;
};
type Directory = { providers: Provider[]; items: Gateway[]; can_write: boolean; credentials_writable: boolean };
type Draft = Gateway & { credentials: Record<string, string>; clear_credentials: string[]; retryText: string };
const base = import.meta.env.VITE_API_BASE_URL ?? "";
const digits = (v: string) => v.replace(/[۰-۹]/g, d => String(d.charCodeAt(0) - 1776)).replace(/[٠-٩]/g, d => String(d.charCodeAt(0) - 1632));

export function PaymentGatewaySettings({ token, categories }: { token: string; categories: { id: string; name: string }[] }) {
  const [directory, setDirectory] = useState<Directory | null>(null);
  const [error, setError] = useState(""), [notice, setNotice] = useState(""), [loading, setLoading] = useState(false), [busy, setBusy] = useState(false);
  const [draft, setDraft] = useState<Draft | null>(null), [tab, setTab] = useState("general"), [fields, setFields] = useState<Record<string, string>>({});
  const generation = useRef(0), saving = useRef(false), pending = useRef<{ fingerprint: string; key: string } | null>(null);
  const load = useCallback(async () => {
    const current = ++generation.current; setLoading(true); setError("");
    try {
      const response = await fetch(base + "/backoffice/commerce/gateways", { headers: { Authorization: "Bearer " + token }, cache: "no-store" });
      if (!response.ok) throw Error(response.status === 403 ? "مجوز مشاهده تنظیمات درگاه را ندارید." : "دریافت تنظیمات درگاه انجام نشد.");
      const data = await response.json() as Directory;
      if (current === generation.current) setDirectory(data);
    } catch (e) { if (current === generation.current) setError(e instanceof Error ? e.message : "دریافت تنظیمات درگاه انجام نشد."); }
    finally { if (current === generation.current) setLoading(false); }
  }, [token]);
  useEffect(() => { setDirectory(null); setDraft(null); pending.current = null; void load(); return () => { generation.current++; }; }, [load]);
  const provider = directory?.providers.find(item => item.id === draft?.provider);
  const open = (item?: Gateway) => {
    const selected = directory?.providers.find(p => p.id === item?.provider) ?? directory?.providers[0];
    if (!selected) return;
    const values = Object.fromEntries(selected.fields.filter(f => f.kind !== "secret").map(f => [f.key, f.default]));
    setDraft({ id: "", version: 0, provider: selected.id, name: selected.name, active: false, min_toman: "0", max_toman: "", all_categories: true, allowed_categories: [], excluded_categories: [], instructions: "", max_automatic_retries: 0, allow_wallet_topup: false, values, configured_credentials: [], execution_ready: false, ...item, credentials: {}, clear_credentials: [], retryText: String(item?.max_automatic_retries ?? 0) });
    setFields({}); setTab("general"); setError(""); setNotice(""); pending.current = null;
  };
  const patch = (next: Partial<Draft>) => {
    setDraft(current => current ? { ...current, ...next } : null);
    setFields({}); setError("");
  };
  const close = () => { if (!saving.current) { setDraft(null); setFields({}); pending.current = null; } };
  async function save(event: FormEvent) {
    event.preventDefault(); if (!draft || saving.current) return;
    const retry = digits(draft.retryText).trim();
    if (!/^\d+$/.test(retry) || Number(retry) > 10) { setFields({ max_automatic_retries: "تعداد تلاش مجدد باید از صفر تا ۱۰ باشد." }); setTab("general"); return; }
    const { configured_credentials: _configured, execution_ready: _ready, retryText: _retryText, ...input } = draft;
    input.max_automatic_retries = Number(retry); input.min_toman = digits(input.min_toman); input.max_toman = digits(input.max_toman);
    const fingerprint = JSON.stringify(input);
    if (pending.current?.fingerprint !== fingerprint) pending.current = { fingerprint, key: crypto.randomUUID() };
    saving.current = true; setBusy(true); setError(""); setFields({});
    const current = generation.current;
    try {
      const response = await fetch(base + "/backoffice/commerce/gateways", { method: "POST", cache: "no-store", headers: { Authorization: "Bearer " + token, "Content-Type": "application/json", "Idempotency-Key": pending.current.key }, body: fingerprint });
      const result = await response.json();
      if (current !== generation.current) return;
      if (!response.ok) {
        if (result.fields) { setFields(result.fields); setTab(Object.keys(result.fields).some(k => k.startsWith("values.") || k.startsWith("credentials")) ? "connection" : "general"); }
        throw Error(result.code === "gateway_key_unavailable" ? "کلید نگهداری امن اطلاعات اتصال در سرور تنظیم نشده یا تغییر کرده است." : response.status === 409 ? "این تنظیمات قبلاً تغییر کرده است؛ پنجره را ببندید و فهرست را به‌روزرسانی کنید." : response.status === 403 ? "مجوز ذخیره تنظیمات درگاه را ندارید." : result.fields ? "فیلدهای مشخص‌شده را اصلاح کنید." : "ذخیره انجام نشد؛ می‌توانید همین درخواست را دوباره ارسال کنید.");
      }
      const saved = result as Gateway;
      setDirectory(data => data ? { ...data, items: [...data.items.filter(g => g.id !== saved.id), saved] } : data);
      setDraft(null); pending.current = null; setNotice("تنظیمات درگاه ذخیره شد؛ اجرای پرداخت در مرحله بعد پیاده‌سازی می‌شود.");
    } catch (e) { if (current === generation.current) setError(e instanceof Error ? e.message : "ذخیره تنظیمات انجام نشد."); }
    finally { saving.current = false; setBusy(false); }
  }
  return <section className="panel payment-gateway-settings" aria-label="تنظیمات درگاه‌های آنلاین" aria-busy={loading || busy}>
    <div className="page-toolbar"><div><h2>درگاه‌های بانکی و آنلاین</h2><p>تنظیمات اختصاصی هر درگاه در کنار روش‌های پرداخت فعلی.</p></div><Button variant="contained" disabled={!directory?.can_write || loading} onClick={() => open()}>افزودن درگاه</Button></div>
    <Alert severity="info">در این مرحله فقط تنظیمات ذخیره می‌شود؛ اتصال زرین‌پال و اجرای پرداخت بعد از بررسی فرم پیاده‌سازی می‌شود.</Alert>
    {!draft && error && <Alert severity="error">{error}</Alert>}{notice && <Alert severity="success">{notice}</Alert>}
    {loading && !directory && <p role="status">دریافت تنظیمات درگاه‌ها…</p>}
    <Button disabled={loading || busy} onClick={() => void load()}>به‌روزرسانی درگاه‌ها</Button>
    {directory && !directory.items.length && <p>هنوز درگاهی تعریف نشده است.</p>}
    {directory?.items.map(item => <article className="gateway-setting-summary" key={item.id}><div><h3>{item.name}</h3><p>{directory.providers.find(p => p.id === item.provider)?.name} · {item.values.environment === "production" ? "محیط واقعی" : "محیط آزمایشی"} · پیش‌نویس تنظیمات</p><p>تلاش مجدد: {item.max_automatic_retries.toLocaleString("fa-IR")} · شارژ مستقل کیف پول: {item.allow_wallet_topup ? "مجاز" : "غیرمجاز"}</p><p>حداقل: {money(item.min_toman)} · حداکثر: {item.max_toman ? money(item.max_toman) : "بدون سقف"}</p></div><Button variant="outlined" onClick={() => open(item)}>{directory.can_write ? "ویرایش تنظیمات" : "مشاهده تنظیمات"}</Button></article>)}
    <Dialog open={!!draft} onClose={close} fullWidth maxWidth="md" PaperProps={{ component: "form", onSubmit: (event: FormEvent) => void save(event), noValidate: true }}>
      <DialogTitle>{draft?.id ? "تنظیمات درگاه «" + draft.name + "»" : "افزودن درگاه آنلاین"}</DialogTitle>
      <DialogContent>{draft && provider && <>
        {error && <Alert severity="error">{error}</Alert>}
        <Alert severity="info">این درگاه هنوز برای مشتری فعال نمی‌شود؛ تغییر محیط یا ذخیره شناسه هیچ تراکنشی ایجاد نمی‌کند.</Alert>
        <div className="gateway-settings-fieldset">
          <EditorTabs label="بخش‌های تنظیمات درگاه" value={tab} onChange={setTab} disabled={busy} panels={[
            { id: "general", label: "تنظیمات عمومی", content: <fieldset disabled={busy || !directory?.can_write} className="gateway-settings-fieldset"><div className="gateway-settings-grid">
              <TextField select fullWidth size="small" label="ارائه‌دهنده درگاه" value={draft.provider} disabled={!!draft.id} onChange={e => {
                const p = directory?.providers.find(item => item.id === e.target.value); if (!p) return;
                patch({ provider: p.id, name: p.name, values: Object.fromEntries(p.fields.filter(f => f.kind !== "secret").map(f => [f.key, f.default])), credentials: {}, clear_credentials: [] });
              }}>{directory?.providers.map(p => <MenuItem key={p.id} value={p.id}>{p.name}</MenuItem>)}</TextField>
              <TextField label="نام نمایشی درگاه" fullWidth size="small" required value={draft.name} error={!!fields.name} helperText={fields.name} onChange={e => patch({ name: e.target.value })} inputProps={{ maxLength: 80 }} />
              <TextField label="حداقل کل سفارش (تومان)" fullWidth size="small" value={draft.min_toman} error={!!fields.min_toman} helperText={fields.min_toman} onChange={e => patch({ min_toman: e.target.value })} inputProps={{ inputMode: "decimal" }} />
              <TextField label="حداکثر (خالی = بدون سقف)" fullWidth size="small" value={draft.max_toman} error={!!fields.max_toman} helperText={fields.max_toman} onChange={e => patch({ max_toman: e.target.value })} inputProps={{ inputMode: "decimal" }} />
              <TextField label="آستانه تحمل خطا سفارشات" fullWidth size="small" value={draft.retryText} error={!!fields.max_automatic_retries} helperText={fields.max_automatic_retries || "تعداد retry اضافی خودکار: صفر تا ۱۰؛ صفر = خاموش. پیگیری همان تراکنش، بدون برداشت مجدد."} onChange={e => patch({ retryText: e.target.value })} inputProps={{ inputMode: "numeric", maxLength: 2 }} />
              <div><FormControlLabel control={<Checkbox checked={draft.allow_wallet_topup} onChange={e => patch({ allow_wallet_topup: e.target.checked })} />} label="امکان شارژ کیف پول" /><p className="gateway-field-help">اجازه شارژ مستقل کیف پول؛ این سیاست پس از پیاده‌سازی اتصال اعمال می‌شود.</p></div>
              <div className="gateway-settings-wide"><FormControlLabel control={<Checkbox checked={draft.all_categories} onChange={e => patch({ all_categories: e.target.checked })} />} label="همه دسته‌ها به جز موارد مستثنا" /></div>
              {!draft.all_categories && <MultiChoiceField label="دسته‌های مجاز" value={draft.allowed_categories} options={categories} onChange={v => patch({ allowed_categories: v })} />}
              <MultiChoiceField label="دسته‌های مستثنا" value={draft.excluded_categories} options={categories} onChange={v => patch({ excluded_categories: v })} />
              {fields.categories && <Alert severity="error" className="gateway-settings-wide">{fields.categories}</Alert>}
              <TextField className="gateway-settings-wide" label="راهنمای پرداخت" fullWidth size="small" multiline minRows={2} maxRows={4} value={draft.instructions} error={!!fields.instructions} helperText={fields.instructions} onChange={e => patch({ instructions: e.target.value })} inputProps={{ maxLength: 1000 }} />
            </div></fieldset> },
            { id: "connection", label: "اتصال درگاه", content: <fieldset disabled={busy || !directory?.can_write} className="gateway-settings-fieldset"><div className="gateway-settings-grid">
              {!directory?.credentials_writable && <Alert severity="warning" className="gateway-settings-wide">کلید امن سرور برای تغییر اطلاعات محرمانه در دسترس نیست؛ پیش‌نویس بدون شناسه پذیرنده قابل ذخیره است.</Alert>}
              {provider.fields.map(field => {
                const secret = field.kind === "secret";
                const path = (secret ? "credentials." : "values.") + field.key;
                const configured = draft.configured_credentials.includes(field.key) && !draft.clear_credentials.includes(field.key);
                const value = secret ? draft.credentials[field.key] ?? "" : draft.values[field.key] ?? "";
                return <div key={field.key} className={field.key === "base_url" ? "gateway-settings-wide" : undefined}>
                  <TextField label={field.label} select={field.kind === "select"} type={secret ? "password" : "text"} fullWidth size="small" value={value} error={!!fields[path]} disabled={secret && !directory?.credentials_writable} autoComplete={secret ? "new-password" : "off"} helperText={fields[path] || (secret ? (configured ? "شناسه ذخیره شده است؛ برای حفظ آن این فیلد را خالی بگذارید. " : "") + field.helper : field.helper)} inputProps={{ readOnly: field.read_only, dir: ["secret", "url"].includes(field.kind) || field.key === "currency" ? "ltr" : undefined, maxLength: secret ? 36 : field.key === "description" ? 500 : 2048 }} onChange={e => {
                    if (secret) { patch({ credentials: { ...draft.credentials, [field.key]: e.target.value }, clear_credentials: draft.clear_credentials.filter(k => k !== field.key) }); return; }
                    const values = { ...draft.values, [field.key]: e.target.value };
                    if (field.key === "environment") values.base_url = e.target.value === "production" ? "https://payment.zarinpal.com/pg/v4/payment" : "https://sandbox.zarinpal.com/pg/v4/payment";
                    patch({ values });
                  }}>{field.options?.map(option => <MenuItem key={option.value} value={option.value}>{option.label}</MenuItem>)}</TextField>
                  {secret && configured && <Button disabled={!directory?.credentials_writable} onClick={() => patch({ credentials: { ...draft.credentials, [field.key]: "" }, clear_credentials: [...draft.clear_credentials, field.key] })}>پاک‌کردن شناسه ذخیره‌شده</Button>}
                  {secret && draft.clear_credentials.includes(field.key) && <p className="gateway-field-help">شناسه پس از ذخیره پاک می‌شود.</p>}
                </div>;
              })}
            </div></fieldset> },
          ]} />
        </div>
      </>}</DialogContent>
      <DialogActions><Button onClick={close} disabled={busy}>بستن</Button><Button type="submit" variant="contained" disabled={busy || !directory?.can_write}>{busy ? "در حال ذخیره…" : "ذخیره تنظیمات درگاه"}</Button></DialogActions>
    </Dialog>
  </section>;
}
