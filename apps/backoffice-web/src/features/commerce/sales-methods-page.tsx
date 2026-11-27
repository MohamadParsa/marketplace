import { ChoiceField, MultiChoiceField } from "../../shared/option-select";
import { useEffect, useRef, useState } from "react";
import {
  Alert,
  Button,
  Checkbox,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
} from "@mui/material";
import { useWorkspace } from "../../shared/workspace";
import { DataTable, PageHeader } from "../../shared/components";
export type Sale = {
  id: string;
  name: string;
  active: boolean;
  markup_bps: number;
  payment_method_ids: string[];
};
type Override = { category_id: string; method_id: string; markup_bps: number };
type Directory = {
  version: number;
  methods: Sale[];
  overrides: Override[];
  can_write: boolean;
  payment_methods: { id: string; name: string }[];
};
async function request<T>(
  token: string,
  body?: unknown,
  key?: string,
): Promise<T> {
  const r = await fetch(
    (import.meta.env.VITE_API_BASE_URL ?? "") + "/backoffice/commerce/sales",
    {
      method: body ? "PUT" : "GET",
      cache: "no-store",
      headers: {
        Authorization: `Bearer ${token}`,
        "Content-Type": "application/json",
        ...(key ? { "Idempotency-Key": key } : {}),
      },
      body: body ? JSON.stringify(body) : undefined,
    },
  );
  if (!r.ok)
    throw Error(
      r.status === 409
        ? "اطلاعات تغییر کرده؛ داده تازه را دریافت و تغییرها را دوباره بررسی کنید."
        : r.status === 403
          ? "مجوز تنظیم روش فروش ندارید."
          : "درصد، ارتباط پرداخت یا اطلاعات واردشده معتبر نیست.",
    );
  return r.json();
}
function useDirectory() {
  const { authToken } = useWorkspace();
  const [data, setData] = useState<Directory>(),
    [error, setError] = useState(""),
    [busy, setBusy] = useState(false),
    [notice, setNotice] = useState("");
  const keys = useRef(new Map<string, string>()),
    running = useRef(false);
  const load = () => {
    setError("");
    return request<Directory>(authToken)
      .then(setData)
      .catch((e) => setError(e.message));
  };
  useEffect(() => {
    void load();
  }, [authToken]);
  const save = async (next: Directory) => {
    if (running.current) return false;
    running.current = true;
    setBusy(true);
    setError("");
    setNotice("");
    const body = {
      version: next.version,
      methods: next.methods,
      overrides: next.overrides,
    };
    const signature = JSON.stringify(body);
    let key = keys.current.get(signature);
    if (!key) {
      key = crypto.randomUUID();
      keys.current.set(signature, key);
    }
    try {
      const updated = await request<Directory>(authToken, body, key);
      setData({ ...next, ...updated });
      keys.current.delete(signature);
      setNotice("تنظیمات روش فروش ذخیره شد.");
      return true;
    } catch (e) {
      setError((e as Error).message);
      return false;
    } finally {
      running.current = false;
      setBusy(false);
    }
  };
  return { data, error, busy, notice, load, save };
}
const percent = (bps: number) => String(bps / 100);
function parsePercent(value: string) {
  value = value
    .trim()
    .replace(/[۰-۹]/g, (c) => String(c.charCodeAt(0) - 1776))
    .replace(/[٠-٩]/g, (c) => String(c.charCodeAt(0) - 1632))
    .replace(/٫/g, ".");
  if (!/^\d{1,3}(\.\d{1,2})?$/.test(value) || Number(value) > 100)
    throw Error("درصد باید از صفر تا صد و حداکثر دو رقم اعشار باشد.");
  return Math.round(Number(value) * 100);
}
export function SalesMethodsPage() {
  const dir = useDirectory();
  const [localError, setLocalError] = useState("");
  const [draft, setDraft] = useState<Sale>(),
    [rate, setRate] = useState("0");
  return (
    <>
      <PageHeader
        title="روش‌های فروش"
        description="مارکاپ پیش‌فرض و ارتباط روش پرداخت؛ کمیسیون و هزینه ارسال مستقل‌اند."
        action={
          <Button
            disabled={!dir.data?.can_write}
            onClick={() => {
              setLocalError("");
              setDraft({
                id: "chr-sale-" + crypto.randomUUID(),
                name: "",
                active: false,
                markup_bps: 0,
                payment_method_ids: [],
              });
              setRate("0");
            }}
          >
            روش فروش جدید
          </Button>
        }
      />
      {dir.error && <Alert severity="error">{dir.error}</Alert>}
      {dir.notice && <Alert severity="success">{dir.notice}</Alert>}
      <Button onClick={() => void dir.load()}>به‌روزرسانی فهرست</Button>
      <DataTable
        rows={dir.data?.methods ?? []}
        columns={[
          { title: "نام", render: (m) => m.name },
          { title: "درصد پیش‌فرض", render: (m) => percent(m.markup_bps) + "٪" },
          {
            title: "روش‌های پرداخت",
            render: (m) =>
              m.payment_method_ids
                .map(
                  (id) =>
                    dir.data?.payment_methods.find((p) => p.id === id)?.name ??
                    id,
                )
                .join("، "),
          },
          { title: "وضعیت", render: (m) => (m.active ? "فعال" : "غیرفعال") },
          {
            title: "عملیات",
            render: (m) => (
              <Button
                onClick={() => {
                  setLocalError("");
                  setDraft({
                    ...m,
                    payment_method_ids: [...m.payment_method_ids],
                  });
                  setRate(percent(m.markup_bps));
                }}
              >
                مشاهده و ویرایش
              </Button>
            ),
          },
        ]}
      />
      {draft && dir.data && (
        <Dialog
          open
          fullWidth
          maxWidth="lg"
          onClose={() => {
            if (!dir.busy) setDraft(undefined);
          }}
        >
          <DialogTitle>تنظیم روش فروش</DialogTitle>
          <DialogContent>
            <TextField
              fullWidth
              disabled={!dir.data.can_write || dir.busy}
              label="نام روش فروش"
              value={draft.name}
              onChange={(e) => setDraft({ ...draft, name: e.target.value })}
            />
            <TextField
              fullWidth
              disabled={!dir.data.can_write || dir.busy}
              label="مارکاپ پیش‌فرض (درصد)"
              value={rate}
              onChange={(e) => setRate(e.target.value)}
              inputProps={{ inputMode: "decimal", dir: "ltr" }}
            />
            <label>
              <Checkbox
                disabled={!dir.data.can_write || dir.busy}
                checked={draft.active}
                onChange={(_, v) => setDraft({ ...draft, active: v })}
              />
              فعال
            </label>
            <p>
              هر روش پرداخت فقط به یک روش فروش وصل می‌شود؛ انتخاب آن، ارتباط روش
              قبلی را منتقل می‌کند.
            </p>
            <MultiChoiceField
              label="روش‌های پرداخت"
              value={draft.payment_method_ids}
              options={dir.data.payment_methods.map((p) => ({
                id: p.id,
                name: p.name,
              }))}
              disabled={!dir.data.can_write || dir.busy}
              onChange={(payment_method_ids) =>
                setDraft({ ...draft, payment_method_ids })
              }
            />
            {(dir.error || localError) && (
              <Alert severity="error">{dir.error || localError}</Alert>
            )}
          </DialogContent>
          <DialogActions>
            <Button disabled={dir.busy} onClick={() => setDraft(undefined)}>
              انصراف
            </Button>
            <Button
              disabled={dir.busy || !dir.data.can_write || !draft.name.trim()}
              onClick={() => {
                let bps: number;
                try {
                  bps = parsePercent(rate);
                } catch (e) {
                  setLocalError((e as Error).message);
                  return;
                }
                setLocalError("");
                const edited = { ...draft, markup_bps: bps };
                const methods = dir
                  .data!.methods.filter((m) => m.id !== draft.id)
                  .map((m) => ({
                    ...m,
                    payment_method_ids: m.payment_method_ids.filter(
                      (id) => !draft.payment_method_ids.includes(id),
                    ),
                  }));
                methods.push(edited);
                void dir.save({ ...dir.data!, methods }).then((ok) => {
                  if (ok) setDraft(undefined);
                });
              }}
            >
              ذخیره روش فروش
            </Button>
          </DialogActions>
        </Dialog>
      )}
    </>
  );
}
export function CategorySalesDialog({
  id,
  close,
}: {
  id: string;
  close: () => void;
}) {
  const { state } = useWorkspace();
  const dir = useDirectory();
  const [values, setValues] = useState<Record<string, string>>({}),
    [localError, setLocalError] = useState("");
  useEffect(() => {
    if (dir.data)
      setValues(
        Object.fromEntries(
          dir.data.methods.map((m) => [
            m.id,
            dir.data!.overrides.find(
              (o) => o.category_id === id && o.method_id === m.id,
            )?.markup_bps == null
              ? ""
              : percent(
                  dir.data!.overrides.find(
                    (o) => o.category_id === id && o.method_id === m.id,
                  )!.markup_bps,
                ),
          ]),
        ),
      );
  }, [dir.data?.version, id]);
  function inherited(m: Sale) {
    let category = (state.categories ?? []).find((c) => c.id === id)?.parent_id;
    const seen = new Set<string>();
    while (category && !seen.has(category)) {
      seen.add(category);
      const rule = dir.data?.overrides.find(
        (o) => o.category_id === category && o.method_id === m.id,
      );
      if (rule)
        return `${percent(rule.markup_bps)}٪ از ${(state.categories ?? []).find((c) => c.id === category)?.name}`;
      category = (state.categories ?? []).find(
        (c) => c.id === category,
      )?.parent_id;
    }
    return `${percent(m.markup_bps)}٪ پیش‌فرض روش`;
  }
  return (
    <Dialog
      open
      fullWidth
      maxWidth="lg"
      onClose={() => {
        if (!dir.busy) close();
      }}
    >
      <DialogTitle>روش‌های فروش و مارکاپ دسته</DialogTitle>
      <DialogContent>
        <p>
          خالی یعنی ارث‌بری؛ صفر صریح نرخ والد را خنثی می‌کند. نرخ روی قیمت نقدی
          پس از تخفیف مستقیم و قبل از کوپن اعمال می‌شود.
        </p>
        {dir.data?.methods.map((m) => (
          <div className="sales-rate-row" key={m.id}>
            <b>
              {m.name} · {m.active ? "فعال" : "غیرفعال"}
            </b>
            <small>نرخ ارث‌رسیده: {inherited(m)}</small>
            <TextField
              label={`مارکاپ ${m.name} در این دسته (درصد)`}
              value={values[m.id] ?? ""}
              disabled={!dir.data?.can_write || dir.busy}
              onChange={(e) => setValues({ ...values, [m.id]: e.target.value })}
            />
            <Button
              disabled={!dir.data?.can_write || dir.busy}
              onClick={() => setValues({ ...values, [m.id]: "" })}
            >
              بازگشت به ارث‌بری
            </Button>
          </div>
        ))}
        {(dir.error || localError) && (
          <Alert severity="error">{dir.error || localError}</Alert>
        )}
      </DialogContent>
      <DialogActions>
        <Button disabled={dir.busy} onClick={close}>
          بستن
        </Button>
        <Button
          disabled={!dir.data?.can_write || dir.busy}
          onClick={() => {
            try {
              const overrides = dir.data!.overrides.filter(
                (o) => o.category_id !== id,
              );
              for (const [method, value] of Object.entries(values)) {
                if (value.trim() !== "")
                  overrides.push({
                    category_id: id,
                    method_id: method,
                    markup_bps: parsePercent(value),
                  });
              }
              setLocalError("");
              void dir.save({ ...dir.data!, overrides }).then((ok) => {
                if (ok) close();
              });
            } catch (e) {
              setLocalError((e as Error).message);
            }
          }}
        >
          ذخیره مارکاپ دسته
        </Button>
      </DialogActions>
    </Dialog>
  );
}
