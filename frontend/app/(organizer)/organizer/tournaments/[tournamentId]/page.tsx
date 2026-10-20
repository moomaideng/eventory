import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { TournamentWorkspace } from "@/features/organizer/components/tournament-workspace";
import { tournamentIdParamSchema } from "@/features/tournaments/schemas";

export const metadata: Metadata = {
  title: "Tournament Workspace - Eventory",
  description:
    "View registrations, participant rosters, configure registration questions, and manage staff for your tournament.",
};

export default async function TournamentWorkspacePage({
  params,
}: {
  params: Promise<{ tournamentId: string }>;
}) {
  const { tournamentId } = await params;
  const parsed = tournamentIdParamSchema.safeParse(tournamentId);
  if (!parsed.success) {
    notFound();
  }

  return <TournamentWorkspace tournamentId={parsed.data} />;
}
