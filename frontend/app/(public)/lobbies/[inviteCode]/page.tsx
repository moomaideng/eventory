import type { Metadata } from "next";
import { TeamLobbyWorkspace } from "@/features/lobbies/components/team-lobby-workspace";

export async function generateMetadata({
  params,
}: {
  params: Promise<{ inviteCode: string }>;
}): Promise<Metadata> {
  const { inviteCode } = await params;
  const normalized = inviteCode.trim().toUpperCase();
  return {
    title: `Team Lobby ${normalized} - Eventory`,
    description: `View and manage team roster for lobby ${normalized}.`,
  };
}

export default async function TeamLobbyPage({
  params,
}: {
  params: Promise<{ inviteCode: string }>;
}) {
  const { inviteCode } = await params;
  return <TeamLobbyWorkspace inviteCode={inviteCode} />;
}
