import { createContext, useContext } from "react";
import type {
  Command,
  Role,
  Workspace,
  ProductWorkspace,
  VariantResolution,
} from "./types";
export interface WorkspaceContext {
  state: Workspace;
  role: Role;
  surface: "backoffice" | "seller";
  loadProduct: (id: string) => Promise<ProductWorkspace>;
  resolveVariant: (
    id: string,
    values: { attribute_id: string; value: string }[],
  ) => Promise<VariantResolution>;
  execute: (command: Command) => Promise<void>;
  uploadIcon: (input: {
    kind: "category" | "brand";
    target: string;
    version: number;
    key: string;
    file: File;
  }) => Promise<void>;
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
