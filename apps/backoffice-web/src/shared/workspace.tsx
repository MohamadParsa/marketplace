import { createContext, useContext } from "react";
import type {
  Command,
  Role,
  Workspace,
  ProductWorkspace,
  VariantResolution,
  BannerList,
  ProductQAPage,
  ProductQAMutation,
} from "./types";
export interface WorkspaceContext {
  loadProfile: () => Promise<{
    name: string;
    email?: string;
    password_available: boolean;
  }>;
  updateProfile: (input: {
    name: string;
    email: string;
    password: string;
  }) => Promise<void>;
  loadCatalogMedia: (id: string) => Promise<Blob>;
  loadBanners: () => Promise<BannerList>;
  loadBannerAsset: (id: string) => Promise<Blob>;
  uploadBanner: (input: {
    id: string;
    slot: "desktop" | "mobile";
    version: number;
    key: string;
    file: File;
  }) => Promise<{ asset_id: string; version: number }>;
  state: Workspace;
  role: Role;
  surface: "backoffice" | "seller";
  loadProductQA: (id: string, after: number) => Promise<ProductQAPage>;
  changeProductQA: (id: string, question: string, mutation: ProductQAMutation, key: string) => Promise<void>;
  loadProduct: (id: string) => Promise<ProductWorkspace>;
  resolveVariant: (
    id: string,
    values: { attribute_id: string; value: string }[],
  ) => Promise<VariantResolution>;

  uploadIcon: (input: {
    kind: "category" | "brand" | "product" | "product-gallery";
    target: string;
    version: number;
    key: string;
    file: File;
  }) => Promise<void>;
  can: (action: string, area?: "management" | "support" | "finance") => boolean;
  execute: (command: Command) => Promise<void>;
  requestAdminCode: () => Promise<{ request_id: string; test_code?: string }>;
  busy: boolean;
}
export const WorkspaceContext = createContext<WorkspaceContext | null>(null);
export function useWorkspace() {
  const value = useContext(WorkspaceContext);
  if (!value) throw new Error("Workspace provider missing");
  return value;
}
export function can(
  role: Role,
  area: "management" | "operations" | "support" | "finance",
) {
  return (
    role === "super-admin" ||
    (area === "management" ? role === "operations" : role === area)
  );
}
