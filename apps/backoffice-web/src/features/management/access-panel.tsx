import { useState } from "react";
import { Alert, Button, Checkbox, FormControlLabel, MenuItem, Stack, TextField } from "@mui/material";
import type { Command, CustomRole, RoleScope } from "../../shared/types";
import { useWorkspace } from "../../shared/workspace";

const permissionNames: Record<string, string> = {
  "grid.export": "خروجی Excel گرید", "commerce.sales.read": "مشاهده روش‌های فروش", "commerce.sales.write": "تنظیم روش‌های فروش و مارکاپ",
  "workspace.read": "مشاهده بک‌آفیس", "finance.read": "مشاهده داده مالی",
  "store.create": "ایجاد فروشگاه", "store.toggle": "تغییر وضعیت فروشگاه",
  "category.create": "ایجاد دسته", "category.archive": "بایگانی دسته",
  "product.publish": "انتشار محصول", "product.reject": "رد محصول", "shipping.update": "ویرایش تعرفه ارسال",
  "case.start": "آغاز رسیدگی", "case.close": "بستن پرونده", "case.refund": "درخواست بازپرداخت",
  "case.assign": "ارجاع تأمین جایگزین", "order.deliver": "ثبت تحویل", "transfer.resolve": "تعیین نتیجه انتقال",
  "rule.toggle": "تغییر وضعیت قانون کمیسیون", "notification.retry": "ارسال دوباره اعلان",
  "offers.read": "مشاهده عرضه‌ها", "stock.write": "ویرایش موجودی",
};
const scopeName = (scope: RoleScope) => scope === "global" ? "بک‌آفیس" : "فروشگاه";

export function AccessPanel() {
  const { state, execute, busy } = useWorkspace();
  const access = state.access;
  const [error, setError] = useState("");
  const [draft, setDraft] = useState<CustomRole | null>(null);
  const [userID, setUserID] = useState("");
  const [roleID, setRoleID] = useState("");
  const [storeID, setStoreID] = useState("");
  if (!access) return null;
  const roles = access.roles ?? [];
  const stores = state.stores ?? [];
  const memberships = (access.memberships ?? []).filter((member) => member.user_id === userID);
  const grants = (access.grants ?? []).filter((grant) => grant.user_id === userID);
  const selectedRole = roles.find((role) => role.id === roleID);
  const isMember = memberships.some((member) => member.store_id === storeID);
  const storeName = (id: string) => stores.find((store) => store.id === id)?.name ?? id;

  async function run(command: Command, onSuccess?: () => void) {
    setError("");
    try {
      await execute(command);
      onSuccess?.();
    } catch (error) {
      setError(error instanceof Error ? error.message : "ثبت تغییر ناموفق بود.");
    }
  }

  return <section className="panel detail">
    {error && <Alert severity="error">{error}</Alert>}
    <Stack spacing={2} component="section" aria-label="نقش‌های سفارشی">
      <h2>نقش‌های سفارشی</h2>
      <ul>{roles.map((role) => <li key={role.id}>
        <b>{role.name}</b> · {scopeName(role.scope)} · {role.permissions.map((key) => permissionNames[key] ?? key).join("، ")}
        <Button disabled={busy} onClick={() => setDraft(role)}>ویرایش نقش</Button>
      </li>)}</ul>
      {draft ? <Stack component="form" spacing={1} onSubmit={(event) => {
        event.preventDefault();
        void run({
          action: draft.id ? "role.update" : "role.create",
          target: draft.id, version: draft.version, name: draft.name,
          scope: draft.scope, permissions: draft.permissions,
        }, () => setDraft(null));
      }}>
        <TextField label="نام نقش سفارشی" required value={draft.name}
          onChange={(event) => setDraft({ ...draft, name: event.target.value })} />
        <TextField select label="محدوده نقش" value={draft.scope} disabled={!!draft.id}
          onChange={(event) => setDraft({ ...draft, scope: event.target.value as RoleScope, permissions: [] })}>
          <MenuItem value="global">بک‌آفیس</MenuItem><MenuItem value="store">فروشگاه</MenuItem>
        </TextField>
        <div>{access.registry[draft.scope].map((key) => <FormControlLabel key={key}
          label={permissionNames[key] ?? key}
          control={<Checkbox checked={draft.permissions.includes(key)} onChange={(_, checked) => {
            setDraft({ ...draft, permissions: checked ? [...draft.permissions, key] : draft.permissions.filter((value) => value !== key) });
          }} />} />)}</div>
        <Stack direction="row" spacing={1}>
          <Button type="submit" variant="contained" disabled={busy || draft.permissions.length === 0}>ذخیره نقش</Button>
          <Button type="button" disabled={busy} onClick={() => setDraft(null)}>انصراف</Button>
          {draft.id && <Button type="button" color="error" disabled={busy} onClick={() => void run({
            action: "role.delete", target: draft.id, version: draft.version,
          }, () => setDraft(null))}>حذف نقش</Button>}
        </Stack>
      </Stack> : <Button onClick={() => setDraft({ id: "", name: "", scope: "global", permissions: [], version: 0 })}>تعریف نقش</Button>}
    </Stack>
    <Stack spacing={2} component="section" aria-label="عضویت و انتساب">
      <h2>عضویت و انتساب</h2>
      <TextField select label="حساب کاربر" value={userID} onChange={(event) => { setUserID(event.target.value); setRoleID(""); setStoreID(""); }}>
        {(state.users ?? []).map((user) => <MenuItem key={user.id} value={user.id}>{user.name}</MenuItem>)}
      </TextField>
      {userID && <>
        <ul>{memberships.map((member) => <li key={member.store_id}>
          {storeName(member.store_id)}
          <Button disabled={busy} onClick={() => void run({ action: "membership.remove", target: userID, store_id: member.store_id }, () => setStoreID(""))}>
            لغو عضویت و نقش‌های فروشگاه
          </Button>
        </li>)}</ul>
        <ul>{grants.map((grant) => <li key={grant.role_id + ":" + (grant.store_id ?? "")}>
          {roles.find((role) => role.id === grant.role_id)?.name} · {grant.store_id ? storeName(grant.store_id) : "بک‌آفیس"}
          <Button disabled={busy} onClick={() => void run({
            action: "user.role-revoke", target: userID, value: grant.role_id, store_id: grant.store_id,
          })}>لغو نقش</Button>
        </li>)}</ul>
        <TextField select label="نقش برای انتساب" value={roleID} onChange={(event) => setRoleID(event.target.value)}>
          <MenuItem value="">انتخاب نشده</MenuItem>
          {roles.map((role) => <MenuItem key={role.id} value={role.id}>{role.name} · {scopeName(role.scope)}</MenuItem>)}
        </TextField>
        <TextField select label="فروشگاه برای عضویت یا نقش" value={storeID} onChange={(event) => setStoreID(event.target.value)}>
          <MenuItem value="">انتخاب نشده</MenuItem>
          {stores.map((store) => <MenuItem key={store.id} value={store.id}>{store.name}</MenuItem>)}
        </TextField>
        <Stack direction="row" spacing={1}>
          <Button disabled={busy || !storeID || isMember}
            onClick={() => void run({ action: "membership.add", target: userID, store_id: storeID })}>افزودن عضویت</Button>
          <Button variant="contained" disabled={busy || !selectedRole || (selectedRole.scope === "store" && !isMember)}
            onClick={() => void run({ action: "user.role-grant", target: userID, value: roleID, store_id: selectedRole?.scope === "store" ? storeID : undefined })}>انتساب نقش</Button>
        </Stack>
      </>}
    </Stack>
  </section>;
}
