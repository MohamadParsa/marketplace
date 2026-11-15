import { useEffect, useState, type FormEvent } from "react";
import { DEMO_CODE, DEMO_PHONE, digits } from "./model";
import { money } from "../catalog";

export function useDemoAuth() {
  const [customer, setCustomer] = useState<string | null>(null);
  const [requestedAt, setRequestedAt] = useState<number | null>(null);
  const [requests, setRequests] = useState<number[]>([]);
  const [attempts, setAttempts] = useState<number[]>([]);
  const [now, setNow] = useState(Date.now());
  const [error, setError] = useState("");
  useEffect(() => {
    if (requestedAt === null) return;
    const timer = setInterval(() => {
      const time = Date.now();
      setNow(time);
      if (time >= requestedAt + 90000) clearInterval(timer);
    }, 1000);
    return () => clearInterval(timer);
  }, [requestedAt]);
  const secondsLeft =
    requestedAt === null
      ? 0
      : Math.max(0, Math.ceil((requestedAt + 90000 - now) / 1000));
  const resendIn =
    requestedAt === null
      ? 0
      : Math.max(0, Math.ceil((requestedAt + 60000 - now) / 1000));
  function request(phone: string) {
    const time = Date.now();
    if (digits(phone) !== DEMO_PHONE) {
      setError(
        "این نسخه فقط شماره حساب نمونه ۰۹۰۰۰۰۰۰۰۰۰۳ را می‌پذیرد؛ پیامک واقعی ارسال نمی‌شود.",
      );
      return false;
    }
    if (requestedAt !== null && time < requestedAt + 60000) {
      setError("برای درخواست دوباره کد، تا پایان شمارش صبر کنید.");
      return false;
    }
    const recent = requests.filter((timestamp) => timestamp > time - 3600000);
    if (
      recent.length >= 5 ||
      attempts.filter((timestamp) => timestamp > time - 3600000).length >= 5
    ) {
      setError("سقف ۵ تلاش در یک ساعت تکمیل شده است.");
      return false;
    }
    setRequests([...recent, time]);
    setRequestedAt(time);
    setNow(time);
    setError("");
    return true;
  }
  function verify(code: string) {
    const time = Date.now();
    const recent = attempts.filter((timestamp) => timestamp > time - 3600000);
    if (recent.length >= 5) {
      setError("سقف ۵ تلاش در یک ساعت تکمیل شده است.");
      return false;
    }
    if (requestedAt === null || time >= requestedAt + 90000) {
      setError("کد منقضی شده است؛ دوباره کد بگیرید.");
      return false;
    }
    setAttempts([...recent, time]);
    if (digits(code) !== DEMO_CODE) {
      setError("کد واردشده صحیح نیست.");
      return false;
    }
    setCustomer(DEMO_PHONE);
    setRequestedAt(null);
    setError("");
    return true;
  }
  return {
    customer,
    requestedAt,
    secondsLeft,
    resendIn,
    error,
    request,
    verify,
  };
}
export type DemoAuth = ReturnType<typeof useDemoAuth>;
export function AuthPanel({
  auth,
  onDone,
}: {
  auth: DemoAuth;
  onDone?: () => void;
}) {
  const [phone, setPhone] = useState(DEMO_PHONE);
  const [code, setCode] = useState("");
  function submit(event: FormEvent) {
    event.preventDefault();
    if (auth.requestedAt === null) auth.request(phone);
    else if (auth.verify(code)) onDone?.();
  }
  if (auth.customer)
    return (
      <div className="auth-success">
        <span>✓</span>
        <h3>به حساب نمونه وارد شده‌اید</h3>
        <p dir="ltr">{auth.customer}</p>
        <p>این ورود فقط برای بررسی رابط است و اعتبار احراز هویت واقعی ندارد.</p>
        {onDone && (
          <button className="primary" onClick={onDone}>
            ادامه خرید
          </button>
        )}
      </div>
    );
  return (
    <div className="auth-panel">
      <span className="overline">یک قدم نزدیک‌تر به انتخابت</span>
      <h2>ورود به چریوا</h2>
      <p className="checkout-muted">سبد انتخاب‌هایت بعد از ورود حفظ می‌شود.</p>
      <p className="demo-note">
        ورود آزمایشی است؛ فقط از شماره نمونه استفاده کن. پیامکی ارسال نمی‌شود.
      </p>
      <form onSubmit={submit}>
        <label htmlFor="customer-phone">شماره همراه حساب</label>
        <input
          id="customer-phone"
          className="text-input ltr-input"
          type="tel"
          inputMode="numeric"
          autoComplete="off"
          value={phone}
          disabled={auth.requestedAt !== null}
          onChange={(event) => setPhone(digits(event.target.value))}
        />
        {auth.requestedAt !== null && (
          <>
            <label htmlFor="customer-code">کد یک‌بارمصرف</label>
            <input
              autoFocus
              id="customer-code"
              className="text-input ltr-input"
              inputMode="numeric"
              autoComplete="off"
              maxLength={6}
              value={code}
              onChange={(event) => setCode(digits(event.target.value))}
            />
            <div className="otp-meta">
              <span>اعتبار کد: {money(auth.secondsLeft)} ثانیه</span>
              <button
                type="button"
                disabled={auth.resendIn > 0}
                onClick={() => {
                  if (auth.request(phone)) setCode("");
                }}
              >
                {auth.resendIn > 0
                  ? `ارسال مجدد پس از ${money(auth.resendIn)} ثانیه`
                  : "دریافت دوباره کد"}
              </button>
            </div>
            <div className="sample-code">
              کد نمونه برای بررسی: <b dir="ltr">{DEMO_CODE}</b>
            </div>
          </>
        )}
        {auth.error && (
          <p className="form-error" role="alert">
            {auth.error}
          </p>
        )}
        <button
          className="primary full"
          type="submit"
          disabled={auth.requestedAt !== null && auth.secondsLeft === 0}
        >
          {auth.requestedAt === null ? "دریافت کد آزمایشی" : "تأیید و ادامه"}
        </button>
      </form>
    </div>
  );
}
