import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useRef,
  useState,
  type ReactNode,
} from "react";
import { call, CommerceError, useCommerce } from "./commerce";
import { BusyRegion } from "./busy-region";
import { navigate } from "./navigation";
import "./wishlist.css";

export function HeartIcon({ filled = false }: { filled?: boolean }) {
  return (
    <svg
      className="wishlist-heart-icon"
      viewBox="0 0 24 24"
      fill={filled ? "currentColor" : "none"}
      stroke="currentColor"
      strokeWidth="1.6"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
    >
      <path d="M20.8 4.6a5.5 5.5 0 0 0-7.8 0L12 5.7l-1.1-1.1a5.5 5.5 0 0 0-7.8 7.8L12 21l8.8-8.6a5.5 5.5 0 0 0 0-7.8Z" />
    </svg>
  );
}
const endpoint = "/storefront/account/wishlist";
type Membership = { product_id: string; saved: boolean; total: number };
type State = {
  session: string;
  total: number | null;
  saved: Record<string, boolean>;
  revision: number;
};
type Context = {
  total: number | null;
  saved: Record<string, boolean>;
  revision: number;
  ensure: (id: string) => Promise<void>;
  setSaved: (id: string, saved: boolean) => Promise<void>;
};
const WishlistContext = createContext<Context | null>(null);
export function useWishlist() {
  const value = useContext(WishlistContext);
  if (!value) throw Error("Wishlist provider missing");
  return value;
}
export function WishlistProvider({ children }: { children: ReactNode }) {
  const { token } = useCommerce();
  return (
    <WishlistSession key={token} token={token}>
      {children}
    </WishlistSession>
  );
}
function WishlistSession({
  children,
  token,
}: {
  children: ReactNode;
  token: string;
}) {
  const identity = useRef(token);
  identity.current = token;
  const epoch = useRef(0);
  const requests = useRef(new Map<string, Promise<void>>());
  const writes = useRef(new Map<string, Promise<void>>());
  const [state, setState] = useState<State>({
    session: token,
    total: null,
    saved: {},
    revision: 0,
  });
  useEffect(() => {
    const ticket = epoch.current;
    let cancelled = false;
    if (token)
      void call<{ total: number }>(endpoint + "?page_size=1", token)
        .then((result) => {
          if (
            !cancelled &&
            identity.current === token &&
            epoch.current === ticket
          )
            setState((old) =>
              old.session === token ? { ...old, total: result.total } : old,
            );
        })
        .catch(() => {});
    return () => {
      cancelled = true;
    };
  }, [token]);
  const ensure = useCallback(
    (id: string): Promise<void> => {
      if (!token) return Promise.resolve();
      const key = token + ":" + id;
      const pending = requests.current.get(key);
      if (pending) return pending;
      const ticket = epoch.current;
      const promise = call<Membership>(
        endpoint + "/" + encodeURIComponent(id),
        token,
      )
        .then((result) => {
          if (identity.current === token && epoch.current === ticket)
            setState((old) =>
              old.session === token
                ? {
                    ...old,
                    total: result.total,
                    saved: { ...old.saved, [id]: result.saved },
                  }
                : old,
            );
        })
        .finally(() => {
          if (requests.current.get(key) === promise)
            requests.current.delete(key);
        });
      requests.current.set(key, promise);
      return promise;
    },
    [token],
  );
  const setSaved = useCallback(
    (id: string, saved: boolean): Promise<void> => {
      if (!token) return Promise.reject(new CommerceError(401, "unauthorized"));
      const key = token + ":" + id;
      if (writes.current.has(key))
        return Promise.reject(new CommerceError(409, "state_conflict"));
      ++epoch.current;
      const promise = call<Membership>(
        endpoint + "/" + encodeURIComponent(id),
        token,
        undefined,
        saved ? "PUT" : "DELETE",
      )
        .then((result) => {
          if (identity.current === token)
            setState((old) =>
              old.session === token
                ? {
                    ...old,
                    total: result.total,
                    saved: { ...old.saved, [id]: result.saved },
                    revision: old.revision + 1,
                  }
                : old,
            );
        })
        .finally(() => {
          if (writes.current.get(key) === promise) writes.current.delete(key);
        });
      writes.current.set(key, promise);
      return promise;
    },
    [token],
  );
  const current =
    state.session === token ? state : { total: null, saved: {}, revision: 0 };
  return (
    <WishlistContext.Provider value={{ ...current, ensure, setSaved }}>
      {children}
    </WishlistContext.Provider>
  );
}
export function WishlistButton({ productID }: { productID: string }) {
  const { token } = useCommerce();
  const { saved, ensure, setSaved } = useWishlist();
  const [busy, setBusy] = useState(false),
    [ready, setReady] = useState(!token),
    [error, setError] = useState("");
  const running = useRef(false);
  useEffect(() => {
    let cancelled = false;
    setError("");
    setReady(!token);
    if (token)
      void ensure(productID)
        .then(() => {
          if (!cancelled) setReady(true);
        })
        .catch((e) => {
          if (!cancelled) {
            setError((e as Error).message);
            setReady(false);
          }
        });
    return () => {
      cancelled = true;
    };
  }, [token, productID, ensure]);
  const selected = !!saved[productID];
  return (
    <div className="pdp-wishlist">
      <BusyRegion busy={busy} label="به‌روزرسانی علاقه‌مندی‌ها">
        <button
          className={`wishlist-toggle ${selected ? "is-saved" : ""}`}
          type="button"
          aria-label={
            selected ? "حذف از علاقه‌مندی‌ها" : "افزودن به علاقه‌مندی‌ها"
          }
          title={
            !ready && error
              ? "تلاش مجدد برای دریافت علاقه‌مندی"
              : selected
                ? "حذف از علاقه‌مندی‌ها"
                : "افزودن به علاقه‌مندی‌ها"
          }
          aria-pressed={selected}
          aria-busy={busy}
          disabled={busy || (!!token && !ready && !error)}
          onClick={() => {
            if (!token) {
              navigate(
                "/login?return=" +
                  encodeURIComponent(location.pathname + location.search),
              );
              return;
            }
            if (running.current) return;
            running.current = true;
            setBusy(true);
            setError("");
            const work = ready
              ? setSaved(productID, !selected)
              : ensure(productID).then(() => setReady(true));
            void work
              .catch((e) => setError((e as Error).message))
              .finally(() => {
                running.current = false;
                setBusy(false);
              });
          }}
        >
          <HeartIcon filled={selected} />
        </button>
        {error && <p role="alert">{error}</p>}
      </BusyRegion>
    </div>
  );
}
