import { StoreEligibilityEditor } from "./store-eligibility-editor";
import { useEffect, useRef, useState } from "react";
import {
  Alert,
  Button,
  Checkbox,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  LinearProgress,
  TextField,
} from "@mui/material";
import { StorefrontOutlined } from "@mui/icons-material";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { Link } from "react-router-dom";
import { useWorkspace } from "../../shared/workspace";
import { ApiError } from "../../shared/api";
import { DataTable, PageHeader } from "../../shared/components";
import type {
  Store,
  StoreProfile,
  StoreSaleMethodChoice,
} from "../../shared/types";
import type { Province } from "../../../../../packages/commerce-client/types";

type StoreMember = {
  user_id: string;
  name: string;
  active: boolean;
  roles: string[];
};
type ManagedStore = Store & { version: number; members?: StoreMember[] };
type Filter = {
  name: string;
  active: string;
  province: string;
  city: string;
  member: string;
  sales_method: string;
};
type Directory = {
  can_eligibility: boolean;
  sales_methods: StoreSaleMethodChoice[];
  rows: ManagedStore[];
  total: number;
  page: number;
  size: number;
  can_write: boolean;
  can_media: boolean;
  can_members: boolean;
  can_users: boolean;
  geography: Province[];
};
const defaults: Filter = {
  name: "",
  active: "",
  province: "",
  city: "",
  member: "",
  sales_method: "",
};
const apiRoot =
  (import.meta.env.VITE_API_BASE_URL ?? "") + "/backoffice/stores";
async function request<T>(
  token: string,
  path: string,
  data?: unknown,
  key?: string,
  method = "POST",
  signal?: AbortSignal,
): Promise<T> {
  const r = await fetch(apiRoot + path, {
    method,
    signal,
    cache: "no-store",
    headers: {
      Authorization: `Bearer ${token}`,
      ...(data instanceof FormData
        ? {}
        : { "Content-Type": "application/json" }),
      ...(key ? { "Idempotency-Key": key } : {}),
    },
    body:
      data instanceof FormData
        ? data
        : data === undefined
          ? undefined
          : JSON.stringify(data),
  });
  if (!r.ok) {
    const body = await r.json().catch(() => null);
    throw new ApiError(r.status, body?.code);
  }
  return (
    r.headers.get("Content-Type")?.includes("image/") ? r.blob() : r.json()
  ) as Promise<T>;
}
const message = (e: unknown) =>
  e instanceof Error ? e.message : "ذخیره انجام نشد؛ دوباره تلاش کنید.";
