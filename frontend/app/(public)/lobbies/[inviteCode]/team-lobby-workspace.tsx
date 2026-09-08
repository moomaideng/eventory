"use client";

import React from "react";
import Link from "next/link";
import {
  ArrowLeft,
  CalendarDays,
  Check,
  ClipboardCopy,
  LockKeyhole,
  MapPin,
  RefreshCw,
  ShieldAlert,
  Trash2,
  UserMinus,
  UsersRound,
} from "lucide-react";
import { useRouter } from "next/navigation";
import { $api, apiClient } from "@/lib/api/client";
import type { components } from "@/lib/api/schema";
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
import { Skeleton } from "@/components/ui/skeleton";
import { Spinner } from "@/components/ui/spinner";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";

type Lobby = components["schemas"]["TeamLobbyResponse"];

export function TeamLobbyWorkspace({ inviteCode }: { inviteCode: string }) {
  const router = useRouter();
  const {
    user,
    isLoading: isAuthLoading,
    authorizationHeader,
    loginAsDev,
  } = useAuth();
  const [actionError, setActionError] = React.useState("");
  const [pendingAction, setPendingAction] = React.useState<string | null>(null);
  const [copied, setCopied] = React.useState(false);
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

  async function runAction(
    action: string,
    request: Promise<{ data?: unknown; error?: unknown }>,
    afterSuccess?: (data: Lobby | undefined) => void
  ) {
    setActionError("");
    setPendingAction(action);
    const { data, error: requestError } = await request;
    setPendingAction(null);
    if (requestError) {
      setActionError(
        problemMessage(requestError, "We could not update this team lobby.")
      );
      return;
    }
    afterSuccess?.(data as Lobby | undefined);
    if (!afterSuccess) {
      await refetch();
    }
  }

  async function copyInvite() {
    const inviteUrl = `${window.location.origin}/lobbies/${normalizedInviteCode}`;
    try {
      await navigator.clipboard.writeText(inviteUrl);
      setCopied(true);
      window.setTimeout(() => setCopied(false), 2_000);
    } catch {
      setActionError(
        "Could not copy the invite link. Share the code shown here instead."
      );
    }
  }

  if (!user && !isAuthLoading) {
    return <SignInRequired onDevLogin={() => loginAsDev("competitor")} />;
  }

  if (isAuthLoading || isLoading) {
    return <WorkspaceSkeleton />;
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

  const members = lobby.members ?? [];
  const isCaptain = lobby.viewerRole === "CAPTAIN";
  const isInvitee = lobby.viewerRole === "INVITEE";
  const isForming = lobby.status === "FORMING";
  const rosterReady = members.length >= lobby.tournament.minTeamSize;
  const rosterFull = members.length >= lobby.tournament.maxTeamSize;

  function joinLobby() {
    if (!authorizationHeader) return;
    void runAction(
      "join",
      apiClient.POST("/api/v1/lobbies/{inviteCode}/join", {
        params: { path: { inviteCode: normalizedInviteCode } },
        headers: { Authorization: authorizationHeader },
      })
    );
  }

  function regenerateInvite() {
    if (!authorizationHeader || !lobby) return;
    void runAction(
      "regenerate",
      apiClient.POST("/api/v1/lobbies/{id}/invite/regenerate", {
        params: { path: { id: lobby.id } },
        headers: { Authorization: authorizationHeader },
      }),
      (updatedLobby) => {
        if (updatedLobby) router.replace(`/lobbies/${updatedLobby.inviteCode}`);
      }
    );
  }

  function lockRoster() {
    if (!authorizationHeader || !lobby) return;
    void runAction(
      "lock",
      apiClient.POST("/api/v1/lobbies/{id}/lock", {
        params: { path: { id: lobby.id } },
        headers: { Authorization: authorizationHeader },
      })
    );
  }

  function removeMember(memberId: string) {
    if (!authorizationHeader || !lobby) return;
    void runAction(
      `remove-${memberId}`,
      apiClient.DELETE("/api/v1/lobbies/{id}/members/{memberId}", {
        params: { path: { id: lobby.id, memberId } },
        headers: { Authorization: authorizationHeader },
      })
    );
  }

  function disbandLobby() {
    if (!authorizationHeader || !lobby) return;
    void runAction(
      "disband",
      apiClient.DELETE("/api/v1/lobbies/{id}", {
        params: { path: { id: lobby.id } },
        headers: { Authorization: authorizationHeader },
      }),
      () => router.replace("/lobbies")
    );
  }

  return (
    <div className="container mx-auto flex w-full max-w-5xl flex-1 flex-col gap-6 px-4 py-12 sm:px-8">
      <Link
        href={`/tournaments/${lobby.tournament.id}`}
        className="text-muted-foreground hover:text-foreground flex w-fit items-center gap-2 text-sm font-medium"
      >
        <ArrowLeft />
        Tournament details
      </Link>

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

      {actionError ? (
        <Alert variant="destructive">
          <ShieldAlert />
          <AlertTitle>Action unavailable</AlertTitle>
          <AlertDescription>{actionError}</AlertDescription>
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
        <Card>
          <CardHeader>
            <div className="flex flex-wrap items-start justify-between gap-3">
              <div className="flex flex-col gap-1">
                <CardTitle>Team roster</CardTitle>
                <CardDescription>
                  {members.length} of {lobby.tournament.maxTeamSize} players ·
                  Minimum {lobby.tournament.minTeamSize}
                </CardDescription>
              </div>
              <Badge variant={rosterReady ? "secondary" : "outline"}>
                {rosterReady
                  ? "Ready to lock"
                  : `${lobby.tournament.minTeamSize - members.length} more needed`}
              </Badge>
            </div>
          </CardHeader>
          <CardContent>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Member</TableHead>
                  <TableHead>Joined</TableHead>
                  {isCaptain && isForming ? (
                    <TableHead className="text-right">Actions</TableHead>
                  ) : null}
                </TableRow>
              </TableHeader>
              <TableBody>
                {members.map((member) => {
                  const isMemberCaptain = member.role === "CAPTAIN";
                  return (
                    <TableRow key={member.id}>
                      <TableCell className="font-medium">
                        <div className="flex items-center gap-2">
                          <span>{member.displayName}</span>
                          <span className="text-muted-foreground text-xs">
                            @{member.handle}
                          </span>
                          {isMemberCaptain ? (
                            <Badge variant="outline">Captain</Badge>
                          ) : null}
                        </div>
                      </TableCell>
                      <TableCell className="text-muted-foreground">
                        {formatDate(member.joinedAt)}
                      </TableCell>
                      {isCaptain && isForming ? (
                        <TableCell className="text-right">
                          {!isMemberCaptain ? (
                            <Button
                              variant="ghost"
                              size="sm"
                              disabled={pendingAction === `remove-${member.id}`}
                              onClick={() => removeMember(member.id)}
                            >
                              {pendingAction === `remove-${member.id}` ? (
                                <Spinner data-icon="inline-start" />
                              ) : (
                                <UserMinus data-icon="inline-start" />
                              )}
                              Remove
                            </Button>
                          ) : null}
                        </TableCell>
                      ) : null}
                    </TableRow>
                  );
                })}
              </TableBody>
            </Table>
          </CardContent>
          {isInvitee ? (
            <CardFooter className="flex flex-wrap items-center justify-between gap-3">
              <p className="text-muted-foreground text-sm">
                {isForming && !rosterFull
                  ? "Join this roster to compete with the team."
                  : "This roster cannot accept new members."}
              </p>
              <Button
                disabled={!isForming || rosterFull || pendingAction === "join"}
                onClick={joinLobby}
              >
                {pendingAction === "join" ? (
                  <Spinner data-icon="inline-start" />
                ) : (
                  <UsersRound data-icon="inline-start" />
                )}
                Join team
              </Button>
            </CardFooter>
          ) : null}
        </Card>

        <div className="flex flex-col gap-6">
          <Card>
            <CardHeader>
              <CardTitle>Invite teammates</CardTitle>
              <CardDescription>
                Share this link or six-character code while the roster is
                forming.
              </CardDescription>
            </CardHeader>
            <CardContent className="flex flex-col gap-3">
              <div className="bg-muted rounded-lg p-3">
                <p className="text-muted-foreground text-xs font-medium">
                  Invite code
                </p>
                <p className="font-mono text-xl font-semibold tracking-[0.2em]">
                  {lobby.inviteCode}
                </p>
              </div>
              <Button
                variant="outline"
                disabled={pendingAction === "copy"}
                onClick={copyInvite}
              >
                {copied ? (
                  <Check data-icon="inline-start" />
                ) : (
                  <ClipboardCopy data-icon="inline-start" />
                )}
                {copied ? "Invite link copied" : "Copy invite link"}
              </Button>
            </CardContent>
            {isCaptain && isForming ? (
              <CardFooter>
                <Button
                  variant="outline"
                  className="w-full"
                  disabled={pendingAction === "regenerate"}
                  onClick={regenerateInvite}
                >
                  {pendingAction === "regenerate" ? (
                    <Spinner data-icon="inline-start" />
                  ) : (
                    <RefreshCw data-icon="inline-start" />
                  )}
                  Regenerate code
                </Button>
              </CardFooter>
            ) : null}
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Registration</CardTitle>
              <CardDescription>
                Team entry fee:{" "}
                {formatEntryFee(
                  lobby.tournament.entryFee,
                  lobby.tournament.currency
                )}
              </CardDescription>
            </CardHeader>
            <CardContent className="text-muted-foreground flex flex-col gap-2 text-sm">
              <span className="flex items-center gap-2">
                <CalendarDays /> Registration closes{" "}
                {formatDate(lobby.tournament.registrationDeadline)}
              </span>
              <span className="flex items-center gap-2">
                <MapPin /> {lobby.tournament.location}
              </span>
            </CardContent>
            {isCaptain && isForming ? (
              <CardFooter>
                <Button
                  className="w-full"
                  disabled={!rosterReady || pendingAction === "lock"}
                  onClick={lockRoster}
                >
                  {pendingAction === "lock" ? (
                    <Spinner data-icon="inline-start" />
                  ) : (
                    <LockKeyhole data-icon="inline-start" />
                  )}
                  Lock roster
                </Button>
              </CardFooter>
            ) : null}
          </Card>

          {isCaptain && isForming ? (
            <Card>
              <CardHeader>
                <CardTitle>Lobby management</CardTitle>
                <CardDescription>
                  Disbanding releases every current member while the roster is
                  still forming.
                </CardDescription>
              </CardHeader>
              <CardFooter className="flex flex-col items-stretch gap-3">
                {confirmDisband ? (
                  <Alert variant="destructive">
                    <ShieldAlert />
                    <AlertTitle>Disband this team?</AlertTitle>
                    <AlertDescription>
                      This permanently removes the forming team and releases its
                      members.
                      <span className="mt-3 flex flex-wrap gap-2">
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => setConfirmDisband(false)}
                        >
                          Cancel
                        </Button>
                        <Button
                          variant="destructive"
                          size="sm"
                          disabled={pendingAction === "disband"}
                          onClick={disbandLobby}
                        >
                          {pendingAction === "disband" ? (
                            <Spinner data-icon="inline-start" />
                          ) : (
                            <Trash2 data-icon="inline-start" />
                          )}
                          Disband team
                        </Button>
                      </span>
                    </AlertDescription>
                  </Alert>
                ) : (
                  <Button
                    variant="destructive"
                    onClick={() => setConfirmDisband(true)}
                  >
                    <Trash2 data-icon="inline-start" />
                    Disband team
                  </Button>
                )}
              </CardFooter>
            </Card>
          ) : null}
        </div>
      </div>
    </div>
  );
}

function SignInRequired({ onDevLogin }: { onDevLogin: () => void }) {
  return (
    <div className="container mx-auto flex w-full max-w-3xl flex-1 items-center px-4 py-12 sm:px-8">
      <Card className="w-full">
        <CardHeader>
          <CardTitle>Sign in to open this invitation</CardTitle>
          <CardDescription>
            You need an Eventory account before you can view or join a team
            roster.
          </CardDescription>
        </CardHeader>
        <CardFooter className="flex flex-wrap gap-3">
          <Link
            href="/login"
            className="text-primary text-sm font-medium underline-offset-4 hover:underline"
          >
            Sign in
          </Link>
          <Button variant="outline" onClick={onDevLogin}>
            Use Dev Quick Login
          </Button>
        </CardFooter>
      </Card>
    </div>
  );
}

function WorkspaceSkeleton() {
  return (
    <div className="container mx-auto flex w-full max-w-5xl flex-1 flex-col gap-6 px-4 py-12 sm:px-8">
      <Skeleton className="h-5 w-32" />
      <Skeleton className="h-12 w-72" />
      <div className="grid gap-6 lg:grid-cols-[minmax(0,1fr)_20rem]">
        <Skeleton className="h-96 w-full" />
        <div className="flex flex-col gap-6">
          <Skeleton className="h-64 w-full" />
          <Skeleton className="h-56 w-full" />
        </div>
      </div>
    </div>
  );
}

function LobbyStatusBadge({ status }: { status: Lobby["status"] }) {
  const labels: Record<Lobby["status"], string> = {
    FORMING: "Forming roster",
    LOCKED: "Roster locked",
    ACCEPTED: "Registration accepted",
    REJECTED: "Registration rejected",
  };
  return <Badge variant="outline">{labels[status]}</Badge>;
}

function formatDate(value: string) {
  return new Intl.DateTimeFormat("en-GB", {
    dateStyle: "medium",
    timeZone: "Asia/Bangkok",
  }).format(new Date(value));
}

function formatEntryFee(amount: number, currency: string) {
  if (amount === 0) return "Free";
  return new Intl.NumberFormat("en-US", {
    style: "currency",
    currency,
    maximumFractionDigits: 0,
  }).format(amount);
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
