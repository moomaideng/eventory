import React from "react";
import { Badge } from "@/components/ui/badge";
import type { Lobby } from "@/features/lobbies/utils";

const STATUS_LABELS: Record<Lobby["status"], string> = {
  FORMING: "Forming roster",
  LOCKED: "Waiting for approval",
  PAYMENT_PENDING: "Awaiting payment",
  ACCEPTED: "Registration accepted",
  REJECTED: "Registration rejected",
};

export function LobbyStatusBadge({ status }: { status: Lobby["status"] }) {
  return <Badge variant="outline">{STATUS_LABELS[status] ?? status}</Badge>;
}
