import { useSearchParams } from "react-router-dom";
import { permissionName } from "./access-panel";
import { useEffect, useRef, useState } from "react";
import {
  Alert,
  Button,
  Chip,
  Checkbox,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
  LinearProgress,
} from "@mui/material";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { useWorkspace } from "../../shared/workspace";
import { ActionDialog, DataTable, PageHeader } from "../../shared/components";
import { ApiError } from "../../shared/api";
import type { Command, CustomRole } from "../../shared/types";
import { WalletCreditDialog } from "./wallet-credit-dialog";
export type ManagedUser = {
  id: string;
  name: string;
  first_name: string;
  last_name: string;
  phone: string;
  email: string;
  gender: string;
  active: boolean;
  version: number;
  base_role: string;
  has_password: boolean;
  password_change_required: boolean;
  store_ids: string[];
  grants: { user_id: string; role_id: string; store_id?: string }[];
};
type Filter = {
  id?: string;
  first_name: string;
  last_name: string;
  phone: string;
  email: string;
  active: string;
  gender: string;
  role: string;
  store: string;
  sort: string;
};
type Directory = {
  rows: ManagedUser[];
  total: number;
  page: number;
  size: number;
  can_write: boolean;
  can_reset_password: boolean;
  can_assign: boolean;
  roles: CustomRole[];
};
const defaults: Filter = {
  first_name: "",
  last_name: "",
  phone: "",
  email: "",
  active: "",
  gender: "",
  role: "",
  store: "",
  sort: "name",
};
export const baseRoles = [
  { id: "", name: "فقط مشتری" },
  { id: "super-admin", name: "ادمین" },
  { id: "operations", name: "عملیات" },
  { id: "support", name: "پشتیبانی" },
  { id: "finance", name: "مالی" },
  { id: "viewer", name: "مشاهده‌گر" },
];
async function usersRequest<T>(
  token: string,
  path: string,
  body: unknown,
  key?: string,
  signal?: AbortSignal,
): Promise<T> {
  const r = await fetch(
    (import.meta.env.VITE_API_BASE_URL ?? "") + "/backoffice/users/" + path,
    {
      method: "POST",
      signal,
      cache: "no-store",
      headers: {
        Authorization: `Bearer ${token}`,
        "Content-Type": "application/json",
        ...(key ? { "Idempotency-Key": key } : {}),
      },
      body: JSON.stringify(body),
    },
  );
  if (!r.ok) {
    const b = await r.json().catch(() => null);
    throw new ApiError(r.status, b?.code);
  }
  return r.json();
}
function Choices({
  label,
  value,
  options,
  change,
  disabled = false,
}: {
  label: string;
  value: string;
  options: { id: string; name: string }[];
  change: (v: string) => void;
  disabled?: boolean;
}) {
  return (
    <fieldset className="user-choice-group">
      <legend>{label}</legend>
      {options.map((o) => (
        <Chip
          key={o.id}
          label={o.name}
          component="button"
          type="button"
          clickable
          disabled={disabled}
          aria-pressed={value === o.id}
          color={value === o.id ? "primary" : "default"}
          onClick={() => change(o.id)}
        />
      ))}
    </fieldset>
  );
}
export function UsersPage() {
  const { authToken, state, role, can } = useWorkspace();
  const cache = useQueryClient();
  const [params, setParams] = useSearchParams();
  const linkedUser = params.get("user");
  const linkedOpened = useRef("");
  const [draft, setDraft] = useState<Filter>({
      ...defaults,
      id: linkedUser ?? undefined,
    }),
    [filter, setFilter] = useState<Filter>({
      ...defaults,
      id: linkedUser ?? undefined,
    }),
    [page, setPage] = useState(0),
    [size, setSize] = useState(20),
    [editing, setEditing] = useState<ManagedUser | "new">(),
    [wallet, setWallet] = useState<ManagedUser>();
  const query = useQuery<Directory>({
    queryKey: ["managed-users", authToken, filter, page, size],
    placeholderData: (previous) => previous,
    queryFn: ({ signal }) =>
      usersRequest<Directory>(
        authToken,
        "query",
        { filter: { ...filter, page, size } },
        undefined,
        signal,
      ),
  });
  const data = query.data;
  useEffect(() => {
    if (linkedUser) {
      setFilter((old) =>
        old.id === linkedUser ? old : { ...defaults, id: linkedUser },
      );
      setDraft((old) =>
        old.id === linkedUser ? old : { ...defaults, id: linkedUser },
      );
      setPage(0);
    } else linkedOpened.current = "";
  }, [linkedUser]);
  useEffect(() => {
    if (linkedUser && linkedOpened.current !== linkedUser) {
      const row = data?.rows.find((r) => r.id === linkedUser);
      if (row) {
        linkedOpened.current = linkedUser;
        setEditing(row);
      }
    }
  }, [linkedUser, data]);
  const closeEditing = () => {
    setEditing(undefined);
    if (linkedUser) {
      const next = new URLSearchParams(params);
      next.delete("user");
      setParams(next, { replace: true });
    }
  };
  const refresh = () =>
    cache.invalidateQueries({ queryKey: ["managed-users"] });
  const setField = (k: keyof Filter, v: string) =>
    setDraft((x) => ({ ...x, [k]: v }));
  const canCredit = can("commerce.wallet.credit", "finance");
  return (
    <>
      <PageHeader
        title="کاربران"
        description="یک حساب برای مشتری، کارکنان و اعضای فروشگاه؛ اطلاعات و انتساب‌ها مستقل مدیریت می‌شوند."
        action={
          <Button
            variant="contained"
            disabled={!data?.can_write}
            onClick={() => setEditing("new")}
          >
            کاربر جدید
          </Button>
        }
      />
      <section className="panel">
        <form
          className="user-filter-form"
          onSubmit={(e) => {
            e.preventDefault();
            setFilter({ ...draft });
            setPage(0);
          }}
        >
          <div className="user-form-grid">
            {(
              [
                ["first_name", "نام"],
                ["last_name", "نام خانوادگی"],
                ["phone", "شماره موبایل"],
                ["email", "ایمیل"],
              ] as const
            ).map(([k, label]) => (
              <TextField
                size="small"
                key={k}
                label={`فیلتر ${label}`}
                value={draft[k]}
                onChange={(e) => setField(k, e.target.value)}
              />
            ))}
          </div>
          <Choices
            label="وضعیت کاربر"
            value={draft.active}
            change={(v) => setField("active", v)}
            options={[
              { id: "", name: "همه" },
              { id: "true", name: "فعال" },
              { id: "false", name: "غیرفعال" },
            ]}
          />
          <Choices
            label="جنسیت"
            value={draft.gender}
            change={(v) => setField("gender", v)}
            options={[
              { id: "", name: "همه" },
              { id: "female", name: "زن" },
              { id: "male", name: "مرد" },
              { id: "unspecified", name: "تعیین‌نشده" },
            ]}
          />
          <details>
            <summary>فیلتر نقش و فروشگاه</summary>
            <Choices
              label="نقش کاربر"
              value={draft.role}
              change={(v) => setField("role", v)}
              options={[
                { id: "", name: "همه نقش‌ها" },
                ...baseRoles.filter((r) => r.id),
                ...(data?.roles ?? []).map((r) => ({ id: r.id, name: r.name })),
              ]}
            />
            <Choices
              label="فروشگاه"
              value={draft.store}
              change={(v) => setField("store", v)}
              options={[
                { id: "", name: "همه فروشگاه‌ها" },
                ...(state.stores ?? []).map((s) => ({
                  id: s.id,
                  name: s.name,
                })),
              ]}
            />
          </details>
          <Choices
            label="مرتب‌سازی کاربران"
            value={draft.sort}
            change={(v) => setField("sort", v)}
            options={[
              { id: "name", name: "نام صعودی" },
              { id: "name-desc", name: "نام نزولی" },
              { id: "id", name: "شناسه" },
            ]}
          />
          <div className="user-inline-actions">
            <Button type="submit" variant="contained">
              اعمال فیلترها
            </Button>
            <Button
              onClick={() => {
                setDraft(defaults);
                setFilter(defaults);
                setPage(0);
              }}
            >
              پاک‌کردن فیلترها
            </Button>
            <Button onClick={() => void query.refetch()}>
              دریافت داده تازه
            </Button>
          </div>
        </form>
        {query.isFetching && <LinearProgress aria-label="دریافت کاربران" />}
        {query.error && <Alert severity="error">{query.error.message}</Alert>}
        <fieldset
          className="user-grid-region"
          disabled={query.isFetching}
          aria-busy={query.isFetching}
        >
          <DataTable
            rows={data?.rows ?? []}
            remote={{
              filterKey: JSON.stringify(filter),
              controls: {
                page: data?.page ?? page,
                pages: Math.max(1, Math.ceil((data?.total ?? 0) / size)),
                size,
                total: data?.total ?? 0,
                setPage,
                setSize: (n) => {
                  setSize(n);
                  setPage(0);
                },
              },
              exportRows: async (ids, signal) =>
                (
                  await usersRequest<Directory>(
                    authToken,
                    "export",
                    { filter, ids: ids ?? [] },
                    undefined,
                    signal,
                  )
                ).rows,
            }}
            columns={[
              {
                title: "نام و نام خانوادگی",
                render: (u) => (
                  <>
                    <b>{u.name}</b>
                    <small>
                      {u.first_name} {u.last_name}
                    </small>
                  </>
                ),
              },
              {
                title: "شماره موبایل",
                render: (u) => <span dir="ltr">{u.phone}</span>,
              },
              {
                title: "ایمیل",
                render: (u) => <span dir="ltr">{u.email || "—"}</span>,
              },
              {
                title: "نقش‌ها",
                render: (u) =>
                  [
                    baseRoles.find((r) => r.id === u.base_role)?.name,
                    ...u.grants.map(
                      (g) =>
                        data?.roles.find((r) => r.id === g.role_id)?.name ??
                        g.role_id,
                    ),
                  ]
                    .filter(Boolean)
                    .join("، "),
              },
              {
                title: "فروشگاه‌ها",
                render: (u) =>
                  u.store_ids
                    .map(
                      (id) =>
                        state.stores?.find((s) => s.id === id)?.name ?? id,
                    )
                    .join("، ") || "—",
              },
              {
                title: "جنسیت",
                render: (u) =>
                  u.gender === "male"
                    ? "مرد"
                    : u.gender === "female"
                      ? "زن"
                      : "تعیین‌نشده",
              },
              {
                title: "وضعیت",
                render: (u) => (u.active ? "فعال" : "غیرفعال"),
              },
              {
                title: "عملیات",
                render: (u) => (
                  <>
                    <Button onClick={() => setEditing(u)}>
                      مشاهده و ویرایش
                    </Button>
                    {canCredit && (
                      <Button onClick={() => setWallet(u)}>شارژ کیف پول</Button>
                    )}
                  </>
                ),
              },
            ]}
          />
        </fieldset>
      </section>
      {editing && data && (
        <UserEditor
          key={typeof editing === "string" ? "new" : editing.id}
          row={editing === "new" ? undefined : editing}
          directory={data}
          close={closeEditing}
          saved={(u) => {
            setEditing(u);
            void refresh();
            void cache.invalidateQueries({ queryKey: ["workspace"] });
          }}
        />
      )}
      {wallet && (
        <WalletCreditDialog
          user={wallet}
          close={() => setWallet(undefined)}
          onCredited={() => void refresh()}
        />
      )}
      {role !== "super-admin" && (
        <p className="muted">
          تغییر نقش و دسترسی حساب‌های کارکنان فقط توسط ادمین انجام می‌شود.
        </p>
      )}
    </>
  );
}
function UserEditor({
  row,
  directory,
  close,
  saved,
}: {
  row?: ManagedUser;
  directory: Directory;
  close: () => void;
  saved: (u: ManagedUser) => void;
}) {
  const { authToken, state, execute, requestAdminCode } = useWorkspace();
  const [current, setCurrent] = useState(row),
    [form, setForm] = useState({
      first_name: row?.first_name ?? "",
      last_name: row?.last_name ?? "",
      phone: row?.phone ?? "",
      email: row?.email ?? "",
      gender: row?.gender ?? "",
      active: row?.active ?? true,
      temporary_password: "",
    }),
    [error, setError] = useState(""),
    [busy, setBusy] = useState(false),
    [pending, setPending] = useState<Command>();
  const keys = useRef(new Map<string, string>());
  const formField = (k: string, v: string | boolean) =>
    setForm((x) => ({ ...x, [k]: v }));
  useEffect(() => {
    if (row) setCurrent(row);
  }, [row]);
  const refreshCurrent = async () => {
    if (current) {
      const d = await usersRequest<Directory>(authToken, "query", {
        filter: { id: current.id },
      });
      if (d.rows[0]) {
        setCurrent(d.rows[0]);
        if (d.rows[0].active !== current.active)
          setForm((x) => ({ ...x, active: d.rows[0].active }));
        saved(d.rows[0]);
      }
    }
  };
  const protectedUser =
    current && (current.base_role || current.grants.some((g) => !g.store_id));
  const writable =
    directory.can_write && (!protectedUser || directory.can_assign);
  const command = (action: string, extra: Partial<Command>) => {
    if (!current) return;
    setPending({
      action,
      target: current.id,
      user_version: current.version,
      role_version:
        action === "user.role-grant"
          ? directory.roles.find((r) => r.id === extra.value)?.version
          : undefined,
      idempotency_key: crypto.randomUUID(),
      ...extra,
    });
  };
  const roleNames = (g: { role_id: string; store_id?: string }) =>
    (directory.roles.find((r) => r.id === g.role_id)?.name ?? g.role_id) +
    (g.store_id
      ? " · " +
        (state.stores?.find((s) => s.id === g.store_id)?.name ?? g.store_id)
      : " · بک‌آفیس");
  return (
    <>
      <Dialog
        open
        fullWidth
        maxWidth="md"
        onClose={() => {
          if (!busy) close();
        }}
      >
        <DialogTitle>
          {current ? "ویرایش کاربر · " + current.name : "ایجاد کاربر"}
        </DialogTitle>
        <DialogContent className="user-editor">
          {error && <Alert severity="error">{error}</Alert>}
          {current && !current.first_name && !current.last_name && (
            <Alert severity="info">
              نام قبلی «{current.name}» بدون حدس‌زدن نام خانوادگی حفظ شده است؛
              نام و نام خانوادگی را می‌توانید جدا تکمیل کنید.
            </Alert>
          )}
          <form
            id="managed-user-form"
            onSubmit={async (e) => {
              e.preventDefault();
              if (busy) return;
              setError("");
              setBusy(true);
              const input = {
                ...form,
                id: current?.id ?? "",
                version: current?.version ?? 0,
              };
              const signature = JSON.stringify(input);
              let key = keys.current.get(signature);
              if (!key) {
                key = crypto.randomUUID();
                keys.current.set(signature, key);
              }
              try {
                const next = await usersRequest<ManagedUser>(
                  authToken,
                  "save",
                  input,
                  key,
                );
                keys.current.delete(signature);
                setCurrent(next);
                setForm((x) => ({ ...x, temporary_password: "" }));
                saved(next);
              } catch (e) {
                setError((e as Error).message);
              } finally {
                setBusy(false);
              }
            }}
          >
            <fieldset disabled={busy || !writable}>
              <legend>اطلاعات حساب</legend>
              <div className="user-form-grid">
                {(
                  [
                    ["first_name", "نام"],
                    ["last_name", "نام خانوادگی"],
                    ["phone", "شماره موبایل"],
                    ["email", "ایمیل"],
                  ] as const
                ).map(([k, label]) => (
                  <TextField
                    key={k}
                    fullWidth
                    label={label}
                    value={form[k]}
                    required={k === "phone"}
                    disabled={k === "phone" && !!current}
                    onChange={(e) => formField(k, e.target.value)}
                    inputProps={{
                      maxLength: k === "email" ? 254 : 160,
                      dir: k === "email" || k === "phone" ? "ltr" : "rtl",
                    }}
                  />
                ))}
              </div>
              <Choices
                label="جنسیت حساب"
                value={form.gender}
                disabled={busy || !writable}
                change={(v) => formField("gender", v)}
                options={[
                  { id: "", name: "تعیین‌نشده" },
                  { id: "female", name: "زن" },
                  { id: "male", name: "مرد" },
                ]}
              />
              <label>
                <Checkbox
                  checked={form.active}
                  disabled={
                    current?.base_role === "super-admin" && !current.active
                  }
                  onChange={(_, v) => formField("active", v)}
                />
                حساب فعال است
              </label>
            </fieldset>
            {directory.can_reset_password && writable && (
              <fieldset disabled={busy}>
                <legend>رمز اولیه یا بازنشانی رمز</legend>
                <TextField
                  fullWidth
                  label="رمز موقت جدید"
                  type="password"
                  autoComplete="new-password"
                  value={form.temporary_password}
                  onChange={(e) =>
                    formField("temporary_password", e.target.value)
                  }
                  helperText="خالی یعنی بدون تغییر رمز. ثبت رمز جدید، نشست‌های قبلی را باطل و تعیین رمز شخصی را در ورود بعدی الزامی می‌کند."
                />
              </fieldset>
            )}
            {current && (
              <p>
                {current.password_change_required
                  ? "وضعیت رمز: تغییر رمز شخصی الزامی است"
                  : current.has_password
                    ? "رمز شخصی تنظیم شده است"
                    : "ورود با OTP؛ هنوز رمز ندارد"}
              </p>
            )}
          </form>
          {current && directory.can_assign && (
            <>
              <section className="user-access-box">
                <h3>نقش پایه بک‌آفیس</h3>
                {current.base_role === "super-admin" && !current.active && (
                  <Button onClick={() => command("user.toggle", {})}>
                    فعال‌سازی ادمین با تأیید OTP
                  </Button>
                )}
                <p>
                  ادمین اختیار کامل دارد؛ ارتقا به ادمین به OTP احراز مجدد نیاز
                  دارد.
                </p>
                <div className="user-inline-actions">
                  {baseRoles.map((r) => (
                    <Chip
                      component="button"
                      type="button"
                      key={r.id}
                      label={r.name}
                      clickable
                      color={current.base_role === r.id ? "primary" : "default"}
                      onClick={() => command("user.role", { value: r.id })}
                    />
                  ))}
                </div>
              </section>
              <section className="user-access-box">
                <h3>عضویت فروشگاه‌ها</h3>
                {(state.stores ?? []).map((s) => (
                  <div className="user-access-row" key={s.id}>
                    <span>
                      {s.name}
                      {current.store_ids.includes(s.id)
                        ? " · عضو"
                        : " · عضو نیست"}
                    </span>
                    <Button
                      onClick={() =>
                        command(
                          current.store_ids.includes(s.id)
                            ? "membership.remove"
                            : "membership.add",
                          { store_id: s.id },
                        )
                      }
                    >
                      {current.store_ids.includes(s.id)
                        ? "لغو عضویت و نقش‌های این فروشگاه"
                        : "افزودن عضویت"}
                    </Button>
                  </div>
                ))}
              </section>
              <section className="user-access-box">
                <h3>نقش‌های منتسب</h3>
                {current.grants.map((g) => (
                  <div className="user-access-row" key={g.role_id + g.store_id}>
                    <span>{roleNames(g)}</span>
                    <Button
                      onClick={() =>
                        command("user.role-revoke", {
                          value: g.role_id,
                          store_id: g.store_id,
                        })
                      }
                    >
                      سلب نقش
                    </Button>
                  </div>
                ))}
                <h4>نقش‌های بک‌آفیس قابل انتساب</h4>
                {directory.roles
                  .filter(
                    (r) =>
                      r.scope === "global" &&
                      !current.grants.some((g) => g.role_id === r.id),
                  )
                  .map((r) => (
                    <Button
                      key={r.id}
                      onClick={() =>
                        command("user.role-grant", { value: r.id })
                      }
                    >
                      انتساب {r.name}
                    </Button>
                  ))}
                {current.store_ids.map((id) => (
                  <div key={id}>
                    <h4>
                      نقش در{" "}
                      {state.stores?.find((s) => s.id === id)?.name ?? id}
                    </h4>
                    {directory.roles
                      .filter(
                        (r) =>
                          r.scope === "store" &&
                          !current.grants.some(
                            (g) => g.role_id === r.id && g.store_id === id,
                          ),
                      )
                      .map((r) => (
                        <Button
                          key={r.id}
                          onClick={() =>
                            command("user.role-grant", {
                              value: r.id,
                              store_id: id,
                            })
                          }
                        >
                          انتساب {r.name}
                        </Button>
                      ))}
                  </div>
                ))}
                <h4>دسترسی مؤثر در هر محدوده</h4>
                {["", ...current.store_ids].map((id) => {
                  const roles = (state.access?.roles ?? []).filter((r) =>
                    current.grants.some(
                      (g) => g.role_id === r.id && (g.store_id ?? "") === id,
                    ),
                  );
                  const permissions = [
                    ...new Set(roles.flatMap((r) => r.permissions)),
                  ];
                  return (
                    <div
                      className="user-permission-summary"
                      key={id || "global"}
                    >
                      <b>
                        {id
                          ? (state.stores?.find((s) => s.id === id)?.name ?? id)
                          : "بک‌آفیس"}
                      </b>
                      <p>
                        منبع:{" "}
                        {roles.map((r) => r.name).join("، ") ||
                          "بدون نقش سفارشی"}
                        {!id && current.base_role
                          ? " · نقش پایه " +
                            (baseRoles.find((r) => r.id === current.base_role)
                              ?.name ?? current.base_role)
                          : ""}
                      </p>
                      <p>
                        {permissions.map(permissionName).join(" · ") ||
                          "مجوز سفارشی ندارد"}
                      </p>
                    </div>
                  );
                })}
              </section>
            </>
          )}
          {!current && (
            <Alert severity="info">
              پس از ایجاد حساب، عضویت فروشگاه و نقش‌های آن را در همین فرم مشخص
              کنید.
            </Alert>
          )}
        </DialogContent>
        <DialogActions>
          <Button disabled={busy} onClick={close}>
            بستن
          </Button>
          <Button
            type="submit"
            form="managed-user-form"
            variant="contained"
            disabled={busy || !writable}
          >
            ذخیره اطلاعات کاربر
          </Button>
        </DialogActions>
      </Dialog>
      {pending && (
        <ActionDialog
          title="تأیید تغییر دسترسی کاربر"
          open
          fields={[]}
          requestAdminCode={requestAdminCode}
          onClose={() => setPending(undefined)}
          onSubmit={async (_values, reauthentication) => {
            await execute({ ...pending, reauthentication });
            await refreshCurrent();
            setPending(undefined);
          }}
        />
      )}
    </>
  );
}
