import type { Metadata } from "next";
import { redirect } from "next/navigation";
import { TeamLobbyJoin } from "@/features/lobbies/components/team-lobby-join";
import { parseTeamLobbyRouteParams } from "@/features/lobbies/route-params";

export const metadata: Metadata = {
  title: "Join team - Eventory",
};

export default async function JoinTeamPage({
  params,
  searchParams,
}: {
  params: Promise<{ tournamentId: string; teamId: string }>;
  searchParams: Promise<{ invite?: string | string[] }>;
}) {
  const { tournamentId, teamId, inviteCode } = await parseTeamLobbyRouteParams(
    params,
    searchParams
  );
  if (!inviteCode) {
    redirect(`/tournaments/${tournamentId}/teams/${teamId}`);
  }

  return (
    <TeamLobbyJoin
      tournamentId={tournamentId}
      teamId={teamId}
      inviteCode={inviteCode}
    />
  );
}
