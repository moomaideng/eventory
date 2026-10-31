"use client";

import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { FieldGroup } from "@/components/ui/field";
import { Skeleton } from "@/components/ui/skeleton";
import { FormFieldAnswerCard } from "@/features/registration-form/components/form-field-answer-card";
import type { RegistrationAnswersState } from "@/features/registration-form/hooks/use-registration-answers";

export function RegistrationAnswerFields({
  state,
}: {
  state: RegistrationAnswersState;
}) {
  if (state.isLoading) return <Skeleton className="h-28 w-full" />;
  if (state.error)
    return (
      <Alert variant="destructive">
        <AlertTitle>Registration questions unavailable</AlertTitle>
        <AlertDescription>
          {state.error}
          <Button type="button" variant="outline" onClick={state.retry}>
            Retry
          </Button>
        </AlertDescription>
      </Alert>
    );
  return (
    <FieldGroup>
      {state.fields.length === 0 ? (
        <p className="text-muted-foreground text-sm">
          No registration questions are required.
        </p>
      ) : (
        state.fields.map((field, index) => (
          <FormFieldAnswerCard
            key={field.id}
            field={field}
            index={index}
            value={state.valueOf(field)}
            error={state.errors[field.id]}
            onChange={(value) => state.change(field, value)}
          />
        ))
      )}
    </FieldGroup>
  );
}
