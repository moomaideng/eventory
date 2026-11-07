import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { TournamentAnnouncements } from "@/features/announcements";
import { tournamentIdParamSchema } from "@/features/tournaments/schemas";

export async function generateMetadata({
  params,
}: {
  params: Promise<{ tournamentId: string }>;
}): Promise<Metadata> {
  const { tournamentId } = await params;
  return {
    title: `Tournament Announcements - Eventory`,
    description: `View official updates and announcements for tournament ${tournamentId}.`,
  };
}

export default async function TournamentAnnouncementsPage({
  params,
}: {
  params: Promise<{ tournamentId: string }>;
}) {
  const { tournamentId } = await params;
  const parsed = tournamentIdParamSchema.safeParse(tournamentId);
  if (!parsed.success) {
    notFound();
  }

  return <TournamentAnnouncements tournamentId={parsed.data} />;
}
