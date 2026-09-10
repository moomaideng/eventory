"use client";

import React from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { cn } from "@/lib/utils";

export interface NavLinkItem {
  label: string;
  href: string;
}

export interface NavbarLeftProps {
  /** Target link for the brand logo, defaults to "/" */
  homeHref?: string;
  /** Section subtitle/badge under or next to Eventory, e.g. "Organizer", "Sponsor" */
  badge?: string;
  /** Accessible title for the logo link, defaults to "Eventory Home" */
  logoTitle?: string;
  /** Navigation links for this section */
  links?: NavLinkItem[];
  /** Optional custom slot or children */
  children?: React.ReactNode;
}

export function NavbarLeft({
  homeHref = "/",
  badge,
  logoTitle = "Eventory Home",
  links,
  children,
}: NavbarLeftProps) {
  const pathname = usePathname();

  return (
    <div className="flex items-center gap-8">
      {/* Brand Logo & Title */}
      <div className="flex items-center gap-2.5">
        <Link
          href={homeHref}
          title={logoTitle}
          className="bg-primary text-primary-foreground flex size-9 shrink-0 items-center justify-center rounded-lg font-black transition-transform select-none hover:scale-105"
        >
          E
        </Link>

        <Link href={homeHref} className="transition-opacity hover:opacity-90">
          {badge ? (
            <div className="flex flex-col text-left leading-none">
              <span className="text-foreground text-sm font-bold tracking-tight">
                Eventory<span className="text-primary">.</span>
              </span>
              <span className="text-muted-foreground mt-0.5 text-[10px] font-bold tracking-wider uppercase">
                {badge}
              </span>
            </div>
          ) : (
            <span className="text-foreground text-lg font-bold tracking-tight">
              Eventory<span className="text-primary">.</span>
            </span>
          )}
        </Link>
      </div>

      {/* Navigation Links */}
      {links && links.length > 0 && (
        <nav className="hidden items-center gap-6 md:flex">
          {links.map((link) => {
            const isActive = pathname === link.href;
            return (
              <Link
                key={link.href}
                href={link.href}
                className={cn(
                  "hover:text-foreground text-sm font-medium transition-colors",
                  isActive
                    ? "text-foreground font-semibold"
                    : "text-muted-foreground"
                )}
              >
                {link.label}
              </Link>
            );
          })}
        </nav>
      )}

      {children}
    </div>
  );
}
