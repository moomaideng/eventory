"use client";

import React, { useEffect, useMemo, useState } from "react";
import { useRouter } from "next/navigation";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { useAuth } from "@/context/auth-context";
import { apiClient } from "@/lib/api/client";
import { PageHeader } from "@/components/page-header";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Spinner } from "@/components/ui/spinner";
import { FormFieldAnswerCard } from "@/features/registration-form/components/form-field-answer-card";
import {
  emptyAnswer,
  formatAnswerForApi,
  validateAnswer,
  type AnswerValue,
} from "@/features/registration-form/utils";
import {
  formFieldSchema,
  type FormField,
} from "@/features/registration-form/schemas";
import type { components } from "@/lib/api/schema";
import { isRegistrationOpen } from "@/features/tournaments/utils";
import { RegistrationPaymentButton } from "@/features/payments/components/registration-payment-button";

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

export function RegistrationFormView({
  tournamentId,
}: {
  tournamentId: string;
}) {
  const { authorizationHeader, isLoading: authLoading } = useAuth();
  const router = useRouter();
  const queryClient = useQueryClient();
  const [submitError, setSubmitError] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [isRedirecting, setIsRedirecting] = useState(false);

  useEffect(() => {
    if (!authLoading && !authorizationHeader) {
      router.push(`/login?redirect=/tournaments/${tournamentId}/form`);
    }
  }, [authLoading, authorizationHeader, router, tournamentId]);

  const tournamentQuery = useQuery({
    queryKey: ["get", "/api/v1/tournaments/{tournamentId}", tournamentId],
    queryFn: async () => {
      const result = await apiClient.GET("/api/v1/tournaments/{tournamentId}", {
        params: { path: { tournamentId } },
      });
      if (result.error || !result.data) {
        throw new Error(result.error?.detail ?? "Tournament not found");
      }
      return result.data;
    },
  });

  const myTeamQuery = useQuery({
    queryKey: [
      "get",
      "/api/v1/tournaments/{tournamentId}/my-team",
      tournamentId,
      authorizationHeader,
    ],
    staleTime: 0,
    refetchOnMount: "always",
    enabled: Boolean(authorizationHeader),
    retry: false,
    queryFn: async () => {
      const result = await apiClient.GET(
        "/api/v1/tournaments/{tournamentId}/my-team",
        {
          params: { path: { tournamentId } },
          headers: { Authorization: authorizationHeader },
        }
      );
      if (result.response.status === 404) return null;
      if (result.error) throw new Error("Failed to check registration status");
      return result.data;
    },
  });

  const { data, isLoading, isError } = useQuery({
    queryKey: ["solo-registration-form", tournamentId],
    enabled: Boolean(authorizationHeader),
    retry: false,
    queryFn: async () => {
      const result = await apiClient.GET(
        "/api/v1/tournaments/{tournamentId}/form",
        {
          params: { path: { tournamentId } },
        }
      );
      if (result.response.status === 404) return null;
      if (result.error || !result.data)
        throw new Error("Unable to load registration questions");
      return result.data;
    },
  });
  const fields = useMemo(
    () =>
      (data?.fields ?? []).map(toFormField).sort((a, b) => a.order - b.order),
    [data]
  );

  const [answers, setAnswers] = useState<Record<string, AnswerValue>>({});
  const [errors, setErrors] = useState<Record<string, string | null>>({});

  const answerOf = (field: FormField): AnswerValue =>
    answers[field.id] ?? emptyAnswer(field);

  const setError = (field: FormField, value: AnswerValue) =>
    setErrors((prev) => ({
      ...prev,
      [field.id]: validateAnswer(field, value),
    }));

  const handleChange = (field: FormField, value: AnswerValue) => {
    setAnswers((prev) => ({ ...prev, [field.id]: value }));
    // Re-check live only once the question has shown an error, so it clears as soon as it is fixed.
    if (errors[field.id]) setError(field, value);
  };

  const handleSubmit = async (e: React.FormEvent) => {
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
      return;
    }
    if (!authorizationHeader || isSubmitting) return;
    setSubmitError(null);
    setIsSubmitting(true);
    try {
      const answersPayload = fields.map((field) => ({
        fieldId: field.id,
        value: formatAnswerForApi(field, answerOf(field)),
      }));

      const result = await apiClient.POST(
        "/api/v1/tournaments/{tournamentId}/solo-register",
        {
          params: { path: { tournamentId } },
          headers: { Authorization: authorizationHeader },
          body: { answers: answersPayload },
        }
      );
      if (result.error || !result.data) {
        setSubmitError(
          result.error?.detail ??
            result.error?.title ??
            "Registration failed. Please try again."
        );
        setIsSubmitting(false);
        return;
      }
      setAnswers({});
      setErrors({});
      setIsRedirecting(true);
      void queryClient.invalidateQueries({
        queryKey: ["get", "/api/v1/tournaments/{tournamentId}"],
      });
      void queryClient.invalidateQueries({
        queryKey: ["get", "/api/v1/tournaments/{tournamentId}/my-team"],
      });
      void queryClient.invalidateQueries({
        queryKey: ["my-solo-registration-answers", tournamentId],
      });
      if (result.data.paymentId) {
        router.push(`/payments/${result.data.paymentId}/checkout`);
      } else {
        router.push(`/tournaments/${tournamentId}`);
      }
    } catch {
      setSubmitError(
        "Registration is temporarily unavailable. Please try again."
      );
      setIsSubmitting(false);
    }
  };

  if (
    authLoading ||
    isLoading ||
    tournamentQuery.isLoading ||
    myTeamQuery.isLoading
  ) {
    return (
      <div className="flex flex-1 items-center justify-center p-12">
        <Spinner className="text-muted-foreground size-6" />
      </div>
    );
  }

  if (tournamentQuery.isError || !tournamentQuery.data?.tournament) {
    return (
      <div className="mx-auto w-full max-w-3xl px-4 py-8">
        <Card>
          <CardContent className="flex flex-col items-center gap-4 pt-6 text-center">
            <h2 className="text-xl font-semibold">Tournament Not Found</h2>
            <p className="text-muted-foreground">
              The requested tournament could not be found.
            </p>
            <Button onClick={() => router.push("/tournaments")}>
              Back to Tournaments
            </Button>
          </CardContent>
        </Card>
      </div>
    );
  }

  const tournament = tournamentQuery.data.tournament;

  if (tournament.registrationMode !== "SOLO") {
    return (
      <div className="mx-auto w-full max-w-3xl px-4 py-8">
        <Card>
          <CardContent className="flex flex-col items-center gap-4 pt-6 text-center">
            <h2 className="text-xl font-semibold">Team Tournament</h2>
            <p className="text-muted-foreground">
              This tournament requires team registration. Solo registration
              forms are not available for team tournaments.
            </p>
            <Button onClick={() => router.push(`/tournaments/${tournamentId}`)}>
              Back to Tournament
            </Button>
          </CardContent>
        </Card>
      </div>
    );
  }

  if (myTeamQuery.data && !isRedirecting) {
    const isRejected = myTeamQuery.data.status === "REJECTED";
    const isPending = ["FORMING", "PAYMENT_PENDING"].includes(
      myTeamQuery.data.status
    );
    return (
      <div className="mx-auto w-full max-w-3xl px-4 py-8">
        <Card>
          <CardContent className="flex flex-col items-center gap-4 pt-6 text-center">
            <h2 className="text-xl font-semibold">
              {isRejected
                ? "Application Rejected"
                : isPending
                  ? "Complete your registration"
                  : "Already Registered"}
            </h2>
            <p className="text-muted-foreground">
              {isRejected
                ? "Your application was not accepted for this tournament."
                : isPending
                  ? "Your answers are saved. Continue your existing registration without filling out the form again."
                  : "You are already registered for this tournament."}
            </p>
            {isPending && (
              <RegistrationPaymentButton
                tournamentId={tournamentId}
                teamId={myTeamQuery.data.id}
              />
            )}
            <Button onClick={() => router.push(`/tournaments/${tournamentId}`)}>
              View Tournament Details
            </Button>
          </CardContent>
        </Card>
      </div>
    );
  }

  if (!isRegistrationOpen(tournament)) {
    return (
      <div className="mx-auto w-full max-w-3xl px-4 py-8">
        <Card>
          <CardContent className="flex flex-col items-center gap-4 pt-6 text-center">
            <h2 className="text-xl font-semibold">Registration Closed</h2>
            <p className="text-muted-foreground">
              Registration is not currently open for this tournament.
            </p>
            <Button onClick={() => router.push(`/tournaments/${tournamentId}`)}>
              Back to Tournament
            </Button>
          </CardContent>
        </Card>
      </div>
    );
  }

  if (isError || myTeamQuery.isError) {
    return (
      <div className="mx-auto w-full max-w-3xl px-4 py-8">
        <Card>
          <CardContent className="text-muted-foreground flex flex-col items-center gap-4 pt-6 text-center">
            <p>Registration questions could not be loaded. Please try again.</p>
            <Button
              variant="outline"
              onClick={() => router.push(`/tournaments/${tournamentId}`)}
            >
              Back to Tournament
            </Button>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <form
      noValidate
      onSubmit={handleSubmit}
      className="mx-auto flex w-full max-w-7xl flex-1 flex-col gap-8 px-4 py-10 sm:px-8"
    >
      <PageHeader
        breadcrumbs={[
          { label: "Tournaments", href: "/tournaments" },
          { label: tournament.name, href: `/tournaments/${tournamentId}` },
          { label: "Registration" },
        ]}
        tags={<Badge variant="outline">Solo registration</Badge>}
        title={data?.title ?? "Individual registration"}
        description={
          data?.description ? (
            <span className="whitespace-pre-wrap">{data.description}</span>
          ) : undefined
        }
      />

      <div className="flex w-full max-w-3xl flex-col gap-6">
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

        <p className="text-muted-foreground text-sm">
          Check your answers before submitting. They cannot be edited afterward.
        </p>

        {submitError && (
          <Alert variant="destructive" aria-live="polite">
            <AlertTitle>Registration was not completed</AlertTitle>
            <AlertDescription>{submitError}</AlertDescription>
          </Alert>
        )}

        <div className="flex justify-end">
          <Button type="submit" disabled={isSubmitting || isRedirecting}>
            {(isSubmitting || isRedirecting) && (
              <Spinner data-icon="inline-start" />
            )}
            {isRedirecting ? "Redirecting..." : "Submit registration"}
          </Button>
        </div>
      </div>
    </form>
  );
}
