import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { LobbyInviteRedirect } from "@/features/lobbies/components/lobby-invite-redirect";
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
    title: `Team Invite ${normalized} - Eventory`,
    description: `Resolve invite code ${normalized} to open the team lobby.`,
  };
}

export default async function LobbyInvitePage({
  params,
}: {
  params: Promise<{ inviteCode: string }>;
}) {
  const { inviteCode } = await params;
  const parsed = lobbyInviteParamSchema.safeParse(inviteCode);
  if (!parsed.success) {
    notFound();
  }

  return <LobbyInviteRedirect inviteCode={parsed.data} />;
}
