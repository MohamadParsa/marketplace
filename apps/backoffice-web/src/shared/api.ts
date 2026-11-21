import type {
  Command,
  Session,
  Workspace,
  ProductWorkspace,
  VariantResolution,
} from "./types";
const base = import.meta.env.VITE_API_BASE_URL ?? "";
const messages: Record<number, string> = {
  400: "اطلاعات واردشده معتبر نیست.",
  401: "نشست منقضی شده؛ دوباره وارد شوید.",
  403: "مجوز انجام این عملیات را ندارید.",
  404: "مورد درخواستی پیدا نشد.",
  409: "وضعیت رکورد تغییر کرده یا عملیات مجاز نیست؛ داده را تازه کنید.",
  413: "حجم تصویر از حد مجاز بیشتر است.",
  429: "تعداد تلاش‌ها بیش از حد مجاز است. کمی بعد دوباره تلاش کنید.",
};
export class ApiError extends Error {
  constructor(public status: number) {
    super(messages[status] ?? "ارتباط با سرور برقرار نشد. دوباره تلاش کنید.");
  }
}
async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
  const response = await fetch(base + path, {
    ...init,
    headers: {
      ...(init.body instanceof FormData
        ? {}
        : { "Content-Type": "application/json" }),
      ...init.headers,
    },
  });
  if (!response.ok) throw new ApiError(response.status);
  if (response.status === 204) return undefined as T;
  return response.json();
}
export const api = {
  productWorkspace: (token: string, id: string, surface = "backoffice") =>
    request<ProductWorkspace>(
      `/${surface}/products/${encodeURIComponent(id)}/workspace`,
      { headers: { Authorization: `Bearer ${token}` } },
    ),
  resolveVariant: (
    token: string,
    id: string,
    values: { attribute_id: string; value: string }[],
    surface = "backoffice",
  ) => {
    const params = new URLSearchParams(
      values.map((v) => [`attr.${v.attribute_id}`, v.value]),
    );
    return request<VariantResolution>(
      `/${surface}/products/${encodeURIComponent(id)}/resolve?${params}`,
      { headers: { Authorization: `Bearer ${token}` } },
    );
  },
  iconUrl: (id: string) => base + "/catalog/icons/" + encodeURIComponent(id),
  uploadIcon: (
    token: string,
    input: {
      kind: "category" | "brand";
      target: string;
      version: number;
      key: string;
      file: File;
    },
  ) => {
    const body = new FormData();
    body.set("file", input.file);
    body.set("version", String(input.version));
    body.set("idempotency_key", input.key);
    return request<void>(
      `/backoffice/catalog/${input.kind}/${encodeURIComponent(input.target)}/icon`,
      {
        method: "POST",
        body,
        headers: { Authorization: `Bearer ${token}` },
      },
    );
  },
  requestCode: (phone: string, surface = "backoffice") =>
    request<{ request_id: string; test_code?: string }>(
      `/${surface}/auth/request-otp`,
      { method: "POST", body: JSON.stringify({ phone }) },
    ),
  verify: (request_id: string, code: string, surface = "backoffice") =>
    request<Session>(`/${surface}/auth/verify-otp`, {
      method: "POST",
      body: JSON.stringify({ request_id, code }),
    }),
  workspace: (token: string, surface = "backoffice") =>
    request<Workspace>(`/${surface}/workspace`, {
      headers: { Authorization: `Bearer ${token}` },
    }),
  execute: (token: string, command: Command, surface = "backoffice") =>
    request<void>(`/${surface}/commands`, {
      method: "POST",
      headers: { Authorization: `Bearer ${token}` },
      body: JSON.stringify({
        ...command,
        idempotency_key: command.idempotency_key ?? crypto.randomUUID(),
      }),
    }),
  logout: (token: string, surface = "backoffice") =>
    request<void>(`/${surface}/auth/logout`, {
      method: "POST",
      headers: { Authorization: `Bearer ${token}` },
    }),
};
