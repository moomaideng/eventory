"use client";

import React from "react";
import Link from "next/link";
import {
  ArrowLeft,
  CalendarDays,
  MapPin,
  ShieldAlert,
  UsersRound,
} from "lucide-react";
import { useRouter } from "next/navigation";
import { $api, apiClient } from "@/lib/api/client";
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
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@/components/ui/empty";
import {
  Field,
  FieldDescription,
  FieldGroup,
  FieldLabel,
} from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Skeleton } from "@/components/ui/skeleton";
import { Spinner } from "@/components/ui/spinner";

export function TeamLobbyCreate({ tournamentId }: { tournamentId: string }) {
  const router = useRouter();
  const {
    user,
    isLoading: isAuthLoading,
    authorizationHeader,
    loginAsDev,
  } = useAuth();
  const [teamName, setTeamName] = React.useState("");
  const [formError, setFormError] = React.useState("");
  const [isSubmitting, setIsSubmitting] = React.useState(false);

  const {
    data: details,
    error,
    isLoading,
  } = $api.useQuery(
    "get",
    "/api/v1/tournaments/{tournamentId}",
    { params: { path: { tournamentId } } },
    { staleTime: 30_000 }
  );

  async function createLobby(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const name = teamName.trim();
    if (!name) {
      setFormError("Choose a team name before creating the lobby.");
      return;
    }
    if (!authorizationHeader) {
      setFormError("Sign in before creating a team lobby.");
      return;
    }

    setFormError("");
    setIsSubmitting(true);
    const { data, error: createError } = await apiClient.POST(
      "/api/v1/tournaments/{tournamentId}/lobbies",
      {
        params: { path: { tournamentId } },
        headers: { Authorization: authorizationHeader },
        body: { name },
      }
    );
    setIsSubmitting(false);
    if (createError || !data) {
      setFormError(
        problemMessage(createError, "We could not create this team lobby.")
      );
      return;
    }
    router.push(`/lobbies/${data.inviteCode}`);
  }

  if (isLoading) {
    return <CreatePageSkeleton />;
  }

  if (error || !details) {
    return (
      <div className="container mx-auto flex w-full max-w-3xl flex-1 px-4 py-12 sm:px-8">
        <Empty className="border">
          <EmptyHeader>
            <EmptyMedia variant="icon">
              <ShieldAlert />
            </EmptyMedia>
            <EmptyTitle>Tournament unavailable</EmptyTitle>
            <EmptyDescription>
              This tournament may not exist or is no longer published.
            </EmptyDescription>
          </EmptyHeader>
        </Empty>
      </div>
    );
  }

  const { tournament } = details;
  const supportsTeams =
    tournament.registrationMode === "TEAM" ||
    tournament.registrationMode === "BOTH";
  const registrationOpen = tournament.status === "REGISTRATION_OPEN";

  return (
    <div className="container mx-auto flex w-full max-w-4xl flex-1 flex-col gap-6 px-4 py-12 sm:px-8">
      <Link
        href={`/tournaments/${tournamentId}`}
        className="text-muted-foreground hover:text-foreground flex w-fit items-center gap-2 text-sm font-medium"
      >
        <ArrowLeft />
        Tournament details
      </Link>

      <div className="flex flex-col gap-3">
        <Badge variant="secondary" className="w-fit">
          Team registration
        </Badge>
        <div className="flex flex-col gap-2">
          <h1 className="text-3xl font-bold tracking-tight sm:text-4xl">
            Create your team
          </h1>
          <p className="text-muted-foreground max-w-2xl">
            Start a roster for {tournament.name}, then share a secure invite
            code with your teammates.
          </p>
        </div>
      </div>

      <Card>
        <CardHeader>
          <div className="flex flex-wrap items-start justify-between gap-3">
            <div className="flex flex-col gap-1">
              <CardTitle>{tournament.name}</CardTitle>
              <CardDescription>
                {tournament.game} · Hosted by {tournament.organizerName}
              </CardDescription>
            </div>
            <Badge variant="outline">
              {tournament.minTeamSize}–{tournament.maxTeamSize} players
            </Badge>
          </div>
        </CardHeader>
        <CardContent className="flex flex-col gap-3 text-sm">
          <p className="text-muted-foreground">{tournament.description}</p>
          <div className="text-muted-foreground flex flex-wrap gap-x-5 gap-y-2">
            <span className="flex items-center gap-2">
              <CalendarDays /> Registration closes{" "}
              {formatDate(tournament.registrationDeadline)}
            </span>
            <span className="flex items-center gap-2">
              <MapPin /> {tournament.location}
            </span>
          </div>
        </CardContent>
      </Card>

      {!supportsTeams || !registrationOpen ? (
        <Alert variant="destructive">
          <ShieldAlert />
          <AlertTitle>Team registration is unavailable</AlertTitle>
          <AlertDescription>
            {!supportsTeams
              ? "This tournament only accepts solo entries."
              : "Registration is not currently open for this tournament."}
          </AlertDescription>
        </Alert>
      ) : !user && !isAuthLoading ? (
        <Card>
          <CardHeader>
            <CardTitle>Sign in to become captain</CardTitle>
            <CardDescription>
              A captain account is required to create and manage a roster.
            </CardDescription>
          </CardHeader>
          <CardFooter className="flex flex-wrap gap-2">
            <Link
              href="/login"
              className="text-primary text-sm font-medium underline-offset-4 hover:underline"
            >
              Sign in
            </Link>
            <Button variant="outline" onClick={() => loginAsDev("competitor")}>
              Use Dev Quick Login
            </Button>
          </CardFooter>
        </Card>
      ) : isAuthLoading ? (
        <Card>
          <CardContent className="pt-6">
            <Skeleton className="h-24 w-full" />
          </CardContent>
        </Card>
      ) : (
        <Card>
          <CardHeader>
            <CardTitle>Name your team</CardTitle>
            <CardDescription>
              Teammates will see this name when they open your invite link.
            </CardDescription>
          </CardHeader>
          <form onSubmit={createLobby}>
            <CardContent>
              <FieldGroup>
                <Field data-invalid={Boolean(formError)}>
                  <FieldLabel htmlFor="team-name">Team name</FieldLabel>
                  <Input
                    id="team-name"
                    value={teamName}
                    onChange={(event) => {
                      setTeamName(event.target.value);
                      setFormError("");
                    }}
                    placeholder="Night Owls"
                    maxLength={120}
                    aria-invalid={Boolean(formError)}
                  />
                  <FieldDescription>
                    Up to 120 characters. You can manage invitations on the next
                    page.
                  </FieldDescription>
                  {formError ? (
                    <FieldDescription className="text-destructive">
                      {formError}
                    </FieldDescription>
                  ) : null}
                </Field>
              </FieldGroup>
            </CardContent>
            <CardFooter className="justify-end">
              <Button type="submit" disabled={isSubmitting}>
                {isSubmitting ? (
                  <Spinner data-icon="inline-start" />
                ) : (
                  <UsersRound data-icon="inline-start" />
                )}
                {isSubmitting ? "Creating lobby…" : "Create team lobby"}
              </Button>
            </CardFooter>
          </form>
        </Card>
      )}
    </div>
  );
}

function CreatePageSkeleton() {
  return (
    <div className="container mx-auto flex w-full max-w-4xl flex-1 flex-col gap-6 px-4 py-12 sm:px-8">
      <Skeleton className="h-5 w-36" />
      <Skeleton className="h-12 w-72" />
      <Skeleton className="h-52 w-full" />
      <Skeleton className="h-48 w-full" />
    </div>
  );
}

function formatDate(value: string) {
  return new Intl.DateTimeFormat("en-GB", {
    dateStyle: "medium",
    timeZone: "Asia/Bangkok",
  }).format(new Date(value));
}

function problemMessage(error: unknown, fallback: string) {
  if (
    error &&
    typeof error === "object" &&
    "detail" in error &&
    typeof error.detail === "string"
  ) {
    return error.detail;
  }
  return fallback;
}
