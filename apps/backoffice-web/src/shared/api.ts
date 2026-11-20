import type { Command, Session, Workspace } from "./types";
const base = import.meta.env.VITE_API_BASE_URL ?? "";
const messages: Record<number, string> = {
  400: "اطلاعات واردشده معتبر نیست.",
  401: "نشست منقضی شده؛ دوباره وارد شوید.",
  403: "مجوز انجام این عملیات را ندارید.",
  409: "وضعیت رکورد تغییر کرده یا عملیات مجاز نیست؛ داده را تازه کنید.",
  429: "تعداد تلاش‌ها بیش از حد مجاز است. کمی بعد دوباره تلاش کنید.",
};
export class ApiError extends Error {
  constructor(public status: number, public code?: string) {
    super(
      code === "invalid_reauthentication"
        ? "کد تأیید معتبر نیست یا منقضی شده است."
        : messages[status] ?? "ارتباط با سرور برقرار نشد. دوباره تلاش کنید.",
    );
  }
}
async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
  const response = await fetch(base + path, {
    ...init,
    headers: { "Content-Type": "application/json", ...init.headers },
  });
  if (!response.ok) {
    const body = await response.json().catch(() => null);
    throw new ApiError(response.status, body?.code);
  }
  if (response.status === 204) return undefined as T;
  return response.json();
}
export const api = {
  requestCode: (phone: string) =>
    request<{ request_id: string; test_code?: string }>(
      "/backoffice/auth/request-otp",
      { method: "POST", body: JSON.stringify({ phone }) },
    ),
  verify: (request_id: string, code: string) =>
    request<Session>("/backoffice/auth/verify-otp", {
      method: "POST",
      body: JSON.stringify({ request_id, code }),
    }),
  requestAdminCode: (token: string) =>
    request<{ request_id: string; test_code?: string }>(
      "/backoffice/auth/request-admin-otp",
      { method: "POST", headers: { Authorization: `Bearer ${token}` } },
    ),
  workspace: (token: string) =>
    request<Workspace>("/backoffice/workspace", {
      headers: { Authorization: `Bearer ${token}` },
    }),
  execute: (token: string, command: Command) =>
    request<void>("/backoffice/commands", {
      method: "POST",
      headers: { Authorization: `Bearer ${token}` },
      body: JSON.stringify({
        ...command,
        idempotency_key: command.idempotency_key ?? crypto.randomUUID(),
      }),
    }),
  logout: (token: string) =>
    request<void>("/backoffice/auth/logout", {
      method: "POST",
      headers: { Authorization: `Bearer ${token}` },
    }),
};
