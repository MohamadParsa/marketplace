import { useState } from "react";
import { Button, Switch } from "@mui/material";
import { Add } from "@mui/icons-material";
import { useWorkspace, can } from "../../shared/workspace";
import {
  ActionDialog,
  DataTable,
  PageHeader,
  Status,
  type Field,
} from "../../shared/components";
import { labels, money } from "../../shared/format";
import type { Command, Category } from "../../shared/types";
import { CategoryBindingsDialog } from "./shared-attributes";
import { Link } from "react-router-dom";
import { ProductQuickDialog } from "./product-workspace";
import {
  CatalogIcon,
  CatalogIconDialog,
  type IconTarget,
} from "./catalog-icon";
const roleOptions = [
  "super-admin",
  "operations",
  "support",
  "finance",
  "viewer",
].map((value) => ({ value, label: value }));
export function ManagementPage({
  section,
}: {
  section:
    | "users"
    | "stores"
    | "categories"
    | "brands"
    | "products"
    | "pricing"
    | "shipping";
}) {
  const { state, role, execute } = useWorkspace();
  const [search, setSearch] = useState("");
  const [iconTarget, setIconTarget] = useState<IconTarget>();
  const [bindingTarget, setBindingTarget] = useState<Category>();
  const [productTarget, setProductTarget] = useState<{
    id: string;
    tab: number;
  }>();
  const [dialog, setDialog] = useState<{
    title: string;
    fields: Field[];
    command: Command;
  } | null>(null);
  const editable = can(role, "management");
  const open = (title: string, command: Command, fields: Field[] = []) =>
    setDialog({ title, fields, command });
  const titles = {
    users: "کاربران و نقش‌ها",
    stores: "فروشگاه‌ها",
    categories: "دسته‌ها و خصوصیات",
    brands: "برندها",
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
          (role === "super-admin" || (section !== "users" && editable)) ? (
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
                render: (u) => <span className="code">{u.role}</span>,
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
                              default: u.role,
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
                    disabled={!editable}
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
              {
                title: "دسته",
                render: (c) => (
                  <div className="catalog-name">
                    <CatalogIcon
                      kind="category"
                      name={c.name}
                      iconId={c.icon_id}
                    />
                    <b>{c.name}</b>
                  </div>
                ),
              },
              {
                title: "خصوصیات",
                render: (c) =>
                  (c.bindings ?? [])
                    .map(
                      (b) =>
                        state.attributes?.find((a) => a.id === b.attribute_id)
                          ?.name,
                    )
                    .filter(Boolean)
                    .join("، ") || "بدون انتساب صریح",
              },
              {
                title: "وضعیت",
                render: (c) => (c.archived ? "آرشیوشده" : "فعال"),
              },
              {
                title: "عملیات",
                render: (c) => (
                  <>
                    <Button
                      disabled={c.archived}
                      onClick={() => setBindingTarget(c)}
                    >
                      تنظیم خصوصیات
                    </Button>
                    <Button
                      disabled={!editable || c.archived}
                      onClick={() => setIconTarget({ ...c, kind: "category" })}
                    >
                      آیکن
                    </Button>
                    <Button
                      disabled={!editable || c.archived}
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
                  </>
                ),
              },
            ]}
          />
        )}
        {section === "brands" && (
          <DataTable
            rows={filtered(state.brands ?? [])}
            columns={[
              {
                title: "برند",
                render: (b) => (
                  <div className="catalog-name">
                    <CatalogIcon
                      kind="brand"
                      name={b.name}
                      iconId={b.icon_id}
                    />
                    <b>{b.name}</b>
                  </div>
                ),
              },
              {
                title: "وضعیت",
                render: (b) => (b.active ? "فعال" : "غیرفعال"),
              },
              {
                title: "عملیات",
                render: (b) => (
                  <Button
                    disabled={!editable || !b.active}
                    onClick={() => setIconTarget({ ...b, kind: "brand" })}
                  >
                    آیکن
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
                    <Link to={`/products/${p.id}`}>جزئیات محصول</Link>
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
                      onClick={() => setProductTarget({ id: p.id, tab: 0 })}
                    >
                      تنوع‌ها
                    </Button>
                    <Button
                      onClick={() => setProductTarget({ id: p.id, tab: 1 })}
                    >
                      قیمت و موجودی
                    </Button>
                    <Button
                      disabled={!editable || p.status === "published"}
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
                      disabled={!editable || p.status === "rejected"}
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
                    disabled={role !== "super-admin"}
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
                      disabled={!editable}
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
      {iconTarget && (
        <CatalogIconDialog
          target={iconTarget}
          close={() => setIconTarget(undefined)}
        />
      )}
      {productTarget && (
        <ProductQuickDialog
          id={productTarget.id}
          initialTab={productTarget.tab}
          close={() => setProductTarget(undefined)}
        />
      )}
      {bindingTarget && (
        <CategoryBindingsDialog
          key={bindingTarget.id}
          category={bindingTarget}
          close={() => setBindingTarget(undefined)}
        />
      )}
      {dialog && (
        <ActionDialog
          key={dialog.command.action + dialog.command.target}
          title={dialog.title}
          open
          fields={dialog.fields}
          onClose={() => setDialog(null)}
          onSubmit={(values) =>
            execute({
              ...dialog.command,
              ...values,
              amount: values.amount ? Number(values.amount) : undefined,
            })
          }
        />
      )}
    </>
  );
}
