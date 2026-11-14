"use client";

import React from "react";
import Link from "next/link";
import { useQuery } from "@tanstack/react-query";
import { ArrowUpRight, GitBranch, Users } from "lucide-react";
import { AnnouncementPreview } from "@/features/announcements";
import { $api, apiClient } from "@/lib/api/client";
import { useAuth } from "@/context/auth-context";
import { PageHeader } from "@/components/page-header";
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
import { PublicSponsorsSection, PublicTiersSection } from "@/features/sponsorship";
import { RegistrationPaymentButton } from "@/features/payments/components/registration-payment-button";
import {
  formatRegistrationType,
  isRegistrationOpen,
} from "../utils";
import { TournamentAboutCard } from "./tournament-about-card";
import { TournamentDetailsSkeleton } from "./tournament-details-skeleton";
import { TournamentFundingCard } from "./tournament-funding-card";
import { TournamentHeaderTags } from "./tournament-header-tags";
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
  const {
    data: myTeam,
    isFetching: isMyTeamFetching,
    isError: isMyTeamError,
  } = useQuery({
    queryKey: [
      "get",
      "/api/v1/tournaments/{tournamentId}/my-team",
      tournamentId,
      authorizationHeader,
    ],
    enabled: Boolean(authorizationHeader) && !sponsorMode,
    retry: false,
    staleTime: 0,
    refetchOnWindowFocus: true,
    queryFn: async () => {
      const result = await apiClient.GET("/api/v1/tournaments/{tournamentId}/my-team", {
        params: { path: { tournamentId } },
        headers: { Authorization: authorizationHeader ?? "" },
      });
      if (result.response.status === 404) return null;
      if (result.error || !result.data) throw new Error("Failed to check registration status");
      return result.data;
    },
  });

  if (isLoading) return <TournamentDetailsSkeleton />;

  if (error || !data) {
    return (
      <div className="container mx-auto flex w-full max-w-7xl flex-1 flex-col gap-6 px-4 py-10 sm:px-8">
        <PageHeader
          workspace={sponsorMode ? "sponsor" : undefined}
          breadcrumbs={[
            {
              label: sponsorMode ? "Find tournaments" : "Tournaments",
              href: catalogPath,
            },
            { label: "Unavailable" },
          ]}
          title="Tournament unavailable"
        />
        <Alert variant="destructive">
          <AlertTitle>Tournament unavailable</AlertTitle>
          <AlertDescription>
            This tournament may not exist, may be unpublished, or the API may be
            temporarily unavailable.
          </AlertDescription>
        </Alert>
      </div>
    );
  }

  const { tournament, funding } = data;
  const teams = data.teams ?? [];
  const spotsLeft = Math.max(
    tournament.capacity - tournament.registeredCount,
    0
  );
  const registrationType = formatRegistrationType(tournament.registrationMode);
  const isSolo = tournament.registrationMode === "SOLO";
  const isCrowdfunding = tournament.status === "CROWDFUNDING";
  const hasPublicBracket = ["REGISTRATION", "ONGOING", "COMPLETED"].includes(
    tournament.status
  );

  const hasFundingOutcome = funding.goalAmountMinor > 0;

  if (sponsorMode && !isCrowdfunding && !hasFundingOutcome) {
    return (
      <main className="container mx-auto flex w-full max-w-7xl flex-1 flex-col gap-6 px-4 py-10 sm:px-8">
        <PageHeader
          workspace="sponsor"
          breadcrumbs={[
            { label: "Find tournaments", href: catalogPath },
            { label: tournament.name },
          ]}
          tags={
            <TournamentHeaderTags
              game={tournament.game}
              status={tournament.status}
              registrationMode={tournament.registrationMode}
              registrationDeadline={tournament.registrationDeadline}
            />
          }
          title={tournament.name}
          description={`Hosted by ${tournament.organizerName}`}
          actions={
            <Button
              render={<Link href={`/tournaments/${tournament.id}`} />}
              nativeButton={false}
              variant="outline"
            >
              View public page
              <ArrowUpRight data-icon="inline-end" />
            </Button>
          }
        />
        <Alert>
          <AlertTitle>Sponsorship unavailable</AlertTitle>
          <AlertDescription className="flex flex-col items-start gap-3 sm:flex-row sm:items-center sm:justify-between">
            <span>
              Only tournaments with crowdfunding are available here.
            </span>
            <Button
              render={<Link href="/sponsor/tournaments" />}
              nativeButton={false}
              variant="outline"
              size="sm"
            >
              Browse crowdfunding tournaments
            </Button>
          </AlertDescription>
        </Alert>
      </main>
    );
  }

  let registrationAction: React.ReactNode = null;
  if (!sponsorMode && !isCrowdfunding) {
    if (myTeam?.status === "REJECTED") {
      registrationAction = <Button disabled>Application rejected</Button>;
    } else if (myTeam?.status === "ACCEPTED") {
      registrationAction = <Button disabled>Confirmed</Button>;
    } else if (myTeam?.status === "LOCKED") {
      registrationAction = <Button disabled>Awaiting review</Button>;
    } else if (!isRegistrationOpen(tournament)) {
      registrationAction = <Button disabled>Registration closed</Button>;
    } else if (!authorizationHeader) {
      registrationAction = isSolo ? (
        <Button
          render={
            <Link href={`/login?redirect=/tournaments/${tournament.id}`} />
          }
          nativeButton={false}
        >
          Sign in to register
        </Button>
      ) : (
        <Button
          render={<Link href={`/tournaments/${tournament.id}/teams/new`} />}
          nativeButton={false}
        >
          <Users data-icon="inline-start" />
          Create a team
        </Button>
      );
    } else if (isMyTeamFetching) {
      registrationAction = <Button disabled>Checking eligibility</Button>;
    } else if (isMyTeamError) {
      registrationAction = <Button disabled>Registration unavailable</Button>;
    } else if (myTeam) {
      if (myTeam.status === "PAYMENT_PENDING") {
        registrationAction = isSolo ? (
          <RegistrationPaymentButton
            tournamentId={tournament.id}
            teamId={myTeam.id}
          />
        ) : (
          <Button disabled>Payment pending</Button>
        );
      } else if (myTeam.status === "FORMING") {
        registrationAction = isSolo ? (
          <RegistrationPaymentButton
            tournamentId={tournament.id}
            teamId={myTeam.id}
          >
            {tournament.entryFeeMinor > 0
              ? "Continue to payment"
              : "Submit registration"}
          </RegistrationPaymentButton>
        ) : (
          <Button disabled>Team forming</Button>
        );
      } else {
        registrationAction = <Button disabled>Registration started</Button>;
      }
    } else if (spotsLeft <= 0) {
      registrationAction = <Button disabled>Tournament full</Button>;
    } else if (isSolo) {
      registrationAction = (
        <Button
          render={<Link href={`/tournaments/${tournament.id}/form`} />}
          nativeButton={false}
        >
          Register as an individual
        </Button>
      );
    } else {
      registrationAction = (
        <Button
          render={<Link href={`/tournaments/${tournament.id}/teams/new`} />}
          nativeButton={false}
        >
          <Users data-icon="inline-start" />
          Create a team
        </Button>
      );
    }
  }

  return (
    <main className="container mx-auto flex w-full max-w-7xl flex-1 flex-col gap-8 px-4 py-10 sm:px-8">
      <PageHeader
        workspace={sponsorMode ? "sponsor" : undefined}
        breadcrumbs={[
          {
            label: sponsorMode ? "Find tournaments" : "Tournaments",
            href: catalogPath,
          },
          { label: tournament.name },
        ]}
        tags={
          <TournamentHeaderTags
            game={tournament.game}
            status={tournament.status}
            registrationMode={tournament.registrationMode}
            registrationDeadline={tournament.registrationDeadline}
          />
        }
        title={tournament.name}
        description={`Hosted by ${tournament.organizerName}`}
        actions={
          sponsorMode ? (
            <Button
              render={<Link href={`/tournaments/${tournament.id}`} />}
              nativeButton={false}
              variant="outline"
            >
              View public page
              <ArrowUpRight data-icon="inline-end" />
            </Button>
          ) : hasPublicBracket || registrationAction ? (
            <>
              {hasPublicBracket ? (
                <Button
                  render={<Link href={`/tournaments/${tournament.id}/bracket`} />}
                  nativeButton={false}
                  variant="outline"
                >
                  <GitBranch data-icon="inline-start" />
                  View bracket
                </Button>
              ) : null}
              {registrationAction}
            </>
          ) : undefined
        }
      />

      {!sponsorMode && isCrowdfunding ? (
        <Alert>
          <AlertTitle>Funding stage</AlertTitle>
          <AlertDescription>
            Sponsorship is open. Participant registration will open after this
            stage.
          </AlertDescription>
        </Alert>
      ) : null}

      <div
        className={
          !sponsorMode || funding.goalAmountMinor > 0
            ? "grid items-start gap-6 lg:grid-cols-[minmax(0,1fr)_22rem]"
            : "flex flex-col gap-6"
        }
      >
        <div className="flex flex-col gap-6">
          {sponsorMode ? (
            <PublicTiersSection
              sponsorMode
              tournamentId={tournament.id}
              currency={tournament.currency}
              status={tournament.status}
              endsAt={tournament.endAt}
            />
          ) : null}
          <TournamentAboutCard
            tournament={tournament}
            spotsLeft={spotsLeft}
            registrationType={registrationType}
          />
          {!sponsorMode && !isCrowdfunding ? (
            <TournamentRegisteredTeamsCard
              tournamentId={tournament.id}
              teams={teams}
              isSolo={isSolo}
            />
          ) : null}
          <PublicSponsorsSection
            tournamentId={tournament.id}
            hideWhenEmpty={!sponsorMode && !isCrowdfunding && !hasFundingOutcome}
          />
        </div>

        {!sponsorMode || funding.goalAmountMinor > 0 ? (
          <div className="flex flex-col gap-6 lg:sticky lg:top-24">
            {!sponsorMode ? (
              <AnnouncementPreview
                key={tournament.id}
                tournamentId={tournament.id}
              />
            ) : null}

            {!sponsorMode && !isMyTeamFetching && !isMyTeamError && myTeam ? (
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
                {!isSolo ? (
                  <CardFooter>
                    <Button
                      render={
                        <Link
                          href={`/tournaments/${tournament.id}/teams/${myTeam.id}`}
                        />
                      }
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

            {sponsorMode && funding.goalAmountMinor > 0 ? (
              <TournamentFundingCard
                funding={funding}
                status={tournament.status}
              />
            ) : null}
          </div>
        ) : null}
      </div>
    </main>
  );
}
