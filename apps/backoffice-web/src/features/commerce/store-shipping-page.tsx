import { ChoiceField } from "../../shared/option-select";
import { FilterPanel } from "../../shared/filter-panel";
import { useEffect, useState } from "react";
import {
  Alert,
  Button,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  LinearProgress,
  TextField,
} from "@mui/material";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { useWorkspace } from "../../shared/workspace";
import { ShippingCapacityDialog } from "./shipping-capacity-dialog";
import { DataTable } from "../../shared/components";
import {
  MethodEditor,
  MethodError,
  type Method,
} from "./shipping-methods-page";
import type { Province } from "../../../../../packages/commerce-client/types";

type Revision = {
  number: number;
  status: string;
  base: Method;
  method: Method;
  reason: string;
  at: string;
};
type Binding = {
  base: Method;
  enabled: boolean;
  mode: string;
  effective?: Method;
  live?: Revision;
  draft?: Revision;
  history: Revision[];
  base_changed: boolean;
};
type StoreRow = {
  id: string;
  name: string;
  version: number;
  can_write: boolean;
  can_review: boolean;
  methods: Binding[];
};
export type StoreShippingDirectory = {
  stores: StoreRow[];
  geography: Province[];
};
type Target = { id: string; version: number };
type Command = {
  action: string;
  method_id: string;
  base_version: number;
  all: boolean;
  stores: Target[];
  reason: string;
  method?: Method;
};
const statuses: Record<string, string> = {
  public: "استفاده از عمومی",
  draft: "پیش‌نویس اختصاصی",
  submitted: "در انتظار بررسی",
  approved: "اختصاصی تأییدشده",
  rejected: "اصلاحات ردشده",
  disabled: "غیرفعال",
  withdrawn: "کنارگذاشته‌شده",
};
const stateOf = (row: Binding) =>
  !row.enabled
    ? "disabled"
    : (row.draft?.status ?? (row.mode === "public" ? "public" : "approved"));
const initialMethod = (row: Binding, bulk = false): Method => {
  const method = bulk
    ? row.base
    : (row.draft?.method ?? row.live?.method ?? row.base);
  return {
    ...method,
    id: row.base.id,
    name: row.base.name,
    active: true,
    collection_mode: row.base.collection_mode,
    store_ids: [],
    fee_toman: row.base.collection_mode === "collect" ? "" : method.fee_toman,
  };
};

