"use client";

import { useState } from "react";
import { Trash2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Checkbox } from "@/components/ui/checkbox";
import {
  Field,
  FieldDescription,
  FieldError,
  FieldGroup,
  FieldLabel,
  FieldSet,
  FieldLegend,
} from "@/components/ui/field";
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import type { RegistrationConfig } from "@/features/registration/schemas";

const types = [
  { value: "TEXT", label: "Text" },
  { value: "DROPDOWN", label: "Dropdown" },
  { value: "FILE", label: "File upload" },
];
const fileTypes = [
  { value: "application/pdf", label: "PDF" },
  { value: "image/png", label: "PNG" },
  { value: "image/jpeg", label: "JPEG" },
];
type Question = RegistrationConfig["questions"][number];

export function QuestionEditor({
  question: q,
  index,
  errors,
  onChange,
  onRemove,
}: {
  question: Question;
  index: number;
  errors: Record<string, string>;
  onChange: (question: Question) => void;
  onRemove: () => void;
}) {
  const [optionsText, setOptionsText] = useState(q.options.join("\n"));
  const prefix = `question-${q.id}`;
  const patch = (changes: Partial<Question>) => onChange({ ...q, ...changes });
  return (
    <FieldSet className="rounded-lg border p-4">
      <FieldLegend>Question {index + 1}</FieldLegend>
      <FieldGroup>
        <Field data-invalid={Boolean(errors.label)}>
          <FieldLabel htmlFor={`${prefix}-label`}>Question label</FieldLabel>
          <Input
            id={`${prefix}-label`}
            value={q.label}
            maxLength={160}
            aria-invalid={Boolean(errors.label)}
            onChange={(e) => patch({ label: e.target.value })}
            placeholder="e.g. In-game ID"
          />
          {errors.label ? <FieldError>{errors.label}</FieldError> : null}
        </Field>
        <Field>
          <FieldLabel htmlFor={`${prefix}-type`}>Answer type</FieldLabel>
          <Select
            items={types}
            value={q.type}
            onValueChange={(value) => {
              if (!value) return;
              patch({
                type: value as Question["type"],
                pattern: "",
                allowedTypes:
                  value === "FILE"
                    ? ["application/pdf", "image/png", "image/jpeg"]
                    : [],
                maxFileBytes: value === "FILE" ? 5 * 1024 * 1024 : 0,
              });
            }}
          >
            <SelectTrigger id={`${prefix}-type`}>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectGroup>
                {types.map((t) => (
                  <SelectItem key={t.value} value={t.value}>
                    {t.label}
                  </SelectItem>
                ))}
              </SelectGroup>
            </SelectContent>
          </Select>
        </Field>
        <Field orientation="horizontal">
          <Checkbox
            id={`${prefix}-required`}
            checked={q.required}
            onCheckedChange={(required) => patch({ required })}
          />
          <FieldLabel htmlFor={`${prefix}-required`}>
            Required answer
          </FieldLabel>
        </Field>
        <Field>
          <FieldLabel htmlFor={`${prefix}-help`}>
            Help text (optional)
          </FieldLabel>
          <Input
            id={`${prefix}-help`}
            value={q.helpText}
            maxLength={500}
            onChange={(e) => patch({ helpText: e.target.value })}
          />
        </Field>
        {q.type === "TEXT" ? (
          <Field>
            <FieldLabel htmlFor={`${prefix}-pattern`}>
              Validation pattern (optional)
            </FieldLabel>
            <Input
              id={`${prefix}-pattern`}
              value={q.pattern}
              maxLength={256}
              onChange={(e) => patch({ pattern: e.target.value })}
              placeholder="^[A-Za-z0-9_]+$"
            />
            <FieldDescription>
              Uses Go RE2 syntax. Leave blank to accept any text. Use ^ and $ to
              match the whole answer.
            </FieldDescription>
          </Field>
        ) : null}
        {q.type === "DROPDOWN" ? (
          <Field data-invalid={Boolean(errors.options)}>
            <FieldLabel htmlFor={`${prefix}-options`}>
              Dropdown options
            </FieldLabel>
            <Textarea
              id={`${prefix}-options`}
              value={optionsText}
              aria-invalid={Boolean(errors.options)}
              onChange={(e) => {
                setOptionsText(e.target.value);
                patch({
                  options: e.target.value
                    .split("\n")
                    .map((o) => o.trim())
                    .filter(Boolean),
                });
              }}
              placeholder={"Gold\nSilver\nBronze"}
            />
            <FieldDescription>One option per line.</FieldDescription>
            {errors.options ? <FieldError>{errors.options}</FieldError> : null}
          </Field>
        ) : null}
        {q.type === "FILE" ? (
          <>
            <FieldSet>
              <FieldLegend variant="label">Allowed file types</FieldLegend>
              <FieldGroup>
                {fileTypes.map((t) => (
                  <Field
                    key={t.value}
                    orientation="horizontal"
                    data-invalid={Boolean(errors.allowedTypes)}
                  >
                    <Checkbox
                      id={`${prefix}-${t.label}`}
                      checked={q.allowedTypes.includes(
                        t.value as Question["allowedTypes"][number]
                      )}
                      aria-invalid={Boolean(errors.allowedTypes)}
                      onCheckedChange={(checked) =>
                        patch({
                          allowedTypes: checked
                            ? [
                                ...q.allowedTypes,
                                t.value as Question["allowedTypes"][number],
                              ]
                            : q.allowedTypes.filter((v) => v !== t.value),
                        })
                      }
                    />
                    <FieldLabel htmlFor={`${prefix}-${t.label}`}>
                      {t.label}
                    </FieldLabel>
                  </Field>
                ))}
                {errors.allowedTypes ? (
                  <FieldError>{errors.allowedTypes}</FieldError>
                ) : null}
              </FieldGroup>
            </FieldSet>
            <Field data-invalid={Boolean(errors.maxFileBytes)}>
              <FieldLabel htmlFor={`${prefix}-size`}>
                Maximum file size (MB)
              </FieldLabel>
              <Input
                id={`${prefix}-size`}
                type="number"
                min={0.001}
                max={5}
                step="any"
                value={q.maxFileBytes / (1024 * 1024)}
                aria-invalid={Boolean(errors.maxFileBytes)}
                onChange={(e) =>
                  patch({
                    maxFileBytes: Math.round(
                      Number(e.target.value) * 1024 * 1024
                    ),
                  })
                }
              />
              {errors.maxFileBytes ? (
                <FieldError>{errors.maxFileBytes}</FieldError>
              ) : null}
            </Field>
          </>
        ) : null}
        <Button
          type="button"
          variant="outline"
          onClick={onRemove}
          className="self-end"
        >
          <Trash2 data-icon="inline-start" />
          Remove question {index + 1}
        </Button>
      </FieldGroup>
    </FieldSet>
  );
}
