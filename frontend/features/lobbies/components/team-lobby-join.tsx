"use client";

import { useEffect } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { ArrowLeft } from "lucide-react";
import { $api, apiClient } from "@/lib/api/client";
import { useAuth } from "@/context/auth-context";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { LobbyWorkspaceSkeleton } from "@/features/lobbies/components/lobby-skeleton";
import { MemberRegistrationForm } from "@/features/lobbies/components/member-registration-form";
import { isRegistrationOpen } from "@/features/tournaments/utils";

export function TeamLobbyJoin({ inviteCode }: { inviteCode: string }) {
  const router = useRouter();
  const queryClient = useQueryClient();
  const { authorizationHeader, isLoading: authLoading } = useAuth();
  const lobbyPath = `/lobbies/${inviteCode}`;
  const {
    data: lobby,
    isPending,
    error,
    isFetching,
  } = $api.useQuery(
    "get",
    "/api/v1/lobbies/{inviteCode}",
    {
      params: { path: { inviteCode } },
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
  const myTeamQuery = useQuery({
    queryKey: [
      "get",
      "/api/v1/tournaments/{tournamentId}/my-team",
      lobby?.tournament.id,
      authorizationHeader,
    ],
    enabled: Boolean(authorizationHeader && lobby?.tournament.id && !alreadyJoined),
    retry: false,
    staleTime: 0,
    queryFn: async () => {
      const result = await apiClient.GET(
        "/api/v1/tournaments/{tournamentId}/my-team",
        {
          params: { path: { tournamentId: lobby!.tournament.id } },
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
      queryKey: ["get", "/api/v1/lobbies/{inviteCode}"],
    });
    router.replace(lobbyPath);
  }

  if (authLoading || isPending || (alreadyJoined && !error)) {
    return <LobbyWorkspaceSkeleton />;
  }

  const unavailable =
    !lobby ||
    lobby.status !== "FORMING" ||
    !isRegistrationOpen(lobby.tournament) ||
    (lobby.members?.length ?? 0) >= lobby.tournament.maxTeamSize;

  return (
    <div className="container mx-auto flex w-full max-w-3xl flex-1 flex-col gap-6 px-4 py-12 sm:px-8">
      <Button
        variant="ghost"
        size="sm"
        render={<Link href={lobbyPath} />}
        nativeButton={false}
        className="self-start"
      >
        <ArrowLeft data-icon="inline-start" />
        Back to lobby
      </Button>
      {myTeamQuery.data?.status === "REJECTED" ? (
        <Alert variant="destructive">
          <AlertTitle>Application rejected</AlertTitle>
          <AlertDescription>
            Your application was not accepted for this tournament, so you cannot
            join another team.
          </AlertDescription>
        </Alert>
      ) : error || unavailable ? (
        <Alert>
          <AlertTitle>Unable to join this team</AlertTitle>
          <AlertDescription>
            {error
              ? "The lobby could not be loaded. Return to the lobby and try again."
              : "This team is full or is no longer accepting new members."}
          </AlertDescription>
        </Alert>
      ) : (
        <>
          <div className="flex flex-col gap-2">
            <h1 className="text-3xl font-bold tracking-tight">
              Join {lobby.name}
            </h1>
            <p className="text-muted-foreground">
              Complete your registration for {lobby.tournament.name}. You will
              return to the lobby after submitting.
            </p>
          </div>
          <MemberRegistrationForm lobby={lobby} onSaved={returnToLobby} />
        </>
      )}
    </div>
  );
}
