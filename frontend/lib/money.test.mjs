import assert from "node:assert/strict";
import test from "node:test";

import {
  DEFAULT_CURRENCY,
  MAX_WHOLE_MONEY,
  formatMinorMoney,
  minorToWhole,
  normalizeCurrency,
  wholeToMinor,
} from "./money.ts";

test("MAX_WHOLE_MONEY matches backend satang safe bound", () => {
  assert.equal(MAX_WHOLE_MONEY, 90_071_992_547_409);
  assert.equal(MAX_WHOLE_MONEY * 100 <= Number.MAX_SAFE_INTEGER, true);
  assert.equal((MAX_WHOLE_MONEY + 1) * 100 > Number.MAX_SAFE_INTEGER, true);
});

test("wholeToMinor and minorToWhole convert cleanly without float drift", () => {
  assert.equal(wholeToMinor(0), 0);
  assert.equal(wholeToMinor(19.99), 1999);
  assert.equal(wholeToMinor(500), 50000);
  assert.equal(minorToWhole(50000), 500);
  assert.equal(minorToWhole(1999), 19.99);
});

test("normalizeCurrency defaults blank values and uppercases", () => {
  assert.equal(normalizeCurrency(""), DEFAULT_CURRENCY);
  assert.equal(normalizeCurrency("   "), DEFAULT_CURRENCY);
  assert.equal(normalizeCurrency(null), DEFAULT_CURRENCY);
  assert.equal(normalizeCurrency(" thb "), "THB");
});

test("formatMinorMoney formats minor currency, zero labels, and unknown currencies", () => {
  assert.equal(formatMinorMoney(0, "THB", "Free entry"), "Free entry");
  assert.equal(formatMinorMoney(0, "THB", "Free"), "Free");
  assert.equal(formatMinorMoney(50000, "THB", "Free"), "THB\u00a0500.00");
  assert.equal(formatMinorMoney(125050, ""), "THB\u00a01,250.50");
  assert.equal(formatMinorMoney(1, "THB"), "THB\u00a00.01");
  assert.equal(
    formatMinorMoney(9007199254740991, "THB"),
    "THB\u00a090,071,992,547,409.91"
  );
  assert.equal(
    formatMinorMoney(50025, "NOT_A_CURRENCY"),
    "500.25 NOT_A_CURRENCY"
  );
});
