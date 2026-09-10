"use client";

import React from "react";
import { useRouter } from "next/navigation";
import { apiClient } from "@/lib/api/client";
import { problemMessage, type Lobby } from "../utils";

export function useLobbyActions({
  lobby,
  authorizationHeader,
  normalizedInviteCode,
  refetch,
}: {
  lobby?: Lobby;
  authorizationHeader?: string | null;
  normalizedInviteCode: string;
  refetch: () => Promise<unknown>;
}) {
  const router = useRouter();
  const [actionError, setActionError] = React.useState("");
  const [pendingAction, setPendingAction] = React.useState<string | null>(null);

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

  return {
    actionError,
    pendingAction,
    joinLobby,
    regenerateInvite,
    lockRoster,
    removeMember,
    disbandLobby,
  };
}
