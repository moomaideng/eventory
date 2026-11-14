"use client";

import React from "react";
import Link from "next/link";
import { ArrowRight, Briefcase, Gamepad2, Trophy } from "lucide-react";
import {
  Card,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import type { UserRole } from "@/lib/role";

export interface PersonaConfig {
  id: UserRole;
  title: string;
  description: string;
  targetHref: string;
  actionText: string;
  icon: React.ElementType;
  profileName?: string;
}

export function PersonaCard({ persona }: { persona: PersonaConfig }) {
  const { title, description, targetHref, actionText, icon: Icon, profileName } =
    persona;

  return (
    <Link href={targetHref} className="group block focus-visible:outline-none">
      <Card className="flex h-full min-h-72 flex-col justify-between transition duration-200 group-hover:border-primary/60 group-hover:bg-muted/20 group-hover:-translate-y-1 group-hover:shadow-md">
        <CardHeader className="gap-2">
          <div className="bg-muted text-foreground group-hover:bg-primary group-hover:text-primary-foreground mb-3 flex size-12 items-center justify-center rounded-xl transition-colors duration-200">
            <Icon />
          </div>
          <CardTitle className="group-hover:text-primary text-xl font-bold tracking-tight transition-colors">
            {title}
          </CardTitle>
          {profileName ? (
            <p className="text-muted-foreground truncate text-xs font-medium">
              {profileName}
            </p>
          ) : null}
          <CardDescription className="text-sm leading-relaxed">
            {description}
          </CardDescription>
        </CardHeader>
        <CardFooter className="pt-4">
          <div className="text-muted-foreground group-hover:text-primary flex items-center gap-1.5 text-xs font-semibold transition-colors duration-200">
            <span>{actionText}</span>
            <ArrowRight data-icon="inline-end" />
          </div>
        </CardFooter>
      </Card>
    </Link>
  );
}

export function PersonaGrid({ handle }: { handle?: string }) {
  const personas: PersonaConfig[] = [
    {
      id: "competitor",
      title: "Competitor",
      description:
        "Browse open tournaments, join team lobbies, and compete in brackets.",
      targetHref: "/tournaments",
      actionText: "Browse Tournaments",
      icon: Gamepad2,
      profileName: handle ? `@${handle}` : undefined,
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

  return (
    <div className="mx-auto grid w-full max-w-4xl gap-6 text-left md:grid-cols-3">
      {personas.map((persona) => (
        <PersonaCard key={persona.id} persona={persona} />
      ))}
    </div>
  );
}