function StorePhoto({ row, token }: { row: ManagedStore; token: string }) {
  const [src, setSrc] = useState("");
  useEffect(() => {
    let alive = true,
      url = "";
    setSrc("");
    if (row.profile?.image_id)
      void request<Blob>(
        token,
        `/${encodeURIComponent(row.id)}/image`,
        undefined,
        undefined,
        "GET",
      )
        .then((blob) => {
          url = URL.createObjectURL(blob);
          if (alive) setSrc(url);
          else URL.revokeObjectURL(url);
        })
        .catch(() => {});
    return () => {
      alive = false;
      if (url) URL.revokeObjectURL(url);
    };
  }, [row.id, row.profile?.image_id, token]);
  return src ? (
    <img
      className="store-photo"
      src={src}
      alt={`تصویر ${row.name}`}
      onError={() => setSrc("")}
    />
  ) : (
    <span
      className="store-photo fallback"
      aria-label={`تصویر پیش‌فرض ${row.name}`}
    >
      <StorefrontOutlined />
    </span>
  );
}
export function StoresPage() {
  const { authToken, can } = useWorkspace(),
    cache = useQueryClient();
  const [draft, setDraft] = useState(defaults),
    [filter, setFilter] = useState(defaults),
    [page, setPage] = useState(0),
    [size, setSize] = useState(20),
    [editing, setEditing] = useState<ManagedStore | "new">();
  const query = useQuery<Directory>({
    queryKey: ["managed-stores", authToken, filter, page, size],
    placeholderData: (old) => old,
    queryFn: ({ signal }) =>
      request<Directory>(
        authToken,
        "/query",
        { filter: { ...filter, page, size } },
        undefined,
        "POST",
        signal,
      ),
  });
  const data = query.data;
  const refresh = () => {
    void cache.invalidateQueries({ queryKey: ["managed-stores"] });
    void cache.invalidateQueries({ queryKey: ["workspace"] });
    void cache.invalidateQueries({ queryKey: ["product"] });
    void cache.invalidateQueries({ queryKey: ["variant-resolution"] });
  };
  if (query.isError && !data)
    return (
      <Alert
        severity="error"
        action={<Button onClick={() => void query.refetch()}>تلاش مجدد</Button>}
      >
        {message(query.error)}
      </Alert>
    );
  return (
    <>
      <PageHeader
        title="فروشگاه‌ها"
        description="پروفایل، تصویر و اطلاعات ارتباطی خصوصی فروشگاه؛ عضویت کاربران از فرم کاربران مدیریت می‌شود."
        action={
          <Button
            variant="contained"
            disabled={!data?.can_write}
            onClick={() => setEditing("new")}
          >
            فروشگاه جدید
          </Button>
        }
      />
      <section className="panel store-directory">
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
                ["name", "نام فروشگاه"],
                ["city", "شهر"],
                ["member", "عضو فروشگاه"],
              ] as const
            )
              .filter(([key]) => key !== "member" || data?.can_members)
              .map(([key, label]) => (
                <TextField
                  key={key}
                  size="small"
                  label={`فیلتر ${label}`}
                  value={draft[key]}
                  onChange={(e) =>
                    setDraft({ ...draft, [key]: e.target.value })
                  }
                />
              ))}
          </div>
          <div
            className="toolbar"
            role="group"
            aria-label="فیلتر وضعیت فروشگاه"
          >
            {[
              ["", "همه وضعیت‌ها"],
              ["true", "فعال"],
              ["false", "غیرفعال"],
            ].map(([id, name]) => (
              <Button
                key={id}
                aria-pressed={draft.active === id}
                variant={draft.active === id ? "contained" : "outlined"}
                onClick={() => setDraft({ ...draft, active: id })}
              >
                {name}
              </Button>
            ))}
          </div>
          <div
            className="toolbar"
            role="group"
            aria-label="فیلتر استان فروشگاه"
          >
            <strong>استان</strong>
            <Button
              aria-pressed={!draft.province}
              onClick={() => setDraft({ ...draft, province: "" })}
            >
              همه استان‌ها
            </Button>
            {data?.geography?.map((p) => (
              <Button
                key={p.id}
                aria-pressed={draft.province === p.name}
                variant={draft.province === p.name ? "contained" : "text"}
                onClick={() => setDraft({ ...draft, province: p.name })}
              >
                {p.name}
              </Button>
            ))}
          </div>
          <div
            className="toolbar"
            role="group"
            aria-label="فیلتر روش فروش فروشگاه"
          >
            <strong>روش فروش</strong>
            <Button
              aria-pressed={!draft.sales_method}
              onClick={() => setDraft({ ...draft, sales_method: "" })}
            >
              همه روش‌ها
            </Button>
            {data?.sales_methods?.map((m) => (
              <Button
                key={m.id}
                aria-pressed={draft.sales_method === m.id}
                variant={draft.sales_method === m.id ? "contained" : "text"}
                onClick={() => setDraft({ ...draft, sales_method: m.id })}
              >
                {m.name}
              </Button>
            ))}
          </div>
          <div className="toolbar">
            <Button
              variant="contained"
              type="submit"
              disabled={query.isFetching}
            >
              اعمال فیلترهای فروشگاه
            </Button>
            <Button
              onClick={() => {
                setDraft(defaults);
                setFilter(defaults);
                setPage(0);
              }}
            >
              پاک‌کردن فیلترهای فروشگاه
            </Button>
            <Button onClick={() => void query.refetch()}>بارگیری مجدد</Button>
          </div>
        </form>
        {query.isFetching && <LinearProgress aria-label="دریافت فروشگاه‌ها" />}
        {query.isError && (
          <Alert severity="error">{message(query.error)}</Alert>
        )}
        {data && (
          <fieldset className="store-grid-region" disabled={query.isFetching}>
            <DataTable
              rows={data.rows}
              remote={{
                filterKey: JSON.stringify(filter),
                controls: {
                  page: data.page,
                  pages: Math.max(1, Math.ceil(data.total / data.size)),
                  size: data.size,
                  total: data.total,
                  setPage,
                  setSize,
                },
                exportRows: async (ids, signal) =>
                  (
                    await request<Directory>(
                      authToken,
                      "/export",
                      { filter, ids: ids ?? [] },
                      undefined,
                      "POST",
                      signal,
                    )
                  ).rows,
              }}
              columns={[
                {
                  title: "فروشگاه",
                  render: (s) => (
                    <div className="catalog-name">
                      <StorePhoto row={s} token={authToken} />
                      <div>
                        <b>{s.name}</b>
                        <small>
                          {s.model === "independent"
                            ? "کاتالوگ مستقل"
                            : "کاتالوگ مشترک"}
                        </small>
                      </div>
                    </div>
                  ),
                },
                {
                  title: "وضعیت",
                  render: (s) => (s.active ? "فعال" : "غیرفعال"),
                },
                {
                  title: "موقعیت",
                  render: (s) =>
                    [s.profile?.province, s.city].filter(Boolean).join(" / ") ||
                    "ثبت نشده",
                },
                {
                  title: "اعضا",
                  render: (s) =>
                    data.can_members
                      ? (s.members ?? []).map((m) => m.name).join("، ") ||
                        "بدون عضو"
                      : "بدون دسترسی مشاهده اعضا",
                },
                {
                  title: "عملیات",
                  render: (s) => (
                    <Button onClick={() => setEditing(s)}>
                      مشاهده و ویرایش
                    </Button>
                  ),
                },
              ]}
            />
          </fieldset>
        )}
      </section>
      {editing && data && (
        <StoreEditor
          key={typeof editing === "string" ? "new" : editing.id}
          row={editing === "new" ? undefined : editing}
          directory={data}
          token={authToken}
          policyAllowed={can("store.content-policy.set", "management")}
          close={() => setEditing(undefined)}
          saved={refresh}
        />
      )}
    </>
  );
}
function StoreEditor({
  row,
  directory,
  token,
  policyAllowed,
  close,
  saved,
}: {
  row?: ManagedStore;
  directory: Directory;
  token: string;
  policyAllowed: boolean;
  close: () => void;
  saved: () => void;
}) {
  const { execute } = useWorkspace();
  const [current, setCurrent] = useState(row),
    [name, setName] = useState(row?.name ?? ""),
    [active, setActive] = useState(row?.active ?? true),
    [city, setCity] = useState(row?.city ?? ""),
    [profile, setProfile] = useState<StoreProfile>(row?.profile ?? {}),
    [busy, setBusy] = useState(false),
    [error, setError] = useState(""),
    [notice, setNotice] = useState(""),
    [key, setKey] = useState(crypto.randomUUID()),
    [file, setFile] = useState<File>(),
    [preview, setPreview] = useState(""),
    [imageKey, setImageKey] = useState(crypto.randomUUID()),
    [policy, setPolicy] = useState(row?.content_policy ?? "review");
  const fileInput = useRef<HTMLInputElement>(null);
  const province = directory.geography?.find(
      (p) => p.name === profile.province,
    ),
    cities = province?.cities ?? [];
  const change = (field: keyof StoreProfile, value: string) => {
    setProfile({ ...profile, [field]: value });
    setKey(crypto.randomUUID());
  };
  useEffect(() => {
    if (!file) {
      setPreview("");
      return;
    }
    const url = URL.createObjectURL(file);
    setPreview(url);
    return () => URL.revokeObjectURL(url);
  }, [file]);
  async function save() {
    if (busy || !directory.can_write) return;
    setBusy(true);
    setError("");
    setNotice("");
    try {
      const { image_id, ...fields } = profile;
      void image_id;
      const value = await request<ManagedStore>(
        token,
        "/save",
        {
          id: current?.id ?? "",
          version: current?.version ?? 0,
          name,
          active,
          city,
          profile: {
            description: "",
            email: "",
            province: "",
            address: "",
            postal_code: "",
            phone: "",
            mobile: "",
            ...fields,
          },
        },
        key,
      );
      setCurrent(value);
      setProfile(value.profile ?? {});
      setName(value.name);
      setCity(value.city);
      setActive(value.active);
      setKey(crypto.randomUUID());
      setNotice("فروشگاه ذخیره شد.");
      saved();
    } catch (e) {
      setError(message(e));
    } finally {
      setBusy(false);
    }
  }
  const profileField = (
    field: keyof StoreProfile,
    label: string,
    limit: number,
    multiline = false,
  ) => (
    <TextField
      key={field}
      label={label}
      value={profile[field] ?? ""}
      multiline={multiline}
      minRows={multiline ? 3 : undefined}
      inputProps={{ maxLength: limit }}
      onChange={(e) => change(field, e.target.value)}
    />
  );
  return (
    <Dialog
      open
      fullWidth
      maxWidth="md"
      onClose={() => {
        if (!busy) close();
      }}
    >
      <DialogTitle>
        {current ? `ویرایش فروشگاه · ${current.name}` : "فروشگاه جدید"}
      </DialogTitle>
      <DialogContent className="store-editor">
        {error && <Alert severity="error">{error}</Alert>}
        {notice && <Alert severity="success">{notice}</Alert>}
        {!directory.can_write && (
          <Alert severity="info">
            مجوز ویرایش ندارید؛ اطلاعات فقط خواندنی است.
          </Alert>
        )}
        <form
          id="store-profile-form"
          onSubmit={(e) => {
            e.preventDefault();
            void save();
          }}
        >
          <fieldset
            disabled={busy || !directory.can_write}
            className="store-form-grid"
          >
            <TextField
              label="نام فروشگاه"
              required
              value={name}
              inputProps={{ maxLength: 160 }}
              onChange={(e) => {
                setName(e.target.value);
                setKey(crypto.randomUUID());
              }}
            />
            <label className="store-active">
              <Checkbox
                checked={active}
                onChange={(e) => {
                  setActive(e.target.checked);
                  setKey(crypto.randomUUID());
                }}
              />
              فروشگاه فعال است
            </label>
            <div className="store-form-wide">
              {profileField("description", "توضیحات فروشگاه", 4000, true)}
            </div>
            {profileField("email", "ایمیل فروشگاه", 254)}
            {profileField("phone", "تلفن ثابت فروشگاه", 16)}
            {profileField("mobile", "موبایل فروشگاه", 11)}
            <label className="store-geo-field">
              استان
              <select
                aria-label="استان فروشگاه"
                value={profile.province ?? ""}
                onChange={(e) => {
                  change("province", e.target.value);
                  setCity("");
                }}
              >
                <option value="">انتخاب استان</option>
                {directory.geography?.map((p) => (
                  <option key={p.id} value={p.name}>
                    {p.name}
                  </option>
                ))}
              </select>
            </label>
            <label className="store-geo-field">
              شهر
              <select
                aria-label="شهر فروشگاه"
                disabled={!province}
                required={!!profile.province}
                value={cities.some((c) => c.name === city) ? city : ""}
                onChange={(e) => {
                  setCity(e.target.value);
                  setKey(crypto.randomUUID());
                }}
              >
                <option value="">انتخاب شهر</option>
                {cities.map((c) => (
                  <option key={c.id} value={c.name}>
                    {c.name}
                  </option>
                ))}
              </select>
            </label>
            {!profile.province && current?.city && (
              <p className="store-form-wide muted">
                شهر قدیمی «{current.city}» تا انتخاب استان حفظ می‌شود؛ استان از
                روی نام شهر حدس زده نمی‌شود.
              </p>
            )}
            <div className="store-form-wide">
              {profileField("address", "آدرس کامل فروشگاه", 1500, true)}
            </div>
            {profileField("postal_code", "کد پستی فروشگاه", 10)}
          </fieldset>
        </form>
        {current && (
          <>
            <section className="store-detail-box">
              <h3>اطلاعات موجود</h3>
              <p>شناسه ثابت · {current.id}</p>
              <p>
                مالک ثبت‌شده: {current.owner || "ثبت نشده"} · وقفه تأمین:{" "}
                {current.lead_days} روز {current.preparation?.day_basis === "calendar" ? "تقویمی" : "کاری"} · مدل:{" "}
                {current.model === "independent" ? "مستقل" : "مشترک"}
              </p>
              <p>این اطلاعات با ویرایش پروفایل بازنویسی نمی‌شوند.</p>
              <Button component={Link} to="/store-preparation">تنظیم وقفه و تقویم فروشگاه</Button>
            </section>
            <section className="store-detail-box">
              <h3>تصویر فروشگاه</h3>
              {preview ? (
                <img
                  className="store-photo large"
                  src={preview}
                  alt="پیش‌نمایش تصویر فروشگاه"
                />
              ) : (
                <StorePhoto row={current} token={token} />
              )}
              <p>PNG، JPEG یا WebP؛ حداکثر۱MiB و هر ضلع۱۰۲۴پیکسل.</p>
              <input
                ref={fileInput}
                type="file"
                aria-label="انتخاب تصویر فروشگاه"
                disabled={busy || !directory.can_media}
                accept="image/png,image/jpeg,image/webp"
                onChange={(e) => {
                  setFile(e.target.files?.[0]);
                  setImageKey(crypto.randomUUID());
                }}
              />
              <div className="toolbar">
                <Button
                  disabled={busy || !directory.can_media || !file}
                  onClick={async () => {
                    if (!file || busy) return;
                    setBusy(true);
                    setError("");
                    try {
                      const form = new FormData();
                      form.set("version", String(current.version));
                      form.set("file", file);
                      const next = await request<ManagedStore>(
                        token,
                        `/${current.id}/image`,
                        form,
                        imageKey,
                      );
                      setCurrent(next);
                      setProfile((p) => ({
                        ...p,
                        image_id: next.profile?.image_id,
                      }));
                      setFile(undefined);
                      if (fileInput.current) fileInput.current.value = "";
                      setImageKey(crypto.randomUUID());
                      setNotice("تصویر فروشگاه ذخیره شد.");
                      saved();
                    } catch (e) {
                      setError(message(e));
                    } finally {
                      setBusy(false);
                    }
                  }}
                >
                  ذخیره تصویر فروشگاه
                </Button>
                <Button
                  disabled={
                    busy || !directory.can_media || !current.profile?.image_id
                  }
                  onClick={async () => {
                    setBusy(true);
                    setError("");
                    try {
                      const next = await request<ManagedStore>(
                        token,
                        `/${current.id}/image`,
                        { version: current.version },
                        imageKey,
                        "DELETE",
                      );
                      setCurrent(next);
                      setProfile((p) => ({ ...p, image_id: undefined }));
                      setImageKey(crypto.randomUUID());
                      setNotice(
                        "تصویر حذف شد؛ نشان پیش‌فرض نمایش داده می‌شود.",
                      );
                      saved();
                    } catch (e) {
                      setError(message(e));
                    } finally {
                      setBusy(false);
                    }
                  }}
                >
                  حذف تصویر فروشگاه
                </Button>
              </div>
            </section>
            <StoreEligibilityEditor
              store={current}
              locked={busy}
              onBusyChange={setBusy}
              methods={directory.sales_methods ?? []}
              allowed={directory.can_eligibility}
              save={(input, key) =>
                request<ManagedStore>(
                  token,
                  `/${current.id}/eligibility`,
                  input,
                  key,
                )
              }
              saved={(value) => {
                setCurrent(value);
                saved();
              }}
            />
            <section className="store-detail-box">
              <h3>سیاست انتشار محتوا</h3>
              <div className="toolbar">
                {(
                  [
                    ["review", "نیازمند تأیید"],
                    ["direct", "انتشار مستقیم"],
                  ] as const
                ).map(([value, label]) => (
                  <Button
                    key={value}
                    aria-pressed={policy === value}
                    disabled={busy || !policyAllowed}
                    variant={policy === value ? "contained" : "outlined"}
                    onClick={() => setPolicy(value)}
                  >
                    {label}
                  </Button>
                ))}
                <Button
                  disabled={
                    busy || !policyAllowed || policy === current.content_policy
                  }
                  onClick={async () => {
                    setBusy(true);
                    setError("");
                    try {
                      await execute({
                        action: "store.content-policy.set",
                        target: current.id,
                        version: current.content_policy_version ?? 0,
                        value: policy,
                        idempotency_key: crypto.randomUUID(),
                      });
                      setCurrent({
                        ...current,
                        content_policy: policy,
                        content_policy_version:
                          (current.content_policy_version ?? 0) + 1,
                      });
                      saved();
                      setNotice("سیاست محتوا ثبت شد.");
                    } catch (e) {
                      setError(message(e));
                    } finally {
                      setBusy(false);
                    }
                  }}
                >
                  ثبت سیاست محتوا
                </Button>
              </div>
            </section>
            <section className="store-detail-box">
              <h3>کاربران متصل به فروشگاه</h3>
              {directory.can_members ? (
                <>
                  {current.members?.length ? (
                    current.members.map((m) => (
                      <div key={m.user_id} className="user-access-row">
                        <div>
                          {directory.can_users ? (
                            <Link
                              to={`/users?user=${encodeURIComponent(m.user_id)}`}
                            >
                              {m.name}
                            </Link>
                          ) : (
                            <b>{m.name}</b>
                          )}
                          <small>
                            {m.active ? "کاربر فعال" : "کاربر غیرفعال"} ·{" "}
                            {m.roles.length
                              ? m.roles.join("، ")
                              : "عضو فروشگاه؛ بدون نقش فروشگاهی"}
                          </small>
                        </div>
                      </div>
                    ))
                  ) : (
                    <p>هنوز کاربری به این فروشگاه متصل نیست.</p>
                  )}
                  <p>مدیریت عضویت و نقش از فرم کاربران انجام می‌شود.</p>
                </>
              ) : (
                <p>مجوز مشاهده اعضای این فروشگاه را ندارید.</p>
              )}
            </section>
          </>
        )}
      </DialogContent>
      <DialogActions>
        <Button disabled={busy} onClick={close}>
          بستن
        </Button>
        <Button
          type="submit"
          form="store-profile-form"
          variant="contained"
          disabled={busy || !directory.can_write}
        >
          ذخیره فروشگاه
        </Button>
      </DialogActions>
    </Dialog>
  );
}
