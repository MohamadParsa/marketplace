import { CreateButton } from "../../shared/create-button";
import { ChoiceField, MultiChoiceField } from "../../shared/option-select";
import { FilterPanel } from "../../shared/filter-panel";
import { OptionSelect } from "../../shared/option-select";
import { useState } from "react";
import { Link } from "react-router-dom";
import {
  Alert,
  Button,
  Checkbox,
  Chip,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  FormControlLabel,
  LinearProgress,
  TextField,
} from "@mui/material";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { useWorkspace } from "../../shared/workspace";
import { ApiError } from "../../shared/api";
import { DataTable, PageHeader } from "../../shared/components";
import { CategoryTree, categoryPath } from "../management/category-tree";
import { WeeklyScheduleEditor, scheduleError } from "./weekly-schedule-editor";
import { StoreShippingPanel } from "./store-shipping-page";
import type {
  Shipping,
  Province,
} from "../../../../../packages/commerce-client/types";

export type Method = Shipping & {
  version: number;
  unresolved_cities?: string[];
};
type Directory = {
  rows: Method[];
  total: number;
  page: number;
  size: number;
  can_write: boolean;
  geography: Province[];
};
type Filters = {
  query: string;
  active: string;
  collection_mode: string;
  delivery_mode: string;
  category_id: string;
  city_id: string;
};
export class MethodError extends ApiError {
  constructor(
    status: number,
    code?: string,
    public fields: Record<string, string> = {},
  ) {
    super(status, code);
  }
}
const defaults: Filters = {
  query: "",
  active: "",
  collection_mode: "",
  delivery_mode: "",
  category_id: "",
  city_id: "",
};
const blank = (): Method => ({
  id: "ship-" + crypto.randomUUID(),
  version: 0,
  name: "",
  active: false,
  all_cities: true,
  cities: [],
  city_ids: [],
  store_ids: [],
  fee_toman: "0",
  min_days: 0,
  max_days: 0,
  day_basis: "working",
  collection_mode: "prepaid",
  delivery_mode: "range",
  category_mode: "all",
  allowed_categories: [],
  denied_categories: [],
});

