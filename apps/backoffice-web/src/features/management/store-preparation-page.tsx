import { useState } from "react";
import { Alert, Button, LinearProgress, TextField } from "@mui/material";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { useWorkspace } from "../../shared/workspace";
import { ApiError } from "../../shared/api";
import { PageHeader } from "../../shared/components";

type Calendar = {
  version: number;
  working_weekdays: number[];
  holidays: string[];
};
type Policy = {
  version: number;
  day_basis: "working" | "calendar";
  calendar_mode: "inherit" | "override";
  working_weekdays: number[];
};
type StorePreparation = {
  id: string;
  name: string;
  version: number;
  lead_days: number;
  policy: Policy;
  site_calendar: Calendar;
  effective_calendar: Calendar;
  timing: {
    ready_time: string;
    cutoff_time: string;
    min_slot_lead_minutes: number;
    allow_same_day_slots: boolean;
  };
  can_write: boolean;
};
const weekdays = [
  [6, "شنبه"],
  [0, "یکشنبه"],
  [1, "دوشنبه"],
  [2, "سه‌شنبه"],
  [3, "چهارشنبه"],
  [4, "پنج‌شنبه"],
  [5, "جمعه"],
] as const;
const dateTime = (value: string) =>
  new Intl.DateTimeFormat("fa-IR", {
    timeZone: "Asia/Tehran",
    dateStyle: "full",
    timeStyle: "short",
  }).format(new Date(value));

export function StorePreparationPage() {
  const { authToken, surface } = useWorkspace();
  const client = useQueryClient();
  const [selected, setSelected] = useState("");
  const [filter, setFilter] = useState("");
  const [editing, setEditing] = useState(false);
  const [revision, setRevision] = useState(0);
  const [notice, setNotice] = useState("");
  const root = (import.meta.env.VITE_API_BASE_URL ?? "") + `/${surface}/stores`;
  const queryKey = ["store-preparation", surface, authToken];
  async function request<T>(
    path: string,
    method = "GET",
    body?: unknown,
    key?: string,
  ): Promise<T> {
    const response = await fetch(root + path, {
      method,
      cache: "no-store",
      headers: {
        Authorization: `Bearer ${authToken}`,
        ...(body ? { "Content-Type": "application/json" } : {}),
        ...(key ? { "Idempotency-Key": key } : {}),
      },
      body: body ? JSON.stringify(body) : undefined,
    });
    if (!response.ok) {
      const error = await response.json().catch(() => null);
      throw new ApiError(response.status, error?.code);
    }
    return response.json();
  }
  const query = useQuery({
    queryKey,
    queryFn: () => request<StorePreparation[]>("/preparation"),
    refetchOnWindowFocus: false,
  });
  const rows = query.data ?? [];
  const store = rows.find((row) => row.id === selected) ?? rows[0];
  return (
    <div className="store-preparation-page">
      <PageHeader
        title="وقفه تأمین و تقویم فروشگاه"
        description="روزهای کاری، واحد وقفه و زمان محاسبهٔ آماده‌سازی هر فروشگاه"
      />
      {query.isPending && <LinearProgress aria-label="دریافت تقویم فروشگاه" />}
      {query.error && (
        <Alert severity="error">
          دریافت تنظیمات انجام نشد.{" "}
          <Button onClick={() => void query.refetch()}>تلاش دوباره</Button>
        </Alert>
      )}
      {notice && <Alert severity="success">{notice}</Alert>}
      {query.isSuccess && !rows.length && (
        <Alert severity="info">
          فروشگاهی با مجوز مشاهدهٔ وقفه تأمین در دسترس نیست. مدیر می‌تواند مجوز
          مشاهده و ویرایش را در نقش فروشگاه تعیین کند.
        </Alert>
      )}
      {!!rows.length && (
        <section className="store-detail-box">
          <TextField
            label="جست‌وجوی فروشگاه"
            size="small"
            value={filter}
            disabled={editing}
            onChange={(event) => setFilter(event.target.value)}
          />
          <div
            className="toolbar"
            role="group"
            aria-label="انتخاب فروشگاه"
            style={{ flexWrap: "wrap", marginTop: 12 }}
          >
            {rows
              .filter((row) => row.name.includes(filter))
              .map((row) => (
                <Button
                  key={row.id}
                  disabled={editing}
                  aria-pressed={store?.id === row.id}
                  variant={store?.id === row.id ? "contained" : "outlined"}
                  onClick={() => setSelected(row.id)}
                >
                  {row.name}
                </Button>
              ))}
          </div>
        </section>
      )}
      {store && (
        <PreparationEditor
          key={`${store.id}:${store.version}:${store.site_calendar.version}:${revision}`}
          store={store}
          busyChanged={setEditing}
          save={(body, key) =>
            request<StorePreparation>(
              `/${encodeURIComponent(store.id)}/preparation`,
              "PUT",
              body,
              key,
            )
          }
          preview={(body) =>
            request<{ ready_at: string; source_at: string }>(
              `/${encodeURIComponent(store.id)}/preparation/preview`,
              "POST",
              body,
            )
          }
          saved={(row) => {
            client.setQueryData<StorePreparation[]>(queryKey, (previous) =>
              previous?.map((old) => (old.id === row.id ? row : old)),
            );
            setNotice("تنظیمات وقفه و تقویم ذخیره شد.");
          }}
          reload={() =>
            void query.refetch().then((result) => {
              if (!result.isError) {
                setRevision((value) => value + 1);
                setNotice("");
              }
            })
          }
        />
      )}
    </div>
  );
}

