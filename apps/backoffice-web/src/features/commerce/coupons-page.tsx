import { useCallback, useEffect, useState } from "react";
import {
  Alert,
  Button,
  Checkbox,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  TextField,
} from "@mui/material";
import { useWorkspace } from "../../shared/workspace";
import { DataTable, PageHeader } from "../../shared/components";
import { money } from "../../../../../packages/commerce-client/types";

type Coupon = {
  id: string;
  version: number;
  code: string;
  name: string;
  active: boolean;
  kind: "fixed" | "percent";
  value: string;
  target: "items" | "order" | "shipping";
  all_categories: boolean;
  categories: string[];
  min_toman: string;
  max_toman: string;
  cap_toman: string;
  starts_at: string | null;
  ends_at: string | null;
  global_limit: number;
  user_limit: number;
  global_unlimited: boolean;
  user_unlimited: boolean;
  platform_bps: number;
  reserved?: number;
  consumed?: number;
};
type Directory = { items: Coupon[]; permissions: Record<string, boolean> };
const targets = {
  items: "کالاهای واجد شرایط",
  order: "کالاها و ارسال واجد شرایط",
  shipping: "فقط ارسال واجد شرایط",
};
const empty = (): Coupon => ({
  id: "",
  version: 0,
  code: "",
  name: "",
  active: false,
  kind: "percent",
  value: "10",
  target: "items",
  all_categories: true,
  categories: [],
  min_toman: "0",
  max_toman: "0",
  cap_toman: "0",
  starts_at: null,
  ends_at: null,
  global_limit: 0,
  user_limit: 0,
  global_unlimited: true,
  user_unlimited: true,
  platform_bps: 10000,
});
const base = import.meta.env.VITE_API_BASE_URL ?? "";
async function api<T>(
  token: string,
  path: string,
  body?: unknown,
  key?: string,
): Promise<T> {
  const r = await fetch(base + "/backoffice/commerce/coupons" + path, {
    method: body ? "POST" : "GET",
    cache: "no-store",
    headers: {
      Authorization: "Bearer " + token,
      "Content-Type": "application/json",
      ...(key ? { "Idempotency-Key": key } : {}),
    },
    body: body ? JSON.stringify(body) : undefined,
  });
  if (!r.ok) {
    const e = await r.json().catch(() => ({}));
    throw Error(
      (
        {
          forbidden: "دسترسی این عملیات را ندارید.",
          state_conflict:
            "نسخه یا کد تکراری است؛ فهرست را تازه کنید. ورودی شما حفظ شد.",
          coupon_not_eligible:
            "شرایط دسته، تاریخ، مبلغ یا سهمیه این کوپن برقرار نیست.",
          invalid_request:
            "اطلاعات کوپن معتبر نیست؛ مبلغ، بازه، سهم هزینه و سهمیه‌ها را بررسی کنید.",
          session_required: "نشست معتبر نیست؛ دوباره وارد شوید.",
        } as Record<string, string>
      )[e.code] ?? "ارتباط برقرار نشد؛ ورودی حفظ شد، دوباره تلاش کنید.",
    );
  }
  return r.json();
}
export function CouponsPage({ token }: { token: string }) {
  const { state, surface } = useWorkspace();
  const [data, setData] = useState<Directory>(),
    [draft, setDraft] = useState<Coupon>(),
    [search, setSearch] = useState(""),
    [error, setError] = useState(""),
    [busy, setBusy] = useState(false),
    [key, setKey] = useState(crypto.randomUUID());
  const [preview, setPreview] = useState(""),
    [offer, setOffer] = useState(""),
    [shipping, setShipping] = useState("0");
  const load = useCallback(async () => {
    setBusy(true);
    setError("");
    try {
      setData(await api<Directory>(token, ""));
    } catch (e) {
      setError(String((e as Error).message));
    } finally {
      setBusy(false);
    }
  }, [token]);
  useEffect(() => {
    if (surface !== "seller") void load();
  }, [load, surface]);
  if (surface === "seller")
    return <Alert severity="error">مدیریت کوپن مختص بک‌آفیس است.</Alert>;
  const permitted = (action: string) =>
    !!data?.permissions["commerce.coupons." + action];
  const change = <K extends keyof Coupon>(field: K, value: Coupon[K]) => {
    setDraft((d) => (d ? { ...d, [field]: value } : d));
    setKey(crypto.randomUUID());
    setPreview("");
  };
  const editable = !!draft && permitted(draft.id ? "edit" : "create");
  const save = async () => {
    if (!draft || busy || !editable) return;
    setBusy(true);
    setError("");
    try {
      const { reserved: _, consumed: __, ...input } = draft;
      void _;
      void __;
      await api(token, "", input, key);
      setDraft(undefined);
      await load();
    } catch (e) {
      setError((e as Error).message);
    } finally {
      setBusy(false);
    }
  };
  return (
    <>
      <PageHeader
        title="کدهای تخفیف"
        description="قانون نسخه‌دار؛ رزرو سهمیه همراه سفارش و ثبت سهم هزینه چریوا و فروشنده"
        action={
          <Button
            variant="contained"
            disabled={busy || !permitted("create")}
            onClick={() => {
              setDraft(empty());
              setKey(crypto.randomUUID());
              setPreview("");
              setError("");
            }}
          >
            کوپن جدید
          </Button>
        }
      />
      {!draft && error && <Alert severity="error">{error}</Alert>}
      <section className="panel">
        <div className="toolbar">
          <input
            className="search"
            aria-label="جست‌وجوی کوپن"
            placeholder="کد یا عنوان…"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
          />
          <Button disabled={busy} onClick={() => void load()}>
            تازه‌سازی فهرست
          </Button>
        </div>
        {!data ? (
          <p role="status">
            {busy ? "در حال دریافت کوپن‌ها…" : "فهرست دریافت نشد."}
          </p>
        ) : (
          <DataTable
            rows={data.items.filter((c) =>
              (c.name + " " + c.code)
                .toLowerCase()
                .includes(search.toLowerCase()),
            )}
            columns={[
              {
                title: "کد / عنوان",
                render: (c) => (
                  <>
                    <b dir="ltr">{c.code}</b>
                    <small>{c.name}</small>
                  </>
                ),
              },
              {
                title: "تخفیف",
                render: (c) =>
                  c.kind === "percent" ? c.value + "٪" : money(c.value),
              },
              {
                title: "هدف / وضعیت",
                render: (c) => (
                  <>
                    {targets[c.target]}
                    <small>{c.active ? "فعال" : "غیرفعال"}</small>
                  </>
                ),
              },
              {
                title: "سهمیه",
                render: (c) => (
                  <>
                    {c.reserved} رزرو · {c.consumed} مصرف
                    <small>
                      {c.global_unlimited
                        ? "کل نامحدود"
                        : "کل " + c.global_limit}{" "}
                      ·{" "}
                      {c.user_unlimited
                        ? "هر مشتری نامحدود"
                        : "هر مشتری " + c.user_limit}
                    </small>
                  </>
                ),
              },
              {
                title: "عملیات",
                render: (c) => (
                  <>
                    <Button
                      disabled={busy || !permitted("activate")}
                      onClick={async () => {
                        setBusy(true);
                        setError("");
                        try {
                          await api(
                            token,
                            "/activate",
                            { id: c.id, version: c.version, active: !c.active },
                            crypto.randomUUID(),
                          );
                          await load();
                        } catch (e) {
                          setError((e as Error).message);
                        } finally {
                          setBusy(false);
                        }
                      }}
                    >
                      {c.active ? "غیرفعال‌کردن" : "فعال‌کردن"}
                    </Button>
                    <Button
                      onClick={() => {
                        setDraft({ ...c, categories: [...c.categories] });
                        setPreview("");
                        setError("");
                        setKey(crypto.randomUUID());
                      }}
                    >
                      مشاهده و ویرایش
                    </Button>
                  </>
                ),
              },
            ]}
          />
        )}
      </section>
      {draft && (
        <Dialog
          open
          maxWidth="md"
          fullWidth
          onClose={() => {
            if (!busy) setDraft(undefined);
          }}
        >
          <DialogTitle>{draft.id ? "ویرایش کوپن" : "کوپن جدید"}</DialogTitle>
          <form
            onSubmit={(e) => {
              e.preventDefault();
              void save();
            }}
          >
            <DialogContent className="coupon-form">
              {error && <Alert severity="error">{error}</Alert>}
              <fieldset disabled={!editable || busy}>
                <legend>تعریف قانون</legend>
                <TextField
                  label="کد کوپن"
                  value={draft.code}
                  required
                  inputProps={{
                    dir: "ltr",
                    maxLength: 64,
                    pattern: "[A-Za-z0-9][A-Za-z0-9-]{2,63}",
                  }}
                  onChange={(e) => change("code", e.target.value)}
                />
                <TextField
                  label="عنوان کوپن"
                  value={draft.name}
                  required
                  inputProps={{ maxLength: 120 }}
                  onChange={(e) => change("name", e.target.value)}
                />
                <div className="coupon-choices">
                  {(["percent", "fixed"] as const).map((k) => (
                    <Button
                      key={k}
                      variant={draft.kind === k ? "contained" : "outlined"}
                      aria-pressed={draft.kind === k}
                      onClick={() => {
                        change("kind", k);
                        change("cap_toman", "0");
                      }}
                    >
                      {k === "percent" ? "درصدی" : "مبلغ ثابت تومان"}
                    </Button>
                  ))}
                </div>
                <TextField
                  label={
                    draft.kind === "percent"
                      ? "درصد تخفیف"
                      : "مبلغ ثابت (تومان)"
                  }
                  value={draft.value}
                  required
                  inputProps={{ inputMode: "decimal" }}
                  onChange={(e) => change("value", e.target.value)}
                />
                <div className="coupon-choices">
                  {Object.entries(targets).map(([k, n]) => (
                    <Button
                      key={k}
                      aria-pressed={draft.target === k}
                      variant={draft.target === k ? "contained" : "outlined"}
                      onClick={() => change("target", k as Coupon["target"])}
                    >
                      {n}
                    </Button>
                  ))}
                </div>
                <TextField
                  label="حداقل مبلغ مبنا (تومان)"
                  value={draft.min_toman}
                  onChange={(e) => change("min_toman", e.target.value)}
                />
                <TextField
                  label="حداکثر مبلغ مبنا (۰ = نامحدود)"
                  value={draft.max_toman}
                  onChange={(e) => change("max_toman", e.target.value)}
                />
                {draft.kind === "percent" && (
                  <TextField
                    label="سقف تخفیف تومان (۰ = بدون سقف)"
                    value={draft.cap_toman}
                    onChange={(e) => change("cap_toman", e.target.value)}
                  />
                )}
                <p>
                  مبنای درصد و بازه: مبلغ واجد شرایط پس از تخفیف مستقیم و
                  مارکاپ، پیش از خود کوپن. روش‌های نقدی فعلی مارکاپ ندارند.
                  ارسال فقط برای فروشگاه‌های دارای کالای واجد شرایط پوشش داده
                  می‌شود.
                </p>
              </fieldset>
              <fieldset disabled={!editable || busy}>
                <legend>دسته‌ها و زیرشاخه‌ها</legend>
                <label>
                  <Checkbox
                    checked={draft.all_categories}
                    onChange={(e) => {
                      change("all_categories", e.target.checked);
                      change("categories", []);
                    }}
                  />
                  همه دسته‌ها
                </label>
                {!draft.all_categories && (
                  <div className="coupon-category-list">
                    {(state.categories ?? [])
                      .filter((c) => !c.archived)
                      .map((c) => (
                        <label key={c.id}>
                          <Checkbox
                            checked={draft.categories.includes(c.id)}
                            onChange={(e) =>
                              change(
                                "categories",
                                e.target.checked
                                  ? [...draft.categories, c.id]
                                  : draft.categories.filter(
                                      (id) => id !== c.id,
                                    ),
                              )
                            }
                          />
                          {c.name}
                        </label>
                      ))}
                  </div>
                )}
              </fieldset>
              <fieldset disabled={!editable || busy}>
                <legend>زمان و سهمیه</legend>
                <p>
                  زمان‌ها با منطقه زمانی همین دستگاه ثبت می‌شوند؛ بدون زمان یعنی
                  بدون محدودیت همان طرف بازه.
                </p>
                {(["starts_at", "ends_at"] as const).map((field) => (
                  <TextField
                    key={field}
                    label={
                      field === "starts_at"
                        ? "شروع (تقویم میلادی، ساعت دستگاه)"
                        : "پایان (تقویم میلادی، ساعت دستگاه)"
                    }
                    type="datetime-local"
                    InputLabelProps={{ shrink: true }}
                    value={
                      draft[field]
                        ? new Date(
                            new Date(draft[field]!).getTime() -
                              new Date(draft[field]!).getTimezoneOffset() *
                                60000,
                          )
                            .toISOString()
                            .slice(0, 16)
                        : ""
                    }
                    onChange={(e) =>
                      change(
                        field,
                        e.target.value
                          ? new Date(e.target.value).toISOString()
                          : null,
                      )
                    }
                  />
                ))}
                <label>
                  <Checkbox
                    checked={draft.global_unlimited}
                    onChange={(e) => {
                      change("global_unlimited", e.target.checked);
                      change("global_limit", e.target.checked ? 0 : 1);
                    }}
                  />
                  سهمیه کل نامحدود
                </label>
                {!draft.global_unlimited && (
                  <TextField
                    label="سهمیه کل"
                    type="number"
                    value={draft.global_limit}
                    inputProps={{ min: 1, step: 1, max: 1000000000 }}
                    onChange={(e) =>
                      change("global_limit", Number(e.target.value))
                    }
                  />
                )}
                <label>
                  <Checkbox
                    checked={draft.user_unlimited}
                    onChange={(e) => {
                      change("user_unlimited", e.target.checked);
                      change("user_limit", e.target.checked ? 0 : 1);
                    }}
                  />
                  سهمیه هر مشتری نامحدود
                </label>
                {!draft.user_unlimited && (
                  <TextField
                    label="سهمیه هر مشتری"
                    type="number"
                    value={draft.user_limit}
                    inputProps={{ min: 1, step: 1, max: 1000000000 }}
                    onChange={(e) =>
                      change("user_limit", Number(e.target.value))
                    }
                  />
                )}
              </fieldset>
              <fieldset disabled={!editable || busy}>
                <legend>تأمین هزینه</legend>
                <div className="coupon-choices">
                  {[
                    [10000, "چریوا"],
                    [0, "فروشنده"],
                    [5000, "ترکیبی"],
                  ].map(([n, label]) => (
                    <Button
                      key={n}
                      variant={
                        draft.platform_bps === n ? "contained" : "outlined"
                      }
                      onClick={() => change("platform_bps", Number(n))}
                    >
                      {label}
                    </Button>
                  ))}
                </div>
                <TextField
                  label="سهم چریوا (درصد)"
                  type="number"
                  value={draft.platform_bps / 100}
                  inputProps={{ min: 0, max: 100, step: 0.01 }}
                  onChange={(e) =>
                    change(
                      "platform_bps",
                      Math.round(Number(e.target.value) * 100),
                    )
                  }
                />
                <p>
                  باقی سهم فروشنده است؛ جمع سهم‌ها در snapshot خرید ثبت می‌شود.
                </p>
              </fieldset>
              <label>
                <Checkbox
                  disabled={!editable || busy || !permitted("activate")}
                  checked={draft.active}
                  onChange={(e) => change("active", e.target.checked)}
                />
                کوپن فعال باشد
              </label>
              {draft.id && (
                <fieldset disabled={busy}>
                  <legend>پیش‌نمایش قانون ذخیره‌شده</legend>
                  <p>
                    تغییرات فرم ابتدا ذخیره شوند. این محاسبه سهمیه رزرو نمی‌کند
                    و فعال بودن را نادیده می‌گیرد.
                  </p>
                  <TextField
                    label="شناسه عرضه برای آزمون"
                    value={offer}
                    onChange={(e) => setOffer(e.target.value)}
                  />
                  <TextField
                    label="هزینه ارسال آزمون (تومان)"
                    value={shipping}
                    onChange={(e) => setShipping(e.target.value)}
                  />
                  <Button
                    disabled={!offer}
                    onClick={async () => {
                      setBusy(true);
                      setError("");
                      try {
                        const v = await api<Record<string, string>>(
                          token,
                          "/preview",
                          {
                            coupon_id: draft.id,
                            offer_id: offer,
                            quantity: 1,
                            shipping_toman: shipping,
                          },
                        );
                        setPreview(
                          `تخفیف ${money(v.discount_toman)}؛ نهایی ${money(v.total_toman)}`,
                        );
                      } catch (e) {
                        setError((e as Error).message);
                      } finally {
                        setBusy(false);
                      }
                    }}
                  >
                    محاسبه پیش‌نمایش
                  </Button>
                  {preview && <Alert severity="success">{preview}</Alert>}
                </fieldset>
              )}
            </DialogContent>
            <DialogActions>
              <Button disabled={busy} onClick={() => setDraft(undefined)}>
                بستن
              </Button>
              <Button
                disabled={busy || !editable}
                type="submit"
                variant="contained"
              >
                ذخیره کوپن
              </Button>
            </DialogActions>
          </form>
        </Dialog>
      )}
    </>
  );
}
