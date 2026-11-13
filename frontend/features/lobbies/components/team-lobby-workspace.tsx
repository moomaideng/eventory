"use client";

import React from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { ArrowLeft, Check, ShieldAlert } from "lucide-react";
import { $api } from "@/lib/api/client";
import { useAuth } from "@/context/auth-context";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@/components/ui/empty";
import { useLobbyActions } from "@/features/lobbies/hooks/use-lobby-actions";
import { LobbyWorkspaceSkeleton } from "@/features/lobbies/components/lobby-skeleton";
import { LobbyRosterCard } from "@/features/lobbies/components/lobby-roster-card";
import { LobbyInviteCard } from "@/features/lobbies/components/lobby-invite-card";
import { LobbyDetailsCard } from "@/features/lobbies/components/lobby-details-card";
import { MemberRegistrationForm } from "@/features/lobbies/components/member-registration-form";

export function TeamLobbyWorkspace({ inviteCode }: { inviteCode: string }) {
  const router = useRouter();
  const { user, isLoading: isAuthLoading, authorizationHeader } = useAuth();
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
    {
      enabled: Boolean(authorizationHeader),
      retry: false,
      staleTime: 0,
      refetchInterval: (query) =>
        query.state.data?.status === "PAYMENT_PENDING" ||
        query.state.data?.status === "FORMING"
          ? 5000
          : false,
    }
  );

  const {
    actionError,
    pendingAction,
    regenerateInvite,
    lockRoster,
    payRegistration,
    removeMember,
    disbandLobby,
  } = useLobbyActions({
    lobby,
    authorizationHeader,
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

  if (isAuthLoading || isLoading || !user) {
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
              This invite may be invalid, revoked, or linked to a disbanded
              lobby.
            </EmptyDescription>
          </EmptyHeader>
        </Empty>
      </div>
    );
  }

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

      <div className="flex flex-col gap-2">
        <h1 className="text-3xl font-bold tracking-tight sm:text-4xl">
          {lobby.name}
        </h1>
        <p className="text-muted-foreground">
          Teaming up for {lobby.tournament.name} · Hosted by{" "}
          {lobby.tournament.organizerName}
        </p>
      </div>

      {displayedError ? (
        <Alert variant="destructive">
          <ShieldAlert />
          <AlertTitle>Action unavailable</AlertTitle>
          <AlertDescription>{displayedError}</AlertDescription>
        </Alert>
      ) : null}

      {lobby.status === "ACCEPTED" ? (
        <Alert>
          <Check />
          <AlertTitle>Registration accepted</AlertTitle>
          <AlertDescription>
            Your team is confirmed to participate in this tournament.
          </AlertDescription>
        </Alert>
      ) : null}

      {lobby.status === "REJECTED" ? (
        <Alert variant="destructive">
          <ShieldAlert />
          <AlertTitle>Application rejected</AlertTitle>
          <AlertDescription>
            Your team&apos;s application was not accepted for this tournament.
          </AlertDescription>
        </Alert>
      ) : null}

      {lobby.status === "LOCKED" ? (
        <Alert>
          <Check />
          <AlertTitle>Waiting for organizer approval</AlertTitle>
          <AlertDescription>
            Your completed roster has been submitted to the organizer for
            review. Team members can no longer be changed.
          </AlertDescription>
        </Alert>
      ) : null}

      {lobby.status === "PAYMENT_PENDING" && (
        <Alert>
          <AlertTitle>Roster locked · awaiting payment</AlertTitle>
          <AlertDescription>
            Your team is not in the organizer review list yet. Complete the full
            team payment to submit. If payment fails, is cancelled, or expires,
            your roster and answers stay locked. Select Pay to try again.
          </AlertDescription>
        </Alert>
      )}
      <div className="grid gap-6 lg:grid-cols-[minmax(0,1fr)_20rem]">
        <div className="flex flex-col gap-6">
          <LobbyRosterCard
            lobby={lobby}
            pendingAction={pendingAction}
            onJoinLobby={() =>
              router.push(`/lobbies/${normalizedInviteCode}/join`)
            }
            onRemoveMember={removeMember}
          />
          {lobby.viewerRole !== "INVITEE" && (
            <MemberRegistrationForm
              key={`${lobby.id}-${lobby.viewerRole}`}
              lobby={lobby}
              onSaved={refetch}
            />
          )}
        </div>

        <div className="flex flex-col gap-6">
          {lobby.inviteCode && lobby.status === "FORMING" ? (
            <LobbyInviteCard
              lobby={lobby}
              pendingAction={pendingAction}
              copied={copied}
              onCopyInvite={copyInvite}
              onRegenerateInvite={regenerateInvite}
            />
          ) : null}

          <LobbyDetailsCard
            lobby={lobby}
            pendingAction={pendingAction}
            confirmDisband={confirmDisband}
            onSetConfirmDisband={setConfirmDisband}
            onLockRoster={lockRoster}
            onPayRegistration={payRegistration}
            onDisbandLobby={disbandLobby}
          />
        </div>
      </div>
    </div>
  );
}
