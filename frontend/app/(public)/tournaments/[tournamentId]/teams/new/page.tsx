import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { TeamLobbyCreate } from "@/features/tournaments/components/team-lobby-create";
import { tournamentIdParamSchema } from "@/features/tournaments/schemas";

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
  const parsed = tournamentIdParamSchema.safeParse(tournamentId);
  if (!parsed.success) {
    notFound();
  }

  return <TeamLobbyCreate tournamentId={parsed.data} />;
}
