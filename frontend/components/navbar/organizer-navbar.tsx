"use client";

import React from "react";
import { NavbarFrame } from "./navbar-frame";
import { NavbarLeft, type NavLinkItem } from "./navbar-left";
import { NavbarRight } from "./navbar-right";

const ORGANIZER_NAV_LINKS: NavLinkItem[] = [
  { label: "My Tournaments", href: "/organizer" },
  { label: "Host a Tournament", href: "/organizer/tournaments/new" },
];

export interface OrganizerNavbarProps {
  className?: string;
}

export function OrganizerNavbar({ className }: OrganizerNavbarProps) {
  return (
    <NavbarFrame
      className={className}
      left={
        <NavbarLeft
          homeHref="/organizer"
          badge="Organizer"
          links={ORGANIZER_NAV_LINKS}
        />
      }
      right={<NavbarRight role="organizer" />}
    />
  );
}
