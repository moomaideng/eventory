"use client";

import React from "react";
import { ArrowRight, UsersRound } from "lucide-react";
import { useRouter } from "next/navigation";
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
  FieldGroup,
  FieldLabel,
} from "@/components/ui/field";
import { Input } from "@/components/ui/input";

export function LobbyCodeForm() {
  const router = useRouter();
  const [inviteCode, setInviteCode] = React.useState("");
  const [error, setError] = React.useState("");

  function submit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const normalizedCode = inviteCode.trim().toUpperCase();
    if (!/^[A-Z0-9]{6}$/.test(normalizedCode)) {
      setError("Enter the six-character code your captain shared.");
      return;
    }
    router.push(`/lobbies/${normalizedCode}`);
  }

  return (
    <div className="container mx-auto flex w-full max-w-2xl flex-1 items-center px-4 py-12 sm:px-8">
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
                {error ? (
                  <FieldDescription className="text-destructive">
                    {error}
                  </FieldDescription>
                ) : null}
              </Field>
            </FieldGroup>
          </CardContent>
          <CardFooter className="justify-end">
            <Button type="submit">
              View lobby
              <ArrowRight data-icon="inline-end" />
            </Button>
          </CardFooter>
        </form>
      </Card>
    </div>
  );
}
