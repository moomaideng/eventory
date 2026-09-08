"use client";

import React from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useAuth } from "@/context/auth-context";
import type { UserRole } from "@/lib/role";
import { Card } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { cn } from "@/lib/utils";
import { Gamepad2, Trophy, Briefcase, ArrowRight } from "lucide-react";

interface PersonaConfig {
  id: UserRole;
  title: string;
  description: string;
  targetHref: string;
  actionText: string;
  icon: React.ElementType;
  profileName?: string;
}

export function ModeHub() {
  const router = useRouter();
  const { user, isLoading } = useAuth();

  const personas: PersonaConfig[] = [
    {
      id: "competitor",
      title: "Competitor",
      description:
        "Browse open tournaments, join team lobbies, and compete in brackets.",
      targetHref: "/tournaments",
      actionText: "Browse Tournaments",
      icon: Gamepad2,
      profileName: user ? `@${user.handle}` : undefined,
    },
    {
      id: "organizer",
      title: "Organizer",
      description:
        "Host tournaments, set crowdfunding prize goals, manage match schedules, and invite staff.",
      targetHref: "/organizer",
      actionText: "Open Organizer Workspace",
      icon: Trophy,
      profileName: "Organizer Workspace",
    },
    {
      id: "sponsor",
      title: "Sponsor",
      description:
        "Pledge funding to tournament prize pools, track sponsorships, and showcase your brand logo.",
      targetHref: "/sponsor",
      actionText: "Open Sponsor Workspace",
      icon: Briefcase,
      profileName: "Sponsor Workspace",
    },
  ];

  const handleSelectPersona = (targetHref: string) => {
    router.push(targetHref);
  };

  if (isLoading) {
    return (
      <div className="container mx-auto flex max-w-5xl flex-1 flex-col items-center justify-center px-4 py-20 sm:px-8">
        <Skeleton className="h-8 w-48 rounded-xl" />
        <Skeleton className="mt-2 h-10 w-72 rounded-xl" />
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
          <span className="block">Welcome back,</span>
          <span className="text-primary mx-auto mt-1.5 block max-w-2xl truncate">
            {user?.displayName || "Player"}
          </span>
        </h1>
        <p className="text-muted-foreground mx-auto mt-3 max-w-xl text-sm sm:text-base">
          Select what profile mode you want to use to interact with Eventory.
        </p>
      </div>

      {/* 3 Interactive Clickable Persona Cards */}
      <div className="grid gap-6 md:grid-cols-3">
        {personas.map(
          ({
            id,
            title,
            description,
            targetHref,
            actionText,
            icon: Icon,
            profileName,
          }) => {
            return (
              <Card
                key={id}
                role="button"
                tabIndex={0}
                onClick={() => handleSelectPersona(targetHref)}
                onKeyDown={(e) => {
                  if (e.key === "Enter" || e.key === " ") {
                    e.preventDefault();
                    handleSelectPersona(targetHref);
                  }
                }}
                className={cn(
                  "group border-border/70 bg-card relative flex cursor-pointer flex-col justify-between rounded-2xl border p-6 transition-all duration-200 select-none",
                  "hover:border-primary/60 hover:bg-muted/20 hover:-translate-y-1.5 hover:shadow-lg active:translate-y-0 active:scale-[0.99]"
                )}
              >
                <div>
                  <div className="flex items-center justify-between gap-2">
                    <div className="bg-muted text-foreground group-hover:bg-primary group-hover:text-primary-foreground flex size-12 items-center justify-center rounded-xl transition-colors duration-200">
                      <Icon className="size-6 transition-transform duration-200 group-hover:scale-110" />
                    </div>
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
                  <span>{actionText}</span>
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
