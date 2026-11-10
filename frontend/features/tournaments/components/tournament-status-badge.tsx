"use client";

import React from "react";
import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";
import { formatTournamentStatus } from "../utils";

export interface TournamentStatusBadgeProps {
  status: string;
  registrationDeadline?: string | null;
  className?: string;
}

export function TournamentStatusBadge({
  status,
  registrationDeadline,
  className,
}: TournamentStatusBadgeProps) {
  const [now] = React.useState(() => Date.now());

  const isClosed =
    status === "REGISTRATION" &&
    Boolean(
      registrationDeadline &&
        new Date(registrationDeadline).getTime() <= now
    );

  let variant: "default" | "secondary" | "outline" = "outline";
  let modifierClass = "";

  if (status === "DRAFT") {
    variant = "outline";
    modifierClass = "border-dashed text-muted-foreground";
  } else if (status === "CROWDFUNDING") {
    variant = "secondary";
  } else if (status === "REGISTRATION") {
    if (isClosed) {
      variant = "outline";
      modifierClass = "text-muted-foreground";
    } else {
      variant = "default";
    }
  } else if (status === "REGISTRATION_OPEN") {
    variant = "default";
  } else if (status === "REGISTRATION_CLOSED") {
    variant = "outline";
    modifierClass = "text-muted-foreground";
  } else if (status === "ONGOING") {
    variant = "default";
  } else if (status === "COMPLETED") {
    variant = "outline";
    modifierClass = "text-muted-foreground";
  }

  const label = formatTournamentStatus(status, registrationDeadline);

  return (
    <Badge
      variant={variant}
      className={cn("capitalize", modifierClass, className)}
    >
      {label}
    </Badge>
  );
}
