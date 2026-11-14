import assert from "node:assert/strict";
import test from "node:test";
import {
  sponsorshipCheckoutRequestSchema,
  sponsorshipCheckoutSchema,
} from "./schemas.ts";
import { checkoutErrorFallback } from "./utils.ts";

test("checkout request requires a UUID requestId", () => {
  assert.equal(
    sponsorshipCheckoutRequestSchema.safeParse({
      requestId: "550e8400-e29b-41d4-a716-446655440000",
    }).success,
    true
  );
  assert.equal(
    sponsorshipCheckoutRequestSchema.safeParse({ requestId: "not-a-uuid" })
      .success,
    false
  );
});

test("checkout rejects a malformed payment response", () => {
  assert.equal(
    sponsorshipCheckoutSchema.safeParse({ status: "PAID" }).success,
    false
  );
});

test("HTTP errors have useful fallbacks without claiming a network failure", () => {
  for (const status of [401, 404, 409, 412, 422, 429, 500, 503]) {
    assert.match(checkoutErrorFallback(status), new RegExp(`HTTP ${status}`));
    assert.doesNotMatch(checkoutErrorFallback(status), /Connection interrupted/);
  }
});
