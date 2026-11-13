import { formatMinorMoney } from "@/lib/money";
import { extractProblemMessage } from "@/lib/schemas";
import type { components } from "@/lib/api/schema";

export type Lobby = components["schemas"]["TeamLobbyResponse"];

const dateFormatter = new Intl.DateTimeFormat("en-GB", {
  dateStyle: "medium",
  timeZone: "Asia/Bangkok",
});

export function formatDate(value: string) {
  return dateFormatter.format(new Date(value));
}

export function formatEntryFee(amountMinor: number, currency: string) {
  return formatMinorMoney(amountMinor, currency, "Free");
}

export function problemMessage(error: unknown, fallback: string) {
  return extractProblemMessage(error, fallback);
}
