import type {
  Command,
  Session,
  Workspace,
  ProductWorkspace,
  VariantResolution,
  BannerList,
  ProductQAPage,
  ProductQAMutation,
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
  constructor(
    public status: number,
    public code?: string,
  ) {
    super(
      code === "invalid_reauthentication"
        ? "کد تأیید معتبر نیست یا منقضی شده است."
        : (messages[status] ?? "ارتباط با سرور برقرار نشد. دوباره تلاش کنید."),
    );
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
  if (!response.ok) {
    const body = await response.json().catch(() => null);
    throw new ApiError(response.status, body?.code);
  }
  if (response.status === 204) return undefined as T;
  return response.json();
}
export const api = {
  profile: (token: string, surface="backoffice") =>
    request<{ name: string; email?: string; password_available: boolean }>(
      `/${surface}/account/profile`,
      { headers: { Authorization: `Bearer ${token}` } },
    ),
  updateProfile: (
    token: string,
    input: { name: string; email: string; password: string },
    surface="backoffice",
  ) =>
    request<void>(`/${surface}/account/profile`, {
      method: "PATCH",
      headers: { Authorization: `Bearer ${token}` },
      body: JSON.stringify(input),
    }),
  catalogMedia: async (token: string, id: string) => {
    const response = await fetch(
      base + `/backoffice/catalog/media/${encodeURIComponent(id)}`,
      { headers: { Authorization: `Bearer ${token}` } },
    );
    if (!response.ok) throw new ApiError(response.status);
    return response.blob();
  },
  banners: (token: string) =>
    request<BannerList>("/backoffice/banners", {
      headers: { Authorization: `Bearer ${token}` },
    }),
  bannerViewUrl: (id: string) =>
    base + `/content/banners/${encodeURIComponent(id)}/view`,
  bannerAsset: async (token: string, id: string) => {
    const response = await fetch(
      base + `/backoffice/banner-media/${encodeURIComponent(id)}`,
      { headers: { Authorization: `Bearer ${token}` } },
    );
    if (!response.ok) throw new ApiError(response.status);
    return response.blob();
  },
  uploadBanner: (
    token: string,
    input: {
      id: string;
      slot: "desktop" | "mobile";
      version: number;
      key: string;
      file: File;
    },
  ) => {
    const body = new FormData();
    body.set("file", input.file);
    body.set("version", String(input.version));
    body.set("idempotency_key", input.key);
    return request<{ asset_id: string; version: number }>(
      `/backoffice/banners/${encodeURIComponent(input.id)}/media/${input.slot}`,
      { method: "POST", body, headers: { Authorization: `Bearer ${token}` } },
    );
  },
  productQA: (token: string, id: string, after: number) =>
    request<ProductQAPage>(
      `/backoffice/products/${encodeURIComponent(id)}/questions${after ? `?after=${after}` : ""}`,
      { cache: "no-store", headers: { Authorization: `Bearer ${token}` } },
    ),
  changeProductQA: (token: string, id: string, question: string, mutation: ProductQAMutation, key: string) =>
    request<void>(
      `/backoffice/products/${encodeURIComponent(id)}/questions/${encodeURIComponent(question)}`,
      { method: "POST", body: JSON.stringify(mutation), headers: { Authorization: `Bearer ${token}`, "Idempotency-Key": key } },
    ),
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
      kind: "category" | "brand" | "product" | "product-gallery";
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
  logout: (token: string, surface = "backoffice") =>
    request<void>(`/${surface}/auth/logout`, {
      method: "POST",
      headers: { Authorization: `Bearer ${token}` },
    }),
  requestCode: (phone: string, surface = "backoffice") =>
    request<{
      request_id: string;
      test_code?: string;
      password_available?: boolean;
    }>(`/${surface}/auth/request-otp`, {
      method: "POST",
      body: JSON.stringify({ phone }),
    }),
  passwordLogin: (phone: string, password: string, surface="backoffice") =>
    request<Session>(`/${surface}/auth/password-login`, {
      method: "POST",
      body: JSON.stringify({ phone, password }),
    }),
  verify: (request_id: string, code: string, surface = "backoffice") =>
    request<Session>(`/${surface}/auth/verify-otp`, {
      method: "POST",
      body: JSON.stringify({ request_id, code }),
    }),
  requestAdminCode: (token: string) =>
    request<{ request_id: string; test_code?: string }>(
      "/backoffice/auth/request-admin-otp",
      { method: "POST", headers: { Authorization: `Bearer ${token}` } },
    ),
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
};
