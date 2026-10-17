"use client";

import { useQueryClient } from "@tanstack/react-query";
import { $api } from "@/lib/api/client";
import { useAuth } from "@/context/auth-context";
import type { MatchResponse } from "../utils";
import type { BracketMatchNode, Stage1SeedAssignment, TeamPlacementInput } from "../schemas";

export function useBracket({ tournamentId }: { tournamentId: string }) {
  const { authorizationHeader } = useAuth();
  const queryClient = useQueryClient();

  const bracketQuery = $api.useQuery(
    "get",
    "/api/v1/tournaments/{id}/bracket",
    {
      params: { path: { id: tournamentId } },
      headers: authorizationHeader ? { Authorization: authorizationHeader } : undefined,
    },
    {
      enabled: Boolean(tournamentId),
      retry: 1,
      refetchInterval: 15_000,
    }
  );

  const dashboardQuery = $api.useQuery(
    "get",
    "/api/v1/tournaments/{id}/dashboard",
    {
      params: { path: { id: tournamentId } },
      headers: authorizationHeader ? { Authorization: authorizationHeader } : undefined,
    },
    {
      enabled: Boolean(tournamentId && authorizationHeader),
      retry: 1,
    }
  );

  const acceptedCount = dashboardQuery.data?.summary?.metrics?.acceptedEntries ?? 0;

  const defineMutation = $api.useMutation("post", "/api/v1/tournaments/{id}/bracket");
  const seedMutation = $api.useMutation("post", "/api/v1/tournaments/{id}/seed");
  const recordMutation = $api.useMutation(
    "post",
    "/api/v1/tournaments/{id}/matches/{matchId}/results"
  );

  const matches: MatchResponse[] = bracketQuery.data?.matches ?? [];

  async function handleDefineBracket(nodes: BracketMatchNode[]) {
    const payload = nodes.map((n) => ({
      tempId: n.tempId,
      name: n.name,
      venue: n.venue,
      stageLevel: n.stageLevel,
      teamCount: n.teamCount,
      advancingCount: n.advancingCount,
      targetTempId: n.targetTempId ?? undefined,
    }));

    const res = await defineMutation.mutateAsync({
      params: { path: { id: tournamentId } },
      headers: authorizationHeader ? { Authorization: authorizationHeader } : undefined,
      body: { matches: payload },
    });

    await queryClient.invalidateQueries();
    return res.matches;
  }

  async function handleSeedStage1(assignments?: Stage1SeedAssignment[]) {
    const res = await seedMutation.mutateAsync({
      params: { path: { id: tournamentId } },
      headers: authorizationHeader ? { Authorization: authorizationHeader } : undefined,
      body: assignments ? { assignments } : {},
    });

    await queryClient.invalidateQueries();
    return res.matches;
  }

  async function handleRecordResults(
    matchId: string,
    placements: TeamPlacementInput[]
  ) {
    await recordMutation.mutateAsync({
      params: { path: { id: tournamentId, matchId } },
      headers: authorizationHeader ? { Authorization: authorizationHeader } : undefined,
      body: { placements },
    });

    await queryClient.invalidateQueries();
  }

  return {
    matches,
    acceptedCount,
    isLoading: bracketQuery.isLoading,
    isFetching: bracketQuery.isFetching,
    error: bracketQuery.error,
    refetch: bracketQuery.refetch,
    defineBracket: handleDefineBracket,
    isDefining: defineMutation.isPending,
    seedStage1: handleSeedStage1,
    isSeeding: seedMutation.isPending,
    recordResults: handleRecordResults,
    isRecording: recordMutation.isPending,
  };
}
