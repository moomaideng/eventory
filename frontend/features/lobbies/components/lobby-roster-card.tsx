"use client";

import React from "react";
import { UserMinus, UsersRound } from "lucide-react";
import { Badge } from "@/components/ui/badge";
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
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { formatDate, type Lobby } from "../utils";

interface LobbyRosterCardProps {
  lobby: Lobby;
  pendingAction: string | null;
  onJoinLobby: () => void;
  onRemoveMember: (memberId: string) => void;
}

export function LobbyRosterCard({
  lobby,
  pendingAction,
  onJoinLobby,
  onRemoveMember,
}: LobbyRosterCardProps) {
  const members = lobby.members ?? [];
  const isCaptain = lobby.viewerRole === "CAPTAIN";
  const isInvitee = lobby.viewerRole === "INVITEE";
  const isForming = lobby.status === "FORMING";
  const rosterReady = members.length >= lobby.tournament.minTeamSize;
  const rosterFull = members.length >= lobby.tournament.maxTeamSize;

  return (
    <Card>
      <CardHeader>
        <div className="flex flex-wrap items-start justify-between gap-3">
          <div className="flex flex-col gap-1">
            <CardTitle>Team roster</CardTitle>
            <CardDescription>
              {members.length} of {lobby.tournament.maxTeamSize} players · Minimum{" "}
              {lobby.tournament.minTeamSize}
            </CardDescription>
          </div>
          <Badge variant={rosterReady ? "secondary" : "outline"}>
            {rosterReady
              ? "Ready to lock"
              : `${lobby.tournament.minTeamSize - members.length} more needed`}
          </Badge>
        </div>
      </CardHeader>
      <CardContent>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Member</TableHead>
              <TableHead>Joined</TableHead>
              {isCaptain && isForming ? (
                <TableHead className="text-right">Actions</TableHead>
              ) : null}
            </TableRow>
          </TableHeader>
          <TableBody>
            {members.map((member) => {
              const isMemberCaptain = member.role === "CAPTAIN";
              return (
                <TableRow key={member.id}>
                  <TableCell className="font-medium">
                    <div className="flex items-center gap-2">
                      <span>{member.displayName}</span>
                      <span className="text-muted-foreground text-xs">
                        @{member.handle}
                      </span>
                      {isMemberCaptain ? (
                        <Badge variant="outline">Captain</Badge>
                      ) : null}
                    </div>
                  </TableCell>
                  <TableCell className="text-muted-foreground">
                    {formatDate(member.joinedAt)}
                  </TableCell>
                  {isCaptain && isForming ? (
                    <TableCell className="text-right">
                      {!isMemberCaptain ? (
                        <Button
                          variant="ghost"
                          size="sm"
                          disabled={pendingAction === `remove-${member.id}`}
                          onClick={() => onRemoveMember(member.id)}
                        >
                          {pendingAction === `remove-${member.id}` ? (
                            <Spinner data-icon="inline-start" />
                          ) : (
                            <UserMinus data-icon="inline-start" />
                          )}
                          Remove
                        </Button>
                      ) : null}
                    </TableCell>
                  ) : null}
                </TableRow>
              );
            })}
          </TableBody>
        </Table>
      </CardContent>
      {isInvitee ? (
        <CardFooter className="flex flex-wrap items-center justify-between gap-3">
          <p className="text-muted-foreground text-sm">
            {isForming && !rosterFull
              ? "Join this roster to compete with the team."
              : "This roster cannot accept new members."}
          </p>
          <Button
            disabled={!isForming || rosterFull || pendingAction === "join"}
            onClick={onJoinLobby}
          >
            {pendingAction === "join" ? (
              <Spinner data-icon="inline-start" />
            ) : (
              <UsersRound data-icon="inline-start" />
            )}
            Join team
          </Button>
        </CardFooter>
      ) : null}
    </Card>
  );
}
