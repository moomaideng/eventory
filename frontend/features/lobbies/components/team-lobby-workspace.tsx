"use client";

import React from "react";
import Link from "next/link";
import { ArrowLeft, Check, ShieldAlert } from "lucide-react";
import { $api } from "@/lib/api/client";
import { useAuth } from "@/context/auth-context";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@/components/ui/empty";
import { useLobbyActions } from "../hooks/use-lobby-actions";
import { LobbyStatusBadge } from "./lobby-status-badge";
import { LobbySignInCard } from "./lobby-sign-in-card";
import { LobbyWorkspaceSkeleton } from "./lobby-skeleton";
import { LobbyRosterCard } from "./lobby-roster-card";
import { LobbyInviteCard } from "./lobby-invite-card";
import { LobbyDetailsCard } from "./lobby-details-card";

export function TeamLobbyWorkspace({ inviteCode }: { inviteCode: string }) {
  const {
    user,
    isLoading: isAuthLoading,
    authorizationHeader,
    loginAsDev,
  } = useAuth();
  const [copied, setCopied] = React.useState(false);
  const [copyError, setCopyError] = React.useState("");
  const [confirmDisband, setConfirmDisband] = React.useState(false);
  const normalizedInviteCode = inviteCode.trim().toUpperCase();

  const {
    data: lobby,
    error,
    isLoading,
    refetch,
  } = $api.useQuery(
    "get",
    "/api/v1/lobbies/{inviteCode}",
    {
      params: { path: { inviteCode: normalizedInviteCode } },
      headers: authorizationHeader
        ? { Authorization: authorizationHeader }
        : {},
    },
    { enabled: Boolean(authorizationHeader), retry: false, staleTime: 0 }
  );

  const {
    actionError,
    pendingAction,
    joinLobby,
    regenerateInvite,
    lockRoster,
    removeMember,
    disbandLobby,
  } = useLobbyActions({
    lobby,
    authorizationHeader,
    normalizedInviteCode,
    refetch,
  });

  async function copyInvite() {
    setCopyError("");
    const inviteUrl = `${window.location.origin}/lobbies/${normalizedInviteCode}`;
    try {
      await navigator.clipboard.writeText(inviteUrl);
      setCopied(true);
      window.setTimeout(() => setCopied(false), 2_000);
    } catch {
      setCopyError(
        "Could not copy the invite link. Share the code shown here instead."
      );
    }
  }

  if (!user && !isAuthLoading) {
    return <LobbySignInCard onDevLogin={() => loginAsDev("competitor")} />;
  }

  if (isAuthLoading || isLoading) {
    return <LobbyWorkspaceSkeleton />;
  }

  if (error || !lobby) {
    return (
      <div className="container mx-auto flex w-full max-w-3xl flex-1 px-4 py-12 sm:px-8">
        <Empty className="border">
          <EmptyHeader>
            <EmptyMedia variant="icon">
              <ShieldAlert />
            </EmptyMedia>
            <EmptyTitle>Lobby unavailable</EmptyTitle>
            <EmptyDescription>
              This invite may be invalid, revoked, or linked to a disbanded lobby.
            </EmptyDescription>
          </EmptyHeader>
        </Empty>
      </div>
    );
  }

  const isCaptain = lobby.viewerRole === "CAPTAIN";
  const displayedError = actionError || copyError;

  return (
    <div className="container mx-auto flex w-full max-w-5xl flex-1 flex-col gap-6 px-4 py-12 sm:px-8">
      <Button
        variant="ghost"
        size="sm"
        render={<Link href={`/tournaments/${lobby.tournament.id}`} />}
        nativeButton={false}
        className="self-start"
      >
        <ArrowLeft data-icon="inline-start" />
        Tournament details
      </Button>

      <div className="flex flex-col gap-3">
        <div className="flex flex-wrap items-center gap-2">
          <Badge variant="secondary">{lobby.tournament.game}</Badge>
          <LobbyStatusBadge status={lobby.status} />
          {isCaptain ? <Badge variant="outline">Captain</Badge> : null}
        </div>
        <div className="flex flex-col gap-2">
          <h1 className="text-3xl font-bold tracking-tight sm:text-4xl">
            {lobby.name}
          </h1>
          <p className="text-muted-foreground">
            Teaming up for {lobby.tournament.name} · Hosted by{" "}
            {lobby.tournament.organizerName}
          </p>
        </div>
      </div>

      {displayedError ? (
        <Alert variant="destructive">
          <ShieldAlert />
          <AlertTitle>Action unavailable</AlertTitle>
          <AlertDescription>{displayedError}</AlertDescription>
        </Alert>
      ) : null}

      {lobby.status === "LOCKED" ? (
        <Alert>
          <Check />
          <AlertTitle>Roster locked</AlertTitle>
          <AlertDescription>
            This team can no longer be changed. The registration and payment
            workflow continues from the tournament organizer&apos;s process.
          </AlertDescription>
        </Alert>
      ) : null}

      <div className="grid gap-6 lg:grid-cols-[minmax(0,1fr)_20rem]">
        <LobbyRosterCard
          lobby={lobby}
          pendingAction={pendingAction}
          onJoinLobby={joinLobby}
          onRemoveMember={removeMember}
        />

        <div className="flex flex-col gap-6">
          <LobbyInviteCard
            lobby={lobby}
            pendingAction={pendingAction}
            copied={copied}
            onCopyInvite={copyInvite}
            onRegenerateInvite={regenerateInvite}
          />

          <LobbyDetailsCard
            lobby={lobby}
            pendingAction={pendingAction}
            confirmDisband={confirmDisband}
            onSetConfirmDisband={setConfirmDisband}
            onLockRoster={lockRoster}
            onDisbandLobby={disbandLobby}
          />
        </div>
      </div>
    </div>
  );
}
