export interface CartItem {
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
export interface Destination {
  id: string;
  card_number: string;
  holder: string;
  bank: string;
}
export interface Business {
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
  store_id: string;
  store_name: string;
  method_id: string;
  method_name: string;
  fee_toman: string;
  lead_days: number;
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
  "order-cancel-credit": "برگشت لغو سفارش",
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
