"use client";

import React from "react";
import { useAuth } from "@/context/auth-context";
import { Skeleton } from "@/components/ui/skeleton";
import { PersonaGrid } from "./persona-card";

export function ModeHub() {
  const { user, isLoading } = useAuth();

  if (isLoading || !user) {
    return (
      <div className="container mx-auto flex w-full max-w-7xl flex-1 flex-col items-center justify-center px-4 py-16 sm:px-8">
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

  return (
    <div className="container mx-auto flex w-full max-w-7xl flex-1 flex-col justify-center px-4 py-16 sm:px-8">
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
      <PersonaGrid handle={user.handle} />
    </div>
  );
}