export function ShippingMethodsPage() {
  const { authToken, state, can } = useWorkspace();
  const client = useQueryClient();
  const [filter, setFilter] = useState<Filters>(defaults),
    [page, setPage] = useState(0),
    [size, setSize] = useState(20),
    [editing, setEditing] = useState<{ method: Method; create: boolean }>(),
    [notice, setNotice] = useState("");
  const [draftFilter, setDraftFilter] = useState<Filters>(defaults);
  const [storeMethod, setStoreMethod] = useState<Method>();
  const [storeBusy, setStoreBusy] = useState(false);
  const root =
    (import.meta.env.VITE_API_BASE_URL ?? "") + "/backoffice/shipping/methods";
  async function request<T>(
    path: string,
    method = "GET",
    body?: unknown,
    key?: string,
    signal?: AbortSignal,
  ): Promise<T> {
    const response = await fetch(root + path, {
      method,
      cache: "no-store",
      signal,
      headers: {
        Authorization: `Bearer ${authToken}`,
        ...(body ? { "Content-Type": "application/json" } : {}),
        ...(key ? { "Idempotency-Key": key } : {}),
      },
      body: body ? JSON.stringify(body) : undefined,
    });
    if (!response.ok) {
      const error = await response.json().catch(() => null);
      throw new MethodError(response.status, error?.code, error?.fields);
    }
    return response.json();
  }
  const query = useQuery({
    queryKey: ["shipping-methods", authToken, filter, page, size],
    queryFn: () =>
      request<Directory>(
        "?" +
          new URLSearchParams({
            ...filter,
            page: String(page),
            size: String(size),
          }),
      ),
  });
  const directory = query.data;
  const change = (patch: Partial<Filters>) => {
    setDraftFilter({ ...draftFilter, ...patch });
  };
  function choices(
    field: "active" | "collection_mode" | "delivery_mode",
    label: string,
    options: [string, string][],
  ) {
    return (
      <div role="group" aria-label={label} className="toolbar">
        <OptionSelect
          label={label}
          multiple={false}
          options={options.map(([value, name]) => ({
            id: String(value),
            name: String("" + name + ""),
            selected: draftFilter[field] === value,
            disabled: false,
            onSelect: () => change({ [field]: value }),
          }))}
        />
      </div>
    );
  }
  return (
    <div className="shipping-methods-page">
      <PageHeader
        title="روش‌های عمومی حمل"
        description="دریافت هزینه، زمان حمل و پوشش مجاز هر روش"
        action={
          <CreateButton
            disabled={!directory?.can_write}
            variant="contained"
            onClick={() => {
              setEditing({ method: blank(), create: true });
              setNotice("");
            }}
          >
            روش حمل جدید
          </CreateButton>
        }
      />
      {notice && <Alert severity="success">{notice}</Alert>}
      <FilterPanel
        onSearch={() => {
          if (JSON.stringify(filter) === JSON.stringify(draftFilter))
            void query.refetch();
          else setFilter({ ...draftFilter });
          setPage(0);
        }}
      >
        <section className="store-detail-box shipping-method-filters">
          <h2 className="filter-panel-title">فیلترهای روش حمل</h2>
          <div className="shipping-filter-fields">
            <TextField
              label="جست‌وجوی نام روش"
              value={draftFilter.query}
              onChange={(event) => change({ query: event.target.value })}
              size="small"
            />
            {choices("active", "وضعیت", [
              ["", "همه"],
              ["true", "فعال"],
              ["false", "غیرفعال"],
            ])}
            {choices("collection_mode", "دریافت هزینه", [
              ["", "همه"],
              ["prepaid", "پیش‌کرایه"],
              ["collect", "پس‌کرایه"],
            ])}
            {choices("delivery_mode", "زمان تحویل", [
              ["", "همه"],
              ["range", "بازهٔ روز"],
              ["slots", "بازهٔ ساعت"],
            ])}
          </div>
          <details>
            <summary>فیلتر دسته و شهر</summary>
            <CategoryTree
              categories={state.categories ?? []}
              value={draftFilter.category_id}
              onChange={(category_id) => change({ category_id })}
            />
            <ChoiceField
              label="شهر"
              value={draftFilter.city_id}
              clearLabel="همه شهرها"
              options={(directory?.geography ?? []).flatMap((p) =>
                p.cities.map((c) => ({
                  id: c.id,
                  name: p.name + " / " + c.name,
                })),
              )}
              onChange={(city_id) => change({ city_id })}
            />
          </details>
          <div className="toolbar">
            <Button
              onClick={() => {
                setFilter(defaults);
                setDraftFilter(defaults);
                setPage(0);
              }}
            >
              پاک‌کردن فیلترها
            </Button>
            <Button component={Link} to="/commerce-settings">
              تقویم عمومی و ساعت آماده‌سازی
            </Button>
          </div>
        </section>
      </FilterPanel>
      {query.isFetching && <LinearProgress aria-label="دریافت روش‌های حمل" />}
      {query.error && (
        <Alert severity="error">
          دریافت روش‌ها انجام نشد.{" "}
          <Button onClick={() => void query.refetch()}>تلاش دوباره</Button>
        </Alert>
      )}
      {directory && (
        <>
          {!directory.can_write && (
            <Alert severity="info">
              فقط مشاهده؛ مجوز ویرایش روش‌های عمومی را ندارید.
            </Alert>
          )}
          <DataTable
            rows={directory.rows}
            remote={{
              controls: {
                page: directory.page,
                pages: Math.max(1, Math.ceil(directory.total / directory.size)),
                size: directory.size,
                total: directory.total,
                setPage,
                setSize: (value) => {
                  setSize(value);
                  setPage(0);
                },
              },
              filterKey: JSON.stringify(filter),
              exportRows: async (ids, signal) => {
                const records: Method[] = [];
                for (let current = 0; ; current++) {
                  const result = await request<Directory>(
                    "?" +
                      new URLSearchParams({
                        ...filter,
                        page: String(current),
                        size: "100",
                      }),
                    "GET",
                    undefined,
                    undefined,
                    signal,
                  );
                  records.push(...result.rows);
                  if (records.length >= result.total || !result.rows.length)
                    break;
                }
                return ids
                  ? records.filter((row) => ids.includes(row.id))
                  : records;
              },
            }}
            columns={[
              {
                title: "روش حمل",
                render: (row) => (
                  <>
                    <b>{row.name}</b>
                    <small>
                      {row.active ? "فعال" : "غیرفعال"} · نسخه{" "}
                      {row.version.toLocaleString("fa-IR")}
                    </small>
                  </>
                ),
              },
              {
                title: "هزینه",
                render: (row) =>
                  row.collection_mode === "collect"
                    ? "پس‌کرایه؛ پرداخت به حمل‌کننده"
                    : row.fee_toman === "0"
                      ? "پیش‌کرایه رایگان"
                      : `${row.fee_toman} تومان پیش‌کرایه`,
              },
              {
                title: "زمان تحویل",
                render: (row) =>
                  row.delivery_mode === "slots"
                    ? "بازه‌های ساعت قابل تعریف"
                    : `${row.min_days} تا ${row.max_days} روز ${row.day_basis === "calendar" ? "تقویمی" : "کاری"} پس از تأمین`,
              },
              {
                title: "ظرفیت روزانه",
                render: (row) =>
                  row.daily_capacity == null
                    ? "تعیین نشده"
                    : row.daily_capacity.toLocaleString("fa-IR"),
              },
              {
                title: "پوشش",
                render: (row) => (
                  <>
                    {row.all_cities
                      ? "همه شهرهای فعال"
                      : `${row.city_ids?.length ?? 0} شهر`}{" "}
                    ·{" "}
                    {row.category_mode === "selected"
                      ? "دسته‌های انتخابی"
                      : "همه دسته‌ها با استثناها"}
                    {!!row.unresolved_cities?.length && (
                      <small>شهر قدیمی نیازمند اصلاح است.</small>
                    )}
                  </>
                ),
              },
              {
                title: "عملیات",
                render: (row) => (
                  <>
                    <Button
                      onClick={() => setEditing({ method: row, create: false })}
                    >
                      {directory.can_write ? "ویرایش" : "مشاهده"}
                    </Button>
                    {can("stores.shipping.read", "management") && (
                      <Button
                        onClick={() => setStoreMethod(row)}
                        aria-label={"تنظیم فروشگاه‌ها برای " + row.name}
                      >
                        تنظیم فروشگاه‌ها
                      </Button>
                    )}
                  </>
                ),
              },
            ]}
          />
          {!directory.rows.length && (
            <Alert severity="info">روشی با این فیلترها پیدا نشد.</Alert>
          )}
        </>
      )}
      {storeMethod && (
        <Dialog
          open
          fullWidth
          maxWidth="xl"
          onClose={() => !storeBusy && setStoreMethod(undefined)}
        >
          <DialogTitle>{storeMethod.name} · تنظیم فروشگاه‌ها</DialogTitle>
          <DialogContent>
            <StoreShippingPanel
              key={storeMethod.id}
              methodID={storeMethod.id}
              allowBulk
              onBusyChange={setStoreBusy}
            />
          </DialogContent>
          <DialogActions>
            <Button
              disabled={storeBusy}
              onClick={() => setStoreMethod(undefined)}
            >
              بستن
            </Button>
          </DialogActions>
        </Dialog>
      )}
      {editing && (
        <MethodEditor
          key={editing.method.id}
          initial={editing.method}
          create={editing.create}
          allowed={directory?.can_write ?? false}
          geography={directory?.geography ?? []}
          close={() => setEditing(undefined)}
          save={async (input, key) => {
            await request<Method>(
              editing.create ? "" : "/" + encodeURIComponent(input.id),
              editing.create ? "POST" : "PUT",
              input,
              key,
            );
            setEditing(undefined);
            setNotice("روش حمل ذخیره شد.");
            await client.invalidateQueries({ queryKey: ["shipping-methods"] });
          }}
          reload={async () => {
            const next = await request<Directory>(
              "?query=" + encodeURIComponent(editing.method.id),
            );
            const row = next.rows.find((row) => row.id === editing.method.id);
            if (!row) throw new Error("missing method");
            return row;
          }}
        />
      )}
    </div>
  );
}

