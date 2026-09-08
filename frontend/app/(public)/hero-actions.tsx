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
  // Natural gaming-oriented wording (replaces stiff "Open Workspace")
  const primaryText = isAuthenticated ? "Enter Eventory" : "Get Started";

  return (
    <div className="flex w-full flex-col items-center justify-center gap-3.5 pt-3 sm:w-auto sm:flex-row">
      <Button
        size="lg"
        render={<Link href={primaryHref} />}
        nativeButton={false}
        className="h-11 w-full rounded-lg px-5.5 has-data-[icon=inline-end]:pr-5 has-data-[icon=inline-start]:pl-5 text-sm font-semibold shadow-xs transition-all hover:opacity-95 sm:w-auto sm:min-w-47.5 justify-center"
      >
        <span className="leading-none">{primaryText}</span>
        <ArrowRight data-icon="inline-end" />
      </Button>

      <Button
        variant="outline"
        size="lg"
        render={<Link href="/tournaments" />}
        nativeButton={false}
        className="h-11 w-full rounded-lg px-5.5 has-data-[icon=inline-end]:pr-5 has-data-[icon=inline-start]:pl-5 text-sm font-semibold sm:w-auto sm:min-w-47.5 justify-center"
      >
        <Gamepad2 data-icon="inline-start" />
        <span className="leading-none">Explore Tournaments</span>
      </Button>
    </div>
  );
}
