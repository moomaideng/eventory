"use client";

import {
  Field,
  FieldDescription,
  FieldError,
  FieldLabel,
} from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import type { RegistrationQuestion as Question } from "@/features/registration/schemas";

const fileTypeLabels: Record<string, string> = {
  "application/pdf": "PDF",
  "image/png": "PNG",
  "image/jpeg": "JPEG",
};

export function RegistrationQuestion({
  question: q,
  value,
  error,
  onChange,
  onFileChange,
}: {
  question: Question;
  value: string;
  error?: string;
  onChange: (value: string) => void;
  onFileChange: (file?: File) => void;
}) {
  const id = `registration-${q.id}`;
  return (
    <Field data-invalid={Boolean(error)}>
      <FieldLabel htmlFor={id}>
        {q.label}
        {q.required ? " *" : " (optional)"}
      </FieldLabel>
      {q.type === "DROPDOWN" ? (
        <Select
          items={[
            { label: "Choose an option", value: null },
            ...(q.options ?? []).map((option) => ({
              label: option,
              value: option,
            })),
          ]}
          value={value || null}
          onValueChange={(v) => onChange(v ?? "")}
        >
          <SelectTrigger
            id={id}
            className="w-full"
            aria-invalid={Boolean(error)}
          >
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectGroup>
              {(q.options ?? []).map((option) => (
                <SelectItem key={option} value={option}>
                  {option}
                </SelectItem>
              ))}
            </SelectGroup>
          </SelectContent>
        </Select>
      ) : q.type === "FILE" ? (
        <Input
          id={id}
          type="file"
          accept={(q.allowedTypes ?? []).join(",")}
          aria-invalid={Boolean(error)}
          onChange={(e) => onFileChange(e.target.files?.[0])}
        />
      ) : (
        <Input
          id={id}
          value={value}
          maxLength={2000}
          aria-invalid={Boolean(error)}
          onChange={(e) => onChange(e.target.value)}
        />
      )}
      {q.helpText ? <FieldDescription>{q.helpText}</FieldDescription> : null}
      {q.type === "FILE" ? (
        <FieldDescription>
          {(q.allowedTypes ?? [])
            .map((type) => fileTypeLabels[type] ?? type)
            .join(", ")}{" "}
          · Up to{" "}
          {q.maxFileBytes < 1024 * 1024
            ? `${Math.ceil(q.maxFileBytes / 1024)} KB`
            : `${(q.maxFileBytes / 1024 / 1024).toFixed(1)} MB`}
        </FieldDescription>
      ) : null}
      {error ? <FieldError>{error}</FieldError> : null}
    </Field>
  );
}
