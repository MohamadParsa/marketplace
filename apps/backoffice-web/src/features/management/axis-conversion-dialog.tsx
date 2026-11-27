import { OptionSelect } from "../../shared/option-select";
import { useEffect, useState } from "react";
import {
  Alert,
  Button,
  Checkbox,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  LinearProgress,
  TextField,
} from "@mui/material";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { useWorkspace } from "../../shared/workspace";
import { ApiError } from "../../shared/api";
import { PaginationControls, usePagedRows } from "../../shared/pagination";
import type { AttributeDefinition, Category } from "../../shared/types";

type Value = { attribute_id: string; value: string };
type Variant = {
  id: string;
  product_id: string;
  sku: string;
  values: Value[];
  version: number;
};
type Source = {
  category: Category;
  attributes: AttributeDefinition[];
  products: { id: string; name: string }[];
  variants: Variant[];
};
type Impact = {
  products: number;
  variants: number;
  offers: number;
  media: number;
  carts: number;
  quotes: number;
  reserved_offers: number;
  reserved_quotes: number;
  preserved_aliases: number;
};
type Preview = {
  fingerprint: string;
  ready: boolean;
  impact: Impact;
  conflicts: {
    code: string;
    variant_id?: string;
    other_variant_id?: string;
    attribute_id?: string;
  }[];
};
type Plan = {
  id: string;
  fingerprint: string;
  expires_at: string;
  impact: Impact;
};
const reasons: Record<string, string> = {
  inconsistent_offer:
    "عرضه به تنوع معتبر همان محصول متصل نیست؛ ابتدا ارتباط عرضه را اصلاح کنید.",
  unchanged_axes: "مجموعه محورهای مقصد تغییری نکرده است.",
  missing_or_invalid_value: "مقدار محور مقصد تعیین نشده یا معتبر نیست.",
  missing_variant_mapping: "نگاشت یک تنوع کامل نشده است.",
  combination_collision: "دو تنوع به ترکیب یکسان می‌رسند؛ نگاشت را اصلاح کنید.",
  active_reservation:
    "رزرو خرید وجود دارد؛ پس از پایان یا لغو عادی آن پیش‌نمایش تازه بگیرید.",
};
async function call<T>(
  token: string,
  path: string,
  data?: unknown,
  key?: string,
): Promise<T> {
  const r = await fetch((import.meta.env.VITE_API_BASE_URL ?? "") + path, {
    method: data === undefined ? "GET" : "POST",
    cache: "no-store",
    headers: {
      Authorization: `Bearer ${token}`,
      "Content-Type": "application/json",
      ...(key ? { "Idempotency-Key": key } : {}),
    },
    body: data === undefined ? undefined : JSON.stringify(data),
  });
  if (!r.ok) {
    const b = await r.json().catch(() => null);
    throw new ApiError(r.status, b?.code);
  }
  return r.json() as Promise<T>;
}
export function AxisConversionDialog({
  category,
  close,
  completed,
}: {
  category: Category;
  close: () => void;
  completed: () => void;
}) {
  const { authToken } = useWorkspace(),
    cache = useQueryClient();
  const source = useQuery<Source>({
    queryKey: ["axis-source", authToken, category.id],
    queryFn: () =>
      call(
        authToken,
        `/backoffice/categories/${encodeURIComponent(category.id)}/axis-conversion`,
      ),
    staleTime: 0,
    refetchOnWindowFocus: false,
    refetchOnReconnect: false,
  });
  const [axes, setAxes] = useState<string[]>([]),
    [mapping, setMapping] = useState<Record<string, Value[]>>({}),
    [search, setSearch] = useState(""),
    [preview, setPreview] = useState<Preview>(),
    [plan, setPlan] = useState<Plan>(),
    [confirmed, setConfirmed] = useState(false),
    [busy, setBusy] = useState(false),
    [error, setError] = useState(""),
    [done, setDone] = useState(false),
    [key, setKey] = useState(crypto.randomUUID());
  const data = source.data;
  useEffect(() => {
    if (data) {
      setAxes(
        (data.category.bindings ?? [])
          .filter((b) => b.axis)
          .map((b) => b.attribute_id),
      );
      setMapping(
        Object.fromEntries(data.variants.map((v) => [v.id, v.values])),
      );
      setPreview(undefined);
      setPlan(undefined);
      setConfirmed(false);
      setKey(crypto.randomUUID());
    }
  }, [data]);
  const reset = () => {
    setPreview(undefined);
    setPlan(undefined);
    setConfirmed(false);
    setError("");
    setKey(crypto.randomUUID());
  };
  const variants = usePagedRows(
    (data?.variants ?? []).filter((v) => {
      const p = data?.products.find((p) => p.id === v.product_id);
      return (
        p?.name +
        " " +
        v.sku +
        " " +
        v.values.map((a) => a.value).join(" ")
      )
        .toLocaleLowerCase()
        .includes(search.trim().toLocaleLowerCase());
    }),
  );
  const variantName = (id?: string) => {
    const v = data?.variants.find((v) => v.id === id);
    return v
      ? (data?.products.find((p) => p.id === v.product_id)?.name ?? "") +
          " · " +
          (v.values.map((a) => a.value).join(" / ") || v.sku || "بدون محور")
      : "";
  };
  const input = () => ({
    category_id: category.id,
    version: data!.category.version,
    axis_ids: axes,
    mapping: data!.variants.map((v) => ({
      variant_id: v.id,
      values: axes.flatMap((id) =>
        (mapping[v.id] ?? []).filter((a) => a.attribute_id === id),
      ),
    })),
  });
  async function run(fn: () => Promise<void>) {
    if (busy) return;
    setBusy(true);
    setError("");
    try {
      await fn();
    } catch (e) {
      setError(e instanceof Error ? e.message : "عملیات انجام نشد.");
    } finally {
      setBusy(false);
    }
  }
  return (
    <Dialog
      open
      fullWidth
      maxWidth="lg"
      onClose={() => {
        if (!busy) close();
      }}
    >
      <DialogTitle>تبدیل محورهای دسته {category.name}</DialogTitle>
      <DialogContent className="axis-conversion-dialog">
        <ol className="axis-steps" aria-label="مراحل تبدیل">
          <li aria-current={!preview ? "step" : undefined}>۱. محور و نگاشت</li>
          <li aria-current={preview && !plan ? "step" : undefined}>
            ۲. بررسی اثر
          </li>
          <li aria-current={plan ? "step" : undefined}>۳. تأیید و اجرا</li>
        </ol>
        {(source.isFetching || busy) && (
          <LinearProgress aria-label="بررسی تبدیل محور" />
        )}
        {error && <Alert severity="error">{error}</Alert>}
        {source.isError && (
          <Alert
            severity="error"
            action={
              <Button onClick={() => void source.refetch()}>تلاش مجدد</Button>
            }
          >
            {source.error.message}
          </Alert>
        )}
        {done ? (
          <Alert severity="success">
            تبدیل کامل شد. قیمت، موجودی، شناسه عرضه‌ها و سفارش‌های قبلی حفظ
            شدند؛ سبدهای مرتبط تأیید مجدد می‌خواهند.
          </Alert>
        ) : (
          data && (
            <>
              <Alert severity="info">
                تنها مقدار محور تنوع تغییر می‌کند. مقدار محور تازه را برای هر
                تنوع انتخاب کنید؛ ادغام دو تنوع مجاز نیست. رزرو خرید مانع
                اجراست.
              </Alert>
              <fieldset disabled={busy || !!plan} className="axis-inputs">
                <legend>۱. محورهای مقصد و نگاشت صریح</legend>
                <div
                  className="axis-options"
                  role="group"
                  aria-label="محورهای مقصد"
                >
                  <OptionSelect
                    label={"محورهای مقصد"}
                    multiple={true}
                    options={data.attributes.map((a) => ({
                      id: String(a.id),
                      name: String("" + a.name + ""),
                      selected: axes.includes(a.id),
                      disabled: false,
                      onSelect: () => {
                        setAxes((old) =>
                          old.includes(a.id)
                            ? old.filter((id) => id !== a.id)
                            : [...old, a.id],
                        );
                        reset();
                      },
                    }))}
                  />
                </div>
                {!axes.length && (
                  <p>
                    بدون محور: فقط محصول دارای یک تنوع می‌تواند ترکیب یکتا داشته
                    باشد.
                  </p>
                )}
                <TextField
                  size="small"
                  label="جست‌وجوی محصول یا تنوع برای نگاشت"
                  value={search}
                  onChange={(e) => setSearch(e.target.value)}
                />
                <p>
                  {data.products.length.toLocaleString("fa-IR")} محصول و{" "}
                  {data.variants.length.toLocaleString("fa-IR")} تنوع؛ نگاشت همه
                  صفحات بررسی می‌شود.
                </p>
                {variants.rows.map((v) => (
                  <article
                    className="axis-variant"
                    key={v.id}
                    data-testid={`axis-variant-${v.id}`}
                  >
                    <h3>{variantName(v.id)}</h3>
                    {axes.map((id) => {
                      const a = data.attributes.find((a) => a.id === id)!;
                      const selected =
                        mapping[v.id]?.find((av) => av.attribute_id === id)
                          ?.value ?? "";
                      return (
                        <details key={id} open={a.options.length <= 5}>
                          <summary>
                            {a.name}: {selected || "انتخاب نشده"}
                          </summary>
                          <div
                            className="axis-options"
                            role="group"
                            aria-label={`مقدار ${a.name}`}
                          >
                            <OptionSelect
                              label={`مقدار ${a.name}`}
                              multiple={false}
                              options={a.options.map((value) => ({
                                id: String(value),
                                name: String("" + value + ""),
                                selected: selected === value,
                                disabled: false,
                                onSelect: () => {
                                  setMapping((old) => ({
                                    ...old,
                                    [v.id]: [
                                      ...(old[v.id] ?? []).filter(
                                        (av) => av.attribute_id !== id,
                                      ),
                                      { attribute_id: id, value },
                                    ],
                                  }));
                                  reset();
                                },
                              }))}
                            />
                          </div>
                        </details>
                      );
                    })}
                  </article>
                ))}
                {!data.variants.length && (
                  <p>این دسته هنوز تنوعی ندارد؛ فقط طرح محور تغییر می‌کند.</p>
                )}
                <PaginationControls {...variants.controls} />
              </fieldset>
              {!plan && (
                <Button
                  variant="outlined"
                  disabled={busy}
                  onClick={() =>
                    void run(async () => {
                      const out = await call<Preview>(
                        authToken,
                        "/backoffice/axis-conversions/preview",
                        { input: input() },
                      );
                      setPreview(out);
                    })
                  }
                >
                  بررسی اثر تبدیل
                </Button>
              )}
              {preview && (
                <section
                  className="axis-impact-box"
                  aria-label="گزارش اثر تبدیل"
                >
                  <h3>۲. نتیجه پیش‌نمایش</h3>
                  <p>
                    محصول: {preview.impact.products} · تنوع:{" "}
                    {preview.impact.variants} · عرضه: {preview.impact.offers} ·
                    تصویر تنوع: {preview.impact.media} · شناسه‌های محفوظ:{" "}
                    {preview.impact.preserved_aliases}
                  </p>
                  <p>
                    سبد نیازمند تأیید: {preview.impact.carts} · پیش‌فاکتور
                    نیازمند بازبینی: {preview.impact.quotes} · رزرو عرضه:{" "}
                    {preview.impact.reserved_offers} · رزرو خرید:{" "}
                    {preview.impact.reserved_quotes}
                  </p>
                  {preview.conflicts.slice(0, 50).map((c, i) => (
                    <Alert severity="warning" key={i}>
                      {reasons[c.code] ?? c.code} {variantName(c.variant_id)}{" "}
                      {c.other_variant_id &&
                        ` / ${variantName(c.other_variant_id)}`}
                    </Alert>
                  ))}
                  {preview.conflicts.length > 50 && (
                    <p>
                      {preview.conflicts.length - 50} مورد دیگر؛ همه نگاشت‌ها را
                      اصلاح و دوباره بررسی کنید.
                    </p>
                  )}
                  {preview.ready && (
                    <Alert severity="success">
                      نگاشت کامل و یکتا است. تا اجرای برنامه، هر تغییر مرتبط
                      نیازمند بررسی تازه خواهد بود.
                    </Alert>
                  )}
                  {!plan && (
                    <Button
                      variant="contained"
                      disabled={busy || !preview.ready}
                      onClick={() =>
                        void run(async () => {
                          setPlan(
                            await call<Plan>(
                              authToken,
                              "/backoffice/axis-conversions/prepare",
                              {
                                input: input(),
                                fingerprint: preview.fingerprint,
                              },
                              key,
                            ),
                          );
                        })
                      }
                    >
                      ساخت برنامه تبدیل
                    </Button>
                  )}
                </section>
              )}
              {plan && (
                <section className="axis-impact-box">
                  <h3>۳. تأیید برنامه و اجرا</h3>
                  <p>
                    اعتبار برنامه تا{" "}
                    {new Date(plan.expires_at).toLocaleString("fa-IR")}؛ تبدیل
                    فقط با نسخه بررسی‌شده اجرا می‌شود.
                  </p>
                  <label className="axis-confirm">
                    <Checkbox
                      checked={confirmed}
                      disabled={busy}
                      onChange={(e) => setConfirmed(e.target.checked)}
                    />
                    نگاشت و اثر آن را بررسی کردم و تبدیل این دسته را تأیید
                    می‌کنم.
                  </label>
                  <div className="toolbar">
                    <Button
                      variant="contained"
                      disabled={!confirmed || busy}
                      onClick={() =>
                        void run(async () => {
                          await call(
                            authToken,
                            "/backoffice/axis-conversions/apply",
                            {
                              plan_id: plan.id,
                              fingerprint: plan.fingerprint,
                              confirm: true,
                            },
                          );
                          setDone(true);
                          void cache.invalidateQueries({
                            queryKey: ["workspace"],
                          });
                          void cache.invalidateQueries({
                            queryKey: ["product"],
                          });
                          void cache.invalidateQueries({
                            queryKey: ["variant-resolution"],
                          });
                          void cache.invalidateQueries({
                            queryKey: ["axis-source"],
                          });
                        })
                      }
                    >
                      اجرای تبدیل تأییدشده
                    </Button>
                    <Button disabled={busy} onClick={reset}>
                      بازگشت به اصلاح نگاشت
                    </Button>
                  </div>
                </section>
              )}
            </>
          )
        )}
      </DialogContent>
      <DialogActions>
        {!done && (
          <Button
            disabled={busy || source.isFetching}
            onClick={() => void source.refetch()}
          >
            بازخوانی مبدأ و شروع نگاشت تازه
          </Button>
        )}
        <Button disabled={busy} onClick={done ? completed : close}>
          {done ? "پایان تبدیل" : "بستن بدون اجرا"}
        </Button>
      </DialogActions>
    </Dialog>
  );
}
