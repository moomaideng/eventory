import assert from "node:assert/strict";
import test from "node:test";
import { sponsorshipBrandSchema, sponsorshipCheckoutSchema } from "./schemas.ts";
import { checkoutErrorFallback } from "./utils.ts";

const logo = new File([new Uint8Array([1])], "logo.png", { type: "image/png" });

test("brand validation accepts and trims a website with a supported logo", () => {
  const result = sponsorshipBrandSchema.safeParse({ websiteUrl: " https://brand.example ", logo });
  assert.equal(result.success, true);
  assert.equal(result.data.websiteUrl, "https://brand.example");
});

test("brand validation rejects unsafe links and unsupported or oversized uploads", () => {
  for (const websiteUrl of ["bad url", "javascript:alert(1)", "https://user:pass@brand.example"]) {
    assert.equal(sponsorshipBrandSchema.safeParse({ websiteUrl, logo }).success, false);
  }
  for (const file of [undefined, new File([], "empty.png", { type: "image/png" }), new File(["<svg/>"], "logo.svg", { type: "image/svg+xml" }), new File([new Uint8Array(1024 * 1024 + 1)], "large.png", { type: "image/png" })]) {
    assert.equal(sponsorshipBrandSchema.safeParse({ websiteUrl: "https://brand.example", logo: file }).success, false);
  }
});

test("checkout rejects a malformed payment response", () => {
  assert.equal(sponsorshipCheckoutSchema.safeParse({ status: "PAID" }).success, false);
});

test("HTTP errors have useful fallbacks without claiming a network failure", () => {
  assert.match(checkoutErrorFallback(413), /too large/);
  for (const status of [401, 404, 409, 422, 429, 500, 503]) {
    assert.match(checkoutErrorFallback(status), new RegExp(`HTTP ${status}`));
    assert.doesNotMatch(checkoutErrorFallback(status), /Connection interrupted/);
  }
});
