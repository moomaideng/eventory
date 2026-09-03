"use client";

import React from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { useRole, UserRole } from "@/context/role-context";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuLabel,
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
  Check,
} from "lucide-react";

const ROLES: { id: UserRole; label: string; icon: React.ElementType }[] = [
  { id: "competitor", label: "Competitor", icon: Gamepad2 },
  { id: "organizer", label: "Organizer", icon: Trophy },
  { id: "sponsor", label: "Sponsor", icon: Briefcase },
];

const NAV_LINKS: Record<UserRole, { label: string; href: string }[]> = {
  competitor: [
    { label: "Tournaments", href: "/tournaments" },
    { label: "Lobby (Demo)", href: "/lobbies/DEMO123" },
  ],
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
  const {
    user,
    activeRole,
    activeProfileName,
    isLoading,
    setRole,
    loginAsDev,
    logout,
  } = useRole();

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

        {/* Right: Role Switcher & Auth Actions */}
        <div className="flex items-center gap-3">
          <ThemeToggle />
          {isLoading ? (
            /* Skeleton Loading State during Hydration (Zero Layout Shift) */
            <Skeleton className="h-9 w-28 rounded-full" />
          ) : user ? (
            /* Role Switcher Dropdown */
            <DropdownMenu>
              <DropdownMenuTrigger
                render={
                  <Button
                    variant="outline"
                    className="flex h-10 items-center gap-2.5 rounded-full px-3"
                  />
                }
              >
                <Avatar className="size-6">
                  <AvatarImage src={user.avatarUrl} alt={user.displayName} />
                  <AvatarFallback className="text-[10px]">
                    {user.displayName.slice(0, 2).toUpperCase()}
                  </AvatarFallback>
                </Avatar>
                <div className="flex flex-col text-left">
                  <span className="text-foreground text-xs leading-tight font-semibold">
                    {activeProfileName}
                  </span>
                  <span className="text-muted-foreground text-[10px] capitalize">
                    {activeRole} Mode
                  </span>
                </div>
                <ChevronDown className="text-muted-foreground" />
              </DropdownMenuTrigger>

              <DropdownMenuContent align="end" className="w-64 p-1">
                <DropdownMenuGroup>
                  <DropdownMenuLabel>Switch Role Context</DropdownMenuLabel>
                  <DropdownMenuSeparator />
                  {ROLES.map(({ id, label, icon: Icon }) => (
                    <DropdownMenuItem
                      key={id}
                      onClick={() => setRole(id)}
                      className="flex items-center justify-between py-2"
                    >
                      <div className="flex items-center gap-2.5">
                        <Icon className="text-primary" />
                        <span className="text-sm font-medium">
                          {label} Mode
                        </span>
                      </div>
                      {activeRole === id && <Check className="text-primary" />}
                    </DropdownMenuItem>
                  ))}
                </DropdownMenuGroup>

                <DropdownMenuSeparator />

                <DropdownMenuGroup>
                  <DropdownMenuItem
                    onClick={logout}
                    variant="destructive"
                    className="flex items-center gap-2"
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
