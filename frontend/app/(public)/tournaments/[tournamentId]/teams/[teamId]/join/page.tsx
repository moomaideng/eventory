import type { Metadata } from "next";
import { notFound, redirect } from "next/navigation";
import { TeamLobbyJoin } from "@/features/lobbies/components/team-lobby-join";
import {
  lobbyInviteParamSchema,
  teamIdParamSchema,
} from "@/features/lobbies/schemas";
import { tournamentIdParamSchema } from "@/features/tournaments/schemas";

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
  if (!parsedInvite?.success) {
    redirect(
      `/tournaments/${parsedTournament.data}/teams/${parsedTeam.data}`
    );
  }

  return (
    <TeamLobbyJoin
      tournamentId={parsedTournament.data}
      teamId={parsedTeam.data}
      inviteCode={parsedInvite.data}
    />
  );
}
