import { FormEvent, StrictMode, useEffect, useState } from "react";
import { createRoot } from "react-dom/client";
import "./styles.css";

const base = import.meta.env.VITE_API_BASE_URL ?? "";
type Session = {
  token: string;
  user: { StoreIDs: string[]; StorePermissions?: Record<string, string[] | null> };
};
type Offer = { id: string; product_id: string; stock: number; version: number };

class RequestError extends Error {
  constructor(readonly status: number) {
    const messages: Record<number, string> = {
      400: "اطلاعات واردشده معتبر نیست.",
      401: "نشست یا کد ورود معتبر نیست. دوباره تلاش کنید.",
      403: "برای این عملیات در این فروشگاه مجوز ندارید.",
      409: "موجودی تغییر کرده است. فهرست را تازه و دوباره بررسی کنید.",
      429: "برای درخواست کد دوباره کمی صبر کنید.",
    };
    super(messages[status] ?? "ارتباط با سرور برقرار نشد.");
  }
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(base + path, init);
  if (!response.ok) throw new RequestError(response.status);
  return response.status === 204 ? undefined as T : response.json();
}

function post(body: object): RequestInit {
  return { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify(body) };
}

function message(error: unknown) {
  return error instanceof Error ? error.message : "عملیات ناموفق بود.";
}

function Login({ onLogin }: { onLogin: (session: Session) => void }) {
  const [phone, setPhone] = useState("");
  const [requestID, setRequestID] = useState("");
  const [code, setCode] = useState("");
  const [status, setStatus] = useState("");
  const [busy, setBusy] = useState(false);
  async function submit(event: FormEvent) {
    event.preventDefault();
    setStatus("");
    setBusy(true);
    try {
      if (requestID) {
        onLogin(await request<Session>("/seller/auth/verify-otp", post({ request_id: requestID, code })));
      } else {
        const result = await request<{ request_id: string; test_code?: string }>("/seller/auth/request-otp", post({ phone }));
        setRequestID(result.request_id);
        setCode(result.test_code ?? "");
        setStatus("کد ورود ارسال شد.");
      }
    } catch (error) { setStatus(message(error)); }
    finally { setBusy(false); }
  }
  return <>
    <h1>ورود فروشنده</h1>
    <form onSubmit={submit}>
      {requestID
        ? <label>کد یک‌بارمصرف<input value={code} onChange={(event) => setCode(event.target.value)} inputMode="numeric" autoComplete="one-time-code" required /></label>
        : <label>شماره تلفن<input value={phone} onChange={(event) => setPhone(event.target.value)} placeholder="09120000002" type="tel" required /></label>}
      <button disabled={busy} type="submit">{requestID ? "ورود" : "دریافت کد"}</button>
      {requestID && <button type="button" disabled={busy} onClick={() => { setRequestID(""); setCode(""); }}>تغییر شماره</button>}
    </form>
    <p role="status">{status}</p>
  </>;
}

function StockRow({ offer, onSave }: { offer: Offer; onSave?: (stock: number) => Promise<void> }) {
  const [stock, setStock] = useState(String(offer.stock));
  const [status, setStatus] = useState("");
  const [busy, setBusy] = useState(false);
  if (!onSave) return <li>{offer.product_id} · موجودی: {offer.stock}</li>;
  const save = onSave;
  return <li>
    <form onSubmit={async (event) => {
      event.preventDefault();
      setBusy(true);
      setStatus("");
      try { await save(Number(stock)); setStatus("موجودی ذخیره شد."); }
      catch (error) { setStatus(message(error)); }
      finally { setBusy(false); }
    }}>
      <label>موجودی {offer.product_id}
        <input type="number" min="0" step="1" required value={stock} onChange={(event) => setStock(event.target.value)} />
      </label>
      <button type="submit" disabled={busy || stock === String(offer.stock)}>ذخیره موجودی</button>
      {status && <span role="status">{status}</span>}
    </form>
  </li>;
}

function Workspace({ session, onLogout }: { session: Session; onLogout: () => void }) {
  const [storeID, setStoreID] = useState(session.user.StoreIDs[0] ?? "");
  const [offers, setOffers] = useState<Offer[] | null>(null);
  const [status, setStatus] = useState("");
  const [revision, setRevision] = useState(0);
  const headers = { Authorization: `Bearer ${session.token}` };
  const path = `/seller/stores/${encodeURIComponent(storeID)}/offers`;
  const canWrite = session.user.StorePermissions?.[storeID]?.includes("stock.write");

  useEffect(() => {
    if (!storeID) return;
    const controller = new AbortController();
    setOffers(null);
    setStatus("");
    request<Offer[]>(path, { headers: { Authorization: `Bearer ${session.token}` }, signal: controller.signal })
      .then((result) => { if (!controller.signal.aborted) setOffers(result); })
      .catch((error) => { if (!controller.signal.aborted) setStatus(message(error)); });
    return () => controller.abort();
  }, [path, revision, session.token, storeID]);

  return <>
    <h1>عرضه‌های فروشگاه</h1>
    <button type="button" onClick={async () => {
      try { await request("/seller/auth/logout", { method: "POST", headers }); }
      catch { /* The local session must still close. */ }
      finally { onLogout(); }
    }}>خروج</button>
    <label>فروشگاه
      <select value={storeID} onChange={(event) => { setOffers(null); setStoreID(event.target.value); }}>
        {session.user.StoreIDs.map((id) => <option value={id} key={id}>{id}</option>)}
      </select>
    </label>
    <button type="button" onClick={() => setRevision((value) => value + 1)}>تازه‌سازی عرضه‌ها</button>
    {status ? <p role="alert">{status}</p> : offers === null
      ? <p role="status">در حال دریافت عرضه‌ها…</p>
      : offers.length === 0 ? <p>عرضه‌ای وجود ندارد.</p>
      : <ul>{offers.map((offer) => <StockRow
          key={`${storeID}:${offer.id}:${offer.version}`} offer={offer}
          onSave={canWrite ? async (stock) => {
            await request(`${path}/${encodeURIComponent(offer.id)}/stock`, {
              method: "PUT", headers: { ...headers, "Content-Type": "application/json" },
              body: JSON.stringify({ stock, version: offer.version }),
            });
            setRevision((value) => value + 1);
          } : undefined}
        />)}</ul>}
  </>;
}

function App() {
  const [session, setSession] = useState<Session | null>(null);
  return <main className="shell" dir="rtl">
    {session ? <Workspace session={session} onLogout={() => setSession(null)} /> : <Login onLogin={setSession} />}
  </main>;
}

createRoot(document.getElementById("root")!).render(<StrictMode><App /></StrictMode>);
