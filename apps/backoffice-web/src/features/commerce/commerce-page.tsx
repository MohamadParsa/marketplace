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
  DialogTitle,
  FormControlLabel,
  MenuItem,
  Tab,
  Tabs,
  TextField,
} from "@mui/material";
import { useWorkspace } from "../../shared/workspace";
import type {
  AdminView,
  Deposit,
  Purchase,
  Settings,
  Method,
  Shipping,
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
    throw Error(
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
}: {
  token: string;
  settingsOnly?: boolean;
}) {
  const { state } = useWorkspace(),
    [data, setData] = useState<AdminView | null>(null),
    [settings, setSettings] = useState<Settings | null>(null),
    [busy, setBusy] = useState(false),
    [error, setError] = useState(""),
    [notice, setNotice] = useState(""),
    [tab, setTab] = useState(0),
    [selected, setSelected] = useState<{
      kind: "deposit" | "purchase" | "refund";
      id: string;
    } | null>(null);
  const running = useRef(false),
    keys = useRef(new Map<string, string>());
  const load = useCallback(async () => {
    const v = await request<AdminView>("/backoffice/commerce", token);
    setData(v);
    if (v.permissions["commerce.settings.read"])
      setSettings(
        await request<Settings>("/backoffice/commerce/settings", token),
      );
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
    await request("/backoffice/commerce/" + path, token, input, key);
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
              : "سفارش‌ها و پرداخت‌های سایت"}
          </h1>
          <p>تأیید واریز ← شارژ کیف پول؛ تأیید سفارش ← مصرف کیف پول</p>
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
          تازه‌سازی
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
          <SettingsForm
            key={settings.version}
            settings={settings}
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
        ) : (
          <Alert severity="warning">دسترسی مشاهده تنظیمات ندارید.</Alert>
        )
      ) : (
        <>
          <Tabs
            value={tab}
            onChange={(_, v: number) => setTab(v)}
            variant="scrollable"
          >
            <Tab label={"سفارش‌ها (" + data.purchases.length + ")"} />
            <Tab label={"اعلام واریز (" + data.deposits.length + ")"} />
            <Tab label={"بازپرداخت (" + data.refunds.length + ")"} />
          </Tabs>
          <section className="panel commerce-list">
            {tab === 0 ? (
              data.purchases.length ? (
                data.purchases
                  .slice()
                  .reverse()
                  .map((p) => (
                    <article key={p.id}>
                      <div>
                        <b>
                          {p.customer_name || "مشتری"} · {p.customer_phone}
                        </b>
                        <small dir="ltr">{p.id}</small>
                        <p>
                          سفارش: {label(p.order_status)} · پرداخت:{" "}
                          {label(p.payment_status)}
                        </p>
                        {p.cancellation_requested && (
                          <Alert severity="warning">
                            درخواست لغو: {p.cancellation_reason}
                          </Alert>
                        )}
                      </div>
                      <div>
                        <strong>{money(p.quote.total_toman)}</strong>
                        <p>{label(p.status)}</p>
                      </div>
                      <Button
                        onClick={() =>
                          setSelected({ kind: "purchase", id: p.id })
                        }
                      >
                        جزئیات سفارش
                      </Button>
                    </article>
                  ))
              ) : (
                <Empty />
              )
            ) : tab === 1 ? (
              data.deposits.length ? (
                data.deposits
                  .slice()
                  .reverse()
                  .map((d) => (
                    <article key={d.id}>
                      <div>
                        <b>
                          {d.customer_name || "مشتری"} · {d.customer_phone}
                        </b>
                        <p>
                          {d.method_name} · {d.date} · مرجع {d.reference}
                        </p>
                        <small>
                          {d.purchase_id
                            ? "مرتبط با سفارش"
                            : "شارژ مستقل کیف پول"}
                        </small>
                      </div>
                      <div>
                        <strong>{money(d.amount_toman)}</strong>
                        <p>{label(d.status)}</p>
                      </div>
                      <Button
                        onClick={() =>
                          setSelected({ kind: "deposit", id: d.id })
                        }
                      >
                        مشاهده و بررسی واریز
                      </Button>
                    </article>
                  ))
              ) : (
                <Empty />
              )
            ) : data.refunds.length ? (
              data.refunds
                .slice()
                .reverse()
                .map((r) => (
                  <article key={r.id}>
                    <div>
                      <b>{r.customer_name || "مشتری"}</b>
                      <p>{r.instructions}</p>
                    </div>
                    <div>
                      <strong>{money(r.amount_toman)}</strong>
                      <p>{label(r.status)}</p>
                    </div>
                    <Button
                      onClick={() => setSelected({ kind: "refund", id: r.id })}
                    >
                      رسیدگی بازپرداخت
                    </Button>
                  </article>
                ))
            ) : (
              <Empty />
            )}
          </section>
          <section className="panel">
            <h2>رویدادهای مالی مجاز</h2>
            {data.audit
              .slice(-30)
              .reverse()
              .map((a, i) => (
                <p key={a.id ?? i}>
                  {a.action} · {a.target}
                </p>
              ))}
          </section>
        </>
      )}
      <Dialog
        open={!!selected}
        onClose={() => {
          if (!busy) setSelected(null);
        }}
        fullWidth
        maxWidth="md"
      >
        <DialogTitle>
          {selectedDeposit
            ? "بررسی اعلام واریز"
            : selectedPurchase
              ? "جزئیات سفارش"
              : "رسیدگی بازپرداخت"}
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
          {selectedPurchase && (
            <PurchaseReview
              key={selectedPurchase.id + ":" + selectedPurchase.version}
              purchase={selectedPurchase}
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
          <Button disabled={busy} onClick={() => setSelected(null)}>
            بستن
          </Button>
        </DialogContent>
      </Dialog>
    </>
  );
}
function Empty() {
  return <p className="empty-state">هنوز رکوردی ثبت نشده است.</p>;
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
        مرجع اعلام‌شده: {d.reference} · تاریخ {d.date}
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
          <Field
            label="تاریخ واقعی واریز"
            type="date"
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
              disabled={!checked || !amount || !ref || busy}
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
function PurchaseReview({
  purchase: p,
  busy,
  allowed,
  submit,
}: {
  purchase: Purchase;
  busy: boolean;
  allowed: boolean;
  submit: (cancel: boolean, reason: string) => void;
}) {
  const [reason, setReason] = useState(""),
    [checked, setChecked] = useState(false);
  return (
    <>
      <States p={p} />
      <div className="commerce-status">
        <div>
          مبلغ نهایی<strong>{money(p.quote.total_toman)}</strong>
        </div>
        <div>
          اعتبار قابل استفاده<strong>{money(p.wallet_available_toman)}</strong>
        </div>
      </div>
      <p>کسری: {money(p.shortfall_toman)}</p>
      {p.quote.lines.map((l) => (
        <p key={l.offer_id}>
          {l.name} · {l.variant_label} · تعداد {l.quantity} ·{" "}
          {money(l.total_toman)}
        </p>
      ))}
      {p.quote.shipments.map((s) => (
        <p key={s.store_id}>
          {s.store_name} · {s.method_name} · {money(s.fee_toman)} ·{" "}
          {s.delivery_from} تا {s.delivery_to}
        </p>
      ))}
      <h3>مقصد</h3>
      <p>
        {p.quote.address.recipient} · {p.quote.address.phone}
      </p>
      <p>
        {p.quote.address.province}، {p.quote.address.city}،{" "}
        {p.quote.address.street} · {p.quote.address.postal_code}
      </p>
      {p.cancellation_requested && (
        <Alert severity="warning">
          درخواست لغو مشتری: {p.cancellation_reason}
        </Alert>
      )}
      <h3>واریزهای مرتبط</h3>
      {p.deposits.map((d) => (
        <p key={d.id}>
          {money(d.amount_toman)} · {label(d.status)} · {d.reference}
        </p>
      ))}
      {allowed && p.status !== "cancelled" && (
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
    [date, setDate] = useState(new Date().toISOString().slice(0, 10)),
    [reason, setReason] = useState(""),
    [checked, setChecked] = useState(false);
  return (
    <>
      <h3>
        {money(r.amount_toman)} · {label(r.status)}
      </h3>
      <p>{r.instructions}</p>
      <p>
        {r.reason} · {r.reference}
      </p>
      {allowed && r.status === "pending" && (
        <fieldset disabled={busy}>
          <Alert severity="warning">
            انتقال بانکی خودکار نیست؛ فقط پس از انجام واقعی پرداخت، نتیجه را ثبت
            کنید. تکمیل، مبلغ رزروشده را از کیف پول کم می‌کند.
          </Alert>
          <Field label="مرجع پرداخت انجام‌شده" value={ref} onChange={setRef} />
          <Field
            label="تاریخ پرداخت"
            type="date"
            value={date}
            onChange={setDate}
          />
          <Field label="نتیجه و دلیل" value={reason} onChange={setReason} />
          <FormControlLabel
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
            disabled={busy || !checked || !ref || !reason.trim()}
            onClick={() =>
              submit({
                id: r.id,
                version: r.version,
                complete: true,
                reference: ref,
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
  settings: s,
  permissions,
  categories,
  stores,
  busy,
  save,
}: {
  settings: Settings;
  permissions: Record<string, boolean>;
  categories: { id: string; name: string }[];
  stores: { id: string; name: string }[];
  busy: boolean;
  save: (body: unknown) => Promise<void>;
}) {
  const [methods, setMethods] = useState(s.methods),
    [destination, setDestination] = useState(s.destination),
    [shipping, setShipping] = useState(s.shipping),
    [days, setDays] = useState(s.closed_weekdays),
    [holidays, setHolidays] = useState(s.holidays.join("\n")),
    [business, setBusiness] = useState(s.business);
  const can = (key: string) => !!permissions["commerce." + key];
  const patchMethod = (i: number, patch: Partial<Method>) =>
    setMethods(
      methods.map((m, index) => (index === i ? { ...m, ...patch } : m)),
    );
  const patchShipping = (i: number, patch: Partial<Shipping>) =>
    setShipping(
      shipping.map((m, index) => (index === i ? { ...m, ...patch } : m)),
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
    <>
      {section(
        "روش‌های پرداخت و مقصد وجه",
        <fieldset disabled={busy || !can("payment.settings.write")}>
          <Alert severity="info">
            تمام مبالغ تومان‌اند. سقف روی کل سفارش شامل ارسال سنجیده می‌شود.
            استثنای یک دسته، همه زیرشاخه‌ها را پوشش می‌دهد. برای پرداخت محلی،
            ریشه طلای واقعی را در مستثناها انتخاب کنید.
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
                  onChange={(min_toman) => patchMethod(i, { min_toman })}
                />
                <Field
                  label="حداکثر (خالی = بدون سقف)"
                  value={m.max_toman}
                  onChange={(max_toman) => patchMethod(i, { max_toman })}
                />
              </div>
              <FormControlLabel
                control={
                  <Checkbox
                    checked={m.all_categories}
                    onChange={(e) =>
                      patchMethod(i, { all_categories: e.target.checked })
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
                onChange={(instructions) => patchMethod(i, { instructions })}
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
              onChange={(holder) => setDestination({ ...destination, holder })}
            />
            <Field
              label="بانک"
              value={destination.bank}
              onChange={(bank) => setDestination({ ...destination, bank })}
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
      {section(
        "روش‌های ارسال و تقویم کاری",
        <fieldset disabled={busy || !can("shipping.write")}>
          {shipping.map((m, i) => (
            <article className="commerce-setting-card" key={m.id}>
              <div className="commerce-grid">
                <Field
                  label="نام ارسال"
                  value={m.name}
                  onChange={(name) => patchShipping(i, { name })}
                />
                <Field
                  label="هزینه هر فروشگاه (تومان)"
                  value={m.fee_toman}
                  onChange={(fee_toman) => patchShipping(i, { fee_toman })}
                />
                <Field
                  label="حداقل روز پس از تأمین"
                  type="number"
                  value={String(m.min_days)}
                  onChange={(v) => patchShipping(i, { min_days: Number(v) })}
                />
                <Field
                  label="حداکثر روز پس از تأمین"
                  type="number"
                  value={String(m.max_days)}
                  onChange={(v) => patchShipping(i, { max_days: Number(v) })}
                />
              </div>
              <FormControlLabel
                control={
                  <Checkbox
                    checked={m.active}
                    onChange={(e) =>
                      patchShipping(i, { active: e.target.checked })
                    }
                  />
                }
                label="فعال"
              />
              <FormControlLabel
                control={
                  <Checkbox
                    checked={m.all_cities}
                    onChange={(e) =>
                      patchShipping(i, { all_cities: e.target.checked })
                    }
                  />
                }
                label="همه شهرها"
              />
              {!m.all_cities && (
                <Field
                  label="شهرهای مجاز، جداشده با ویرگول"
                  value={m.cities.join("،")}
                  onChange={(v) =>
                    patchShipping(i, {
                      cities: v
                        .split(/[,،]/)
                        .map((x) => x.trim())
                        .filter(Boolean),
                    })
                  }
                />
              )}
              <Multiple
                caption="فروشگاه‌های مجاز (خالی = همه)"
                value={m.store_ids ?? []}
                options={stores}
                onChange={(store_ids) => patchShipping(i, { store_ids })}
              />
            </article>
          ))}
          <Button
            disabled={busy || !can("shipping.write")}
            onClick={() =>
              setShipping([
                ...shipping,
                {
                  id: "ship-" + crypto.randomUUID(),
                  name: "روش ارسال جدید",
                  active: false,
                  all_cities: true,
                  cities: [],
                  store_ids: [],
                  fee_toman: "0",
                  min_days: 1,
                  max_days: 3,
                },
              ])
            }
          >
            افزودن روش ارسال
          </Button>
          <h3>روزهای تعطیل هفته</h3>
          {[
            "یکشنبه",
            "دوشنبه",
            "سه‌شنبه",
            "چهارشنبه",
            "پنج‌شنبه",
            "جمعه",
            "شنبه",
          ].map((d, i) => (
            <FormControlLabel
              key={i}
              control={
                <Checkbox
                  checked={days.includes(i)}
                  onChange={(e) =>
                    setDays(
                      e.target.checked
                        ? [...days, i]
                        : days.filter((v) => v !== i),
                    )
                  }
                />
              }
              label={d}
            />
          ))}
          <Field
            label="تعطیلات استثنایی؛ هر خط یک تاریخ YYYY-MM-DD"
            value={holidays}
            onChange={setHolidays}
            multiline
          />
          <Button
            variant="contained"
            disabled={busy || !can("shipping.write")}
            onClick={() =>
              void save({
                version: s.version,
                shipping: {
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
      {section(
        "اطلاعات عمومی و قوانین کسب‌وکار",
        <fieldset disabled={busy || !can("business.write")}>
          <Alert severity="warning">
            متن‌ها و اطلاعات واقعی باید پیش از انتشار عمومی توسط تیم تکمیل شوند؛
            نشان مجوز یا تعهد ساختگی نمایش داده نمی‌شود.
          </Alert>
          {(
            [
              ["name", "نام کسب‌وکار"],
              ["support_phone", "شماره پشتیبانی"],
              ["support_email", "ایمیل پشتیبانی"],
              ["about", "درباره ما"],
              ["contact", "تماس با ما"],
              ["shipping_policy", "شرایط ارسال"],
              ["return_policy", "شرایط بازگشت"],
              ["privacy", "حریم خصوصی"],
              ["terms", "قوانین خرید"],
            ] as const
          ).map(([key, caption]) => (
            <Field
              key={key}
              label={caption}
              value={business[key]}
              onChange={(v) => setBusiness({ ...business, [key]: v })}
              multiline={
                !["name", "support_phone", "support_email"].includes(key)
              }
            />
          ))}
          <Button
            variant="contained"
            disabled={busy || !can("business.write")}
            onClick={() => void save({ version: s.version, business })}
          >
            ذخیره محتوای عمومی
          </Button>
        </fieldset>,
      )}
    </>
  );
}

export function SellerCheckoutPage({ token }: { token: string }) {
  const [rows, setRows] = useState<
      | {
          id: string;
          store_id: string;
          city: string;
          status: string;
          payment_status: string;
          shipping_method: string;
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
        <Button onClick={() => void load()}>تازه‌سازی</Button>
      </div>
      {error && <Alert severity="error">{error}</Alert>}
      <section className="panel commerce-list">
        {rows === null ? (
          <p>دریافت سفارش…</p>
        ) : !rows.length ? (
          <p>سفارشی در محدوده مجاز شما وجود ندارد.</p>
        ) : (
          rows.map((o) => (
            <article key={o.id}>
              <div>
                <b>
                  {o.city} · {o.shipping_method}
                </b>
                <p>
                  وضعیت سفارش: {label(o.status)} · وضعیت پرداخت:{" "}
                  {label(o.payment_status)}
                </p>
                {o.items.map((i, index) => (
                  <p key={index}>
                    {i.name} · تعداد {i.quantity} · {money(i.unit_price_toman)}
                  </p>
                ))}
              </div>
            </article>
          ))
        )}
      </section>
    </>
  );
}
