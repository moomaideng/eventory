import type { FormField } from "@/features/registration-form/schemas";

/** Text, number and single-choice answers are strings; multi-choice answers are arrays. */
export type AnswerValue = string | string[];

export function emptyAnswer(field: FormField): AnswerValue {
  return field.type === "SELECT_MULTIPLE" ? [] : "";
}

const REQUIRED_MESSAGE = "This is a required question";

/** Returns an error message for the answer, or null when it is valid. */
export function validateAnswer(field: FormField, value: AnswerValue | undefined): string | null {
  switch (field.type) {
    case "SHORT_TEXT":
    case "LONG_TEXT": {
      const text = typeof value === "string" ? value.trim() : "";
      if (!text) return field.required ? REQUIRED_MESSAGE : null;
      const { minLength, maxLength } = field.config;
      if (minLength !== undefined && text.length < minLength) {
        return `Must be at least ${minLength} characters`;
      }
      if (maxLength !== undefined && text.length > maxLength) {
        return `Must be at most ${maxLength} characters`;
      }
      return null;
    }
    case "NUMBER": {
      const raw = typeof value === "string" ? value.trim() : "";
      if (!raw) return field.required ? REQUIRED_MESSAGE : null;
      const num = Number(raw);
      if (!Number.isFinite(num)) return "Enter a valid number";
      const { min, max, allowDecimals } = field.config;
      if (!allowDecimals && !Number.isInteger(num)) return "Must be a whole number";
      if (min !== undefined && num < min) return `Must be at least ${min}`;
      if (max !== undefined && num > max) return `Must be at most ${max}`;
      return null;
    }
    case "SELECT_ONE": {
      const picked = typeof value === "string" ? value : "";
      return !picked && field.required ? REQUIRED_MESSAGE : null;
    }
    case "SELECT_MULTIPLE": {
      const picked = Array.isArray(value) ? value : [];
      if (picked.length === 0) return field.required ? REQUIRED_MESSAGE : null;
      const { minSelections, maxSelections } = field.config;
      if (minSelections !== undefined && picked.length < minSelections) {
        return `Select at least ${minSelections}`;
      }
      if (maxSelections !== undefined && picked.length > maxSelections) {
        return `Select at most ${maxSelections}`;
      }
      return null;
    }
  }
}
