import React from "react";
import { Badge } from "@/components/ui/badge";
import type { Lobby } from "../utils";

const STATUS_LABELS: Record<Lobby["status"], string> = {
  FORMING: "Forming roster",
  LOCKED: "Roster locked",
  ACCEPTED: "Registration accepted",
  REJECTED: "Registration rejected",
};

export function LobbyStatusBadge({ status }: { status: Lobby["status"] }) {
  return <Badge variant="outline">{STATUS_LABELS[status] ?? status}</Badge>;
}
