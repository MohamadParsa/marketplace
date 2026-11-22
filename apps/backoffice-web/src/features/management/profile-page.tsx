import { useEffect, useState } from "react";
import { Alert, Button, TextField } from "@mui/material";
import { useWorkspace } from "../../shared/workspace";
export function ProfilePage() {
  const { loadProfile, updateProfile } = useWorkspace();
  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [hasPassword, setHasPassword] = useState(false);
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [notice, setNotice] = useState("");
  useEffect(() => {
    let active = true;
    void loadProfile()
      .then((p) => {
        if (active) {
          setName(p.name);
          setEmail(p.email ?? "");
          setHasPassword(p.password_available);
        }
      })
      .catch(() => {
        if (active) setError("دریافت اطلاعات حساب انجام نشد.");
      })
      .finally(() => {
        if (active) setLoading(false);
      });
    return () => {
      active = false;
    };
  }, [loadProfile]);
  return (
    <>
      <div className="page-heading">
        <div>
          <h1>پروفایل و تنظیم رمز</h1>
          <p>یک حساب برای خرید، مدیریت و عضویت فروشگاه</p>
        </div>
      </div>
      <section className="panel home-admin-panel" style={{ maxWidth: 650 }}>
        {error && <Alert severity="error">{error}</Alert>}
        {notice && <Alert severity="success">{notice}</Alert>}
        {loading ? (
          <p role="status">در حال دریافت حساب…</p>
        ) : (
          <form
            className="home-admin-form"
            onSubmit={(e) => {
              e.preventDefault();
              if (busy) return;
              setBusy(true);
              setError("");
              setNotice("");
              void updateProfile({ name, email, password })
                .then(() => {
                  setHasPassword(hasPassword || !!password);
                  setPassword("");
                  setNotice("اطلاعات حساب ذخیره شد.");
                })
                .catch((e) =>
                  setError(e instanceof Error ? e.message : "ذخیره انجام نشد."),
                )
                .finally(() => setBusy(false));
            }}
          >
            <TextField
              label="نام"
              value={name}
              onChange={(e) => setName(e.target.value)}
              required
              inputProps={{ maxLength: 160 }}
            />
            <TextField
              label="ایمیل اختیاری"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              type="email"
              inputProps={{ dir: "ltr", maxLength: 254 }}
              autoComplete="email"
              helperText="ورود با ایمیل فقط با رمز عبور انجام می‌شود."
            />
            <TextField
              label={hasPassword ? "رمز عبور جدید" : "تعریف رمز عبور"}
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              type="password"
              autoComplete="new-password"
              inputProps={{ minLength: 8, maxLength: 72, dir: "ltr" }}
              helperText="خالی بگذارید تا رمز فعلی حفظ شود. شماره موبایل با رمز یا OTP قابل استفاده است."
            />
            <Button type="submit" variant="contained" disabled={busy}>
              ذخیره حساب
            </Button>
          </form>
        )}
      </section>
    </>
  );
}
