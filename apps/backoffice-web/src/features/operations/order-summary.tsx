import { useState, type ReactNode } from "react";
import ReceiptLongOutlined from "@mui/icons-material/ReceiptLongOutlined";
import PersonOutline from "@mui/icons-material/PersonOutline";
import LocalShippingOutlined from "@mui/icons-material/LocalShippingOutlined";
import ImageOutlined from "@mui/icons-material/ImageOutlined";
import { money, type Purchase } from "../../../../../packages/commerce-client/types";
import { dateTime, number } from "../../shared/format";
import "./order-summary.css";

type SummaryPurchase = Purchase & { customer_name?: string; customer_phone?: string };
const digits = (value: string) => value.replace(/\d/g, digit => "۰۱۲۳۴۵۶۷۸۹"[Number(digit)]);
const text = (value?: string) => value?.trim() ? digits(value) : "ثبت نشده";
const amount = (value?: string) => value ? digits(money(value)) : "ثبت نشده";

function Fact({ label, children, wide = false }: { label: string; children: ReactNode; wide?: boolean }) {
  return <div className={wide ? "order-summary-fact is-wide" : "order-summary-fact"}><dt>{label}</dt><dd>{children}</dd></div>;
}

function ProductImage({ src, name }: { src: string; name: string }) {
  const [failed, setFailed] = useState(false);
  return <div className="order-summary-image">{src && !failed
    ? <img src={src} alt={name} loading="lazy" onError={() => setFailed(true)} />
    : <ImageOutlined aria-label="تصویر موجود نیست" />}</div>;
}

export function OrderSummary({ purchase: p }: { purchase: SummaryPurchase }) {
  const lines = p.quote.lines ?? [];
  const shipments = p.quote.shipments ?? [];
  const address = p.quote.address;
  const registeredAt = p.created_at && Number.isFinite(Date.parse(p.created_at)) ? dateTime(p.created_at) : "ثبت نشده";
  const location = [address?.province, address?.city].filter(Boolean).join("، ");
  const street = address?.full_address || [address?.street, address?.house_number && `پلاک ${address.house_number}`, address?.unit && `واحد ${address.unit}`].filter(Boolean).join("، ");
  const base = import.meta.env.VITE_API_BASE_URL ?? "";

  return <div className="order-summary" aria-label="خلاصه سفارش و اقلام">
    <div className="order-summary-overview">
      <section className="order-summary-card" aria-labelledby="order-summary-information">
        <h2 id="order-summary-information"><ReceiptLongOutlined aria-hidden="true" />اطلاعات سفارش</h2>
        <dl className="order-summary-facts">
          <Fact label="شماره سفارش"><bdi>{text(p.number)}</bdi></Fact>
          <Fact label="تاریخ ثبت">{registeredAt}</Fact>
          <Fact label="مشتری">{text(p.customer_name)}</Fact>
          <Fact label="موبایل مشتری"><bdi dir="ltr">{text(p.customer_phone)}</bdi></Fact>
          <Fact label="تعداد کالا">{number(lines.reduce((count, line) => count + line.quantity, 0))} عدد</Fact>
          <Fact label="تعداد مرسوله">{number(shipments.length)} مرسوله</Fact>
        </dl>
      </section>
      <section className="order-summary-card" aria-labelledby="order-summary-recipient">
        <h2 id="order-summary-recipient"><PersonOutline aria-hidden="true" />اطلاعات گیرنده</h2>
        <dl className="order-summary-facts">
          <Fact label="نام گیرنده">{text(address?.recipient)}</Fact>
          <Fact label="تلفن گیرنده"><bdi dir="ltr">{text(address?.phone)}</bdi></Fact>
          <Fact label="استان و شهر">{text(location)}</Fact>
          <Fact label="کد پستی"><bdi dir="ltr">{text(address?.postal_code)}</bdi></Fact>
          <Fact label="نشانی تحویل" wide>{text(street)}</Fact>
        </dl>
      </section>
    </div>

    <section className="order-summary-card" aria-labelledby="order-summary-items">
      <div className="order-summary-section-heading"><h2 id="order-summary-items">اقلام سفارش</h2><span>{number(lines.length)} ردیف</span></div>
      {lines.length ? <div className="order-summary-lines">{lines.map(line => {
        const store = shipments.find(shipment => shipment.store_id === line.store_id);
        const image = line.image_url ? (/^https?:\/\//.test(line.image_url) ? line.image_url : base + line.image_url) : "";
        return <article key={line.offer_id} className="order-summary-line">
          <ProductImage key={image} src={image} name={line.name} />
          <div className="order-summary-product"><h3>{text(line.name)}</h3>{line.variant_label && <p>{text(line.variant_label)}</p>}{store?.store_name && <span>فروشگاه: {text(store.store_name)}</span>}</div>
          <dl className="order-summary-line-totals">
            <Fact label="تعداد">{number(line.quantity)} عدد</Fact>
            <Fact label="قیمت واحد">{amount(line.unit_price_toman)}</Fact>
            <Fact label="جمع قلم">{amount(line.total_toman)}</Fact>
          </dl>
        </article>;
      })}</div> : <p className="order-summary-empty">قلمی برای این سفارش ثبت نشده است.</p>}
      <div className="order-summary-total"><span>مبلغ کل سفارش</span><strong>{amount(p.quote.total_toman)}</strong></div>
    </section>

    <section className="order-summary-card" aria-labelledby="order-summary-shipments">
      <div className="order-summary-section-heading"><h2 id="order-summary-shipments"><LocalShippingOutlined aria-hidden="true" />خلاصه مرسوله‌ها</h2><span>{number(shipments.length)} مرسوله</span></div>
      {shipments.length ? <div className="order-summary-shipments">{shipments.map(shipment => <article key={shipment.store_id}>
        <h3>{text(shipment.store_name)}</h3>
        <dl className="order-summary-facts">
          <Fact label="روش ارسال">{text(shipment.method_name)}</Fact>
          <Fact label="هزینه ارسال">{shipment.collection_mode === "collect" ? "پس‌کرایه" : amount(shipment.fee_toman)}</Fact>
        </dl>
      </article>)}</div> : <p className="order-summary-empty">اطلاعات مرسوله برای این سفارش ثبت نشده است.</p>}
    </section>
  </div>;
}
