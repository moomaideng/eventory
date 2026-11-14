"use client";

import React from "react";
import { useQueryClient } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import { apiClient } from "@/lib/api/client";
import { problemMessage, type Lobby } from "@/features/lobbies/utils";

export function useLobbyActions({
  lobby,
  authorizationHeader,
  refetch,
}: {
  lobby?: Lobby;
  authorizationHeader?: string | null;
  refetch: () => Promise<unknown>;
}) {
  const router = useRouter();
  const queryClient = useQueryClient();
  const [actionError, setActionError] = React.useState("");
  const [pendingAction, setPendingAction] = React.useState<string | null>(null);

  async function runAction(
    action: string,
    request: Promise<{ data?: unknown; error?: unknown }>,
    afterSuccess?: (data: Lobby | undefined) => void
  ) {
    setActionError("");
    setPendingAction(action);
    try {
      const { data, error: requestError } = await request;
      if (requestError) {
        setActionError(
          problemMessage(requestError, "We could not update this team lobby.")
        );
        await refetch();
        return;
      }
      await queryClient.invalidateQueries({
        queryKey: ["my-registration-answers", lobby?.tournament.id],
      });
      await queryClient.invalidateQueries({
        queryKey: ["my-solo-registration-answers", lobby?.tournament.id],
      });
      afterSuccess?.(data as Lobby | undefined);
      if (!afterSuccess) {
        await refetch();
      }
    } catch {
      setActionError(
        "Connection error. Please check your network and try again."
      );
    } finally {
      setPendingAction(null);
    }
  }

  function regenerateInvite() {
    if (!authorizationHeader || !lobby) return;
    void runAction(
      "regenerate",
      apiClient.POST("/api/v1/tournaments/{tournamentId}/teams/{teamId}/invite/regenerate", {
        params: {
          path: { tournamentId: lobby.tournament.id, teamId: lobby.id },
        },
        headers: { Authorization: authorizationHeader },
      }),
      () => {
        void refetch();
      }
    );
  }

  function submitRoster(openCheckout: boolean) {
    if (!authorizationHeader || !lobby) return;
    void runAction(
      openCheckout ? "pay" : "lock",
      apiClient.POST("/api/v1/tournaments/{tournamentId}/teams/{teamId}/lock", {
        params: {
          path: { tournamentId: lobby.tournament.id, teamId: lobby.id },
        },
        headers: { Authorization: authorizationHeader },
      }),
      openCheckout
        ? (updatedLobby) => {
            void refetch();
            if (updatedLobby?.status === "PAYMENT_PENDING" && updatedLobby.paymentId)
              router.push(`/payments/${updatedLobby.paymentId}/checkout`);
          }
        : undefined
    );
  }

  function lockRoster() {
    submitRoster(false);
  }

  function payRegistration() {
    submitRoster(true);
  }

  function removeMember(memberId: string) {
    if (!authorizationHeader || !lobby) return;
    void runAction(
      `remove-${memberId}`,
      apiClient.DELETE(
        "/api/v1/tournaments/{tournamentId}/teams/{teamId}/members/{memberId}",
        {
          params: {
            path: {
              tournamentId: lobby.tournament.id,
              teamId: lobby.id,
              memberId,
            },
          },
          headers: { Authorization: authorizationHeader },
        }
      )
    );
  }

  function disbandLobby() {
    if (!authorizationHeader || !lobby) return;
    void runAction(
      "disband",
      apiClient.DELETE("/api/v1/tournaments/{tournamentId}/teams/{teamId}", {
        params: {
          path: { tournamentId: lobby.tournament.id, teamId: lobby.id },
        },
        headers: { Authorization: authorizationHeader },
      }),
      () => router.replace(`/tournaments/${lobby.tournament.id}`)
    );
  }

  return {
    actionError,
    pendingAction,
    regenerateInvite,
    lockRoster,
    payRegistration,
    removeMember,
    disbandLobby,
  };
}