type Input = {
  version: number;
  lead_days: number;
  day_basis: Policy["day_basis"];
  calendar_mode: Policy["calendar_mode"];
  working_weekdays: number[];
};
function PreparationEditor({
  store,
  save,
  preview,
  saved,
  reload,
  busyChanged,
}: {
  store: StorePreparation;
  save: (body: Input, key: string) => Promise<StorePreparation>;
  preview: (body: Input) => Promise<{ ready_at: string; source_at: string }>;
  saved: (row: StorePreparation) => void;
  reload: () => void;
  busyChanged: (busy: boolean) => void;
}) {
  const [lead, setLead] = useState(String(store.lead_days));
  const [basis, setBasis] = useState(store.policy.day_basis);
  const [mode, setMode] = useState(store.policy.calendar_mode);
  const [days, setDays] = useState(store.policy.working_weekdays);
  const [key, setKey] = useState(crypto.randomUUID());
  const [busy, setBusy] = useState(false);
  const [dirty, setDirty] = useState(false);
  const [error, setError] = useState("");
  const [conflict, setConflict] = useState(false);
  const [notice, setNotice] = useState("");
  const [example, setExample] = useState<{
    ready_at: string;
    source_at: string;
  }>();
  const effective =
    mode === "inherit" ? store.site_calendar.working_weekdays : days;
  const valid =
    /^\d+$/.test(lead) &&
    Number(lead) <= 365 &&
    (mode === "inherit" || days.length > 0);
  function change() {
    setDirty(true);
    setKey(crypto.randomUUID());
    setNotice("");
    setExample(undefined);
  }
  const input = (): Input => ({
    version: store.version,
    lead_days: Number(lead),
    day_basis: basis,
    calendar_mode: mode,
    working_weekdays: mode === "inherit" ? [] : days,
  });
  async function run(saving: boolean) {
    if (busy || !valid || (saving && !store.can_write)) return;
    setBusy(true);
    busyChanged(true);
    setError("");
    setConflict(false);
    try {
      if (saving) {
        const next = await save(input(), key);
        setDirty(false);
        setNotice("تنظیمات ذخیره شد.");
        saved(next);
      } else setExample(await preview(input()));
    } catch (error) {
      setConflict(error instanceof ApiError && error.status === 409);
      setError(
        error instanceof ApiError && error.status === 403
          ? "مجوز تغییر این فروشگاه را ندارید؛ تنظیمات و دسترسی را تازه کنید."
          : error instanceof ApiError && error.status === 409
            ? "نسخهٔ تنظیمات تغییر کرده است؛ پیش از ذخیره دوباره نسخهٔ تازه را دریافت کنید."
            : "عملیات انجام نشد؛ ورودی‌ها را بررسی و دوباره تلاش کنید. پیش‌نویس حفظ شده است.",
      );
    } finally {
      setBusy(false);
      busyChanged(false);
    }
  }
  return (
    <section
      className="store-detail-box"
      aria-label={`تنظیم وقفه ${store.name}`}
    >
      <h2>{store.name}</h2>
      {!store.can_write && (
        <Alert severity="info">
          مجوز ویرایش ندارید؛ فقط مشاهده و پیش‌نمایش در دسترس است.
        </Alert>
      )}
      {notice && <Alert severity="success">{notice}</Alert>}
      {error && (
        <Alert severity="error">
          {error}
          {conflict && (
            <Button disabled={busy} onClick={reload}>
              دریافت نسخهٔ تازه (حذف پیش‌نویس)
            </Button>
          )}
        </Alert>
      )}
      {busy && <LinearProgress aria-label="محاسبه یا ذخیرهٔ وقفه" />}
      <fieldset disabled={busy || !store.can_write} className="axis-inputs">
        <legend>وقفهٔ تأمین فروشگاه</legend>
        <TextField
          label="تعداد روز وقفه"
          type="number"
          inputProps={{ min: 0, max: 365 }}
          value={lead}
          error={!/^\d+$/.test(lead) || Number(lead) > 365}
          helperText="از صفر تا ۳۶۵ روز؛ مبدأ، پرداخت قطعی سفارش است."
          onChange={(event) => {
            setLead(event.target.value);
            change();
          }}
        />
        <div className="toolbar" role="group" aria-label="واحد وقفه تأمین">
          {(
            [
              ["working", "روز کاری"],
              ["calendar", "روز تقویمی"],
            ] as const
          ).map(([value, label]) => (
            <Button
              key={value}
              aria-pressed={basis === value}
              variant={basis === value ? "contained" : "outlined"}
              onClick={() => {
                setBasis(value);
                change();
              }}
            >
              {label}
            </Button>
          ))}
        </div>
        <div className="toolbar" role="group" aria-label="منبع روزهای کاری">
          {(
            [
              ["inherit", "ارث‌بری از سایت"],
              ["override", "روزهای اختصاصی فروشگاه"],
            ] as const
          ).map(([value, label]) => (
            <Button
              key={value}
              aria-pressed={mode === value}
              variant={mode === value ? "contained" : "outlined"}
              onClick={() => {
                setMode(value);
                if (value === "override" && !days.length)
                  setDays(store.site_calendar.working_weekdays);
                change();
              }}
            >
              {label}
            </Button>
          ))}
        </div>
        {mode === "override" && (
          <div
            className="toolbar"
            role="group"
            aria-label="روزهای کاری اختصاصی"
            style={{ flexWrap: "wrap" }}
          >
            {weekdays.map(([day, label]) => (
              <Button
                key={day}
                aria-pressed={days.includes(day)}
                variant={days.includes(day) ? "contained" : "outlined"}
                onClick={() => {
                  setDays(
                    days.includes(day)
                      ? days.filter((value) => value !== day)
                      : [...days, day],
                  );
                  change();
                }}
              >
                {label}
              </Button>
            ))}
          </div>
        )}
      </fieldset>
      {!valid && (
        <Alert severity="warning">
          تعداد روز باید بین صفر و ۳۶۵ باشد و حداقل یک روز کاری انتخاب شود.
        </Alert>
      )}
      <h3>تقویم مؤثر</h3>
      <p>
        {weekdays
          .filter(([day]) => effective.includes(day))
          .map(([, label]) => label)
          .join("، ") || "هیچ روز کاری انتخاب نشده است."}
      </p>
      <p>
        تعطیلات عمومی سایت در هر دو حالت اعمال می‌شوند (
        {store.site_calendar.holidays.length.toLocaleString("fa-IR")} تاریخ).
      </p>
      {store.site_calendar.holidays.length > 0 && (
        <details>
          <summary>مشاهدهٔ تعطیلات عمومی</summary>
          {store.site_calendar.holidays.map((value) => (
            <p key={value}>
              {new Intl.DateTimeFormat("fa-IR", {
                timeZone: "Asia/Tehran",
                dateStyle: "full",
              }).format(new Date(value + "T12:00:00+03:30"))}
            </p>
          ))}
        </details>
      )}
      <p>
        زمان تهران · آماده‌سازی {store.timing.ready_time} · مهلت شروع همان روز{" "}
        {store.timing.cutoff_time}. پایان وقفه در تعطیلی به اولین روز کاری منتقل
        می‌شود.
      </p>
      <p>
        پرداخت از ساعت مهلت به بعد، از اولین روز کاری بعد محاسبه می‌شود. روز
        شروع در وقفهٔ مثبت شمرده نمی‌شود؛ وقفهٔ صفر زودتر از زمان پرداخت آماده
        نمی‌شود.
      </p>
      <div className="toolbar" style={{ flexWrap: "wrap" }}>
        <Button disabled={busy || !valid} onClick={() => void run(false)}>
          پیش‌نمایش با فرض پرداخت الان
        </Button>
        <Button
          variant="contained"
          disabled={busy || !valid || !dirty || !store.can_write || conflict}
          onClick={() => void run(true)}
        >
          ذخیرهٔ وقفه و تقویم
        </Button>
        <Button disabled={busy} onClick={reload}>
          دریافت آخرین تنظیمات
        </Button>
      </div>
      {example && (
        <Alert severity="info">
          با فرض پرداخت در {dateTime(example.source_at)}، آماده‌سازی در{" "}
          {dateTime(example.ready_at)} است. زمان تحویل به مشتری پس از افزودن مدت
          حمل محاسبه می‌شود.
        </Alert>
      )}
    </section>
  );
}
