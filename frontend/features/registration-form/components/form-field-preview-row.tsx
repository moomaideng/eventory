import React from "react";
import { Badge } from "@/components/ui/badge";
import { FIELD_TYPE_OPTIONS } from "@/features/registration-form/components/form-field-card";
import {
  LONG_TEXT_MAX_LENGTH,
  SHORT_TEXT_MAX_LENGTH,
  type FormField,
} from "@/features/registration-form/schemas";

function range(min: number | undefined, max: number | undefined, unit: string) {
  const suffix = unit ? ` ${unit}` : "";
  if (min !== undefined && max !== undefined) return `${min} to ${max}${suffix}`;
  if (min !== undefined) return `at least ${min}${suffix}`;
  if (max !== undefined) return `up to ${max}${suffix}`;
  return null;
}

function constraints(field: FormField): string[] {
  const rules: (string | null)[] = [];
  switch (field.type) {
    case "SHORT_TEXT":
      rules.push(range(undefined, SHORT_TEXT_MAX_LENGTH, "characters"));
      break;
    case "LONG_TEXT":
      rules.push(range(undefined, LONG_TEXT_MAX_LENGTH, "characters"));
      break;
    case "NUMBER":
      rules.push(
        range(field.config.min, field.config.max, ""),
        field.config.allowDecimals ? "decimals allowed" : "whole numbers only"
      );
      break;
    case "SELECT_MULTIPLE":
      rules.push(
        range(field.config.minSelections, field.config.maxSelections, "selections")
      );
      break;
    case "SELECT_ONE":
      break;
  }
  return rules.filter((rule): rule is string => rule !== null);
}

/** One compact, read-only question line. Render inside a `<ul>`. */
export function FormFieldPreviewRow({
  field,
  index,
}: {
  field: FormField;
  index: number;
}) {
  const typeLabel =
    FIELD_TYPE_OPTIONS.find((opt) => opt.value === field.type)?.label ?? field.type;
  const rules = constraints(field);
  const options =
    field.type === "SELECT_ONE" || field.type === "SELECT_MULTIPLE"
      ? field.config.options
      : [];
  const meta = [typeLabel, ...rules].join(" · ");

  return (
    <li
      id={`field-${field.id}`}
      className="flex scroll-mt-6 gap-3 py-3 first:pt-0 last:pb-0"
    >
      <span className="w-5 shrink-0 pt-px text-right text-sm tabular-nums text-muted-foreground">
        {index + 1}
      </span>
      <div className="flex min-w-0 flex-1 flex-col gap-1">
        <p className="text-sm font-medium leading-snug">
          {field.label}
          {field.required ? (
            <span className="ml-1 text-destructive" aria-label="required">
              *
            </span>
          ) : null}
        </p>
        {field.description ? (
          <p className="text-sm text-muted-foreground">{field.description}</p>
        ) : null}
        {options.length > 0 ? (
          <div className="flex flex-wrap gap-1.5 pt-0.5">
            {options.map((opt) => (
              <Badge key={opt.id} variant="secondary" className="font-normal">
                {opt.label}
              </Badge>
            ))}
          </div>
        ) : null}
        <p className="text-xs text-muted-foreground">{meta}</p>
      </div>
    </li>
  );
}
