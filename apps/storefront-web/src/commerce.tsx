import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useRef,
  useState,
  type ReactNode,
} from "react";
import type {
  Cart,
  Config,
  CustomerInfo,
} from "../../../packages/commerce-client/types";
export const apiBase = import.meta.env.VITE_API_BASE_URL ?? "";
const errors: Record<string, string> = {
  unauthorized: "نشست شما منقضی شده؛ دوباره وارد شوید.",
  forbidden: "اجازه این عملیات را ندارید.",
  not_found: "اطلاعات پیدا نشد.",
  state_conflict: "اطلاعات تغییر کرده است؛ تازه‌سازی و دوباره تلاش کنید.",
  insufficient_stock: "موجودی کافی نیست؛ سبد را بررسی کنید.",
  checkout_changed:
    "قیمت یا شرایط سفارش تغییر کرده؛ مبلغ تازه را دوباره بررسی کنید.",
  insufficient_wallet: "موجودی قابل استفاده کیف پول کافی نیست.",
  payment_method_not_eligible: "این روش برای مبلغ یا دسته‌های سفارش مجاز نیست.",
  address_unavailable:
    "استان و شهر این آدرس فعال یا معتبر نیست؛ آدرس را ویرایش کنید.",
  coupon_not_eligible: "این کد تخفیف برای مبلغ، کالاها، تاریخ یا سهمیه این خرید معتبر نیست.",
  invalid_request: "اطلاعات واردشده معتبر نیست.",
  invalid_receipt: "تصویر رسید معتبر نیست؛ PNG، JPEG یا WebP انتخاب کنید.",
  receipt_too_large: "حجم رسید باید حداکثر ۲ مگابایت باشد.",
};
export class CommerceError extends Error {
  constructor(
    public status: number,
    public code: string,
  ) {
    super(
      errors[code] ??
        (status === 429
          ? "تعداد تلاش زیاد است؛ کمی بعد دوباره تلاش کنید."
          : "ارتباط برقرار نشد؛ دوباره تلاش کنید."),
    );
  }
}
export const operationKey = () => crypto.randomUUID();
export async function call<T>(
  path: string,
  token = "",
  body?: unknown,
  method?: string,
  key?: string,
  guest?: string,
): Promise<T> {
  const response = await fetch(apiBase + path, {
    method: method ?? (body === undefined ? "GET" : "POST"),
    cache: "no-store",
    headers: {
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...(guest ? { "X-Cart-Token": guest } : {}),
      ...(body instanceof FormData
        ? {}
        : { "Content-Type": "application/json" }),
      ...(body !== undefined
        ? { "Idempotency-Key": key ?? operationKey() }
        : {}),
    },
    body:
      body === undefined
        ? undefined
        : body instanceof FormData
          ? body
          : JSON.stringify(body),
  });
  if (!response.ok) {
    const value = await response.json().catch(() => ({}));
    if (response.status === 401 && token)
      dispatchEvent(new Event("cherriva:logout"));
    throw new CommerceError(response.status, value.code ?? "unknown");
  }
  if (response.status === 204) return undefined as T;
  return response.json();
}
interface Context {
  token: string;
  cart: Cart | null;
  config: Config | null;
  profile: CustomerInfo | null;
  cartError: string;
  refreshCart: () => Promise<Cart>;
  mutate: (input: {
    action: string;
    offer_id: string;
    quantity: number;
    version: number;
    new_offer_id?: string;
  }) => Promise<Cart>;
  login: (token: string) => Promise<void>;
  logout: () => void;
}
const CommerceContext = createContext<Context | null>(null);
export function useCommerce() {
  const value = useContext(CommerceContext);
  if (!value) throw Error("Commerce provider missing");
  return value;
}
const guestKey = "cherriva:guest-cart",
  tokenKey = "cherriva:customer-token";
