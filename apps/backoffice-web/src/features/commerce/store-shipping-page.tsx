import { ChoiceField, MultiChoiceField } from "../../shared/option-select";
import { FilterPanel } from "../../shared/filter-panel";
import { OptionSelect } from "../../shared/option-select";
import { useState } from "react";
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
import { DataTable, PageHeader } from "../../shared/components";
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
type Directory = { stores: StoreRow[]; geography: Province[] };
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
  public: "عمومی",
  draft: "پیش‌نویس",
  submitted: "در انتظار بررسی",
  approved: "تأییدشده",
  rejected: "ردشده",
  disabled: "غیرفعال",
  withdrawn: "کنارگذاشته‌شده",
};
const stateOf = (row: Binding) =>
  !row.enabled
    ? "disabled"
    : (row.draft?.status ?? (row.mode === "public" ? "public" : "approved"));
const initialMethod = (row: Binding): Method => ({
  ...(row.draft?.method ?? row.live?.method ?? row.base),
  id: row.base.id,
  name: row.base.name,
  active: true,
  collection_mode: row.base.collection_mode,
  store_ids: [],
  fee_toman:
    row.base.collection_mode === "collect"
      ? ""
      : (row.draft?.method ?? row.live?.method ?? row.base).fee_toman,
});

export function StoreShippingPage() {
  const { authToken, surface } = useWorkspace();
  const client = useQueryClient();
  const queryKey = ["store-shipping", surface, authToken];
  const root =
    (import.meta.env.VITE_API_BASE_URL ?? "") + `/${surface}/stores/shipping`;
  const [selected, setSelected] = useState(""),
    [search, setSearch] = useState(""),
    [methodSearch, setMethodSearch] = useState(""),
    [status, setStatus] = useState(""),
    [all, setAll] = useState(false),
    [notice, setNotice] = useState("");
  const [draftMethodSearch, setDraftMethodSearch] = useState("");
  const [draftStatus, setDraftStatus] = useState("");
  const [editing, setEditing] = useState<{
    row: Binding;
    targets: Target[];
    all: boolean;
  }>();
  const [decision, setDecision] = useState<{
    row: Binding;
    command: Command;
  }>();
  const [capacity, setCapacity] = useState<Binding>();
  async function request(body?: Command, key?: string): Promise<Directory> {
    const response = await fetch(root, {
      method: body ? "POST" : "GET",
      cache: "no-store",
      headers: {
        Authorization: `Bearer ${authToken}`,
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
    queryFn: () => request(),
    refetchOnWindowFocus: false,
  });
  const directory = query.data;
  const store =
    directory?.stores.find((s) => s.id === selected) ?? directory?.stores[0];
  const busy = !!editing || !!decision || !!capacity;
  const allAllowed =
    surface === "backoffice" &&
    !!directory?.stores.length &&
    directory.stores.every((s) => s.can_write);
  const targets = (): Target[] =>
    (all && allAllowed ? (directory?.stores ?? []) : store ? [store] : []).map(
      (s) => ({ id: s.id, version: s.version }),
    );
  function action(row: Binding, action: string) {
    const review = action === "approve" || action === "reject";
    setDecision({
      row,
      command: {
        action,
        method_id: row.base.id,
        base_version: row.base.version,
        all: !review && all && allAllowed,
        stores:
          review && store
            ? [{ id: store.id, version: store.version }]
            : targets(),
        reason: "",
      },
    });
  }
  async function refresh(next: Directory) {
    client.setQueryData(queryKey, next);
    await client.invalidateQueries({ queryKey: ["checkout-prepare"] });
  }
  return (
    <div className="store-shipping-page">
      <PageHeader
        title="روش‌های حمل فروشگاه"
        description="انتساب عمومی، پیش‌نویس اختصاصی و بررسی پلتفرم؛ فقط نسخهٔ تأییدشده در خرید استفاده می‌شود"
      />
      <Button
        disabled={busy || query.isFetching}
        onClick={() => void query.refetch()}
      >
        دریافت آخرین تنظیمات
      </Button>
      {query.isPending && <LinearProgress aria-label="دریافت حمل فروشگاه" />}
      {query.error && (
        <Alert severity="error">
          دریافت تنظیمات حمل انجام نشد. {query.error.message}{" "}
          <Button onClick={() => void query.refetch()}>تلاش دوباره</Button>
        </Alert>
      )}
      {notice && <Alert severity="success">{notice}</Alert>}
      {directory && !directory.stores.length && (
        <Alert severity="info">
          فروشگاهی با مجوز مشاهدهٔ حمل در دسترس نیست.
        </Alert>
      )}
      {store && (
        <>
          <FilterPanel
            onSearch={() => {
              setMethodSearch(draftMethodSearch);
              setStatus(draftStatus);
            }}
          >
            <section className="catalog-filters">
              <h2 className="filter-panel-title">فروشگاه و فیلتر روش‌ها</h2>
              <TextField
                label="جست‌وجوی فروشگاه"
                size="small"
                value={search}
                disabled={busy}
                onChange={(e) => setSearch(e.target.value)}
              />
              <TextField
                label="جست‌وجوی روش حمل"
                size="small"
                value={draftMethodSearch}
                onChange={(e) => setDraftMethodSearch(e.target.value)}
              />
              <div
                role="group"
                aria-label="انتخاب فروشگاه حمل"
                className="store-shipping-stores"
              >
                <OptionSelect
                  label={"انتخاب فروشگاه حمل"}
                  multiple={false}
                  options={directory?.stores
                    .filter((s) => s.name.includes(search))
                    .map((s) => ({
                      id: String(s.id),
                      name: String("" + s.name + ""),
                      selected: store.id === s.id,
                      disabled: busy,
                      onSelect: () => setSelected(s.id),
                    }))}
                />
              </div>
              {allAllowed && (
                <ChoiceField
                  label="محدوده تغییر حمل"
                  value={all ? "all" : "one"}
                  disabled={busy}
                  options={[
                    { id: "one", name: "همین فروشگاه" },
                    { id: "all", name: "همه فروشگاه‌ها" },
                  ]}
                  onChange={(value) => setAll(value === "all")}
                />
              )}
              {all && allAllowed && (
                <Alert severity="warning">
                  انتساب، غیرفعال‌سازی و ساخت/ارسال پیش‌نویس روی همهٔ فروشگاه‌ها
                  اعمال می‌شود. بررسی و تأیید هر نسخه جداگانه انجام می‌شود.
                </Alert>
              )}
              <div
                role="group"
                aria-label="وضعیت روش حمل فروشگاه"
                className="toolbar"
              >
                <OptionSelect
                  label={"وضعیت روش حمل فروشگاه"}
                  multiple={false}
                  options={[
                    ["", "همه"],
                    ...Object.entries(statuses).filter(
                      ([key]) => key !== "withdrawn",
                    ),
                  ].map(([key, label]) => ({
                    id: String(key),
                    name: String("" + label + ""),
                    selected: draftStatus === key,
                    disabled: false,
                    onSelect: () => setDraftStatus(key),
                  }))}
                />
              </div>
            </section>
          </FilterPanel>
          <h2>{store.name}</h2>
          <DataTable
            key={`${store.id}:${status}:${methodSearch}`}
            searchable={false}
            rows={store.methods
              .filter(
                (row) =>
                  (!status || stateOf(row) === status) &&
                  row.base.name.includes(methodSearch),
              )
              .map((row) => ({ ...row, id: row.base.id }))}
            columns={[
              {
                title: "روش عمومی",
                render: (row) => (
                  <>
                    <b>{row.base.name}</b>
                    <small>
                      {row.base.active ? "فعال" : "عمومی غیرفعال"} · نسخه{" "}
                      {row.base.version.toLocaleString("fa-IR")}
                    </small>
                  </>
                ),
              },
              {
                title: "وضعیت فروشگاه",
                render: (row) => (
                  <>
                    {statuses[stateOf(row)]}
                    <small>
                      {row.mode === "public"
                        ? "ارث‌بری تنظیم عمومی جاری"
                        : row.live
                          ? `نسخهٔ منتشرشده ${row.live.number.toLocaleString("fa-IR")}`
                          : "هنوز نسخهٔ منتشرشده ندارد"}
                    </small>
                    {row.base_changed && (
                      <small>
                        روش عمومی تغییر کرده؛ پیش‌نویس باید بازبینی شود.
                      </small>
                    )}
                  </>
                ),
              },
              {
                title: "هزینه و مدت مؤثر",
                render: (row) =>
                  row.effective ? (
                    <>
                      {row.effective.collection_mode === "collect"
                        ? "پس‌کرایه؛ دریافت توسط حمل‌کننده"
                        : `${row.effective.fee_toman} تومان پیش‌کرایه`}
                      <small>
                        {row.effective.min_days} تا {row.effective.max_days} روز{" "}
                        {row.effective.day_basis === "calendar"
                          ? "تقویمی"
                          : "کاری"}
                      </small>
                    </>
                  ) : (
                    "قابل ارائه در خرید نیست"
                  ),
              },
              {
                title: "عملیات",
                render: (row) => (
                  <div className="toolbar">
                    {store.can_write && (
                      <>
                        <Button
                          disabled={busy || !row.base.active}
                          onClick={() => action(row, "public")}
                        >
                          پذیرش عمومی
                        </Button>
                        <Button
                          disabled={busy || !row.enabled}
                          onClick={() => action(row, "disable")}
                        >
                          غیرفعال در فروشگاه
                        </Button>
                        <Button
                          disabled={busy || !row.base.active}
                          onClick={() =>
                            setEditing({
                              row,
                              targets: targets(),
                              all: all && allAllowed,
                            })
                          }
                        >
                          تنظیم اختصاصی
                        </Button>
                        {row.draft?.status === "draft" && (
                          <Button
                            disabled={busy || row.base_changed}
                            onClick={() => action(row, "submit")}
                          >
                            ارسال برای بررسی
                          </Button>
                        )}
                      </>
                    )}
                    <Button disabled={busy} onClick={() => action(row, "view")}>
                      جزئیات و تفاوت‌ها
                    </Button>
                    <Button disabled={busy} onClick={() => setCapacity(row)}>
                      ظرفیت روز
                    </Button>
                    {store.can_review && row.draft?.status === "submitted" && (
                      <>
                        <Button
                          disabled={
                            busy || row.base_changed || !row.base.active
                          }
                          onClick={() => action(row, "approve")}
                        >
                          تأیید نسخه
                        </Button>
                        <Button
                          disabled={busy}
                          color="error"
                          onClick={() => action(row, "reject")}
                        >
                          رد نسخه
                        </Button>
                      </>
                    )}
                  </div>
                ),
              },
            ]}
          />
        </>
      )}
      {editing && (
        <MethodEditor
          key={editing.row.base.id}
          initial={initialMethod(editing.row)}
          create={false}
          custom
          allowed={store?.can_write ?? false}
          geography={directory?.geography ?? []}
          bulk={editing.targets.length > 1}
          scopeLabel={
            editing.all
              ? `همهٔ ${editing.targets.length.toLocaleString("fa-IR")} فروشگاه`
              : store?.name
          }
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
              "پیش‌نویس اختصاصی ذخیره شد؛ برای انتشار، درخواست بررسی بفرستید.",
            );
          }}
          reload={async () => {
            const next = await request();
            await refresh(next);
            const freshStore = next.stores.find((s) => s.id === store?.id);
            const fresh = freshStore?.methods.find(
              (row) => row.base.id === editing.row.base.id,
            );
            if (!fresh || !freshStore) throw Error();
            setEditing({
              ...editing,
              row: fresh,
              targets: (editing.all ? next.stores : [freshStore]).map((s) => ({
                id: s.id,
                version: s.version,
              })),
            });
            return initialMethod(fresh);
          }}
        />
      )}
      {decision && (
        <ShippingDecision
          geography={directory?.geography ?? []}
          row={decision.row}
          command={decision.command}
          close={() => setDecision(undefined)}
          save={async (command, key) => {
            const next = await request(command, key);
            await refresh(next);
            setDecision(undefined);
            setNotice("عملیات حمل فروشگاه ثبت شد.");
          }}
        />
      )}
      {capacity && store && (
        <ShippingCapacityDialog
          storeID={store.id}
          methodID={capacity.base.id}
          name={`${store.name} / ${capacity.base.name}`}
          close={() => setCapacity(undefined)}
        />
      )}
    </div>
  );
}

