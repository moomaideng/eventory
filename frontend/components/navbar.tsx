'use client';

import React, { useState, useRef, useEffect } from 'react';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { useRole } from '@/context/role-context';
import { buttonVariants } from '@/components/ui/button';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { cn } from '@/lib/utils';
import {
  Gamepad2,
  Trophy,
  Briefcase,
  ChevronDown,
  LogOut,
  Sparkles,
  LogIn,
  Check,
} from 'lucide-react';

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
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  // Role visual configuration
  const roleConfig = {
    competitor: {
      label: 'Competitor',
      icon: Gamepad2,
      iconColor: 'text-emerald-500',
      bgColor: 'bg-emerald-500/10',
    },
    organizer: {
      label: 'Organizer',
      icon: Trophy,
      iconColor: 'text-amber-500',
      bgColor: 'bg-amber-500/10',
    },
    sponsor: {
      label: 'Sponsor',
      icon: Briefcase,
      iconColor: 'text-blue-500',
      bgColor: 'bg-blue-500/10',
    },
  };

  const CurrentRoleIcon = roleConfig[activeRole].icon;

  // Dynamic navigation links based on active role
  const getNavLinks = () => {
    switch (activeRole) {
      case 'organizer':
        return [
          { href: '#', label: 'My Tournaments' },
          { href: '#', label: '+ Host Tournament' },
        ];
      case 'sponsor':
        return [
          { href: '#', label: 'Sponsor Dashboard' },
          { href: '#', label: 'Fund Tournaments' },
        ];
      case 'competitor':
      default:
        return [
          { href: '#', label: 'Tournaments' },
          { href: '#', label: 'Lobby (Demo)' },
        ];
    }
  };

  return (
    <header className="sticky top-0 z-50 w-full border-b bg-background/95 backdrop-blur supports-backdrop-filter:bg-background/60">
      <div className="container mx-auto flex h-16 max-w-7xl items-center justify-between px-4 sm:px-8">
        {/* Left: Brand Logo */}
        <div className="flex items-center gap-8">
          <Link href="/" className="flex items-center gap-2">
            <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-primary text-primary-foreground font-black text-lg shadow-xs">
              E
            </div>
            <span className="text-xl font-bold tracking-tight text-foreground">
              Eventory<span className="text-primary">.</span>
            </span>
          </Link>

          {/* Center: Dynamic Navigation Links */}
          <nav className="hidden md:flex items-center gap-6">
            {getNavLinks().map((link, index) => (
              <span
                key={index}
                className="text-sm font-medium text-muted-foreground transition-colors hover:text-foreground cursor-pointer"
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
                className="flex items-center gap-2.5 h-10 px-3 py-1.5 rounded-full border border-border bg-card shadow-xs hover:bg-accent transition cursor-pointer"
                aria-expanded={isOpen}
              >
                <Avatar className="h-6 w-6">
                  <AvatarImage src={user.avatarUrl} alt={user.displayName} />
                  <AvatarFallback className="text-[10px] bg-muted">
                    {user.displayName.slice(0, 2).toUpperCase()}
                  </AvatarFallback>
                </Avatar>
                <div className="flex flex-col text-left">
                  <span className="text-xs font-semibold leading-tight text-foreground">
                    {activeProfileName}
                  </span>
                  <span className="text-[10px] text-muted-foreground capitalize flex items-center gap-1">
                    <CurrentRoleIcon
                      className={`h-2.5 w-2.5 ${roleConfig[activeRole].iconColor}`}
                    />
                    {roleConfig[activeRole].label} Mode
                  </span>
                </div>
                <ChevronDown
                  className={cn(
                    'h-3.5 w-3.5 text-muted-foreground ml-1 transition-transform',
                    isOpen && 'rotate-180'
                  )}
                />
              </button>

              {/* Dropdown Menu Popup */}
              {isOpen && (
                <div className="absolute right-0 mt-2 w-72 rounded-xl border border-border bg-popover p-2 text-popover-foreground shadow-xl z-50 animate-in fade-in-0 zoom-in-95">
                  <div className="px-2.5 py-1.5">
                    <p className="text-xs font-medium text-muted-foreground">
                      Switch Role Context
                    </p>
                  </div>

                  <div className="h-px bg-border my-1" />

                  <div className="space-y-1">
                    {/* 1. Competitor Mode */}
                    <button
                      type="button"
                      onClick={() => {
                        setRole('competitor');
                        setIsOpen(false);
                      }}
                      className={cn(
                        'w-full flex items-center justify-between p-2 rounded-lg text-left transition hover:bg-accent cursor-pointer',
                        activeRole === 'competitor' && 'bg-accent/80'
                      )}
                    >
                      <div className="flex items-center gap-2.5">
                        <div className="flex h-7 w-7 items-center justify-center rounded-md bg-emerald-500/10 text-emerald-500">
                          <Gamepad2 className="h-4 w-4" />
                        </div>
                        <div className="flex flex-col">
                          <span className="text-sm font-medium">Competitor Mode</span>
                          <span className="text-xs text-muted-foreground">
                            {user.displayName}
                          </span>
                        </div>
                      </div>
                      {activeRole === 'competitor' && (
                        <Check className="h-4 w-4 text-primary" />
                      )}
                    </button>

                    {/* 2. Organizer Mode */}
                    <button
                      type="button"
                      onClick={() => {
                        setRole('organizer');
                        setIsOpen(false);
                      }}
                      className={cn(
                        'w-full flex items-center justify-between p-2 rounded-lg text-left transition hover:bg-accent cursor-pointer',
                        activeRole === 'organizer' && 'bg-accent/80'
                      )}
                    >
                      <div className="flex items-center gap-2.5">
                        <div className="flex h-7 w-7 items-center justify-center rounded-md bg-amber-500/10 text-amber-500">
                          <Trophy className="h-4 w-4" />
                        </div>
                        <div className="flex flex-col">
                          <span className="text-sm font-medium">Organizer Mode</span>
                          <span className="text-xs text-muted-foreground">
                            {organizerProfile?.name || 'Chula Esports Club'}
                          </span>
                        </div>
                      </div>
                      {activeRole === 'organizer' && (
                        <Check className="h-4 w-4 text-primary" />
                      )}
                    </button>

                    {/* 3. Sponsor Mode */}
                    <button
                      type="button"
                      onClick={() => {
                        setRole('sponsor');
                        setIsOpen(false);
                      }}
                      className={cn(
                        'w-full flex items-center justify-between p-2 rounded-lg text-left transition hover:bg-accent cursor-pointer',
                        activeRole === 'sponsor' && 'bg-accent/80'
                      )}
                    >
                      <div className="flex items-center gap-2.5">
                        <div className="flex h-7 w-7 items-center justify-center rounded-md bg-blue-500/10 text-blue-500">
                          <Briefcase className="h-4 w-4" />
                        </div>
                        <div className="flex flex-col">
                          <span className="text-sm font-medium">Sponsor Mode</span>
                          <span className="text-xs text-muted-foreground">
                            {sponsorProfile?.companyName || 'Red Bull Gaming'}
                          </span>
                        </div>
                      </div>
                      {activeRole === 'sponsor' && (
                        <Check className="h-4 w-4 text-primary" />
                      )}
                    </button>
                  </div>

                  <div className="h-px bg-border my-1.5" />

                  {/* Sign out */}
                  <button
                    type="button"
                    onClick={() => {
                      logout();
                      setIsOpen(false);
                    }}
                    className="w-full flex items-center gap-2 p-2 rounded-lg text-xs font-medium text-destructive hover:bg-destructive/10 transition cursor-pointer"
                  >
                    <LogOut className="h-4 w-4" />
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
                onClick={() => loginAsDev('competitor')}
                className={cn(
                  buttonVariants({ variant: 'outline', size: 'sm' }),
                  'hidden sm:flex items-center gap-1.5 text-xs text-primary border-primary/30 hover:bg-primary/10 cursor-pointer'
                )}
              >
                <Sparkles className="h-3.5 w-3.5" />
                Dev Quick Login
              </button>

              <Link
                href="/login"
                className={cn(
                  buttonVariants({ size: 'sm' }),
                  'flex items-center gap-1.5 text-xs cursor-pointer shadow-xs'
                )}
              >
                <LogIn className="h-3.5 w-3.5" />
                Sign In
              </Link>
            </div>
          )}
        </div>
      </div>
    </header>
  );
}
