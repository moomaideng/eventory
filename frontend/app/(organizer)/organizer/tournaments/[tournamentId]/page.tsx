import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { TournamentWorkspace } from "@/features/organizer/components/tournament-workspace";
import { isWorkspaceTab } from "@/features/organizer/workspace-tabs";
import { tournamentIdParamSchema } from "@/features/tournaments/schemas";

export const metadata: Metadata = {
  title: "Tournament Workspace - Eventory",
  description:
    "View registrations, participant rosters, configure registration questions, and manage staff for your tournament.",
};

export default async function TournamentWorkspacePage({
  params,
  searchParams,
}: {
  params: Promise<{ tournamentId: string }>;
  searchParams: Promise<{ tab?: string | string[] }>;
}) {
  const [{ tournamentId }, { tab }] = await Promise.all([params, searchParams]);
  const parsed = tournamentIdParamSchema.safeParse(tournamentId);
  if (!parsed.success) {
    notFound();
  }

  return (
    <TournamentWorkspace
      tournamentId={parsed.data}
      initialTab={isWorkspaceTab(tab) ? tab : "home"}
    />
  );
}
