import assert from "node:assert/strict";
import test from "node:test";
import { getRedirectDestinationLabel } from "../../lib/auth-redirect.ts";
import { lobbyCodeSchema, teamIdParamSchema } from "./schemas.ts";

test("teamIdParamSchema validates 8-character alphanumeric team IDs without changing case", () => {
  const parsed = teamIdParamSchema.safeParse(" Ab12Cd34 ");
  assert.equal(parsed.success, true);
  assert.equal(parsed.data, "Ab12Cd34");

  for (const invalid of [
    "short",
    "toolong12",
    "ab12-d34",
    "11111111-1111-4111-8111-111111111111",
  ]) {
    assert.equal(teamIdParamSchema.safeParse(invalid).success, false);
  }
});

test("lobbyCodeSchema validates and uppercases 6-character invite codes", () => {
  const parsed = lobbyCodeSchema.safeParse(" ab12cd ");
  assert.equal(parsed.success, true);
  assert.equal(parsed.data, "AB12CD");

  for (const invalid of ["AB12C", "AB12CDE", "AB12-D"]) {
    assert.equal(lobbyCodeSchema.safeParse(invalid).success, false);
  }
});

test("getRedirectDestinationLabel recognizes tournament team routes", () => {
  const tourId = "11111111-1111-4111-8111-111111111111";
  assert.equal(
    getRedirectDestinationLabel(`/tournaments/${tourId}/teams/new`),
    "Team Registration"
  );
  assert.equal(
    getRedirectDestinationLabel(`/tournaments/${tourId}/teams/Ab12Cd34`),
    "Team Lobby"
  );
  assert.equal(
    getRedirectDestinationLabel(`/tournaments/${tourId}/teams/Ab12Cd34/join?invite=ABC123`),
    "Join Team"
  );
});
