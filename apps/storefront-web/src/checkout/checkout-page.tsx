import { ToggleSwitch } from "../toggle-switch";
import { useEffect, useRef, useState } from "react";
import { products, stores, money, lineKey, type CartLine } from "../catalog";
import { Icon, Brand } from "../components";
import { productPhoto } from "../product-photos";
import { AuthPanel, type DemoAuth } from "./demo-auth";
import { AddressStep } from "./address-step";
import { PaymentStep } from "./payment-step";
import {
  CHECKOUT_DURATION,
  cartGroups,
  itemsTotal,
  selectedShipping,
  shippingOptions,
  validCart,
  validateAddress,
  type PaymentSession,
  type CheckoutDraft,
} from "./model";

type Step = "account" | "address" | "shipping" | "payment";
const steps: { id: Step; title: string }[] = [
  { id: "account", title: "ورود" },
  { id: "address", title: "اطلاعات و آدرس" },
  { id: "shipping", title: "ارسال" },
  { id: "payment", title: "پرداخت" },
];
export function CheckoutPage({
  cart,
  auth,
  draft,
  onChange,
  onExit,
}: {
  cart: CartLine[];
  auth: DemoAuth;
  draft: CheckoutDraft;
  onChange: (draft: CheckoutDraft) => void;
  onExit: () => void;
}) {
  const [step, setStep] = useState<Step>(auth.customer ? "address" : "account");
  const [accepted, setAccepted] = useState(false);
  const [session, setSession] = useState<PaymentSession | null>(null);
  const [now, setNow] = useState(Date.now());
  const [error, setError] = useState("");
  const heading = useRef<HTMLHeadingElement>(null);
  useEffect(() => {
    heading.current?.focus();
    window.scrollTo(0, 0);
  }, [step]);
  useEffect(() => {
    if (!session) return;
    const update = () => setNow(Date.now());
    const timer = setInterval(update, 1000);
    document.addEventListener("visibilitychange", update);
    return () => {
      clearInterval(timer);
      document.removeEventListener("visibilitychange", update);
    };
  }, [session]);
  const groups = cartGroups(cart);
  const address = draft.addresses.find(
    (address) => address.id === draft.addressId,
  );
  const choices = selectedShipping(cart, draft);
  const allShipping =
    groups.length > 0 && choices.every((choice) => choice.option);
  const subtotal = itemsTotal(cart);
  const shipping = choices.reduce(
    (total, choice) => total + (choice.option?.cost ?? 0),
    0,
  );
  const expired = session !== null && now >= session.expiresAt;
  function backTo(step: Step) {
    if (expired) {
      onExit();
      return;
    }
    setSession(null);
    setAccepted(false);
    setError("");
    setStep(step);
  }
  function beginPayment() {
    if (
      !auth.customer ||
      !address ||
      validateAddress(address) ||
      draft.profile.firstName.trim().length < 2 ||
      draft.profile.lastName.trim().length < 2
    ) {
      setError("اطلاعات حساب و آدرس باید کامل باشند.");
      return;
    }
    if (!validCart(cart) || !allShipping || !accepted) {
      setError("سبد، روش ارسال و تأیید مبلغ را دوباره بررسی کنید.");
      return;
    }
    const time = Date.now();
    setNow(time);
    setSession({
      expiresAt: time + CHECKOUT_DURATION,
      subtotal,
      shipping,
      total: subtotal + shipping,
      address: { ...address },
    });
    setError("");
    setStep("payment");
  }
  return (
    <div className="checkout-page">
      <div className="announcement">
        نسخه بررسی · ورود، آدرس‌ها و مبالغ نمونه‌اند؛ پرداخت واقعی فعال نیست.
      </div>
      <header className="checkout-header wrap">
        <Brand onClick={onExit} />
        <button className="text-button" onClick={onExit}>
          بازگشت به سبد خرید <Icon name="bag" />
        </button>
      </header>
      <main className="wrap checkout-main">
        <div className="checkout-heading">
          <div>
            <span className="overline">خرید به سلیقه تو</span>
            <h1 ref={heading} tabIndex={-1}>
              {step === "payment"
                ? "یک قدم تا نهایی‌شدن انتخابت."
                : "انتخاب‌های خوب، در مسیر تو."}
            </h1>
          </div>
          <span className="checkout-safe">◈ تجربه خرید چریوا</span>
        </div>
        <nav className="checkout-stepper" aria-label="مراحل خرید">
          {steps.map((entry, index) => (
            <button
              key={entry.id}
              aria-current={step === entry.id ? "step" : undefined}
              disabled={
                index >= steps.findIndex((entry) => entry.id === step) ||
                (expired && entry.id !== step)
              }
              onClick={() =>
                backTo(
                  entry.id === "account" && auth.customer
                    ? "address"
                    : entry.id,
                )
              }
              className={
                index <= steps.findIndex((entry) => entry.id === step)
                  ? "reached"
                  : ""
              }
            >
              <b>
                {index < steps.findIndex((entry) => entry.id === step)
                  ? "✓"
                  : money(index + 1)}
              </b>
              <span>{entry.title}</span>
            </button>
          ))}
        </nav>
        {!validCart(cart) ? (
          <section className="checkout-card empty">
            <Icon name="bag" />
            <h2>سبد قابل ادامه نیست.</h2>
            <p>یک کالای موجود با تعداد معتبر انتخاب کن.</p>
            <button className="primary" onClick={onExit}>
              بازگشت به سبد
            </button>
          </section>
        ) : (
          <div className="checkout-layout">
            <div className="checkout-content">
              {step === "account" && (
                <section className="checkout-card">
                  <AuthPanel auth={auth} onDone={() => setStep("address")} />
                </section>
              )}
              {step === "address" && (
                <AddressStep
                  draft={draft}
                  onChange={onChange}
                  onContinue={() => {
                    setAccepted(false);
                    setStep("shipping");
                  }}
                />
              )}
              {step === "shipping" && (
                <>
                  <div className="checkout-section-title">
                    <span>۰۲</span>
                    <div>
                      <h2>انتخاب روش ارسال</h2>
                      <p>بسته‌های هر فروشگاه جداگانه به یک مقصد می‌رسند.</p>
                    </div>
                  </div>
                  <section className="checkout-card destination-summary">
                    <div>
                      <span className="checkout-muted">ارسال به</span>
                      <h3>
                        {address?.title} · {address?.city}
                      </h3>
                      <p>{address?.street}</p>
                      <small>گیرنده: {address?.recipient}</small>
                    </div>
                    <button
                      className="text-button"
                      onClick={() => backTo("address")}
                    >
                      تغییر آدرس
                    </button>
                  </section>
                  {groups.map((store) => {
                    const options = shippingOptions(
                      store.id,
                      address?.city ?? "",
                    );
                    return (
                      <section
                        className="checkout-card shipping-group"
                        key={store.id}
                      >
                        <div className="shipping-group-head">
                          <div>
                            <h3>{store.name}</h3>
                            <p>آماده‌سازی نمونه: {store.preparation}</p>
                          </div>
                          <span>
                            {money(
                              store.lines.reduce(
                                (count, line) => count + line.quantity,
                                0,
                              ),
                            )}{" "}
                            کالا
                          </span>
                        </div>
                        <div className="shipment-products">
                          {store.lines.map((line) => {
                            const product = products.find(
                              (product) => product.id === line.productId,
                            )!;
                            return (
                              <div key={lineKey(line)}>
                                <img
                                  src={productPhoto(product, line.color)}
                                  alt=""
                                />
                                <span>
                                  {product.name}
                                  <small>
                                    {line.color} · {money(line.quantity)} عدد
                                  </small>
                                </span>
                              </div>
                            );
                          })}
                        </div>
                        <fieldset className="shipping-options">
                          <legend>روش ارسال {store.name}</legend>
                          {options.length ? (
                            options.map((option) => (
                              <label
                                key={option.id}
                                className={
                                  draft.shipping[store.id] === option.id
                                    ? "shipping-option selected"
                                    : "shipping-option"
                                }
                              >
                                <input
                                  type="radio"
                                  name={`shipping-${store.id}`}
                                  checked={
                                    draft.shipping[store.id] === option.id
                                  }
                                  onChange={() => {
                                    onChange({
                                      ...draft,
                                      shipping: {
                                        ...draft.shipping,
                                        [store.id]: option.id,
                                      },
                                    });
                                    setAccepted(false);
                                  }}
                                />
                                <span>
                                  <strong>{option.title}</strong>
                                  <small>زمان حمل: {option.transit}</small>
                                  <small>
                                    برآورد دریافت نمونه: {option.arrival}
                                  </small>
                                </span>
                                <b>
                                  {money(option.cost)} <small>تومان</small>
                                </b>
                              </label>
                            ))
                          ) : (
                            <p className="form-error" role="alert">
                              برای این مقصد روش ارسال نمونه نداریم. آدرس تهران
                              یا کرج را انتخاب کن.
                            </p>
                          )}
                        </fieldset>
                      </section>
                    );
                  })}
                  <section className="checkout-card">
                    <label className="confirmation">
                      <ToggleSwitch
                        checked={accepted}
                        disabled={!allShipping}
                        onChange={setAccepted}
                      />
                      <span>
                        اقلام، آدرس، هزینه و زمان ارسال نمونه را بررسی کردم و
                        مبلغ <strong>{money(subtotal + shipping)} تومان</strong>{" "}
                        را تأیید می‌کنم.
                      </span>
                    </label>
                    <p className="checkout-muted">
                      با «ادامه پرداخت» شمارش مهلت ۱۰ دقیقه‌ای شروع می‌شود؛ این
                      نسخه موجودی واقعی رزرو نمی‌کند.
                    </p>
                    {error && (
                      <p className="form-error" role="alert">
                        {error}
                      </p>
                    )}
                    <button
                      className="primary full"
                      disabled={!allShipping || !accepted}
                      onClick={beginPayment}
                    >
                      ادامه پرداخت <Icon name="arrow" />
                    </button>
                  </section>
                  <button
                    className="text-button"
                    onClick={() => backTo("address")}
                  >
                    → بازگشت به اطلاعات و آدرس
                  </button>
                </>
              )}
              {step === "payment" && session && (
                <PaymentStep
                  session={session}
                  now={now}
                  expired={expired}
                  onExit={onExit}
                  onBack={() => backTo("shipping")}
                />
              )}
            </div>
            <aside className="checkout-summary">
              <section className="checkout-card">
                <h2>خلاصه خرید تو</h2>
                <div className="summary-products">
                  {cart.map((line) => {
                    const product = products.find(
                      (product) => product.id === line.productId,
                    )!;
                    return (
                      <div key={lineKey(line)}>
                        <img src={productPhoto(product, line.color)} alt="" />
                        <span>
                          <strong>{product.name}</strong>
                          <small>
                            {line.color} · {money(line.quantity)} عدد
                          </small>
                          <small>
                            {
                              stores.find((store) => store.id === line.storeId)!
                                .name
                            }
                          </small>
                        </span>
                      </div>
                    );
                  })}
                </div>
                <dl className="checkout-totals">
                  <div>
                    <dt>جمع اقلام</dt>
                    <dd>{money(session?.subtotal ?? subtotal)} تومان</dd>
                  </div>
                  {choices.map(({ store, option }) => (
                    <div key={store.id}>
                      <dt>ارسال {store.name.replace("فروشگاه ", "")}</dt>
                      <dd>
                        {option
                          ? `${money(option.cost)} تومان`
                          : "پس از انتخاب"}
                      </dd>
                    </div>
                  ))}
                  <div className="summary-total">
                    <dt>
                      {allShipping || session
                        ? "جمع قابل پرداخت نمونه"
                        : "جمع اقلام تا این مرحله"}
                    </dt>
                    <dd>
                      {money(session?.total ?? subtotal + shipping)}
                      <small> تومان</small>
                    </dd>
                  </div>
                </dl>
                <p className="demo-note">
                  این خرید آزمایشی است. هیچ سفارش یا تراکنش واقعی ثبت نمی‌شود.
                </p>
              </section>
              <p className="checkout-privacy">
                اطلاعات فقط در حافظه همین صفحه می‌ماند؛ با بارگذاری مجدد پاک
                می‌شود.
              </p>
            </aside>
          </div>
        )}
      </main>
      <footer className="checkout-footer wrap">
        Cherriva · انتخاب‌های متفاوت، یک تجربه مشترک
      </footer>
    </div>
  );
}
