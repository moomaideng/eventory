import { AcceptApplicationDialog } from "@/features/organizer/components/accept-application-dialog";
import { RejectApplicationDialog } from "@/features/organizer/components/reject-application-dialog";
import type { DashboardEntry } from "@/features/organizer/utils";

/**
 * Review decisions for participant applications.
 *
 * NOTE(design): Review is a one-time decision on LOCKED applications only.
 * Once an entry is ACCEPTED or REJECTED, the decision is final
 * and no further review action buttons are presented.
 */
export function ReviewActions({
  tournamentId,
  entry,
}: {
  tournamentId: string;
  entry: Pick<DashboardEntry, "id" | "name" | "status">;
}) {
  if (entry.status !== "LOCKED") return null;

  return (
    <div className="flex flex-wrap gap-2">
      <AcceptApplicationDialog
        tournamentId={tournamentId}
        teamId={entry.id}
        teamName={entry.name}
      />
      <RejectApplicationDialog
        tournamentId={tournamentId}
        teamId={entry.id}
        teamName={entry.name}
      />
    </div>
  );
}
