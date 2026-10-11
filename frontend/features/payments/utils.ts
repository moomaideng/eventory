import type { components } from "@/lib/api/schema";

export type Payment = components["schemas"]["PaymentResponse"];

export function formatAmount(
  payment: Pick<Payment, "amountMinor" | "currency">
) {
  const minor = BigInt(payment.amountMinor);
  const fraction = (minor % BigInt(100)).toString().padStart(2, "0");
  // Keep cents exact even at the API's maximum safe integer amount.
  return new Intl.NumberFormat("en-GB", {
    style: "currency",
    currency: payment.currency,
  })
    .formatToParts(minor / BigInt(100))
    .map((part) => (part.type === "fraction" ? fraction : part.value))
    .join("");
}

export function formatPaymentDate(value: string) {
  return new Intl.DateTimeFormat("en-GB", {
    dateStyle: "medium",
    timeStyle: "short",
    timeZone: "UTC",
  }).format(new Date(value));
}