export function MethodEditor({
  initial,
  create,
  allowed,
  geography,
  close,
  save,
  reload,
  custom = false,
  scopeLabel,
  bulk = false,
}: {
  initial: Method;
  create: boolean;
  allowed: boolean;
  geography: Province[];
  close: () => void;
  save: (body: Method, key: string) => Promise<void>;
  reload: () => Promise<Method>;
  custom?: boolean;
  scopeLabel?: string;
  bulk?: boolean;
}) {
  const { state } = useWorkspace();
  const categories = state.categories ?? [];
  const [draft, setDraft] = useState(initial),
    [key, setKey] = useState(crypto.randomUUID()),
    [busy, setBusy] = useState(false),
    [error, setError] = useState(""),
    [fields, setFields] = useState<Record<string, string>>({}),
    [stale, setStale] = useState(false),
    [category, setCategory] = useState(""),
    [citySearch, setCitySearch] = useState("");
  const [batchConfirmed, setBatchConfirmed] = useState(false);
  function patch(value: Partial<Method>) {
    setDraft({ ...draft, ...value });
    setKey(crypto.randomUUID());
    setFields({});
  }
  const dayValid =
    Number.isInteger(draft.min_days) &&
    Number.isInteger(draft.max_days) &&
    draft.min_days >= 0 &&
    draft.max_days >= draft.min_days &&
    draft.max_days <= 60;
  const capacityValid =
    draft.daily_capacity != null &&
    Number.isInteger(draft.daily_capacity) &&
    draft.daily_capacity >= 0 &&
    draft.daily_capacity <= 1000000;
  const valid =
    (draft.delivery_mode !== "slots" || !scheduleError(draft.schedule)) &&
    (!bulk || batchConfirmed) &&
    !!draft.name.trim() &&
    dayValid &&
    capacityValid &&
    (!draft.active || draft.all_cities || !!draft.city_ids?.length) &&
    (!draft.active ||
      draft.category_mode !== "selected" ||
      !!draft.allowed_categories?.length);
  async function submit() {
    if (!valid || busy || !allowed || stale) return;
    setBusy(true);
    setError("");
    setFields({});
    try {
      const { unresolved_cities: _legacy, ...input } = draft;
      await save(
        {
          ...input,
          cities: [],
          city_ids: draft.all_cities ? [] : draft.city_ids,
        },
        key,
      );
    } catch (error) {
      if (error instanceof MethodError) {
        setFields(error.fields);
        setStale(error.status === 409);
        setError(
          error.status === 409
            ? "نسخهٔ روش تغییر کرده؛ پیش‌نویس حفظ شده است. نسخهٔ تازه را دریافت کنید."
            : Object.values(error.fields).join(" ") || error.message,
        );
      } else
        setError("ذخیره انجام نشد؛ پیش‌نویس حفظ شده است. دوباره تلاش کنید.");
    } finally {
      setBusy(false);
    }
  }
  const list = (field: "allowed_categories" | "denied_categories") => (
    <MultiChoiceField
      label={
        field === "allowed_categories" ? "دسته‌های مجاز" : "دسته‌های ممنوع"
      }
      value={draft[field] ?? []}
      options={categories.map((c) => ({
        id: c.id,
        name: categoryPath(categories, c.id),
      }))}
      disabled={busy || !allowed}
      onChange={(ids) => patch({ [field]: ids })}
    />
  );
  return (
    <Dialog open fullWidth maxWidth="xl" onClose={() => !busy && close()}>
      <DialogTitle>
        {custom
          ? `پیش‌نویس اختصاصی ${draft.name}`
          : create
            ? "روش حمل جدید"
            : `روش حمل ${draft.name}`}
      </DialogTitle>
      <DialogContent className="shipping-methods-page">
        {custom && (
          <Alert severity="info">
            نوع دریافت هزینه از روش عمومی ارث می‌رسد و فقط‌خواندنی است. پوشش
            شهر/دسته فقط می‌تواند محدودتر شود. ذخیره، پیش‌نویس می‌سازد؛ انتشار
            پس از ارسال و تأیید پلتفرم است. {scopeLabel}
          </Alert>
        )}
        {bulk && (
          <FormControlLabel
            label="ساخت پیش‌نویس برای همهٔ فروشگاه‌های نمایش‌داده‌شده را تأیید می‌کنم"
            control={
              <Checkbox
                checked={batchConfirmed}
                onChange={(e) => setBatchConfirmed(e.target.checked)}
                disabled={busy}
              />
            }
          />
        )}
        {!allowed && (
          <Alert severity="info">فقط مشاهده؛ دسترسی ویرایش ندارید.</Alert>
        )}
        {error && <Alert severity="error">{error}</Alert>}
        {stale && (
          <Button
            disabled={busy}
            onClick={() => {
              setBusy(true);
              void reload()
                .then((next) => {
                  setDraft(next);
                  setKey(crypto.randomUUID());
                  setStale(false);
                  setError("");
                  setFields({});
                })
                .catch(() =>
                  setError(
                    "دریافت نسخهٔ تازه انجام نشد؛ پیش‌نویس حفظ شده است.",
                  ),
                )
                .finally(() => setBusy(false));
            }}
          >
            دریافت نسخهٔ تازه (حذف پیش‌نویس)
          </Button>
        )}
        {!!draft.unresolved_cities?.length && (
          <Alert severity="warning">
            نام‌های شهر قدیمی به شناسه یکتا نگاشت نشدند:{" "}
            {draft.unresolved_cities.join("، ")}. پوشش را از فهرست شهرها دوباره
            انتخاب کنید.
          </Alert>
        )}
        {busy && <LinearProgress aria-label="ثبت روش حمل" />}
        <fieldset disabled={busy || !allowed} className="axis-inputs">
          <TextField
            fullWidth
            label="نام روش حمل"
            disabled={custom}
            value={draft.name}
            error={!!fields.name}
            helperText={fields.name}
            onChange={(event) => patch({ name: event.target.value })}
          />
          <FormControlLabel
            label="فعال"
            control={
              <Checkbox
                disabled={custom}
                checked={draft.active}
                onChange={(event) => patch({ active: event.target.checked })}
              />
            }
          />
          <div className="toolbar" role="group" aria-label="شیوه دریافت هزینه">
            <OptionSelect
              label={"شیوه دریافت هزینه"}
              multiple={false}
              options={(
                [
                  ["prepaid", "پیش‌کرایه"],
                  ["collect", "پس‌کرایه"],
                ] as const
              ).map(([value, label]) => ({
                id: String(value),
                name: String("" + label + ""),
                selected: draft.collection_mode === value,
                disabled: custom,
                onSelect: () =>
                  patch({
                    collection_mode: value,
                    fee_toman:
                      value === "collect" ? "" : draft.fee_toman || "0",
                  }),
              }))}
            />
          </div>
          {draft.collection_mode === "collect" ? (
            <Alert severity="info">
              هزینه را حمل‌کننده هنگام تحویل دریافت می‌کند؛ مبلغ آن در چریوا
              وصول نمی‌شود و این روش رایگان نیست.
            </Alert>
          ) : (
            <TextField
              label="هزینه هر مرسوله (تومان)"
              value={draft.fee_toman}
              error={!!fields.fee_toman}
              helperText={fields.fee_toman || "صفر یعنی پیش‌کرایه رایگان."}
              onChange={(event) => patch({ fee_toman: event.target.value })}
            />
          )}
          <div className="toolbar" role="group" aria-label="نوع زمان تحویل">
            <OptionSelect
              label={"نوع زمان تحویل"}
              multiple={false}
              options={(
                [
                  ["range", "بازهٔ روز"],
                  ["slots", "بازهٔ ساعت"],
                ] as const
              ).map(([value, label]) => ({
                id: String(value),
                name: String("" + label + ""),
                selected: draft.delivery_mode === value,
                disabled: false,
                onSelect: () =>
                  patch({
                    delivery_mode: value,
                    ...(value === "slots"
                      ? {
                          min_days: 0,
                          max_days: 0,
                          schedule: draft.schedule ?? { windows: [], days: [] },
                        }
                      : {}),
                  }),
              }))}
            />
          </div>
          {draft.delivery_mode === "range" ? (
            <>
              <div className="commerce-grid">
                <TextField
                  type="number"
                  label="حداقل مدت حمل"
                  value={draft.min_days}
                  inputProps={{ min: 0, max: 60 }}
                  error={!!fields.min_days || !dayValid}
                  onChange={(event) =>
                    patch({ min_days: Number(event.target.value) })
                  }
                />
                <TextField
                  type="number"
                  label="حداکثر مدت حمل"
                  value={draft.max_days}
                  inputProps={{ min: 0, max: 60 }}
                  error={!dayValid}
                  onChange={(event) =>
                    patch({ max_days: Number(event.target.value) })
                  }
                />
              </div>
              <div className="toolbar" role="group" aria-label="واحد مدت حمل">
                <OptionSelect
                  label={"واحد مدت حمل"}
                  multiple={false}
                  options={(
                    [
                      ["working", "روز کاری"],
                      ["calendar", "روز تقویمی"],
                    ] as const
                  ).map(([value, label]) => ({
                    id: String(value),
                    name: String("" + label + ""),
                    selected: draft.day_basis === value,
                    disabled: false,
                    onSelect: () => patch({ day_basis: value }),
                  }))}
                />
              </div>
            </>
          ) : (
            <>
              <div className="commerce-grid">
                <TextField
                  type="number"
                  label="حداقل فرصت حمل تا بازه (روز)"
                  value={draft.min_days}
                  inputProps={{ min: 0, max: 60 }}
                  onChange={(e) =>
                    patch({
                      min_days: Number(e.target.value),
                      max_days: Number(e.target.value),
                    })
                  }
                />
                <div
                  className="toolbar"
                  role="group"
                  aria-label="واحد فرصت حمل"
                >
                  <OptionSelect
                    label={"واحد فرصت حمل"}
                    multiple={false}
                    options={(
                      [
                        ["working", "روز کاری"],
                        ["calendar", "روز تقویمی"],
                      ] as const
                    ).map(([value, label]) => ({
                      id: String(value),
                      name: String(label),
                      selected: draft.day_basis === value,
                      disabled: false,
                      onSelect: () => patch({ day_basis: value }),
                    }))}
                  />
                </div>
              </div>
              <WeeklyScheduleEditor
                value={draft.schedule ?? { windows: [], days: [] }}
                onChange={(schedule) => patch({ schedule })}
                disabled={busy || !allowed}
                error={fields.schedule}
              />
            </>
          )}
          <TextField
            type="number"
            label="ظرفیت روزانه هر فروشگاه و روش"
            inputProps={{ min: 0, max: 1000000, step: 1 }}
            value={draft.daily_capacity ?? ""}
            error={
              !!fields.daily_capacity ||
              (draft.daily_capacity != null && !capacityValid)
            }
            helperText={
              fields.daily_capacity ||
              "هر مرسوله یک واحد؛ صفر یعنی ظرفیت بسته. مقدار را خودتان تعیین کنید."
            }
            onChange={(event) =>
              patch({
                daily_capacity:
                  event.target.value === "" ? null : Number(event.target.value),
              })
            }
          />
          <h3>پوشش شهر</h3>
          <ChoiceField
            label="پوشش شهر"
            value={draft.all_cities ? "all" : "selected"}
            options={[
              { id: "all", name: "همه شهرهای فعال" },
              { id: "selected", name: "شهرهای انتخابی" },
            ]}
            onChange={(mode) =>
              patch(
                mode === "all"
                  ? { all_cities: true, city_ids: [], cities: [] }
                  : { all_cities: false },
              )
            }
          />
          {draft.all_cities ? (
            <p>شهر فعال تازه نیز خودکار پوشش داده می‌شود.</p>
          ) : (
            <MultiChoiceField
              label="شهرهای تحت پوشش"
              value={draft.city_ids ?? []}
              options={geography.flatMap((p) =>
                p.cities.map((c) => ({
                  id: c.id,
                  name: p.name + " / " + c.name,
                })),
              )}
              disabled={busy || !allowed}
              onChange={(city_ids) => patch({ city_ids })}
            />
          )}
          {!draft.all_cities && <></>}
          <h3>پوشش دسته</h3>
          <div className="toolbar" role="group" aria-label="حالت پوشش دسته">
            <OptionSelect
              label={"حالت پوشش دسته"}
              multiple={false}
              options={(
                [
                  ["all", "همهٔ دسته‌ها"],
                  ["selected", "فقط دسته‌های انتخابی"],
                ] as const
              ).map(([value, label]) => ({
                id: String(value),
                name: String("" + label + ""),
                selected: draft.category_mode === value,
                disabled: false,
                onSelect: () => patch({ category_mode: value }),
              }))}
            />
          </div>
          <p>زیرشاخه‌ها نیز تابع انتخاب‌اند؛ ممنوعیت بر مجازبودن مقدم است.</p>
          <CategoryTree
            categories={categories}
            value={category}
            onChange={setCategory}
            allowRoot={false}
            disabled={busy || !allowed}
          />
          <div className="toolbar">
            <CreateButton
              disabled={!category || draft.category_mode !== "selected"}
              onClick={() =>
                patch({
                  allowed_categories: Array.from(
                    new Set([...(draft.allowed_categories ?? []), category]),
                  ),
                })
              }
            >
              افزودن به مجازها
            </CreateButton>
            <CreateButton
              disabled={!category}
              onClick={() =>
                patch({
                  denied_categories: Array.from(
                    new Set([...(draft.denied_categories ?? []), category]),
                  ),
                })
              }
            >
              افزودن به ممنوع‌ها
            </CreateButton>
          </div>
          {draft.category_mode === "selected" && (
            <>
              <h4>دسته‌های مجاز</h4>
              {list("allowed_categories")}
            </>
          )}
          <h4>دسته‌های ممنوع</h4>
          {list("denied_categories")}
        </fieldset>
      </DialogContent>
      <DialogActions>
        <Button disabled={busy} onClick={close}>
          بستن
        </Button>
        {allowed && (
          <Button
            variant="contained"
            disabled={busy || !valid || stale}
            onClick={() => void submit()}
          >
            {custom ? "ذخیرهٔ پیش‌نویس اختصاصی" : "ذخیرهٔ روش حمل"}
          </Button>
        )}
      </DialogActions>
    </Dialog>
  );
}
