"use client";

import React from "react";
import { ArrowRight, UsersRound } from "lucide-react";
import { useRouter } from "next/navigation";
import { useAuth } from "@/context/auth-context";
import { apiClient } from "@/lib/api/client";
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
  Field,
  FieldDescription,
  FieldError,
  FieldGroup,
  FieldLabel,
} from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Spinner } from "@/components/ui/spinner";

import { lobbyCodeSchema } from "../schemas";
import { problemMessage } from "../utils";

export function LobbyCodeForm() {
  const router = useRouter();
  const { authorizationHeader } = useAuth();
  const [inviteCode, setInviteCode] = React.useState("");
  const [error, setError] = React.useState("");
  const [isResolving, setIsResolving] = React.useState(false);

  async function submit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const result = lobbyCodeSchema.safeParse(inviteCode);
    if (!result.success) {
      setError(result.error.issues[0]?.message ?? "Invalid invite code.");
      return;
    }
    if (!authorizationHeader) {
      setError("Sign in to resolve this invite code.");
      return;
    }
    setError("");
    setIsResolving(true);
    try {
      const { data, error: resolveError } = await apiClient.GET(
        "/api/v1/lobbies/{inviteCode}",
        {
          params: { path: { inviteCode: result.data } },
          headers: { Authorization: authorizationHeader },
        }
      );
      if (resolveError || !data) {
        setError(
          problemMessage(
            resolveError,
            "No active team lobby matched that invite code."
          )
        );
        return;
      }
      router.push(
        `/tournaments/${data.tournamentId}/teams/${data.teamId}?invite=${result.data}`
      );
    } catch {
      setError("Connection error. Please check your network and try again.");
    } finally {
      setIsResolving(false);
    }
  }

  return (
    <div className="container mx-auto flex w-full max-w-md flex-1 items-center justify-center px-4 py-12 sm:px-8">
      <Card className="w-full">
        <CardHeader>
          <div className="flex items-center gap-3">
            <div className="bg-muted flex size-10 items-center justify-center rounded-lg">
              <UsersRound />
            </div>
            <div className="flex flex-col gap-1">
              <CardTitle>Join a team lobby</CardTitle>
              <CardDescription>
                Enter the invite code from your team captain to view the roster.
              </CardDescription>
            </div>
          </div>
        </CardHeader>
        <form onSubmit={submit}>
          <CardContent>
            <FieldGroup>
              <Field data-invalid={Boolean(error)}>
                <FieldLabel htmlFor="lobby-code">Invite code</FieldLabel>
                <Input
                  id="lobby-code"
                  value={inviteCode}
                  onChange={(event) => {
                    setInviteCode(event.target.value.toUpperCase());
                    setError("");
                  }}
                  placeholder="ABC123"
                  autoComplete="off"
                  maxLength={6}
                  aria-invalid={Boolean(error)}
                  className="font-mono tracking-[0.3em] uppercase"
                />
                <FieldDescription>
                  Codes are six characters and are not case-sensitive.
                </FieldDescription>
                {error ? <FieldError>{error}</FieldError> : null}
              </Field>
            </FieldGroup>
          </CardContent>
          <CardFooter className="justify-end">
            <Button type="submit" disabled={isResolving}>
              {isResolving ? (
                <Spinner data-icon="inline-start" />
              ) : (
                <ArrowRight data-icon="inline-end" />
              )}
              View lobby
            </Button>
          </CardFooter>
        </form>
      </Card>
    </div>
  );
}
