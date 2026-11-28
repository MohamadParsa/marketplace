export interface CartItem {
 inactive?: boolean;
 restricted?: boolean;
 review_required?: boolean;
 previous_variant_label?: string;
 variant_version?: number;
  product_url: string;
  product_id: string;
  variant_id: string;
  offer_id: string;
  store_id: string;
  store_name: string;
  name: string;
  variant_label: string;
  image_url: string;
  quantity: number;
  unit_price_toman: string;
  previous_price_toman: string;
  total_toman: string;
  available: boolean;
  price_changed: boolean;
  lead_days: number;
  lead_day_basis?: string;
}
export interface Cart {
  id: string;
  version: number;
  items: CartItem[];
  subtotal_toman: string;
  item_count: number;
  can_checkout: boolean;
  currency: string;
}
export interface Address {
  detail_version?: number;
  title?: string;
  full_address?: string;
  house_number?: string;
  unit?: string;
  recipient_self?: boolean;
  recipient_national_id?: string;
  latitude?: number;
  longitude?: number;
  id: string;
  version: number;
  recipient: string;
  phone: string;
  province: string;
  city: string;
  street: string;
  postal_code: string;
}
export interface Method {
  total_toman?: string;
  sales_method?: { id: string; name: string; markup_bps: number };
  id: string;
  name: string;
  kind: string;
  active: boolean;
  min_toman: string;
  max_toman: string;
  all_categories: boolean;
  allowed_categories: string[];
  excluded_categories: string[];
  instructions: string;
}
export interface Shipping {
  dispatch_identifier?: "" | "none" | "tracking" | "delivery";
  schedule?: ShippingSchedule | null;
  version?: number;
  collection_mode?: "prepaid" | "collect";
  delivery_mode?: "range" | "slots";
  daily_capacity?: number | null;
  category_mode?: "all" | "selected";
  allowed_categories?: string[];
  denied_categories?: string[];
  city_ids?: string[];
  day_basis?: "working" | "calendar";
  id: string;
  name: string;
  active: boolean;
  all_cities: boolean;
  cities: string[];
  store_ids: string[];
  fee_toman: string;
  min_days: number;
  max_days: number;
}
export interface ShippingSchedule {
  windows: {id:string;start:string;end:string;active:boolean}[];
  days: {weekday:number;active:boolean;slots:{window_id:string;capacity:number|null}[]}[];
}
export interface DeliveryPlan {
  capacity_day:string;
  dispatch_date:string;
  delivery_from:string;
  delivery_to:string;
  method_version:number;
  store_revision:number;
  slot?:{window_id:string;start_at:string;end_at:string;capacity:number};
}
export interface DeliverySelection { date:string; window_id:string }
export interface ShippingAvailability {
  store_id:string;
  store_name:string;
  reason?:string;
  methods:{method:Shipping;range?:DeliveryPlan;range_remaining?:number;days:{date:string;open:boolean;daily_remaining:number;slots:{plan:DeliveryPlan;remaining:number}[]}[]}[];
}
export interface Destination {
  id: string;
  card_number: string;
  holder: string;
  bank: string;
}
export interface SocialLink { id:string; title:string; kind:string; url:string; position:number; active:boolean }
export interface Business {
 published_pages?: Record<string,boolean>;
 social_links?:SocialLink[];
  name: string;
  support_phone: string;
  support_email: string;
  about: string;
  contact: string;
  shipping_policy: string;
  return_policy: string;
  privacy: string;
  terms: string;
}
export interface Province {
  id: string;
  name: string;
  active: boolean;
  cities: { id: string; name: string; active: boolean }[];
}
export interface Config {
  geography: Province[];
  methods: Method[];
  destination: Destination;
  business: Business;
  currency: string;
  reservation_minutes: number;
  server_date: string;
}
export interface Settings {
  shipping_timing: { ready_time: string; cutoff_time: string; min_slot_lead_minutes: number; allow_same_day_slots: boolean };
  geography: Province[];
  version: number;
  methods: Method[];
  shipping: Shipping[];
  closed_weekdays: number[];
  holidays: string[];
  destination: Destination;
  business: Business;
  currency: string;
  reservation_minutes: number;
}
export interface QuoteLine {
  reference_unit_price_toman?: string;
  direct_discount_toman?: string;
  offer_id: string;
  quantity: number;
  product_id: string;
  variant_id: string;
  store_id: string;
  name: string;
  variant_label: string;
  image_url: string;
  unit_price_toman: string;
  total_toman: string;
}
export interface Shipment {
  delivery?: DeliveryPlan;
  collection_mode?: "prepaid" | "collect";
  transport_day_basis?: "working" | "calendar";
  ready_at?: string;
  store_id: string;
  store_name: string;
  method_id: string;
  method_name: string;
  fee_toman: string;
  lead_days: number;
  lead_day_basis?: string;
  min_days: number;
  max_days: number;
  delivery_from: string;
  delivery_to: string;
}
export interface Quote {
  cash_subtotal_toman?: string;
  markup_toman?: string;
  sales_method?: { id: string; name: string; markup_bps: number };
  direct_discount_toman?: string;
  contract_version?: number;
  terms_digest?: string;
  coupon_code?: string;
  discount_toman?: string;
  wallet_selected_toman?: string;
  funding_toman?: string;
  destination: Destination;
  id: string;
  status: string;
  address: Address;
  lines: QuoteLine[];
  shipments: Shipment[];
  payment_method: Method;
  subtotal_toman: string;
  shipping_toman: string;
  total_toman: string;
  expires_at: string;
  reserved_until: string;
  currency: string;
}
export interface OrderEvent {
  customer_description?: string;
  id: string; order_id?:string; request_id?:string; from?:string; to?:string;
  action: string;
  flow_status: string;
  order_status: string;
  payment_status: string;
  actor_kind: string;
  at: string;
}
export interface Deposit {
  id: string;
  purchase_id?: string;
  method_id: string;
  method_name: string;
  amount_toman: string;
  date: string;
  reference: string;
  status: string;
  version: number;
  receipt_id?: string;
  destination: Destination;
  instructions: string;
  reason?: string;
  created_at: string;
}
export interface Purchase {
  wallet_consumed_toman?: string;
  payment_stage?: string;
  payment_attempts?: PaymentAttempt[];
  fulfillment_status?: string;
  discount_funding?: {
    platform_toman: string;
    seller_toman: string;
    rule_version: number;
  };
  credit_sources?: {
    lot_id: string;
    title?: string;
    origin_lot_id?: string;
    withdrawable: boolean;
    source: string;
    amount_toman: string;
    kind: string;
    status: string;
  }[];
  number: string;
  id: string;
  status: string;
  version: number;
  order_status: string;
  payment_status: string;
  events: OrderEvent[];
  cancellation_requested: boolean;
  cancellation_reason?: string;
  quote: Quote;
  order_ids: string[];
  created_at: string;
  confirmed_at?: string;
  wallet_balance_toman: string;
  wallet_available_toman: string;
  shortfall_toman: string;
  deposits: Deposit[];
  reason?: string;
}
export interface Refund {
  id: string;
  amount_toman: string;
  instructions: string;
  status: string;
  version: number;
  reference?: string;
  date?: string;
  reason?: string;
  created_at: string;
}
export interface Entry {
  id: string;
  source: string;
  kind: string;
  amount_toman: string;
  at: string;
}
export interface Wallet {
  withdrawable_toman: string;
  purchase_only_toman: string;
  credits: {
    id: string;
    title: string;
    description?: string;
    source: string;
    origin_lot_id?: string;
    amount_toman: string;
    spent_toman: string;
    held_toman: string;
    free_toman: string;
    withdrawable: boolean;
    at: string;
  }[];
  balance_toman: string;
  held_toman: string;
  available_toman: string;
  entries: Entry[];
  deposits: Deposit[];
  refunds: Refund[];
  funding_methods: Method[];
  destination: Destination;
  currency: string;
}
export interface Prepare {
  cart: Cart;
  addresses: Address[];
  shipping: Shipping[];
  methods: Method[];
  wallet_balance_toman: string;
  wallet_available_toman: string;
}
export interface Preview {
  quote: Quote;
  methods: (Method & { eligible: boolean; reason?: string })[];
}
export interface QuoteInput {
  shipping_selections?: Record<string,DeliverySelection>;
  contract_version?: number;
  coupon_code?: string;
  wallet_toman?: string;
  cart_version: number;
  address_id: string;
  shipping: Record<string, string>;
  payment_method_id: string;
  purchase_id?: string;
}
export interface CustomerInfo {
  national_id?: string;
  name: string;
  email?: string;
  password_available: boolean;
  phone?: string;
}
export interface AdminCustomer {
  user_id: string;
  customer_name: string;
  customer_phone: string;
}
export interface AdminView {
  purchases: (Purchase & AdminCustomer)[];
  deposits: (Deposit & AdminCustomer)[];
  refunds: (Refund & AdminCustomer)[];
  permissions: Record<string, boolean>;
  audit: { id: string; action: string; target: string; created_at: string }[];
}
export const statuses: Record<string, string> = {
  "fulfillment.start": "شروع پردازش",
  "fulfillment.supply": "تأیید تأمین",
  "fulfillment.pack": "شروع آماده‌سازی",
  "fulfillment.ship": "ثبت ارسال",
  "fulfillment.deliver": "تأیید تحویل",
  "fulfillment.configure-dispatch": "تعیین شناسه ارسال مرسوله",
  "fulfillment.correct": "اصلاح مشخصات با دلیل",
  "fulfillment.items-cancelled": "لغو اقلام تأیید شد",
  "fulfillment.stock-returned": "کالای سالم به موجودی برگشت",
  "request.created": "ثبت درخواست کنسلی یا مرجوعی",
  "request.approve": "تأیید درخواست",
  "request.reject": "رد درخواست",
  "request.receive": "دریافت فیزیکی کالای برگشتی",
  "request.inspect": "ثبت بازرسی کالای برگشتی",
  "refund.wallet": "بازگشت وجه به کیف پول",
  "refund.bank": "رزرو وجه برای انتقال دستی",
  "refund.bank-completed": "ثبت نتیجه انتقال بانکی",
  "refund.bank-rejected": "پایان مسیر بانکی؛ وجه در کیف پول",
  requested: "درخواست پرداخت",
  "awaiting-result": "در انتظار نتیجه",
  failed: "ناموفق",
  authorized: "مجاز؛ غیرقطعی",
  placed: "ثبت سفارش",
  processing: "در حال پردازش",
  supplied: "تأمین شده",
  packing: "در حال آماده‌سازی",
  shipped: "ارسال شده",
  delivered: "تحویل شده",
  mixed: "مراحل متفاوت مرسوله‌ها",
  "awaiting-confirmation": "در انتظار تأیید تیم",
  "requires-reconfirmation": "نیازمند تأیید دوباره شما",
  confirmed: "تأیید شده",
  cancelled: "لغو شده",
  "awaiting-payment": "در انتظار پرداخت",
  "awaiting-supply": "در انتظار تأمین",
  unpaid: "پرداخت نشده",
  paid: "پرداخت شده",
  "returned-to-wallet": "برگشت به کیف پول",
  pending: "در انتظار بررسی",
  approved: "تأیید و شارژ شد",
  rejected: "رد شده",
  completed: "تکمیل شده",
  "admin-credit": "شارژ مدیریتی",
  "deposit-credit": "شارژ کیف پول",
  "order-payment": "پرداخت سفارش",
  "order-cancel-credit": "بازگشت وجه سفارش",
  "manual-refund": "بازپرداخت دستی",
  "order.accepted": "ثبت سفارش و رزرو",
  "order.paid": "پرداخت و تأیید نهایی",
  "order.cancelled": "لغو سفارش",
  "order.cancellation-requested": "درخواست لغو",
  "deposit.approved": "تأیید واریز و شارژ کیف پول",
  "deposit.rejected": "رد واریز",
  "reservation.expired": "پایان رزرو",
  "order.reconfirmation-required": "نیازمند تأیید مجدد",
};
export const label = (value: string) => statuses[value] ?? value;
// Formatting only; all calculations and validation of financial amounts are server-side.
export function money(value: string) {
  if (!value) return "—";
  const negative = value.startsWith("-");
  const [whole, fraction] = value.replace(/^-/, "").split(".");
  return `${negative ? "−" : ""}${whole.replace(/\B(?=(\d{3})+(?!\d))/g, "٬")}${fraction ? "٫" + fraction : ""} تومان`;
}

