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
import { PublicTiersSection } from "@/features/sponsorship";
import { formatRegistrationType, formatTournamentStatus } from "../utils";
import { TournamentAboutCard } from "./tournament-about-card";
import { TournamentDetailsSkeleton } from "./tournament-details-skeleton";
import { TournamentFundingCard } from "./tournament-funding-card";
import { TournamentRegisteredTeamsCard } from "./tournament-registered-teams-card";

export function TournamentDetails({
  tournamentId,
  sponsorMode = false,
}: {
  tournamentId: string;
  sponsorMode?: boolean;
}) {
  const catalogPath = sponsorMode ? "/sponsor/tournaments" : "/tournaments";
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
      enabled: Boolean(authorizationHeader) && !sponsorMode,
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
          render={<Link href={catalogPath} />}
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
  const status = formatTournamentStatus(tournament.status);
  const registrationType = formatRegistrationType(tournament.registrationMode);
  const isSolo = tournament.registrationMode === "SOLO";
  const isCrowdfunding = tournament.status === "CROWDFUNDING";
  const canCreateTeam =
    !isSolo &&
    tournament.status === "REGISTRATION_OPEN" &&
    !isMyTeamLoading &&
    !myTeam;

  return (
    <main className="container mx-auto flex w-full max-w-7xl flex-1 flex-col gap-8 px-4 py-10 sm:px-8">
      <Button
        render={<Link href={catalogPath} />}
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
          {!isCrowdfunding ? (
            <Badge variant="outline">{registrationType}</Badge>
          ) : null}
        </div>
        <div className="flex flex-col gap-2">
          <h1 className="max-w-4xl text-3xl font-bold tracking-tight sm:text-5xl">
            {tournament.name}
          </h1>
          <p className="text-muted-foreground text-base">
            Hosted by {tournament.organizerName}
          </p>
        </div>
        {!sponsorMode && canCreateTeam ? (
          <Button
            render={<Link href={`/tournaments/${tournament.id}/team`} />}
            nativeButton={false}
            className="self-start"
          >
            <Users data-icon="inline-start" />
            Create a team
          </Button>
        ) : null}
        {isCrowdfunding ? (
          <Alert>
            <AlertTitle>Funding stage</AlertTitle>
            <AlertDescription>
              Sponsorship is open. Participant registration will open after this
              stage.
            </AlertDescription>
          </Alert>
        ) : null}
        {tournament.status === "CANCELLED" ? (
          <Alert>
            <AlertTitle>Tournament cancelled</AlertTitle>
            <AlertDescription>
              This event will not take place. Participant registration is
              closed.
            </AlertDescription>
          </Alert>
        ) : null}
        {!sponsorMode && isSolo && !isCrowdfunding ? (
          tournament.status !== "REGISTRATION_OPEN" ? (
            <Button disabled className="self-start">
              Registration closed
            </Button>
          ) : !authorizationHeader ? (
            <Button
              render={
                <Link href={`/login?redirect=/tournaments/${tournament.id}`} />
              }
              nativeButton={false}
              className="self-start"
            >
              Sign in to register
            </Button>
          ) : isMyTeamLoading ? (
            <Button disabled className="self-start">
              Checking eligibility
            </Button>
          ) : myTeam ? (
            <div className="flex max-w-xl flex-col gap-3">
              {myTeam.status === "LOCKED" ? (
                <>
                  <Alert>
                    <AlertTitle>Registration submitted</AlertTitle>
                    <AlertDescription>
                      Your registration has been submitted and is currently awaiting organizer review.
                    </AlertDescription>
                  </Alert>
                  <Button disabled className="self-start">
                    Awaiting review
                  </Button>
                </>
              ) : myTeam.status === "PAYMENT_PENDING" ? (
                <>
                  <Alert>
                    <AlertTitle>Payment pending</AlertTitle>
                    <AlertDescription>
                      Please complete your entry fee payment to finalize your registration.
                    </AlertDescription>
                  </Alert>
                  {myTeam.paymentId ? (
                    <Button
                      render={<Link href={`/payments/${myTeam.paymentId}/checkout`} />}
                      nativeButton={false}
                      className="self-start"
                    >
                      Complete payment
                    </Button>
                  ) : (
                    <Button disabled className="self-start">
                      Payment pending
                    </Button>
                  )}
                </>
              ) : myTeam.status === "ACCEPTED" ? (
                <>
                  <Alert className="border-emerald-500/40 text-emerald-800 dark:text-emerald-200">
                    <AlertTitle>Registration confirmed</AlertTitle>
                    <AlertDescription>
                      You are confirmed to participate in this tournament!
                    </AlertDescription>
                  </Alert>
                  <Button disabled className="self-start">
                    Confirmed
                  </Button>
                </>
              ) : myTeam.status === "REJECTED" ? (
                <>
                  <Alert variant="destructive">
                    <AlertTitle>Application rejected</AlertTitle>
                    <AlertDescription>
                      Your application was not accepted for this tournament.
                    </AlertDescription>
                  </Alert>
                  <Button disabled className="self-start">
                    Application rejected
                  </Button>
                </>
              ) : (
                <Button disabled className="self-start">
                  Registered
                </Button>
              )}
            </div>
          ) : spotsLeft <= 0 ? (
            <Button disabled className="self-start">
              Tournament full
            </Button>
          ) : (
            <Button
              render={<Link href={`/tournaments/${tournament.id}/form`} />}
              nativeButton={false}
              className="self-start"
            >
              Register as an individual
            </Button>
          )
        ) : null}
      </header>

      <div className="grid items-start gap-6 lg:grid-cols-[minmax(0,1fr)_22rem]">
        <div className="flex flex-col gap-6">
          {sponsorMode && (
            <PublicTiersSection
              tournamentId={tournament.id}
              currency={tournament.currency}
              status={tournament.status}
              endsAt={tournament.endAt}
            />
          )}
          <TournamentAboutCard
            tournament={tournament}
            spotsLeft={spotsLeft}
            registrationType={registrationType}
          />
          {!sponsorMode && (
            <PublicTiersSection
              tournamentId={tournament.id}
              currency={tournament.currency}
              status={tournament.status}
              endsAt={tournament.endAt}
            />
          )}
          {!isCrowdfunding ? (
            <TournamentRegisteredTeamsCard teams={teams} />
          ) : null}
        </div>

        <div className="flex flex-col gap-6 lg:sticky lg:top-24">
          {!sponsorMode && myTeam ? (
            <Card>
              <CardHeader>
                <CardTitle>
                  {isSolo ? "Your registration" : "Your team"}
                </CardTitle>
                <CardDescription>
                  {isSolo
                    ? "Your registration for this tournament."
                    : "Your registration team for this tournament."}
                </CardDescription>
              </CardHeader>
              <CardContent className="flex flex-col gap-3">
                <div className="flex items-center justify-between gap-3">
                  <p className="font-medium">{myTeam.name}</p>
                  <Badge variant="outline" className="capitalize">
                    {myTeam.status.toLowerCase().replace("_", " ")}
                  </Badge>
                </div>
                {!isSolo && (
                  <p className="text-muted-foreground text-sm">
                    {myTeam.members?.length ?? 0} of {tournament.maxTeamSize}{" "}
                    players
                  </p>
                )}
              </CardContent>
              {!isSolo && myTeam.inviteCode ? (
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
              ) : null}
            </Card>
          ) : null}

          {funding.goalAmount > 0 ? (
            <TournamentFundingCard
              funding={funding}
              status={tournament.status}
            />
          ) : null}
        </div>
      </div>
    </main>
  );
}
