export type Role =
  | "super-admin"
  | "operations"
  | "support"
  | "finance"
  | "viewer"
  | "";
export type RoleScope = "global" | "store";
export interface CustomRole {
  id: string;
  name: string;
  scope: RoleScope;
  permissions: string[];
  version: number;
}
interface StoreMembership {
  user_id: string;
  store_id: string;
}
interface RoleGrant {
  user_id: string;
  role_id: string;
  store_id?: string;
}
export interface User {
  id: string;
  name: string;
  role: Role;
  active: boolean;
}
export interface Store {
  content_policy?: "review" | "direct";
  content_policy_version?: number;
  id: string;
  name: string;
  owner: string;
  city: string;
  model: string;
  active: boolean;
  lead_days: number;
}
export interface CategorySEO {
  meta_title: string;
  meta_keywords: string;
  meta_description: string;
}
export interface Category {
  seo?: CategorySEO;
  english_name?: string;
  description?: string;
  published: boolean;
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
export interface ProductContent {
  short_description: string;
  full_description: string;
  advantages: string[];
  disadvantages: string[];
  reviews_enabled: boolean;
  qa_enabled: boolean;
  shipping: {
    weight?: number;
    length?: number;
    width?: number;
    height?: number;
    weight_unit: string;
    dimension_unit: string;
  };
  seo: CategorySEO;
  noindex: boolean;
  nofollow: boolean;
}
export interface ProductContentRevision {
  id: string;
  state: string;
  base_fingerprint: string;
  submitted_by: string;
  submitted_at: string;
  reviewed_by?: string;
  reviewed_at?: string;
  reason?: string;
  name: string;
  english_name: string;
  description: string;
  brand_id: string;
  content?: ProductContent;
  attributes: { attribute_id: string; value: string }[];
}
export interface Product {
  content_revision?: ProductContentRevision;
  content?: ProductContent;
  gallery_ids?: string[];
  english_name?: string;
  description?: string;
  brand_id?: string;
  image_id?: string;
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
  regular_price?: number;
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
  image_id?: string;
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
    edit_content: boolean;
    review_content: boolean;
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
export interface HomeSettings {
  version: number;
  rails: { key: string; category_id: string; limit: number }[];
}
export interface CatalogInput {
  model?: "shared" | "independent";
  owner_store_id?: string;
  content?: ProductContent;
  seo?: CategorySEO;
  attributes?: { attribute_id: string; value: string }[];
  name: string;
  english_name: string;
  description: string;
  category_id: string;
  brand_id: string;
  parent_id: string;
  published: boolean;
}
export interface Workspace {
  variants?: Variant[];
  home_settings: HomeSettings;
  access?: {
    roles: CustomRole[] | null;
    grants: RoleGrant[] | null;
    memberships?: StoreMembership[];
    registry: Record<RoleScope, string[]>;
  };
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
  user: {
    ID: string;
    Role: Role;
    Surface: "backoffice" | "seller";
    Permissions?: string[];
  };
}
export interface Command {
  product_media?: {
    gallery_ids: string[];
    variant_images: Record<string, string>;
  };
  catalog?: CatalogInput;
  home?: HomeSettings;
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
    regular_price?: number;
    stock?: number;
  };
  reauthentication?: { request_id: string; code: string };
  scope?: RoleScope;
  permissions?: string[];
  store_id?: string;
}

export interface BannerConfig {
  placement?: "hero" | "promo";
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
