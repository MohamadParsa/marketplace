import { CustomerRequests } from "./order-requests";
import { PersianDateField } from "../../../packages/date-controls/persian-date-field";
import { tehranCapacityDate, persianCapacityDate, persianDigits } from "../../../packages/date-controls/calendar";
import { CustomerOrderHistory, CustomerShipments } from "./order-shipments";
import { CustomerPurchaseProgress, CustomerOrderFinance } from "./customer-order-details";
import { OrderItems } from "./order-items";
import { WalletTransactions } from "./wallet-transactions";
import { ShippingChoiceCards } from "./shipping-choice-cards";
import { deliveryPromise, shippingCharge } from "../../../packages/commerce-client/shipping-presentation";
import { usePublicBusiness, businessPages } from "./business-content";
import {TemporaryPassword} from "../../../packages/identity-client/temporary-password";
import { AddressEditor } from "./address-editor";
import { CatalogImage } from "./catalog-image";
import {
  CheckoutAdjustmentBox,
  type AdjustmentFeedback,
} from "./checkout-adjustment-box";
import { dispatchTime } from "./dispatch-time";
import { BusyRegion, LoadingPlaceholder } from "./busy-region";
import { ToggleSwitch } from "./toggle-switch";
import { ChoiceGroup } from "./choice-group";
import { PurchaseControl } from "./purchase-control";
import {
  useCallback,
  useEffect,
  useRef,
  useState,
  type FormEvent,
  type ReactNode,
} from "react";
import { SiteHeader, SiteFooter } from "./home-page";
import { Modal } from "./components";
import { navigate } from "./navigation";
import {
  apiBase,
  call,
  CommerceError,
  operationKey,
  useCommerce,
} from "./commerce";
import {
  label,
  money,
  type Address,
  type Cart,
  type Deposit,
  type Purchase,
  type Prepare,
  type Preview,
  type Quote,
  type QuoteInput,
  type ShippingAvailability,
  type DeliverySelection,
  type Wallet,
  type Method,
  type Province,
} from "../../../packages/commerce-client/types";
export function useAction() {
  const [busy, setBusy] = useState(false),
    [error, setError] = useState(""),
    [notice, setNotice] = useState("");
  const running = useRef(false),
    keys = useRef(new Map<string, string>());
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
      key = operationKey();
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
  return { busy, error, notice, setNotice, run, setError };
}
function Feedback({ action }: { action: ReturnType<typeof useAction> }) {
  return (
    <>
      {action.error && (
        <p className="buy-alert" role="alert">
          {action.error}
        </p>
      )}
      {action.notice && (
        <p className="buy-notice" role="status">
          {action.notice}
        </p>
      )}
    </>
  );
}
export function CommerceShell({
  title,
  children,
}: {
  title: string;
  children: ReactNode;
}) {
  useEffect(() => {
    document.title = title + " | Cherriva";
    scrollTo(0, 0);
    let meta = document.querySelector<HTMLMetaElement>('meta[name="robots"]');
    if (!meta) {
      meta = document.createElement("meta");
      meta.name = "robots";
      document.head.append(meta);
    }
    meta.content = "noindex,nofollow";
  }, [title]);
  return (
    <div className="launch-home">
      <SiteHeader />
      <main className="lh-wrap buy-page">
        <nav className="buy-breadcrumb" aria-label="مسیر صفحه">
          <a href="/">خانه</a>
          <span className="breadcrumb-separator" dir="ltr" aria-hidden="true" />
          <span>{title}</span>
        </nav>
        <h1>{title}</h1>
        {children}
      </main>
      <SiteFooter categories={[]} />
    </div>
  );
}
function Input({
  name,
  label: caption,
  ...props
}: {
  name: string;
  label: string;
} & React.InputHTMLAttributes<HTMLInputElement>) {
  return (
    <label className="buy-field">
      {caption}
      <input name={name} {...props} />
    </label>
  );
}
function values(event: FormEvent<HTMLFormElement>) {
  event.preventDefault();
  return Object.fromEntries(new FormData(event.currentTarget)) as Record<
    string,
    string
  >;
}
const digits = (s: string) =>
  s.replace(/[۰-۹٠-٩]/g, (c) =>
    String(
      c.charCodeAt(0) - (c >= "۰" ? "۰".charCodeAt(0) : "٠".charCodeAt(0)),
    ),
  );
