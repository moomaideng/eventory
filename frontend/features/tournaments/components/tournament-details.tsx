"use client";

import React from "react";
import Link from "next/link";
import { ArrowLeft, Users } from "lucide-react";
import { $api } from "@/lib/api/client";
import { useAuth } from "@/context/auth-context";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { formatRegistrationType } from "../utils";
import { TournamentAboutCard } from "./tournament-about-card";
import { TournamentDetailsSkeleton } from "./tournament-details-skeleton";
import { TournamentFundingCard } from "./tournament-funding-card";
import { TournamentRegisteredTeamsCard } from "./tournament-registered-teams-card";

export function TournamentDetails({ tournamentId }: { tournamentId: string }) {
  const { authorizationHeader } = useAuth();
  const { data, error, isLoading } = $api.useQuery(
    "get",
    "/api/v1/tournaments/{tournamentId}",
    { params: { path: { tournamentId } } },
    { staleTime: 30_000 }
  );
  const { data: myTeam, isLoading: isMyTeamLoading } = $api.useQuery(
    "get",
    "/api/v1/tournaments/{tournamentId}/my-team",
    {
      params: { path: { tournamentId } },
      headers: authorizationHeader
        ? { Authorization: authorizationHeader }
        : {},
    },
    {
      enabled: Boolean(authorizationHeader),
      retry: false,
      staleTime: 0,
    }
  );

  if (isLoading) return <TournamentDetailsSkeleton />;

  if (error || !data) {
    return (
      <div className="container mx-auto flex w-full max-w-3xl flex-1 flex-col gap-6 px-4 py-12 sm:px-8">
        <Alert variant="destructive">
          <AlertTitle>Tournament unavailable</AlertTitle>
          <AlertDescription>
            This tournament may not exist, may be unpublished, or the API may be
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

  const { tournament, funding } = data;
  const teams = data.teams ?? [];
  const spotsLeft = Math.max(
    tournament.capacity - tournament.registeredCount,
    0
  );
  const status = tournament.status.replaceAll("_", " ").toLowerCase();
  const registrationType = formatRegistrationType(tournament.registrationMode);
  const supportsTeams =
    tournament.registrationMode === "TEAM" ||
    tournament.registrationMode === "BOTH";
  const canCreateTeam =
    supportsTeams &&
    tournament.status === "REGISTRATION_OPEN" &&
    !isMyTeamLoading &&
    !myTeam;

  return (
    <main className="container mx-auto flex w-full max-w-7xl flex-1 flex-col gap-8 px-4 py-10 sm:px-8">
      <Button
        render={<Link href="/tournaments" />}
        nativeButton={false}
        variant="ghost"
        className="self-start"
      >
        <ArrowLeft data-icon="inline-start" />
        All tournaments
      </Button>

      <header className="flex flex-col gap-4">
        <div className="flex flex-wrap gap-2">
          <Badge>{tournament.game}</Badge>
          <Badge variant="outline" className="capitalize">
            {status}
          </Badge>
          <Badge variant="outline">{registrationType}</Badge>
        </div>
        <div className="flex flex-col gap-2">
          <h1 className="max-w-4xl text-3xl font-bold tracking-tight sm:text-5xl">
            {tournament.name}
          </h1>
          <p className="text-muted-foreground text-base">
            Hosted by {tournament.organizerName}
          </p>
        </div>
        {canCreateTeam ? (
          <Button
            render={<Link href={`/tournaments/${tournament.id}/team`} />}
            nativeButton={false}
            className="self-start"
          >
            <Users data-icon="inline-start" />
            Create a team
          </Button>
        ) : null}
      </header>

      <div className="grid items-start gap-6 lg:grid-cols-[minmax(0,1fr)_22rem]">
        <div className="flex flex-col gap-6">
          <TournamentAboutCard
            tournament={tournament}
            spotsLeft={spotsLeft}
            registrationType={registrationType}
          />
          <TournamentRegisteredTeamsCard teams={teams} />
        </div>

        <div className="flex flex-col gap-6 lg:sticky lg:top-24">
          {myTeam ? (
            <Card>
              <CardHeader>
                <CardTitle>Your team</CardTitle>
                <CardDescription>
                  Your registration team for this tournament.
                </CardDescription>
              </CardHeader>
              <CardContent className="flex flex-col gap-3">
                <div className="flex items-center justify-between gap-3">
                  <p className="font-medium">{myTeam.name}</p>
                  <Badge variant="outline" className="capitalize">
                    {myTeam.status.toLowerCase()}
                  </Badge>
                </div>
                <p className="text-muted-foreground text-sm">
                  {myTeam.members?.length ?? 0} of {tournament.maxTeamSize} players
                </p>
              </CardContent>
              <CardFooter>
                <Button
                  render={<Link href={`/lobbies/${myTeam.inviteCode}`} />}
                  nativeButton={false}
                  className="w-full"
                >
                  <Users data-icon="inline-start" />
                  Open team lobby
                </Button>
              </CardFooter>
            </Card>
          ) : null}

          <TournamentFundingCard funding={funding} />
        </div>
      </div>
    </main>
  );
}
