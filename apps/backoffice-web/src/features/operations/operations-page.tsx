import { useState } from "react";
import { deliveryPromise, shippingCharge } from "../../../../../packages/commerce-client/shipping-presentation";
import { Link, useParams } from "react-router-dom";
import { Alert, Button } from "@mui/material";
import { useWorkspace } from "../../shared/workspace";
import {
  ActionDialog,
  DataTable,
  PageHeader,
  Status,
  type Field,
} from "../../shared/components";
import { date, money, number, labels } from "../../shared/format";
import type { Command } from "../../shared/types";
export function OperationsPage({
  section,
}: {
  section: "orders" | "support" | "assignments" | "notifications" | "reports";
}) {
  const { state, can, execute } = useWorkspace();
  const { id } = useParams();
  const [dialog, setDialog] = useState<{
    title: string;
    fields: Field[];
    command: Command;
  } | null>(null);
  const [search, setSearch] = useState("");
  const open = (
    title: string,
    command: Command,
    fields: Field[] = [{ name: "reason", label: "دلیل / یادداشت عملیات" }],
  ) => setDialog({ title, command, fields });
  const titles = {
    orders: "سفارش‌ها",
    support: "پشتیبانی و پرونده‌ها",
    assignments: "تأمین جایگزین",
    notifications: "اعلان‌ها",
    reports: "گزارش‌ها و تاریخچه",
  };
  const filter = <T,>(rows: T[]) =>
    rows.filter((row) =>
      JSON.stringify(row).toLowerCase().includes(search.toLowerCase()),
    );
  const seller = (storeId: string) =>
    state.stores?.find((s) => s.id === storeId)?.name ?? storeId;
  const product = (productId: string) =>
    state.products?.find((p) => p.id === productId)?.name ?? productId;
  const selected = state.orders?.find((o) => o.id === id);
  return (
    <>
      <PageHeader
        title={
          selected
            ? `جزئیات سفارش ${selected.purchase_number ?? selected.id}`
            : titles[section]
        }
        description="پیگیری مرحله‌های عملیات و بررسی تاریخچه؛ هر تصمیم با عامل و دلیل ثبت می‌شود."
      />
      {selected ? (
        <>
          <Link to="/orders">← بازگشت به سفارش‌ها</Link>
          <div className="detail-grid">
            <section className="panel detail">
              <h2>اطلاعات سفارش</h2>
              <dl>
                <dt>خرید چندفروشگاهی</dt>
                <dd>{selected.purchase_number ?? selected.purchase_id}</dd>
                <dt>فروشگاه</dt>
                <dd>
                  {selected.store_display_name || seller(selected.store_id)}
                </dd>
                <dt>شهر مقصد</dt>
                <dd>{selected.city}</dd>
                <dt>زمان ثبت</dt>
                <dd>{date(selected.created_at)}</dd>
                <dt>پرداخت</dt>
                <dd>
                  <Status value={selected.payment} />
                </dd>
                <dt>عملیات</dt>
                <dd>
                  <Status value={selected.status} />
                </dd>
                <dt>بسته‌ها</dt>
                <dd>
                  {selected.packages.join("، ") || "هنوز بسته‌ای ثبت نشده"}
                </dd>
              </dl>
              {selected.delivery_from && <section aria-label="وعده ثبت‌شدهٔ حمل">
                <h3>{selected.shipping_terms?.name ?? selected.shipping_method}</h3>
                <p>{shippingCharge(selected.shipping_terms?.collection_mode,String((selected.shipping_terms?.fee ?? 0)/10))}</p>
                <p>{deliveryPromise(selected.delivery,selected.delivery_from,selected.delivery_to ?? selected.delivery_from)}</p>
              </section>}
              {selected.status === "shipped" && (
                <Button
                  disabled={!can("order.deliver", "support")}
                  variant="contained"
                  onClick={() =>
                    open("تأیید دریافت توسط پشتیبانی", {
                      action: "order.deliver",
                      target: selected.id,
                      version: selected.version,
                    })
                  }
                >
                  ثبت دریافت با دلیل
                </Button>
              )}
            </section>
            <section className="panel detail">
              <h2>اقلام و snapshot خرید</h2>
              <DataTable
                rows={selected.items}
                columns={[
                  { title: "محصول", render: (i) => product(i.product_id) },
                  { title: "تعداد", render: (i) => number(i.quantity) },
                  { title: "قیمت ثبت‌شده", render: (i) => money(i.unit_price) },
                  {
                    title: "وضعیت",
                    render: (i) => <Status value={i.status} />,
                  },
                ]}
              />
              <Alert severity="info" sx={{ mt: 3 }}>
                ویرایش محصول یا تعرفه جاری، داده تاریخی این خرید را تغییر
                نمی‌دهد.
              </Alert>
            </section>
          </div>
        </>
      ) : (
        <section className="panel">
          <div className="toolbar">
            <input
              aria-label="جست‌وجو"
              className="search"
              placeholder="جست‌وجو با شناسه، فروشگاه یا وضعیت…"
              value={search}
              onChange={(e) => setSearch(e.target.value)}
            />
            <span className="muted">محیط آزمون</span>
          </div>
          {section === "orders" && (
            <DataTable
              rows={filter(state.orders ?? [])}
              columns={[
                {
                  title: "سفارش / خرید",
                  render: (o) => (
                    <>
                      <Link to={"/orders/" + o.id}>
                        <b>سفارش {o.purchase_number ?? o.id}</b>
                      </Link>
                      <small>
                        {o.purchase_number
                          ? "گروه ارسال فروشگاه"
                          : o.purchase_id}
                      </small>
                    </>
                  ),
                },
                {
                  title: "فروشگاه",
                  render: (o) => o.store_display_name || seller(o.store_id),
                },
                { title: "مقصد", render: (o) => o.city },
                {
                  title: "مبلغ کالا",
                  render: (o) =>
                    money(
                      o.items.reduce(
                        (s, i) => s + i.unit_price * i.quantity,
                        0,
                      ),
                    ),
                },
                {
                  title: "پرداخت",
                  render: (o) => <Status value={o.payment} />,
                },
                { title: "وضعیت", render: (o) => <Status value={o.status} /> },
                {
                  title: "جزئیات",
                  render: (o) => (
                    <Button component={Link} to={"/orders/" + o.id}>
                      مشاهده
                    </Button>
                  ),
                },
              ]}
            />
          )}
          {section === "support" && (
            <DataTable
              rows={filter(state.cases ?? [])}
              columns={[
                {
                  title: "پرونده",
                  render: (c) => (
                    <>
                      <b>{c.id}</b>
                      <small>{c.kind}</small>
                    </>
                  ),
                },
                {
                  title: "سفارش",
                  render: (c) => (
                    <Link to={"/orders/" + c.order_id}>{c.order_id}</Link>
                  ),
                },
                {
                  title: "دلیل / مهلت",
                  render: (c) => (
                    <>
                      {c.reason}
                      <small>{date(c.deadline)}</small>
                    </>
                  ),
                },
                { title: "وضعیت", render: (c) => <Status value={c.status} /> },
                {
                  title: "رسیدگی",
                  render: (c) => (
                    <div className="action-group">
                      <Button
                        disabled={
                          !can("case.start", "support") || c.status !== "open"
                        }
                        onClick={() =>
                          open("شروع رسیدگی", {
                            action: "case.start",
                            target: c.id,
                            version: c.version,
                          })
                        }
                      >
                        شروع
                      </Button>
                      <Button
                        disabled={
                          !can("case.refund", "support") ||
                          c.status === "closed" ||
                          c.status === "refund-pending"
                        }
                        onClick={() =>
                          open("درخواست بازپرداخت قلم", {
                            action: "case.refund",
                            target: c.id,
                            version: c.version,
                          })
                        }
                      >
                        بازپرداخت
                      </Button>
                      <Button
                        disabled={
                          !can("case.assign", "support") ||
                          c.status === "closed" ||
                          c.status === "assigned"
                        }
                        onClick={() =>
                          open(
                            "واگذاری تأمین جایگزین",
                            {
                              action: "case.assign",
                              target: c.id,
                              version: c.version,
                            },
                            [
                              {
                                name: "value",
                                label: "فروشگاه جایگزین",
                                options: (state.stores ?? [])
                                  .filter(
                                    (s) =>
                                      s.active &&
                                      s.id !==
                                        state.orders?.find(
                                          (o) => o.id === c.order_id,
                                        )?.store_id,
                                  )
                                  .map((s) => ({ value: s.id, label: s.name })),
                              },
                              { name: "reason", label: "دلیل واگذاری" },
                            ],
                          )
                        }
                      >
                        جایگزین
                      </Button>
                      <Button
                        disabled={
                          !can("case.close", "support") || c.status === "closed"
                        }
                        onClick={() =>
                          open("بستن پرونده", {
                            action: "case.close",
                            target: c.id,
                            version: c.version,
                          })
                        }
                      >
                        بستن
                      </Button>
                    </div>
                  ),
                },
              ]}
            />
          )}
          {section === "assignments" && (
            <>
              <div className="calendar-note">
                <b>تأمین مستقل، پرداخت مشتری ثابت</b>
                <span>
                  نسخه قرارداد هنگام تأیید واگذاری ثبت می‌شود. ایجاد واگذاری از
                  صفحه پشتیبانی انجام می‌شود.
                </span>
              </div>
              <DataTable
                rows={filter(state.assignments ?? [])}
                empty="هنوز واگذاری ثبت نشده. از صفحه پشتیبانی، یک قلم را به فروشگاه جایگزین واگذار کنید."
                columns={[
                  { title: "واگذاری", render: (a) => a.id.slice(0, 12) },
                  { title: "پرونده", render: (a) => a.case_id },
                  { title: "تأمین‌کننده", render: (a) => seller(a.store_id) },
                  { title: "قلم", render: (a) => a.item_id },
                  {
                    title: "نسخه قرارداد",
                    render: (a) => a.contract_version || "پس از تأیید فروشنده",
                  },
                  {
                    title: "وضعیت",
                    render: (a) => <Status value={a.status} />,
                  },
                ]}
              />
            </>
          )}
          {section === "notifications" && (
            <DataTable
              rows={filter(state.notifications ?? [])}
              columns={[
                { title: "رخداد", render: (n) => <b>{n.event}</b> },
                { title: "گیرنده", render: (n) => n.recipient },
                {
                  title: "نتیجه آزمون",
                  render: (n) => <Status value={n.status} />,
                },
                {
                  title: "تلاش مجدد",
                  render: (n) => `${number(n.retries)} / ۳`,
                },
                {
                  title: "عملیات",
                  render: (n) => (
                    <Button
                      disabled={
                        !can("notification.retry") ||
                        n.retries >= 3 ||
                        n.status === "success"
                      }
                      onClick={() =>
                        open("تلاش مجدد اعلان آزمایشی", {
                          action: "notification.retry",
                          target: n.id,
                        })
                      }
                    >
                      تلاش مجدد
                    </Button>
                  ),
                },
              ]}
            />
          )}
          {section === "reports" && (
            <DataTable
              rows={filter([...(state.audit ?? [])].reverse())}
              empty="هنوز عملیاتی ثبت نشده است."
              columns={[
                {
                  title: "نوع عملیات",
                  render: (a) => <span className="code">{a.action}</span>,
                },
                { title: "عامل", render: (a) => a.actor },
                { title: "هدف", render: (a) => a.target || "رکورد جدید" },
                { title: "دلیل", render: (a) => a.reason || "—" },
                { title: "زمان", render: (a) => date(a.at) },
              ]}
            />
          )}
        </section>
      )}
      {dialog && (
        <ActionDialog
          open
          title={dialog.title}
          fields={dialog.fields}
          onClose={() => setDialog(null)}
          onSubmit={(values) => execute({ ...dialog.command, ...values })}
        />
      )}
    </>
  );
}
