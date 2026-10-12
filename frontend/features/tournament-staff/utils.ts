import type { components } from "@/lib/api/schema";

export type StaffMember = components["schemas"]["TournamentStaffResponse"];
export type StaffRole = StaffMember["role"];

export const ROLE_OPTIONS: { value: "ADMIN" | "REFEREE"; label: string }[] = [
  { value: "ADMIN", label: "Admin" },
  { value: "REFEREE", label: "Referee" },
];

export const ROLE_PRIORITY: Record<string, number> = {
  OWNER: 1,
  ADMIN: 2,
  REFEREE: 3,
};

export function sortStaffMembers(staff: StaffMember[]): StaffMember[] {
  return [...staff].sort((a, b) => {
    const priorityDiff =
      (ROLE_PRIORITY[a.role] ?? 99) - (ROLE_PRIORITY[b.role] ?? 99);
    if (priorityDiff !== 0) return priorityDiff;
    return new Date(a.createdAt).getTime() - new Date(b.createdAt).getTime();
  });
}

export function formatRoleTitle(role: string): string {
  switch (role) {
    case "OWNER":
      return "Owner";
    case "ADMIN":
      return "Admin";
    case "REFEREE":
      return "Referee";
    default:
      return role.charAt(0).toUpperCase() + role.slice(1).toLowerCase();
  }
}

export function getErrorMessage(error: unknown): string {
  if (
    error &&
    typeof error === "object" &&
    "detail" in error &&
    typeof (error as { detail: unknown }).detail === "string"
  ) {
    return (error as { detail: string }).detail;
  }
  return "Something went wrong. Please try again.";
}
