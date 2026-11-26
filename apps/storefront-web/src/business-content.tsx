import "./business-content.css";
import { useEffect, useState } from "react";
import type { Business } from "../../../packages/commerce-client/types";
const base = import.meta.env.VITE_API_BASE_URL ?? "";
export const businessPages = [
  ["about", "about", "آشنایی با چریوا"],
  ["contact", "contact", "تماس با ما"],
  ["shipping", "shipping_policy", "شرایط ارسال"],
  ["returns", "return_policy", "شرایط بازگشت"],
  ["terms", "terms", "قوانین خرید"],
  ["privacy", "privacy", "حریم خصوصی"],
] as const;
let pending: Promise<Business> | undefined;
function read() {
  if (!pending) {
    const controller = new AbortController();
    const deadline = setTimeout(() => controller.abort(), 10000);
    pending = fetch(base + "/storefront/business", {
      cache: "no-store",
      signal: controller.signal,
    })
      .then((r) => {
        if (!r.ok) throw Error("اطلاعات عمومی دریافت نشد.");
        return r.json() as Promise<Business>;
      })
      .finally(() => {
        clearTimeout(deadline);
        pending = undefined;
      });
  }
  return pending;
}
export function usePublicBusiness() {
  const [data, setData] = useState<Business>(),
    [loading, setLoading] = useState(true),
    [error, setError] = useState(""),
    [attempt, setAttempt] = useState(0);
  useEffect(() => {
    let alive = true;
    const load = () => {
      setLoading(true);
      read()
        .then((b) => {
          if (alive) {
            setData(b);
            setError("");
          }
        })
        .catch(() => {
          if (alive) {
            setData(undefined);
            setError("اطلاعات عمومی دریافت نشد.");
          }
        })
        .finally(() => {
          if (alive) setLoading(false);
        });
    };
    load();
    addEventListener("focus", load);
    addEventListener("popstate", load);
    return () => {
      alive = false;
      removeEventListener("focus", load);
      removeEventListener("popstate", load);
    };
  }, [attempt]);
  return { data, loading, error, refresh: () => setAttempt((v) => v + 1) };
}
export function SocialIcon({ kind }: { kind: string }) {
  return (
    <svg
      width="22"
      height="22"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.6"
      aria-hidden="true"
    >
      {kind === "instagram" ? (
        <>
          <rect x="3" y="3" width="18" height="18" rx="5" />
          <circle cx="12" cy="12" r="4" />
          <circle cx="17" cy="7" r=".8" fill="currentColor" />
        </>
      ) : kind === "telegram" ? (
        <path d="m3 10 18-7-4 18-6-6-4 3 1-6 9-6-7 8-7-4Z" />
      ) : kind === "x" ? (
        <>
          <path d="m4 3 16 18H16L4 3Zm16 0L4 21" />
        </>
      ) : kind === "youtube" || kind === "aparat" ? (
        <>
          <rect x="2" y="5" width="20" height="14" rx="4" />
          <path d="m10 9 5 3-5 3Z" />
        </>
      ) : kind === "linkedin" ? (
        <>
          <rect x="3" y="3" width="18" height="18" rx="3" />
          <path d="M7 10v7m0-10v.2M11 17v-7m0 3c0-4 6-4 6 0v4" />
        </>
      ) : kind === "whatsapp" ? (
        <>
          <path d="M5 19 2 22l1-7a9 9 0 1 1 4 5Z" />
          <path d="M8 7c-2 4 3 9 7 8l2-2-3-2-1 1-2-2 1-1-2-3Z" />
        </>
      ) : (
        <>
          <path d="m10 7 2-2a5 5 0 0 1 7 7l-2 2M14 17l-2 2a5 5 0 0 1-7-7l2-2m2 5 6-6" />
        </>
      )}
    </svg>
  );
}
