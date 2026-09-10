import type { components } from "@/lib/api/schema";

export type Lobby = components["schemas"]["TeamLobbyResponse"];

const dateFormatter = new Intl.DateTimeFormat("en-GB", {
  dateStyle: "medium",
  timeZone: "Asia/Bangkok",
});

export function formatDate(value: string) {
  return dateFormatter.format(new Date(value));
}

export function formatEntryFee(amount: number, currency: string) {
  if (amount === 0) return "Free";
  return new Intl.NumberFormat("en-US", {
    style: "currency",
    currency,
    maximumFractionDigits: 0,
  }).format(amount);
}

export function problemMessage(error: unknown, fallback: string) {
  if (
    error &&
    typeof error === "object" &&
    "detail" in error &&
    typeof error.detail === "string"
  ) {
    return error.detail;
  }
  return fallback;
}
