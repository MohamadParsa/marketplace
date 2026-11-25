import { Link } from "react-router-dom";
import {
  ArrowBack,
  Storefront,
  Inventory2,
  ReceiptLong,
  SupportAgent,
} from "@mui/icons-material";
import { useWorkspace } from "../../shared/workspace";
import { DataTable, PageHeader, Status } from "../../shared/components";
import { money, number, date } from "../../shared/format";
export function DashboardPage() {
  const { state } = useWorkspace();
  const orders = state.orders ?? [];
  const stores = state.stores ?? [];
  const cases = state.cases ?? [];
  const cards = [
    {
      label: "فروشگاه‌های فعال",
      value: stores.filter((s) => s.active).length,
      icon: <Storefront />,
      note: "از " + number(stores.length) + " فروشگاه",
      link: "/stores",
    },
    {
      label: "محصول در انتظار بررسی",
      value: (state.products ?? []).filter((p) => p.status === "pending")
        .length,
      icon: <Inventory2 />,
      note: "آماده بررسی محتوا",
      link: "/products",
    },
    {
      label: "سفارش‌های در جریان",
      value: orders.filter((o) => o.status !== "delivered").length,
      icon: <ReceiptLong />,
      note: "نیازمند پیگیری عملیات",
      link: "/orders",
    },
    {
      label: "پرونده‌های باز",
      value: cases.filter((c) => c.status !== "closed").length,
      icon: <SupportAgent />,
      note: "در صف رسیدگی پشتیبانی",
      link: "/support",
    },
  ];
  return (
    <>
      <PageHeader
        title="نمای کلی بازارگاه"
        description="وضعیت امروز، کارهای در انتظار و آخرین فعالیت‌های مجموعه"
        action={<span className="today">{date(new Date().toISOString())}</span>}
      />
      <div className="metric-grid">
        {cards.map((card) => (
          <Link className="metric" key={card.label} to={card.link}>
            <div className="metric-top">
              <span>{card.label}</span>
              <i>{card.icon}</i>
            </div>
            <strong>{number(card.value)}</strong>
            <div className="metric-note">
              {card.note}
              <ArrowBack fontSize="small" />
            </div>
          </Link>
        ))}
      </div>
      <div className="dashboard-grid">
        <section className="panel">
          <div className="section-heading">
            <div>
              <h2>آخرین سفارش‌ها</h2>
              <p>خریدهای چندفروشگاهی، با وضعیت مستقل هر سفارش</p>
            </div>
            <Link to="/orders">مشاهده همه ←</Link>
          </div>
          <DataTable
            rows={orders}
            columns={[
              {
                title: "سفارش",
                render: (o) => <Link to={"/orders/" + o.id}>{o.id}</Link>,
              },
              {
                title: "فروشگاه",
                render: (o) => stores.find((s) => s.id === o.store_id)?.name,
              },
              {
                title: "مبلغ",
                render: (o) =>
                  money(
                    o.items.reduce(
                      (sum, i) => sum + i.quantity * i.unit_price,
                      0,
                    ),
                  ),
              },
              { title: "وضعیت", render: (o) => <Status value={o.status} /> },
            ]}
          />
        </section>
        <section className="panel attention">
          <span className="overline">نیازمند توجه شما</span>
          <h2>صف کارهای امروز</h2>
          <Link to="/support">
            <span className="queue-number">{number(cases.length)}</span>
            <div>
              پرونده پشتیبانی<small>عدم تأمین و کسری بسته</small>
            </div>
            <ArrowBack fontSize="small" />
          </Link>
          <Link to="/products">
            <span className="queue-number">
              {number(
                (state.products ?? []).filter((p) => p.status === "pending")
                  .length,
              )}
            </span>
            <div>
              بررسی محتوای محصول<small>در انتظار تصمیم بک‌آفیس</small>
            </div>
            <ArrowBack fontSize="small" />
          </Link>
          <Link to="/notifications">
            <span className="queue-number">
              {number(
                (state.notifications ?? []).filter(
                  (n) => n.status !== "success",
                ).length,
              )}
            </span>
            <div>
              اعلان ارسال‌نشده<small>بازبینی و تلاش مجدد</small>
            </div>
            <ArrowBack fontSize="small" />
          </Link>
          <div className="test-note">
            <b>محیط بررسی با داده آزمون</b>
            <p>
              فرایندها را با داده آزمون بررسی کنید. وضعیت این محیط به معنی ثبت
              سفارش یا پرداخت واقعی نیست.
            </p>
          </div>
        </section>
      </div>
      <section className="panel">
        <div className="section-heading">
          <div>
            <h2>فعالیت‌های اخیر</h2>
            <p>تغییرات ثبت‌شده در همین محیط</p>
          </div>
          <Link to="/reports">گزارش فعالیت ←</Link>
        </div>
        <DataTable
          rows={[...(state.audit ?? [])].reverse().slice(0, 5)}
          empty="هنوز فعالیتی ثبت نشده؛ اولین عملیات شما اینجا نمایش داده می‌شود."
          columns={[
            { title: "نوع عملیات", render: (a) => a.action },
            { title: "عامل", render: (a) => a.actor },
            { title: "دلیل", render: (a) => a.reason || "—" },
            { title: "زمان", render: (a) => date(a.at) },
          ]}
        />
      </section>
    </>
  );
}
