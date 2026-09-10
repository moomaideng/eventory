"use client";

import React from "react";
import {
  CalendarDays,
  LockKeyhole,
  MapPin,
  ShieldAlert,
  Trash2,
} from "lucide-react";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
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
import { formatDate, formatEntryFee, type Lobby } from "../utils";

interface LobbyDetailsCardProps {
  lobby: Lobby;
  pendingAction: string | null;
  confirmDisband: boolean;
  onSetConfirmDisband: (confirm: boolean) => void;
  onLockRoster: () => void;
  onDisbandLobby: () => void;
}

export function LobbyDetailsCard({
  lobby,
  pendingAction,
  confirmDisband,
  onSetConfirmDisband,
  onLockRoster,
  onDisbandLobby,
}: LobbyDetailsCardProps) {
  const members = lobby.members ?? [];
  const isCaptain = lobby.viewerRole === "CAPTAIN";
  const isForming = lobby.status === "FORMING";
  const rosterReady = members.length >= lobby.tournament.minTeamSize;

  return (
    <div className="flex flex-col gap-6">
      <Card>
        <CardHeader>
          <CardTitle>Registration</CardTitle>
          <CardDescription>
            Team entry fee:{" "}
            {formatEntryFee(
              lobby.tournament.entryFee,
              lobby.tournament.currency
            )}
          </CardDescription>
        </CardHeader>
        <CardContent className="text-muted-foreground flex flex-col gap-2 text-sm">
          <span className="flex items-center gap-2">
            <CalendarDays className="size-4 shrink-0" />
            Registration closes {formatDate(lobby.tournament.registrationDeadline)}
          </span>
          <span className="flex items-center gap-2">
            <MapPin className="size-4 shrink-0" />
            {lobby.tournament.location}
          </span>
        </CardContent>
        {isCaptain && isForming ? (
          <CardFooter>
            <Button
              className="w-full"
              disabled={!rosterReady || pendingAction === "lock"}
              onClick={onLockRoster}
            >
              {pendingAction === "lock" ? (
                <Spinner data-icon="inline-start" />
              ) : (
                <LockKeyhole data-icon="inline-start" />
              )}
              Lock roster
            </Button>
          </CardFooter>
        ) : null}
      </Card>

      {isCaptain && isForming ? (
        <Card>
          <CardHeader>
            <CardTitle>Lobby management</CardTitle>
            <CardDescription>
              Disbanding releases every current member while the roster is still
              forming.
            </CardDescription>
          </CardHeader>
          <CardContent>
            {confirmDisband ? (
              <Alert variant="destructive">
                <ShieldAlert />
                <AlertTitle>Disband this team?</AlertTitle>
                <AlertDescription>
                  This permanently removes the forming team and releases its members.
                </AlertDescription>
              </Alert>
            ) : null}
          </CardContent>
          <CardFooter className="flex flex-col items-stretch gap-2">
            {confirmDisband ? (
              <div className="flex justify-end gap-2">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => onSetConfirmDisband(false)}
                >
                  Cancel
                </Button>
                <Button
                  variant="destructive"
                  size="sm"
                  disabled={pendingAction === "disband"}
                  onClick={onDisbandLobby}
                >
                  {pendingAction === "disband" ? (
                    <Spinner data-icon="inline-start" />
                  ) : (
                    <Trash2 data-icon="inline-start" />
                  )}
                  Disband team
                </Button>
              </div>
            ) : (
              <Button
                variant="destructive"
                onClick={() => onSetConfirmDisband(true)}
              >
                <Trash2 data-icon="inline-start" />
                Disband team
              </Button>
            )}
          </CardFooter>
        </Card>
      ) : null}
    </div>
  );
}
