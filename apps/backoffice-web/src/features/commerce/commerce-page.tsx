import { EditorTabs } from "../../shared/editor-tabs";
import { PaymentGatewaySettings } from "./payment-gateway-settings";
import { normalizeTransferReference, transferReferenceError } from "../../../../../packages/commerce-client/transfer-reference";
import { PersianDateField, PersianDatesField } from "../../../../../packages/date-controls/persian-date-field";
import { tehranCapacityDate, persianCapacityDate, persianDigits, validISODate } from "../../../../../packages/date-controls/calendar";
import { PurchaseRequests, PurchaseHistory } from "../operations/order-requests-panel";
import { OrderOperationsPanel } from "../operations/order-operations-panel";
import { MultiChoiceField } from "../../shared/option-select";
import {
  deliveryPromise,
  shippingCharge,
} from "../../../../../packages/commerce-client/shipping-presentation";
import { SupplyOverduePanel } from "../operations/supply-overdue-panel";
import { Link } from "react-router-dom";
import {
  BusinessContentEditor,
  businessDraft,
} from "./business-content-editor";
import { GeographyEditor } from "./geography-editor";
import {
  useCallback,
  useEffect,
  useRef,
  useState,
  type ReactNode,
} from "react";
import {
  Alert,
  Button,
  Checkbox,
  Dialog,
  DialogContent,
  DialogActions,
  DialogTitle,
  FormControlLabel,
  MenuItem,
  Tab,
  Tabs,
  TextField,
} from "@mui/material";
import { useWorkspace } from "../../shared/workspace";
import { DataTable } from "../../shared/components";
import type {
  AdminView,
  Deposit,
  Purchase,
  Settings,
  Method,
  Refund,
} from "../../../../../packages/commerce-client/types";
import { label, money } from "../../../../../packages/commerce-client/types";
const base = import.meta.env.VITE_API_BASE_URL ?? "";
async function request<T>(
  path: string,
  token: string,
  body?: unknown,
  key?: string,
  method?: string,
): Promise<T> {
  const r = await fetch(base + path, {
    method: method ?? (body ? "POST" : "GET"),
    cache: "no-store",
    headers: {
      Authorization: "Bearer " + token,
      "Content-Type": "application/json",
      ...(body ? { "Idempotency-Key": key ?? crypto.randomUUID() } : {}),
    },
    body: body ? JSON.stringify(body) : undefined,
  });
  if (!r.ok) {
    const e = await r.json().catch(() => ({}));
    const fieldMessages:Record<string,string>={
      invalid_transfer_reference:"کد پیگیری روی رسید بانک معتبر نیست؛ ۳ تا ۸۰ نویسه شامل عدد، حروف لاتین یا خط تیره وارد کنید.",
      invalid_transfer_date:"تاریخ واقعی انتقال را وارد کنید؛ تاریخ آینده معتبر نیست.",
      refund_reason_required:"شرح ثبت نتیجه انتقال بانکی را وارد کنید.",
    };
    throw Error(
      fieldMessages[e.code] ??
      (
        {
          401: "نشست منقضی شده است؛ دوباره وارد شوید.",
          403: "مجوز این عملیات را ندارید.",
          409:
            e.code === "insufficient_wallet"
              ? "اعتبار قابل استفاده مشتری کافی نیست."
              : "اطلاعات تغییر کرده یا عملیات قبلاً انجام شده؛ وضعیت را تازه کنید.",
          400: "اطلاعات یا محدودیت‌های روش پرداخت معتبر نیست.",
        } as Record<number, string>
      )[r.status] ?? "دریافت یا ثبت انجام نشد؛ دوباره تلاش کنید.",
    );
  }
  return r.json();
}
function Receipt({ id, token }: { id: string; token: string }) {
  const [url, setURL] = useState(""),
    [error, setError] = useState("");
  useEffect(() => {
    let active = true,
      object = "";
    fetch(base + "/backoffice/commerce/receipts/" + id, {
      headers: { Authorization: "Bearer " + token },
      cache: "no-store",
    })
      .then(async (r) => {
        if (!r.ok) throw Error("دریافت رسید مجاز نیست یا انجام نشد.");
        const b = await r.blob();
        if (active) {
          object = URL.createObjectURL(b);
          setURL(object);
        }
      })
      .catch((e) => {
        if (active) setError(e.message);
      });
    return () => {
      active = false;
      if (object) URL.revokeObjectURL(object);
    };
  }, [id, token]);
  return error ? (
    <Alert severity="error">{error}</Alert>
  ) : url ? (
    <a target="_blank" href={url} rel="noreferrer">
      <img src={url} className="commerce-receipt" alt="تصویر رسید خصوصی" />
    </a>
  ) : (
    <p>دریافت رسید…</p>
  );
}
function States({ p }: { p: Purchase }) {
  return (
    <>
      <div className="commerce-status">
        <div>
          <small>وضعیت سفارش</small>
          <strong>{label(p.order_status)}</strong>
        </div>
        <div>
          <small>وضعیت پرداخت</small>
          <strong>{label(p.payment_status)}</strong>
        </div>
      </div>
      <p>پیگیری: {label(p.status)}</p>
      <details>
        <summary>تاریخچه سفارش و پرداخت</summary>
        <ol>
          {p.events.map((e, i) => (
            <li key={i}>
              {label(e.action)} · {new Date(e.at).toLocaleString("fa-IR")}
              <small>
                سفارش: {label(e.order_status)} · پرداخت:{" "}
                {label(e.payment_status)}
              </small>
            </li>
          ))}
        </ol>
      </details>
    </>
  );
}
export function CommercePage({
  token,
  settingsOnly = false,
  financialOnly = false,
}: {
  token: string;
  settingsOnly?: boolean;
  financialOnly?: boolean;
}) {
  const { state } = useWorkspace(),
    [data, setData] = useState<AdminView | null>(null),
    [settings, setSettings] = useState<Settings | null>(null),
    [busy, setBusy] = useState(false),
    [error, setError] = useState(""),
    [notice, setNotice] = useState(""),
    [tab, setTab] = useState(financialOnly ? 1 : 0),
    [settingsTab, setSettingsTab] = useState("geography"),
    [selected, setSelected] = useState<{
      kind: "deposit" | "purchase" | "refund";
      id: string;
    } | null>(null);
  const running = useRef(false),
    keys = useRef(new Map<string, string>());
  const load = useCallback(async () => {
    setSettings(null);
    const v = await request<AdminView>("/backoffice/commerce", token);
    setData(v);
    if (v.permissions["commerce.settings.read"]) {
      const next = await request<Settings>(
        "/backoffice/commerce/settings",
        token,
      );
      const timing = next.shipping_timing;
      if (
        !timing ||
        typeof timing.ready_time !== "string" ||
        typeof timing.cutoff_time !== "string" ||
        !Number.isInteger(timing.min_slot_lead_minutes) ||
        typeof timing.allow_same_day_slots !== "boolean"
      ) {
        throw Error(
          "تنظیمات زمان‌بندی از سرور دریافت نشد. نسخهٔ رابط و سرویس باید هماهنگ شوند؛ پس از به‌روزرسانی سرویس، به‌روزرسانی فهرست را بزنید.",
        );
      }
      setSettings(next);
    }
  }, [token]);
  const run = async (
    work: (key: string) => Promise<void>,
    signature = "action",
  ) => {
    if (running.current) return;
    running.current = true;
    setBusy(true);
    setError("");
    setNotice("");
    let key = keys.current.get(signature);
    if (!key) {
      key = crypto.randomUUID();
      keys.current.set(signature, key);
    }
    try {
      await work(key);
      keys.current.delete(signature);
    } catch (e) {
      setError((e as Error).message);
    } finally {
      running.current = false;
      setBusy(false);
    }
  };
  useEffect(() => {
    void run(async () => {
      await load();
    });
  }, [load]);
  const mutation = async (path: string, input: unknown, key: string) => {
    await request(
      "/backoffice/commerce/" + path,
      token,
      input,
      key,
      path === "settings" ? "PATCH" : undefined,
    );
    await load();
    setSelected(null);
    setNotice("عملیات ثبت شد؛ وضعیت جدید را بررسی کنید.");
  };
  const selectedDeposit = data?.deposits.find(
      (d) => selected?.kind === "deposit" && d.id === selected.id,
    ),
    selectedPurchase = data?.purchases.find(
      (p) => selected?.kind === "purchase" && p.id === selected.id,
    ),
    selectedRefund = data?.refunds.find(
      (r) => selected?.kind === "refund" && r.id === selected.id,
    );
  return (
    <>
      <div className="page-heading">
        <div>
          <h1>
            {settingsOnly
              ? "تنظیمات خرید، حمل و کسب‌وکار"
              : financialOnly ? "واریز و برداشت کیف پول" : "سفارش‌ها و پرداخت‌های سایت"}
          </h1>
          <p>
            {settingsOnly
              ? "مدیریت مقصدها، روش‌های پرداخت، تقویم و اطلاعات کسب‌وکار"
              : "تأیید واریز ← شارژ کیف پول؛ تأیید سفارش ← مصرف کیف پول"}
          </p>
        </div>
        <Button
          variant="outlined"
          disabled={busy}
          onClick={() =>
            void run(async () => {
              await load();
            })
          }
        >
          به‌روزرسانی فهرست
        </Button>
      </div>
      {error && <Alert severity="error">{error}</Alert>}
      {notice && <Alert severity="success">{notice}</Alert>}
      {!data ? (
        <section className="panel">
          <p role="status">
            {error ? "اطلاعات در دسترس نیست." : "در حال دریافت…"}
          </p>
        </section>
      ) : settingsOnly ? (
        settings ? (
          <>
            <Button component={Link} to="/sales-methods">
              تنظیم روش‌های فروش و مارکاپ
            </Button>
            <SettingsForm
              token={token}
              key={settings.version}
              settings={settings}
              activeTab={settingsTab}
              onTabChange={setSettingsTab}
              permissions={data.permissions}
              categories={state.categories ?? []}
              stores={state.stores ?? []}
              busy={busy}
              save={(body) =>
                run(async (key) => {
                  await mutation("settings", body, key);
                }, JSON.stringify(body))
              }
            />
          </>
        ) : !data.permissions["commerce.settings.read"] ? (
          <Alert severity="warning">دسترسی مشاهده تنظیمات ندارید.</Alert>
        ) : !error ? (
          <p role="status">در حال دریافت تنظیمات…</p>
        ) : null
      ) : (
        <>
          <Tabs
            value={tab}
            onChange={(_, v: number) => setTab(v)}
            variant="scrollable"
          >
            {!financialOnly && <Tab value={0} label={"سفارش‌ها (" + data.purchases.length + ")"} />}
            <Tab value={1} label={"اعلام واریز (" + data.deposits.length + ")"} />
            <Tab value={2} label={"بازپرداخت (" + data.refunds.length + ")"} />
          </Tabs>
          <section className="panel commerce-list">
            {tab === 0 ? (
              <DataTable
                key="purchases"
                rows={data.purchases.slice().reverse()}
                columns={[
                  { title: "شماره سفارش", render: (p) => `#${p.number}` },
                  {
                    title: "مشتری",
                    render: (p) =>
                      `${p.customer_name || "مشتری"} · ${p.customer_phone || ""}`,
                  },
                  { title: "مبلغ", render: (p) => money(p.quote.total_toman) },
                  {
                    title: "وضعیت سفارش",
                    render: (p) => label(p.order_status),
                  },
                  {
                    title: "وضعیت پرداخت",
                    render: (p) => label(p.payment_status),
                  },
                  {
                    title: "درخواست لغو",
                    render: (p) =>
                      p.cancellation_requested ? "در انتظار رسیدگی" : "ندارد",
                  },
                  {
                    title: "عملیات",
                    render: (p) => (
                      <Button
                        onClick={() =>
                          setSelected({ kind: "purchase", id: p.id })
                        }
                      >
                        جزئیات سفارش
                      </Button>
                    ),
                  },
                ]}
              />
            ) : tab === 1 ? (
              <DataTable
                key="deposits"
                rows={data.deposits.slice().reverse()}
                columns={[
                  {
                    title: "مشتری",
                    render: (d) =>
                      `${d.customer_name || "مشتری"} · ${d.customer_phone || ""}`,
                  },
                  { title: "روش", render: (d) => d.method_name },
                  { title: "تاریخ", render: (d) => persianDigits(persianCapacityDate(d.date)) },
                  { title: "مرجع", render: (d) => d.reference },
                  { title: "مبلغ", render: (d) => money(d.amount_toman) },
                  {
                    title: "نوع شارژ",
                    render: (d) =>
                      d.purchase_id ? "مرتبط با سفارش" : "شارژ مستقل کیف پول",
                  },
                  { title: "وضعیت", render: (d) => label(d.status) },
                  {
                    title: "عملیات",
                    render: (d) => (
                      <Button
                        onClick={() =>
                          setSelected({ kind: "deposit", id: d.id })
                        }
                      >
                        مشاهده و بررسی واریز
                      </Button>
                    ),
                  },
                ]}
              />
            ) : (
              <DataTable
                key="refunds"
                rows={data.refunds.slice().reverse()}
                columns={[
                  { title: "مشتری", render: (r) => r.customer_name || "مشتری" },
                  { title: "مبلغ", render: (r) => money(r.amount_toman) },
                  { title: "وضعیت", render: (r) => label(r.status) },
                  {
                    title: "عملیات",
                    render: (r) => (
                      <Button
                        onClick={() =>
                          setSelected({ kind: "refund", id: r.id })
                        }
                      >
                        رسیدگی بازپرداخت
                      </Button>
                    ),
                  },
                ]}
              />
            )}
          </section>
          <details className="panel commerce-audit">
            <summary>تاریخچه عملیات مالی</summary>
            <p>سابقه ثبت و بررسی عملیات در محدوده دسترسی شما؛ این بخش برای پیگیری است و اقدامی برای تکمیل پرداخت ندارد.</p>
            {data.audit
              .slice(-30)
              .reverse()
              .map((a, i) => (
                <p key={a.id ?? i}>
                  {({"commerce.deposit.review":"بررسی اعلام واریز","commerce.refund.resolve":"ثبت نتیجه برداشت بانکی","commerce.refund.request":"ثبت درخواست برداشت","commerce.purchase.confirm":"ثبت پرداخت سفارش","commerce.purchase.accept":"ثبت سفارش","commerce.purchase.admin-cancel":"لغو سفارش"} as Record<string,string>)[a.action] ?? "ثبت عملیات مالی"} · {data.purchases.find(p => p.id === a.target)?.number ?? data.deposits.find(d => d.id === a.target)?.customer_name ?? data.refunds.find(r => r.id === a.target)?.customer_name ?? "سابقه ثبت‌شده"}
                </p>
              ))}
          </details>
        </>
      )}
      <Dialog
        open={!!selected}
        onClose={() => {
          if (!busy) setSelected(null);
        }}
        fullWidth
        maxWidth={selectedPurchase?"xl":selectedRefund?"sm":"md"}
      >
        <DialogTitle>
          {selectedDeposit
            ? "بررسی اعلام واریز"
            : selectedPurchase
              ? "جزئیات سفارش"
              : "ثبت نتیجه انتقال بانکی"}
        </DialogTitle>
        <DialogContent>
          {error && <Alert severity="error">{error}</Alert>}
          {selectedDeposit && (
            <DepositReview
              key={selectedDeposit.id}
              deposit={selectedDeposit}
              token={token}
              busy={busy}
              allowed={!!data?.permissions["commerce.deposits.review"]}
              submit={(input) =>
                void run(async (key) => {
                  await mutation("deposits/review", input, key);
                }, JSON.stringify(input))
              }
            />
          )}{" "}
          {selectedPurchase && <PurchaseRequests id={selectedPurchase.id} onChanged={load}/>}
          {selectedPurchase && <PurchaseHistory id={selectedPurchase.id} version={selectedPurchase.version}/>}
          {selectedPurchase && (
            <PurchaseReview
              key={selectedPurchase.id + ":" + selectedPurchase.version}
              purchase={selectedPurchase}
              token={token}
              canReadDeposits={!!data?.permissions["commerce.deposits.read"]}
              busy={busy}
              allowed={!!data?.permissions["commerce.orders.confirm"]}
              submit={(cancel, reason) =>
                void run(
                  async (key) => {
                    await mutation(
                      cancel ? "orders/cancel" : "orders/confirm",
                      {
                        id: selectedPurchase.id,
                        version: selectedPurchase.version,
                        reason,
                      },
                      key,
                    );
                  },
                  "order:" +
                    selectedPurchase.id +
                    selectedPurchase.version +
                    cancel +
                    reason,
                )
              }
            />
          )}{" "}
          {selectedRefund && (
            <RefundReview
              key={selectedRefund.id}
              refund={selectedRefund}
              busy={busy}
              allowed={!!data?.permissions["commerce.refunds.resolve"]}
              submit={(input) =>
                void run(async (key) => {
                  await mutation("refunds/resolve", input, key);
                }, JSON.stringify(input))
              }
            />
          )}
        </DialogContent>
        <DialogActions><Button disabled={busy} onClick={() => setSelected(null)}>
            بستن
          </Button></DialogActions>
      </Dialog>
    </>
  );
}
function Field({
  label: caption,
  value,
  onChange,
  ...props
}: {
  label: string;
  value: string;
  onChange: (value: string) => void;
  type?: string;
  multiline?: boolean;
  required?: boolean;
  disabled?: boolean;
}) {
  return (
    <TextField
      label={caption}
      value={value}
      onChange={(e) => onChange(e.target.value)}
      fullWidth
      margin="normal"
      size="small"
      {...props}
    />
  );
}
function DepositReview({
  deposit: d,
  token,
  busy,
  allowed,
  submit,
}: {
  deposit: Deposit;
  token: string;
  busy: boolean;
  allowed: boolean;
  submit: (input: unknown) => void;
}) {
  const [ref, setRef] = useState(""),
    [date, setDate] = useState(d.date),
    [amount, setAmount] = useState(""),
    [reason, setReason] = useState(""),
    [checked, setChecked] = useState(false);
  return (
    <>
      <h3>
        {d.method_name} · {money(d.amount_toman)}
      </h3>
      <p>
        مرجع اعلام‌شده: {d.reference} · تاریخ {persianDigits(persianCapacityDate(d.date))}
      </p>
      {d.destination.card_number && (
        <p>
          مقصد: <b dir="ltr">{d.destination.card_number}</b> ·{" "}
          {d.destination.holder}
        </p>
      )}
      <Alert severity="info">
        تأیید این فرم فقط کیف پول را شارژ می‌کند؛ سفارش جداگانه بررسی و تأیید
        می‌شود.
      </Alert>
      {d.receipt_id ? (
        <Receipt id={d.receipt_id} token={token} />
      ) : (
        <p>تصویر رسید ارائه نشده؛ تطبیق با تراکنش واقعی همچنان الزامی است.</p>
      )}
      {d.status !== "pending" ? (
        <p>
          {label(d.status)} · {d.reason}
        </p>
      ) : allowed ? (
        <fieldset disabled={busy}>
          <Field
            label="مبلغ واقعی تطبیق‌شده (تومان)"
            value={amount}
            onChange={setAmount}
          />
          <Field label="مرجع واقعی تراکنش" value={ref} onChange={setRef} />
          <PersianDateField
            label="تاریخ واقعی واریز"
            required
            value={date}
            onChange={setDate}
          />
          <Field
            label="توضیح یا دلیل رد"
            value={reason}
            onChange={setReason}
            multiline
          />
          <FormControlLabel
            control={
              <Checkbox
                checked={checked}
                onChange={(e) => setChecked(e.target.checked)}
              />
            }
            label="پرداخت واقعی را با سوابق بانکی / پرداخت حضوری تطبیق داده‌ام."
          />
          <div className="commerce-actions">
            <Button
              variant="contained"
              disabled={!checked || !amount || !ref || busy || !validISODate(date)}
              onClick={() =>
                submit({
                  id: d.id,
                  version: d.version,
                  approve: true,
                  reason,
                  verified_amount_toman: amount,
                  verified_reference: ref,
                  verified_date: date,
                })
              }
            >
              تأیید و شارژ کیف پول
            </Button>
            <Button
              color="error"
              disabled={!reason.trim() || busy}
              onClick={() =>
                submit({ id: d.id, version: d.version, approve: false, reason })
              }
            >
              رد با دلیل
            </Button>
          </div>
        </fieldset>
      ) : (
        <Alert severity="warning">مجوز بررسی واریز ندارید.</Alert>
      )}
    </>
  );
}
export function PurchaseReview({
  purchase: p,
  mode = "all",
  token,
  canReadDeposits,
  busy,
  allowed,
  submit,
}: {
  purchase: Purchase;
  mode?: "all" | "payment";
  token: string;
  canReadDeposits: boolean;
  busy: boolean;
  allowed: boolean;
  submit: (cancel: boolean, reason: string) => void;
}) {
  const [reason, setReason] = useState(""),
    [checked, setChecked] = useState(false);
  return (
    <>
      <div className="commerce-order-heading">
        <div>
          <small>شماره سفارش</small>
          <h2>#{p.number}</h2>
        </div>
        <span>{new Date(p.created_at).toLocaleString("fa-IR")}</span>
      </div>
      <States p={p} />
      {!!p.payment_attempts?.length && (
        <section className="commerce-order-box" aria-label="سابقه تلاش‌های پرداخت">
          <h3>سابقه تلاش‌های پرداخت</h3>
          {p.payment_attempts.map((attempt,index) => <div key={attempt.id} className="commerce-order-heading">
            <span>تلاش {(index+1).toLocaleString("fa-IR")} · {attempt.purpose === "wallet-funding" ? "تأمین وجه کیف پول" : "پرداخت سفارش"} · {attempt.method_name}</span>
            <span>{money(attempt.amount_toman)} · {label(attempt.status)}</span>
          </div>)}
          <small>تأیید تأمین وجه فقط کیف پول را شارژ می‌کند؛ پرداخت سفارش با مصرف جداگانه قطعی می‌شود.</small>
        </section>
      )}
      <div className="commerce-order-grid">
        <section
          className="commerce-order-box commerce-payment-box"
          aria-label="اطلاعات پرداخت"
        >
          <h3>اطلاعات پرداخت</h3>
          <dl>
            <div>
              <dt>روش انتخاب‌شده</dt>
              <dd>{p.quote.payment_method.name || "بدون نیاز به تأمین وجه"}</dd>
            </div>
            <div>
              <dt>جمع کالاها</dt>
              <dd>
                {money(p.quote.cash_subtotal_toman ?? p.quote.subtotal_toman)}
              </dd>
            </div>
            {Number(p.quote.markup_toman ?? "0") > 0 && (
              <div>
                <dt>مارکاپ {p.quote.sales_method?.name}</dt>
                <dd>{money(p.quote.markup_toman!)}</dd>
              </div>
            )}
            <div>
              <dt>هزینه ارسال</dt>
              <dd>{money(p.quote.shipping_toman)}</dd>
            </div>
            {p.quote.contract_version === 2 && (
              <>
                <div>
                  <dt>کد / مبلغ تخفیف</dt>
                  <dd>
                    {p.quote.coupon_code || "بدون کد"} ·{" "}
                    {money(p.quote.discount_toman ?? "0")}
                  </dd>
                </div>
                <div>
                  <dt>اعتبار انتخابی کیف پول</dt>
                  <dd>{money(p.quote.wallet_selected_toman ?? "0")}</dd>
                </div>
                <div>
                  <dt>نیاز به تأمین وجه</dt>
                  <dd>{money(p.quote.funding_toman ?? "0")}</dd>
                </div>
              </>
            )}
            <div className="commerce-order-grand">
              <dt>مبلغ نهایی</dt>
              <dd>{money(p.quote.total_toman)}</dd>
            </div>
            <div>
              <dt>اعتبار قابل استفاده</dt>
              <dd>
                {p.wallet_available_toman
                  ? money(p.wallet_available_toman)
                  : "—"}
              </dd>
            </div>
            <div>
              <dt>کسری اعتبار</dt>
              <dd>{p.shortfall_toman ? money(p.shortfall_toman) : "—"}</dd>
            </div>
          </dl>
        </section>
        {p.discount_funding && (
          <section className="commerce-order-box" aria-label="سهم هزینه تخفیف">
            <h3>سهم هزینه تخفیف</h3>
            <p>نسخه قانون: {p.discount_funding.rule_version}</p>
            <p>
              چریوا: <b>{money(p.discount_funding.platform_toman)}</b>
            </p>
            <p>
              فروشنده: <b>{money(p.discount_funding.seller_toman)}</b>
            </p>
          </section>
        )}
        {!!p.credit_sources?.length && (
          <section
            className="commerce-order-box"
            aria-label="منابع اعتبار خرید"
          >
            <h3>منابع اعتبار خرید</h3>
            {p.credit_sources.map((a, i) => (
              <article key={i}>
                <b>
                  {a.kind === "selected-balance"
                    ? "اعتبار قبلی انتخابی"
                    : "وجه جدید تخصیص‌یافته"}
                </b>
                <p>
                  {money(a.amount_toman)} ·{" "}
                  {a.status === "held"
                    ? "رزرو"
                    : a.status === "posted"
                      ? "مصرف‌شده"
                      : "آزادشده"}
                </p>
                <p>
                  {a.title || "اعتبار کیف پول"} ·{" "}
                  {a.withdrawable ? "قابل برداشت" : "فقط قابل خرید"}
                </p>
              </article>
            ))}
          </section>
        )}
        {mode === "all" && <><section className="commerce-order-box" aria-label="گیرنده و آدرس">
          <h3>گیرنده و آدرس</h3>
          <b>{p.quote.address.recipient}</b>
          <p dir="ltr" className="commerce-phone">
            {p.quote.address.phone}
          </p>
          <p>
            {p.quote.address.province}، {p.quote.address.city}،{" "}
            {p.quote.address.full_address || p.quote.address.street}
          </p>
          {p.quote.address.house_number && (
            <p>
              پلاک {p.quote.address.house_number} · واحد{" "}
              {p.quote.address.unit || "—"}
            </p>
          )}
          {p.quote.address.title && <p>عنوان آدرس: {p.quote.address.title}</p>}
          {p.quote.address.recipient_national_id && (
            <p>کد ملی گیرنده: {p.quote.address.recipient_national_id}</p>
          )}
          <small>کد پستی: {p.quote.address.postal_code}</small>
        </section>
        <section
          className="commerce-order-box commerce-order-wide"
          aria-label="اقلام سفارش"
        >
          <h3>
            اقلام سفارش <small>({p.quote.lines.length})</small>
          </h3>
          {p.quote.lines.map((l) => (
            <article className="commerce-order-item" key={l.offer_id}>
              <img src={base + l.image_url} alt={l.name} />
              <div>
                <b>{l.name}</b>
                <small>{l.variant_label || "تنوع پیش‌فرض"}</small>
                <small>قیمت واحد: {money(l.unit_price_toman)}</small>
              </div>
              <div>
                <small>تعداد {l.quantity.toLocaleString("fa-IR")}</small>
                <strong>{money(l.total_toman)}</strong>
              </div>
            </article>
          ))}
        </section>
        <section
          className="commerce-order-box commerce-order-wide"
          aria-label="مرسوله‌ها و ارسال"
        >
          <h3>مرسوله‌ها و ارسال</h3>
          <div className="commerce-order-grid">
            {p.quote.shipments.map((s) => (
              <article key={s.store_id} className="commerce-shipment-card">
                <b>{s.store_name}</b>
                <p>
                  {s.method_name} ·{" "}
                  {shippingCharge(s.collection_mode, s.fee_toman)}
                </p>
                <small>
                  وقفه تأمین: {s.lead_days} روز{" "}
                  {s.lead_day_basis === "calendar" ? "تقویمی" : "کاری"}
                </small>
                <p>
                  {deliveryPromise(s.delivery, s.delivery_from, s.delivery_to)}
                </p>
                <small>
                  {p.quote.lines
                    .filter((l) => l.store_id === s.store_id)
                    .map((l) => l.name)
                    .join("، ")}
                </small>
              </article>
            ))}
          </div>
        </section>
        </>}
        <section
          className="commerce-order-box commerce-order-wide"
          aria-label="واریزهای مرتبط"
        >
          <h3>واریزهای مرتبط</h3>
          {!canReadDeposits ? (
            <p>مجوز مشاهده واریزها را ندارید.</p>
          ) : p.deposits.length ? (
            p.deposits.map((d) => (
              <article className="commerce-shipment-card" key={d.id}>
                <div className="commerce-order-heading">
                  <strong>{money(d.amount_toman)}</strong>
                  <span className={"commerce-deposit-state " + d.status}>
                    {label(d.status)}
                  </span>
                </div>
                <p>
                  {d.method_name} · {persianDigits(persianCapacityDate(d.date))}
                </p>
                <p>
                  شماره پیگیری: <b dir="ltr">{d.reference}</b>
                </p>
                {d.reason && <p>{d.reason}</p>}
                {d.receipt_id && <Receipt id={d.receipt_id} token={token} />}
              </article>
            ))
          ) : (
            <p className="commerce-order-empty">
              هنوز واریزی برای این سفارش ثبت نشده است.
            </p>
          )}
        </section>
      </div>
      {p.cancellation_requested && (
        <Alert severity="warning">
          درخواست لغو مشتری: {p.cancellation_reason}
        </Alert>
      )}
      {allowed && p.status !== "cancelled" && p.status !== "confirmed" && (
        <fieldset disabled={busy}>
          <Field
            label="دلیل لغو / توضیح رسیدگی"
            value={reason}
            onChange={setReason}
          />
          {p.status !== "confirmed" && (
            <>
              <Alert severity="info">
                تأیید نهایی، مبلغ را یک‌بار از کیف پول کم و موجودی رزروشده را
                مصرف می‌کند. رزرو تا{" "}
                {new Date(p.quote.reserved_until).toLocaleString("fa-IR")} معتبر
                است.
              </Alert>
              <FormControlLabel
                control={
                  <Checkbox
                    checked={checked}
                    onChange={(e) => setChecked(e.target.checked)}
                  />
                }
                label="سفارش، تأمین اعتبار و زمان رزرو را بررسی کرده‌ام."
              />
              <Button
                variant="contained"
                disabled={
                  busy ||
                  !checked ||
                  p.status === "requires-reconfirmation" ||
                  p.shortfall_toman !== "0"
                }
                onClick={() => submit(false, reason)}
              >
                تأیید نهایی و پرداخت سفارش
              </Button>
            </>
          )}
          <Button
            color="error"
            disabled={busy || !reason.trim()}
            onClick={() => submit(true, reason)}
          >
            لغو سفارش{p.status === "confirmed" ? " و برگشت وجه به کیف پول" : ""}
          </Button>
        </fieldset>
      )}
    </>
  );
}
function RefundReview({
  refund: r,
  busy,
  allowed,
  submit,
}: {
  refund: Refund;
  busy: boolean;
  allowed: boolean;
  submit: (input: unknown) => void;
}) {
  const [ref, setRef] = useState(""),
    [date, setDate] = useState(tehranCapacityDate),
    [reason, setReason] = useState(""),
    [checked, setChecked] = useState(false);
  const referenceError=transferReferenceError(ref);
  const dateValid=validISODate(date)&&date<=tehranCapacityDate();
  return (
    <>
      <h3>
        {money(r.amount_toman)} · {label(r.status)}
      </h3>
      <p>{r.instructions}</p>
      {(r.reason||r.reference)&&<p>{r.reason}{r.reference&&" · کد پیگیری: "+r.reference}</p>}
      {allowed && r.status === "pending" && (
        <fieldset disabled={busy} className="refund-review-form">
          <Alert severity="info" className="refund-review-wide">
            این درخواست برای برداشت از کیف پول و واریز به حساب مشتری است. پس از انتقال واقعی مبلغ دقیق، کد پیگیری و تاریخ رسید بانک را وارد کنید؛ ثبت تکمیل مبلغ رزروشده را از کیف پول کم می‌کند.
          </Alert>
          <TextField label="کد پیگیری انتقال بانکی" required fullWidth value={ref} onChange={event=>setRef(event.target.value)} error={!!ref&&!!referenceError} helperText={ref&&referenceError?referenceError:"کد روی رسید بانک؛ ۳ تا ۸۰ نویسه شامل عدد، حروف لاتین یا خط تیره."} inputProps={{dir:"ltr",maxLength:80}} />
          <PersianDateField
            label="تاریخ انتقال بانکی"
            required
            value={date}
            onChange={setDate}
            max={tehranCapacityDate()}
          />
          <TextField label="شرح ثبت نتیجه" required value={reason} onChange={event=>setReason(event.target.value)} multiline minRows={2} maxRows={4} className="refund-review-wide" inputProps={{maxLength:1000}} />
          <FormControlLabel
            className="refund-review-wide"
            control={
              <Checkbox
                checked={checked}
                onChange={(e) => setChecked(e.target.checked)}
              />
            }
            label="مبلغ دقیق درخواست را به مقصد تأییدشده بازپرداخت کرده‌ام."
          />
          <Button
            variant="contained"
            disabled={busy || !checked || !!referenceError || !reason.trim() || !dateValid}
            onClick={() =>
              submit({
                id: r.id,
                version: r.version,
                complete: true,
                verified_amount_toman: r.amount_toman,
                reference: normalizeTransferReference(ref),
                date,
                reason,
              })
            }
          >
            ثبت بازپرداخت انجام‌شده
          </Button>
          <Button
            color="error"
            disabled={busy || !reason.trim()}
            onClick={() =>
              submit({ id: r.id, version: r.version, complete: false, reason })
            }
          >
            رد و آزادسازی مبلغ
          </Button>
        </fieldset>
      )}
    </>
  );
}
function SettingsForm({
  token,
  settings: s,
  activeTab,
  onTabChange,
  permissions,
  categories,
  busy,
  save,
}: {
  token: string;
  settings: Settings;
  activeTab: string;
  onTabChange: (tab: string) => void;
  permissions: Record<string, boolean>;
  categories: { id: string; name: string }[];
  stores: { id: string; name: string }[];
  busy: boolean;
  save: (body: unknown) => Promise<void>;
}) {
  const [methods, setMethods] = useState(s.methods),
    [destination, setDestination] = useState(s.destination),
    [shipping] = useState(s.shipping),
    [days, setDays] = useState(s.closed_weekdays ?? []),
    [timing, setTiming] = useState(s.shipping_timing),
    [holidays, setHolidays] = useState((s.holidays ?? []).join("\n")),
    [business, setBusiness] = useState(businessDraft(s.business));
  const can = (key: string) => !!permissions["commerce." + key];
  const patchMethod = (i: number, patch: Partial<Method>) =>
    setMethods(
      methods.map((m, index) => (index === i ? { ...m, ...patch } : m)),
    );
  function Multiple({
    caption,
    value,
    options,
    onChange,
  }: {
    caption: string;
    value: string[];
    options: { id: string; name: string }[];
    onChange: (values: string[]) => void;
  }) {
    return (
      <TextField
        select
        fullWidth
        size="small"
        margin="normal"
        label={caption}
        value={value}
        SelectProps={{ multiple: true }}
        onChange={(e) =>
          onChange(
            typeof e.target.value === "string"
              ? e.target.value.split(",")
              : e.target.value,
          )
        }
      >
        {options.map((o) => (
          <MenuItem key={o.id} value={o.id}>
            {o.name}
          </MenuItem>
        ))}
      </TextField>
    );
  }
  const section = (title: string, children: ReactNode) => (
    <section className="panel commerce-settings">
      <h2>{title}</h2>
      {children}
    </section>
  );
  return (
    <div className="commerce-settings-tabs">
      <EditorTabs
        label="بخش‌های تنظیمات خرید و حمل"
        value={activeTab}
        onChange={onTabChange}
        disabled={busy}
        panels={[
          {
            id: "geography",
            label: "استان‌ها و شهرها",
            content: (
              <>
                <GeographyEditor
                  value={s.geography ?? []}
                  busy={busy}
                  allowed={can("geography.write")}
                  save={(geography) => save({ version: s.version, geography })}
                />
              </>
            ),
          },
          {
            id: "payment",
            label: "پرداخت و مقصد وجه",
            content: (
              <>
                <PaymentGatewaySettings token={token} categories={categories} />
                {section(
                  "روش‌های پرداخت و مقصد وجه",
                  <fieldset disabled={busy || !can("payment.settings.write")}>
                    <Alert severity="info">
                      تمام مبالغ تومان‌اند. سقف روی کل سفارش شامل ارسال سنجیده
                      می‌شود. استثنای یک دسته، همه زیرشاخه‌ها را پوشش می‌دهد.
                      برای پرداخت محلی، ریشه طلای واقعی را در مستثناها انتخاب
                      کنید.
                    </Alert>
                    {methods.map((m, i) => (
                      <article className="commerce-setting-card" key={m.id}>
                        <h3>{m.name}</h3>
                        <FormControlLabel
                          control={
                            <Checkbox
                              checked={m.active}
                              onChange={(e) =>
                                patchMethod(i, { active: e.target.checked })
                              }
                            />
                          }
                          label="فعال"
                        />
                        <div className="commerce-grid">
                          <Field
                            label="نام روش"
                            value={m.name}
                            onChange={(name) => patchMethod(i, { name })}
                          />
                          <Field
                            label="حداقل کل سفارش (تومان)"
                            value={m.min_toman}
                            onChange={(min_toman) =>
                              patchMethod(i, { min_toman })
                            }
                          />
                          <Field
                            label="حداکثر (خالی = بدون سقف)"
                            value={m.max_toman}
                            onChange={(max_toman) =>
                              patchMethod(i, { max_toman })
                            }
                          />
                        </div>
                        <FormControlLabel
                          control={
                            <Checkbox
                              checked={m.all_categories}
                              onChange={(e) =>
                                patchMethod(i, {
                                  all_categories: e.target.checked,
                                })
                              }
                            />
                          }
                          label="همه دسته‌ها به جز موارد مستثنا"
                        />
                        {!m.all_categories && (
                          <Multiple
                            caption="دسته‌های مجاز"
                            value={m.allowed_categories}
                            options={categories}
                            onChange={(allowed_categories) =>
                              patchMethod(i, { allowed_categories })
                            }
                          />
                        )}
                        <Multiple
                          caption="دسته‌های مستثنا"
                          value={m.excluded_categories}
                          options={categories}
                          onChange={(excluded_categories) =>
                            patchMethod(i, { excluded_categories })
                          }
                        />
                        <Field
                          label="راهنمای پرداخت"
                          value={m.instructions}
                          onChange={(instructions) =>
                            patchMethod(i, { instructions })
                          }
                          multiline
                        />
                      </article>
                    ))}
                    <div className="commerce-grid">
                      <Field
                        label="شماره کارت مقصد"
                        value={destination.card_number ?? ""}
                        onChange={(card_number) =>
                          setDestination({ ...destination, card_number })
                        }
                      />
                      <Field
                        label="نام صاحب حساب"
                        value={destination.holder}
                        onChange={(holder) =>
                          setDestination({ ...destination, holder })
                        }
                      />
                      <Field
                        label="بانک"
                        value={destination.bank}
                        onChange={(bank) =>
                          setDestination({ ...destination, bank })
                        }
                      />
                    </div>
                    <Button
                      variant="contained"
                      disabled={busy || !can("payment.settings.write")}
                      onClick={() =>
                        void save({
                          version: s.version,
                          payment: { methods, destination },
                        })
                      }
                    >
                      ذخیره روش‌های پرداخت و مقصد
                    </Button>
                  </fieldset>,
                )}
              </>
            ),
          },
          {
            id: "shipping",
            label: "حمل و تقویم",
            content: (
              <>
                {section(
                  "روش‌های ارسال و تقویم کاری",
                  <fieldset disabled={busy || !can("shipping.write")}>
                    <p>
                      روش‌های عمومی، پوشش شهر/دسته و دریافت هزینه از صفحهٔ
                      روش‌های حمل مدیریت می‌شوند.
                    </p>
                    <Button component={Link} to="/shipping">
                      مدیریت روش‌های حمل
                    </Button>
                    <h3>قواعد محاسبهٔ موعد آماده‌شدن و تحویل</h3>
                    <p>این ساعت‌ها برای محاسبهٔ موعد سفارش‌اند؛ ساعت کاری یا بازهٔ حمل نیستند. روزهای کاری سایت در پایین تعریف می‌شوند و هر فروشگاه می‌تواند روزهای کاری خودش را تنظیم کند.</p>
                    <div className="commerce-grid">
                      <Field
                        label="ساعت آماده‌شدن در روز پایان وقفه (تهران)"
                        value={timing.ready_time}
                        onChange={(ready_time) =>
                          setTiming({ ...timing, ready_time })
                        }
                      />
                      <Field
                        label="مهلت پرداخت برای شروع محاسبه از همان روز"
                        value={timing.cutoff_time}
                        onChange={(cutoff_time) =>
                          setTiming({ ...timing, cutoff_time })
                        }
                      />
                      <Field
                        label="حداقل فرصت از آماده‌شدن تا شروع بازهٔ تحویل (دقیقه)"
                        type="number"
                        value={String(timing.min_slot_lead_minutes)}
                        onChange={(value) =>
                          setTiming({
                            ...timing,
                            min_slot_lead_minutes: Number(value),
                          })
                        }
                      />
                    </div>
                    <FormControlLabel
                      label="اسلات همان روز با رعایت فرصت آماده‌سازی مجاز باشد"
                      control={
                        <Checkbox
                          checked={timing.allow_same_day_slots}
                          onChange={(event) =>
                            setTiming({
                              ...timing,
                              allow_same_day_slots: event.target.checked,
                            })
                          }
                        />
                      }
                    />
                    <p>
                      پایان وقفه در تعطیلی به اولین روز کاری منتقل می‌شود. روز
                      شروع در وقفهٔ مثبت شمرده نمی‌شود. محاسبه از ساعت مهلت به
                      بعد، از اولین روز کاری بعد آغاز می‌شود.
                    </p>
                    <p>مثال: با مهلت ۱۲:۰۰، پرداخت ساعت ۱۳:۰۰ از اولین روز کاری بعد محاسبه می‌شود. اگر پایان وقفه دوشنبه باشد و ساعت آماده‌شدن ۰۹:۰۰، بازهٔ تحویل باید حداقل به‌اندازهٔ فرصت تعیین‌شده پس از آن شروع شود.</p>
                    <h3>روزهای کاری سایت</h3>
                    <MultiChoiceField
                      label="روزهای کاری سایت"
                      value={[0, 1, 2, 3, 4, 5, 6].filter((day) => !days.includes(day)).map(String)}
                      options={[
                        "یکشنبه",
                        "دوشنبه",
                        "سه‌شنبه",
                        "چهارشنبه",
                        "پنج‌شنبه",
                        "جمعه",
                        "شنبه",
                      ].map((name, i) => ({ id: String(i), name }))}
                      disabled={busy || !can("shipping.write")}
                      onChange={(ids) => setDays([0, 1, 2, 3, 4, 5, 6].filter((day) => !ids.includes(String(day))))}
                    />
                    <p>فروشگاه‌ها به‌صورت پیش‌فرض از این روزها استفاده می‌کنند؛ تغییر روزهای کاری هر فروشگاه از «وقفه و تقویم فروشگاه» است. تعطیلات تاریخ‌دار سایت در هر دو حالت اعمال می‌شوند.</p>
                    <PersianDatesField
                      label="تعطیلات استثنایی سایت"
                      values={holidays ? holidays.split("\n") : []}
                      onChange={values=>setHolidays(values.join("\n"))}
                      disabled={busy || !can("shipping.write")}
                    />
                    <Button
                      variant="contained"
                      disabled={busy || !can("shipping.write") || holidays.split("\n").filter(Boolean).some(day=>!validISODate(day))}
                      onClick={() =>
                        void save({
                          version: s.version,
                          shipping: {
                            timing,
                            methods: shipping,
                            closed_weekdays: days,
                            holidays: holidays
                              .split("\n")
                              .map((v) => v.trim())
                              .filter(Boolean),
                          },
                        })
                      }
                    >
                      ذخیره ارسال و تقویم
                    </Button>
                  </fieldset>,
                )}
              </>
            ),
          },
          {
            id: "business",
            label: "کسب‌وکار و قوانین",
            content: (
              <>
                {section(
                  "اطلاعات عمومی و قوانین کسب‌وکار",
                  <fieldset disabled={busy || !can("business.write")}>
                    <BusinessContentEditor
                      value={business}
                      onChange={setBusiness}
                    />
                    <Button
                      variant="contained"
                      disabled={busy || !can("business.write")}
                      onClick={() =>
                        void save({ version: s.version, business })
                      }
                    >
                      ذخیره محتوای عمومی
                    </Button>
                  </fieldset>,
                )}
              </>
            ),
          },
        ]}
      />
    </div>
  );
}

