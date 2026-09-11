import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { TournamentDashboard } from "@/features/organizer/components/tournament-dashboard";
import { tournamentIdParamSchema } from "@/features/tournaments/schemas";

export const metadata: Metadata = {
  title: "Tournament Dashboard - Eventory",
  description:
    "View registrations, participant rosters, and funding metrics for your tournament.",
};

export default async function TournamentDashboardPage({
  params,
}: {
  params: Promise<{ tournamentId: string }>;
}) {
  const { tournamentId } = await params;
  const parsed = tournamentIdParamSchema.safeParse(tournamentId);
  if (!parsed.success) {
    notFound();
  }

  return <TournamentDashboard tournamentId={parsed.data} />;
}
