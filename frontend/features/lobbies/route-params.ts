import { notFound } from "next/navigation";
import { lobbyInviteParamSchema, teamIdParamSchema } from "@/features/lobbies/schemas";
import { tournamentIdParamSchema } from "@/features/tournaments/schemas";

export async function parseTeamLobbyRouteParams(
  params: Promise<{ tournamentId: string; teamId: string }>,
  searchParams: Promise<{ invite?: string | string[] }>
) {
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

  return {
    tournamentId: parsedTournament.data,
    teamId: parsedTeam.data,
    inviteCode: parsedInvite?.success ? parsedInvite.data : undefined,
  };
}
