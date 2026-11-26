import { useRef, useState } from "react";
import {
  Alert,
  Button,
  Checkbox,
  Chip,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
} from "@mui/material";
import type { CustomRole, RoleScope } from "../../shared/types";
import { useWorkspace } from "../../shared/workspace";
import { DataTable, PageHeader } from "../../shared/components";
const names: Record<string, string> = {
 "category.axes.convert": "تبدیل کنترل‌شده محورهای تنوع",
  "users.read": "مشاهده فهرست کاربران",
  "users.write": "ایجاد و ویرایش کاربران",
  "users.password.reset": "تعیین و بازنشانی رمز موقت",
  "grid.export": "خروجی Excel گرید",
  "workspace.read": "مشاهده بک‌آفیس",
  "finance.read": "مشاهده مالی",
  "commerce.sales.read": "مشاهده روش فروش",
  "commerce.sales.write": "تنظیم روش فروش",
  "offers.read": "مشاهده عرضه‌ها",
  "stock.write": "ویرایش موجودی",
  "offers.create": "ایجاد عرضه",
  "offers.price.write": "ویرایش قیمت عرضه",
  "orders.read": "مشاهده سفارش فروشگاه",
  "products.content.write": "ویرایش محتوای فروشگاه",
};
const words: Record<string, string> = {
  commerce: "خرید و مالی",
  wallet: "کیف پول",
  credit: "شارژ اعتبار",
  coupons: "کد تخفیف",
  activate: "فعال‌سازی",
  create: "ایجاد",
  edit: "ویرایش",
  settings: "تنظیمات",
  payment: "پرداخت",
  shipping: "ارسال",
  geography: "استان و شهر",
  business: "تقویم کسب‌وکار",
  orders: "سفارش‌ها",
  confirm: "تأیید",
  deposits: "واریزها",
  review: "بررسی و تأیید",
  refunds: "بازپرداخت‌ها",
  resolve: "رسیدگی",
  store: "فروشگاه",
  content: "محتوا",
  policy: "سیاست",
  set: "تنظیم",
  product: "محصول",
  category: "دسته‌بندی",
  publish: "انتشار",
  banner: "بنر",
  save: "ذخیره",
  media: "رسانه",
  remove: "حذف",
  icon: "تصویر آیکن",
  brand: "برند",
  attribute: "خصوصیت",
  groups: "گروه‌ها",
  update: "ویرایش",
  delete: "حذف",
  bindings: "انتساب خصوصیات",
  card: "کارت",
  limit: "حداکثر",
  sales: "فروش",
  window: "بازه آمار",
  variant: "تنوع",
  offer: "عرضه",
  price: "قیمت",
  stock: "موجودی",
  homepage: "صفحه اصلی",
  catalog: "کاتالوگ",
  toggle: "تغییر وضعیت",
  archive: "بایگانی",
  reject: "رد",
  case: "پرونده پشتیبانی",
  start: "شروع رسیدگی",
  close: "بستن",
  refund: "بازپرداخت",
  assign: "ارجاع",
  order: "سفارش",
  deliver: "ثبت تحویل",
  transfer: "انتقال مالی",
  rule: "قانون قیمت",
  notification: "اعلان",
  retry: "تلاش مجدد",
  read: "مشاهده",
  write: "ویرایش",
};
export const permissionName = (key: string) =>
  names[key] ??
  key
    .split(/[.-]/)
    .map((part) => words[part] ?? part)
    .join(" / ");
