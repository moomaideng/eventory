import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { TeamLobbyWorkspace } from "@/features/lobbies/components/team-lobby-workspace";
import {
  lobbyInviteParamSchema,
  teamIdParamSchema,
} from "@/features/lobbies/schemas";
import { tournamentIdParamSchema } from "@/features/tournaments/schemas";

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
  const [{ tournamentId, teamId }, { invite }] = await Promise.all([
    params,
    searchParams,
  ]);
  const parsedTournament = tournamentIdParamSchema.safeParse(tournamentId);
  const parsedTeam = teamIdParamSchema.safeParse(teamId);
  if (!parsedTournament.success || !parsedTeam.success) {
    notFound();
  }

  const parsedInvite =
    typeof invite === "string" ? lobbyInviteParamSchema.safeParse(invite) : null;

  return (
    <TeamLobbyWorkspace
      tournamentId={parsedTournament.data}
      teamId={parsedTeam.data}
      inviteCode={parsedInvite?.success ? parsedInvite.data : undefined}
    />
  );
}
