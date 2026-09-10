"use client";

import React from "react";
import { Check, ClipboardCopy, RefreshCw } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Spinner } from "@/components/ui/spinner";
import type { Lobby } from "../utils";

interface LobbyInviteCardProps {
  lobby: Lobby;
  pendingAction: string | null;
  copied: boolean;
  onCopyInvite: () => void;
  onRegenerateInvite: () => void;
}

export function LobbyInviteCard({
  lobby,
  pendingAction,
  copied,
  onCopyInvite,
  onRegenerateInvite,
}: LobbyInviteCardProps) {
  const isCaptain = lobby.viewerRole === "CAPTAIN";
  const isForming = lobby.status === "FORMING";

  return (
    <Card>
      <CardHeader>
        <CardTitle>Invite teammates</CardTitle>
        <CardDescription>
          Share this link or six-character code while the roster is forming.
        </CardDescription>
      </CardHeader>
      <CardContent className="flex flex-col gap-3">
        <div className="bg-muted rounded-lg p-3">
          <p className="text-muted-foreground text-xs font-medium">Invite code</p>
          <p className="font-mono text-xl font-semibold tracking-[0.2em]">
            {lobby.inviteCode}
          </p>
        </div>
        <Button
          variant="outline"
          disabled={pendingAction === "copy"}
          onClick={onCopyInvite}
        >
          {copied ? (
            <Check data-icon="inline-start" />
          ) : (
            <ClipboardCopy data-icon="inline-start" />
          )}
          {copied ? "Invite link copied" : "Copy invite link"}
        </Button>
      </CardContent>
      {isCaptain && isForming ? (
        <CardFooter>
          <Button
            variant="outline"
            className="w-full"
            disabled={pendingAction === "regenerate"}
            onClick={onRegenerateInvite}
          >
            {pendingAction === "regenerate" ? (
              <Spinner data-icon="inline-start" />
            ) : (
              <RefreshCw data-icon="inline-start" />
            )}
            Regenerate code
          </Button>
        </CardFooter>
      ) : null}
    </Card>
  );
}
