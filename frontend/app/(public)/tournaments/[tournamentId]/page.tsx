import { TournamentDetails } from "./tournament-details";

export default async function TournamentDetailsPage({
  params,
}: {
  params: Promise<{ tournamentId: string }>;
}) {
  const { tournamentId } = await params;

  return <TournamentDetails tournamentId={tournamentId} />;
}
