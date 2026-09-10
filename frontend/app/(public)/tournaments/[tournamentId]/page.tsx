import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { TournamentDetails } from "@/features/tournaments/components/tournament-details";
import { tournamentIdParamSchema } from "@/features/tournaments/schemas";

export async function generateMetadata({
  params,
}: {
  params: Promise<{ tournamentId: string }>;
}): Promise<Metadata> {
  const { tournamentId } = await params;
  return {
    title: `Tournament Details - Eventory`,
    description: `View tournament schedule, funding progress, rules, and registered teams for tournament ${tournamentId}.`,
  };
}

export default async function TournamentDetailsPage({
  params,
}: {
  params: Promise<{ tournamentId: string }>;
}) {
  const { tournamentId } = await params;
  const parsed = tournamentIdParamSchema.safeParse(tournamentId);
  if (!parsed.success) {
    notFound();
  }

  return <TournamentDetails tournamentId={parsed.data} />;
}
