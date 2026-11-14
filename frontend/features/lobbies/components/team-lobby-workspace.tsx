"use client";

import React from "react";
import { useRouter } from "next/navigation";
import { Check, ShieldAlert } from "lucide-react";
import { $api, apiClient } from "@/lib/api/client";
import { useAuth } from "@/context/auth-context";
import { PageHeader } from "@/components/page-header";
import { LobbyStatusBadge } from "./lobby-status-badge";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
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
import { useLobbyActions } from "@/features/lobbies/hooks/use-lobby-actions";
import { LobbyWorkspaceSkeleton } from "@/features/lobbies/components/lobby-skeleton";
import { LobbyRosterCard } from "@/features/lobbies/components/lobby-roster-card";
import { LobbyInviteCard } from "@/features/lobbies/components/lobby-invite-card";
import { LobbyDetailsCard } from "@/features/lobbies/components/lobby-details-card";
import { MemberRegistrationForm } from "@/features/lobbies/components/member-registration-form";
import { lobbyCodeSchema } from "@/features/lobbies/schemas";
import { problemMessage } from "@/features/lobbies/utils";

export function TeamLobbyWorkspace({
  tournamentId,
  teamId,
  inviteCode,
}: {
  tournamentId: string;
  teamId: string;
  inviteCode?: string;
}) {
  const router = useRouter();
  const { user, isLoading: isAuthLoading, authorizationHeader } = useAuth();
  const [copied, setCopied] = React.useState(false);
  const [copyError, setCopyError] = React.useState("");
  const [confirmDisband, setConfirmDisband] = React.useState(false);
  const [inviteDialogOpen, setInviteDialogOpen] = React.useState(false);
  const [enteredInviteCode, setEnteredInviteCode] = React.useState("");
  const [inviteDialogError, setInviteDialogError] = React.useState("");
  const [isCheckingInvite, setIsCheckingInvite] = React.useState(false);
  const normalizedInviteCode = inviteCode?.trim().toUpperCase();

  const {
    data: lobby,
    error,
    isLoading,
    refetch,
  } = $api.useQuery(
    "get",
    "/api/v1/tournaments/{tournamentId}/teams/{teamId}",
    {
      params: { path: { tournamentId, teamId } },
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
    if (!lobby?.inviteCode) return;
    setCopyError("");
    const inviteUrl = `${window.location.origin}/lobbies/${lobby.inviteCode}`;
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

  function handleJoinLobby() {
    if (normalizedInviteCode) {
      router.push(
        `/tournaments/${tournamentId}/teams/${teamId}/join?invite=${normalizedInviteCode}`
      );
      return;
    }
    setEnteredInviteCode("");
    setInviteDialogError("");
    setInviteDialogOpen(true);
  }

  async function submitInviteCode(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const parsed = lobbyCodeSchema.safeParse(enteredInviteCode);
    if (!parsed.success) {
      setInviteDialogError(
        parsed.error.issues[0]?.message ?? "Invalid invite code."
      );
      return;
    }
    if (!authorizationHeader) {
      setInviteDialogError("Sign in to join this team.");
      return;
    }
    setInviteDialogError("");
    setIsCheckingInvite(true);
    try {
      const { data, error: resolveError } = await apiClient.GET(
        "/api/v1/lobbies/{inviteCode}",
        {
          params: { path: { inviteCode: parsed.data } },
          headers: { Authorization: authorizationHeader },
        }
      );
      if (resolveError || !data) {
        setInviteDialogError(
          problemMessage(
            resolveError,
            "Invalid invite code. Check with your team captain."
          )
        );
        return;
      }
      if (data.tournamentId !== tournamentId || data.teamId !== teamId) {
        setInviteDialogError("This invite code belongs to a different team.");
        return;
      }
      setInviteDialogOpen(false);
      router.push(
        `/tournaments/${tournamentId}/teams/${teamId}/join?invite=${parsed.data}`
      );
    } catch {
      setInviteDialogError(
        "Connection error. Please check your network and try again."
      );
    } finally {
      setIsCheckingInvite(false);
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
              This team lobby may not exist or may have been disbanded.
            </EmptyDescription>
          </EmptyHeader>
        </Empty>
      </div>
    );
  }

  const displayedError = actionError || copyError;

  return (
    <div className="container mx-auto flex w-full max-w-5xl flex-1 flex-col gap-6 px-4 py-12 sm:px-8">
      <PageHeader
        breadcrumbs={[
          { label: "Tournaments", href: "/tournaments" },
          {
            label: lobby.tournament.name,
            href: `/tournaments/${lobby.tournament.id}`,
          },
          { label: lobby.name },
        ]}
        tags={<LobbyStatusBadge status={lobby.status} />}
        title={lobby.name}
        description={`Teaming up for ${lobby.tournament.name} · Hosted by ${lobby.tournament.organizerName}`}
      />

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
            onJoinLobby={handleJoinLobby}
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

      <Dialog open={inviteDialogOpen} onOpenChange={setInviteDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Enter invite code</DialogTitle>
            <DialogDescription>
              Ask your team captain for the six-character invite code to join{" "}
              {lobby.name}.
            </DialogDescription>
          </DialogHeader>
          <form onSubmit={submitInviteCode} className="flex flex-col gap-6">
            <FieldGroup>
              <Field data-invalid={Boolean(inviteDialogError)}>
                <FieldLabel htmlFor="team-invite-code">Invite code</FieldLabel>
                <Input
                  id="team-invite-code"
                  autoFocus
                  value={enteredInviteCode}
                  onChange={(event) => {
                    setEnteredInviteCode(event.target.value.toUpperCase());
                    setInviteDialogError("");
                  }}
                  placeholder="ABC123"
                  autoComplete="off"
                  autoCorrect="off"
                  autoCapitalize="characters"
                  spellCheck={false}
                  maxLength={6}
                  aria-invalid={Boolean(inviteDialogError)}
                  className="font-mono tracking-[0.3em] uppercase"
                />
                <FieldDescription>
                  Codes are six characters and are not case-sensitive.
                </FieldDescription>
                {inviteDialogError ? (
                  <FieldError>{inviteDialogError}</FieldError>
                ) : null}
              </Field>
            </FieldGroup>
            <DialogFooter>
              <Button
                type="button"
                variant="outline"
                onClick={() => setInviteDialogOpen(false)}
                disabled={isCheckingInvite}
              >
                Cancel
              </Button>
              <Button type="submit" disabled={isCheckingInvite}>
                {isCheckingInvite ? (
                  <Spinner data-icon="inline-start" />
                ) : null}
                Continue
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </div>
  );
}
