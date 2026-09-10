"use client";

import React from "react";
import { NavbarFrame } from "./navbar-frame";
import { NavbarLeft } from "./navbar-left";
import { NavbarRight } from "./navbar-right";

export interface AuthNavbarProps {
  className?: string;
}

export function AuthNavbar({ className }: AuthNavbarProps) {
  return (
    <NavbarFrame
      className={className}
      left={<NavbarLeft homeHref="/" />}
      right={<NavbarRight showAuth={false} />}
    />
  );
}
