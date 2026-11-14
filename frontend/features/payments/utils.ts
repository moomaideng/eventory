import type { components } from "@/lib/api/schema";
import { formatMinorMoney } from "@/lib/money";

export type Payment = components["schemas"]["PaymentResponse"];

export function formatAmount(
  payment: { amountMinor: number; currency: string }
) {
  return formatMinorMoney(payment.amountMinor, payment.currency);
}

export function formatPaymentDate(value: string) {
  return new Intl.DateTimeFormat("en-GB", {
    dateStyle: "medium",
    timeStyle: "short",
    timeZone: "UTC",
  }).format(new Date(value));
}
