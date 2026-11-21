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
import { apiBase, call, operationKey, useCommerce } from "./commerce";
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
  type Wallet,
  type Method,
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
      <SiteHeader categories={[]} />
      <main className="lh-wrap buy-page">
        <nav className="buy-breadcrumb" aria-label="مسیر صفحه">
          <a href="/">خانه</a>
          <span> / </span>
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
  return (
    <CommerceShell title="ورود به چریوا">
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
          <Recovery identifier={identifier} onDone={() => setRecover(false)} />
        ) : !email && mode === "otp" ? (
          <>
            {request ? (
              <form
                onSubmit={(e) => {
                  const v = values(e);
                  void action.run(
                    async () => {
                      const r = await call<{ token: string }>(
                        "/storefront/auth/verify-otp",
                        "",
                        {
                          request_id: request.request_id,
                          code: digits(v.code),
                        },
                      );
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
                  const r = await call<{ token: string }>(
                    "/storefront/auth/password-login",
                    "",
                    { phone: digits(identifier), password: v.password },
                  );
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
            <p role="status">در حال دریافت سبد…</p>
          ) : cart.items.length ? (
            cart.items.map((item) => (
              <article className="buy-panel buy-cart-row" key={item.offer_id}>
                <a
                  href={
                    "/product/" +
                    item.product_id +
                    "?variant=" +
                    item.variant_id +
                    "&offer=" +
                    item.offer_id
                  }
                >
                  <img src={apiBase + item.image_url} alt={item.name} />
                </a>
                <div>
                  <a href={"/product/" + item.product_id}>
                    <h2>{item.name}</h2>
                  </a>
                  <p>{item.variant_label}</p>
                  <p>
                    فروشنده: {item.store_name} · تأمین: {item.lead_days} روز
                    کاری
                  </p>
                  <strong>{money(item.unit_price_toman)}</strong>
                  {item.price_changed && (
                    <p className="buy-alert">
                      قیمت از {money(item.previous_price_toman)} به‌روز شده؛
                      مبلغ نهایی را در پرداخت تأیید کنید.
                    </p>
                  )}
                  {!item.available && (
                    <p className="buy-alert">
                      این تعداد یا عرضه فعلاً قابل خرید نیست؛ تعداد یا فروشنده
                      را اصلاح کنید.
                    </p>
                  )}
                  <div className="buy-row">
                    <button
                      aria-label={"کاهش تعداد " + item.name}
                      disabled={action.busy || item.quantity <= 1}
                      onClick={() =>
                        void action.run(async () => {
                          await mutate({
                            action: "set",
                            offer_id: item.offer_id,
                            quantity: item.quantity - 1,
                            version: cart.version,
                          });
                        })
                      }
                    >
                      −
                    </button>
                    <b>{item.quantity}</b>
                    <button
                      aria-label={"افزایش تعداد " + item.name}
                      disabled={action.busy || item.quantity >= 99}
                      onClick={() =>
                        void action.run(async () => {
                          await mutate({
                            action: "set",
                            offer_id: item.offer_id,
                            quantity: item.quantity + 1,
                            version: cart.version,
                          });
                        })
                      }
                    >
                      +
                    </button>
                    <button
                      className="buy-link"
                      disabled={action.busy}
                      onClick={() =>
                        void action.run(async () => {
                          await mutate({
                            action: "remove",
                            offer_id: item.offer_id,
                            quantity: 0,
                            version: cart.version,
                          });
                        })
                      }
                    >
                      حذف
                    </button>
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
                <strong>{money(item.total_toman)}</strong>
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
        <Modal title="انتخاب تنوع و فروشنده" onClose={() => setEdit(null)}>
          <label className="buy-field">
            تنوع
            <select
              value={variant}
              onChange={(e) =>
                void action.run(async () => {
                  await loadOffers(
                    cart.items.find((i) => i.offer_id === edit)!.product_id,
                    e.target.value,
                  );
                })
              }
            >
              {variants.map((v) => (
                <option key={v.id} value={v.id}>
                  {v.values.map((x) => x.value).join(" / ") || "پیش‌فرض"}
                </option>
              ))}
            </select>
          </label>
          <label className="buy-field">
            فروشنده
            <select
              value={newOffer}
              onChange={(e) => setNewOffer(e.target.value)}
            >
              <option value="">انتخاب کنید</option>
              {offers.map((o) => (
                <option key={o.id} value={o.id} disabled={!o.available}>
                  {o.store_name} · {money(String(o.price / 10))}
                  {!o.available ? " · ناموجود" : ""}
                </option>
              ))}
            </select>
          </label>
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
};
export function AddressBook({
  onSelect,
  selected,
}: {
  onSelect?: (id: string) => void;
  selected?: string;
}) {
  const { token } = useCommerce(),
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
  return (
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
        <p>در حال دریافت آدرس‌ها…</p>
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
              <label>
                <input
                  type="radio"
                  name="address"
                  value={a.id}
                  checked={selected === a.id}
                  onChange={() => onSelect(a.id)}
                />{" "}
                انتخاب این آدرس
              </label>
            )}
            <h3>
              {a.recipient} · {a.city}
            </h3>
            <p>
              {a.province}، {a.street}
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
          title={draft.id ? "ویرایش آدرس" : "آدرس جدید"}
          onClose={() => {
            if (!action.busy) setDraft(null);
          }}
        >
          <form
            onSubmit={(e) => {
              const fields = values(e),
                input = {
                  id: draft.id,
                  version: draft.version,
                  recipient: fields.recipient,
                  phone: fields.phone,
                  province: fields.province,
                  city: fields.city,
                  street: fields.street,
                  postal_code: fields.postal_code,
                };
              void action.run(async (key) => {
                const a = await call<Address>(
                  "/storefront/account/addresses",
                  token,
                  input,
                  undefined,
                  key,
                );
                await load();
                onSelect?.(a.id);
                setDraft(null);
              }, JSON.stringify(input));
            }}
          >
            <fieldset disabled={action.busy} className="buy-fields">
              {[
                ["recipient", "نام گیرنده"],
                ["phone", "موبایل گیرنده"],
                ["province", "استان"],
                ["city", "شهر"],
                ["street", "نشانی کامل"],
                ["postal_code", "کد پستی"],
              ].map(([name, caption]) => (
                <Input
                  key={name}
                  name={name}
                  label={caption}
                  defaultValue={String(draft[name as keyof Address])}
                  required
                  maxLength={name === "street" ? 1000 : 100}
                />
              ))}
              <Feedback action={action} />
              <button className="buy-primary">ذخیره آدرس</button>
            </fieldset>
          </form>
        </Modal>
      )}
    </section>
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
  const { token, refreshCart } = useCommerce(),
    action = useAction();
  const [prepare, setPrepare] = useState<Prepare | null>(null),
    [address, setAddress] = useState(""),
    [shipping, setShipping] = useState<Record<string, string>>({}),
    [preview, setPreview] = useState<Preview | null>(null),
    [method, setMethod] = useState(""),
    [quote, setQuote] = useState<Quote | null>(null),
    [consent, setConsent] = useState(false);
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
        setAddress(old.quote.address.id);
      }
    });
  }, [token, purchaseID]);
  const change = () => {
    setPreview(null);
    setQuote(null);
    setConsent(false);
  };
  const input: QuoteInput = {
    cart_version: prepare?.cart.version ?? 0,
    address_id: address,
    shipping,
    payment_method_id: method,
    ...(purchaseID ? { purchase_id: purchaseID } : {}),
  };
  const stores = oldPurchase
    ? oldPurchase.quote.shipments.map((s) => ({
        id: s.store_id,
        name: s.store_name,
      }))
    : Array.from(
        new Map(
          prepare?.cart.items.map((i) => [
            i.store_id,
            { id: i.store_id, name: i.store_name },
          ]),
        ).values(),
      );
  return (
    <>
      <ol className="buy-steps">
        <li className="active">۱ آدرس</li>
        <li className={address ? "active" : ""}>۲ ارسال</li>
        <li className={preview ? "active" : ""}>۳ پرداخت</li>
        <li className={quote ? "active" : ""}>۴ ثبت سفارش</li>
      </ol>
      <Feedback action={action} />
      {!prepare ? (
        <p role="status">در حال دریافت شرایط خرید…</p>
      ) : (
        <div className="buy-layout">
          <section>
            <AddressBook
              selected={address}
              onSelect={(id) => {
                setAddress(id);
                change();
              }}
            />
            <section className="buy-panel">
              <h2>روش ارسال برای هر فروشنده</h2>
              {stores.map((st) => (
                <label key={st.id} className="buy-field">
                  {st.name}
                  <select
                    value={shipping[st.id] ?? ""}
                    onChange={(e) => {
                      setShipping({ ...shipping, [st.id]: e.target.value });
                      change();
                    }}
                  >
                    <option value="">انتخاب روش ارسال</option>
                    {prepare.shipping
                      .filter(
                        (s) =>
                          !s.store_ids?.length || s.store_ids.includes(st.id),
                      )
                      .map((s) => (
                        <option key={s.id} value={s.id}>
                          {s.name} · {money(s.fee_toman)} · {s.min_days} تا{" "}
                          {s.max_days} روز کاری پس از تأمین
                        </option>
                      ))}
                  </select>
                </label>
              ))}
              {!prepare.shipping.length && (
                <p className="buy-alert">
                  هنوز روش ارسال فعالی تنظیم نشده است.
                </p>
              )}
              <button
                className="buy-secondary"
                disabled={
                  action.busy ||
                  !address ||
                  !stores.length ||
                  stores.some((s) => !shipping[s.id])
                }
                onClick={() =>
                  void action.run(async () => {
                    const result = await call<Preview>(
                      "/storefront/checkout/preview",
                      token,
                      { ...input, payment_method_id: "" },
                    );
                    setPreview(result);
                    setMethod("");
                    setQuote(null);
                  })
                }
              >
                محاسبه مبلغ و روش‌های پرداخت
              </button>
            </section>
            {preview && (
              <section className="buy-panel">
                <h2>روش پرداخت</h2>
                <p>
                  موجودی قابل استفاده کیف پول:{" "}
                  <b>{money(prepare.wallet_available_toman)}</b>
                </p>
                {preview.methods.map((m) => (
                  <label
                    key={m.id}
                    className={"buy-choice " + (!m.eligible ? "disabled" : "")}
                  >
                    <input
                      type="radio"
                      name="payment_method"
                      checked={method === m.id}
                      disabled={!m.eligible || action.busy}
                      onChange={() => {
                        setMethod(m.id);
                        setQuote(null);
                        setConsent(false);
                      }}
                    />
                    <span>
                      <b>{m.name}</b>
                      <small>{m.eligible ? m.instructions : m.reason}</small>
                      {m.max_toman && (
                        <small>سقف کل سفارش: {money(m.max_toman)}</small>
                      )}
                    </span>
                  </label>
                ))}
                <p className="buy-note">
                  هر واریز ابتدا کیف پول را شارژ می‌کند. تأیید واریز به‌تنهایی
                  سفارش را پرداخت‌شده نمی‌کند؛ تیم پس از کنترل اعتبار و رزرو،
                  سفارش را تأیید می‌کند.
                </p>
                <button
                  className="buy-primary"
                  disabled={action.busy || !method}
                  onClick={() =>
                    void action.run(async (key) => {
                      setQuote(
                        await call<Quote>(
                          "/storefront/checkout/quotes",
                          token,
                          input,
                          undefined,
                          key,
                        ),
                      );
                      setConsent(false);
                    }, JSON.stringify(input))
                  }
                >
                  بررسی نهایی سفارش
                </button>
              </section>
            )}
            {quote && (
              <section className="buy-panel">
                <h2>بازبینی و تأیید مبلغ</h2>
                <OrderSummary quote={quote} />
                <label className="buy-choice">
                  <input
                    type="checkbox"
                    checked={consent}
                    onChange={(e) => setConsent(e.target.checked)}
                  />
                  <span>
                    مبلغ {money(quote.total_toman)}، آدرس و شرایط ارسال را تأیید
                    می‌کنم. رزرو ۳۰ دقیقه اعتبار دارد.
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
                        },
                        undefined,
                        key,
                      );
                      await refreshCart();
                      navigate("/account/orders/" + result.id);
                    }, "accept:" + quote.id)
                  }
                >
                  ثبت سفارش و ادامه پرداخت
                </button>
              </section>
            )}
          </section>
          <aside className="buy-panel buy-summary">
            <h2>خلاصه خرید</h2>
            {preview ? (
              <OrderSummary quote={preview.quote} />
            ) : (
              <>
                <p>{prepare.cart.item_count} کالا</p>
                <strong>{money(prepare.cart.subtotal_toman)}</strong>
                <p>هزینه ارسال پس از انتخاب آدرس و روش ارسال محاسبه می‌شود.</p>
              </>
            )}
            <a href="/cart">بازگشت به سبد</a>
          </aside>
        </div>
      )}
    </>
  );
}
export function OrderSummary({ quote }: { quote: Quote }) {
  return (
    <div className="buy-order-summary">
      {quote.lines.map((l) => (
        <div className="buy-total" key={l.offer_id}>
          <span>
            {l.name}
            <small>
              {l.variant_label} · تعداد {l.quantity}
            </small>
          </span>
          <strong>{money(l.total_toman)}</strong>
        </div>
      ))}
      {quote.shipments.map((s) => (
        <div className="buy-shipment" key={s.store_id}>
          <b>
            {s.store_name} · {s.method_name}
          </b>
          <p>هزینه ارسال: {money(s.fee_toman)}</p>
          <p>
            تأمین {s.lead_days} روز کاری؛ تحویل از{" "}
            {new Date(s.delivery_from).toLocaleDateString("fa-IR")} تا{" "}
            {new Date(s.delivery_to).toLocaleDateString("fa-IR")}
          </p>
        </div>
      ))}
      <div className="buy-total">
        <span>جمع کالاها</span>
        <b>{money(quote.subtotal_toman)}</b>
      </div>
      <div className="buy-total">
        <span>جمع ارسال</span>
        <b>{money(quote.shipping_toman)}</b>
      </div>
      <div className="buy-total buy-grand">
        <span>مبلغ نهایی</span>
        <b>{money(quote.total_toman)}</b>
      </div>
    </div>
  );
}
export function PurchaseStatus({ purchase: p }: { purchase: Purchase }) {
  return (
    <>
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
      {p.reason && <p className="buy-alert">{p.reason}</p>}
      <details className="buy-history">
        <summary>تاریخچه سفارش و پرداخت</summary>
        <ol>
          {p.events.map((e, i) => (
            <li key={i}>
              <b>{label(e.action)}</b>
              <time>{new Date(e.at).toLocaleString("fa-IR")}</time>
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
        <p>در حال دریافت سفارش…</p>
      ) : !rows.length ? (
        <div className="buy-panel">هنوز سفارشی ثبت نشده است.</div>
      ) : (
        rows
          .slice()
          .reverse()
          .map((p) => (
            <article className="buy-panel" key={p.id}>
              <h2>
                {id ? (
                  "سفارش شما"
                ) : (
                  <a href={"/account/orders/" + p.id}>مشاهده سفارش</a>
                )}
              </h2>
              <p className="buy-note" dir="ltr">
                {p.id}
              </p>
              <PurchaseStatus purchase={p} />
              {!id ? (
                <strong>{money(p.quote.total_toman)}</strong>
              ) : (
                <div className="buy-layout">
                  <section>
                    <OrderSummary quote={p.quote} />
                    <h3>گیرنده و آدرس</h3>
                    <p>
                      {p.quote.address.recipient} · {p.quote.address.phone}
                    </p>
                    <p>
                      {p.quote.address.province}، {p.quote.address.city}،{" "}
                      {p.quote.address.street} · {p.quote.address.postal_code}
                    </p>
                    <h3>اعلام واریزها</h3>
                    <Deposits
                      rows={p.deposits}
                      token={token}
                      onRefresh={load}
                    />
                  </section>
                  <aside>
                    <h3>{p.quote.payment_method.name}</h3>
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
                        {p.quote.payment_method.kind !== "wallet" ? (
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
                          پس از تأمین اعتبار، منتظر تأیید مستقل تیم بمانید. ثبت
                          رسید یا داشتن اعتبار، پرداخت سفارش محسوب نمی‌شود.
                        </p>
                      </>
                    )}
                    {p.status !== "cancelled" && (
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
                {label(d.status)} · {d.date} · مرجع {d.reference}
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
                            "/storefront/wallet/deposits/" + d.id + "/receipt",
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
    <p>دریافت رسید…</p>
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
    [method, setMethod] = useState(methods[0]?.id ?? "");
  const chosen = methods.find((m) => m.id === method);
  return (
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
        <label className="buy-field">
          روش تأمین اعتبار
          <select value={method} onChange={(e) => setMethod(e.target.value)}>
            {methods
              .filter((m) => m.kind !== "wallet")
              .map((m) => (
                <option key={m.id} value={m.id}>
                  {m.name}
                </option>
              ))}
          </select>
        </label>
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
        <Input
          name="date"
          label="تاریخ واریز (میلادی)"
          type="date"
          defaultValue={config?.server_date}
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
  );
}
export function WalletPage() {
  const { token } = useCommerce();
  return (
    <CommerceShell title="کیف پول">
      <RequireLogin>{token && <WalletContent />}</RequireLogin>
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
        <p>دریافت کیف پول…</p>
      ) : (
        <>
          <div className="buy-wallet-card">
            <span>اعتبار قابل استفاده</span>
            <strong>{money(wallet.available_toman)}</strong>
            <small>
              کل موجودی: {money(wallet.balance_toman)} · در انتظار بازپرداخت:{" "}
              {money(wallet.held_toman)}
            </small>
            <div className="buy-row">
              <button
                onClick={() => setTopup(true)}
                disabled={!wallet.funding_methods.length}
              >
                شارژ کیف پول
              </button>
              <button onClick={() => setRefund(true)}>درخواست بازپرداخت</button>
            </div>
          </div>
          <section className="buy-panel">
            <h2>گردش کیف پول</h2>
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
                  مبلغ تا رسیدگی تیم از اعتبار قابل استفاده کنار گذاشته می‌شود.
                  پرداخت بانکی به‌صورت دستی انجام خواهد شد.
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
  );
}
export function AccountPage({ addresses = false }: { addresses?: boolean }) {
  const { token, profile, logout } = useCommerce(),
    action = useAction();
  return (
    <CommerceShell title={addresses ? "آدرس‌های من" : "حساب کاربری"}>
      <RequireLogin>
        {addresses ? (
          <AddressBook />
        ) : (
          <>
            <nav className="buy-account-nav">
              <a href="/account/orders">سفارش‌ها</a>
              <a href="/account/addresses">آدرس‌ها</a>
              <a href="/wallet">کیف پول</a>
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
                ایمیل اختیاری است و برای ورود با رمز استفاده می‌شود؛ ورود با OTP
                از طریق موبایل است.
              </p>
              <form
                key={profile?.name + ":" + profile?.email}
                onSubmit={(e) => {
                  const v = values(e);
                  void action.run(async () => {
                    await call(
                      "/storefront/account/profile",
                      token,
                      { name: v.name, email: v.email, password: v.password },
                      "PATCH",
                    );
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
    </CommerceShell>
  );
}
export function BusinessPage({ slug }: { slug: string }) {
  const { config } = useCommerce();
  const keys: Record<
    string,
    [string, keyof NonNullable<typeof config>["business"]]
  > = {
    about: ["درباره چریوا", "about"],
    contact: ["تماس با ما", "contact"],
    shipping: ["شرایط ارسال", "shipping_policy"],
    returns: ["شرایط بازگشت", "return_policy"],
    privacy: ["حریم خصوصی", "privacy"],
    terms: ["قوانین خرید", "terms"],
  };
  const entry = keys[slug];
  return (
    <CommerceShell title={entry?.[0] ?? "صفحه پیدا نشد"}>
      <section className="buy-panel buy-prose">
        {entry
          ? config?.business[entry[1]] ||
            "اطلاعات این صفحه هنوز تکمیل نشده است."
          : ""}
        {config?.business.support_phone && (
          <p>
            پشتیبانی:{" "}
            <a href={"tel:" + config.business.support_phone}>
              {config.business.support_phone}
            </a>
          </p>
        )}
      </section>
    </CommerceShell>
  );
}
