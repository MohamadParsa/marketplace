import { createContext, useContext } from "react";
import type { Command, Role, Workspace } from "./types";
export interface WorkspaceContext {
  state: Workspace;
  role: Role;
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
