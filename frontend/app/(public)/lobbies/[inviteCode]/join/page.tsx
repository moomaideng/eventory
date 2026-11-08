import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { TeamLobbyJoin } from "@/features/lobbies/components/team-lobby-join";
import { lobbyInviteParamSchema } from "@/features/lobbies/schemas";

export const metadata: Metadata = {
  title: "Join team - Eventory",
};

export default async function JoinTeamPage({
  params,
}: {
  params: Promise<{ inviteCode: string }>;
}) {
  const { inviteCode } = await params;
  const parsed = lobbyInviteParamSchema.safeParse(inviteCode);
  if (!parsed.success) notFound();
  return <TeamLobbyJoin inviteCode={parsed.data} />;
}
