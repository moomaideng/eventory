import { Badge } from "@/components/ui/badge";
import {
  formatRegistrationType,
  formatTournamentStatus,
} from "@/features/tournaments/utils";

export function TournamentHeaderTags({
  game,
  status,
  registrationMode,
  registrationDeadline,
}: {
  game: string;
  status: string;
  registrationMode: string;
  registrationDeadline?: string | null;
}) {
  const formattedStatus = formatTournamentStatus(status, registrationDeadline);
  const registrationType = formatRegistrationType(registrationMode);
  const isCrowdfunding = status === "CROWDFUNDING";

  return (
    <>
      <Badge>{game}</Badge>
      <Badge variant="outline" className="capitalize">
        {formattedStatus}
      </Badge>
      {!isCrowdfunding ? (
        <Badge variant="outline">{registrationType}</Badge>
      ) : null}
    </>
  );
}
