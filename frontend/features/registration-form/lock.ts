/** The form schema can only be changed while the tournament is DRAFT or CROWDFUNDING. */
export function isFormEditable(status: string) {
  return status === "DRAFT" || status === "CROWDFUNDING";
}

/** Explains why the form is locked; players start answering once registration opens. */
export function formLockReason(status: string) {
  return status === "REGISTRATION"
    ? "Registration is open and players are already answering these questions. Changing them would invalidate the answers players have submitted."
    : "Registration has closed. These questions are final and match what players answered.";
}
