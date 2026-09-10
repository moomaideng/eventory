import type { Metadata } from "next";
import { TournamentDashboard } from "@/features/organizer/components/tournament-dashboard";

export const metadata: Metadata = { title: "Tournament Dashboard - Eventory" };

export default async function TournamentDashboardPage({
  params,
}: {
  params: Promise<{ tournamentId: string }>;
}) {
  const { tournamentId } = await params;
  return <TournamentDashboard tournamentId={tournamentId} />;
}
