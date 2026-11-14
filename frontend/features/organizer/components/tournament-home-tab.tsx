"use client";

import React, { useState } from "react";
import { TournamentStatusTimeline } from "./tournament-status-timeline";
import { TournamentDetailsCard } from "./tournament-details-card";
import {
  EditTournamentDialog,
} from "./edit-tournament-dialog";
import type { DetailsField } from "@/features/organizer/tournament-details-form";
import {
  formatRelative,
  type OrganizerSummary,
} from "@/features/organizer/utils";
import type { WorkspaceTab } from "@/features/organizer/workspace-tabs";
import { formatMoney } from "@/features/tournaments/utils";

export function TournamentHomeTab({
  tournamentId,
  summary,
  isOwner,
  onRefresh,
  onSelectTab,
}: {
  tournamentId: string;
  summary: OrganizerSummary;
  isOwner: boolean;
  onRefresh: () => void;
  onSelectTab: (tab: WorkspaceTab) => void;
}) {
  const [editOpen, setEditOpen] = useState(false);
  const [editField, setEditField] = useState<DetailsField>("name");

  function openEditor(field: DetailsField = "name") {
    setEditField(field);
    setEditOpen(true);
  }

  return (
    <div className="flex flex-col gap-6">
      <TournamentStatusTimeline
        tournamentId={tournamentId}
        summary={summary}
        isOwner={isOwner}
        onSuccess={onRefresh}
        onEditField={openEditor}
        onSelectTab={onSelectTab}
      />

      <TournamentStats summary={summary} onSelectTab={onSelectTab} />

      <TournamentDetailsCard
        summary={summary}
        isOwner={isOwner}
        onEdit={openEditor}
      />

      <EditTournamentDialog
        key={editOpen ? `open-${editField}` : "closed"}
        tournamentId={tournamentId}
        summary={summary}
        open={editOpen}
        onOpenChange={setEditOpen}
        initialField={editField}
        onSuccess={onRefresh}
      />
    </div>
  );
}

interface Stat {
  label: string;
  value: string;
  detail?: string | null;
  onClick?: () => void;
}

function TournamentStats({
  summary,
  onSelectTab,
}: {
  summary: OrganizerSummary;
  onSelectTab: (tab: WorkspaceTab) => void;
}) {
  const [now] = useState(() => Date.now());
  const { tournament, metrics, funding } = summary;
  const status = tournament.status;
  if (status === "DRAFT") return null;

  const stats: Stat[] = [];
  const fundingStat: Stat = {
    label: "Funding raised",
    value: `${formatMoney(funding.raisedAmountMinor, funding.currency)} / ${formatMoney(funding.goalAmountMinor, funding.currency)}`,
    detail: `${funding.supporterCount} supporter${funding.supporterCount === 1 ? "" : "s"}`,
  };

  if (status === "CROWDFUNDING") {
    stats.push(fundingStat, {
      label: "Funding closes",
      value: formatRelative(funding.deadline, now) ?? "Not set",
    });
  } else {
    const unresolved =
      metrics.totalEntries - metrics.acceptedEntries - metrics.rejectedEntries;
    stats.push({
      label: "Accepted",
      value: `${metrics.acceptedEntries} / ${tournament.capacity}`,
      detail: `${metrics.availableSpots} spot${metrics.availableSpots === 1 ? "" : "s"} left`,
    });
    if (status === "REGISTRATION") {
      stats.push(
        {
          label: "Awaiting review",
          value: String(metrics.lockedEntries),
          detail:
            unresolved > metrics.lockedEntries
              ? `${unresolved - metrics.lockedEntries} still forming or paying`
              : null,
          onClick: () => onSelectTab("registration"),
        },
        {
          label: "Registration closes",
          value: formatRelative(tournament.registrationDeadline, now) ?? "Not set",
        }
      );
    }
    if (funding.goalAmountMinor > 0) stats.push(fundingStat);
  }
  if (status !== "COMPLETED") {
    stats.push({
      label: status === "ONGOING" ? "Ends" : "Starts",
      value:
        formatRelative(
          status === "ONGOING" ? tournament.endAt : tournament.startAt,
          now
        ) ?? "Not set",
    });
  }

  return (
    <div className="grid grid-cols-2 gap-3 lg:grid-cols-4">
      {stats.map((stat) => {
        const body = (
          <>
            <p className="text-muted-foreground text-xs font-medium">
              {stat.label}
            </p>
            <p className="mt-1 text-lg font-semibold tabular-nums">
              {stat.value}
            </p>
            {stat.detail ? (
              <p className="text-muted-foreground text-xs">{stat.detail}</p>
            ) : null}
          </>
        );
        return stat.onClick ? (
          <button
            key={stat.label}
            type="button"
            onClick={stat.onClick}
            className="bg-card hover:bg-muted/50 focus-visible:ring-ring rounded-xl border p-4 text-left transition-colors outline-none focus-visible:ring-2"
          >
            {body}
          </button>
        ) : (
          <div key={stat.label} className="bg-card rounded-xl border p-4">
            {body}
          </div>
        );
      })}
    </div>
  );
}
