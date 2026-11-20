import { useState } from "react";
import { Button, Switch } from "@mui/material";
import { Add } from "@mui/icons-material";
import { useWorkspace } from "../../shared/workspace";
import { AccessPanel } from "./access-panel";
import {
  ActionDialog,
  DataTable,
  PageHeader,
  Status,
  type Field,
} from "../../shared/components";
import { labels, money } from "../../shared/format";
import type { Command } from "../../shared/types";
const roleOptions = [
  "super-admin",
  "operations",
  "support",
  "finance",
  "viewer",
].map((value) => ({ value, label: value === "super-admin" ? "ادمین" : value }));
export function ManagementPage({
  section,
}: {
  section:
    "users" | "stores" | "categories" | "products" | "pricing" | "shipping";
}) {
  const { state, role, can, execute, requestAdminCode } = useWorkspace();
  const [search, setSearch] = useState("");
  const [dialog, setDialog] = useState<{
    title: string;
    fields: Field[];
    command: Command;
  } | null>(null);
  const open = (title: string, command: Command, fields: Field[] = []) =>
    setDialog({ title, fields, command: { ...command, idempotency_key: crypto.randomUUID() } });
  const titles = {
    users: "کاربران و نقش‌ها",
    stores: "فروشگاه‌ها",
    categories: "دسته‌ها و خصوصیات",
    products: "محصولات و بررسی محتوا",
    pricing: "قوانین کمیسیون",
    shipping: "ارسال و تقویم کاری",
  };
  const create = () =>
    section === "users"
      ? open("کاربر جدید", { action: "user.create" }, [
          { name: "name", label: "نام کاربر" },
          { name: "phone", label: "شماره حساب آزمون (۱۱ رقم)" },
          { name: "value", label: "نقش", options: roleOptions },
        ])
      : section === "stores"
        ? open("فروشگاه جدید", { action: "store.create" }, [
            { name: "name", label: "نام فروشگاه" },
            { name: "value", label: "شهر" },
          ])
        : open("دسته جدید", { action: "category.create" }, [
            { name: "name", label: "عنوان دسته" },
          ]);
  const filtered = <T,>(rows: T[]) =>
    rows.filter((row) =>
      JSON.stringify(row)
        .toLocaleLowerCase()
        .includes(search.toLocaleLowerCase()),
    );
  return (
    <>
      <PageHeader
        title={titles[section]}
        description="مشاهده، بررسی و مدیریت اطلاعات با ثبت تاریخچه تغییرات"
        action={
          ["users", "stores", "categories"].includes(section) &&
          (section === "users" ? role === "super-admin" : can(section === "stores" ? "store.create" : "category.create", "management")) ? (
            <Button variant="contained" startIcon={<Add />} onClick={create}>
              افزودن{" "}
              {section === "users"
                ? "کاربر"
                : section === "stores"
                  ? "فروشگاه"
                  : "دسته"}
            </Button>
          ) : undefined
        }
      />
      <section className="panel">
        <div className="toolbar">
          <input
            aria-label="جست‌وجو"
            className="search"
            placeholder="جست‌وجو در این بخش…"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
          />
          <span className="muted">اطلاعات آزمون · به‌روز از سرور</span>
        </div>
        {section === "users" && (
          <DataTable
            rows={filtered(state.users ?? [])}
            columns={[
              {
                title: "نام و شناسه",
                render: (u) => (
                  <>
                    <b>{u.name}</b>
                    <small>{u.id}</small>
                  </>
                ),
              },
              {
                title: "نقش",
                render: (u) => <span className="code">{u.role === "super-admin" ? "ادمین" : u.role || "بدون نقش پیش‌فرض"}</span>,
              },
              {
                title: "وضعیت",
                render: (u) => (u.active ? "فعال" : "غیرفعال"),
              },
              {
                title: "عملیات",
                render: (u) => (
                  <>
                    <Button
                      disabled={role !== "super-admin"}
                      onClick={() =>
                        open(
                          "تغییر نقش",
                          { action: "user.role", target: u.id },
                          [
                            {
                              name: "value",
                              label: "نقش جدید",
                              options: roleOptions,
                              default: u.role || "viewer",
                            },
                          ],
                        )
                      }
                    >
                      تغییر نقش
                    </Button>
                    <Button
                      color="warning"
                      disabled={role !== "super-admin"}
                      onClick={() =>
                        open(
                          "تغییر وضعیت کاربر",
                          { action: "user.toggle", target: u.id },
                          [{ name: "reason", label: "دلیل تغییر" }],
                        )
                      }
                    >
                      {u.active ? "غیرفعال‌سازی" : "فعال‌سازی"}
                    </Button>
                  </>
                ),
              },
            ]}
          />
        )}
        {section === "stores" && (
          <DataTable
            rows={filtered(state.stores ?? [])}
            columns={[
              {
                title: "فروشگاه",
                render: (s) => (
                  <>
                    <b>{s.name}</b>
                    <small>{s.id}</small>
                  </>
                ),
              },
              { title: "مالک", render: (s) => s.owner },
              { title: "شهر", render: (s) => s.city },
              { title: "مدل کاتالوگ", render: (s) => labels[s.model] },
              {
                title: "وضعیت",
                render: (s) => (
                  <span className={s.active ? "active-dot" : "muted"}>
                    {s.active ? "فعال" : "غیرفعال"}
                  </span>
                ),
              },
              {
                title: "عملیات",
                render: (s) => (
                  <Button
                    disabled={!can("store.toggle", "management")}
                    onClick={() =>
                      open(
                        "تغییر وضعیت فروشگاه",
                        { action: "store.toggle", target: s.id },
                        [{ name: "reason", label: "دلیل تغییر" }],
                      )
                    }
                  >
                    {s.active ? "توقف فعالیت" : "فعال‌سازی"}
                  </Button>
                ),
              },
            ]}
          />
        )}
        {section === "categories" && (
          <DataTable
            rows={filtered(state.categories ?? [])}
            columns={[
              { title: "دسته", render: (c) => <b>{c.name}</b> },
              {
                title: "خصوصیات",
                render: (c) =>
                  c.attributes.join("، ") || "هنوز تعریفی ثبت نشده",
              },
              {
                title: "وضعیت",
                render: (c) => (c.archived ? "آرشیوشده" : "فعال"),
              },
              {
                title: "عملیات",
                render: (c) => (
                  <Button
                    disabled={!can("category.archive", "management") || c.archived}
                    onClick={() =>
                      open(
                        "آرشیو دسته",
                        { action: "category.archive", target: c.id },
                        [{ name: "reason", label: "دلیل آرشیو" }],
                      )
                    }
                  >
                    آرشیو
                  </Button>
                ),
              },
            ]}
          />
        )}
        {section === "products" && (
          <DataTable
            rows={filtered(state.products ?? [])}
            columns={[
              {
                title: "محصول",
                render: (p) => (
                  <>
                    <b>{p.name}</b>
                    <small>
                      نسخه {p.version} · {p.id}
                    </small>
                  </>
                ),
              },
              {
                title: "دسته",
                render: (p) =>
                  state.categories?.find((c) => c.id === p.category_id)?.name,
              },
              { title: "نوع", render: (p) => labels[p.model] },
              { title: "وضعیت", render: (p) => <Status value={p.status} /> },
              {
                title: "عملیات",
                render: (p) => (
                  <>
                    <Button
                      disabled={!can("product.publish", "management") || p.status === "published"}
                      onClick={() =>
                        open(
                          "تأیید و انتشار محتوا",
                          {
                            action: "product.publish",
                            target: p.id,
                            version: p.version,
                          },
                          [{ name: "reason", label: "یادداشت بررسی" }],
                        )
                      }
                    >
                      انتشار
                    </Button>
                    <Button
                      color="error"
                      disabled={!can("product.reject", "management") || p.status === "rejected"}
                      onClick={() =>
                        open(
                          "رد محتوای محصول",
                          {
                            action: "product.reject",
                            target: p.id,
                            version: p.version,
                          },
                          [{ name: "reason", label: "دلیل رد" }],
                        )
                      }
                    >
                      رد
                    </Button>
                  </>
                ),
              },
            ]}
          />
        )}
        {section === "pricing" && (
          <DataTable
            rows={filtered(state.rules ?? [])}
            columns={[
              { title: "قانون", render: (r) => <b>{r.name}</b> },
              { title: "سطح", render: (r) => labels[r.level] },
              { title: "شیوه", render: (r) => labels[r.kind] },
              {
                title: "مقدار",
                render: (r) =>
                  r.kind === "percent" ? `${r.value}٪` : money(r.value),
              },
              {
                title: "فعال",
                render: (r) => (
                  <Switch
                    checked={r.active}
                    disabled={!can("rule.toggle")}
                    onChange={() =>
                      open(
                        "تغییر اعتبار قانون",
                        { action: "rule.toggle", target: r.id },
                        [{ name: "reason", label: "دلیل تغییر" }],
                      )
                    }
                  />
                ),
              },
            ]}
          />
        )}
        {section === "shipping" && (
          <>
            <div className="calendar-note">
              <b>تقویم کاری: شنبه تا پنج‌شنبه</b>
              <span>
                منطقه زمانی تهران · شروع مهلت از روز کاری بعد · تعطیلات ثبت‌شده
                حذف می‌شوند
              </span>
            </div>
            <DataTable
              rows={filtered(state.shipping ?? [])}
              columns={[
                { title: "روش ارسال", render: (s) => <b>{s.method}</b> },
                { title: "شهر مقصد", render: (s) => s.city },
                { title: "تعرفه", render: (s) => money(s.tariff) },
                { title: "مدت", render: (s) => `${s.days} روز کاری` },
                {
                  title: "عملیات",
                  render: (s) => (
                    <Button
                      disabled={!can("shipping.update", "management")}
                      onClick={() =>
                        open(
                          "ویرایش تعرفه",
                          { action: "shipping.update", target: s.id },
                          [
                            {
                              name: "amount",
                              label: "تعرفه (ریال)",
                              type: "number",
                              default: String(s.tariff),
                            },
                            { name: "reason", label: "دلیل تغییر" },
                          ],
                        )
                      }
                    >
                      ویرایش تعرفه
                    </Button>
                  ),
                },
              ]}
            />
          </>
        )}
      </section>
      {section === "users" && <AccessPanel />}
      {dialog && (
        <ActionDialog
          key={dialog.command.action + dialog.command.target}
          title={dialog.title}
          open
          fields={dialog.fields}
          requestAdminCode={dialog.command.action.startsWith("user.") ? requestAdminCode : undefined}
          onClose={() => setDialog(null)}
          onSubmit={(values, reauthentication) =>
            execute({
              ...dialog.command,
              ...values,
              amount: values.amount ? Number(values.amount) : undefined,
              reauthentication,
            })
          }
        />
      )}
    </>
  );
}
