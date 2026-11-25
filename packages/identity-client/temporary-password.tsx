import { useState } from "react";
import "./temporary-password.css";
export function TemporaryPassword<T>({
  token,
  surface,
  base = "",
  onComplete,
  onCancel,
}: {
  token: string;
  surface: string;
  base?: string;
  onComplete: (value: T) => void | Promise<void>;
  onCancel: () => void;
}) {
  const [password, setPassword] = useState(""),
    [confirmation, setConfirmation] = useState(""),
    [error, setError] = useState(""),
    [busy, setBusy] = useState(false);
  return (
    <section className="credential-change" dir="rtl">
      <span>امنیت حساب چریوا</span>
      <h1>رمز شخصی خود را تعیین کنید</h1>
      <p>
        رمز تعیین‌شده توسط ادمین موقت است؛ برای ادامه یک رمز تازه انتخاب کنید.
      </p>
      <form
        onSubmit={async (e) => {
          e.preventDefault();
          if (busy) return;
          setError("");
          if (password !== confirmation) {
            setError("رمز و تکرار آن یکسان نیستند.");
            return;
          }
          if (
            new TextEncoder().encode(password).length < 8 ||
            new TextEncoder().encode(password).length > 72
          ) {
            setError("رمز باید بین ۸ تا ۷۲ بایت باشد.");
            return;
          }
          setBusy(true);
          try {
            const r = await fetch(base + `/${surface}/auth/complete-password`, {
              method: "POST",
              cache: "no-store",
              headers: {
                "Content-Type": "application/json",
                Authorization: `Bearer ${token}`,
              },
              body: JSON.stringify({ password }),
            });
            if (!r.ok)
              throw Error(
                r.status === 400
                  ? "رمز تازه معتبر نیست؛ از رمز موقت متفاوت باشد."
                  : "نشست معتبر نیست؛ دوباره وارد شوید.",
              );
            const next = (await r.json()) as T;
            setPassword("");
            setConfirmation("");
            await onComplete(next);
          } catch (e) {
            setError((e as Error).message);
          } finally {
            setBusy(false);
          }
        }}
      >
        <label>
          رمز شخصی جدید
          <input
            type="password"
            autoComplete="new-password"
            required
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            disabled={busy}
          />
        </label>
        <label>
          تکرار رمز جدید
          <input
            type="password"
            autoComplete="new-password"
            required
            value={confirmation}
            onChange={(e) => setConfirmation(e.target.value)}
            disabled={busy}
          />
        </label>
        {error && <p role="alert">{error}</p>}
        <button disabled={busy}>ثبت رمز شخصی و ادامه</button>
        <button type="button" disabled={busy} onClick={onCancel}>
          خروج و بازگشت به ورود
        </button>
      </form>
    </section>
  );
}
