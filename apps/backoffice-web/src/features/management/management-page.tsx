import { CreateButton } from "../../shared/create-button";
import { TextField } from "@mui/material";
import { FilterPanel } from "../../shared/filter-panel";
import {
  ProductCatalogFilters,
  filterProductRows,
  emptyProductFilters,
} from "./product-catalog-filters";
import { WalletCreditDialog } from "./wallet-credit-dialog";
import type { User } from "../../shared/types";
import type { WalletCreditResult } from "../../../../../packages/commerce-client/types";
import { CategorySalesDialog } from "../commerce/sales-methods-page";
import { useEffect, useState } from "react";
import { Alert, Button, Switch } from "@mui/material";
import { Add } from "@mui/icons-material";

import { useWorkspace, can } from "../../shared/workspace";
import {
  ActionDialog,
  DataTable,
  PageHeader,
  Status,
  type Field,
} from "../../shared/components";
import { labels, money, recordNumber } from "../../shared/format";
import type { Command, Category, Product } from "../../shared/types";
import { CategoryBindingsDialog } from "./shared-attributes";
import { Link, useNavigate, useSearchParams } from "react-router-dom";
import { categoryPath } from "./category-tree";
import { ProductQuickDialog } from "./product-workspace";
import { ProductThumbnail } from "./product-thumbnail";
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
  onCatalogEdit,
}: {
  onCatalogEdit?: (row?: Product) => void;
  section:
    | "users"
    | "stores"
    | "categories"
    | "brands"
    | "products"
    | "pricing"
    | "shipping";
}) {
  const navigate = useNavigate();
  const [params] = useSearchParams();
  const { state, role, execute, busy, surface, can: allowed } = useWorkspace();
  const [creditTarget, setCreditTarget] = useState<User>();
  const [walletUpdates, setWalletUpdates] = useState<
    Record<string, WalletCreditResult>
  >({});
  const categoryParam = params.get("category") ?? "";
  useEffect(
    () =>
      setProductFilters((previous) => ({
        ...previous,
        category: categoryParam,
      })),
    [categoryParam],
  );
  const [policyError, setError] = useState("");
  const [salesCategory, setSalesCategory] = useState<string>();
  const [search, setSearch] = useState("");
  const [productFilters, setProductFilters] = useState({
    ...emptyProductFilters,
    category: params.get("category") ?? "",
  });
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
    pricing: "راهنمای قیمت و کمیسیون",
    shipping: "ارسال و تقویم کاری",
  };
  const create = () =>
    section === "categories"
      ? navigate("/categories?create=category")
      : section === "users"
        ? open("کاربر جدید", { action: "user.create" }, [
            { name: "name", label: "نام کاربر" },
            { name: "phone", label: "شماره موبایل (۱۱ رقم)" },
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
  const [searchDraft, setSearchDraft] = useState(search);
  return (
    <>
      <PageHeader
        title={titles[section]}
        description={section === "pricing" ? "مسیر تنظیم قیمت فروش، مارکاپ و وضعیت تعریف قواعد کمیسیون" : "مشاهده، بررسی و مدیریت اطلاعات با ثبت تاریخچه تغییرات"}
        action={
          ["users", "stores", "categories", "products"].includes(section) &&
          (section === "products"
            ? surface === "backoffice" &&
              allowed("catalog.product.create", "management")
            : role === "super-admin" || (section !== "users" && editable)) ? (
            <CreateButton
              variant="contained"
              startIcon={<Add />}
              onClick={() =>
                section === "products" ? onCatalogEdit?.() : create()
              }
            >
              افزودن{" "}
              {section === "users"
                ? "کاربر"
                : section === "stores"
                  ? "فروشگاه"
                  : section === "products"
                    ? "محصول"
                    : "دسته"}
            </CreateButton>
          ) : undefined
        }
      />
      {policyError && <Alert severity="error">{policyError}</Alert>}
      <section className="panel">
        {section !== "products" && section !== "pricing" && (
          <FilterPanel onSearch={() => setSearch(searchDraft)}>
            <TextField
              label="جست‌وجو"

              placeholder="جست‌وجو در این بخش…"
              value={searchDraft}
              onChange={(e) => setSearchDraft(e.target.value)}
            />
          </FilterPanel>
        )}
        {section === "users" && (
          <DataTable
            searchable={false}
            rows={filtered(state.users ?? [])}
            columns={[
              {
                title: "نام و شناسه",
                render: (u) => (
                  <>
                    <b>{u.name}</b>
                    <small>{recordNumber(state, "user", u.id)}</small>
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
                      disabled={
                        !u.active ||
                        !allowed("commerce.wallet.credit", "finance")
                      }
                      onClick={() => setCreditTarget(u)}
                    >
                      شارژ کیف پول
                    </Button>
                    {walletUpdates[u.id] && (
                      <small>
                        موجودی پس از شارژ: {walletUpdates[u.id].balance_toman}{" "}
                        تومان
                      </small>
                    )}
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
            searchable={false}
            rows={filtered(state.stores ?? [])}
            columns={[
              {
                title: "فروشگاه",
                render: (s) => (
                  <>
                    <b>{s.name}</b>
                    <small>{recordNumber(state, "store", s.id)}</small>
                  </>
                ),
              },
              { title: "مالک", render: (s) => s.owner },
              { title: "شهر", render: (s) => s.city },
              { title: "مدل کاتالوگ", render: (s) => labels[s.model] },
              {
                title: "بررسی محتوای فروشنده",
                render: (s) => (
                  <select
                    className="content-policy-select"
                    aria-label={"سیاست محتوا " + s.name}
                    disabled={
                      busy || !allowed("store.content-policy.set", "management")
                    }
                    value={s.content_policy || "review"}
                    onChange={(e) => {
                      setError("");
                      void execute({
                        action: "store.content-policy.set",
                        target: s.id,
                        version: s.content_policy_version ?? 0,
                        value: e.target.value,
                        idempotency_key: crypto.randomUUID(),
                      }).catch((e) => setError(e.message));
                    }}
                  >
                    <option value="review">نیازمند تأیید</option>
                    <option value="direct">انتشار مستقیم</option>
                  </select>
                ),
              },
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
            searchable={false}
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
                    <div>
                      <b>{c.name}</b>
                      <small>
                        {categoryPath(state.categories ?? [], c.id)}
                      </small>
                      <small>
                        {c.display_priority == null
                          ? "بدون اولویت"
                          : `اولویت نمایش: ${c.display_priority.toLocaleString("fa-IR")}`}
                      </small>
                    </div>
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
                      component={Link}
                      to={`/categories?category=${encodeURIComponent(c.id)}`}
                      disabled={c.archived || !editable}
                    >
                      ویرایش
                    </Button>
                    <Button
                      disabled={c.archived}
                      onClick={() => setBindingTarget(c)}
                    >
                      تنظیم خصوصیات
                    </Button>
                    <Button onClick={() => setSalesCategory(c.id)}>
                      مارکاپ روش فروش
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
            searchable={false}
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
          <>
            <ProductCatalogFilters
              state={state}
              value={productFilters}
              onChange={setProductFilters}
            />
            <DataTable
              searchable={false}
              key={JSON.stringify(productFilters)}
              rows={filtered(filterProductRows(state, productFilters))}
              columns={[
                {
                  title: "محصول",
                  render: (p) => (
                    <>
                      <div className="product-grid-name">{surface === "backoffice" && <ProductThumbnail imageID={p.image_id || p.gallery_ids?.[0]} name={p.name} />}<b>{p.name}</b></div>
                      {p.content_revision?.state === "pending" && (
                        <small>
                          ویرایش در انتظار بررسی؛ نسخه عمومی قبلی محفوظ است.
                        </small>
                      )}
                      <small>
                        نسخه {p.version} · {p.public_code || "کد در حال تخصیص"}
                      </small>
                    </>
                  ),
                },
                {
                  title: "دسته",
                  render: (p) =>
                    state.categories?.find((c) => c.id === p.category_id)?.name,
                },
                { title: "نام انگلیسی", render: (p) => p.english_name || "—" },
                {
                  title: "برند",
                  render: (p) =>
                    state.brands?.find((b) => b.id === p.brand_id)?.name || "—",
                },
                {
                  title: "فعال‌بودن",
                  render: (p) => (p.active === false ? "غیرفعال" : "فعال"),
                },
                { title: "نوع", render: (p) => labels[p.model] },
                { title: "وضعیت", render: (p) => <Status value={p.status} /> },
                {
                  title: "عملیات",
                  render: (p) => (
                    <>
                      <Button component={Link} to={`/products/${p.id}`}>جزئیات محصول</Button>
                      <Button
                        onClick={() => onCatalogEdit?.(p)}
                        disabled={
                          surface !== "backoffice" ||
                          !allowed("catalog.product.update", "management")
                        }
                      >
                        ویرایش
                      </Button>
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
                        disabled={
                          !allowed("product.publish", "management") ||
                          p.status === "published" ||
                          p.content_revision?.state === "pending"
                        }
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
                        disabled={
                          !allowed("product.reject", "management") ||
                          p.status === "rejected" ||
                          p.content_revision?.state === "pending"
                        }
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
          </>
        )}
        {section === "pricing" && <section>
          <h2>قیمت فروش و قرارداد فروشگاه</h2>
          <p>قیمت پایه و موجودی برای عرضهٔ هر فروشگاه در «محصولات و محتوا ← قیمت و موجودی» تنظیم می‌شوند. مارکاپ هر روش فروش از تنظیمات همان دسته و روش فروش خوانده می‌شود.</p>
          <Alert severity="info">این صفحه قرارداد فروشگاه نیست. تعریف و ویرایش قانون کمیسیون، پس از تعیین قواعد قرارداد و تسویه در دسترس قرار می‌گیرد.</Alert>
          <div className="toolbar"><Button component={Link} to="/products">قیمت و موجودی محصولات</Button><Button component={Link} to="/categories">مارکاپ دسته‌ها</Button><Button component={Link} to="/sales-methods">روش‌های فروش</Button></div>
        </section>}
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
              searchable={false}
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

      {creditTarget && (
        <WalletCreditDialog
          key={creditTarget.id}
          user={creditTarget}
          close={() => setCreditTarget(undefined)}
          onCredited={(result) =>
            setWalletUpdates((old) => ({ ...old, [result.user_id]: result }))
          }
        />
      )}
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
      {salesCategory && (
        <CategorySalesDialog
          id={salesCategory}
          close={() => setSalesCategory(undefined)}
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
