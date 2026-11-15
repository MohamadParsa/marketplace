import { useState } from "react";
import { money } from "../catalog";
import { DEMO_WALLET, digits, type PaymentSession } from "./model";
export function PaymentStep({
  session,
  now,
  expired,
  onExit,
  onBack,
}: {
  session: PaymentSession;
  now: number;
  expired: boolean;
  onExit: () => void;
  onBack: () => void;
}) {
  const [method, setMethod] = useState<"wallet" | "bnpl">("wallet");
  const deficit = Math.max(0, session.total - DEMO_WALLET);
  const [topUp, setTopUp] = useState(String(deficit));
  const remaining = Math.max(0, Math.ceil((session.expiresAt - now) / 1000));
  const topUpAmount = Number(topUp);
  const validTopUp =
    Number.isSafeInteger(topUpAmount) &&
    topUpAmount >= deficit &&
    topUpAmount > 0;
  return (
    <>
      <div className="checkout-section-title">
        <span>۰۳</span>
        <div>
          <h2>صفحه پرداخت</h2>
          <p>انتخاب روش پرداخت و بررسی مبلغ نهایی نمونه</p>
        </div>
      </div>
      <div
        className={`checkout-timer ${expired ? "expired" : ""}`}
        role="status"
      >
        <span>
          {expired ? "مهلت جلسه نمونه تمام شد" : "زمان باقی‌مانده جلسه نمونه"}
        </span>
        <strong dir="ltr" aria-label="زمان باقی‌مانده پرداخت">
          {String(Math.floor(remaining / 60)).padStart(2, "0")}:
          {String(remaining % 60).padStart(2, "0")}
        </strong>
      </div>
      {expired ? (
        <section className="checkout-card expiry-panel">
          <h3>برای ادامه، سبد را دوباره بررسی کن.</h3>
          <p>
            این جلسه دیگر قابل استفاده نیست. قیمت و موجودی نمونه باید از مسیر
            سبد بررسی شوند.
          </p>
          <button className="primary full" onClick={onExit}>
            بازگشت به سبد و بررسی دوباره
          </button>
        </section>
      ) : (
        <>
          <section className="checkout-card">
            <h3>روش پرداخت</h3>
            <label
              className={`payment-method ${method === "wallet" ? "selected" : ""}`}
            >
              <input
                type="radio"
                name="payment-method"
                checked={method === "wallet"}
                onChange={() => setMethod("wallet")}
              />
              <span>
                <strong>کیف پول و شارژ نقدی</strong>
                <small>کل مبلغ خرید از کیف پول پرداخت می‌شود.</small>
              </span>
              <b>◈</b>
            </label>
            <label
              className={`payment-method ${method === "bnpl" ? "selected" : ""}`}
            >
              <input
                type="radio"
                name="payment-method"
                checked={method === "bnpl"}
                onChange={() => setMethod("bnpl")}
              />
              <span>
                <strong>خرید اعتباری — BNPL</strong>
                <small>مسیر مستقل از کیف پول</small>
              </span>
              <b>◇</b>
            </label>
            {method === "wallet" ? (
              <div className="wallet-panel">
                <dl className="checkout-totals">
                  <div>
                    <dt>موجودی نمونه کیف پول</dt>
                    <dd>{money(DEMO_WALLET)} تومان</dd>
                  </div>
                  <div>
                    <dt>مبلغ خرید نمونه</dt>
                    <dd>{money(session.total)} تومان</dd>
                  </div>
                  <div className="wallet-deficit">
                    <dt>کسری کیف پول</dt>
                    <dd>{money(deficit)} تومان</dd>
                  </div>
                </dl>
                {deficit > 0 ? (
                  <>
                    <label htmlFor="wallet-topup">
                      مبلغ شارژ کیف پول (تومان)
                    </label>
                    <input
                      id="wallet-topup"
                      className="text-input ltr-input"
                      inputMode="numeric"
                      autoComplete="off"
                      value={topUp}
                      onChange={(event) => {
                        const value = digits(event.target.value);
                        if (/^\d*$/.test(value)) setTopUp(value);
                      }}
                    />
                    <p className="checkout-muted">
                      حداقل {money(deficit)} تومان؛ شارژ بیشتر از کسری هم مجاز
                      است. پس از شارژ، کل مبلغ خرید از کیف پول برداشت می‌شود.
                    </p>
                    {!validTopUp && (
                      <p className="form-error" role="alert">
                        مبلغ شارژ باید حداقل برابر کسری و یک عدد معتبر باشد.
                      </p>
                    )}
                    {validTopUp && (
                      <p className="checkout-muted">
                        اعتبار باقی‌مانده پس از خرید نمونه:{" "}
                        {money(DEMO_WALLET + topUpAmount - session.total)} تومان
                      </p>
                    )}
                    <button className="primary full" disabled>
                      شارژ کیف پول و ادامه — درگاه متصل نیست
                    </button>
                  </>
                ) : (
                  <button className="primary full" disabled>
                    پرداخت از کیف پول — سرویس متصل نیست
                  </button>
                )}
                <p className="demo-note">
                  صفحه تا انتخاب روش پرداخت آماده است؛ شارژ و برداشت واقعی انجام
                  نمی‌شود.
                </p>
              </div>
            ) : (
              <div className="bnpl-panel">
                <h3>خرید اعتباری، جدا از کیف پول</h3>
                <p>
                  ارائه‌دهنده خرید اعتباری هنوز انتخاب نشده است. شرایط اعتبار و
                  اقساط پس از اتصال سرویس نمایش داده می‌شوند.
                </p>
                <button className="primary full" disabled>
                  ادامه خرید اعتباری — سرویس متصل نیست
                </button>
              </div>
            )}
          </section>
          <button className="text-button" onClick={onBack}>
            → بازگشت و تغییر ارسال
          </button>
        </>
      )}
    </>
  );
}
