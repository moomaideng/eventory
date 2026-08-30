"use client";

import React, { useState, useRef, useEffect } from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { useRole } from "@/context/role-context";
import { buttonVariants } from "@/components/ui/button";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { cn } from "@/lib/utils";
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

export function Navbar() {
  const pathname = usePathname();
  const {
    user,
    activeRole,
    activeProfileName,
    organizerProfile,
    sponsorProfile,
    setRole,
    loginAsDev,
    logout,
  } = useRole();

  // State to control dropdown visibility
  const [isOpen, setIsOpen] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);

  // Close dropdown on outside click
  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (
        dropdownRef.current &&
        !dropdownRef.current.contains(event.target as Node)
      ) {
        setIsOpen(false);
      }
    }
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, []);

  // Role visual configuration using semantic tokens
  const roleConfig = {
    competitor: {
      label: "Competitor",
      icon: Gamepad2,
      badgeColor: "bg-primary/10 text-primary",
    },
    organizer: {
      label: "Organizer",
      icon: Trophy,
      badgeColor: "bg-amber-500/10 text-amber-600 dark:text-amber-400",
    },
    sponsor: {
      label: "Sponsor",
      icon: Briefcase,
      badgeColor: "bg-blue-500/10 text-blue-600 dark:text-blue-400",
    },
  };

  const CurrentRoleIcon = roleConfig[activeRole].icon;

  // Dynamic navigation links based on active role
  const getNavLinks = () => {
    switch (activeRole) {
      case "organizer":
        return [
          { href: "#", label: "My Tournaments" },
          { href: "#", label: "+ Host Tournament" },
        ];
      case "sponsor":
        return [
          { href: "#", label: "Sponsor Dashboard" },
          { href: "#", label: "Fund Tournaments" },
        ];
      case "competitor":
      default:
        return [
          { href: "#", label: "Tournaments" },
          { href: "#", label: "Lobby (Demo)" },
        ];
    }
  };

  return (
    <header className="bg-background/95 supports-[backdrop-filter]:bg-background/60 sticky top-0 z-50 w-full border-b backdrop-blur">
      <div className="container mx-auto flex h-16 max-w-7xl items-center justify-between px-4 sm:px-8">
        {/* Left: Brand Logo */}
        <div className="flex items-center gap-8">
          <Link href="/" className="flex items-center gap-2">
            <div className="bg-primary text-primary-foreground flex size-9 items-center justify-center rounded-lg text-lg font-black shadow-xs">
              E
            </div>
            <span className="text-foreground text-xl font-bold tracking-tight">
              Eventory<span className="text-primary">.</span>
            </span>
          </Link>

          {/* Center: Dynamic Navigation Links */}
          <nav className="hidden items-center gap-6 md:flex">
            {getNavLinks().map((link, index) => (
              <span
                key={index}
                className="text-muted-foreground hover:text-foreground cursor-pointer text-sm font-medium transition-colors"
              >
                {link.label}
              </span>
            ))}
          </nav>
        </div>

        {/* Right: Role Switcher & Auth Actions */}
        <div className="flex items-center gap-3">
          {user ? (
            /* Logged In: Role Switcher Dropdown */
            <div className="relative" ref={dropdownRef}>
              <button
                type="button"
                onClick={() => setIsOpen(!isOpen)}
                className="border-border bg-card hover:bg-accent flex h-10 cursor-pointer items-center gap-2.5 rounded-full border px-3 py-1.5 shadow-xs transition"
                aria-expanded={isOpen}
              >
                <Avatar className="size-6">
                  <AvatarImage src={user.avatarUrl} alt={user.displayName} />
                  <AvatarFallback className="bg-muted text-[10px]">
                    {user.displayName.slice(0, 2).toUpperCase()}
                  </AvatarFallback>
                </Avatar>
                <div className="flex flex-col text-left">
                  <span className="text-foreground text-xs leading-tight font-semibold">
                    {activeProfileName}
                  </span>
                  <span className="text-muted-foreground flex items-center gap-1 text-[10px] capitalize">
                    <CurrentRoleIcon className="size-2.5" />
                    {roleConfig[activeRole].label} Mode
                  </span>
                </div>
                <ChevronDown
                  className={cn(
                    "text-muted-foreground ml-1 size-3.5 transition-transform",
                    isOpen && "rotate-180"
                  )}
                />
              </button>

              {/* Dropdown Menu Popup */}
              {isOpen && (
                <div className="border-border bg-popover text-popover-foreground animate-in fade-in-0 zoom-in-95 absolute right-0 z-50 mt-2 w-72 rounded-xl border p-2 shadow-xl">
                  <div className="px-2.5 py-1.5">
                    <p className="text-muted-foreground text-xs font-medium">
                      Switch Role Context
                    </p>
                  </div>

                  <div className="bg-border my-1 h-px" />

                  <div className="flex flex-col gap-1">
                    {/* 1. Competitor Mode */}
                    <button
                      type="button"
                      onClick={() => {
                        setRole("competitor");
                        setIsOpen(false);
                      }}
                      className={cn(
                        "hover:bg-accent flex w-full cursor-pointer items-center justify-between rounded-lg p-2 text-left transition",
                        activeRole === "competitor" && "bg-accent/80"
                      )}
                    >
                      <div className="flex items-center gap-2.5">
                        <div className="bg-primary/10 text-primary flex size-7 items-center justify-center rounded-md">
                          <Gamepad2 className="size-4" />
                        </div>
                        <div className="flex flex-col">
                          <span className="text-sm font-medium">
                            Competitor Mode
                          </span>
                          <span className="text-muted-foreground text-xs">
                            {user.displayName}
                          </span>
                        </div>
                      </div>
                      {activeRole === "competitor" && (
                        <Check className="text-primary size-4" />
                      )}
                    </button>

                    {/* 2. Organizer Mode */}
                    <button
                      type="button"
                      onClick={() => {
                        setRole("organizer");
                        setIsOpen(false);
                      }}
                      className={cn(
                        "hover:bg-accent flex w-full cursor-pointer items-center justify-between rounded-lg p-2 text-left transition",
                        activeRole === "organizer" && "bg-accent/80"
                      )}
                    >
                      <div className="flex items-center gap-2.5">
                        <div className="flex size-7 items-center justify-center rounded-md bg-amber-500/10 text-amber-600 dark:text-amber-400">
                          <Trophy className="size-4" />
                        </div>
                        <div className="flex flex-col">
                          <span className="text-sm font-medium">
                            Organizer Mode
                          </span>
                          <span className="text-muted-foreground text-xs">
                            {organizerProfile?.name || "Chula Esports Club"}
                          </span>
                        </div>
                      </div>
                      {activeRole === "organizer" && (
                        <Check className="text-primary size-4" />
                      )}
                    </button>

                    {/* 3. Sponsor Mode */}
                    <button
                      type="button"
                      onClick={() => {
                        setRole("sponsor");
                        setIsOpen(false);
                      }}
                      className={cn(
                        "hover:bg-accent flex w-full cursor-pointer items-center justify-between rounded-lg p-2 text-left transition",
                        activeRole === "sponsor" && "bg-accent/80"
                      )}
                    >
                      <div className="flex items-center gap-2.5">
                        <div className="flex size-7 items-center justify-center rounded-md bg-blue-500/10 text-blue-600 dark:text-blue-400">
                          <Briefcase className="size-4" />
                        </div>
                        <div className="flex flex-col">
                          <span className="text-sm font-medium">
                            Sponsor Mode
                          </span>
                          <span className="text-muted-foreground text-xs">
                            {sponsorProfile?.companyName || "Red Bull Gaming"}
                          </span>
                        </div>
                      </div>
                      {activeRole === "sponsor" && (
                        <Check className="text-primary size-4" />
                      )}
                    </button>
                  </div>

                  <div className="bg-border my-1.5 h-px" />

                  {/* Sign out */}
                  <button
                    type="button"
                    onClick={() => {
                      logout();
                      setIsOpen(false);
                    }}
                    className="text-destructive hover:bg-destructive/10 flex w-full cursor-pointer items-center gap-2 rounded-lg p-2 text-xs font-medium transition"
                  >
                    <LogOut className="size-4" />
                    <span>Sign Out</span>
                  </button>
                </div>
              )}
            </div>
          ) : (
            /* Unauthenticated: Sign In & Quick Dev Login */
            <div className="flex items-center gap-2">
              <button
                type="button"
                onClick={() => loginAsDev("competitor")}
                className={cn(
                  buttonVariants({ variant: "outline", size: "sm" }),
                  "text-primary border-primary/30 hover:bg-primary/10 hidden cursor-pointer items-center gap-1.5 text-xs sm:flex"
                )}
              >
                <Sparkles className="size-3.5" />
                Dev Quick Login
              </button>

              <Link
                href="/login"
                className={cn(
                  buttonVariants({ size: "sm" }),
                  "flex cursor-pointer items-center gap-1.5 text-xs shadow-xs"
                )}
              >
                <LogIn className="size-3.5" />
                Sign In
              </Link>
            </div>
          )}
        </div>
      </div>
    </header>
  );
}
