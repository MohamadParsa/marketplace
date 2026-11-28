import { useEffect, useState } from "react";
import {
  Alert,
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  LinearProgress,
} from "@mui/material";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { useWorkspace } from "../../shared/workspace";
import { ApiError } from "../../shared/api";
import { ChoiceField } from "../../shared/option-select";
import { PaginationControls, usePagedRows } from "../../shared/pagination";
import type {
  AttributeBinding,
  AttributeDefinition,
  Category,
  Variant,
} from "../../shared/types";

type Value = { attribute_id: string; value: string };
type Source = {
  category: Category;
  attributes: AttributeDefinition[];
  products: { id: string; name: string }[];
  variants: Variant[];
};
type Preview = {
  fingerprint: string;
  ready: boolean;
  impact: {
    products: number;
    variants: number;
    offers: number;
    carts: number;
    quotes: number;
  };
  conflicts: {
    code: string;
    variant_id?: string;
    other_variant_id?: string;
    attribute_id?: string;
  }[];
};
type Plan = { id: string; fingerprint: string };
const reasons: Record<string, string> = {
  unchanged_axes: "محورها تغییری نکرده‌اند.",
  inconsistent_offer: "ارتباط یک عرضه با تنوع نیاز به اصلاح دارد.",
  missing_or_invalid_value: "مقدار محور تازه را تکمیل کنید.",
  missing_variant_mapping: "مقادیر یک تنوع کامل نیست.",
  combination_collision:
    "دو تنوع به ترکیب یکسان می‌رسند؛ مقدارها را تغییر دهید.",
  active_reservation:
    "خرید در حال انجام است؛ پس از پایان رزرو دوباره بررسی کنید.",
};
async function call<T>(
  token: string,
  path: string,
  data?: unknown,
  key?: string,
): Promise<T> {
  const response = await fetch(
    (import.meta.env.VITE_API_BASE_URL ?? "") + path,
    {
      method: data === undefined ? "GET" : "POST",
      cache: "no-store",
      headers: {
        Authorization: "Bearer " + token,
        "Content-Type": "application/json",
        ...(key ? { "Idempotency-Key": key } : {}),
      },
      body: data === undefined ? undefined : JSON.stringify(data),
    },
  );
  if (!response.ok) {
    const body = await response.json().catch(() => null);
    throw new ApiError(response.status, body?.code);
  }
  return response.json();
}
export function AxisConversionDialog({
  category,
  proposedBindings,
  close,
  completed,
}: {
  category: Category;
  proposedBindings?: AttributeBinding[];
  close: () => void;
  completed: () => void;
}) {
  const { authToken, state } = useWorkspace(),
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
  const data = source.data;
  const bindings = proposedBindings ?? data?.category.bindings ?? [];
  const axes = bindings.filter((b) => b.axis).map((b) => b.attribute_id);
  const [mapping, setMapping] = useState<Record<string, Value[]>>({});
  const [initialized, setInitialized] = useState(false),
    [busy, setBusy] = useState(false),
    [error, setError] = useState("");
  const [conflicted, setConflicted] = useState(false);
  const stale = !!data && data.category.version !== category.version;
  const [key, setKey] = useState(crypto.randomUUID());
  useEffect(() => {
    if (!data) return;
    const axisIDs = (proposedBindings ?? data.category.bindings)
      .filter((b) => b.axis)
      .map((b) => b.attribute_id);
    setMapping(
      Object.fromEntries(
        data.variants.map((v) => [
          v.id,
          axisIDs.flatMap((id) => {
            const a = data.attributes.find((a) => a.id === id);
            const value =
              v.values.find((value) => value.attribute_id === id)?.value ??
              state.products
                ?.find((p) => p.id === v.product_id)
                ?.attributes?.find((value) => value.attribute_id === id)
                ?.value ??
              (a?.options.length === 1 ? a.options[0] : undefined);
            return value && a?.options.includes(value)
              ? [{ attribute_id: id, value }]
              : [];
          }),
        ]),
      ),
    );
    setInitialized(true);
    setKey(crypto.randomUUID());
  }, [data, proposedBindings]);
  const input = {
    category_id: category.id,
    version: category.version,
    axis_ids: axes,
    bindings,
    mapping: (data?.variants ?? []).map((v) => ({
      variant_id: v.id,
      values: axes.flatMap((id) =>
        (mapping[v.id] ?? []).filter((value) => value.attribute_id === id),
      ),
    })),
  };
  const complete =
    initialized &&
    (data?.variants ?? []).every((v) =>
      axes.every(
        (id) => !!mapping[v.id]?.some((value) => value.attribute_id === id),
      ),
    );
  const preview = useQuery<Preview>({
    queryKey: ["axis-preview", authToken, input],
    queryFn: () =>
      call(authToken, "/backoffice/axis-conversions/preview", { input }),
    enabled: !!data && complete && !busy && !stale && !conflicted,
    retry: false,
  });
  const rows = usePagedRows(data?.variants ?? []);
  const label = (id?: string) => {
    const variant = data?.variants.find((v) => v.id === id);
    return variant
      ? (data?.products.find((p) => p.id === variant.product_id)?.name ??
          "محصول") +
          " · " +
          (variant.values.map((value) => value.value).join(" / ") ||
            "بدون محور")
      : "";
  };
  const mappingNeeded =
    axes.some(
      (id) =>
        !data?.category.bindings.some((b) => b.axis && b.attribute_id === id),
    ) || !!preview.data?.conflicts.length;
  return (
    <Dialog open fullWidth maxWidth="lg" onClose={() => !busy && close()}>
      <DialogTitle>تکمیل تغییر محورها · {category.name}</DialogTitle>
      <DialogContent className="axis-conversion-dialog">
        {(stale || conflicted) && (
          <Alert
            severity="warning"
            action={
              <Button
                disabled={busy}
                onClick={async () => {
                  await cache.invalidateQueries({ queryKey: ["workspace"] });
                  completed();
                }}
              >
                دریافت تنظیمات تازه و بستن پیش‌نویس
              </Button>
            }
          >
            دسته از زمان بازشدن زون‌ها تغییر کرده است؛ برای جلوگیری از جایگزینی
            تغییرات دیگر، تنظیمات تازه را دریافت کنید.
          </Alert>
        )}
        {(source.isFetching || preview.isFetching || busy) && (
          <LinearProgress aria-label="بررسی تغییر محور" />
        )}
        <Alert severity="info">
          قیمت، موجودی و سوابق حفظ می‌شوند. مقدارهای قابل تشخیص از تنوع فعلی
          تکمیل شده‌اند؛ فقط مقدارهای باقی‌مانده را انتخاب و تغییر را ثبت کنید.
        </Alert>
        {(error || source.error || preview.error) && (
          <Alert severity="error">
            {error || source.error?.message || preview.error?.message}
            <Button
              disabled={busy}
              onClick={() => {
                if (source.error) void source.refetch();
                else void preview.refetch();
              }}
            >
              تلاش دوباره
            </Button>
          </Alert>
        )}
        {data && (
          <>
            <p>
              {data.products.length.toLocaleString("fa-IR")} محصول ·{" "}
              {data.variants.length.toLocaleString("fa-IR")} تنوع · محورهای
              مقصد:{" "}
              {axes
                .map((id) => data.attributes.find((a) => a.id === id)?.name)
                .join("، ") || "بدون محور"}
            </p>
            {mappingNeeded && (
              <div className="axis-mapping-compact">
                {rows.rows.map((v) => (
                  <article key={v.id} className="axis-variant">
                    <b>{label(v.id)}</b>
                    <div className="commerce-grid">
                      {axes.map((id) => {
                        const a = data.attributes.find((a) => a.id === id);
                        if (!a) return null;
                        return (
                          <ChoiceField
                            key={id}
                            label={a.name}
                            value={
                              mapping[v.id]?.find(
                                (value) => value.attribute_id === id,
                              )?.value ?? ""
                            }
                            options={a.options.map((value) => ({
                              id: value,
                              name: value,
                            }))}
                            disabled={busy}
                            clearLabel="انتخاب مقدار"
                            onChange={(value) => {
                              setMapping((old) => ({
                                ...old,
                                [v.id]: [
                                  ...(old[v.id] ?? []).filter(
                                    (av) => av.attribute_id !== id,
                                  ),
                                  ...(value
                                    ? [{ attribute_id: id, value }]
                                    : []),
                                ],
                              }));
                              setKey(crypto.randomUUID());
                              setError("");
                            }}
                          />
                        );
                      })}
                    </div>
                  </article>
                ))}
                <PaginationControls {...rows.controls} />
              </div>
            )}
            {!complete && (
              <Alert severity="warning">
                مقدار محورهای تازه را برای همهٔ تنوع‌ها تکمیل کنید.
              </Alert>
            )}
            {preview.data?.conflicts.slice(0, 30).map((conflict, i) => (
              <Alert key={i} severity="warning">
                {reasons[conflict.code] ?? "تغییر نیاز به بررسی دارد."}{" "}
                {label(conflict.variant_id)}
                {conflict.other_variant_id
                  ? " / " + label(conflict.other_variant_id)
                  : ""}
              </Alert>
            ))}
            {preview.data?.ready && (
              <Alert severity="success">
                تغییر قابل ثبت است؛{" "}
                {preview.data.impact.offers.toLocaleString("fa-IR")} عرضه با
                قیمت و موجودی فعلی حفظ می‌شود.
              </Alert>
            )}
          </>
        )}
      </DialogContent>
      <DialogActions>
        <Button disabled={busy} onClick={close}>
          بازگشت به زون‌ها
        </Button>
        <Button
          variant="contained"
          disabled={
            busy ||
            stale ||
            conflicted ||
            !complete ||
            !preview.data?.ready ||
            preview.isFetching
          }
          onClick={async () => {
            if (busy || !preview.data?.ready) return;
            setBusy(true);
            setError("");
            try {
              const plan = await call<Plan>(
                authToken,
                "/backoffice/axis-conversions/prepare",
                { input, fingerprint: preview.data.fingerprint },
                key,
              );
              await call(authToken, "/backoffice/axis-conversions/apply", {
                plan_id: plan.id,
                fingerprint: plan.fingerprint,
                confirm: true,
              });
              await Promise.all(
                ["workspace", "product", "variant-resolution"].map((queryKey) =>
                  cache.invalidateQueries({ queryKey: [queryKey] }),
                ),
              );
              completed();
            } catch (cause) {
              if (cause instanceof ApiError && cause.status === 409)
                setConflicted(true);
              setError(
                cause instanceof ApiError && cause.status === 409
                  ? "اطلاعات تغییر کرده است؛ برای بازبینی تازه به زون‌ها برگردید و دوباره ذخیره کنید."
                  : cause instanceof Error
                    ? cause.message
                    : "تغییر ثبت نشد.",
              );
            } finally {
              setBusy(false);
            }
          }}
        >
          {busy ? "در حال ثبت…" : "تأیید و ثبت تغییر محورها"}
        </Button>
      </DialogActions>
    </Dialog>
  );
}