export function useStoreShippingDirectory(enabled = true) {
  const { authToken, surface } = useWorkspace();
  const queryKey = ["store-shipping", surface, authToken];
  const root =
    (import.meta.env.VITE_API_BASE_URL ?? "") +
    "/" +
    surface +
    "/stores/shipping";
  async function request(
    body?: Command,
    key?: string,
    signal?: AbortSignal,
  ): Promise<StoreShippingDirectory> {
    const response = await fetch(root, {
      method: body ? "POST" : "GET",
      signal,
      cache: "no-store",
      headers: {
        Authorization: "Bearer " + authToken,
        ...(body
          ? {
              "Content-Type": "application/json",
              "Idempotency-Key": key ?? crypto.randomUUID(),
            }
          : {}),
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
    queryKey,
    queryFn: ({ signal }) => request(undefined, undefined, signal),
    enabled,
    refetchOnWindowFocus: false,
  });
  return { query, request, queryKey };
}

export function StoreShippingPanel({
  storeID,
  methodID,
  allowBulk = false,
  enabled = true,
  onBusyChange,
}: {
  storeID?: string;
  methodID?: string;
  allowBulk?: boolean;
  enabled?: boolean;
  onBusyChange?: (busy: boolean) => void;
}) {
  const { surface } = useWorkspace();
  const client = useQueryClient();
  const { query, request, queryKey } = useStoreShippingDirectory(enabled);
  const directory = query.data;
  const [selected, setSelected] = useState(""),
    [all, setAll] = useState(false),
    [notice, setNotice] = useState("");
  const [methodSearch, setMethodSearch] = useState(""),
    [status, setStatus] = useState(""),
    [draftMethodSearch, setDraftMethodSearch] = useState(""),
    [draftStatus, setDraftStatus] = useState("");
  const [editing, setEditing] = useState<{
    row: Binding;
    targets: Target[];
    all: boolean;
  }>();
  const [decision, setDecision] = useState<{
    row: Binding;
    command: Command;
    scopeLabel: string;
  }>();
  const [capacity, setCapacity] = useState<Binding>();
  const busy = !!editing || !!decision || !!capacity;
  useEffect(() => {
    onBusyChange?.(busy);
    return () => onBusyChange?.(false);
  }, [busy, onBusyChange]);
  const allAllowed =
    allowBulk &&
    !storeID &&
    surface === "backoffice" &&
    !!directory?.stores.length &&
    directory.stores.every((s) => s.can_write);
  const bulk = all && allAllowed;
  const store =
    directory?.stores.find((s) => s.id === (storeID ?? selected)) ??
    (bulk ? directory?.stores[0] : undefined);
  const scopeLabel = bulk
    ? "همهٔ " + directory!.stores.length.toLocaleString("fa-IR") + " فروشگاه"
    : (store?.name ?? "");
  const targets = (): Target[] =>
    (bulk ? directory!.stores : store ? [store] : []).map((s) => ({
      id: s.id,
      version: s.version,
    }));
  function action(row: Binding, action: string) {
    const review = action === "approve" || action === "reject";
    setDecision({
      row,
      scopeLabel,
      command: {
        action,
        method_id: row.base.id,
        base_version: row.base.version,
        all: !review && bulk,
        stores:
          review && store
            ? [{ id: store.id, version: store.version }]
            : targets(),
        reason: "",
      },
    });
  }
  async function refresh(next: StoreShippingDirectory) {
    client.setQueryData(queryKey, next);
    await client.invalidateQueries({ queryKey: ["checkout-prepare"] });
  }
  return (
    <section
      className="store-shipping-panel"
      aria-label="تنظیم روش‌های حمل فروشگاه"
    >
      <div className="toolbar">
        <Button
          disabled={busy || query.isFetching}
          onClick={() => void query.refetch()}
        >
          به‌روزرسانی روش‌های حمل
        </Button>
      </div>
      {query.isPending && enabled && (
        <LinearProgress aria-label="دریافت روش‌های حمل فروشگاه" />
      )}
      {query.error && (
        <Alert severity="error">
          دریافت تنظیمات حمل انجام نشد. {query.error.message}
          <Button onClick={() => void query.refetch()}>تلاش دوباره</Button>
        </Alert>
      )}
      {notice && <Alert severity="success">{notice}</Alert>}
      {directory && !directory.stores.length && (
        <Alert severity="info">
          فروشگاهی با مجوز مشاهدهٔ حمل در دسترس نیست.
        </Alert>
      )}
      {!storeID && !!directory?.stores.length && (
        <div className="admin-shipping-scope">
          <ChoiceField
            label="فروشگاه موردنظر"
            value={selected}
            clearLabel="انتخاب فروشگاه"
            options={directory.stores.map((s) => ({ id: s.id, name: s.name }))}
            disabled={busy || bulk}
            onChange={(id) => {
              setSelected(id);
              setNotice("");
            }}
          />
          {allAllowed && (
            <ChoiceField
              label="محدوده اعمال روش حمل"
              value={bulk ? "all" : "one"}
              options={[
                { id: "one", name: "یک فروشگاه" },
                { id: "all", name: "همه فروشگاه‌ها" },
              ]}
              disabled={busy}
              onChange={(value) => {
                setAll(value === "all");
                setNotice("");
              }}
            />
          )}
        </div>
      )}
      {directory && !!directory.stores.length && !store && (
        <Alert severity={storeID ? "warning" : "info"}>
          {storeID
            ? "مجوز مشاهدهٔ حمل این فروشگاه را ندارید یا فروشگاه در دسترس نیست."
            : "یک فروشگاه یا محدودهٔ همهٔ فروشگاه‌ها را انتخاب کنید."}
        </Alert>
      )}
      {store && (
        <>
          <Alert severity={bulk ? "warning" : "info"}>
            {bulk
              ? "تغییرات مشترک برای همهٔ فروشگاه‌ها اعمال می‌شوند؛ بررسی اصلاحات و ظرفیت روز برای هر فروشگاه جداست."
              : "برای هر روش، تنظیمات عمومی را بپذیرید یا اصلاحات اختصاصی پیشنهاد دهید. اصلاحات پس از تأیید قابل استفاده می‌شوند."}
          </Alert>
          {!methodID && (
            <FilterPanel
              onSearch={() => {
                setMethodSearch(draftMethodSearch);
                setStatus(draftStatus);
              }}
            >
              <div className="admin-shipping-filters">
                <TextField
                  label="جست‌وجوی روش حمل"
                  value={draftMethodSearch}
                  onChange={(e) => setDraftMethodSearch(e.target.value)}
                />
                <ChoiceField
                  label="وضعیت روش حمل"
                  value={draftStatus}
                  clearLabel="همه وضعیت‌ها"
                  options={Object.entries(statuses)
                    .filter(([id]) => id !== "withdrawn")
                    .map(([id, name]) => ({ id, name }))}
                  onChange={setDraftStatus}
                />
              </div>
            </FilterPanel>
          )}
          <DataTable
            key={store.id + ":" + status + ":" + methodSearch + ":" + bulk}
            searchable={false}
            rows={store.methods
              .filter(
                (row) =>
                  (!methodID || row.base.id === methodID) &&
                  (bulk || !status || stateOf(row) === status) &&
                  row.base.name.includes(methodSearch),
              )
              .map((row) => ({ ...row, id: row.base.id }))}
            columns={[
              {
                title: "روش حمل",
                render: (row) => (
                  <>
                    <b>{row.base.name}</b>
                    <small>
                      {row.base.active ? "عمومی فعال" : "روش عمومی غیرفعال"}
                    </small>
                  </>
                ),
              },
              {
                title: "تنظیمات مورد استفاده",
                render: (row) =>
                  bulk ? (
                    "اعمال مشترک؛ وضعیت هر فروشگاه مستقل است"
                  ) : (
                    <>
                      {statuses[stateOf(row)]}
                      <small>
                        {row.mode === "public"
                          ? "ارث‌بری از تنظیمات عمومی جاری"
                          : row.live
                            ? "تنظیمات اختصاصی تأییدشده فعال است"
                            : "تا تأیید، تنظیمات اختصاصی قابل استفاده نیست"}
                      </small>
                      {row.base_changed && (
                        <small>
                          روش عمومی تغییر کرده؛ اصلاحات نیاز به بازبینی دارد.
                        </small>
                      )}
                    </>
                  ),
              },
              {
                title: "هزینه و زمان حمل",
                render: (row) =>
                  bulk ? (
                    "برای مشاهده، یک فروشگاه انتخاب کنید"
                  ) : row.effective ? (
                    <>
                      {row.effective.collection_mode === "collect"
                        ? "پس‌کرایه؛ پرداخت به حمل‌کننده"
                        : row.effective.fee_toman + " تومان پیش‌کرایه"}
                      <small>
                        {row.effective.min_days} تا {row.effective.max_days} روز{" "}
                        {row.effective.day_basis === "calendar"
                          ? "تقویمی"
                          : "کاری"}
                      </small>
                    </>
                  ) : (
                    "در خرید ارائه نمی‌شود"
                  ),
              },
              {
                title: "عملیات",
                render: (row) => {
                  const submitReady = bulk
                    ? directory!.stores.every((s) => {
                        const binding = s.methods.find(
                          (m) => m.base.id === row.base.id,
                        );
                        return (
                          binding?.draft?.status === "draft" &&
                          !binding.base_changed
                        );
                      })
                    : row.draft?.status === "draft";
                  return (
                    <div className="table-row-actions">
                      {(bulk || store.can_write) && (
                        <>
                          <Button
                            disabled={
                              busy ||
                              !row.base.active ||
                              (!bulk &&
                                row.enabled &&
                                row.mode === "public" &&
                                !row.draft)
                            }
                            onClick={() => action(row, "public")}
                          >
                            استفاده از تنظیمات عمومی
                          </Button>
                          <Button
                            disabled={busy || (!bulk && !row.enabled)}
                            onClick={() => action(row, "disable")}
                          >
                            غیرفعال‌کردن
                          </Button>
                          <Button
                            disabled={busy || !row.base.active}
                            onClick={() =>
                              setEditing({ row, targets: targets(), all: bulk })
                            }
                          >
                            اصلاح تنظیمات
                          </Button>
                          {submitReady && (
                            <Button
                              disabled={busy || (!bulk && row.base_changed)}
                              onClick={() => action(row, "submit")}
                            >
                              ارسال اصلاحات برای بررسی
                            </Button>
                          )}
                        </>
                      )}
                      {!bulk && (row.draft || row.live) && (
                        <Button
                          disabled={busy}
                          onClick={() => action(row, "view")}
                        >
                          مشاهده اصلاحات
                        </Button>
                      )}
                      {!bulk && (
                        <Button
                          disabled={busy}
                          onClick={() => setCapacity(row)}
                        >
                          ظرفیت روز
                        </Button>
                      )}
                      {!bulk &&
                        store.can_review &&
                        row.draft?.status === "submitted" && (
                          <>
                            <Button
                              disabled={
                                busy || row.base_changed || !row.base.active
                              }
                              onClick={() => action(row, "approve")}
                            >
                              بررسی و تأیید
                            </Button>
                            <Button
                              disabled={busy}
                              color="error"
                              onClick={() => action(row, "reject")}
                            >
                              بررسی و رد
                            </Button>
                          </>
                        )}
                    </div>
                  );
                },
              },
            ]}
          />
        </>
      )}
      {editing && (
        <MethodEditor
          key={editing.row.base.id}
          initial={initialMethod(editing.row, editing.all)}
          create={false}
          custom
          allowed={store?.can_write ?? false}
          geography={directory?.geography ?? []}
          bulk={editing.targets.length > 1}
          scopeLabel={scopeLabel}
          close={() => setEditing(undefined)}
          save={async (method, key) => {
            const next = await request(
              {
                action: "draft",
                method_id: editing.row.base.id,
                base_version: editing.row.base.version,
                all: editing.all,
                stores: editing.targets,
                reason: "",
                method,
              },
              key,
            );
            await refresh(next);
            setEditing(undefined);
            setNotice(
              "اصلاحات ذخیره شد؛ برای انتشار آن را برای بررسی ارسال کنید.",
            );
          }}
          reload={async () => {
            const next = await request();
            await refresh(next);
            const freshStore = next.stores.find((s) => s.id === store?.id);
            const fresh = freshStore?.methods.find(
              (row) => row.base.id === editing.row.base.id,
            );
            if (!fresh || !freshStore)
              throw Error("تنظیمات دیگر در دسترس نیست");
            setEditing({
              ...editing,
              row: fresh,
              targets: (editing.all ? next.stores : [freshStore]).map((s) => ({
                id: s.id,
                version: s.version,
              })),
            });
            return initialMethod(fresh, editing.all);
          }}
        />
      )}
      {decision && (
        <ShippingDecision
          geography={directory?.geography ?? []}
          row={decision.row}
          command={decision.command}
          scopeLabel={decision.scopeLabel}
          close={() => setDecision(undefined)}
          save={async (command, key) => {
            const next = await request(command, key);
            await refresh(next);
            setDecision(undefined);
            setNotice("تنظیم روش حمل ثبت شد.");
          }}
        />
      )}
      {capacity && store && (
        <ShippingCapacityDialog
          storeID={store.id}
          methodID={capacity.base.id}
          name={store.name + " / " + capacity.base.name}
          close={() => setCapacity(undefined)}
        />
      )}
    </section>
  );
}

function ShippingDecision({
  row,
  command,
  save,
  close,
  geography,
  scopeLabel,
}: {
  row: Binding;
  command: Command;
  save: (command: Command, key: string) => Promise<void>;
  close: () => void;
  geography: Province[];
  scopeLabel: string;
}) {
  const [reason, setReason] = useState(""),
    [busy, setBusy] = useState(false),
    [error, setError] = useState("");
  const [key, setKey] = useState(crypto.randomUUID());
  const review = command.action === "approve" || command.action === "reject";
  const view = command.action === "view";
  const showDiff = view || review;
  const actionLabel = (
    {
      public: "استفاده از تنظیمات عمومی",
      disable: "غیرفعال‌کردن این روش",
      submit: "ارسال اصلاحات برای بررسی",
      approve: "تأیید و انتشار اصلاحات",
      reject: "رد اصلاحات",
    } as Record<string, string>
  )[command.action];
  const fields: [keyof Method, string][] = [
    ["schedule", "برنامهٔ روزها و ساعت"],
    ["collection_mode", "دریافت هزینه"],
    ["delivery_mode", "نوع تحویل"],
    ["fee_toman", "پیش‌کرایه (تومان)"],
    ["min_days", "حداقل مدت حمل"],
    ["max_days", "حداکثر مدت حمل"],
    ["day_basis", "واحد مدت"],
    ["daily_capacity", "سقف روزانه"],
    ["all_cities", "همهٔ شهرها"],
    ["city_ids", "شهرهای انتخابی"],
    ["category_mode", "محدودهٔ دسته"],
    ["allowed_categories", "دسته‌های مجاز"],
    ["denied_categories", "دسته‌های ممنوع"],
  ];
  const names: Record<string, string> = {
    prepaid: "پیش‌کرایه",
    collect: "پس‌کرایه",
    range: "بازهٔ روز",
    slots: "بازهٔ ساعت",
    working: "روز کاری",
    calendar: "روز تقویمی",
    all: "همه",
    selected: "انتخابی",
  };
  const { state } = useWorkspace();
  const display = (value: unknown) => {
    if (
      value &&
      typeof value === "object" &&
      "windows" in value &&
      "days" in value
    ) {
      const schedule = value as NonNullable<Method["schedule"]>;
      const days = [
        "یکشنبه",
        "دوشنبه",
        "سه‌شنبه",
        "چهارشنبه",
        "پنج‌شنبه",
        "جمعه",
        "شنبه",
      ];
      return (
        schedule.days
          .map(
            (day) =>
              `${days[day.weekday]} ${day.active ? "باز" : "بسته"}: ${day.slots
                .map((slot) => {
                  const window = schedule.windows.find(
                    (w) => w.id === slot.window_id,
                  );
                  return `${window?.start ?? "—"} تا ${window?.end ?? "—"} / ظرفیت ${slot.capacity ?? "—"}`;
                })
                .join("، ")}`,
          )
          .join("؛ ") || "برنامهٔ خالی"
      );
    }
    if (Array.isArray(value))
      return (
        value
          .map(
            (v) =>
              (state.categories ?? []).find((c) => c.id === v)?.name ??
              geography.flatMap((p) => p.cities).find((c) => c.id === v)
                ?.name ??
              v,
          )
          .join("، ") || "—"
      );
    if (typeof value === "boolean") return value ? "بله" : "خیر";
    return value == null || value === ""
      ? "—"
      : (names[String(value)] ?? String(value));
  };
  return (
    <Dialog
      open
      fullWidth
      maxWidth={showDiff ? "lg" : "sm"}
      onClose={() => !busy && close()}
    >
      <DialogTitle>
        {showDiff ? "بررسی اصلاحات " + row.base.name : actionLabel}
      </DialogTitle>
      <DialogContent>
        {error && <Alert severity="error">{error}</Alert>}
        <p>
          <strong>{row.base.name}</strong> · {scopeLabel}
        </p>
        {showDiff && row.base_changed && (
          <Alert severity="warning">
            روش عمومی از زمان ساخت درخواست تغییر کرده است؛ پیش از تأیید باید
            اصلاحات بازبینی و دوباره ارسال شوند.
          </Alert>
        )}
        {!view && (
          <Alert severity={command.all ? "warning" : "info"}>
            {command.action === "public"
              ? "این روش با تنظیمات عمومی جاری استفاده می‌شود و درخواست اختصاصی جاری کنار گذاشته خواهد شد."
              : command.action === "disable"
                ? "این روش فقط در محدودهٔ انتخاب‌شده ارائه نمی‌شود؛ تعریف عمومی روش تغییر نمی‌کند."
                : command.action === "submit"
                  ? "اصلاحات برای بررسی ارسال می‌شوند؛ تنظیمات قابل استفادهٔ قبلی تا تأیید حفظ خواهند شد."
                  : "تصمیم با دلیل ثبت می‌شود؛ سفارش‌های قبلی تغییر نمی‌کنند."}
            {command.all && (
              <p>
                این عملیات برای همهٔ{" "}
                {command.stores.length.toLocaleString("fa-IR")} فروشگاه اعمال
                می‌شود.
              </p>
            )}
          </Alert>
        )}
        {showDiff && (
          <>
            <div className="store-shipping-diff">
              <table>
                <thead>
                  <tr>
                    <th>تنظیم</th>
                    <th>عمومی جاری</th>
                    <th>مبنای درخواست</th>
                    <th>نسخهٔ منتشرشده</th>
                    <th>درخواست جاری</th>
                  </tr>
                </thead>
                <tbody>
                  {fields.map(([field, label]) => (
                    <tr key={field}>
                      <th>{label}</th>
                      {[
                        row.base,
                        row.draft?.base,
                        row.live?.method,
                        row.draft?.method,
                      ].map((method, i) => (
                        <td key={i}>{method ? display(method[field]) : "—"}</td>
                      ))}
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
            {row.draft?.reason && <p>دلیل تصمیم اخیر: {row.draft.reason}</p>}
            <details>
              <summary>
                تاریخچه نسخه‌ها ({row.history.length.toLocaleString("fa-IR")})
              </summary>
              {row.history.map((r, i) => (
                <p key={i}>
                  نسخه {r.number.toLocaleString("fa-IR")} ·{" "}
                  {statuses[r.status] ?? r.status} ·{" "}
                  {new Date(r.at).toLocaleString("fa-IR")} · {r.reason}
                </p>
              ))}
            </details>
          </>
        )}
        {review && (
          <TextField
            fullWidth
            multiline
            required
            label="دلیل بررسی و تصمیم"
            value={reason}
            disabled={busy}
            onChange={(e) => {
              setReason(e.target.value);
              setKey(crypto.randomUUID());
            }}
            inputProps={{ maxLength: 1000 }}
          />
        )}
      </DialogContent>
      <DialogActions>
        <Button disabled={busy} onClick={close}>
          بستن
        </Button>
        {!view && (
          <Button
            variant="contained"
            disabled={busy || (review && !reason.trim())}
            onClick={() => {
              setBusy(true);
              setError("");
              void save({ ...command, reason }, key)
                .catch((error) =>
                  setError(
                    error instanceof MethodError && error.status === 409
                      ? "اطلاعات تغییر کرده است؛ پنجره را ببندید و آخرین تنظیمات را دریافت کنید."
                      : error.message,
                  ),
                )
                .finally(() => setBusy(false));
            }}
          >
            {busy ? "در حال ثبت…" : actionLabel}
          </Button>
        )}
      </DialogActions>
    </Dialog>
  );
}
