import {Link} from "react-router-dom";
import { useState } from "react";
import { Alert, Button, TextField } from "@mui/material";
import {
  LockOutlined,
  PersonOutline,
  StorefrontOutlined,
  ArrowBack,
} from "@mui/icons-material";
import { api } from "../shared/api";
import type { Session } from "../shared/types";
const normalize = (value: string) =>
  value
    .trim()
    .replace(/[۰-۹]/g, (c) => String(c.charCodeAt(0) - 1776))
    .replace(/[٠-٩]/g, (c) => String(c.charCodeAt(0) - 1632));
export function LoginPage({ onLogin, surface="backoffice" }: { onLogin: (s: Session) => void; surface?:"backoffice"|"seller" }) {
  const [identifier, setIdentifier] = useState("");
  const [stage, setStage] = useState<
    "identifier" | "method" | "password" | "otp"
  >("identifier");
  const [password, setPassword] = useState("");
  const [code, setCode] = useState("");
  const [challenge, setChallenge] = useState("");
  const [testCode, setTestCode] = useState("");
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  const isEmail = identifier.includes("@");
  async function run(fn: () => Promise<void>) {
    if (busy) return;
    setBusy(true);
    setError("");
    try {
      await fn();
    } catch (e) {
      setError(e instanceof Error ? e.message : "ارتباط با سرور برقرار نشد.");
    } finally {
      setBusy(false);
    }
  }
  function reset() {
    setStage("identifier");
    setPassword("");
    setCode("");
    setChallenge("");
    setError("");
    setTestCode("");
  }
  return (
    <div className="login-layout design-a-login">
      <section className="login-story">
        <div className="login-brand-row">
          <img
            src="/brand/cherriva-wordmark-white.svg"
            alt="Cherriva"
            width="180"
            height="50"
          />
          <span>
            یک حساب
            <br />
            برای همه نقش‌ها
          </span>
        </div>
        <div className="login-story-copy">
          <span className="overline">همان شما، در هر نقش</span>
          <h1>
            یک حساب،
            <br />
            چند تجربه.
            <br />
            <em>ساده و یکپارچه.</em>
          </h1>
          <p>
            خرید کنید، در بک‌آفیس فعالیت کنید یا عضو فروشگاه باشید؛
            <br />
            همه با یک هویت و دسترسی‌های مشخص.
          </p>
        </div>
        <div className="login-identity-card">
          <div>
            <PersonOutline />
            <span>
              <b>کاربر واحد</b>
              <small>خریدار، مدیر و عضو فروشگاه</small>
            </span>
          </div>
          <div>
            <StorefrontOutlined />
            <span>
              <b>دسترسی متناسب با نقش</b>
              <small>یک تجربه مشترک، با محدوده‌های مستقل</small>
            </span>
          </div>
        </div>
        <div className="story-footer">
          <span>Cherriva / فضای مدیریت</span>
        </div>
      </section>
      <section className="login-form">
        <div className="login-card">
          <div className="login-icon-box">
            <LockOutlined />
          </div>
          <h2>
            {stage === "identifier"
              ? "ورود به فضای مدیریت"
              : stage === "method"
                ? "روش ورود را انتخاب کنید"
                : stage === "password"
                  ? "ورود با رمز عبور"
                  : "کد یک‌بارمصرف را وارد کنید"}
          </h2>
          <p>
            {stage === "identifier"
              ? "با شماره موبایل یا ایمیل متصل به حساب خود وارد شوید."
              : isEmail
                ? "ورود با ایمیل، با رمز عبور حساب انجام می‌شود."
                : "ورود با رمز عبور یا دریافت کد روی شماره موبایل."}
          </p>
          {stage !== "identifier" && (
            <div className="login-identifier">
              <b dir="ltr">{identifier}</b>
              <button disabled={busy} onClick={reset}>
                تغییر نام کاربری
              </button>
            </div>
          )}
          {error && (
            <Alert severity="error" sx={{ my: 2 }}>
              {error}
            </Alert>
          )}
          {stage === "identifier" && (
            <form
              onSubmit={(e) => {
                e.preventDefault();
                setError("");
                const v = normalize(identifier);
                const email = /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(v);
                if (!/^09\d{9}$/.test(v) && !email) {
                  setError("شماره موبایل یا ایمیل معتبر وارد کنید.");
                  return;
                }
                setIdentifier(email ? v.toLowerCase() : v);
                setStage(email ? "password" : "method");
              }}
            >
              <TextField
                fullWidth
                label="شماره موبایل یا ایمیل"
                autoComplete="username"
                value={identifier}
                onChange={(e) => setIdentifier(e.target.value)}
                inputProps={{ dir: "ltr", maxLength: 254 }}
                required
              />
              <Button
                type="submit"
                fullWidth
                variant="contained"
                size="large"
                endIcon={<ArrowBack />}
              >
                ادامه
              </Button>
            </form>
          )}
          {stage === "method" && (
            <div className="login-methods">
              <Button
                fullWidth
                variant="contained"
                size="large"
                disabled={busy}
                onClick={() => {
                  setError("");
                  setStage("password");
                }}
              >
                ورود با رمز عبور
              </Button>
              <Button
                fullWidth
                variant="outlined"
                size="large"
                disabled={busy}
                onClick={() =>
                  void run(async () => {
                    const r = await api.requestCode(identifier,surface);
                    setChallenge(r.request_id);
                    setTestCode(r.test_code ?? "");
                    setStage("otp");
                  })
                }
              >
                دریافت کد یک‌بارمصرف
              </Button>
            </div>
          )}
          {stage === "password" && (
            <form
              onSubmit={(e) => {
                e.preventDefault();
                void run(async () =>
                  onLogin(await api.passwordLogin(identifier, password,surface)),
                );
              }}
            >
              <input type="hidden" name="username" autoComplete="username" value={identifier}/><TextField
                fullWidth
                label="رمز عبور"
                type="password"
                autoComplete="current-password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                inputProps={{ dir: "ltr", maxLength: 128 }}
                required
              />
              <Button
                type="submit"
                fullWidth
                variant="contained"
                size="large"
                disabled={busy}
              >
                ورود با رمز عبور
              </Button>
              {!isEmail && (
                <Button
                  fullWidth
                  disabled={busy}
                  onClick={() => {
                    setPassword("");
                    setError("");
                    setStage("method");
                  }}
                >
                  انتخاب روش دیگر
                </Button>
              )}
            </form>
          )}
          {stage === "otp" && (
            <form
              onSubmit={(e) => {
                e.preventDefault();
                void run(async () =>
                  onLogin(await api.verify(challenge, normalize(code),surface)),
                );
              }}
            >
              {testCode && (
                <Alert severity="warning" sx={{ my: 2 }}>
                  کد آزمون: {testCode}
                </Alert>
              )}
              <TextField
                fullWidth
                label="کد یک‌بارمصرف"
                autoComplete="one-time-code"
                value={code}
                onChange={(e) => setCode(normalize(e.target.value))}
                inputProps={{
                  dir: "ltr",
                  inputMode: "numeric",
                  pattern: "[0-9]{6}",
                  maxLength: 6,
                }}
                required
              />
              <Button
                fullWidth
                type="submit"
                variant="contained"
                size="large"
                disabled={busy}
              >
                ورود به پنل
              </Button>
              <Button
                fullWidth
                disabled={busy}
                onClick={() => {
                  setCode("");
                  setError("");
                  setStage("method");
                }}
              >
                انتخاب روش دیگر
              </Button>
            </form>
          )}
          <p className="login-security-note">
            ورود به این بخش فقط برای کاربران دارای دسترسی مجاز است.<br/><Link to={surface==="seller"?"/":"/seller-login"}>{surface==="seller"?"ورود کارکنان بک‌آفیس":"ورود اعضای فروشگاه"}</Link>
          </p>
        </div>
      </section>
    </div>
  );
}
