import assert from "node:assert/strict";
import test from "node:test";
import { formatAmount, formatPaymentDate } from "./utils.ts";
import { historyDatesSchema, paymentIdSchema, sponsorPaymentRouteSchema } from "./schemas.ts";

test("sponsor payment routes accept only invoice, checkout and receipt paths", () => {
  const id = "66666666-6666-4666-8666-666666666666";
  for (const path of [[id], [id, "checkout"], [id, "receipt"]]) {
    assert.equal(sponsorPaymentRouteSchema.safeParse(path).success, true);
  }
  for (const path of [[], ["invalid"], [id, "other"], [id, "receipt", "extra"]]) {
    assert.equal(sponsorPaymentRouteSchema.safeParse(path).success, false);
  }
});

test("money display preserves satang, including the largest allowed amount", () => {
  assert.equal(
    formatAmount({ amountMinor: 15000, currency: "THB" }),
    "THB 150.00"
  );
  assert.equal(formatAmount({ amountMinor: 1, currency: "THB" }), "THB 0.01");
  assert.equal(
    formatAmount({ amountMinor: 9007199254740991, currency: "THB" }),
    "THB 90,071,992,547,409.91"
  );
});

test("history uses UTC even when a timestamp contains a different offset", () => {
  assert.equal(
    formatPaymentDate("2026-10-11T00:30:00+07:00"),
    "10 Oct 2026, 17:30"
  );
});

test("date filters reject impossible or reversed dates", () => {
  assert.ok(historyDatesSchema.safeParse({ from: "", to: "" }).success);
  assert.ok(
    historyDatesSchema.safeParse({ from: "2024-02-29", to: "2024-02-29" })
      .success
  );
  assert.equal(
    historyDatesSchema.safeParse({ from: "2025-02-29", to: "" }).success,
    false
  );
  assert.equal(
    historyDatesSchema.safeParse({ from: "2026-10-11", to: "2026-10-10" })
      .success,
    false
  );
  assert.equal(paymentIdSchema.safeParse("not-a-payment-id").success, false);
});
