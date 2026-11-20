import { useState } from "react";
import { Alert, Button, TextField } from "@mui/material";
import { ArrowBack, ShieldOutlined } from "@mui/icons-material";
import { useForm } from "react-hook-form";
import { z } from "zod";
import { zodResolver } from "@hookform/resolvers/zod";
import { api } from "../shared/api";
import type { Session } from "../shared/types";
const phoneSchema = z.object({
  phone: z.string().regex(/^09\d{9}$/, "شماره باید ۱۱ رقم و با ۰۹ شروع شود."),
});
const accounts = [
  { name: "مدیر کل", phone: "09000000001" },
  { name: "عملیات", phone: "09000000002" },
  { name: "پشتیبانی", phone: "09000000003" },
  { name: "مالی", phone: "09000000004" },
  { name: "مشاهده‌گر", phone: "09000000005" },
];
export function LoginPage({
  onLogin,
}: {
  onLogin: (session: Session) => void;
}) {
  const [challenge, setChallenge] = useState("");
  const [code, setCode] = useState("");
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  const [testCode, setTestCode] = useState("");
  const [phone, setPhone] = useState("");
  const [passwordAvailable, setPasswordAvailable] = useState(false);
  const [passwordMode, setPasswordMode] = useState(false);
  const [password, setPassword] = useState("");
  const form = useForm<{ phone: string }>({
    resolver: zodResolver(phoneSchema),
    defaultValues: { phone: "09000000001" },
  });
  const handleError = (error: unknown) =>
    setError(
      error instanceof Error ? error.message : "ارتباط با سرور برقرار نشد.",
    );
  return (
    <div className="login-layout">
      <section className="login-story">
        <div className="brand">
          <span className="brand-mark">م</span>
          <span>
            مارکت‌پلیس<small>مرکز مدیریت بازارگاه</small>
          </span>
        </div>
        <div>
          <span className="overline">همه‌چیز، در یک نگاه.</span>
          <h1>
            یک فضای روشن
            <br />
            برای مدیریت
            <br />
            <em>تمام بازارگاه.</em>
          </h1>
          <p>
            از اولین عرضه تا آخرین تسویه؛
            <br />
            عملیات، فروشندگان و تجربه مشتری را دنبال کنید.
          </p>
        </div>
        <span className="story-footer">بک‌آفیس بازارگاه · نسخه بررسی</span>
      </section>
      <section className="login-form">
        <div className="login-card">
          <ShieldOutlined className="login-icon" />
          <h2>ورود به بک‌آفیس</h2>
          <p>برای دسترسی به پنل، شماره حساب آزمایشی را وارد کنید.</p>
          <Alert severity="info" sx={{ my: 3 }}>
            محیط آزمون با داده ساختگی؛ ارسال پیامک و انتقال وجه واقعی غیرفعال
            است.
          </Alert>
          {!challenge ? (
            <form
              onSubmit={form.handleSubmit(async (values) => {
                setBusy(true);
                setError("");
                try {
                  const result = await api.requestCode(values.phone);
                  setPhone(values.phone);
                  setPasswordAvailable(result.password_available === true);
                  setChallenge(result.request_id);
                  setTestCode(result.test_code ?? "");
                } catch (error) {
                  handleError(error);
                } finally {
                  setBusy(false);
                }
              })}
            >
              <TextField
                label="شماره تلفن"
                fullWidth
                dir="ltr"
                {...form.register("phone")}
                error={!!form.formState.errors.phone}
                helperText={form.formState.errors.phone?.message}
              />
              <Button
                fullWidth
                size="large"
                type="submit"
                variant="contained"
                disabled={busy}
                sx={{ mt: 2 }}
                endIcon={<ArrowBack />}
              >
                {busy ? "در حال درخواست…" : "دریافت کد ورود"}
              </Button>
            </form>
          ) : (
            <form
              onSubmit={async (event) => {
                event.preventDefault();
                setBusy(true);
                setError("");
                try {
                  onLogin(
                    passwordMode
                      ? await api.passwordLogin(phone, password)
                      : await api.verify(challenge, code),
                  );
                } catch (error) {
                  handleError(error);
                } finally {
                  setBusy(false);
                }
              }}
            >
              {passwordMode ? (
                <>
                  <TextField
                    fullWidth
                    label="رمز ورود"
                    type="password"
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    autoFocus
                  />
                  <Button
                    fullWidth
                    size="large"
                    type="submit"
                    variant="contained"
                    disabled={busy || password.length === 0}
                    sx={{ mt: 2 }}
                  >
                    {busy ? "در حال بررسی…" : "ورود با رمز"}
                  </Button>
                  <Button type="button" onClick={() => setPasswordMode(false)} sx={{ mt: 1 }}>
                    بازگشت به ورود با کد
                  </Button>
                </>
              ) : (
                <>
              {testCode && (
                <Alert severity="warning" sx={{ mb: 2 }}>
                  کد آزمون: <b dir="ltr">{testCode}</b>
                </Alert>
              )}
              <TextField
                fullWidth
                label="کد یک‌بارمصرف"
                value={code}
                onChange={(e) => setCode(e.target.value)}
                inputProps={{ inputMode: "numeric", maxLength: 6 }}
              />
              <Button
                fullWidth
                size="large"
                type="submit"
                variant="contained"
                disabled={busy}
                sx={{ mt: 2 }}
              >
                {busy ? "در حال بررسی…" : "ورود به پنل"}
              </Button>
              {passwordAvailable && (
                <Button type="button" fullWidth onClick={() => setPasswordMode(true)} sx={{ mt: 1 }}>
                  ورود با رمز
                </Button>
              )}
              <Button
                onClick={() => {
                  setChallenge("");
                  setCode("");
                  setPassword("");
                  setPasswordMode(false);
                  setPasswordAvailable(false);
                }}
                sx={{ mt: 1 }}
              >
                تغییر شماره
              </Button>
                </>
              )}
            </form>
          )}
          {error && (
            <Alert severity="error" sx={{ mt: 2 }}>
              {error}
            </Alert>
          )}
          <div className="test-accounts">
            <span>حساب‌های ساختگی برای بررسی نقش‌ها</span>
            <div>
              {accounts.map((account) => (
                <button
                  key={account.phone}
                  disabled={!!challenge}
                  onClick={() => form.setValue("phone", account.phone)}
                >
                  {account.name}
                </button>
              ))}
            </div>
          </div>
          <small className="muted">
            ورود صرفاً برای محیط آزمون است. سیاست رمزعبور بک‌آفیس هنوز نهایی نشده
            است.
          </small>
        </div>
      </section>
    </div>
  );
}
