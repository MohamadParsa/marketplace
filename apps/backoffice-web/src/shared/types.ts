export type Role =
  | "super-admin"
  | "operations"
  | "support"
  | "finance"
  | "viewer";
export interface User {
  id: string;
  name: string;
  role: Role;
  active: boolean;
}
export interface Store {
  id: string;
  name: string;
  owner: string;
  city: string;
  model: string;
  active: boolean;
  lead_days: number;
}
export interface Category {
  id: string;
  name: string;
  parent_id?: string;
  bindings: AttributeBinding[];
  archived: boolean;
  icon_id?: string;
  version: number;
}
export interface Brand {
  id: string;
  name: string;
  active: boolean;
  icon_id?: string;
  version: number;
}
export interface IconPolicy {
  max_bytes: number;
  max_dimension: number;
  content_types: string[];
}
export interface Product {
  id: string;
  name: string;
  category_id: string;
  model: string;
  status: string;
  version: number;
  attributes?: { attribute_id: string; value: string }[];
  owner_store_id?: string;
  variant_version: number;
}
export interface AttributeDefinition {
  id: string;
  code: string;
  name: string;
  kind: "enum" | "text" | "number";
  unit?: string;
  options: string[];
  version: number;
}
export interface AttributeBinding {
  attribute_id: string;
  axis: boolean;
  card: boolean;
  filter: boolean;
  position: number;
}
export interface Offer {
  id: string;
  product_id: string;
  store_id: string;
  price: number;
  stock: number;
  variant_id: string;
  reserved: number;
  version: number;
}
export interface Variant {
  id: string;
  product_id: string;
  sku: string;
  values: { attribute_id: string; value: string }[];
  version: number;
}
export interface OfferView extends Offer {
  available: number;
  lead_days: number;
  store_active: boolean;
}
export interface ProductWorkspace {
  product: Product;
  axes: AttributeDefinition[];
  variants: Variant[];
  offers: OfferView[];
  stores: Store[];
  permissions: {
    create_variant: boolean;
    create_offer_stores: string[];
    price_stores: string[];
    stock_stores: string[];
  };
}
export interface VariantResolution {
  variant: Variant;
  offers: OfferView[];
}
export interface OrderItem {
  id: string;
  product_id: string;
  quantity: number;
  unit_price: number;
  status: string;
}
export interface Order {
  id: string;
  purchase_id: string;
  store_id: string;
  city: string;
  status: string;
  payment: string;
  created_at: string;
  items: OrderItem[];
  packages: string[];
  version: number;
}
export interface Case {
  id: string;
  order_id: string;
  item_id: string;
  kind: string;
  status: string;
  assignee: string;
  reason: string;
  deadline: string;
  version: number;
}
export interface Assignment {
  id: string;
  case_id: string;
  item_id: string;
  store_id: string;
  status: string;
  contract_version: string;
}
export interface Transfer {
  reserved: boolean;
  id: string;
  kind: string;
  order_id: string;
  item_id: string;
  store_id: string;
  amount: number;
  status: string;
  source: string;
  reference: string;
  version: number;
}
export interface Workspace {
  users?: User[];
  stores?: Store[];
  categories?: Category[];
  attributes?: AttributeDefinition[];
  display_settings: { card_attribute_limit: number; version: number };
  card_previews?: {
    product_id: string;
    features: {
      attribute_id: string;
      name: string;
      value: string;
      unit?: string;
    }[];
  }[];
  brands?: Brand[];
  icon_policy?: IconPolicy;
  products?: Product[];
  offers?: Offer[];
  orders?: Order[];
  cases?: Case[];
  assignments?: Assignment[];
  transfers?: Transfer[];
  rules?: {
    id: string;
    name: string;
    level: string;
    kind: string;
    value: number;
    active: boolean;
  }[];
  shipping?: {
    id: string;
    method: string;
    city: string;
    tariff: number;
    days: number;
  }[];
  notifications?: {
    id: string;
    event: string;
    recipient: string;
    status: string;
    retries: number;
  }[];
  audit?: {
    id: string;
    actor: string;
    action: string;
    target: string;
    reason: string;
    at: string;
  }[];
  ledger?: {
    id: string;
    transfer_id: string;
    amount: number;
    kind: string;
    source: string;
  }[];
}
export interface Session {
  token: string;
  expires_at: string;
  user: { ID: string; Role: Role; Surface: "backoffice" | "seller" };
}
export interface Command {
  banner?: BannerConfig;
  action: string;
  target?: string;
  name?: string;
  phone?: string;
  value?: string;
  reason?: string;
  amount?: number;
  version?: number;
  idempotency_key?: string;
  attribute?: AttributeDefinition;
  bindings?: AttributeBinding[];
  variant?: {
    product_id: string;
    sku: string;
    values: { attribute_id: string; value: string }[];
  };
  offer?: {
    product_id?: string;
    variant_id?: string;
    store_id?: string;
    price?: number;
    stock?: number;
  };
}

export interface BannerConfig {
  desktop_id: string;
  mobile_id: string;
  alt: string;
  link: string;
  enabled: boolean;
}
export interface Banner {
  id: string;
  name: string;
  version: number;
  draft: BannerConfig;
  published?: { config: BannerConfig; revision: number };
}
export interface BannerList {
  banners: Banner[];
  can_edit: boolean;
  policy: {
    max_bytes: number;
    max_dimension: number;
    max_pixels: number;
    content_types: string[];
  };
}
