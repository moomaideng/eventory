"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { $api, apiClient } from "@/lib/api/client";
import { useAuth } from "@/context/auth-context";
import { PageHeader } from "@/components/page-header";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { LobbyWorkspaceSkeleton } from "@/features/lobbies/components/lobby-skeleton";
import { MemberRegistrationForm } from "@/features/lobbies/components/member-registration-form";
import { TournamentHeaderTags } from "@/features/tournaments/components/tournament-header-tags";
import { isRegistrationOpen } from "@/features/tournaments/utils";

export function TeamLobbyJoin({
  tournamentId,
  teamId,
  inviteCode,
}: {
  tournamentId: string;
  teamId: string;
  inviteCode: string;
}) {
  const router = useRouter();
  const queryClient = useQueryClient();
  const { authorizationHeader, isLoading: authLoading } = useAuth();
  const lobbyPath = `/tournaments/${tournamentId}/teams/${teamId}`;
  const normalizedInviteCode = inviteCode.trim().toUpperCase();
  const {
    data: lobby,
    isPending,
    error,
    isFetching,
  } = $api.useQuery(
    "get",
    "/api/v1/tournaments/{tournamentId}/teams/{teamId}",
    {
      params: { path: { tournamentId, teamId } },
      headers: { Authorization: authorizationHeader ?? "" },
    },
    {
      enabled: Boolean(authorizationHeader),
      retry: false,
      staleTime: 0,
      refetchInterval: 5000,
    }
  );
  const alreadyJoined = Boolean(lobby && lobby.viewerRole !== "INVITEE");
  const inviteQuery = $api.useQuery(
    "get",
    "/api/v1/lobbies/{inviteCode}",
    {
      params: { path: { inviteCode: normalizedInviteCode } },
      headers: { Authorization: authorizationHeader ?? "" },
    },
    {
      enabled: Boolean(authorizationHeader && lobby && !alreadyJoined),
      retry: false,
      staleTime: 0,
    }
  );
  const myTeamQuery = useQuery({
    queryKey: [
      "get",
      "/api/v1/tournaments/{tournamentId}/my-team",
      tournamentId,
      authorizationHeader,
    ],
    enabled: Boolean(authorizationHeader && lobby && !alreadyJoined),
    retry: false,
    staleTime: 0,
    queryFn: async () => {
      const result = await apiClient.GET(
        "/api/v1/tournaments/{tournamentId}/my-team",
        {
          params: { path: { tournamentId } },
          headers: { Authorization: authorizationHeader ?? "" },
        }
      );
      if (result.response.status === 404) return null;
      if (result.error || !result.data)
        throw new Error("Failed to check registration status");
      return result.data;
    },
  });

  useEffect(() => {
    if (alreadyJoined && !isFetching && !error) router.replace(lobbyPath);
  }, [alreadyJoined, isFetching, error, router, lobbyPath]);

  async function returnToLobby() {
    await queryClient.invalidateQueries({
      queryKey: ["get", "/api/v1/tournaments/{tournamentId}/teams/{teamId}"],
    });
    router.replace(lobbyPath);
  }

  if (
    authLoading ||
    isPending ||
    (!alreadyJoined && Boolean(lobby) && inviteQuery.isPending) ||
    (alreadyJoined && !error)
  ) {
    return <LobbyWorkspaceSkeleton />;
  }

  const inviteMismatch =
    !alreadyJoined &&
    Boolean(lobby) &&
    (Boolean(inviteQuery.error) ||
      !inviteQuery.data ||
      inviteQuery.data.tournamentId !== tournamentId ||
      inviteQuery.data.teamId !== teamId);

  const unavailable =
    !lobby ||
    inviteMismatch ||
    lobby.status !== "FORMING" ||
    !isRegistrationOpen(lobby.tournament) ||
    (lobby.members?.length ?? 0) >= lobby.tournament.maxTeamSize;

  return (
    <div className="container mx-auto flex w-full max-w-7xl flex-1 flex-col gap-8 px-4 py-10 sm:px-8">
      {myTeamQuery.data?.status === "REJECTED" ? (
        <>
          <PageHeader
            breadcrumbs={[
              { label: "Tournaments", href: "/tournaments" },
              ...(lobby
                ? [
                    {
                      label: lobby.tournament.name,
                      href: `/tournaments/${tournamentId}`,
                    },
                    { label: lobby.name, href: lobbyPath },
                  ]
                : []),
              { label: "Join team" },
            ]}
            title="Unable to join this team"
          />
          <Alert variant="destructive" className="max-w-3xl">
            <AlertTitle>Application rejected</AlertTitle>
            <AlertDescription>
              Your application was not accepted for this tournament, so you
              cannot join another team.
            </AlertDescription>
          </Alert>
        </>
      ) : error || unavailable ? (
        <>
          <PageHeader
            breadcrumbs={[
              { label: "Tournaments", href: "/tournaments" },
              ...(lobby
                ? [
                    {
                      label: lobby.tournament.name,
                      href: `/tournaments/${tournamentId}`,
                    },
                    { label: lobby.name, href: lobbyPath },
                  ]
                : []),
              { label: "Join team" },
            ]}
            title="Unable to join this team"
          />
          <Alert className="max-w-3xl">
            <AlertTitle>Unable to join this team</AlertTitle>
            <AlertDescription>
              {error
                ? "The lobby could not be loaded. Return to the lobby and try again."
                : inviteMismatch
                  ? "This invite code is invalid or does not match this team."
                  : "This team is full or is no longer accepting new members."}
            </AlertDescription>
          </Alert>
        </>
      ) : (
        <>
          <PageHeader
            breadcrumbs={[
              { label: "Tournaments", href: "/tournaments" },
              {
                label: lobby.tournament.name,
                href: `/tournaments/${tournamentId}`,
              },
              { label: lobby.name, href: lobbyPath },
              { label: "Join team" },
            ]}
            title={`Join ${lobby.name}`}
            tags={
              <TournamentHeaderTags
                game={lobby.tournament.game}
                status={lobby.tournament.status}
                registrationMode={lobby.tournament.registrationMode}
                registrationDeadline={lobby.tournament.registrationDeadline}
              />
            }
            description={`Complete your registration for ${lobby.tournament.name}. You will return to the lobby after submitting.`}
          />
          <div className="w-full max-w-3xl">
            <MemberRegistrationForm
              lobby={lobby}
              inviteCode={normalizedInviteCode}
              onSaved={returnToLobby}
            />
          </div>
        </>
      )}
    </div>
  );
}