export function SellerCheckoutPage({ token }: { token: string }) {
  const [selectedOrder,setSelectedOrder] = useState<string | null>(null);
  const [rows, setRows] = useState<
      | {
          id: string;
          store_id: string;
          city: string;
          status: string;
          order_status: string;
          payment_status: string;
          shipping_method: string;
          collection_mode?: string;
          shipping_fee_toman?: string;
          delivery?: import("../../../../../packages/commerce-client/types").DeliveryPlan;
          delivery_from: string;
          delivery_to: string;
          items: { name: string; quantity: number; unit_price_toman: string }[];
        }[]
      | null
    >(null),
    [error, setError] = useState("");
  const load = useCallback(async () => {
    try {
      setRows(await request("/seller/checkout/orders", token));
      setError("");
    } catch (e) {
      setError((e as Error).message);
    }
  }, [token]);
  useEffect(() => {
    void load();
  }, [load]);
  return (
    <>
      <div className="page-heading">
        <div>
          <h1>سفارش‌های فروشگاه</h1>
          <p>نمایش فقط با عضویت و مجوز مشاهده سفارش همان فروشگاه</p>
        </div>
        <Button onClick={() => void load()}>به‌روزرسانی فهرست</Button>
      </div>
      <SupplyOverduePanel />
      {error && <Alert severity="error">{error}</Alert>}
      <section className="panel commerce-list">
        {rows === null ? (
          <p>دریافت سفارش…</p>
        ) : !rows.length ? (
          <p>سفارشی در محدوده مجاز شما وجود ندارد.</p>
        ) : (
          rows.map((o) => (
            <article key={o.id} className="seller-order-card">
              <div>
                <b>
                  {o.city} · {o.shipping_method}
                </b>
                <p>
                  {shippingCharge(
                    o.collection_mode,
                    o.shipping_fee_toman ?? "0",
                  )}
                </p>
                {o.delivery_from && (
                  <p>
                    {deliveryPromise(
                      o.delivery,
                      o.delivery_from,
                      o.delivery_to,
                    )}
                  </p>
                )}
                <p>
                  وضعیت سفارش: {label(o.order_status)} · عملیات: {label(o.status)} · وضعیت پرداخت:{" "}
                  {label(o.payment_status)}
                </p>
                <Button onClick={() => setSelectedOrder(current => current === o.id ? null : o.id)}>{selectedOrder === o.id ? "بستن عملیات" : "مشاهده و عملیات سفارش"}</Button>
                {o.items.map((i, index) => (
                  <p key={index}>
                    {i.name} · تعداد {i.quantity} · {money(i.unit_price_toman)}
                  </p>
                ))}
              </div>
              {selectedOrder === o.id && <OrderOperationsPanel id={o.id} onChanged={load} />}
            </article>
          ))
        )}
      </section>
    </>
  );
}
