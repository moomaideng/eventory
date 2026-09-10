"use client";

import React from "react";
import Link from "next/link";
import { useAuth } from "@/context/auth-context";
import { Button } from "@/components/ui/button";
import { ArrowRight, Gamepad2 } from "lucide-react";

export function HeroActions({
  serverAuthenticated = false,
}: {
  serverAuthenticated?: boolean;
}) {
  const { user } = useAuth();
  const isAuthenticated = Boolean(user) || serverAuthenticated;

  // Smart routing: unauthenticated goes to /login, authenticated jumps straight into /hub
  const primaryHref = isAuthenticated ? "/hub" : "/login";
  const primaryText = isAuthenticated ? "Enter Eventory" : "Get Started";

  return (
    <div className="flex w-full flex-col items-center justify-center gap-3 pt-2 sm:w-auto sm:flex-row">
      <Button
        size="lg"
        render={<Link href={primaryHref} />}
        nativeButton={false}
        className="w-full sm:w-auto sm:min-w-48"
      >
        <span>{primaryText}</span>
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
