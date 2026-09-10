"use client";

import React from "react";
import Link from "next/link";
import { ArrowRight } from "lucide-react";
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
      <Card className="flex h-full flex-col justify-between transition-all duration-200 group-hover:border-primary/60 group-hover:bg-muted/20 group-hover:-translate-y-1 group-hover:shadow-md">
        <CardHeader>
          <div className="bg-muted text-foreground group-hover:bg-primary group-hover:text-primary-foreground mb-2 flex size-12 items-center justify-center rounded-xl transition-colors duration-200">
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
        <CardFooter className="pt-2">
          <div className="text-muted-foreground group-hover:text-primary flex items-center gap-1.5 text-xs font-semibold transition-[color,transform] duration-200 group-hover:translate-x-1">
            <span>{actionText}</span>
            <ArrowRight data-icon="inline-end" />
          </div>
        </CardFooter>
      </Card>
    </Link>
  );
}
