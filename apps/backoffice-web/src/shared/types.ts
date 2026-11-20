export type Role =
  "super-admin" | "operations" | "support" | "finance" | "viewer" | "";
export type RoleScope = "global" | "store";
export interface CustomRole {
  id: string;
  name: string;
  scope: RoleScope;
  permissions: string[];
  version: number;
}
interface StoreMembership { user_id: string; store_id: string }
interface RoleGrant { user_id: string; role_id: string; store_id?: string }
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
}
export interface Category {
  id: string;
  name: string;
  attributes: string[];
  archived: boolean;
}
export interface Product {
  id: string;
  name: string;
  category_id: string;
  model: string;
  status: string;
  version: number;
}
export interface Offer {
  id: string;
  product_id: string;
  store_id: string;
  price: number;
  stock: number;
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
  access?: {
    roles: CustomRole[] | null;
    grants: RoleGrant[] | null;
    memberships?: StoreMembership[];
    registry: Record<RoleScope, string[]>;
  };
  users?: User[];
  stores?: Store[];
  categories?: Category[];
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
  user: { ID: string; Role: Role; Permissions?: string[] };
}
export interface Command {
  action: string;
  target?: string;
  name?: string;
  phone?: string;
  value?: string;
  reason?: string;
  amount?: number;
  version?: number;
  idempotency_key?: string;
  reauthentication?: { request_id: string; code: string };
  scope?: RoleScope;
  permissions?: string[];
  store_id?: string;
}