export function CommerceProvider({ children }: { children: ReactNode }) {
  const [token, setToken] = useState(
    () => sessionStorage.getItem(tokenKey) ?? "",
  );
  const tokenRef = useRef(token);
  tokenRef.current = token;
  const [cart, setCart] = useState<Cart | null>(null),
    [config, setConfig] = useState<Config | null>(null),
    [profile, setProfile] = useState<CustomerInfo | null>(null),
    [cartError, setCartError] = useState("");
  const guestPromise = useRef<Promise<string> | null>(null);
  const guest = useCallback(async () => {
    const old = localStorage.getItem(guestKey);
    if (old) return old;
    if (!guestPromise.current)
      guestPromise.current = call<{ cart_token: string }>(
        "/storefront/cart/guest",
        "",
        {},
      )
        .then((v) => {
          localStorage.setItem(guestKey, v.cart_token);
          return v.cart_token;
        })
        .finally(() => {
          guestPromise.current = null;
        });
    return guestPromise.current;
  }, []);
  const refreshCart = useCallback(async () => {
    const current = tokenRef.current;
    try {
      const v = await call<Cart>(
        "/storefront/cart",
        current,
        undefined,
        undefined,
        undefined,
        current ? undefined : await guest(),
      );
      setCart(v);
      setCartError("");
      return v;
    } catch (e) {
      setCartError((e as Error).message);
      throw e;
    }
  }, [guest]);
  const logout = useCallback(() => {
    sessionStorage.removeItem(tokenKey);
    setToken("");
    tokenRef.current = "";
    setProfile(null);
    setCart(null);
  }, []);
  useEffect(() => {
    const fn = () => logout();
    addEventListener("cherriva:logout", fn);
    return () => removeEventListener("cherriva:logout", fn);
  }, [logout]);
  useEffect(() => {
    void refreshCart().catch(() => {});
    if (token)
      void call<CustomerInfo>("/storefront/account/me", token)
        .then(setProfile)
        .catch(() => {});
    void call<Config>("/storefront/commerce/config")
      .then(setConfig)
      .catch(() => {});
  }, [token, refreshCart]);
  const login = async (next: string) => {
    sessionStorage.setItem(tokenKey, next);
    tokenRef.current = next;
    const old = localStorage.getItem(guestKey);
    if (old) {
      await call("/storefront/cart/merge", next, { guest_token: old });
      localStorage.removeItem(guestKey);
    }
    await refreshCart();
    setToken(next);
    setProfile(await call<CustomerInfo>("/storefront/account/me", next));
  };
  const mutation = useRef<{ signature: string; promise: Promise<Cart> } | null>(
    null,
  );
  const mutationKeys = useRef(new Map<string, string>());
  const mutate = (input: Parameters<Context["mutate"]>[0]): Promise<Cart> => {
    const currentToken = tokenRef.current;
    const signature = JSON.stringify([currentToken, input]);
    if (mutation.current) {
      if (mutation.current.signature === signature)
        return mutation.current.promise;
      return Promise.reject(
        new Error("عملیات قبلی سبد هنوز پایان نیافته است."),
      );
    }
    let key = mutationKeys.current.get(signature);
    if (!key) {
      key = operationKey();
      mutationKeys.current.set(signature, key);
    }
    const promise = (async () => {
      try {
        const v = await call<Cart>(
          "/storefront/cart",
          currentToken,
          input,
          undefined,
          key,
          currentToken ? undefined : await guest(),
        );
        if (currentToken === tokenRef.current) setCart(v);
        mutationKeys.current.delete(signature);
        return v;
      } finally {
        mutation.current = null;
      }
    })();
    mutation.current = { signature, promise };
    return promise;
  };

  return (
    <CommerceContext.Provider
      value={{
        token,
        cart,
        config,
        profile,
        cartError,
        refreshCart,
        mutate,
        login,
        logout,
      }}
    >
      {children}
    </CommerceContext.Provider>
  );
}
