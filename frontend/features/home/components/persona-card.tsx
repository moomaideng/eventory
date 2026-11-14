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
}

export function PersonaCard({ persona }: { persona: PersonaConfig }) {
  const { title, description, targetHref, actionText, icon: Icon } = persona;

  return (
    <Link href={targetHref} className="group block focus-visible:outline-none">
      <Card className="flex h-full min-h-60 flex-col justify-between transition duration-200 group-hover:border-primary/60 group-hover:bg-muted/20 group-hover:-translate-y-1 group-hover:shadow-md group-focus-visible:border-ring group-focus-visible:ring-[3px] group-focus-visible:ring-ring/50">
        <CardHeader className="gap-2">
          <div className="bg-muted text-foreground group-hover:bg-primary group-hover:text-primary-foreground mx-auto mb-3 flex size-12 items-center justify-center rounded-xl transition-colors duration-200">
            <Icon />
          </div>
          <CardTitle className="group-hover:text-primary text-xl font-bold tracking-tight transition-colors">
            {title}
          </CardTitle>
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

export function PersonaGrid() {
  const personas: PersonaConfig[] = [
    {
      id: "competitor",
      title: "Compete in a tournament",
      description:
        "Browse open tournaments, join team lobbies, and compete in brackets.",
      targetHref: "/tournaments",
      actionText: "Browse tournaments",
      icon: Gamepad2,
    },
    {
      id: "organizer",
      title: "Organize a tournament",
      description:
        "Create events, set crowdfunding prize goals, manage match schedules, and invite staff.",
      targetHref: "/organizer",
      actionText: "Organizer workspace",
      icon: Trophy,
    },
    {
      id: "sponsor",
      title: "Sponsor a tournament",
      description:
        "Back tournament prize pools, track sponsorships, and showcase your brand logo.",
      targetHref: "/sponsor",
      actionText: "Sponsor workspace",
      icon: Briefcase,
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
