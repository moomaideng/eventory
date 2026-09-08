"use client";

import React from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useAuth } from "@/context/auth-context";
import { useRole, UserRole } from "@/context/role-context";
import { Card } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { cn } from "@/lib/utils";
import {
  Gamepad2,
  Trophy,
  Briefcase,
  ArrowRight,
  CheckCircle2,
} from "lucide-react";

interface PersonaConfig {
  id: UserRole;
  title: string;
  description: string;
  targetHref: string;
  icon: React.ElementType;
  profileName?: string;
}

export function ModeHub() {
  const router = useRouter();
  const { user, isLoading } = useAuth();
  const { activeRole, organizerProfile, sponsorProfile, setRole } = useRole();

  const personas: PersonaConfig[] = [
    {
      id: "competitor",
      title: "Competitor",
      description:
        "Browse open tournaments, join team lobbies, and compete in brackets.",
      targetHref: "/tournaments",
      icon: Gamepad2,
      profileName: user ? `@${user.handle}` : undefined,
    },
    {
      id: "organizer",
      title: "Organizer",
      description:
        "Host tournaments, set crowdfunding prize goals, manage match schedules, and invite staff.",
      targetHref: "/organizer",
      icon: Trophy,
      profileName: organizerProfile?.name || "Organizer Profile",
    },
    {
      id: "sponsor",
      title: "Sponsor",
      description:
        "Pledge funding to tournament prize pools, track sponsorships, and showcase your brand logo.",
      targetHref: "/sponsor",
      icon: Briefcase,
      profileName: sponsorProfile?.companyName || "Sponsor Profile",
    },
  ];

  const handleSelectPersona = (targetRole: UserRole, targetHref: string) => {
    setRole(targetRole);
    router.push(targetHref);
  };

  if (isLoading) {
    return (
      <div className="container mx-auto flex max-w-5xl flex-1 flex-col items-center justify-center px-4 py-20 sm:px-8">
        <Skeleton className="h-10 w-64 rounded-xl" />
        <Skeleton className="mt-3 h-5 w-80 rounded-md" />
        <div className="mt-12 grid w-full gap-6 md:grid-cols-3">
          <Skeleton className="h-72 rounded-xl" />
          <Skeleton className="h-72 rounded-xl" />
          <Skeleton className="h-72 rounded-xl" />
        </div>
      </div>
    );
  }

  if (!user) {
    return (
      <div className="container mx-auto flex max-w-md flex-1 flex-col items-center justify-center px-4 py-24 text-center">
        <h2 className="text-2xl font-bold tracking-tight">Sign In Required</h2>
        <p className="text-muted-foreground mt-2 text-sm">
          Please sign in to select your workspace mode and access your profile.
        </p>
        <Button
          className="mt-6"
          render={<Link href="/login" />}
          nativeButton={false}
        >
          Sign In
        </Button>
      </div>
    );
  }

  return (
    <div className="container mx-auto flex max-w-6xl flex-1 flex-col justify-center px-4 py-16 sm:px-8">
      {/* Header Greeting */}
      <div className="mb-12 text-center">
        <h1 className="text-3xl font-extrabold tracking-tight sm:text-5xl">
          Welcome back,{" "}
          <span className="text-primary">{user?.displayName || "Player"}</span>
        </h1>
        <p className="text-muted-foreground mx-auto mt-3 max-w-xl text-sm sm:text-base">
          Select what profile mode you want to use to interact with Eventory.
        </p>
      </div>

      {/* 3 Interactive Clickable Persona Cards */}
      <div className="grid gap-6 md:grid-cols-3">
        {personas.map(
          ({ id, title, description, targetHref, icon: Icon, profileName }) => {
            const isActive = activeRole === id;

            return (
              <Card
                key={id}
                role="button"
                tabIndex={0}
                onClick={() => handleSelectPersona(id, targetHref)}
                onKeyDown={(e) => {
                  if (e.key === "Enter" || e.key === " ") {
                    e.preventDefault();
                    handleSelectPersona(id, targetHref);
                  }
                }}
                className={cn(
                  "group relative flex cursor-pointer flex-col justify-between rounded-2xl p-6 transition-all duration-200 select-none",
                  "hover:-translate-y-1.5 hover:shadow-lg active:translate-y-0 active:scale-[0.99]",
                  isActive
                    ? "border-primary/80 bg-primary/3 ring-primary/40 shadow-xs ring-1"
                    : "hover:border-primary/60 hover:bg-muted/20"
                )}
              >
                <div>
                  <div className="flex items-center justify-between gap-2">
                    <div
                      className={cn(
                        "flex size-12 items-center justify-center rounded-xl transition-all duration-200",
                        isActive
                          ? "bg-primary text-primary-foreground shadow-xs"
                          : "bg-muted text-foreground group-hover:bg-primary group-hover:text-primary-foreground"
                      )}
                    >
                      <Icon className="size-6 transition-transform duration-200 group-hover:scale-110" />
                    </div>
                    {isActive ? (
                      <Badge variant="default" className="gap-1 text-xs">
                        <CheckCircle2 data-icon="inline-start" />
                        Current Mode
                      </Badge>
                    ) : null}
                  </div>

                  <div className="mt-6">
                    <h3 className="text-foreground group-hover:text-primary text-xl font-bold tracking-tight transition-colors">
                      {title}
                    </h3>
                    {profileName ? (
                      <p className="text-muted-foreground mt-1 truncate text-xs font-medium">
                        {profileName}
                      </p>
                    ) : null}
                  </div>

                  <p className="text-muted-foreground mt-3 text-sm leading-relaxed">
                    {description}
                  </p>
                </div>

                <div className="text-muted-foreground group-hover:text-primary mt-8 flex items-center gap-1.5 text-xs font-semibold transition-[color,transform] duration-200 group-hover:translate-x-1">
                  <span>
                    {isActive
                      ? "Continue as " + title
                      : "Enter " + title + " Mode"}
                  </span>
                  <ArrowRight className="size-3.5" />
                </div>
              </Card>
            );
          }
        )}
      </div>
    </div>
  );
}