export function RolesPage() {
  const { state, role, execute } = useWorkspace();
  const access = state.access;
  const [draft, setDraft] = useState<CustomRole>(),
    [search, setSearch] = useState(""),
    [error, setError] = useState(""),
    [busy, setBusy] = useState(false);
  const keys = useRef(new Map<string, string>());
  if (role !== "super-admin" || !access)
    return (
      <Alert severity="warning">مدیریت نقش‌ها فقط برای ادمین مجاز است.</Alert>
    );
  async function run(action: string, value: CustomRole) {
    if (busy) return;
    setError("");
    setBusy(true);
    const body = {
      action,
      target: value.id,
      version: value.version,
      name: value.name,
      scope: value.scope,
      permissions: value.permissions,
    };
    const signature = JSON.stringify(body);
    let key = keys.current.get(signature);
    if (!key) {
      key = crypto.randomUUID();
      keys.current.set(signature, key);
    }
    try {
      await execute({ ...body, idempotency_key: key });
      keys.current.delete(signature);
      setDraft(undefined);
    } catch (e) {
      setError((e as Error).message);
    } finally {
      setBusy(false);
    }
  }
  const roles = access.roles ?? [];
  const used = (id: string) =>
    (access.grants ?? []).filter((g) => g.role_id === id).length;
  const scopes: { id: RoleScope; name: string }[] = [
    { id: "global", name: "بک‌آفیس" },
    { id: "store", name: "فروشگاه" },
  ];
  const registry = draft ? (access.registry[draft.scope] ?? []) : [];
  const visible = registry.filter((p) =>
    (p + " " + permissionName(p))
      .toLocaleLowerCase()
      .includes(search.toLocaleLowerCase()),
  );
  return (
    <>
      <PageHeader
        title="نقش‌ها و مجوزها"
        description="تعریف نقش مستقل از انتساب است؛ برای اختصاص نقش به یک حساب، فرم کاربران را باز کنید."
        action={
          <Button
            variant="contained"
            onClick={() => {
              setDraft({
                id: "",
                name: "",
                scope: "global",
                permissions: [],
                version: 0,
              });
              setError("");
              setSearch("");
            }}
          >
            نقش جدید
          </Button>
        }
      />
      <Alert severity="info">
        ادمین و نقش‌های پایه سیستم از فرم کاربر انتخاب می‌شوند. نقش‌های این
        صفحه، مجموعه مجوزهای سفارشی در بک‌آفیس یا یک فروشگاه هستند.
      </Alert>
      <section className="panel">
        <DataTable
          rows={roles}
          columns={[
            { title: "نام نقش", render: (r) => <b>{r.name}</b> },
            {
              title: "محدوده",
              render: (r) => (r.scope === "global" ? "بک‌آفیس" : "فروشگاه"),
            },
            { title: "تعداد مجوز", render: (r) => r.permissions.length },
            { title: "انتساب فعال", render: (r) => used(r.id) },
            {
              title: "مجوزها",
              render: (r) =>
                r.permissions.map((p) => permissionName(p)).join("، "),
            },
            {
              title: "عملیات",
              render: (r) => (
                <Button
                  onClick={() => {
                    setDraft({ ...r, permissions: [...r.permissions] });
                    setError("");
                    setSearch("");
                  }}
                >
                  مشاهده و ویرایش نقش
                </Button>
              ),
            },
          ]}
        />
      </section>
      {draft && (
        <Dialog
          open
          fullWidth
          maxWidth="md"
          onClose={() => {
            if (!busy) setDraft(undefined);
          }}
        >
          <DialogTitle>{draft.id ? "ویرایش نقش" : "تعریف نقش"}</DialogTitle>
          <DialogContent className="user-editor">
            {error && <Alert severity="error">{error}</Alert>}
            <TextField
              fullWidth
              label="نام نقش"
              value={draft.name}
              disabled={busy}
              inputProps={{ maxLength: 160 }}
              onChange={(e) => setDraft({ ...draft, name: e.target.value })}
            />
            <fieldset className="user-choice-group">
              <legend>محدوده نقش</legend>
              {scopes.map((s) => (
                <Chip
                  key={s.id}
                  component="button"
                  type="button"
                  clickable
                  disabled={busy || !!draft.id}
                  label={s.name}
                  color={draft.scope === s.id ? "primary" : "default"}
                  aria-pressed={draft.scope === s.id}
                  onClick={() =>
                    setDraft({ ...draft, scope: s.id, permissions: [] })
                  }
                />
              ))}
            </fieldset>
            <p>
              مجوز فروشگاه فقط هنگام انتساب به همان فروشگاه اثر دارد؛ به
              فروشگاه‌های دیگر تعمیم داده نمی‌شود.
            </p>
            <TextField
              fullWidth
              size="small"
              label="جست‌وجوی مجوزها"
              value={search}
              onChange={(e) => setSearch(e.target.value)}
            />
            <p>
              {draft.permissions.length.toLocaleString("fa-IR")} مجوز انتخاب‌شده
            </p>
            <div className="role-permission-grid">
              {visible.map((p) => (
                <label key={p}>
                  <Checkbox
                    disabled={busy}
                    checked={draft.permissions.includes(p)}
                    onChange={(_, v) =>
                      setDraft({
                        ...draft,
                        permissions: v
                          ? [...draft.permissions, p]
                          : draft.permissions.filter((x) => x !== p),
                      })
                    }
                  />
                  <span>
                    {permissionName(p)}
                    <small dir="ltr">{p}</small>
                  </span>
                </label>
              ))}
            </div>
            {!visible.length && <p>مجوزی مطابق جست‌وجو وجود ندارد.</p>}
            {draft.id && (
              <p>
                این نقش {used(draft.id).toLocaleString("fa-IR")} انتساب دارد؛
                نقش در حال استفاده حذف نمی‌شود.
              </p>
            )}
          </DialogContent>
          <DialogActions>
            <Button disabled={busy} onClick={() => setDraft(undefined)}>
              انصراف
            </Button>
            {draft.id && (
              <Button
                color="error"
                disabled={busy || used(draft.id) > 0}
                onClick={() => void run("role.delete", draft)}
              >
                حذف نقش بدون انتساب
              </Button>
            )}
            <Button
              variant="contained"
              disabled={busy || !draft.name.trim() || !draft.permissions.length}
              onClick={() =>
                void run(draft.id ? "role.update" : "role.create", draft)
              }
            >
              ذخیره نقش
            </Button>
          </DialogActions>
        </Dialog>
      )}
    </>
  );
}
