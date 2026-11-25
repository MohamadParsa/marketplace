import { useEffect, useRef, useState } from "react";
import { apiBase, call, useCommerce } from "./commerce";
type Point = { lat: number; lng: number };
type MapConfig = { enabled: boolean; web_key: string };
type Suggestion = {
  full_address: string;
  province: string;
  city: string;
  geography_matched: boolean;
};
type MapView = {
  on: (event: string, fn: (e: { latlng: Point }) => void) => void;
  remove: () => void;
  stop: () => void;
  setView: (
    point: Point | number[],
    zoom?: number,
    options?: { animate: boolean },
  ) => void;
};
type Marker = {
  addTo: (map: MapView) => Marker;
  setLatLng: (point: Point) => void;
  getLatLng: () => Point;
  on: (event: string, fn: () => void) => void;
  remove: () => void;
};
type SDK = {
  Map: new (element: HTMLElement, options: object) => MapView;
  marker: (point: Point, options: object) => Marker;
  divIcon: (options: object) => object;
};
const sdkURL =
  "https://static.neshan.org/sdk/leaflet/v1.9.4/neshan-sdk/v1.0.8/index.js";
let sdkPromise: Promise<SDK> | null = null;
function loadSDK() {
  if (!sdkPromise)
    sdkPromise = new Promise<SDK>((resolve, reject) => {
      const css = document.createElement("link");
      css.rel = "stylesheet";
      css.href = sdkURL.replace("index.js", "index.css");
      document.head.append(css);
      const script = document.createElement("script");
      script.src = sdkURL;
      script.async = true;
      const fail = () => {
        script.remove();
        css.remove();
        reject(Error("map-sdk-failed"));
      };
      const timer = setTimeout(fail, 8000);
      script.onload = () => {
        clearTimeout(timer);
        const sdk = (window as unknown as { L?: SDK }).L;
        if (sdk) resolve(sdk);
        else fail();
      };
      script.onerror = () => {
        clearTimeout(timer);
        fail();
      };
      document.head.append(script);
    }).catch((e) => {
      sdkPromise = null;
      throw e;
    });
  return sdkPromise;
}
export function AddressMap({
  latitude,
  longitude,
  disabled,
  onPick,
  onSuggestion,
}: {
  latitude?: number;
  longitude?: number;
  disabled: boolean;
  onPick: (point: Point | null) => void;
  onSuggestion: (s: Suggestion) => void;
}) {
  const { token } = useCommerce();
  const [open, setOpen] = useState(false),
    [config, setConfig] = useState<MapConfig | null>(null),
    [error, setError] = useState(""),
    [busy, setBusy] = useState(false),
    [suggestion, setSuggestion] = useState<Suggestion | null>(null);
  const [lat, setLat] = useState(latitude?.toString() ?? ""),
    [lng, setLng] = useState(longitude?.toString() ?? "");
  const element = useRef<HTMLDivElement>(null),
    map = useRef<MapView | null>(null),
    marker = useRef<Marker | null>(null),
    sdk = useRef<SDK | null>(null),
    sequence = useRef(0),
    abort = useRef<AbortController | null>(null);
  const callbacks = useRef({ onPick, onSuggestion, disabled });
  callbacks.current = { onPick, onSuggestion, disabled };
  const reverse = async (point: Point) => {
    abort.current?.abort();
    const controller = new AbortController();
    abort.current = controller;
    const current = ++sequence.current;
    setSuggestion(null);
    setError("");
    if (!config?.enabled) return;
    setBusy(true);
    const timer = setTimeout(() => controller.abort(), 6500);
    try {
      const response = await fetch(
        apiBase + "/storefront/account/address-map/reverse",
        {
          method: "POST",
          cache: "no-store",
          signal: controller.signal,
          headers: {
            Authorization: `Bearer ${token}`,
            "Content-Type": "application/json",
          },
          body: JSON.stringify({ latitude: point.lat, longitude: point.lng }),
        },
      );
      if (!response.ok)
        throw Error(
          response.status === 429
            ? "سقف درخواست نقشه پر شده؛ نشانی را دستی وارد کنید."
            : "دریافت نشانی انجام نشد؛ ورود دستی همچنان ممکن است.",
        );
      const result: Suggestion = await response.json();
      if (current === sequence.current) setSuggestion(result);
    } catch (e) {
      if (current === sequence.current)
        setError(
          controller.signal.aborted
            ? "زمان دریافت نشانی تمام شد؛ ورود دستی همچنان ممکن است."
            : (e as Error).message,
        );
    } finally {
      clearTimeout(timer);
      if (current === sequence.current) setBusy(false);
    }
  };
  const pickRef = useRef<(point: Point) => void>(() => {});
  const icon = (L: SDK) =>
    L.divIcon({
      className: "address-map-pin",
      html: '<span aria-hidden="true">●</span>',
      iconSize: [28, 28],
    });
  pickRef.current = (point) => {
    if (callbacks.current.disabled) return;
    const chosen = {
      lat: Number(point.lat.toFixed(6)),
      lng: Number(point.lng.toFixed(6)),
    };
    setLat(chosen.lat.toString());
    setLng(chosen.lng.toString());
    callbacks.current.onPick(chosen);
    if (map.current && sdk.current) {
      if (!marker.current) {
        marker.current = sdk.current
          .marker(chosen, { draggable: true, icon: icon(sdk.current) })
          .addTo(map.current);
        marker.current.on("dragend", () => {
          if (marker.current) pickRef.current(marker.current.getLatLng());
        });
      } else marker.current.setLatLng(chosen);
    }
    void reverse(chosen);
  };
  useEffect(() => {
    if (!open) return;
    let disposed = false;
    setError("");
    void call<MapConfig>("/storefront/account/address-map", token)
      .then((c) => {
        if (!disposed) setConfig(c);
      })
      .catch(() => {
        if (!disposed)
          setError("پیکربندی نقشه دریافت نشد؛ نشانی را دستی وارد کنید.");
      });
    return () => {
      disposed = true;
      sequence.current++;
      abort.current?.abort();
      setBusy(false);
    };
  }, [open, token]);
  useEffect(() => {
    if (!open || !config?.enabled || !element.current) return;
    let disposed = false;
    void loadSDK()
      .then((L) => {
        if (disposed || !element.current) return;
        sdk.current = L;
        const point = {
          lat: latitude ?? 35.699756,
          lng: longitude ?? 51.338076,
        };
        const view = new L.Map(element.current, {
          key: config.web_key,
          maptype: "dreamy",
          center: [point.lat, point.lng],
          zoom: 14,
          zoomAnimation: false,
          fadeAnimation: false,
          markerZoomAnimation: false,
        });
        map.current = view;
        view.on("click", (e) => pickRef.current(e.latlng));
        // An initial map centre is not a selected delivery location.
        if (latitude !== undefined && longitude !== undefined) {
          marker.current = L.marker(point, {
            draggable: true,
            icon: icon(L),
          }).addTo(view);
          marker.current.on("dragend", () => {
            if (marker.current) pickRef.current(marker.current.getLatLng());
          });
        }
      })
      .catch(() => {
        if (!disposed)
          setError(
            "نقشه بارگذاری نشد؛ می‌توانید نشانی یا مختصات را دستی وارد کنید.",
          );
      });
    return () => {
      disposed = true;
      map.current?.stop();
      map.current?.remove();
      map.current = null;
      marker.current = null;
    };
  }, [open, config?.enabled, config?.web_key]);
  return (
    <section className="address-map-section">
      <button
        type="button"
        className="buy-link"
        aria-expanded={open}
        onClick={() => setOpen((v) => !v)}
      >
        انتخاب موقعیت روی نقشه {open ? "−" : "+"}
      </button>
      {open && (
        <>
          <p className="buy-note">
            موقعیت فقط برای همین آدرس است. ورود دستی همیشه امکان‌پذیر است.
          </p>
          {config?.enabled ? (
            <div
              className="address-map-canvas"
              ref={element}
              aria-label="نقشه انتخاب مقصد"
            />
          ) : (
            <p className="buy-note">
              نقشه هنوز فعال نشده است؛ نشانی را دستی تکمیل کنید.
            </p>
          )}
          <div className="address-map-coordinates">
            <label className="buy-field">
              عرض جغرافیایی
              <input
                value={lat}
                inputMode="decimal"
                dir="ltr"
                onChange={(e) => setLat(e.target.value)}
              />
            </label>
            <label className="buy-field">
              طول جغرافیایی
              <input
                value={lng}
                inputMode="decimal"
                dir="ltr"
                onChange={(e) => setLng(e.target.value)}
              />
            </label>
            <button
              type="button"
              className="buy-secondary"
              disabled={disabled}
              onClick={() => {
                const a = Number(lat),
                  b = Number(lng);
                if (
                  !lat.trim() ||
                  !lng.trim() ||
                  !Number.isFinite(a) ||
                  !Number.isFinite(b) ||
                  a < -90 ||
                  a > 90 ||
                  b < -180 ||
                  b > 180
                ) {
                  setError("مختصات معتبر وارد کنید.");
                  return;
                }
                pickRef.current({ lat: a, lng: b });
                map.current?.setView([a, b], 16, { animate: false });
              }}
            >
              انتخاب این مختصات
            </button>
          </div>
          <div className="buy-row">
            <button
              type="button"
              className="buy-link"
              disabled={disabled}
              onClick={() => {
                if (!navigator.geolocation) {
                  setError(
                    "مکان‌یابی پشتیبانی نمی‌شود؛ نشانی را دستی وارد کنید.",
                  );
                  return;
                }
                const current = ++sequence.current;
                abort.current?.abort();
                setBusy(false);
                setSuggestion(null);
                navigator.geolocation.getCurrentPosition(
                  (p) => {
                    if (current !== sequence.current) return;
                    pickRef.current({
                      lat: p.coords.latitude,
                      lng: p.coords.longitude,
                    });
                    map.current?.setView(
                      [p.coords.latitude, p.coords.longitude],
                      16,
                      { animate: false },
                    );
                  },
                  () => {
                    if (current === sequence.current)
                      setError(
                        "دسترسی به موقعیت ممکن نشد؛ از نقشه یا ورود دستی استفاده کنید.",
                      );
                  },
                  { timeout: 8000, maximumAge: 0 },
                );
              }}
            >
              موقعیت فعلی من
            </button>
            {(latitude !== undefined || longitude !== undefined) && (
              <button
                type="button"
                className="buy-link"
                onClick={() => {
                  ++sequence.current;
                  abort.current?.abort();
                  setBusy(false);
                  setSuggestion(null);
                  setLat("");
                  setLng("");
                  marker.current?.remove();
                  marker.current = null;
                  callbacks.current.onPick(null);
                }}
              >
                حذف موقعیت
              </button>
            )}
          </div>
          {busy && (
            <div className="address-map-progress" role="status">
              در حال دریافت پیشنهاد نشانی…
            </div>
          )}
          {suggestion && (
            <div className="address-map-suggestion">
              <p>{suggestion.full_address}</p>
              {!suggestion.geography_matched && (
                <p className="buy-alert">
                  استان یا شهر پیشنهادشده در فهرست فعال پیدا نشد؛ مقصد را دستی
                  انتخاب کنید.
                </p>
              )}
              <button
                type="button"
                className="buy-secondary"
                disabled={disabled}
                onClick={() => {
                  callbacks.current.onSuggestion(suggestion);
                  setSuggestion(null);
                }}
              >
                استفاده از نشانی پیشنهادی
              </button>
            </div>
          )}
          {error && (
            <p className="buy-alert" role="alert">
              {error}
            </p>
          )}
        </>
      )}
    </section>
  );
}
