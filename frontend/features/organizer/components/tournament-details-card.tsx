"use client";

import type { ReactNode } from "react";
import { Lock, Pencil, Plus } from "lucide-react";
import { Button } from "@/components/ui/button";
import { MarkdownContent } from "@/components/ui/markdown-content";
import { Progress } from "@/components/ui/progress";
import type { DetailsSection } from "@/features/organizer/components/edit-tournament-dialog";
import type { DetailsField } from "@/features/organizer/tournament-details-form";
import {
  SECTION_LOCK_REASON,
  isSectionEditable,
} from "@/features/organizer/tournament-details-form";
import {
  formatRelative,
  type OrganizerSummary,
} from "@/features/organizer/utils";
import {
  formatMoney,
  formatPercentage,
  formatTournamentDate,
} from "@/features/tournaments/utils";

function DetailGroup({
  section,
  title,
  status,
  children,
}: {
  section: DetailsSection;
  title: string;
  status: string;
  children: ReactNode;
}) {
  const locked = !isSectionEditable(section, status);
  return (
    <section className="flex flex-col gap-4">
      <div className="flex flex-wrap items-center justify-between gap-x-3 gap-y-1 border-b pb-2">
        <h3 className="text-muted-foreground text-xs font-semibold tracking-wider uppercase">
          {title}
        </h3>
        {locked ? (
          <span className="text-muted-foreground inline-flex items-center gap-1 text-xs">
            <Lock className="size-3" aria-hidden />
            {SECTION_LOCK_REASON[section]}
          </span>
        ) : null}
      </div>
      <dl className="flex flex-col gap-4">{children}</dl>
    </section>
  );
}

function DetailItem({
  label,
  children,
}: {
  label: string;
  children: ReactNode;
}) {
  return (
    <div className="flex flex-col gap-1">
      <dt className="text-muted-foreground text-xs">{label}</dt>
      <dd className="text-sm font-medium wrap-anywhere">{children}</dd>
    </div>
  );
}

function DateValue({ value }: { value?: string | null }) {
  if (!value) return <span className="text-muted-foreground">Not set</span>;
  const relative = formatRelative(value);
  return (
    <>
      {formatTournamentDate(value)}
      {relative ? (
        <span className="text-muted-foreground font-normal"> · {relative}</span>
      ) : null}
    </>
  );
}

export function TournamentDetailsCard({
  summary,
  isOwner,
  onEdit,
}: {
  summary: OrganizerSummary;
  isOwner: boolean;
  onEdit: (field?: DetailsField) => void;
}) {
  const { tournament, funding } = summary;
  const status = tournament.status;
  const canEdit = isOwner && status !== "COMPLETED";
  const isSolo = tournament.registrationMode === "SOLO";

  const format = isSolo
    ? "Solo"
    : `Team · ${
        tournament.minTeamSize === tournament.maxTeamSize
          ? tournament.minTeamSize
          : `${tournament.minTeamSize}–${tournament.maxTeamSize}`
      } players per team`;

  const showFundingGroup =
    status === "DRAFT" || status === "CROWDFUNDING" || funding.goalAmount > 0;
  const goal =
    funding.goalAmount > 0 ? (
      formatMoney(funding.goalAmount, funding.currency)
    ) : (
      <span className="text-muted-foreground">Not set</span>
    );

  return (
    <section className="bg-card rounded-xl border">
      <div className="flex items-center justify-between gap-4 border-b px-5 py-4">
        <h2 className="text-base font-semibold">Tournament details</h2>
        {canEdit ? (
          <Button variant="outline" size="sm" onClick={() => onEdit()}>
            <Pencil data-icon="inline-start" />
            Edit
          </Button>
        ) : null}
      </div>

      <div className="grid gap-8 p-5 lg:grid-cols-2 lg:gap-10">
        <DetailGroup section="general" title="General" status={status}>
          <DetailItem label="Name">{tournament.name}</DetailItem>
          <DetailItem label="Game">{tournament.game}</DetailItem>
          <DetailItem label="Location">{tournament.location}</DetailItem>
          <DetailItem label="Description">
            <span className="font-normal whitespace-pre-line">
              {tournament.description || (
                <span className="text-muted-foreground">
                  No description yet
                </span>
              )}
            </span>
          </DetailItem>
          <div className="flex flex-col gap-1">
            <dt className="text-muted-foreground text-xs">Rules & details</dt>
            <dd className="text-sm">
              {tournament.detail?.trim() ? (
                <div className="bg-muted/30 max-h-96 overflow-y-auto rounded-lg border p-4">
                  <MarkdownContent content={tournament.detail} />
                </div>
              ) : canEdit ? (
                <Button
                  variant="ghost"
                  size="sm"
                  className="-ml-2.5"
                  onClick={() => onEdit("detail")}
                >
                  <Plus data-icon="inline-start" />
                  Add rules
                </Button>
              ) : (
                <span className="text-muted-foreground">No rules yet</span>
              )}
            </dd>
          </div>
        </DetailGroup>

        <div className="flex flex-col gap-8">
          <DetailGroup
            section="registration"
            title="Registration"
            status={status}
          >
            <DetailItem label="Format">{format}</DetailItem>
            <DetailItem label="Capacity">
              {tournament.capacity > 0 ? (
                `${tournament.capacity} ${isSolo ? "players" : "teams"}`
              ) : (
                <span className="text-muted-foreground">Not set</span>
              )}
            </DetailItem>
            <DetailItem label="Entry fee">
              {formatMoney(tournament.entryFee, tournament.currency, true)}
            </DetailItem>
          </DetailGroup>

          <DetailGroup section="schedule" title="Schedule" status={status}>
            <DetailItem label="Registration closes">
              <DateValue value={tournament.registrationDeadline} />
            </DetailItem>
            <DetailItem label="Starts">
              <DateValue value={tournament.startAt} />
            </DetailItem>
            <DetailItem label="Ends">
              <DateValue value={tournament.endAt} />
            </DetailItem>
          </DetailGroup>

          {showFundingGroup ? (
            <DetailGroup
              section="crowdfunding"
              title="Crowdfunding"
              status={status}
            >
              {status === "DRAFT" ? (
                <>
                  <DetailItem label="Goal">{goal}</DetailItem>
                  <DetailItem label="Deadline">
                    <DateValue value={funding.deadline} />
                  </DetailItem>
                </>
              ) : status === "CROWDFUNDING" ? (
                <>
                  <DetailItem label="Goal">{goal}</DetailItem>
                  <DetailItem label="Raised">
                    <div className="flex flex-col gap-2">
                      <span>
                        {formatMoney(funding.raisedAmount, funding.currency)}
                        <span className="text-muted-foreground font-normal">
                          {" "}
                          · {formatPercentage(funding.percentage)}
                        </span>
                      </span>
                      <Progress
                        value={Math.min(funding.percentage, 100)}
                        aria-label="Funding progress"
                      />
                    </div>
                  </DetailItem>
                  <DetailItem label="Supporters">
                    {funding.supporterCount}
                  </DetailItem>
                  <DetailItem label="Deadline">
                    <DateValue value={funding.deadline} />
                  </DetailItem>
                </>
              ) : (
                <DetailItem label="Result">
                  Raised {formatMoney(funding.raisedAmount, funding.currency)}{" "}
                  of {formatMoney(funding.goalAmount, funding.currency)} ·{" "}
                  {funding.supporterCount}{" "}
                  {funding.supporterCount === 1 ? "supporter" : "supporters"}
                </DetailItem>
              )}
            </DetailGroup>
          ) : null}
        </div>
      </div>
    </section>
  );
}
