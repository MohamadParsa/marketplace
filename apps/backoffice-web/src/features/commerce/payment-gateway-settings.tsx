import { useCallback, useEffect, useRef, useState, type FormEvent } from "react";
import { Alert, Button, Checkbox, Dialog, DialogActions, DialogContent, DialogTitle, FormControlLabel, MenuItem, Switch, TextField } from "@mui/material";
import { EditorTabs } from "../../shared/editor-tabs";
import { MultiChoiceField } from "../../shared/option-select";
import { money } from "../../../../../packages/commerce-client/types";

type ProviderField = { key: string; label: string; kind: string; section?: string; format?: string; max_length?: number; default: string; required_for_activation: boolean; read_only: boolean; helper: string; options?: { value: string; label: string }[] };
type ProviderSection = { id: string; label: string; description: string; expanded: boolean };
type Provider = { id: string; name: string; execution_ready: boolean; fields: ProviderField[]; sections?: ProviderSection[] };
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
const isCredential = (field: ProviderField) => field.kind === "secret" || field.kind === "credential";

export function PaymentGatewaySettings({ token, categories }: { token: string; categories: { id: string; name: string }[] }) {
  const [directory, setDirectory] = useState<Directory | null>(null);
  const [error, setError] = useState(""), [notice, setNotice] = useState(""), [loading, setLoading] = useState(false), [busy, setBusy] = useState(false);
  const [draft, setDraft] = useState<Draft | null>(null), [tab, setTab] = useState("general"), [fields, setFields] = useState<Record<string, string>>({});
  const [openSections, setOpenSections] = useState<Record<string, boolean>>({});
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
    const values = Object.fromEntries(selected.fields.filter(f => !isCredential(f)).map(f => [f.key, f.default]));
    setDraft({ id: "", version: 0, provider: selected.id, name: selected.name, active: false, min_toman: "0", max_toman: "", all_categories: true, allowed_categories: [], excluded_categories: [], instructions: "", max_automatic_retries: 0, allow_wallet_topup: false, values, configured_credentials: [], execution_ready: false, ...item, credentials: {}, clear_credentials: [], retryText: String(item?.max_automatic_retries ?? 0) });
    setFields({}); setOpenSections({}); setTab("general"); setError(""); setNotice(""); pending.current = null;
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
        if (result.fields) {
          setFields(result.fields); setTab(Object.keys(result.fields).some(k => k.startsWith("values.") || k.startsWith("credentials")) ? "connection" : "general");
          setOpenSections(current => ({ ...current, ...Object.fromEntries((provider?.fields ?? []).filter(f => result.fields[(isCredential(f) ? "credentials." : "values.") + f.key]).map(f => [f.section ?? "", true])) }));
        }
        throw Error(result.code === "gateway_key_unavailable" ? "کلید نگهداری امن اطلاعات اتصال در سرور تنظیم نشده یا تغییر کرده است." : response.status === 409 ? "این تنظیمات قبلاً تغییر کرده است؛ پنجره را ببندید و فهرست را به‌روزرسانی کنید." : response.status === 403 ? "مجوز ذخیره تنظیمات درگاه را ندارید." : result.fields ? "فیلدهای مشخص‌شده را اصلاح کنید." : "ذخیره انجام نشد؛ می‌توانید همین درخواست را دوباره ارسال کنید.");
      }
      const saved = result as Gateway;
      setDirectory(data => data ? { ...data, items: data.items.map(g => g.provider === saved.provider ? saved : g) } : data);
      setDraft(null); pending.current = null; setNotice("تنظیمات درگاه ذخیره شد؛ اجرای پرداخت در مرحله بعد پیاده‌سازی می‌شود.");
    } catch (e) { if (current === generation.current) setError(e instanceof Error ? e.message : "ذخیره تنظیمات انجام نشد."); }
    finally { saving.current = false; setBusy(false); }
  }
  async function toggleActive(item: Gateway, active: boolean) {
    if (saving.current || !directory?.can_write) return;
    const { configured_credentials: _configured, execution_ready: _ready, ...input } = item;
    input.active = active;
    const fingerprint = JSON.stringify(input);
    if (pending.current?.fingerprint !== fingerprint) pending.current = { fingerprint, key: crypto.randomUUID() };
    saving.current = true; setBusy(true); setError(""); setNotice("");
    const current = generation.current;
    try {
      const response = await fetch(base + "/backoffice/commerce/gateways", { method: "POST", cache: "no-store", headers: { Authorization: "Bearer " + token, "Content-Type": "application/json", "Idempotency-Key": pending.current.key }, body: fingerprint });
      const result = await response.json();
      if (current !== generation.current) return;
      if (!response.ok) throw Error(response.status === 409 ? "تنظیمات تغییر کرده است؛ فهرست را به‌روزرسانی کنید." : result.code === "gateway_key_unavailable" ? "کلید امن اطلاعات اتصال در سرور در دسترس نیست." : "تغییر وضعیت ذخیره نشد؛ می‌توانید دوباره تلاش کنید.");
      const saved = result as Gateway;
      setDirectory(data => data ? { ...data, items: data.items.map(g => g.provider === saved.provider ? saved : g) } : data);
      pending.current = null; setNotice("وضعیت درگاه در تنظیمات ذخیره شد.");
    } catch (e) { if (current === generation.current) setError(e instanceof Error ? e.message : "ذخیره وضعیت انجام نشد."); }
    finally { saving.current = false; setBusy(false); }
  }
  function connectionField(field: ProviderField) {
    if (!draft) return null;
    const credential = isCredential(field), secret = field.kind === "secret";
    const path = (credential ? "credentials." : "values.") + field.key;
    const configured = draft.configured_credentials.includes(field.key) && !draft.clear_credentials.includes(field.key);
    const value = credential ? draft.credentials[field.key] ?? "" : draft.values[field.key] ?? field.default;
    return <div key={field.key} className={field.kind === "url" ? "gateway-settings-wide" : undefined}>
      <TextField label={field.label} select={field.kind === "select"} type={secret ? "password" : "text"} fullWidth size="small" value={value} error={!!fields[path]} disabled={credential && !directory?.credentials_writable} autoComplete={credential ? "new-password" : "off"} helperText={fields[path] || (credential ? (configured ? "مقدار ذخیره شده است؛ برای حفظ آن این فیلد را خالی بگذارید. " : "") + field.helper : field.helper)} inputProps={{ readOnly: field.read_only, dir: credential || field.kind === "url" || field.key === "currency" ? "ltr" : undefined, inputMode: field.format === "positive-integer" ? "numeric" : undefined, maxLength: field.max_length || 2048 }} onChange={e => {
        if (credential) { patch({ credentials: { ...draft.credentials, [field.key]: e.target.value }, clear_credentials: draft.clear_credentials.filter(k => k !== field.key) }); return; }
        const values = { ...draft.values, [field.key]: e.target.value };
        if (field.key === "environment") values.base_url = e.target.value === "production" ? "https://payment.zarinpal.com/pg/v4/payment" : "https://sandbox.zarinpal.com/pg/v4/payment";
        patch({ values });
      }}>{field.options?.map(option => <MenuItem key={option.value} value={option.value}>{option.label}</MenuItem>)}</TextField>
      {credential && configured && <Button disabled={!directory?.credentials_writable} onClick={() => patch({ credentials: { ...draft.credentials, [field.key]: "" }, clear_credentials: [...draft.clear_credentials, field.key] })}>پاک‌کردن {field.label}</Button>}
      {credential && draft.clear_credentials.includes(field.key) && <p className="gateway-field-help">این مقدار پس از ذخیره پاک می‌شود.</p>}
    </div>;
  }
  return <section className="panel payment-gateway-settings" aria-label="تنظیمات درگاه‌های آنلاین" aria-busy={loading || busy}>
    <div className="page-toolbar"><div><h2>درگاه‌های بانکی و آنلاین</h2><p>تنظیمات همه درگاه‌های پشتیبانی‌شده همیشه در این بخش نمایش داده می‌شود.</p></div></div>
    <Alert severity="info">در این مرحله فقط تنظیمات ذخیره می‌شود؛ اتصال زرین‌پال و اجرای پرداخت بعد از بررسی فرم پیاده‌سازی می‌شود.</Alert>
    {!draft && error && <Alert severity="error">{error}</Alert>}{notice && <Alert severity="success">{notice}</Alert>}
    {loading && !directory && <p role="status">دریافت تنظیمات درگاه‌ها…</p>}
    <Button disabled={loading || busy} onClick={() => void load()}>به‌روزرسانی درگاه‌ها</Button>
    {directory?.items.map(item => <details className="gateway-provider-section" key={item.provider}><summary><strong>{item.name}</strong><span>{item.active ? "فعال در تنظیمات" : "غیرفعال"}</span></summary><div className="gateway-setting-summary"><div><p>{directory.providers.find(p => p.id === item.provider)?.name} · {item.values.environment === "production" ? "محیط واقعی" : "محیط آزمایشی"}</p><p>تلاش مجدد: {item.max_automatic_retries.toLocaleString("fa-IR")} · شارژ مستقل کیف پول: {item.allow_wallet_topup ? "مجاز" : "غیرمجاز"}</p><p>حداقل: {money(item.min_toman)} · حداکثر: {item.max_toman ? money(item.max_toman) : "بدون سقف"}</p></div><div><FormControlLabel control={<Switch checked={item.active} disabled={!directory.can_write || loading || busy} onChange={event => void toggleActive(item, event.target.checked)} inputProps={{ "aria-label": "فعال بودن درگاه " + item.name }} />} label={item.active ? "فعال در تنظیمات" : "غیرفعال"} /><Button variant="outlined" disabled={loading || busy} onClick={() => open(item)}>{directory.can_write ? "ویرایش تنظیمات" : "مشاهده تنظیمات"}</Button></div></div></details>)}
    <Dialog open={!!draft} onClose={close} fullWidth maxWidth="md" PaperProps={{ component: "form", onSubmit: (event: FormEvent) => void save(event), noValidate: true }}>
      <DialogTitle>{draft ? "تنظیمات درگاه «" + draft.name + "»" : "تنظیمات درگاه"}</DialogTitle>
      <DialogContent>{draft && provider && <>
        {error && <Alert severity="error">{error}</Alert>}
        <Alert severity="info">این درگاه هنوز برای مشتری فعال نمی‌شود؛ تغییر محیط یا ذخیره شناسه هیچ تراکنشی ایجاد نمی‌کند.</Alert>
        <div className="gateway-settings-fieldset">
          <EditorTabs label="بخش‌های تنظیمات درگاه" value={tab} onChange={setTab} disabled={busy} panels={[
            { id: "general", label: "تنظیمات عمومی", content: <fieldset disabled={busy || !directory?.can_write} className="gateway-settings-fieldset"><div className="gateway-settings-grid">
              <TextField fullWidth size="small" label="ارائه‌دهنده درگاه" value={provider.name} inputProps={{ readOnly: true }} />
              <TextField label="نام نمایشی درگاه" fullWidth size="small" required value={draft.name} error={!!fields.name} helperText={fields.name} onChange={e => patch({ name: e.target.value })} inputProps={{ maxLength: 80 }} />
              <TextField label="حداقل کل سفارش (تومان)" fullWidth size="small" value={draft.min_toman} error={!!fields.min_toman} helperText={fields.min_toman} onChange={e => patch({ min_toman: e.target.value })} inputProps={{ inputMode: "decimal" }} />
              <TextField label="حداکثر (خالی = بدون سقف)" fullWidth size="small" value={draft.max_toman} error={!!fields.max_toman} helperText={fields.max_toman} onChange={e => patch({ max_toman: e.target.value })} inputProps={{ inputMode: "decimal" }} />
              <TextField label="آستانه تحمل خطا سفارشات" fullWidth size="small" value={draft.retryText} error={!!fields.max_automatic_retries} helperText={fields.max_automatic_retries || "تعداد retry اضافی خودکار: صفر تا ۱۰؛ صفر = خاموش. پیگیری همان تراکنش، بدون برداشت مجدد."} onChange={e => patch({ retryText: e.target.value })} inputProps={{ inputMode: "numeric", maxLength: 2 }} />
              <div><FormControlLabel control={<Checkbox checked={draft.allow_wallet_topup} onChange={e => patch({ allow_wallet_topup: e.target.checked })} />} label="امکان شارژ کیف پول" /><p className="gateway-field-help">اجازه شارژ مستقل کیف پول؛ این سیاست پس از پیاده‌سازی اتصال اعمال می‌شود.</p></div>
              <div className="gateway-settings-wide"><FormControlLabel control={<Switch checked={draft.active} onChange={event => patch({ active: event.target.checked })} />} label="فعال بودن درگاه در تنظیمات" /></div>
              <div className="gateway-settings-wide"><FormControlLabel control={<Checkbox checked={draft.all_categories} onChange={e => patch({ all_categories: e.target.checked })} />} label="همه دسته‌ها به جز موارد مستثنا" /></div>
              {!draft.all_categories && <MultiChoiceField label="دسته‌های مجاز" value={draft.allowed_categories} options={categories} onChange={v => patch({ allowed_categories: v })} />}
              <MultiChoiceField label="دسته‌های مستثنا" value={draft.excluded_categories} options={categories} onChange={v => patch({ excluded_categories: v })} />
              {fields.categories && <Alert severity="error" className="gateway-settings-wide">{fields.categories}</Alert>}
              <TextField className="gateway-settings-wide" label="راهنمای پرداخت" fullWidth size="small" multiline minRows={2} maxRows={4} value={draft.instructions} error={!!fields.instructions} helperText={fields.instructions} onChange={e => patch({ instructions: e.target.value })} inputProps={{ maxLength: 1000 }} />
            </div></fieldset> },
            { id: "connection", label: "اتصال درگاه", content: <fieldset disabled={busy || !directory?.can_write} className="gateway-settings-fieldset">
              {!directory?.credentials_writable && <Alert severity="warning">کلید امن سرور برای تغییر اطلاعات ورود در دسترس نیست؛ پیش‌نویس بدون credential قابل ذخیره است.</Alert>}
              {(provider.sections ?? [{ id: "", label: "مشخصات اتصال", description: "", expanded: true }]).map(section => <details className="gateway-connection-section" key={section.id} open={openSections[section.id] ?? section.expanded} onToggle={event => { const expanded = event.currentTarget.open; setOpenSections(current => current[section.id] === expanded ? current : { ...current, [section.id]: expanded }); }}>
                <summary>{section.label}</summary>{section.description && <p className="gateway-field-help">{section.description}</p>}
                <div className="gateway-settings-grid">{provider.fields.filter(f => !provider.sections || f.section === section.id).map(connectionField)}</div>
              </details>)}
            </fieldset> },
          ]} />
        </div>
      </>}</DialogContent>
      <DialogActions><Button onClick={close} disabled={busy}>بستن</Button><Button type="submit" variant="contained" disabled={busy || !directory?.can_write}>{busy ? "در حال ذخیره…" : "ذخیره تنظیمات درگاه"}</Button></DialogActions>
    </Dialog>
  </section>;
}
