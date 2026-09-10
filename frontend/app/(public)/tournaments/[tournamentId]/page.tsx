import type { Metadata } from "next";
import { TournamentDetails } from "@/features/tournaments/components/tournament-details";

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
  return <TournamentDetails tournamentId={tournamentId} />;
}
