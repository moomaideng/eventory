export type UserRole = "competitor" | "organizer" | "sponsor";

export const DEFAULT_ROLE: UserRole = "competitor";

export function isValidRole(value: unknown): value is UserRole {
  return value === "competitor" || value === "organizer" || value === "sponsor";
}

/**
 * Derives the active workspace role directly from the URL pathname.
 * Aside from /organizer and /sponsor, everything else is treated as competitor.
 */
export function getWorkspaceRoleFromPath(pathname: string): UserRole {
  if (pathname === "/organizer" || pathname.startsWith("/organizer/")) {
    return "organizer";
  }
  if (pathname === "/sponsor" || pathname.startsWith("/sponsor/")) {
    return "sponsor";
  }
  return "competitor";
}

export interface OrganizerProfile {
  id: string;
  name: string;
  bio?: string;
  logoUrl?: string;
}

export interface SponsorProfile {
  id: string;
  companyName: string;
  websiteUrl?: string;
  logoUrl?: string;
}
