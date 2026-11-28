import type { Session } from "./types";

const key = "cherriva.staff-session.v1";
export type StoredStaffSession = { token: string; surface: "backoffice" | "seller"; expires_at: string };
// Tab-scoped bearer storage keeps reloads usable without extending the server session lifetime.
export function readStaffSession(): StoredStaffSession | null {
  try {
    const saved = JSON.parse(sessionStorage.getItem(key) ?? "null") as StoredStaffSession | null;
    if (saved && typeof saved.token === "string" && saved.token.length > 0 &&
      (saved.surface === "backoffice" || saved.surface === "seller") && Date.parse(saved.expires_at) > Date.now()) return saved;
    sessionStorage.removeItem(key);
  } catch { /* Storage may be unavailable; normal login remains usable. */ }
  return null;
}
export function saveStaffSession(session: Session | null) {
  try {
    if (session) sessionStorage.setItem(key, JSON.stringify({ token: session.token, surface: session.user.Surface, expires_at: session.expires_at }));
    else sessionStorage.removeItem(key);
  } catch { /* Keep the active in-memory session usable when storage is disabled. */ }
}
