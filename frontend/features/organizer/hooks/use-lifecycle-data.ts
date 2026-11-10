"use client";

import { useEffect, useState } from "react";
import { $api } from "@/lib/api/client";
import { useAuth } from "@/context/auth-context";
import type { Remote } from "@/features/organizer/lifecycle";
import type { components } from "@/lib/api/schema";

type Match = components["schemas"]["MatchResponse"];
type Tier = components["schemas"]["SponsorshipTierResponse"];

/**
 * Loads what the next-stage checklist needs beyond the dashboard summary:
 * sponsorship tiers while in DRAFT (owner only) and the bracket while
 * REGISTRATION or ONGOING. Also keeps a minute-resolution clock so deadline
 * checks flip without a reload.
 */
export function useLifecycleData({
  tournamentId,
  status,
  isOwner,
}: {
  tournamentId: string;
  status: string;
  isOwner: boolean;
}) {
  const { authorizationHeader } = useAuth();
  const [now, setNow] = useState(() => Date.now());

  useEffect(() => {
    const id = window.setInterval(() => setNow(Date.now()), 60_000);
    return () => window.clearInterval(id);
  }, []);

  const tiersEnabled = status === "DRAFT" && isOwner && Boolean(authorizationHeader);
  const tiersQuery = $api.useQuery(
    "get",
    "/api/v1/tournaments/{tournamentId}/sponsorship-tiers/manage",
    {
      params: { path: { tournamentId } },
      headers: { Authorization: authorizationHeader ?? "" },
    },
    { enabled: tiersEnabled, retry: false }
  );

  const bracketEnabled =
    (status === "REGISTRATION" || status === "ONGOING") &&
    Boolean(authorizationHeader);
  const bracketQuery = $api.useQuery(
    "get",
    "/api/v1/tournaments/{id}/bracket",
    {
      params: { path: { id: tournamentId } },
      headers: authorizationHeader ? { Authorization: authorizationHeader } : undefined,
    },
    { enabled: bracketEnabled, retry: 1, refetchInterval: 30_000 }
  );

  const tiers: Remote<Tier> = !isOwner
    ? null
    : tiersQuery.isError
      ? null
      : tiersQuery.data
        ? (tiersQuery.data.items ?? [])
        : undefined;

  const matches: Remote<Match> = bracketQuery.isError
    ? null
    : bracketQuery.data
      ? (bracketQuery.data.matches ?? [])
      : undefined;

  return { now, tiers, matches };
}
