"use client";

import React from "react";
import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { useAuth } from "@/context/auth-context";
import { useRole, UserRole } from "@/context/role-context";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
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
import {
  Gamepad2,
  Trophy,
  Briefcase,
  ChevronDown,
  LogOut,
  Sparkles,
  LogIn,
  LayoutGrid,
} from "lucide-react";

const ROLE_ICONS: Record<UserRole, React.ElementType> = {
  competitor: Gamepad2,
  organizer: Trophy,
  sponsor: Briefcase,
};

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
  const { activeRole } = useRole();

  const ActiveRoleIcon = ROLE_ICONS[activeRole] || Gamepad2;

  return (
    <header className="bg-background/95 sticky top-0 z-50 w-full border-b backdrop-blur">
      <div className="container mx-auto flex h-16 max-w-7xl items-center justify-between px-4 sm:px-8">
        {/* Left: Brand Logo & Navigation */}
        <div className="flex items-center gap-8">
          <Link href="/" className="flex items-center gap-2">
            <div className="bg-primary text-primary-foreground flex size-9 items-center justify-center rounded-lg font-black">
              E
            </div>
            <span className="text-foreground text-xl font-bold">
              Eventory<span className="text-primary">.</span>
            </span>
          </Link>

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
            <Skeleton className="h-9 w-28 rounded-full" />
          ) : user ? (
            /* Unified Profile & Mode Dropdown */
            <DropdownMenu>
              <DropdownMenuTrigger
                render={
                  <Button
                    variant="outline"
                    className="flex h-10 items-center gap-2 rounded-full pr-3 pl-2.5"
                  />
                }
              >
                <Avatar className="size-6">
                  <AvatarImage src={user.avatarUrl} alt={user.displayName} />
                  <AvatarFallback className="text-[10px]">
                    {user.displayName.slice(0, 2).toUpperCase()}
                  </AvatarFallback>
                </Avatar>
                <span className="text-foreground hidden max-w-32.5 truncate text-xs font-semibold sm:inline">
                  {user.displayName}
                </span>
                <Badge
                  variant="secondary"
                  className="gap-1 px-2 py-0.5 text-[11px] font-medium capitalize"
                >
                  <ActiveRoleIcon data-icon="inline-start" />
                  <span>{activeRole}</span>
                </Badge>
                <ChevronDown
                  className="text-muted-foreground"
                  data-icon="inline-end"
                />
              </DropdownMenuTrigger>

              <DropdownMenuContent align="end" className="w-56 p-1">
                <DropdownMenuGroup>
                  {/* Compact User Header */}
                  <div className="px-3 py-2">
                    <p className="text-foreground truncate text-xs font-semibold">
                      {user.displayName}
                    </p>
                    <p className="text-muted-foreground truncate text-[11px]">
                      @{user.handle}
                    </p>
                  </div>

                  <DropdownMenuSeparator />

                  {/* Single Cohesive Action to Navigate to Mode Hub */}
                  <DropdownMenuItem
                    onClick={() => router.push("/hub")}
                    className="flex cursor-pointer items-center gap-2.5 py-2"
                  >
                    <LayoutGrid className="text-primary" />
                    <span className="text-sm font-medium">Switch Mode Hub</span>
                  </DropdownMenuItem>
                </DropdownMenuGroup>

                <DropdownMenuSeparator />

                <DropdownMenuGroup>
                  <DropdownMenuItem
                    onClick={logout}
                    variant="destructive"
                    className="flex cursor-pointer items-center gap-2"
                  >
                    <LogOut />
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
