"use client";

import { useEffect } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { ShieldAlert } from "lucide-react";
import { $api } from "@/lib/api/client";
import { useAuth } from "@/context/auth-context";
import { Button } from "@/components/ui/button";
import {
  Empty,
  EmptyContent,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@/components/ui/empty";
import { LobbyWorkspaceSkeleton } from "@/features/lobbies/components/lobby-skeleton";

export function LobbyInviteRedirect({ inviteCode }: { inviteCode: string }) {
  const router = useRouter();
  const { user, isLoading: isAuthLoading, authorizationHeader } = useAuth();
  const normalizedInviteCode = inviteCode.trim().toUpperCase();

  const { data, error, isPending } = $api.useQuery(
    "get",
    "/api/v1/lobbies/{inviteCode}",
    {
      params: { path: { inviteCode: normalizedInviteCode } },
      headers: authorizationHeader
        ? { Authorization: authorizationHeader }
        : {},
    },
    {
      enabled: Boolean(authorizationHeader),
      retry: false,
      staleTime: 0,
    }
  );

  useEffect(() => {
    if (data?.tournamentId && data.teamId) {
      router.replace(
        `/tournaments/${data.tournamentId}/teams/${data.teamId}?invite=${normalizedInviteCode}`
      );
    }
  }, [data, normalizedInviteCode, router]);

  if (isAuthLoading || isPending || !user || (data && !error)) {
    return <LobbyWorkspaceSkeleton />;
  }

  return (
    <div className="container mx-auto flex w-full max-w-3xl flex-1 px-4 py-12 sm:px-8">
      <Empty className="border">
        <EmptyHeader>
          <EmptyMedia variant="icon">
            <ShieldAlert />
          </EmptyMedia>
          <EmptyTitle>Invite unavailable</EmptyTitle>
          <EmptyDescription>
            This invite may be invalid, revoked, or linked to a disbanded team.
          </EmptyDescription>
        </EmptyHeader>
        <EmptyContent>
          <Button variant="outline" render={<Link href="/lobbies" />} nativeButton={false}>
            Enter another code
          </Button>
        </EmptyContent>
      </Empty>
    </div>
  );
}
