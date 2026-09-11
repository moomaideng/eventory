import {
  CheckCheck,
  LockKeyhole,
  UserRoundPlus,
  Users,
} from "lucide-react";
import type { OrganizerSummary } from "@/features/organizer/utils";
import { formatRegistrationType } from "@/features/tournaments/utils";
import { TournamentAboutCard } from "@/features/tournaments/components/tournament-about-card";
import { TournamentFundingCard } from "@/features/tournaments/components/tournament-funding-card";

export function DashboardOverview({ summary }: { summary: OrganizerSummary }) {
  const { tournament, metrics, funding } = summary;
  const entryUnit =
    tournament.registrationMode === "TEAM"
      ? "team entries"
      : tournament.registrationMode === "SOLO"
        ? "solo entries"
        : "entries";
  const organizerMetrics = [
    {
      label: "Accepted entries",
      value: `${metrics.acceptedEntries} / ${tournament.capacity}`,
      detail: entryUnit,
      icon: CheckCheck,
    },
    {
      label: "Confirmed participants",
      value: metrics.confirmedParticipants,
      detail: "On accepted entries",
      icon: Users,
    },
    {
      label: "Locked entries",
      value: metrics.lockedEntries,
      detail: "Roster locked",
      icon: LockKeyhole,
    },
    {
      label: "Available spots",
      value: metrics.availableSpots,
      detail: entryUnit,
      icon: UserRoundPlus,
    },
  ];

  return (
    <div className="flex flex-col gap-8">
      <dl className="grid grid-cols-2 gap-3 lg:grid-cols-4">
        {organizerMetrics.map(({ label, value, detail, icon: Icon }) => (
          <div
            key={label}
            className="bg-card flex min-w-0 flex-col gap-3 rounded-lg border p-4 shadow-xs sm:p-5"
          >
            <div className="bg-muted text-primary flex size-9 items-center justify-center rounded-md">
              <Icon className="size-4" aria-hidden="true" />
            </div>
            <dt className="text-muted-foreground min-h-10 text-sm">{label}</dt>
            <dd className="flex flex-col gap-1">
              <span className="text-2xl font-semibold tabular-nums">
                {value}
              </span>
              <span className="text-muted-foreground text-xs">{detail}</span>
            </dd>
          </div>
        ))}
      </dl>

      <div className="grid items-start gap-6 lg:grid-cols-[minmax(0,1fr)_22rem]">
        <TournamentAboutCard
          tournament={tournament}
          spotsLeft={metrics.availableSpots}
          registrationType={formatRegistrationType(tournament.registrationMode)}
        />
        <div className="lg:sticky lg:top-24">
          <TournamentFundingCard funding={funding} />
        </div>
      </div>
    </div>
  );
}
