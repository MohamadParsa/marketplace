import { FormEvent, StrictMode, useState } from "react";
import { createRoot } from "react-dom/client";
import "./styles.css";

function App() {
  const [phone, setPhone] = useState(""); const [requestId, setRequestId] = useState(""); const [code, setCode] = useState(""); const [message, setMessage] = useState("");
  const base = import.meta.env.VITE_API_BASE_URL ?? "http://localhost:8080";
  async function requestOtp(event: FormEvent) { event.preventDefault(); const response = await fetch(`${base}/seller/auth/request-otp`, { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify({ phone }) }); const data = await response.json(); if (!response.ok) return setMessage("شماره ثبت نشده یا حساب غیرفعال است."); setRequestId(data.request_id); setCode(data.test_code ?? ""); setMessage("کد ورود ارسال شد."); }
  async function verifyOtp(event: FormEvent) { event.preventDefault(); const response = await fetch(`${base}/seller/auth/verify-otp`, { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify({ request_id: requestId, code }) }); setMessage(response.ok ? "ورود با موفقیت انجام شد." : "کد ورود معتبر نیست."); }
  return <main className="shell"><p className="eyebrow">marketplace / seller portal</p><h1>ورود فروشنده</h1><form onSubmit={requestOtp}><label>شماره تلفن<input value={phone} onChange={(event) => setPhone(event.target.value)} placeholder="09120000002" /></label><button type="submit">دریافت کد</button></form>{requestId && <form onSubmit={verifyOtp}><label>کد یک‌بارمصرف<input value={code} onChange={(event) => setCode(event.target.value)} inputMode="numeric" /></label><button type="submit">ورود</button></form>}<p role="status">{message}</p></main>;
}

createRoot(document.getElementById("root")!).render(<StrictMode><App /></StrictMode>);
