export type WorkspaceTab =
  | "home"
  | "registration"
  | "announcements"
  | "staff"
  | "form"
  | "sponsorship";

const WORKSPACE_TABS: readonly WorkspaceTab[] = [
  "home",
  "registration",
  "announcements",
  "staff",
  "form",
  "sponsorship",
];

export function isWorkspaceTab(value: unknown): value is WorkspaceTab {
  return WORKSPACE_TABS.includes(value as WorkspaceTab);
}