function safeReturn() {
  const u = new URLSearchParams(location.search).get("return");
  return u &&
    u.startsWith("/") &&
    !u.startsWith("//") &&
    !u.startsWith("/login")
    ? u
    : "/account";
}
export function LoginPage() {
  const [restricted,setRestricted]=useState<{token:string;user?:{PasswordChangeRequired?:boolean}}|null>(null);
  const returnTo = useRef(safeReturn()).current;
  const { token, login } = useCommerce(),
    action = useAction();
  const [identifier, setIdentifier] = useState(""),
    [mode, setMode] = useState<"password" | "otp">("password"),
    [request, setRequest] = useState<{
      request_id: string;
      test_code?: string;
    } | null>(null),
    [resend, setResend] = useState(0),
    [now, setNow] = useState(Date.now()),
    [recover, setRecover] = useState(false);
  const email = identifier.includes("@");
  useEffect(() => {
    const t = setInterval(() => setNow(Date.now()), 1000);
    return () => clearInterval(t);
  }, []);
  useEffect(() => {
    if (token) navigate(returnTo);
  }, [token]);
  const send = () =>
    action.run(async () => {
      const d = await call<{ request_id: string; test_code?: string }>(
        "/storefront/auth/request-otp",
        "",
        { phone: digits(identifier) },
      );
      setRequest(d);
      setResend(Date.now() + 60000);
    }, "otp:" + identifier);
  if(restricted)return <CommerceShell title="تعیین رمز شخصی"><TemporaryPassword<{token:string}> token={restricted.token} surface="storefront" base={apiBase} onCancel={()=>setRestricted(null)} onComplete={r=>login(r.token).then(()=>navigate(returnTo))}/></CommerceShell>;
  return (
    <CommerceShell title="ورود به چریوا">
      <BusyRegion busy={action.busy}>
        <section className="buy-panel buy-auth">
          <img src="/brand/cherriva-wordmark.svg" alt="Cherriva" width="174" />
          <h2>خوش آمدید</h2>
          <p>شماره موبایل یا ایمیل حساب خود را وارد کنید.</p>
          <Input
            name="identifier"
            label="موبایل یا ایمیل"
            autoComplete="username"
            value={identifier}
            onChange={(e) => {
              setIdentifier(e.target.value);
              setRequest(null);
              setRecover(false);
            }}
            dir="ltr"
            required
          />
          {!email && (
            <div className="buy-tabs">
              <button
                className={mode === "password" ? "selected" : ""}
                onClick={() => {
                  setMode("password");
                  setRequest(null);
                }}
              >
                ورود با رمز
              </button>
              <button
                className={mode === "otp" ? "selected" : ""}
                onClick={() => setMode("otp")}
              >
                ورود با OTP / ثبت‌نام
              </button>
            </div>
          )}
          {recover ? (
            <Recovery
              identifier={identifier}
              onDone={() => setRecover(false)}
            />
          ) : !email && mode === "otp" ? (
            <>
              {request ? (
                <form
                  onSubmit={(e) => {
                    const v = values(e);
                    void action.run(
                      async () => {
                        const r = await call<{ token: string;user?:{PasswordChangeRequired?:boolean} }>(
                          "/storefront/auth/verify-otp",
                          "",
                          {
                            request_id: request.request_id,
                            code: digits(v.code),
                          },
                        );
                        if(r.user?.PasswordChangeRequired){setRestricted(r);return}
 await login(r.token);
                        navigate(returnTo);
                      },
                      "verify:" + request.request_id + v.code,
                    );
                  }}
                >
                  <Input
                    name="code"
                    label="کد تأیید"
                    inputMode="numeric"
                    autoComplete="one-time-code"
                    required
                  />
                  {request.test_code && (
                    <p className="buy-note">
                      کد محیط آزمون: <b dir="ltr">{request.test_code}</b>
                    </p>
                  )}
                  <button className="buy-primary" disabled={action.busy}>
                    تأیید و ورود
                  </button>
                </form>
              ) : (
                <p>حساب جدید با تأیید شماره موبایل ساخته می‌شود.</p>
              )}
              <button
                className={request ? "buy-secondary" : "buy-primary"}
                disabled={action.busy || now < resend || !identifier.trim()}
                onClick={send}
              >
                {now < resend
                  ? `ارسال مجدد تا ${Math.ceil((resend - now) / 1000)} ثانیه`
                  : request
                    ? "ارسال مجدد کد"
                    : "دریافت کد تأیید"}
              </button>
            </>
          ) : (
            <form
              onSubmit={(e) => {
                const v = values(e);
                void action.run(
                  async () => {
                    const r = await call<{ token: string;user?:{PasswordChangeRequired?:boolean} }>(
                      "/storefront/auth/password-login",
                      "",
                      { phone: digits(identifier), password: v.password },
                    );
                    if(r.user?.PasswordChangeRequired){setRestricted(r);return}
 await login(r.token);
                    navigate(returnTo);
                  },
                  "password:" + identifier + v.password,
                );
              }}
            >
              <Input
                name="password"
                label="رمز عبور"
                type="password"
                autoComplete="current-password"
                required
              />
              <button
                className="buy-primary"
                disabled={action.busy || !identifier.trim()}
              >
                ورود با رمز عبور
              </button>
              <button
                type="button"
                className="buy-link"
                onClick={() => setRecover(true)}
              >
                رمز عبور را فراموش کرده‌ام
              </button>
            </form>
          )}
          <Feedback action={action} />
        </section>
      </BusyRegion>
    </CommerceShell>
  );
}
function Recovery({
  identifier,
  onDone,
}: {
  identifier: string;
  onDone: () => void;
}) {
  const [recoveryPhone, setRecoveryPhone] = useState(
    identifier.includes("@") ? "" : identifier,
  );
  const action = useAction(),
    [request, setRequest] = useState<{
      request_id: string;
      test_code?: string;
    } | null>(null);
  return (
    <BusyRegion busy={action.busy}>
      <section>
        <p>بازیابی رمز با کد روی موبایل متصل به حساب انجام می‌شود.</p>
        <Input
          name="recovery_phone"
          label="موبایل حساب برای بازیابی"
          value={recoveryPhone}
          onChange={(e) => setRecoveryPhone(e.target.value)}
          inputMode="tel"
        />
        {!request ? (
          <button
            className="buy-secondary"
            disabled={action.busy}
            onClick={() =>
              void action.run(async () =>
                setRequest(
                  await call(
                    "/storefront/auth/password-recovery/request-otp",
                    "",
                    { phone: digits(recoveryPhone) },
                  ),
                ),
              )
            }
          >
            ارسال کد بازیابی
          </button>
        ) : (
          <form
            onSubmit={(e) => {
              const v = values(e);
              void action.run(async () => {
                await call("/storefront/auth/password-recovery/reset", "", {
                  request_id: request.request_id,
                  code: digits(v.code),
                  password: v.password,
                });
                action.setNotice("رمز ثبت شد؛ با رمز تازه وارد شوید.");
                onDone();
              });
            }}
          >
            <Input name="code" label="کد بازیابی" required />
            {request.test_code && <p>کد آزمون: {request.test_code}</p>}
            <Input
              name="password"
              label="رمز جدید"
              type="password"
              autoComplete="new-password"
              minLength={10}
              required
            />
            <button className="buy-primary" disabled={action.busy}>
              ثبت رمز جدید
            </button>
          </form>
        )}
        <button className="buy-link" onClick={onDone}>
          برگشت به ورود
        </button>
        <Feedback action={action} />
      </section>
    </BusyRegion>
  );
}
function RequireLogin({ children }: { children: ReactNode }) {
  const { token } = useCommerce();
  return token ? (
    <>{children}</>
  ) : (
    <section className="buy-panel">
      <h2>برای ادامه وارد حساب شوید</h2>
      <p>سبد مهمان شما پس از ورود به حساب منتقل می‌شود.</p>
      <a
        className="buy-primary"
        href={
          "/login?return=" +
          encodeURIComponent(location.pathname + location.search)
        }
      >
        ورود و ادامه
      </a>
    </section>
  );
}
export function CartPage() {
  const { cart, cartError, refreshCart, mutate } = useCommerce(),
    action = useAction();
  useEffect(() => {
    void refreshCart().catch(() => {});
  }, [refreshCart]);
  const [edit, setEdit] = useState<string | null>(null),
    [offers, setOffers] = useState<
      {
        id: string;
        store_name: string;
        price: number;
        available: boolean;
        variant_id?: string;
      }[]
    >([]),
    [variants, setVariants] = useState<
      { id: string; values: { attribute_id: string; value: string }[] }[]
    >([]),
    [variant, setVariant] = useState(""),
    [newOffer, setNewOffer] = useState("");
  const loadOffers = async (product: string, variantID: string) => {
    const d = await call<{
      offers: typeof offers;
      variants: typeof variants;
      selected_variant_id: string;
    }>(
      "/storefront/products/" +
        encodeURIComponent(product) +
        (variantID ? "?variant=" + encodeURIComponent(variantID) : ""),
    );
    setVariants(d.variants);
    setOffers(d.offers);
    setVariant(d.selected_variant_id);
    setNewOffer("");
  };
  return (
    <CommerceShell title="سبد خرید">
      <BusyRegion busy={action.busy}>
        <div className="buy-layout">
          <section>
            <Feedback action={action} />
            <button
              className="buy-link"
              disabled={action.busy}
              onClick={() =>
                void action.run(async () => {
                  await refreshCart();
                })
              }
            >
              تازه‌سازی سبد
            </button>
            {cartError && (
              <p role="alert">
                {cartError}{" "}
                <button onClick={() => void refreshCart().catch(() => {})}>
                  تلاش دوباره
                </button>
              </p>
            )}
            {!cart ? (
              <LoadingPlaceholder label="دریافت سبد" />
            ) : cart.items.length ? (
              cart.items.map((item) => (
                <article className="buy-panel buy-cart-row" key={item.offer_id}>
                  <a
                    href={
                      item.product_url +
                      "?variant=" +
                      item.variant_id +
                      "&offer=" +
                      item.offer_id
                    }
                  >
                    <img src={apiBase + item.image_url} alt={item.name} />
                  </a>
                  <div>
                    <a href={item.product_url || undefined}>
                      <h2>{item.name}</h2>
                    </a>
                    <p>{item.variant_label}</p>
                    <p>
                      فروشنده: {item.store_name} ·{" "}
                      {dispatchTime(item.lead_days, item.lead_day_basis)}
                    </p>
                    <strong>{item.inactive ? "غیرقابل خرید" : money(item.unit_price_toman)}</strong>
                    {item.review_required && (
                      <div className="buy-alert" role="status">
                        <p>تنوع این کالا تغییر کرده؛ انتخاب قبلی: {item.previous_variant_label || "بدون محور"} · انتخاب فعلی: {item.variant_label || "بدون محور"}</p>
                        <button className="buy-link" disabled={action.busy || !item.available} onClick={() => void action.run(async () => {
                          await mutate({ action: "review", offer_id: item.offer_id, quantity: item.quantity, version: cart.version, review_variant_version: item.variant_version });
                        })}>تأیید انتخاب فعلی</button>
                      </div>
                    )}
                    {item.price_changed && (
                      <p className="buy-alert">
                        قیمت از {money(item.previous_price_toman)} به‌روز شده؛
                        مبلغ نهایی را در پرداخت تأیید کنید.
                      </p>
                    )}
                    {!item.available && (
                      <p className="buy-alert">
                        {item.inactive ? "این محصول غیرفعال و غیرقابل خرید است؛ آن را از سبد حذف کنید." : item.restricted ? "این فروشنده برای دسته یا فروش نقدی این کالا مجوز ندارد؛ فروشنده را تغییر دهید." : "این تعداد یا عرضه فعلاً قابل خرید نیست؛ تعداد یا فروشنده را اصلاح کنید."}
                      </p>
                    )}
                    <div className="buy-row">
                      <PurchaseControl
                        name={item.name}
                        quantity={item.quantity}
                        canIncrease={item.available}
                        busy={action.busy}
                        onChange={(next) =>
                          void action.run(async () => {
                            await mutate({
                              action: next === 0 ? "remove" : "set",
                              offer_id: item.offer_id,
                              quantity: next,
                              version: cart.version,
                            });
                          })
                        }
                      />
                      <button
                        className="buy-link"
                        disabled={action.busy}
                        onClick={() =>
                          void action.run(async () => {
                            await loadOffers(item.product_id, item.variant_id);
                            setEdit(item.offer_id);
                          })
                        }
                      >
                        تغییر تنوع یا فروشنده
                      </button>
                    </div>
                  </div>
                  <strong>{item.inactive ? "—" : money(item.total_toman)}</strong>
                </article>
              ))
            ) : (
              <div className="buy-panel buy-empty">
                <h2>سبد شما خالی است</h2>
                <a href="/">دیدن محصولات</a>
              </div>
            )}
          </section>
          <aside className="buy-panel buy-summary">
            <h2>خلاصه سبد</h2>
            <p>{cart?.item_count ?? 0} کالا</p>
            <div className="buy-total">
              <span>جمع کالاها</span>
              <b>{money(cart?.subtotal_toman ?? "0")}</b>
            </div>
            <p className="buy-note">
              هزینه ارسال هر فروشنده در مرحله بعد مشخص می‌شود.
            </p>
            {cart?.can_checkout && (
              <a className="buy-primary" href="/checkout">
                ادامه و انتخاب آدرس
              </a>
            )}
            <a className="buy-link" href="/">
              ادامه خرید
            </a>
          </aside>
        </div>
        {edit && cart && (
          <Modal
            busy={action.busy}
            title="انتخاب تنوع و فروشنده"
            onClose={() => setEdit(null)}
          >
            <ChoiceGroup
              label="تنوع"
              value={variant}
              disabled={action.busy}
              options={variants.map((v) => ({
                value: v.id,
                label: v.values.map((x) => x.value).join(" / ") || "پیش‌فرض",
              }))}
              onChange={(id) =>
                void action.run(async () => {
                  await loadOffers(
                    cart.items.find((i) => i.offer_id === edit)!.product_id,
                    id,
                  );
                })
              }
            />
            <ChoiceGroup
              label="فروشنده"
              value={newOffer}
              disabled={action.busy}
              options={offers.map((o) => ({
                value: o.id,
                label: o.store_name,
                detail: o.available ? money(String(o.price / 10)) : "ناموجود",
                disabled: !o.available,
              }))}
              onChange={setNewOffer}
            />
            <p>تغییر تنها با تأیید شما در سبد ثبت می‌شود.</p>
            <button
              className="buy-primary"
              disabled={action.busy || !newOffer}
              onClick={() =>
                void action.run(async () => {
                  await mutate({
                    action: "replace",
                    offer_id: edit,
                    new_offer_id: newOffer,
                    quantity: cart.items.find((i) => i.offer_id === edit)!
                      .quantity,
                    version: cart.version,
                  });
                  setEdit(null);
                })
              }
            >
              تأیید تغییر سبد
            </button>
            <Feedback action={action} />
          </Modal>
        )}
      </BusyRegion>
    </CommerceShell>
  );
}
const emptyAddress: Address = {
  id: "",
  version: 0,
  recipient: "",
  phone: "",
  province: "",
  city: "",
  street: "",
  postal_code: "",
  recipient_self: true,
};
function destinationAvailable(address: Address, geography: Province[]) {
  const norm = (v: string) =>
    v.trim().replace(/ي/g, "ی").replace(/ك/g, "ک").toLowerCase();
  return geography.some(
    (p) =>
      p.active &&
      norm(p.name) === norm(address.province) &&
      p.cities.some((c) => c.active && norm(c.name) === norm(address.city)),
  );
}
export function AddressBook({
  onSelect,
  selected,
}: {
  onSelect?: (id: string, address?: Address) => void;
  selected?: string;
}) {
  const { token, config } = useCommerce(),
    action = useAction(),
    [rows, setRows] = useState<Address[] | null>(null),
    [draft, setDraft] = useState<Address | null>(null);
  const load = useCallback(
    async () =>
      setRows(await call<Address[]>("/storefront/account/addresses", token)),
    [token],
  );
  useEffect(() => {
    if (token)
      void action.run(async () => {
        await load();
      });
  }, [token, load]);
  const geography = config?.geography ?? [];
  return (
    <BusyRegion busy={action.busy}>
      <section>
        <div className="buy-heading">
          <h2>آدرس‌ها</h2>
          <button
            className="buy-secondary"
            onClick={() => setDraft({ ...emptyAddress })}
          >
            افزودن آدرس
          </button>
        </div>
        <Feedback action={action} />
        {rows === null ? (
          <LoadingPlaceholder label="دریافت آدرس‌ها" />
        ) : !rows.length ? (
          <p className="buy-note">هنوز آدرسی ثبت نکرده‌اید.</p>
        ) : (
          rows.map((a) => (
            <article
              key={a.id}
              className={
                "buy-panel buy-address " + (selected === a.id ? "selected" : "")
              }
            >
              {onSelect && (
                <label className="buy-address-choice">
                  <input
                    type="radio"
                    name="address"
                    value={a.id}
                    checked={selected === a.id}
                    aria-describedby={"address-details-" + a.id}
                    disabled={
                      !destinationAvailable(a, geography) ||
                      a.detail_version !== 2
                    }
                    onChange={() => onSelect(a.id, a)}
                  />
                  <span className="buy-address-radio" aria-hidden="true" />
                  <span>
                    {a.title || a.recipient} · {a.city}
                  </span>
                </label>
              )}
              {a.detail_version !== 2 && (
                <p className="buy-alert">
                  این آدرس قدیمی است؛ پلاک و اطلاعات گیرنده را با ویرایش تکمیل
                  کنید.
                </p>
              )}
              {!destinationAvailable(a, geography) && (
                <p className="buy-alert">
                  مقصد این آدرس فعال نیست؛ استان و شهر را ویرایش کنید.
                </p>
              )}
              {!onSelect && (
                <h3>
                  {a.title || a.recipient} · {a.city}
                </h3>
              )}
              <p id={"address-details-" + a.id}>
                {a.full_address || `${a.province}، ${a.street}`}{" "}
                {a.house_number && ` · پلاک ${a.house_number}`}{" "}
                {a.unit && ` · واحد ${a.unit}`}
              </p>
              <p>
                {a.phone} · کد پستی {a.postal_code}
              </p>
              <div className="buy-row">
                <button className="buy-link" onClick={() => setDraft(a)}>
                  ویرایش
                </button>
                <button
                  className="buy-link"
                  disabled={action.busy}
                  onClick={() =>
                    void action.run(async (key) => {
                      await call(
                        "/storefront/account/addresses",
                        token,
                        { id: a.id, version: a.version, delete: true },
                        undefined,
                        key,
                      );
                      await load();
                      if (selected === a.id) onSelect?.("");
                    }, "delete-address:" + a.id)
                  }
                >
                  حذف آدرس
                </button>
              </div>
            </article>
          ))
        )}
        {draft && (
          <Modal
            busy={action.busy}
            className="address-modal"
            title={draft.id ? "ویرایش آدرس" : "آدرس جدید"}
            onClose={() => {
              if (!action.busy) setDraft(null);
            }}
          >
            <AddressEditor
              key={draft.id || "new"}
              initial={draft}
              allowOtherRecipient={!!onSelect}
              geography={geography}
              busy={action.busy}
              error={action.error}
              onSave={(input) => {
                void action.run(async (key) => {
                  const saved = await call<Address>(
                    "/storefront/account/addresses",
                    token,
                    input,
                    undefined,
                    key,
                  );
                  await load();
                  onSelect?.(saved.id, saved);
                  setDraft(null);
                }, JSON.stringify(input));
              }}
            />
          </Modal>
        )}
      </section>
    </BusyRegion>
  );
}
export function CheckoutPage() {
  const { token } = useCommerce();
  return (
    <CommerceShell title="تکمیل خرید">
      <RequireLogin>{token && <CheckoutForm />}</RequireLogin>
    </CommerceShell>
  );
}
function CheckoutForm() {
  const { token, refreshCart, config } = useCommerce(),
    action = useAction();
  const [prepare, setPrepare] = useState<Prepare | null>(null),
    [address, setAddress] = useState<Address | null>(null),
    [shipping, setShipping] = useState<Record<string, string>>({}),
    [shippingSelections, setShippingSelections] = useState<Record<string, DeliverySelection>>({}),
    [shippingChoices, setShippingChoices] = useState<ShippingAvailability[] | null>(null),
    [shippingLoading, setShippingLoading] = useState(false),
    [shippingError, setShippingError] = useState(""),
    [shippingReload, setShippingReload] = useState(0),
    [preview, setPreview] = useState<Preview | null>(null),
    [method, setMethod] = useState(""),
    [quote, setQuote] = useState<Quote | null>(null),
    [consent, setConsent] = useState(false),
    [step, setStep] = useState<1 | 2>(1),
    [couponDraft, setCouponDraft] = useState(""),
    [couponCode, setCouponCode] = useState(""),
    [walletDraft, setWalletDraft] = useState("0"),
    [walletToman, setWalletToman] = useState("0");
  const [adjustmentFeedback, setAdjustmentFeedback] = useState<
    Partial<Record<"coupon" | "wallet", AdjustmentFeedback>>
  >({});
  const adjustmentSequence = useRef(0);
  const heading = useRef<HTMLHeadingElement>(null);
  const purchaseID = new URLSearchParams(location.search).get("purchase") ?? "";
  const [oldPurchase, setOldPurchase] = useState<Purchase | null>(null);
  useEffect(() => {
    void action.run(async () => {
      const p = await call<Prepare>("/storefront/checkout", token);
      setPrepare(p);
      if (purchaseID) {
        const old = await call<Purchase>(
          "/storefront/account/orders/" + purchaseID,
          token,
        );
        setOldPurchase(old);
        setShipping(Object.fromEntries(old.quote.shipments.map(s=>[s.store_id,s.method_id])));
        setShippingSelections(Object.fromEntries(old.quote.shipments.filter(s=>s.delivery?.slot).map(s=>[s.store_id,{date:s.delivery!.capacity_day,window_id:s.delivery!.slot!.window_id}])));

        setAddress(
          p.addresses.find((a) => a.id === old.quote.address.id) ?? null,
        );
      }
    });
  }, [token, purchaseID]);
  useEffect(() => {
    heading.current?.focus();
  }, [step]);
  const change = () => {
    setPreview(null);
    setMethod("");
    setQuote(null);
    setConsent(false);
    setStep(1);
    setCouponDraft("");
    setCouponCode("");
    setWalletDraft("0");
    setWalletToman("0");
    setAdjustmentFeedback({});
  };
  useEffect(() => {
    if (!address || !prepare || (purchaseID && !oldPurchase)) return;
    let active=true;
    setShippingLoading(true);
    setShippingError("");
    const params=new URLSearchParams({address_id:address.id,...(oldPurchase?{purchase_id:oldPurchase.id}:{})});
    void call<ShippingAvailability[]>("/storefront/checkout/shipping-options?"+params, token).then(rows=>{
      if (!active) return;
      setShippingChoices(rows);
      setShipping(old=>Object.fromEntries(Object.entries(old).filter(([store,id])=>rows.some(r=>r.store_id===store && r.methods.some(m=>m.method.id===id)))));
      setShippingSelections(old=>Object.fromEntries(Object.entries(old).filter(([store,selected])=>rows.some(r=>r.store_id===store && r.methods.some(m=>m.days.some(d=>d.date===selected.date && d.slots.some(s=>s.plan.slot?.window_id===selected.window_id)))))));
    }).catch((e:Error)=>{if(active)setShippingError(e instanceof CommerceError ? e.message : "دریافت موعدهای تحویل انجام نشد؛ ارتباط را بررسی و دوباره تلاش کنید.");}).finally(()=>{if(active)setShippingLoading(false);});
    return ()=>{active=false;};
  },[token,address?.id,address?.version,oldPurchase?.id,prepare?.cart.version,shippingReload]);
  const shippingReady = !!shippingChoices?.length && shippingChoices.every(row=>{
    const chosen=row.methods.find(m=>m.method.id===shipping[row.store_id]);
    if (!chosen) return false;
    if (chosen.range) return true;
    const selection=shippingSelections[row.store_id];
    return chosen.days.some(d=>d.open && d.date===selection?.date && d.slots.some(s=>s.plan.slot?.window_id===selection?.window_id));
  });
  const shippingFailure = (error:unknown) => {
    if (error instanceof CommerceError && ["shipping_unavailable","slot_full","slot_expired","checkout_changed","invalid_request","address_unavailable","address_incomplete"].includes(error.code)) {
      change(); setShipping({}); setShippingSelections({}); setShippingReload(n=>n+1);
    }
    throw error;
  };
  const input: QuoteInput = {
    contract_version: 2,
    coupon_code: couponCode,
    wallet_toman: walletToman,
    cart_version: prepare?.cart.version ?? 0,
    address_id: address?.id ?? "",
    shipping,
    shipping_selections: shippingSelections,
    payment_method_id: method,
    ...(purchaseID ? { purchase_id: oldPurchase?.id ?? purchaseID } : {}),
  };
  const resetConsent = () => {
    setQuote(null);
    setConsent(false);
  };
  const adjust = (scope: "coupon" | "wallet", removing = false) => {
    const feedback = (kind: "success" | "error", text: string) =>
      setAdjustmentFeedback((old) => ({
        ...old,
        [scope]: { id: ++adjustmentSequence.current, kind, text },
      }));
    const normalizedMoney = walletDraft
      .trim()
      .replace(/[۰-۹]/g, (c) => String(c.charCodeAt(0) - 1776))
      .replace(/[٠-٩]/g, (c) => String(c.charCodeAt(0) - 1632));
    if (!removing && scope === "coupon" && !couponDraft.trim()) {
      feedback("error", "کد تخفیف را وارد کنید.");
      return;
    }
    if (
      !removing &&
      scope === "wallet" &&
      (!/^\d+(?:\.\d)?$/.test(normalizedMoney) ||
        !/[1-9]/.test(normalizedMoney))
    ) {
      feedback(
        "error",
        "مبلغی معتبر و بیشتر از صفر وارد کنید؛ واحد مبلغ تومان است.",
      );
      return;
    }
    void action.run(async () => {
      resetConsent();
      setAdjustmentFeedback((old) => ({ ...old, [scope]: undefined }));
      const code =
        scope === "coupon" ? (removing ? "" : couponDraft.trim()) : couponCode;
      const wallet =
        scope === "wallet" ? (removing ? "0" : normalizedMoney) : walletToman;
      try {
        const result = await call<Preview>(
          "/storefront/checkout/preview",
          token,
          {
            ...input,
            coupon_code: code,
            wallet_toman: wallet,
            payment_method_id: method,
          },
        );
        setPreview(result);
        setCouponCode(result.quote.coupon_code ?? "");
        setCouponDraft(result.quote.coupon_code ?? "");
        setWalletToman(result.quote.wallet_selected_toman ?? "0");
        setWalletDraft(result.quote.wallet_selected_toman ?? "0");
        feedback(
          "success",
          removing
            ? scope === "coupon"
              ? "کد تخفیف حذف شد."
              : "استفاده از اعتبار کیف پول لغو شد."
            : scope === "coupon"
              ? "کد تخفیف با موفقیت اعمال شد."
              : "اعتبار کیف پول با موفقیت اعمال شد.",
        );
      } catch (error) {
        if (error instanceof CommerceError && ["shipping_unavailable","slot_full","slot_expired","checkout_changed","invalid_request"].includes(error.code)) shippingFailure(error);
        feedback("error", (error as Error).message);
      }
    });
  };
  return (
    <BusyRegion busy={action.busy}>
      <>
        <nav aria-label="مراحل تکمیل خرید">
          <ol className="buy-steps">
            <li
              className={step === 1 ? "active" : "complete"}
              aria-current={step === 1 ? "step" : undefined}
            >
              <button
                type="button"
                disabled={action.busy || step === 1}
                onClick={change}
              >
                <span className="buy-step-number">
                  {step === 2 ? "✓" : "۱"}
                </span>
                <span>
                  آدرس و ارسال
                  <small>
                    {step === 2 ? "تأیید شد" : "انتخاب مقصد و تحویل"}
                  </small>
                </span>
              </button>
            </li>
            <li
              className={step === 2 ? "active" : ""}
              aria-current={step === 2 ? "step" : undefined}
            >
              <span className="buy-step-number">۲</span>
              <span>
                پرداخت و ثبت سفارش
                <small>{step === 2 ? "مرحله جاری" : "پس از تأیید ارسال"}</small>
              </span>
            </li>
          </ol>
        </nav>
        <h2 ref={heading} tabIndex={-1} className="buy-step-title">
          {step === 1 ? "آدرس و روش ارسال" : "روش پرداخت و تأیید سفارش"}
        </h2>
        <Feedback action={action} />
        {!prepare ? (
          <LoadingPlaceholder label="دریافت شرایط خرید" />
        ) : (
          <div className="buy-layout">
            <section className="buy-wizard-stage" key={step}>
              {step === 1 ? (
                <>
                  <fieldset disabled={action.busy}>
                    <AddressBook
                      selected={address?.id}
                      onSelect={(_id, a) => {
                        setAddress(a ?? null);
                        setShipping({});
                        setShippingSelections({});
                        setShippingChoices(null);
                        change();
                      }}
                    />
                  </fieldset>
                  <section className="buy-panel" aria-label="روش‌های ارسال">
                    <h2>ارسال سفارش شما</h2>
                    {!address ? (
                      <p className="buy-note">
                        ابتدا آدرس تحویل را انتخاب کنید تا روش‌های ارسال همان
                        مقصد نمایش داده شوند.
                      </p>
                    ) : (
                      <BusyRegion busy={shippingLoading} label="دریافت موعدهای تحویل">
                        {!shippingChoices ? <LoadingPlaceholder label="دریافت روش‌های ارسال" className="buy-shipping-loading"/> : shippingChoices.map(row=><ShippingChoiceCards
                          key={row.store_id} row={row} methodID={shipping[row.store_id]} selection={shippingSelections[row.store_id]}
                          items={(oldPurchase?.quote.lines ?? prepare.cart.items).filter(i=>i.store_id===row.store_id)}
                          disabled={action.busy || shippingLoading || !!shippingError}
                          onMethod={id=>{setShipping(old=>({...old,[row.store_id]:id}));setShippingSelections(old=>{const copy={...old};delete copy[row.store_id];return copy;});change();}}
                          onSelection={value=>{setShippingSelections(old=>({...old,[row.store_id]:value}));change();}}
                        />)}
                      </BusyRegion>
                    )}
                    {shippingError && <p className="buy-alert" role="alert">{shippingError}</p>}
                    {address && <button type="button" className="buy-secondary" disabled={action.busy || shippingLoading} onClick={()=>{change();setShippingReload(n=>n+1);}}>دریافت دوبارهٔ موعدهای تحویل</button>}
                    <button
                      className="buy-primary"
                      disabled={
                        action.busy ||
                        !address ||
                        !destinationAvailable(
                          address,
                          config?.geography ?? [],
                        ) ||
                        shippingLoading || !!shippingError || !shippingReady
                      }
                      onClick={() =>
                        void action.run(async () => {
                          const result = await call<Preview>(
                            "/storefront/checkout/preview",
                            token,
                            { ...input, payment_method_id: "" },
                          ).catch(shippingFailure);
                          setPreview(result);
                          setMethod("");
                          setQuote(null);
                          setConsent(false);
                          setStep(2);
                        })
                      }
                    >
                      تأیید آدرس و ارسال؛ ادامه به پرداخت
                    </button>
                  </section>
                </>
              ) : (
                preview && (
                  <>
                    <button
                      className="buy-link buy-wizard-back"
                      disabled={action.busy}
                      onClick={change}
                    >
                      ← بازگشت به مرحله آدرس و ارسال
                    </button>
                    <section className="buy-panel buy-delivery-confirmed">
                      <div className="buy-heading">
                        <h2>آدرس و ارسال تأییدشده</h2>
                        <button
                          className="buy-link"
                          disabled={action.busy}
                          onClick={change}
                        >
                          ویرایش آدرس و ارسال
                        </button>
                      </div>
                      <p>
                        {preview.quote.address.recipient} ·{" "}
                        {preview.quote.address.city}،{" "}
                        {preview.quote.address.street}
                      </p>
                      {preview.quote.shipments.map((s) => (
                        <div className="buy-shipment" key={s.store_id}>
                          <b>
                            {s.store_name} · {s.method_name}
                          </b>
                          <p>{deliveryPromise(s.delivery,s.delivery_from,s.delivery_to)}</p>
                          <small>{shippingCharge(s.collection_mode,s.fee_toman)}</small>
                        </div>
                      ))}
                    </section>
                    <CheckoutAdjustmentBox
                      kind="coupon"
                      title="کد تخفیف"
                      busy={action.busy}
                      applied={
                        !!couponCode &&
                        Number(preview.quote.discount_toman ?? "0") > 0
                      }
                      feedback={adjustmentFeedback.coupon}
                      onRemove={() => adjust("coupon", true)}
                      summary={
                        <>
                          مبلغ{" "}
                          <b>{money(preview.quote.discount_toman ?? "0")}</b>{" "}
                          بابت کد <bdi>{couponCode}</bdi> اعمال شده است.
                        </>
                      }
                    >
                      <form
                        onSubmit={(e) => {
                          e.preventDefault();
                          adjust("coupon");
                        }}
                      >
                        <fieldset
                          className="buy-adjustment-form"
                          disabled={action.busy}
                        >
                          <label className="buy-field">
                            کد تخفیف
                            <input
                              aria-label="کد تخفیف"
                              dir="ltr"
                              value={couponDraft}
                              maxLength={64}
                              onChange={(e) => setCouponDraft(e.target.value)}
                            />
                          </label>
                          <button className="buy-secondary" type="submit">
                            اعمال کد تخفیف
                          </button>
                        </fieldset>
                      </form>
                    </CheckoutAdjustmentBox>
                    <CheckoutAdjustmentBox
                      kind="wallet"
                      title="اعتبار کیف پول"
                      subtitle={
                        <>
                          موجودی قابل استفاده:{" "}
                          {money(prepare.wallet_available_toman)}
                        </>
                      }
                      busy={action.busy}
                      applied={Number(walletToman) > 0}
                      feedback={adjustmentFeedback.wallet}
                      onRemove={() => adjust("wallet", true)}
                      summary={
                        <>
                          مبلغ{" "}
                          <b>
                            {money(preview.quote.wallet_selected_toman ?? "0")}
                          </b>{" "}
                          از کیف پول در این خرید استفاده شده است.
                        </>
                      }
                    >
                      <p className="buy-note">
                        مبلغ دلخواه از موجودی قبلی را وارد کنید؛ باقی موجودی
                        مصرف نمی‌شود.
                      </p>
                      <form
                        onSubmit={(e) => {
                          e.preventDefault();
                          adjust("wallet");
                        }}
                      >
                        <fieldset
                          className="buy-adjustment-form"
                          disabled={action.busy}
                        >
                          <label className="buy-field">
                            مبلغ مصرف از کیف پول (تومان)
                            <input
                              aria-label="مبلغ مصرف از کیف پول"
                              inputMode="decimal"
                              value={walletDraft}
                              maxLength={18}
                              onChange={(e) => setWalletDraft(e.target.value)}
                            />
                          </label>
                          <button className="buy-secondary" type="submit">
                            اعمال اعتبار
                          </button>
                        </fieldset>
                      </form>
                    </CheckoutAdjustmentBox>
                    <section
                      className="buy-panel"
                      aria-label="انتخاب روش پرداخت"
                    >
                      <h2>روش پرداخت</h2>
                      {preview.quote.funding_toman === "0" && (
                        <p className="buy-notice">
                          مبلغی برای تأمین وجه باقی نمانده؛ انتخاب روش پرداخت
                          لازم نیست. سفارش پس از ثبت، منتظر تأیید تیم می‌ماند.
                        </p>
                      )}
                      {preview.quote.funding_toman !== "0" &&
                        preview.methods.map((m) => (
                          <label
                            key={m.id}
                            className={
                              "buy-choice " + (!m.eligible ? "disabled" : "")
                            }
                          >
                            <input
                              type="radio"
                              name="payment_method"
                              checked={method === m.id}
                              disabled={!m.eligible || action.busy}
                              onChange={() => {
                                void action.run(async () => {
                                  const updated = await call<Preview>(
                                    "/storefront/checkout/preview",
                                    token,
                                    { ...input, payment_method_id: m.id },
                                  ).catch(shippingFailure);
                                  setPreview(updated);
                                  setMethod(m.id);
                                  setQuote(null);
                                  setConsent(false);
                                });
                              }}
                            />
                            <span>
                              <b>{m.name}</b>
                              {m.sales_method && m.eligible && (
                                <small>
                                  {m.sales_method.name} · مبلغ این روش:{" "}
                                  {money(m.total_toman ?? "0")}
                                </small>
                              )}
                              <small>
                                {m.eligible ? m.instructions : m.reason}
                              </small>
                              {m.max_toman && (
                                <small>
                                  سقف کل سفارش: {money(m.max_toman)}
                                </small>
                              )}
                            </span>
                          </label>
                        ))}
                      <p className="buy-note">
                        هر واریز ابتدا کیف پول را شارژ می‌کند. تیم پس از بررسی
                        واریز، اعتبار و رزرو، سفارش را جداگانه تأیید می‌کند.
                      </p>
                      <button
                        className="buy-primary"
                        disabled={
                          action.busy ||
                          (!method && preview.quote.funding_toman !== "0")
                        }
                        onClick={() =>
                          void action.run(async (key) => {
                            setQuote(
                              await call<Quote>(
                                "/storefront/checkout/quotes",
                                token,
                                input,
                                undefined,
                                key,
                              ).catch(shippingFailure),
                            );
                            setConsent(false);
                          }, JSON.stringify(input))
                        }
                      >
                        بررسی نهایی سفارش
                      </button>
                    </section>
                    {quote && (
                      <section className="buy-panel">
                        <h2>بازبینی و تأیید مبلغ</h2>
                        <OrderSummary quote={quote} />
                        <label className="buy-choice">
                          <ToggleSwitch
                            checked={consent}
                            onChange={setConsent}
                          />
                          <span>
                            مبلغ {money(quote.total_toman)}، آدرس و شرایط ارسال
                            را تأیید می‌کنم. رزرو ۳۰ دقیقه اعتبار دارد.
                          </span>
                        </label>
                        <button
                          className="buy-primary"
                          disabled={action.busy || !consent}
                          onClick={() =>
                            void action.run(async (key) => {
                              const result = await call<Purchase>(
                                "/storefront/checkout/purchases",
                                token,
                                {
                                  quote_id: quote.id,
                                  accepted_total_toman: quote.total_toman,
                                  terms_digest: quote.terms_digest,
                                },
                                undefined,
                                key,
                              ).catch((error: unknown) => {
                                setConsent(false);
                                if (
                                  error instanceof CommerceError &&
                                  [400, 409].includes(error.status)
                                )
                                  setQuote(null);
                                shippingFailure(error);
                                throw error;
                              });
                              await refreshCart();
                              navigate("/account/orders/" + result.number);
                            }, "accept:" + quote.id)
                          }
                        >
                          ثبت سفارش و ادامه پرداخت
                        </button>
                      </section>
                    )}
                  </>
                )
              )}
            </section>
            <aside className="buy-panel buy-summary">
              <h2>خلاصه خرید</h2>
              {preview ? (
                <OrderSummary quote={quote ?? preview.quote} />
              ) : (
                <>
                  <OrderItems lines={prepare.cart.items} />
                  <p>{prepare.cart.item_count} کالا</p>
                  <strong>{money(prepare.cart.subtotal_toman)}</strong>
                  <p>
                    هزینه ارسال پس از انتخاب آدرس و روش ارسال محاسبه می‌شود.
                  </p>
                </>
              )}
              <a href="/cart">بازگشت به سبد</a>
            </aside>
          </div>
        )}
      </>
    </BusyRegion>
  );
}
export function OrderSummary({ quote, showTotals = true }: { quote: Quote; showTotals?:boolean }) {
  return (
    <div className="buy-order-summary">
      <OrderItems lines={quote.lines} />
      {quote.shipments.map((s) => (
        <div className="buy-shipment" key={s.store_id}>
          <b>
            {s.store_name} · {s.method_name}
          </b>
          <p>{shippingCharge(s.collection_mode,s.fee_toman)}</p>
          <p>{deliveryPromise(s.delivery,s.delivery_from,s.delivery_to)}</p>
          <small>{dispatchTime(s.lead_days,s.lead_day_basis)}</small>
        </div>
      ))}
      {showTotals && <><div className="buy-total">
        <span>جمع کالاها</span>
        <b>{money(quote.cash_subtotal_toman ?? quote.subtotal_toman)}</b>
      </div>
      {Number(quote.markup_toman ?? "0") > 0 && (
        <p className="buy-row">
          <span>مارکاپ {quote.sales_method?.name}</span>
          <b>{money(quote.markup_toman!)}</b>
        </p>
      )}
      {Number(quote.direct_discount_toman ?? "0") > 0 && (
        <div className="buy-total buy-direct-saving">
          <span title="این تخفیف قبلاً در قیمت کالاها منظور شده است">
            سود شما از خرید
          </span>
          <b>{money(quote.direct_discount_toman!)}</b>
        </div>
      )}
      <div className="buy-total">
        <span>جمع ارسال</span>
        <b>{money(quote.shipping_toman)}</b>
      </div>
      {quote.contract_version === 2 &&
        Number(quote.discount_toman ?? "0") > 0 && (
          <>
            <div className="buy-total">
              <span>
                کد تخفیف {quote.coupon_code && <bdi>{quote.coupon_code}</bdi>}
              </span>
              <b>− {money(quote.discount_toman ?? "0")}</b>
            </div>
          </>
        )}
      <div className="buy-total buy-grand">
        <span>مبلغ نهایی</span>
        <b>{money(quote.total_toman)}</b>
      </div>
      {quote.contract_version === 2 &&
        Number(quote.wallet_selected_toman ?? "0") > 0 && (
          <>
            <div className="buy-total">
              <span>اعتبار انتخابی کیف پول</span>
              <b>{money(quote.wallet_selected_toman ?? "0")}</b>
            </div>
            <div className="buy-total">
              <span>نیاز به تأمین وجه</span>
              <b>{money(quote.funding_toman ?? "0")}</b>
            </div>
          </>
        )}
      </>}
    </div>
  );
}
export function PurchaseStatus({ purchase: p, showCreditSources = true }: { purchase: Purchase; showCreditSources?:boolean }) {
  return (
    <>
      {showCreditSources && !!p.credit_sources?.length && (
        <section
          className="buy-panel buy-credit-sources"
          aria-label="منابع اعتبار این سفارش"
        >
          <h2>منابع اعتبار این سفارش</h2>
          {p.credit_sources.map((c, i) => (
            <article key={c.lot_id + ":" + i}>
              <b>
                {c.title ||
                  (c.kind === "selected-balance"
                    ? "اعتبار قبلی انتخابی"
                    : "وجه جدید این خرید")}
              </b>
              <p>
                {money(c.amount_toman)} ·{" "}
                {c.withdrawable ? "قابل برداشت" : "فقط قابل خرید"} ·{" "}
                {c.status === "posted"
                  ? "مصرف‌شده"
                  : c.status === "held"
                    ? "رزرو"
                    : "آزادشده"}
              </p>
            </article>
          ))}
        </section>
      )}

      <div className="buy-statuses">
        <div>
          <small>وضعیت سفارش</small>
          <strong>{label(p.order_status)}</strong>
        </div>
        <div>
          <small>وضعیت پرداخت</small>
          <strong>{label(p.payment_status)}</strong>
        </div>
      </div>
      <p className="buy-note">پیگیری خرید: {label(p.status)}</p>
      {p.fulfillment_status&&<p className="buy-note">عملیات مرسوله‌ها: {label(p.fulfillment_status)}</p>}
      {p.reason && <p className="buy-alert">{p.reason}</p>}
    </>
  );
}
export function OrderPage({ id }: { id?: string }) {
  const { token } = useCommerce();
  return (
    <CommerceShell title={id ? "جزئیات سفارش" : "سفارش‌های من"}>
      <RequireLogin>{token && <OrderContent id={id} />}</RequireLogin>
    </CommerceShell>
  );
}
function OrderContent({ id }: { id?: string }) {
  const { token, config } = useCommerce(),
    action = useAction(),
    [rows, setRows] = useState<Purchase[] | null>(null),
    [deposit, setDeposit] = useState(false),
    [reason, setReason] = useState(""),
    [clock, setClock] = useState(Date.now());
  const load = useCallback(async () => {
    const v = await call<Purchase[] | Purchase>(
      "/storefront/account/orders" + (id ? "/" + encodeURIComponent(id) : ""),
      token,
    );
    setRows(Array.isArray(v) ? v : [v]);
  }, [id, token]);
  useEffect(() => {
    void action.run(async () => {
      await load();
    });
    const timer = setInterval(() => {
      setClock(Date.now());
      void load().catch(() => {});
    }, 15000);
    return () => clearInterval(timer);
  }, [load]);
  return (
    <BusyRegion busy={action.busy}>
      <>
        <Feedback action={action} />
        <button
          className="buy-secondary"
          disabled={action.busy}
          onClick={() =>
            void action.run(async () => {
              await load();
            })
          }
        >
          تازه‌سازی وضعیت
        </button>
        {rows === null ? (
          <LoadingPlaceholder label="دریافت سفارش" />
        ) : !rows.length ? (
          <div className="buy-panel">هنوز سفارشی ثبت نشده است.</div>
        ) : (
          rows
            .slice()
            .reverse()
            .map((p) => (
              <article className="buy-panel" key={p.id}>
                <div className="buy-order-list-heading"><h2>{id?"سفارش شما":"سفارش "+p.number.replace(/\d/g,d=>"۰۱۲۳۴۵۶۷۸۹"[Number(d)])}</h2>
                {!id&&<a className="buy-primary order-view-link" href={"/account/orders/"+p.number}>مشاهده جزئیات سفارش <span aria-hidden="true">←</span></a>}</div>
                <p className="buy-order-number">
                  شماره سفارش: <b>{p.number.replace(/\d/g,d=>"۰۱۲۳۴۵۶۷۸۹"[Number(d)])}</b>
                </p>
                <PurchaseStatus purchase={p} showCreditSources={false} />
                {!id ? (
                  <strong>{money(p.quote.total_toman)}</strong>
                ) : (
                  <div className="buy-layout">
                    <section>
                      <CustomerPurchaseProgress purchase={p}/>
                      <CustomerOrderFinance purchase={p}/>
                      <CustomerShipments id={p.id} orders={p.order_ids} quote={p.quote} version={p.version} onChanged={load}/>
                      <CustomerOrderHistory id={p.id} version={p.version}/>
                      <h3>گیرنده و آدرس</h3>
                      <p>
                        {p.quote.address.recipient} · {p.quote.address.phone}
                      </p>
                      <p>
                        {p.quote.address.province}، {p.quote.address.city}،{" "}
                        {p.quote.address.full_address || p.quote.address.street}{" "}
                        · پلاک {p.quote.address.house_number || "—"} ·{" "}
                        {p.quote.address.postal_code}
                      </p>
                      <h3>اعلام واریزها</h3>
                      <Deposits
                        rows={p.deposits}
                        token={token}
                        onRefresh={load}
                      />
                    </section>
                    <aside>
                      <h3>
                        {p.quote.payment_method.name ||
                          "بدون نیاز به تأمین وجه"}
                      </h3>
                      {p.status !== "confirmed" && p.status !== "cancelled" && (
                        <>
                          <p>
                            موجودی قابل استفاده:{" "}
                            <b>{money(p.wallet_available_toman)}</b>
                          </p>
                          <p>
                            کسری اعتبار: <b>{money(p.shortfall_toman)}</b>
                          </p>
                          {p.status === "requires-reconfirmation" ||
                          Date.parse(p.quote.reserved_until) <= clock ? (
                            <>
                              <p className="buy-alert">
                                رزرو پایان یافته یا شرایط نیاز به بررسی دارد؛
                                واریز دیرهنگام فقط کیف پول را شارژ می‌کند.
                              </p>
                              <a
                                className="buy-primary"
                                href={"/checkout?purchase=" + p.id}
                              >
                                بررسی قیمت و تأیید دوباره
                              </a>
                            </>
                          ) : (
                            <p className="buy-notice">
                              رزرو تا{" "}
                              {new Date(
                                p.quote.reserved_until,
                              ).toLocaleTimeString("fa-IR")}{" "}
                              · حدود{" "}
                              {Math.max(
                                0,
                                Math.ceil(
                                  (Date.parse(p.quote.reserved_until) - clock) /
                                    60000,
                                ),
                              )}{" "}
                              دقیقه
                            </p>
                          )}
                          {p.quote.payment_method.id &&
                          p.quote.payment_method.kind !== "wallet" &&
                          (p.quote.contract_version !== 2 ||
                            p.shortfall_toman !== "0") ? (
                            <button
                              className="buy-primary"
                              onClick={() => setDeposit(true)}
                            >
                              اعلام پرداخت / واریز
                            </button>
                          ) : (
                            <a className="buy-secondary" href="/wallet">
                              شارژ مستقل کیف پول
                            </a>
                          )}
                          <p className="buy-note">
                            پس از تأمین اعتبار، منتظر تأیید مستقل تیم بمانید.
                            ثبت رسید یا داشتن اعتبار، پرداخت سفارش محسوب
                            نمی‌شود.
                          </p>
                        </>
                      )}
                      {p.confirmed_at && <CustomerRequests id={p.id} token={token} onChanged={load}/>}
                      {p.status !== "cancelled" && p.status !== "confirmed" && (
                        <form
                          onSubmit={(e) => {
                            e.preventDefault();
                            void action.run(
                              async (key) => {
                                await call(
                                  "/storefront/account/orders/" +
                                    p.id +
                                    "/cancel",
                                  token,
                                  { reason },
                                  undefined,
                                  key,
                                );
                                await load();
                              },
                              "cancel:" + p.id + reason,
                            );
                          }}
                        >
                          <label className="buy-field">
                            دلیل لغو
                            <input
                              value={reason}
                              onChange={(e) => setReason(e.target.value)}
                              required
                              maxLength={1000}
                            />
                          </label>
                          <button
                            className="buy-secondary"
                            disabled={action.busy || p.cancellation_requested}
                          >
                            {p.cancellation_requested
                              ? "درخواست لغو در حال بررسی"
                              : p.status === "confirmed"
                                ? "درخواست لغو سفارش"
                                : "لغو سفارش"}
                          </button>
                        </form>
                      )}
                    </aside>
                  </div>
                )}
                {id && deposit && config && (
                  <Modal title="اعلام پرداخت" onClose={() => setDeposit(false)}>
                    <DepositForm
                      methods={[p.quote.payment_method]}
                      destination={
                        p.quote.payment_method.kind === "card-transfer"
                          ? p.quote.destination
                          : config.destination
                      }
                      purchaseID={p.id}
                      defaultAmount={
                        p.shortfall_toman === "0"
                          ? p.quote.total_toman
                          : p.shortfall_toman
                      }
                      onDone={async () => {
                        setDeposit(false);
                        await load();
                      }}
                    />
                  </Modal>
                )}
              </article>
            ))
        )}
      </>
    </BusyRegion>
  );
}
export function Deposits({
  rows,
  token,
  onRefresh,
}: {
  rows: Deposit[];
  token: string;
  onRefresh: () => Promise<void>;
}) {
  const action = useAction();
  return (
    <BusyRegion busy={action.busy}>
      <>
        <Feedback action={action} />
        {!rows.length ? (
          <p className="buy-note">اعلام واریزی ثبت نشده است.</p>
        ) : (
          rows
            .slice()
            .reverse()
            .map((d) => (
              <div className="buy-deposit" key={d.id}>
                <b>
                  {d.method_name} · {money(d.amount_toman)}
                </b>
                <p>
                  {label(d.status)} · {persianDigits(persianCapacityDate(d.date))} · مرجع {d.reference}
                </p>
                {d.reason && <p>{d.reason}</p>}
                {d.status === "pending" && (
                  <label className="buy-field">
                    {d.receipt_id
                      ? "جایگزینی تصویر رسید (اختیاری)"
                      : "تصویر رسید (اختیاری)"}
                    <input
                      type="file"
                      accept="image/png,image/jpeg,image/webp"
                      disabled={action.busy}
                      onChange={(e) => {
                        const file = e.target.files?.[0];
                        if (!file) return;
                        void action.run(
                          async (key) => {
                            if (file.size > 2 * 1024 * 1024)
                              throw Error("حجم رسید حداکثر ۲ مگابایت است.");
                            const f = new FormData();
                            f.set("file", file);
                            f.set("version", String(d.version));
                            await call(
                              "/storefront/wallet/deposits/" +
                                d.id +
                                "/receipt",
                              token,
                              f,
                              undefined,
                              key,
                            );
                            await onRefresh();
                            action.setNotice(
                              "تصویر رسید ثبت شد؛ واریز همچنان در انتظار بررسی است.",
                            );
                          },
                          "receipt:" + d.id + file.name + file.size,
                        );
                      }}
                    />
                  </label>
                )}
                {d.receipt_id && (
                  <PrivateReceipt id={d.receipt_id} token={token} />
                )}
              </div>
            ))
        )}
      </>
    </BusyRegion>
  );
}
export function PrivateReceipt({
  id,
  token,
  admin = false,
}: {
  id: string;
  token: string;
  admin?: boolean;
}) {
  const [url, setURL] = useState(""),
    [error, setError] = useState("");
  useEffect(() => {
    let active = true,
      object = "";
    fetch(
      apiBase +
        (admin
          ? "/backoffice/commerce/receipts/"
          : "/storefront/wallet/receipts/") +
        id,
      { headers: { Authorization: `Bearer ${token}` }, cache: "no-store" },
    )
      .then(async (r) => {
        if (!r.ok) throw Error("دریافت رسید مجاز نیست یا انجام نشد.");
        const blob = await r.blob();
        if (active) {
          object = URL.createObjectURL(blob);
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
  }, [id, token, admin]);
  return error ? (
    <p role="alert">{error}</p>
  ) : url ? (
    <a href={url} target="_blank" rel="noreferrer">
      <img className="buy-receipt" src={url} alt="تصویر خصوصی رسید" />
    </a>
  ) : (
    <LoadingPlaceholder label="دریافت رسید" />
  );
}
export function DepositForm({
  methods,
  destination,
  purchaseID,
  defaultAmount = "",
  onDone,
}: {
  methods: Method[];
  destination: {
    id: string;
    card_number: string;
    holder: string;
    bank: string;
  };
  purchaseID?: string;
  defaultAmount?: string;
  onDone: () => Promise<void>;
}) {
  const { token, config } = useCommerce(),
    action = useAction(),
    [method, setMethod] = useState(methods[0]?.id ?? ""),
    [depositDate,setDepositDate]=useState(tehranCapacityDate);
  const chosen = methods.find((m) => m.id === method);
  return (
    <BusyRegion busy={action.busy}>
      <form
        onSubmit={(e) => {
          const v = values(e),
            input = {
              method_id: method,
              purchase_id: purchaseID,
              destination_id:
                chosen?.kind === "card-transfer" ? destination.id : "",
              amount_toman: digits(v.amount),
              date: v.date,
              reference: digits(v.reference),
            };
          void action.run(async (key) => {
            await call(
              "/storefront/wallet/deposits",
              token,
              input,
              undefined,
              key,
            );
            await onDone();
          }, JSON.stringify(input));
        }}
      >
        <fieldset disabled={action.busy}>
          <ChoiceGroup
            label="روش تأمین اعتبار"
            value={method}
            onChange={setMethod}
            options={methods
              .filter((m) => m.kind !== "wallet")
              .map((m) => ({ value: m.id, label: m.name }))}
          />
          {chosen?.kind === "card-transfer" && (
            <div className="buy-destination">
              <p>شماره کارت مقصد</p>
              <b dir="ltr">{destination.card_number}</b>
              <p>
                {destination.holder} · {destination.bank}
              </p>
            </div>
          )}
          <p>{chosen?.instructions}</p>
          {chosen?.max_toman && (
            <p className="buy-note">سقف این واریز: {money(chosen.max_toman)}</p>
          )}
          <Input
            name="amount"
            label="مبلغ واریز (تومان)"
            inputMode="decimal"
            defaultValue={defaultAmount}
            required
          />
          <PersianDateField
            name="date"
            label="تاریخ واریز"
            value={depositDate}
            onChange={setDepositDate}
            required
          />
          <Input
            name="reference"
            label="شماره پیگیری / مرجع"
            required
            minLength={3}
            maxLength={80}
          />
          <p className="buy-note">
            فقط پرداخت انجام‌شده را اعلام کنید. تصویر رسید را پس از ثبت، در همین
            صفحه می‌توانید اضافه کنید. شارژ پس از بررسی دستی انجام می‌شود.
          </p>
          <Feedback action={action} />
          <button
            className="buy-primary"
            disabled={!method || chosen?.kind === "wallet"}
          >
            ثبت اعلام پرداخت
          </button>
        </fieldset>
      </form>
    </BusyRegion>
  );
}
export function WalletPage() {
  const { token } = useCommerce();
  const transactions=new URLSearchParams(location.search).get("view")==="transactions";
  return (
    <CommerceShell title={transactions?"تراکنش‌های کیف پول":"کیف پول"}>
      <RequireLogin>{token && (transactions?<WalletTransactions/>:<WalletContent />)}</RequireLogin>
    </CommerceShell>
  );
}
function WalletContent() {
  const { token } = useCommerce(),
    action = useAction(),
    [wallet, setWallet] = useState<Wallet | null>(null),
    [topup, setTopup] = useState(false),
    [refund, setRefund] = useState(false);
  const load = useCallback(
    async () => setWallet(await call<Wallet>("/storefront/wallet", token)),
    [token],
  );
  useEffect(() => {
    void action.run(async () => {
      await load();
    });
  }, [load]);
  return (
    <BusyRegion busy={action.busy}>
      <>
        <Feedback action={action} />
        <button
          className="buy-secondary"
          disabled={action.busy}
          onClick={() =>
            void action.run(async () => {
              await load();
            })
          }
        >
          تازه‌سازی کیف پول
        </button>
        {!wallet ? (
          <LoadingPlaceholder label="دریافت کیف پول" />
        ) : (
          <>
            <div className="buy-wallet-card">
              <span>اعتبار قابل استفاده</span>
              <strong>{money(wallet.available_toman)}</strong>
              <p>
                قابل برداشت:{" "}
                <b>
                  {money(wallet.withdrawable_toman ?? wallet.available_toman)}
                </b>{" "}
                · فقط‌خرید: <b>{money(wallet.purchase_only_toman ?? "0")}</b>
              </p>
              <small>
                کل موجودی: {money(wallet.balance_toman)} · در انتظار بازپرداخت:{" "}
                {money(wallet.held_toman)}
              </small>
              <div className="buy-row">
                <a className="buy-secondary" href="/wallet?view=transactions">مشاهده تراکنش‌ها</a>
                <button
                  onClick={() => setTopup(true)}
                  disabled={!wallet.funding_methods.length}
                >
                  شارژ کیف پول
                </button>
                <button onClick={() => setRefund(true)}>
                  درخواست بازپرداخت
                </button>
              </div>
            </div>
            <section className="buy-panel">
              <h2>گردش کیف پول</h2>
              {!!wallet.credits?.length && (
                <section
                  className="buy-credit-lots"
                  aria-label="منابع اعتبار کیف پول"
                >
                  <h2>اعتبارهای کیف پول</h2>
                  {[...wallet.credits].reverse().map((l) => (
                    <article key={l.id}>
                      <div>
                        <b>{l.title}</b>
                        <span className="buy-credit-kind">
                          {l.withdrawable ? "قابل برداشت" : "فقط قابل خرید"}
                        </span>
                      </div>
                      <p>{l.description}</p>
                      <p>
                        مانده آزاد: <b>{money(l.free_toman)}</b> · رزروشده:{" "}
                        {money(l.held_toman)}
                      </p>
                      <small>
                        {new Date(l.at).toLocaleDateString("fa-IR")}
                      </small>
                    </article>
                  ))}
                </section>
              )}
              {wallet.entries.length ? (
                wallet.entries
                  .slice()
                  .reverse()
                  .map((e) => (
                    <div key={e.id} className="buy-total">
                      <span>
                        {label(e.kind)}
                        <small>{new Date(e.at).toLocaleString("fa-IR")}</small>
                      </span>
                      <b>{money(e.amount_toman)}</b>
                    </div>
                  ))
              ) : (
                <p>تراکنشی ثبت نشده است.</p>
              )}
            </section>
            <section className="buy-panel">
              <h2>اعلام واریزها</h2>
              <Deposits rows={wallet.deposits} token={token} onRefresh={load} />
            </section>
            <section className="buy-panel">
              <h2>درخواست‌های بازپرداخت</h2>
              {wallet.refunds.length ? (
                wallet.refunds.map((r) => (
                  <div key={r.id}>
                    <b>
                      {money(r.amount_toman)} · {label(r.status)}
                    </b>
                    <p>{r.instructions}</p>
                    <p>
                      {r.reason} {r.reference && " · مرجع " + r.reference}
                    </p>
                  </div>
                ))
              ) : (
                <p>درخواستی ثبت نشده است.</p>
              )}
            </section>
            {topup && (
              <Modal title="شارژ کیف پول" onClose={() => setTopup(false)}>
                <DepositForm
                  methods={wallet.funding_methods}
                  destination={wallet.destination}
                  onDone={async () => {
                    setTopup(false);
                    await load();
                  }}
                />
              </Modal>
            )}
            {refund && (
              <Modal
                busy={action.busy}
                title="درخواست بازپرداخت دستی"
                onClose={() => setRefund(false)}
              >
                <form
                  onSubmit={(e) => {
                    const v = values(e),
                      input = {
                        amount_toman: digits(v.amount),
                        instructions: v.instructions,
                      };
                    void action.run(async (key) => {
                      await call(
                        "/storefront/wallet/refunds",
                        token,
                        input,
                        undefined,
                        key,
                      );
                      setRefund(false);
                      await load();
                    }, JSON.stringify(input));
                  }}
                >
                  <Input
                    name="amount"
                    label="مبلغ بازپرداخت (تومان)"
                    inputMode="decimal"
                    required
                  />
                  <label className="buy-field">
                    اطلاعات مقصد و توضیح درخواست
                    <textarea name="instructions" maxLength={2000} required />
                  </label>
                  <p>
                    مبلغ تا رسیدگی تیم از اعتبار قابل استفاده کنار گذاشته
                    می‌شود. پرداخت بانکی به‌صورت دستی انجام خواهد شد.
                  </p>
                  <Feedback action={action} />
                  <button className="buy-primary" disabled={action.busy}>
                    ثبت درخواست رسیدگی
                  </button>
                </form>
              </Modal>
            )}
          </>
        )}
      </>
    </BusyRegion>
  );
}
export function AccountPage({ addresses = false }: { addresses?: boolean }) {
  const { token, profile, logout, refreshProfile } = useCommerce(),
    action = useAction();
  return (
    <CommerceShell title={addresses ? "آدرس‌های من" : "حساب کاربری"}>
      <BusyRegion busy={action.busy}>
        <RequireLogin>
          {addresses ? (
            <AddressBook />
          ) : (
            <>
              <nav className="buy-account-nav">
                <a href="/account/wishlist">علاقه‌مندی‌ها</a>
                <a href="/account/recently-viewed">بازدیدهای اخیر</a>
                <a href="/account/orders">سفارش‌ها</a>
                <a href="/account/addresses">آدرس‌ها</a>
                <a href="/wallet">کیف پول</a>
                <a href="/wallet?view=transactions">تراکنش‌های کیف پول</a>
                <button
                  onClick={() => {
                    void call("/storefront/auth/logout", token, {}).finally(
                      logout,
                    );
                  }}
                >
                  خروج
                </button>
              </nav>
              <section className="buy-panel buy-auth">
                <h2>اطلاعات حساب</h2>
                <p>
                  ایمیل اختیاری است و برای ورود با رمز استفاده می‌شود؛ ورود با
                  OTP از طریق موبایل است.
                </p>
                <form
                  key={profile?.name + ":" + profile?.email}
                  onSubmit={(e) => {
                    const v = values(e);
                    void action.run(async () => {
                      await call(
                        "/storefront/account/profile",
                        token,
                        {
                          name: v.name,
                          email: v.email,
                          password: v.password,
                          ...(v.national_id
                            ? { national_id: v.national_id }
                            : {}),
                        },
                        "PATCH",
                      );
                      await refreshProfile();
                      action.setNotice(
                        "اطلاعات ثبت شد؛ پس از تغییر رمز ممکن است لازم باشد دوباره وارد شوید.",
                      );
                    });
                  }}
                >
                  <Input
                    name="name"
                    label="نام"
                    defaultValue={profile?.name}
                    required
                  />
                  <Input
                    name="national_id"
                    label="کد ملی (برای تحویل سفارش به خودتان الزامی)"
                    defaultValue={profile?.national_id ?? ""}
                    maxLength={10}
                    inputMode="numeric"
                    dir="ltr"
                  />
                  <Input
                    name="email"
                    label="ایمیل (اختیاری)"
                    type="email"
                    defaultValue={profile?.email ?? ""}
                  />
                  <Input
                    name="password"
                    label="ثبت یا تغییر رمز (اختیاری)"
                    type="password"
                    autoComplete="new-password"
                    minLength={10}
                  />
                  <Feedback action={action} />
                  <button className="buy-primary" disabled={action.busy}>
                    ذخیره پروفایل
                  </button>
                </form>
              </section>
            </>
          )}
        </RequireLogin>
      </BusyRegion>
    </CommerceShell>
  );
}
export function BusinessPage({ slug }: { slug: string }) {
 const {data:business,loading,error,refresh}=usePublicBusiness();
 const entry=businessPages.find(([key])=>key===slug);
 const text=entry?business?.[entry[1]]:'';
 return <CommerceShell title={entry?.[2]??'صفحه پیدا نشد'}><section className="buy-panel buy-prose business-public-page">
 {loading&&!business?<p role="status">در حال دریافت اطلاعات صفحه…</p>:error?<div role="alert"><p>{error}</p><button onClick={refresh}>تلاش مجدد</button></div>:text?<div className="business-page-body">{text}</div>:<><p>این صفحه در دسترس نیست.</p><a href="/">بازگشت به صفحه اصلی</a></>}
 </section></CommerceShell>
}
