import { TeamLobbyWorkspace } from "./team-lobby-workspace";

export default async function TeamLobbyPage({
  params,
}: {
  params: Promise<{ inviteCode: string }>;
}) {
  const { inviteCode } = await params;
  return <TeamLobbyWorkspace inviteCode={inviteCode} />;
}
