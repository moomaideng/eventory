import type { Metadata } from "next";
import { TeamLobbyWorkspace } from "@/features/lobbies/components/team-lobby-workspace";
import { parseTeamLobbyRouteParams } from "@/features/lobbies/route-params";
import { teamIdParamSchema } from "@/features/lobbies/schemas";

export async function generateMetadata({
  params,
}: {
  params: Promise<{ tournamentId: string; teamId: string }>;
}): Promise<Metadata> {
  const { teamId } = await params;
  const parsedTeam = teamIdParamSchema.safeParse(teamId);
  const label = parsedTeam.success ? parsedTeam.data : teamId.trim();
  return {
    title: `Team ${label} - Eventory`,
    description: `View and manage team roster for team ${label}.`,
  };
}

export default async function TeamLobbyPage({
  params,
  searchParams,
}: {
  params: Promise<{ tournamentId: string; teamId: string }>;
  searchParams: Promise<{ invite?: string | string[] }>;
}) {
  const resolved = await parseTeamLobbyRouteParams(params, searchParams);
  return <TeamLobbyWorkspace {...resolved} />;
}
