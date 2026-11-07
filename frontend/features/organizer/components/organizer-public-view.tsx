"use client";

import React from "react";
import Link from "next/link";
import {
  ArrowLeft,
  Building2,
  CalendarDays,
  Mail,
  Trophy,
} from "lucide-react";
import { $api } from "@/lib/api/client";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardDescription, CardTitle } from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@/components/ui/empty";
import { TournamentCard } from "@/features/tournaments/components/tournament-card";
import { TournamentGridSkeleton } from "@/features/tournaments/components/tournament-grid-skeleton";
import { formatTournamentDate } from "@/features/tournaments/utils";
import { OrganizerDetailsSkeleton } from "./organizer-details-skeleton";

export function OrganizerPublicView({ organizerId }: { organizerId: string }) {
  const {
    data: profile,
    error: profileError,
    isLoading: isProfileLoading,
  } = $api.useQuery(
    "get",
    "/api/v1/accounts/{id}/organizer-profile",
    { params: { path: { id: organizerId } } },
    { staleTime: 60_000 }
  );

  const {
    data: tournamentsData,
    error: tournamentsError,
    isLoading: isTournamentsLoading,
  } = $api.useQuery(
    "get",
    "/api/v1/tournaments",
    {
      params: {
        query: {
          organizerId,
          pageSize: 50,
        },
      },
    },
    { staleTime: 30_000 }
  );

  if (isProfileLoading) {
    return <OrganizerDetailsSkeleton />;
  }

  if (profileError || !profile) {
    return (
      <div className="container mx-auto flex w-full max-w-3xl flex-1 flex-col gap-6 px-4 py-12 sm:px-8">
        <Alert variant="destructive">
          <AlertTitle>Organizer unavailable</AlertTitle>
          <AlertDescription>
            This organizer profile could not be found, or the profile service is
            temporarily unavailable.
          </AlertDescription>
        </Alert>
        <Button
          render={<Link href="/tournaments" />}
          nativeButton={false}
          variant="outline"
          className="self-start"
        >
          <ArrowLeft data-icon="inline-start" />
          Back to tournaments
        </Button>
      </div>
    );
  }

  const tournaments = tournamentsData?.items ?? [];
  const tournamentCount = tournamentsData?.total ?? tournaments.length;

  return (
    <div className="container mx-auto flex w-full max-w-7xl flex-1 flex-col gap-8 px-4 py-10 sm:px-8">
      <Button
        render={<Link href="/tournaments" />}
        nativeButton={false}
        variant="ghost"
        className="self-start"
      >
        <ArrowLeft data-icon="inline-start" />
        All tournaments
      </Button>

      {/* Organizer Profile Header Card */}
      <Card className="overflow-hidden border py-0">
        <div className="grid grid-cols-1 items-stretch gap-6 p-6 sm:p-8 md:grid-cols-[1fr_auto_1fr] md:gap-8">
          {/* Column 1: Identity */}
          <div className="flex min-w-0 items-center gap-4">
            <div className="bg-primary/10 text-primary flex size-16 shrink-0 items-center justify-center rounded-xl font-bold">
              <Building2 className="size-8" aria-hidden="true" />
            </div>
            <div className="flex min-w-0 flex-col gap-1">
              <div className="flex flex-wrap items-center gap-2">
                <CardTitle className="text-2xl font-bold tracking-tight sm:text-3xl">
                  {profile.organizerName}
                </CardTitle>
                <Badge variant="secondary">Organizer</Badge>
              </div>
              <CardDescription>
                Public organizer profile and tournament hosting record
              </CardDescription>
            </div>
          </div>

          {/* Divider */}
          <Separator orientation="vertical" className="hidden self-stretch md:block" />
          <Separator orientation="horizontal" className="md:hidden" />

          {/* Column 2: Details */}
          <div className="flex flex-col justify-center gap-3 text-sm">
            <div className="text-muted-foreground flex items-center gap-2.5">
              <Mail className="size-4 shrink-0" />
              <span>
                {profile.organizerEmail || (
                  <span className="italic">No public email provided</span>
                )}
              </span>
            </div>
            {profile.createdAt ? (
              <div className="text-muted-foreground flex items-center gap-2.5">
                <CalendarDays className="size-4 shrink-0" />
                <span>Host since {formatTournamentDate(profile.createdAt)}</span>
              </div>
            ) : null}
            <div className="text-muted-foreground flex items-center gap-2.5">
              <Trophy className="size-4 shrink-0" />
              <span>
                {tournamentCount}{" "}
                {tournamentCount === 1 ? "tournament" : "tournaments"} hosted
              </span>
            </div>
          </div>
        </div>
      </Card>

      {/* Hosted Tournaments Section */}
      <section className="flex flex-col gap-6">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <h2 className="text-xl font-bold tracking-tight sm:text-2xl">
              Hosted Tournaments
            </h2>
            <Badge variant="outline">{tournamentCount}</Badge>
          </div>
        </div>

        {isTournamentsLoading ? (
          <TournamentGridSkeleton />
        ) : tournamentsError ? (
          <Alert variant="destructive">
            <AlertTitle>Failed to load hosted tournaments</AlertTitle>
            <AlertDescription>
              We could not load this organizer&apos;s tournaments. Please try
              refreshing the page.
            </AlertDescription>
          </Alert>
        ) : tournaments.length === 0 ? (
          <Empty className="border">
            <EmptyHeader>
              <EmptyMedia variant="icon">
                <Trophy />
              </EmptyMedia>
              <EmptyTitle>No hosted tournaments</EmptyTitle>
              <EmptyDescription>
                This organizer has not hosted any published tournaments yet.
              </EmptyDescription>
            </EmptyHeader>
          </Empty>
        ) : (
          <div className="grid gap-5 md:grid-cols-2 xl:grid-cols-3">
            {tournaments.map((tournament) => (
              <TournamentCard key={tournament.id} tournament={tournament} />
            ))}
          </div>
        )}
      </section>
    </div>
  );
}
