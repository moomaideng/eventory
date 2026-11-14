"use client";

import React from "react";
import Link from "next/link";
import { useAuth } from "@/context/auth-context";
import { Button } from "@/components/ui/button";
import { PersonaGrid } from "@/features/home/components/persona-card";
import { ArrowRight, Gamepad2 } from "lucide-react";

export function HeroActions({
  serverAuthenticated = false,
}: {
  serverAuthenticated?: boolean;
}) {
  const { user, isLoading } = useAuth();
  const isAuthenticated = Boolean(user) || (isLoading && serverAuthenticated);

  if (isAuthenticated) {
    return (
      <section
        aria-labelledby="landing-action-heading"
        className="flex w-full max-w-4xl flex-col items-center gap-5"
      >
        <h2
          id="landing-action-heading"
          className="text-foreground text-center text-xl font-semibold tracking-tight sm:text-2xl"
        >
          What do you want to do?
        </h2>
        <PersonaGrid />
      </section>
    );
  }

  return (
    <div className="flex w-full flex-col items-center justify-center gap-3 pt-2 sm:w-auto sm:flex-row">
      <Button
        size="lg"
        render={<Link href="/login" />}
        nativeButton={false}
        className="w-full sm:w-auto sm:min-w-48"
      >
        <span>Get Started</span>
        <ArrowRight data-icon="inline-end" />
      </Button>

      <Button
        variant="outline"
        size="lg"
        render={<Link href="/tournaments" />}
        nativeButton={false}
        className="w-full sm:w-auto sm:min-w-48"
      >
        <Gamepad2 data-icon="inline-start" />
        <span>Explore Tournaments</span>
      </Button>
    </div>
  );
}
