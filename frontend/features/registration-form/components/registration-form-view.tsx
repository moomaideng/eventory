"use client";

import React, { useMemo, useState } from "react";
import { useAuth } from "@/context/auth-context";
import { $api } from "@/lib/api/client";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Spinner } from "@/components/ui/spinner";
import { FormFieldAnswerCard } from "@/features/registration-form/components/form-field-answer-card";
import {
  emptyAnswer,
  validateAnswer,
  type AnswerValue,
} from "@/features/registration-form/utils";
import { formFieldSchema, type FormField } from "@/features/registration-form/schemas";
import type { components } from "@/lib/api/schema";

type ApiFormField = components["schemas"]["FormField"];

function toFormField(f: ApiFormField, i: number): FormField {
  const candidate = {
    id: f.id,
    type: f.type,
    label: f.label,
    description: f.description ?? "",
    required: f.required ?? false,
    order: f.order ?? i + 1,
    config: typeof f.config === "object" && f.config !== null ? f.config : {},
  };
  const parsed = formFieldSchema.safeParse(candidate);
  if (parsed.success) return parsed.data;
  return { ...candidate, type: "SHORT_TEXT", config: {} };
}

export function RegistrationFormView({ tournamentId }: { tournamentId: string }) {
  const { authorizationHeader, isLoading: authLoading } = useAuth();

  const { data, isLoading, isError } = $api.useQuery(
    "get",
    "/api/v1/tournaments/{tournamentId}/form",
    {
      params: { path: { tournamentId } },
      headers: { Authorization: authorizationHeader ?? "" },
    },
    {
      enabled: Boolean(authorizationHeader),
      retry: false,
    }
  );

  const fields = useMemo(
    () =>
      (data?.fields ?? [])
        .map(toFormField)
        .sort((a, b) => a.order - b.order),
    [data]
  );

  const [answers, setAnswers] = useState<Record<string, AnswerValue>>({});
  const [errors, setErrors] = useState<Record<string, string | null>>({});

  const answerOf = (field: FormField) => answers[field.id] ?? emptyAnswer(field);

  const setError = (field: FormField, value: AnswerValue) =>
    setErrors((prev) => ({ ...prev, [field.id]: validateAnswer(field, value) }));

  const handleChange = (field: FormField, value: AnswerValue) => {
    setAnswers((prev) => ({ ...prev, [field.id]: value }));
    // Re-check live only once the question has shown an error, so it clears as soon as it is fixed.
    if (errors[field.id]) setError(field, value);
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    const next: Record<string, string | null> = {};
    for (const field of fields) {
      next[field.id] = validateAnswer(field, answerOf(field));
    }
    setErrors(next);
    const firstInvalid = fields.find((f) => next[f.id]);
    if (firstInvalid) {
      document
        .getElementById(`field-card-${firstInvalid.id}`)
        ?.scrollIntoView({ behavior: "smooth", block: "center" });
    }
    // Submitting the answers is not wired up yet.
  };

  if (authLoading || isLoading) {
    return (
      <div className="flex flex-1 items-center justify-center p-12">
        <Spinner className="size-6 text-muted-foreground" />
      </div>
    );
  }

  if (isError || !data) {
    return (
      <div className="mx-auto w-full max-w-3xl px-4 py-8">
        <Card>
          <CardContent className="pt-6 text-center text-muted-foreground">
            This tournament has no registration form yet.
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <form
      noValidate
      onSubmit={handleSubmit}
      className="mx-auto flex w-full max-w-3xl flex-col gap-6 px-4 py-8"
    >
      <Card className="border-border border-t-4 border-t-primary shadow-xs">
        <CardContent className="flex flex-col gap-3 pt-6">
          <h1 className="font-heading text-2xl font-bold leading-normal tracking-tight text-foreground">
            {data.title}
          </h1>
          {data.description && (
            <p className="whitespace-pre-wrap text-sm text-muted-foreground">
              {data.description}
            </p>
          )}
        </CardContent>
      </Card>

      <div className="flex flex-col gap-6">
        {fields.map((field, idx) => (
          <FormFieldAnswerCard
            key={field.id}
            field={field}
            index={idx}
            value={answerOf(field)}
            error={errors[field.id]}
            onChange={(value) => handleChange(field, value)}
            onBlur={() => setError(field, answerOf(field))}
          />
        ))}
      </div>

      <div className="flex justify-end">
        <Button type="submit">Submit</Button>
      </div>
    </form>
  );
}
