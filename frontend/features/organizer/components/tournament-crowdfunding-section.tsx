"use client";

import React from "react";
import { Coins, Lock, Pencil } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
  formatMoney,
  formatPercentage,
  formatTournamentDate,
} from "@/features/tournaments/utils";
import type { OrganizerSummary } from "@/features/organizer/utils";

export function TournamentCrowdfundingSection({
  summary,
  canEdit,
  onEdit,
}: {
  summary: OrganizerSummary;
  canEdit: boolean;
  onEdit: () => void;
}) {
  const { funding } = summary;

  return (
    <section className="space-y-4">
      <div className="border-border flex items-center justify-between border-b pb-2.5">
        <div className="flex items-center gap-2">
          <Coins className="text-primary size-5" />
          <h2 className="text-lg font-semibold tracking-tight">
            Crowdfunding & Prize Pool
          </h2>
        </div>
        {canEdit ? (
          <Button
            variant="ghost"
            size="sm"
            onClick={onEdit}
            aria-label="Edit Crowdfunding Settings"
            className="text-muted-foreground hover:text-foreground"
          >
            <Pencil data-icon="inline-start" />
            Edit
          </Button>
        ) : (
          <Badge variant="outline" className="text-muted-foreground gap-1">
            <Lock className="size-3" /> Locked
          </Badge>
        )}
      </div>

      <dl className="grid grid-cols-1 gap-x-6 gap-y-5 sm:grid-cols-2 lg:grid-cols-4">
        <div>
          <dt className="text-muted-foreground text-xs font-medium uppercase tracking-wider">
            Funding Goal
          </dt>
          <dd className="mt-1 text-sm font-medium">
            {formatMoney(funding.goalAmount, funding.currency)}
          </dd>
        </div>

        <div>
          <dt className="text-muted-foreground text-xs font-medium uppercase tracking-wider">
            Campaign Deadline
          </dt>
          <dd className="mt-1 text-sm font-medium">
            {funding.deadline ? formatTournamentDate(funding.deadline) : "Not set"}
          </dd>
        </div>

        <div>
          <dt className="text-muted-foreground text-xs font-medium uppercase tracking-wider">
            Raised Amount
          </dt>
          <dd className="mt-1 text-sm font-medium">
            {formatMoney(funding.raisedAmount, funding.currency)} (
            {formatPercentage(funding.percentage)})
          </dd>
        </div>

        <div>
          <dt className="text-muted-foreground text-xs font-medium uppercase tracking-wider">
            Supporters Count
          </dt>
          <dd className="mt-1 text-sm font-medium">
            {funding.supporterCount} backers
          </dd>
        </div>
      </dl>
    </section>
  );
}
