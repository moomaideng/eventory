// The status can remain OPEN after the deadline; both rules must hold.
export function registrationUnavailableReason(tournament: {
  status: string;
  registrationDeadline: string;
}) {
  if (tournament.status !== "REGISTRATION_OPEN") {
    return "Registration is not currently open for this tournament.";
  }
  if (!(Date.parse(tournament.registrationDeadline) > Date.now())) {
    return "The registration deadline has passed. Ask the organizer to update it before registering.";
  }
  return "";
}
