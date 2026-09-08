"use client";

import React from "react";
import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { useAuth } from "@/context/auth-context";
import { getWorkspaceRoleFromPath, type UserRole } from "@/lib/role";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { cn } from "@/lib/utils";
import { ThemeToggle } from "@/components/theme-toggle";
import { ChevronDown, LogOut, Sparkles, LogIn, Check } from "lucide-react";

const PERSONA_CONFIG: {
  id: UserRole;
  label: string;
  href: string;
}[] = [
  {
    id: "competitor",
    label: "Competitor",
    href: "/tournaments",
  },
  {
    id: "organizer",
    label: "Organizer",
    href: "/organizer",
  },
  {
    id: "sponsor",
    label: "Sponsor",
    href: "/sponsor",
  },
];

const NAV_LINKS: Record<UserRole, { label: string; href: string }[]> = {
  competitor: [{ label: "Tournaments", href: "/tournaments" }],
  organizer: [
    { label: "My Tournaments", href: "/organizer" },
    { label: "+ Host Tournament", href: "/organizer/tournaments/new" },
  ],
  sponsor: [
    { label: "Sponsor Dashboard", href: "/sponsor" },
    { label: "Fund Tournaments", href: "/tournaments?filter=crowdfunding" },
  ],
};

export function Navbar() {
  const pathname = usePathname();
  const router = useRouter();
  const { user, isLoading, loginAsDev, logout } = useAuth();
  const activeRole = getWorkspaceRoleFromPath(pathname);

  const isOrganizerWorkspace = pathname.startsWith("/organizer");
  const isSponsorWorkspace = pathname.startsWith("/sponsor");

  const handleSelectRole = (targetHref: string) => {
    router.push(targetHref);
  };

  return (
    <header className="bg-background/95 sticky top-0 z-50 w-full border-b backdrop-blur">
      <div className="container mx-auto flex h-16 max-w-7xl items-center justify-between px-4 sm:px-8">
        {/* Left: Brand Logo & Navigation */}
        <div className="flex items-center gap-8">
          <div className="flex items-center gap-2.5">
            <Link
              href="/"
              title="Eventory Home"
              className="bg-primary text-primary-foreground flex size-9 shrink-0 items-center justify-center rounded-lg font-black transition-transform select-none hover:scale-105"
            >
              E
            </Link>

            <Link
              href={
                isOrganizerWorkspace
                  ? "/organizer"
                  : isSponsorWorkspace
                    ? "/sponsor"
                    : "/"
              }
              className="transition-opacity hover:opacity-90"
            >
              {isOrganizerWorkspace ? (
                <div className="flex flex-col text-left leading-none">
                  <span className="text-foreground text-sm font-bold tracking-tight">
                    Eventory
                  </span>
                  <span className="text-muted-foreground mt-0.5 text-[10px] font-bold tracking-wider uppercase">
                    Organizer
                  </span>
                </div>
              ) : isSponsorWorkspace ? (
                <div className="flex flex-col text-left leading-none">
                  <span className="text-foreground text-sm font-bold tracking-tight">
                    Eventory
                  </span>
                  <span className="text-muted-foreground mt-0.5 text-[10px] font-bold tracking-wider uppercase">
                    Sponsor
                  </span>
                </div>
              ) : (
                <span className="text-foreground text-lg font-bold tracking-tight">
                  Eventory
                </span>
              )}
            </Link>
          </div>

          {/* Dynamic Navigation Links */}
          <nav className="hidden items-center gap-6 md:flex">
            {NAV_LINKS[activeRole].map((link) => {
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
        </div>

        {/* Right: Theme Toggle & Unified Profile Trigger */}
        <div className="flex items-center gap-2.5">
          <ThemeToggle />
          {isLoading ? (
            /* Skeleton Loading State during Hydration (Zero Layout Shift) */
            <Skeleton className="h-9 w-28 rounded-lg" />
          ) : user ? (
            /* Unified Profile & Mode Dropdown with Squircle Styling */
            <DropdownMenu>
              <DropdownMenuTrigger
                render={
                  <Button
                    variant="outline"
                    className="group flex h-9 items-center gap-2 rounded-lg px-2.5 py-1 transition-colors"
                  />
                }
              >
                <Avatar className="ring-border/80 size-6 rounded-md ring-1 after:rounded-md">
                  <AvatarImage
                    src={user.avatarUrl}
                    alt={user.displayName}
                    className="rounded-md"
                  />
                  <AvatarFallback className="rounded-md text-[10px] font-bold">
                    {user.displayName.slice(0, 2).toUpperCase()}
                  </AvatarFallback>
                </Avatar>

                <span className="text-foreground hidden max-w-32 truncate text-xs font-semibold sm:inline">
                  {user.displayName}
                </span>

                <ChevronDown className="text-muted-foreground group-hover:text-foreground size-3 opacity-60 transition-transform duration-200 group-data-popup-open:rotate-180" />
              </DropdownMenuTrigger>

              <DropdownMenuContent
                align="end"
                className="w-56 rounded-xl border p-2 shadow-xl"
              >
                <DropdownMenuGroup>
                  {/* Compact User Header */}
                  <div className="flex items-center gap-3 px-2 py-1.5">
                    <Avatar className="ring-border size-8 rounded-lg ring-1 after:rounded-lg">
                      <AvatarImage
                        src={user.avatarUrl}
                        alt={user.displayName}
                        className="rounded-lg"
                      />
                      <AvatarFallback className="rounded-lg text-[10px] font-bold">
                        {user.displayName.slice(0, 2).toUpperCase()}
                      </AvatarFallback>
                    </Avatar>
                    <div className="flex min-w-0 flex-col">
                      <p className="text-foreground truncate text-sm leading-tight font-semibold">
                        {user.displayName}
                      </p>
                      <p className="text-muted-foreground truncate text-xs">
                        @{user.handle}
                      </p>
                    </div>
                  </div>

                  <DropdownMenuSeparator className="my-1.5" />

                  {/* Quick Switch Section Header */}
                  <div className="px-1 pb-1">
                    <span className="text-muted-foreground text-[10px] font-bold tracking-wider uppercase">
                      Quick Switch
                    </span>
                  </div>

                  {/* Vertical Segmented Slider Track */}
                  <div className="bg-muted/50 border-border/40 flex flex-col gap-1 rounded-lg border p-1 dark:bg-black/30">
                    {PERSONA_CONFIG.map(({ id, label, href }) => {
                      const isActive = activeRole === id;
                      return (
                        <DropdownMenuItem
                          key={id}
                          onClick={() => handleSelectRole(href)}
                          className={cn(
                            "flex cursor-pointer items-center justify-between rounded-md px-2.5 py-1.5 text-xs transition-all duration-150 select-none",
                            isActive
                              ? "bg-background text-foreground border-border/60 dark:bg-muted focus:bg-background dark:focus:bg-muted focus:text-foreground border font-semibold shadow-xs dark:border-white/10"
                              : "text-muted-foreground hover:bg-background/30 hover:text-foreground focus:bg-background/30 focus:text-foreground dark:hover:bg-white/5 dark:focus:bg-white/5"
                          )}
                        >
                          <span>{label}</span>
                          {isActive && (
                            <Check className="text-foreground/70 ml-2 size-3.5 shrink-0" />
                          )}
                        </DropdownMenuItem>
                      );
                    })}
                  </div>
                </DropdownMenuGroup>

                <DropdownMenuSeparator className="my-1.5" />

                <DropdownMenuGroup>
                  <DropdownMenuItem
                    onClick={logout}
                    variant="destructive"
                    className="text-destructive hover:bg-destructive/10 flex cursor-pointer items-center gap-2 rounded-lg px-2.5 py-2 text-xs font-medium"
                  >
                    <LogOut className="size-3.5" />
                    <span>Sign Out</span>
                  </DropdownMenuItem>
                </DropdownMenuGroup>
              </DropdownMenuContent>
            </DropdownMenu>
          ) : (
            /* Unauthenticated: Dev Login & Sign In */
            <div className="flex items-center gap-2">
              <Button
                variant="outline"
                size="sm"
                onClick={() => loginAsDev("competitor")}
                className="text-primary border-primary/30 hidden text-xs sm:inline-flex"
              >
                <Sparkles data-icon="inline-start" />
                Dev Quick Login
              </Button>

              <Button
                size="sm"
                render={<Link href="/login" />}
                nativeButton={false}
                className="text-xs"
              >
                <LogIn data-icon="inline-start" />
                Sign In
              </Button>
            </div>
          )}
        </div>
      </div>
    </header>
  );
}
