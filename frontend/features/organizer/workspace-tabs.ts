export type WorkspaceTab =
  | "home"
  | "registration"
  | "announcements"
  | "staff"
  | "form"
  | "sponsorship"
  | "bracket";

const WORKSPACE_TABS: readonly WorkspaceTab[] = [
  "home",
  "registration",
  "announcements",
  "staff",
  "form",
  "sponsorship",
  "bracket",
];

export function isWorkspaceTab(value: unknown): value is WorkspaceTab {
  return WORKSPACE_TABS.includes(value as WorkspaceTab);
}
