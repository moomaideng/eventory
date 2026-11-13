export const DEFAULT_CURRENCY = "THB";

/**
 * Maximum whole-unit value whose minor-unit representation (* 100) stays within
 * Number.MAX_SAFE_INTEGER (9,007,199,254,740,991).
 */
export const MAX_WHOLE_MONEY = Math.floor(Number.MAX_SAFE_INTEGER / 100);

const currencyFormatters = new Map<string, Intl.NumberFormat>();

export function normalizeCurrency(currency?: string | null): string {
  const trimmed = currency?.trim().toUpperCase();
  return trimmed || DEFAULT_CURRENCY;
}

/** Converts a user-facing major amount (e.g. THB) into integer minor units (satang). */
export function wholeToMinor(whole: number): number {
  return Math.round(whole * 100);
}

/** Converts integer minor units (satang) into a major-unit number (THB) for form inputs. */
export function minorToWhole(amountMinor: number): number {
  return amountMinor / 100;
}

/**
 * Formats an integer minor-unit amount (satang, cents) in standardized display format
 * (e.g. "THB 1,250.00"). Returns `zeroLabel` when provided and `amountMinor === 0`.
 */
export function formatMinorMoney(
  amountMinor: number,
  currency?: string | null,
  zeroLabel?: string
): string {
  if (zeroLabel !== undefined && amountMinor === 0) {
    return zeroLabel;
  }
  const safeCurrency = normalizeCurrency(currency);
  const minor = BigInt(Math.trunc(amountMinor));
  const whole = minor / BigInt(100);
  const absFraction = (minor < BigInt(0) ? -minor : minor) % BigInt(100);
  const fraction = absFraction.toString().padStart(2, "0");

  let formatter = currencyFormatters.get(safeCurrency);
  if (!formatter) {
    try {
      formatter = new Intl.NumberFormat("en-GB", {
        style: "currency",
        currency: safeCurrency,
        minimumFractionDigits: 2,
        maximumFractionDigits: 2,
      });
      currencyFormatters.set(safeCurrency, formatter);
    } catch {
      return `${whole.toLocaleString("en-GB")}.${fraction} ${safeCurrency}`;
    }
  }
  return formatter
    .formatToParts(whole)
    .map((part) => (part.type === "fraction" ? fraction : part.value))
    .join("");
}
