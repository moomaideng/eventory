"use client";

import React from "react";
import { NavbarFrame } from "./navbar-frame";
import { NavbarLeft, type NavLinkItem } from "./navbar-left";
import { NavbarRight } from "./navbar-right";

const COMPETITOR_NAV_LINKS: NavLinkItem[] = [
  { label: "Tournaments", href: "/tournaments" },
  { label: "Join a Lobby", href: "/lobbies" },
];

export interface CompetitorNavbarProps {
  className?: string;
}

export function CompetitorNavbar({ className }: CompetitorNavbarProps) {
  return (
    <NavbarFrame
      className={className}
      left={<NavbarLeft homeHref="/" links={COMPETITOR_NAV_LINKS} />}
      right={<NavbarRight role="competitor" />}
    />
  );
}