function ShippingDecision({
  row,
  command,
  save,
  close,
  geography,
}: {
  row: Binding;
  command: Command;
  save: (command: Command, key: string) => Promise<void>;
  close: () => void;
  geography: Province[];
}) {
  const [reason, setReason] = useState(""),
    [busy, setBusy] = useState(false),
    [error, setError] = useState("");
  const [key, setKey] = useState(crypto.randomUUID());
  const review = command.action === "approve" || command.action === "reject";
  const view = command.action === "view";
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
    <Dialog open fullWidth maxWidth="lg" onClose={() => !busy && close()}>
      <DialogTitle>{row.base.name} — جزئیات و بررسی نسخه</DialogTitle>
      <DialogContent>
        {error && <Alert severity="error">{error}</Alert>}
        {row.base_changed && (
          <Alert severity="warning">
            روش عمومی از زمان ساخت درخواست تغییر کرده است؛ قبل از تأیید،
            پیش‌نویس را با روش عمومی جاری بازبینی و دوباره ارسال کنید.
          </Alert>
        )}
        {!view && (
          <Alert severity="info">
            عملیات «
            {
              (
                {
                  public: "پذیرش عمومی و کنارگذاشتن درخواست جاری",
                  disable: "غیرفعال‌سازی در فروشگاه",
                  submit: "ارسال برای بررسی",
                  approve: "تأیید و انتشار",
                  reject: "رد درخواست",
                } as Record<string, string>
              )[command.action]
            }
            » برای {command.stores.length.toLocaleString("fa-IR")} فروشگاه انجام
            می‌شود.
          </Alert>
        )}
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
            {busy ? "در حال ثبت…" : "تأیید عملیات"}
          </Button>
        )}
      </DialogActions>
    </Dialog>
  );
}
