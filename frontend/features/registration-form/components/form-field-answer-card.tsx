"use client";

import React from "react";
import { Badge } from "@/components/ui/badge";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group";
import { Textarea } from "@/components/ui/textarea";
import { cn } from "@/lib/utils";
import type { AnswerValue } from "@/features/registration-form/utils";
import type { FormField } from "@/features/registration-form/schemas";

export interface FormFieldAnswerCardProps {
  field: FormField;
  index: number;
  value: AnswerValue;
  error?: string | null;
  onChange: (value: AnswerValue) => void;
  onBlur?: () => void;
}

interface FieldInputProps {
  field: FormField;
  value: AnswerValue;
  invalid: boolean;
  onChange: (value: AnswerValue) => void;
  onBlur?: () => void;
}

function FieldInput({ field, value, invalid, onChange, onBlur }: FieldInputProps) {
  const inputId = `ans-${field.id}`;
  const text = typeof value === "string" ? value : "";

  switch (field.type) {
    case "SHORT_TEXT":
      return (
        <Input
          id={inputId}
          value={text}
          maxLength={field.config.maxLength}
          aria-invalid={invalid}
          onChange={(e) => onChange(e.target.value)}
          onBlur={onBlur}
          placeholder="Your answer"
        />
      );
    case "LONG_TEXT":
      return (
        <Textarea
          id={inputId}
          value={text}
          maxLength={field.config.maxLength}
          rows={4}
          aria-invalid={invalid}
          onChange={(e) => onChange(e.target.value)}
          onBlur={onBlur}
          placeholder="Your answer"
        />
      );
    case "NUMBER":
      return (
        <Input
          id={inputId}
          type="number"
          value={text}
          min={field.config.min}
          max={field.config.max}
          step={field.config.allowDecimals ? "any" : 1}
          aria-invalid={invalid}
          onChange={(e) => onChange(e.target.value)}
          onBlur={onBlur}
          placeholder="Your answer"
        />
      );
    case "SELECT_ONE":
      return (
        <RadioGroup
          className="gap-2"
          value={text}
          onValueChange={(val) => onChange(String(val))}
        >
          {field.config.options.map((opt) => (
            <div key={opt.id} className="flex items-center gap-2">
              <RadioGroupItem id={`${inputId}-${opt.id}`} value={opt.id} />
              <Label htmlFor={`${inputId}-${opt.id}`} className="cursor-pointer font-normal">
                {opt.label}
              </Label>
            </div>
          ))}
        </RadioGroup>
      );
    case "SELECT_MULTIPLE": {
      const picked = Array.isArray(value) ? value : [];
      const max = field.config.maxSelections;
      const atLimit = max !== undefined && picked.length >= max;
      return (
        <div className="flex flex-col gap-2">
          {field.config.options.map((opt) => {
            const checked = picked.includes(opt.id);
            const blocked = atLimit && !checked;
            return (
              <div key={opt.id} className="flex items-center gap-2">
                <input
                  type="checkbox"
                  id={`${inputId}-${opt.id}`}
                  checked={checked}
                  disabled={blocked}
                  onChange={(e) =>
                    onChange(
                      e.target.checked
                        ? [...picked, opt.id]
                        : picked.filter((id) => id !== opt.id)
                    )
                  }
                  className="size-4 cursor-pointer rounded border-input accent-primary disabled:cursor-not-allowed disabled:opacity-50"
                />
                <Label
                  htmlFor={`${inputId}-${opt.id}`}
                  className={cn("cursor-pointer font-normal", blocked && "opacity-50")}
                >
                  {opt.label}
                </Label>
              </div>
            );
          })}
        </div>
      );
    }
  }
}

function constraintHint(field: FormField): string | null {
  if (field.type !== "SELECT_MULTIPLE") return null;
  const { minSelections, maxSelections } = field.config;
  if (minSelections !== undefined && maxSelections !== undefined) {
    return `Select ${minSelections} to ${maxSelections}`;
  }
  if (maxSelections !== undefined) return `Select up to ${maxSelections}`;
  if (minSelections !== undefined) return `Select at least ${minSelections}`;
  return null;
}

export function FormFieldAnswerCard({
  field,
  index,
  value,
  error,
  onChange,
  onBlur,
}: FormFieldAnswerCardProps) {
  const hint = constraintHint(field);

  return (
    <Card
      id={`field-card-${field.id}`}
      className={cn("border-border", error && "border-destructive")}
    >
      <CardHeader className="flex flex-row items-center gap-3 pb-0">
        <Badge variant="outline" className="shrink-0 font-mono">
          #{index + 1}
        </Badge>
        <CardTitle className="min-w-0 font-heading text-base font-medium leading-normal text-card-foreground md:text-base">
          {field.label}
          {field.required && (
            <span className="ml-1 text-destructive" aria-label="required">
              *
            </span>
          )}
        </CardTitle>
      </CardHeader>
      <CardContent className="flex flex-col gap-3">
        {field.description && (
          <p className="text-sm text-muted-foreground">{field.description}</p>
        )}
        {hint && <p className="text-xs text-muted-foreground">{hint}</p>}
        <FieldInput
          field={field}
          value={value}
          invalid={Boolean(error)}
          onChange={onChange}
          onBlur={onBlur}
        />
        {error && (
          <p role="alert" className="text-sm font-medium text-destructive">
            {error}
          </p>
        )}
      </CardContent>
    </Card>
  );
}
