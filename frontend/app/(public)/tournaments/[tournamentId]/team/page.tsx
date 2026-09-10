import type { Metadata } from "next";
import { TeamLobbyCreate } from "@/features/tournaments/components/team-lobby-create";

export async function generateMetadata({
  params,
}: {
  params: Promise<{ tournamentId: string }>;
}): Promise<Metadata> {
  const { tournamentId } = await params;
  return {
    title: `Create Team - Eventory`,
    description: `Create a team lobby and invite teammates for tournament ${tournamentId}.`,
  };
}

export default async function TeamLobbyCreatePage({
  params,
}: {
  params: Promise<{ tournamentId: string }>;
}) {
  const { tournamentId } = await params;
  return <TeamLobbyCreate tournamentId={tournamentId} />;
}
