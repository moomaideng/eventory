"use client";

import React from "react";
import Link from "next/link";
import { ArrowLeft, ShieldAlert, UsersRound } from "lucide-react";
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
  FieldError,
  FieldGroup,
  FieldLabel,
} from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Spinner } from "@/components/ui/spinner";
import { problemMessage } from "../utils";
import { createTeamLobbySchema } from "../schemas";
import { TeamCreateSkeleton } from "./team-create-skeleton";
import { TournamentSummaryCard } from "./tournament-summary-card";
import { RegistrationDialog } from "@/features/registration/components/registration-dialog";
import type { RegistrationSubmission } from "@/features/registration/schemas";
import { registrationUnavailableReason } from "@/features/registration/utils";

export function TeamLobbyCreate({ tournamentId }: { tournamentId: string }) {
  const router = useRouter();
  const { user, isLoading: isAuthLoading, authorizationHeader } = useAuth();
  const [teamName, setTeamName] = React.useState("");
  const [formError, setFormError] = React.useState("");
  const [isSubmitting, setIsSubmitting] = React.useState(false);
  const [dialogOpen, setRegistrationOpen] = React.useState(false);

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

  function openRegistration(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!details) return;
    const unavailable = registrationUnavailableReason(details.tournament);
    if (unavailable) {
      setFormError(unavailable);
      return;
    }
    const result = createTeamLobbySchema.safeParse({ name: teamName });
    if (!result.success) {
      setFormError(result.error.issues[0]?.message ?? "Invalid team name.");
      return;
    }
    if (!authorizationHeader) {
      setFormError("Sign in before creating a team lobby.");
      return;
    }

    setFormError("");
    setRegistrationOpen(true);
  }

  async function createLobby(registration: RegistrationSubmission) {
    if (!authorizationHeader) throw new Error("Sign in to create a team.");
    setIsSubmitting(true);
    try {
      const { data, error: createError } = await apiClient.POST(
        "/api/v1/tournaments/{tournamentId}/lobbies",
        {
          params: { path: { tournamentId } },
          headers: { Authorization: authorizationHeader },
          body: { name: teamName.trim(), registration },
        }
      );
      if (createError || !data) {
        throw new Error(
          problemMessage(createError, "We could not create this team lobby.")
        );
      }
      router.push(`/lobbies/${data.inviteCode}`);
    } finally {
      setIsSubmitting(false);
    }
  }

  if (isLoading) {
    return <TeamCreateSkeleton />;
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
  const supportsTeams = tournament.registrationMode === "TEAM";
  const registrationUnavailable = registrationUnavailableReason(tournament);

  return (
    <div className="container mx-auto flex w-full max-w-4xl flex-1 flex-col gap-6 px-4 py-12 sm:px-8">
      <Button
        variant="ghost"
        size="sm"
        render={<Link href={`/tournaments/${tournamentId}`} />}
        nativeButton={false}
        className="self-start"
      >
        <ArrowLeft data-icon="inline-start" />
        Tournament details
      </Button>

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

      <TournamentSummaryCard tournament={tournament} />

      {!supportsTeams || registrationUnavailable ? (
        <Alert variant="destructive">
          <ShieldAlert />
          <AlertTitle>Team registration is unavailable</AlertTitle>
          <AlertDescription>
            {!supportsTeams
              ? "This tournament only accepts solo entries."
              : registrationUnavailable}
          </AlertDescription>
        </Alert>
      ) : isAuthLoading || !user ? (
        <TeamCreateSkeleton />
      ) : (
        <Card>
          <CardHeader>
            <CardTitle>Name your team</CardTitle>
            <CardDescription>
              Teammates will see this name when they open your invite link.
            </CardDescription>
          </CardHeader>
          <form onSubmit={openRegistration}>
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
                  {formError ? <FieldError>{formError}</FieldError> : null}
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
      <RegistrationDialog
        open={dialogOpen}
        onOpenChange={(open) => {
          if (!isSubmitting) setRegistrationOpen(open);
        }}
        tournamentId={tournamentId}
        authorizationHeader={authorizationHeader ?? ""}
        title="Your captain registration"
        submitLabel="Submit and create team"
        onSubmit={createLobby}
      />
    </div>
  );
}
