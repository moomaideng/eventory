"use client";

import { useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { apiClient } from "@/lib/api/client";
import { useAuth } from "@/context/auth-context";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Spinner } from "@/components/ui/spinner";
import { RegistrationAnswerFields } from "@/features/registration-form/components/registration-answer-fields";
import { useRegistrationAnswers } from "@/features/registration-form/hooks/use-registration-answers";
import { problemMessage, type Lobby } from "@/features/lobbies/utils";

export function MemberRegistrationForm({
  lobby,
  onSaved,
}: {
  lobby: Lobby;
  onSaved: () => Promise<unknown>;
}) {
  const { authorizationHeader } = useAuth();
  const queryClient = useQueryClient();
  const state = useRegistrationAnswers(
    lobby.tournament.id,
    authorizationHeader
  );
  const [pending, setPending] = useState(false);
  const [error, setError] = useState("");
  const [saved, setSaved] = useState(false);
  const joining = lobby.viewerRole === "INVITEE";
  const available =
    lobby.status === "FORMING" &&
    lobby.tournament.status === "REGISTRATION_OPEN" &&
    (!joining || (lobby.members?.length ?? 0) < lobby.tournament.maxTeamSize);

  async function submit(event: React.FormEvent) {
    event.preventDefault();
    const answers = state.validate();
    if (!answers || !authorizationHeader || pending || !available) return;
    setPending(true);
    setSaved(false);
    setError("");
    try {
      const result = joining
        ? await apiClient.POST("/api/v1/lobbies/{inviteCode}/join", {
            params: { path: { inviteCode: lobby.inviteCode! } },
            headers: { Authorization: authorizationHeader },
            body: { answers },
          })
        : await apiClient.PUT("/api/v1/lobbies/{id}/registration", {
            params: { path: { id: lobby.id } },
            headers: { Authorization: authorizationHeader },
            body: { answers },
          });
      if (result.error) {
        setError(
          problemMessage(result.error, "Your registration could not be saved.")
        );
        return;
      }
      await queryClient.invalidateQueries({
        queryKey: ["get", "/api/v1/tournaments/{tournamentId}/my-team"],
      });
      await state.refresh();
      await onSaved();
      setSaved(true);
    } catch {
      setError(
        "Connection error. Refresh the lobby to check whether your answers were saved before retrying."
      );
    } finally {
      setPending(false);
    }
  }

  if (lobby.status !== "FORMING") return null;
  return (
    <Card id="member-registration">
      <CardHeader>
        <CardTitle>
          {joining ? "Register to join this team" : "Your registration answers"}
        </CardTitle>
        <CardDescription>
          Answer for yourself. All members must complete the questions before
          the captain can submit. The organizer accepts or rejects the whole
          team.
        </CardDescription>
      </CardHeader>
      <form noValidate onSubmit={submit}>
        <CardContent className="flex flex-col gap-4">
          <fieldset disabled={pending} className="flex min-w-0 flex-col gap-4">
            <legend className="sr-only">Your registration answers</legend>
            <RegistrationAnswerFields state={state} />
          </fieldset>
          {error && (
            <Alert variant="destructive">
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}
          {saved && (
            <Alert>
              <AlertDescription>
                Your registration answers have been saved.
              </AlertDescription>
            </Alert>
          )}
        </CardContent>
        <CardFooter>
          <Button
            type="submit"
            disabled={!available || state.blocked || pending}
          >
            {pending && <Spinner data-icon="inline-start" />}
            {joining ? "Submit answers & join team" : "Save my answers"}
          </Button>
        </CardFooter>
      </form>
    </Card>
  );
}
