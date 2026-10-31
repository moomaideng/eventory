"use client";

import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { apiClient } from "@/lib/api/client";
import { formFieldSchema } from "@/features/registration-form/schemas";
import {
  emptyAnswer,
  formatAnswerForApi,
  validateAnswer,
  type AnswerValue,
} from "@/features/registration-form/utils";

export function useRegistrationAnswers(
  tournamentId: string,
  authorizationHeader?: string | null
) {
  const [answers, setAnswers] = useState<Record<string, AnswerValue>>({});
  const [errors, setErrors] = useState<Record<string, string | null>>({});
  const form = useQuery({
    queryKey: ["team-registration-questions", tournamentId],
    enabled: Boolean(tournamentId && authorizationHeader),
    retry: false,
    queryFn: async () => {
      const result = await apiClient.GET(
        "/api/v1/tournaments/{tournamentId}/form",
        { params: { path: { tournamentId } } }
      );
      if (result.response.status === 404) return [];
      if (result.error || !result.data)
        throw new Error("Unable to load registration questions.");
      return (result.data.fields ?? [])
        .map((field, index) => {
          const parsed = formFieldSchema.safeParse({
            ...field,
            order: field.order ?? index + 1,
            config: field.config ?? {},
          });
          if (!parsed.success)
            throw new Error(
              "The registration form contains an unsupported question. Contact the organizer."
            );
          return parsed.data;
        })
        .sort((a, b) => a.order - b.order);
    },
  });
  const saved = useQuery({
    queryKey: ["my-registration-answers", tournamentId, authorizationHeader],
    enabled: Boolean(tournamentId && authorizationHeader),
    retry: false,
    queryFn: async () => {
      const result = await apiClient.GET(
        "/api/v1/tournaments/{tournamentId}/form/responses/me",
        {
          params: { path: { tournamentId } },
          headers: { Authorization: authorizationHeader! },
        }
      );
      if (result.response.status === 404) return [];
      if (result.error || !result.data)
        throw new Error("Unable to load your saved answers.");
      return result.data.answers;
    },
  });
  const fields = form.data ?? [];
  const valueOf = (field: (typeof fields)[number]): AnswerValue => {
    if (answers[field.id] !== undefined) return answers[field.id];
    const previous = saved.data?.find(
      (answer) => answer.fieldId === field.id
    )?.value;
    if (field.type === "SELECT_MULTIPLE" && Array.isArray(previous))
      return previous.filter(
        (item): item is string => typeof item === "string"
      );
    if (typeof previous === "string" || typeof previous === "number")
      return String(previous);
    return emptyAnswer(field);
  };
  const blocked =
    form.isPending || saved.isPending || form.isError || saved.isError;
  function validate() {
    if (blocked) return null;
    const next = Object.fromEntries(
      fields.map((field) => [field.id, validateAnswer(field, valueOf(field))])
    );
    setErrors(next);
    const invalid = fields.find((field) => next[field.id]);
    if (invalid) {
      document
        .getElementById(`field-card-${invalid.id}`)
        ?.scrollIntoView({ behavior: "smooth", block: "center" });
      return null;
    }
    return fields.map((field) => ({
      fieldId: field.id,
      value: formatAnswerForApi(field, valueOf(field)),
    }));
  }
  return {
    fields,
    errors,
    valueOf,
    validate,
    blocked,
    isLoading: form.isPending || saved.isPending,
    error: form.error?.message ?? saved.error?.message,
    retry: () => {
      void form.refetch();
      void saved.refetch();
    },
    change: (field: (typeof fields)[number], value: AnswerValue) => {
      setAnswers((previous) => ({ ...previous, [field.id]: value }));
      setErrors((previous) => ({
        ...previous,
        [field.id]: validateAnswer(field, value),
      }));
    },
    refresh: () => saved.refetch(),
  };
}

export type RegistrationAnswersState = ReturnType<
  typeof useRegistrationAnswers
>;
