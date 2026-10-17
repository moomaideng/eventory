import { notFound } from "next/navigation";
import { TournamentDetails } from "@/features/tournaments/components/tournament-details";
import { tournamentIdParamSchema } from "@/features/tournaments/schemas";

export const metadata = { title: "Sponsor a tournament | Eventory" };

export default async function SponsorTournamentPage({
  params,
}: {
  params: Promise<{ tournamentId: string }>;
}) {
  const parsed = tournamentIdParamSchema.safeParse((await params).tournamentId);
  if (!parsed.success) notFound();
  return <TournamentDetails tournamentId={parsed.data} sponsorMode />;
}
