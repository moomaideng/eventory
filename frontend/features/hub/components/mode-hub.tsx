"use client";

import React from "react";
import Link from "next/link";
import { useAuth } from "@/context/auth-context";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Empty,
  EmptyContent,
  EmptyDescription,
  EmptyHeader,
  EmptyTitle,
} from "@/components/ui/empty";
import { Gamepad2, Trophy, Briefcase } from "lucide-react";
import { PersonaCard, type PersonaConfig } from "./persona-card";

export function ModeHub() {
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
      <div className="container mx-auto flex max-w-md flex-1 items-center justify-center px-4 py-24">
        <Empty className="border">
          <EmptyHeader>
            <EmptyTitle>Sign In Required</EmptyTitle>
            <EmptyDescription>
              Please sign in to select your workspace mode and access your profile.
            </EmptyDescription>
          </EmptyHeader>
          <EmptyContent>
            <Button
              render={<Link href="/login" />}
              nativeButton={false}
            >
              Sign In
            </Button>
          </EmptyContent>
        </Empty>
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
            {user.displayName || "Player"}
          </span>
        </h1>
        <p className="text-muted-foreground mx-auto mt-3 max-w-xl text-sm sm:text-base">
          Select what profile mode you want to use to interact with Eventory.
        </p>
      </div>

      {/* 3 Interactive Clickable Persona Cards */}
      <div className="grid gap-6 md:grid-cols-3">
        {personas.map((persona) => (
          <PersonaCard key={persona.id} persona={persona} />
        ))}
      </div>
    </div>
  );
}
