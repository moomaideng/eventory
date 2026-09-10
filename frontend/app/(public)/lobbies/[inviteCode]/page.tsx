import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { TeamLobbyWorkspace } from "@/features/lobbies/components/team-lobby-workspace";
import { lobbyInviteParamSchema } from "@/features/lobbies/schemas";

export async function generateMetadata({
  params,
}: {
  params: Promise<{ inviteCode: string }>;
}): Promise<Metadata> {
  const { inviteCode } = await params;
  const parsed = lobbyInviteParamSchema.safeParse(inviteCode);
  const normalized = parsed.success ? parsed.data : inviteCode.trim().toUpperCase();
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
  const parsed = lobbyInviteParamSchema.safeParse(inviteCode);
  if (!parsed.success) {
    notFound();
  }

  return <TeamLobbyWorkspace inviteCode={parsed.data} />;
}
