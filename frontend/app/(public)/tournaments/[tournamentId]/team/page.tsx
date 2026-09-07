import { TeamLobbyCreate } from "./team-lobby-create";

export default async function TeamLobbyCreatePage({
  params,
}: {
  params: Promise<{ tournamentId: string }>;
}) {
  const { tournamentId } = await params;
  return <TeamLobbyCreate tournamentId={tournamentId} />;
}
