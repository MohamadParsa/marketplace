import { createContext, useContext } from "react";
import type { Command, Role, Workspace } from "./types";
export interface WorkspaceContext {
  state: Workspace;
  role: Role;
  execute: (command: Command) => Promise<void>;
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
