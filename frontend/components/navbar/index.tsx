"use client";

import React from "react";
import { usePathname } from "next/navigation";
import { getWorkspaceRoleFromPath } from "@/lib/role";
import { CompetitorNavbar } from "./competitor-navbar";
import { OrganizerNavbar } from "./organizer-navbar";
import { SponsorNavbar } from "./sponsor-navbar";

export * from "./navbar-frame";
export * from "./navbar-left";
export * from "./navbar-right";
export * from "./competitor-navbar";
export * from "./organizer-navbar";
export * from "./sponsor-navbar";
export * from "./auth-navbar";

export interface NavbarProps {
  className?: string;
}

/**
 * Route-aware Navbar dispatcher.
 * Selects between CompetitorNavbar, OrganizerNavbar, and SponsorNavbar
 * based on current workspace path context.
 */
export function Navbar({ className }: NavbarProps) {
  const pathname = usePathname();
  const role = getWorkspaceRoleFromPath(pathname);

  if (role === "organizer") {
    return <OrganizerNavbar className={className} />;
  }
  if (role === "sponsor") {
    return <SponsorNavbar className={className} />;
  }
  return <CompetitorNavbar className={className} />;
}