export interface PaymentAttempt {
  id: string;
  purpose: "order-payment" | "wallet-funding";
  method_name: string;
  amount_toman: string;
  status: string;
  failure_code?: string;
  created_at: string;
  updated_at: string;
}

export interface OrderOperations {
  id: string;
  purchase_id: string;
  purchase_number: string;
  store_name: string;
  version: number;
  order_status: string;
  status: string;
  payment_status: string;
  shipping_name: string;
  dispatch_identifier: string;
  identifiers: string[] | null;
  items: { id: string; name: string; quantity: number; cancelled_quantity?: number; returned_quantity?: number; status: string }[];
  address?: Address;
  allowed_actions: string[];
  events: { id?: string; action: string; from: string; to: string; actor_kind: string; at: string; reason?: string; customer_description?:string }[];
}

export interface OrderOperationInput {
  version: number;
  action: string;
  reason: string;
  dispatch_identifier?: string;
  identifiers?: string[];
  all_items: boolean;
}

export interface RequestItem {
  order_id: string; item_id: string; name: string; store_name: string;
  quantity: number; available_cancel: number; available_return: number;
}
export interface OrderRequest {
	refund?: OrderRefundSummary; refundable_toman?: string; can_refund: boolean; can_bank: boolean;
  id: string; number: string; kind: string; status: string; version: number;
  reason?: string; review_reason?: string; created_at: string; reviewed_at?: string;
  goods_received: boolean; inspected: boolean; healthy: boolean; stock_applied: boolean;
  items: RequestItem[]; allowed_actions: string[];
}
export interface OrderRefundSummary {
 method: string; amount_toman: string; shipping_toman: string; status: string; at: string;
 settlement_status?: string;
 withdrawal?: {id:string;version:number;status:string;amount_toman:string;instructions:string;reference?:string;date?:string;reason?:string};
}
export interface OrderRequests {
  purchase_id: string; purchase_number: string; version: number;
  can_create: boolean; can_review: boolean; legacy_requested: boolean; legacy_reason?: string;
  items: RequestItem[]; requests: OrderRequest[];
}

export interface WalletCreditInput {
  user_id: string;
  amount_toman: string;
  title: string;
  description: string;
  withdrawable: boolean;
}
export interface WalletCreditResult extends WalletCreditInput {
  id: string;
  customer_name: string;
  balance_toman: string;
  available_toman: string;
  withdrawable_toman: string;
  at: string;
}
export interface CustomerCreditWallet {
  user_id: string;
  customer_name: string;
  balance_toman: string;
  available_toman: string;
  withdrawable_toman: string;
  credits: Wallet["credits"];
  previous_operation?: WalletCreditResult;
}
