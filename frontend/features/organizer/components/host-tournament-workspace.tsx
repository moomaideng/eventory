"use client";

import React, { useState } from "react";
import Link from "next/link";
import {
  ArrowLeft,
  ArrowRight,
  ShieldAlert,
  Trophy,
} from "lucide-react";
import { $api } from "@/lib/api/client";
import { useAuth } from "@/context/auth-context";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { HostTournamentForm } from "./host-tournament-form";
import { HostTournamentPreview } from "./host-tournament-preview";
import type { CreateTournamentFormInput } from "../schemas";

export function HostTournamentWorkspace() {
  const { authorizationHeader, isLoading: authLoading } = useAuth();

  const {
    data: profile,
    error: profileError,
    isLoading: profileLoading,
  } = $api.useQuery(
    "get",
    "/api/v1/accounts/me/organizer-profile",
    {
      headers: authorizationHeader
        ? { Authorization: authorizationHeader }
        : {},
    },
    {
      enabled: Boolean(authorizationHeader),
      retry: false,
      staleTime: 60_000,
    }
  );

  const [draft, setDraft] = useState<CreateTournamentFormInput>({
    name: "",
    game: "",
    location: "Online",
    description: "",
    registrationMode: "TEAM",
    minTeamSize: 5,
    maxTeamSize: 5,
    capacity: 16,
    entryFee: 0,
    registrationDeadline: "",
    startAt: "",
    endAt: "",
  });

  const loading = authLoading || profileLoading;

  if (loading) {
    return (
      <div className="mx-auto flex w-full max-w-7xl flex-1 flex-col gap-8 px-4 py-10 sm:px-8">
        <Skeleton className="h-9 w-36 rounded-md" />
        <div className="flex flex-col gap-2">
          <Skeleton className="h-8 w-64" />
          <Skeleton className="h-4 w-96" />
        </div>
        <div className="grid grid-cols-1 gap-8 lg:grid-cols-12">
          <div className="lg:col-span-7">
            <Skeleton className="h-[520px] w-full rounded-xl" />
          </div>
          <div className="lg:col-span-5">
            <Skeleton className="h-[360px] w-full rounded-xl" />
          </div>
        </div>
      </div>
    );
  }

  // If the user has not configured their organizer profile yet
  if (profileError || !profile) {
    return (
      <div className="mx-auto flex w-full max-w-3xl flex-1 flex-col gap-8 px-4 py-12 sm:px-8">
        <Button
          variant="ghost"
          size="sm"
          render={<Link href="/organizer" />}
          nativeButton={false}
          className="self-start"
        >
          <ArrowLeft data-icon="inline-start" />
          Back to Tournaments
        </Button>

        <Card className="border-border">
          <CardHeader className="text-center sm:text-left">
            <div className="bg-primary/10 text-primary mb-3 flex size-12 items-center justify-center rounded-lg">
              <ShieldAlert aria-hidden="true" />
            </div>
            <CardTitle className="text-2xl">Organizer Profile Required</CardTitle>
            <CardDescription className="text-sm">
              Before hosting a tournament on Eventory, you must configure your
              organizer brand or organization name and contact email.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <p className="text-muted-foreground text-sm leading-relaxed">
              Your organizer identity will be displayed publicly on every
              tournament you create, allowing competitors and potential sponsors
              to identify and reach your staff.
            </p>
          </CardContent>
          <CardFooter className="flex flex-col gap-3 sm:flex-row sm:justify-end">
            <Button
              variant="outline"
              render={<Link href="/organizer" />}
              nativeButton={false}
            >
              Back to dashboard
            </Button>
            <Button
              render={<Link href="/organizer/profile" />}
              nativeButton={false}
            >
              Set up organizer profile
              <ArrowRight data-icon="inline-end" />
            </Button>
          </CardFooter>
        </Card>
      </div>
    );
  }

  return (
    <div className="mx-auto flex w-full max-w-7xl flex-1 flex-col gap-8 px-4 py-10 sm:px-8">
      {/* Top Header & Breadcrumb */}
      <div className="flex flex-col gap-4">
        <Button
          variant="ghost"
          size="sm"
          render={<Link href="/organizer" />}
          nativeButton={false}
          className="self-start text-muted-foreground hover:text-foreground"
        >
          <ArrowLeft data-icon="inline-start" />
          My Tournaments
        </Button>

        <div className="flex items-center gap-4">
          <div className="bg-primary/10 text-primary flex size-12 shrink-0 items-center justify-center rounded-lg">
            <Trophy aria-hidden="true" />
          </div>
          <div className="flex flex-col">
            <p className="text-muted-foreground text-xs font-semibold tracking-wider uppercase">
              Organizer Workspace
            </p>
            <h1 className="text-2xl font-bold tracking-tight sm:text-3xl">
              Host a Tournament
            </h1>
          </div>
        </div>
      </div>

      {/* Main Form & Preview Workspace */}
      <div className="grid grid-cols-1 items-start gap-8 lg:grid-cols-12">
        {/* Form Column */}
        <div className="lg:col-span-7 xl:col-span-8">
          <HostTournamentForm
            organizerName={profile.organizerName}
            onValuesChange={setDraft}
          />
        </div>

        {/* Sticky Preview Column */}
        <div className="lg:sticky lg:top-20 lg:col-span-5 xl:col-span-4">
          <HostTournamentPreview
            name={draft.name}
            game={draft.game}
            location={draft.location}
            description={draft.description}
            registrationMode={draft.registrationMode}
            minTeamSize={draft.minTeamSize}
            maxTeamSize={draft.maxTeamSize}
            capacity={draft.capacity}
            entryFee={draft.entryFee}
            registrationDeadline={draft.registrationDeadline}
            startAt={draft.startAt}
            organizerName={profile.organizerName}
          />
        </div>
      </div>
    </div>
  );
}
