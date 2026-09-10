"use client";

import React from "react";
import { NavbarFrame } from "./navbar-frame";
import { NavbarLeft, type NavLinkItem } from "./navbar-left";
import { NavbarRight } from "./navbar-right";

const SPONSOR_NAV_LINKS: NavLinkItem[] = [
  { label: "Sponsor Dashboard", href: "/sponsor" },
  { label: "Fund Tournaments", href: "/tournaments?filter=crowdfunding" },
];

export interface SponsorNavbarProps {
  className?: string;
}

export function SponsorNavbar({ className }: SponsorNavbarProps) {
  return (
    <NavbarFrame
      className={className}
      left={
        <NavbarLeft
          homeHref="/sponsor"
          badge="Sponsor"
          links={SPONSOR_NAV_LINKS}
        />
      }
      right={<NavbarRight role="sponsor" />}
    />
  );
}
