"use client";

import React, { useState } from "react";
import { Plus, Save } from "lucide-react";
import { useAuth } from "@/context/auth-context";
import { $api } from "@/lib/api/client";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Field,
  FieldGroup,
  FieldLabel,
} from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Spinner } from "@/components/ui/spinner";
import { Textarea } from "@/components/ui/textarea";
import { FormFieldCard } from "@/features/organizer/components/form-field-card";
import type { FormField } from "@/features/organizer/schemas";
import type { components } from "@/lib/api/schema";

type RegistrationFormResponse = components["schemas"]["RegistrationFormResponse"];

interface RegistrationFormEditorProps {
  tournamentId: string;
  initialData?: RegistrationFormResponse;
}

function RegistrationFormEditor({
  tournamentId,
  initialData,
}: RegistrationFormEditorProps) {
  const { authorizationHeader } = useAuth();
  const [title, setTitle] = useState(initialData?.title ?? "Tournament Registration Form");
  const [description, setDescription] = useState(initialData?.description ?? "");
  const [fields, setFields] = useState<FormField[]>(() => {
    if (initialData?.fields && initialData.fields.length > 0) {
      return initialData.fields.map((f, i) => ({
        id: f.id,
        type: f.type,
        label: f.label,
        description: f.description ?? "",
        required: f.required ?? false,
        order: f.order ?? i + 1,
        config: {
          minLength: f.config?.minLength ?? undefined,
          maxLength: f.config?.maxLength ?? undefined,
          min: f.config?.min ?? undefined,
          max: f.config?.max ?? undefined,
          allowDecimals: f.config?.allowDecimals ?? undefined,
          options: f.config?.options ?? undefined,
          minSelections: f.config?.minSelections ?? undefined,
          maxSelections: f.config?.maxSelections ?? undefined,
        },
      }));
    }
    return [];
  });

  const [saveSuccess, setSaveSuccess] = useState(false);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  const saveMutation = $api.useMutation(
    "put",
    "/api/v1/tournaments/{tournamentId}/form"
  );

  const handleAddField = () => {
    const newField: FormField = {
      id: `fld_${Date.now()}`,
      type: "SHORT_TEXT",
      label: "",
      description: "",
      required: false,
      order: fields.length + 1,
      config: {},
    };
    setFields((prev) => [...prev, newField]);
  };

  const handleUpdateField = (index: number, updated: FormField) => {
    setFields((prev) => {
      const next = [...prev];
      next[index] = updated;
      return next;
    });
  };

  const handleDeleteField = (index: number) => {
    setFields((prev) => prev.filter((_, i) => i !== index));
  };

  const handleMoveField = (index: number, direction: -1 | 1) => {
    const targetIndex = index + direction;
    if (targetIndex < 0 || targetIndex >= fields.length) return;
    setFields((prev) => {
      const next = [...prev];
      const temp = next[index];
      next[index] = next[targetIndex];
      next[targetIndex] = temp;
      return next.map((f, idx) => ({ ...f, order: idx + 1 }));
    });
  };

  const handleSave = async () => {
    setErrorMessage(null);
    setSaveSuccess(false);

    if (!title.trim()) {
      setErrorMessage("Form title is required.");
      return;
    }

    for (let i = 0; i < fields.length; i++) {
      if (!fields[i].label.trim()) {
        setErrorMessage(`Field #${i + 1} must have a label.`);
        return;
      }
      if (
        (fields[i].type === "SELECT_ONE" || fields[i].type === "SELECT_MULTIPLE") &&
        (!fields[i].config?.options || (fields[i].config?.options?.length ?? 0) === 0)
      ) {
        setErrorMessage(`Field #${i + 1} (${fields[i].label}) must have at least one choice.`);
        return;
      }
    }

    try {
      await saveMutation.mutateAsync({
        params: { path: { tournamentId } },
        headers: { Authorization: authorizationHeader ?? "" },
        body: {
          title: title.trim(),
          description: description.trim() || undefined,
          fields: fields.map((f, i) => ({
            id: f.id,
            type: f.type,
            label: f.label.trim(),
            description: f.description?.trim() || undefined,
            required: f.required,
            order: i + 1,
            config: f.config,
          })),
        },
      });
      setSaveSuccess(true);
    } catch (err: unknown) {
      const message =
        err instanceof Error ? err.message : "Failed to save registration form.";
      setErrorMessage(message);
    }
  };

  return (
    <div className="mx-auto flex w-full max-w-3xl flex-col gap-6 px-4 py-8">
      {/* Form Title & Description Card */}
      <Card className="border-border">
        <CardHeader>
          <CardTitle className="text-lg">Registration Form Details</CardTitle>
        </CardHeader>
        <CardContent>
          <FieldGroup>
            <Field>
              <FieldLabel htmlFor="form-title">Form Title</FieldLabel>
              <Input
                id="form-title"
                value={title}
                onChange={(e) => setTitle(e.target.value)}
                placeholder="e.g. Tournament Registration Form"
              />
            </Field>
            <Field>
              <FieldLabel htmlFor="form-description">Description</FieldLabel>
              <Textarea
                id="form-description"
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                placeholder="Guidelines or instructions for applicants..."
                rows={3}
              />
            </Field>
          </FieldGroup>
        </CardContent>
      </Card>

      {/* Field Cards List */}
      <div className="flex flex-col gap-4">
        {fields.map((field, idx) => (
          <FormFieldCard
            key={field.id}
            field={field}
            index={idx}
            total={fields.length}
            onUpdate={(updated) => handleUpdateField(idx, updated)}
            onDelete={() => handleDeleteField(idx)}
            onMoveUp={() => handleMoveField(idx, -1)}
            onMoveDown={() => handleMoveField(idx, 1)}
          />
        ))}
      </div>

      {/* Actions */}
      <div className="flex flex-col gap-4">
        {saveSuccess && (
          <Alert className="border-emerald-500/40 bg-emerald-500/10 text-emerald-900 dark:text-emerald-200">
            <AlertTitle>Success</AlertTitle>
            <AlertDescription>Form has been saved successfully.</AlertDescription>
          </Alert>
        )}

        {errorMessage && (
          <Alert variant="destructive">
            <AlertTitle>Error</AlertTitle>
            <AlertDescription>{errorMessage}</AlertDescription>
          </Alert>
        )}

        <div className="flex items-center justify-between gap-3 pt-2">
          <Button
            type="button"
            variant="outline"
            onClick={handleAddField}
          >
            <Plus data-icon="inline-start" />
            Add Field
          </Button>

          <Button
            type="button"
            onClick={handleSave}
            disabled={saveMutation.isPending}
          >
            {saveMutation.isPending ? (
              <Spinner data-icon="inline-start" className="size-4" />
            ) : (
              <Save data-icon="inline-start" />
            )}
            Save Form
          </Button>
        </div>
      </div>
    </div>
  );
}

export function RegistrationFormBuilder({
  tournamentId,
}: {
  tournamentId: string;
}) {
  const { authorizationHeader, isLoading: authLoading } = useAuth();

  const { data, isLoading } = $api.useQuery(
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

  if (authLoading || isLoading) {
    return (
      <div className="flex flex-1 items-center justify-center p-12">
        <Spinner className="size-6 text-muted-foreground" />
      </div>
    );
  }

  return (
    <RegistrationFormEditor
      key={data?.id ?? "new"}
      tournamentId={tournamentId}
      initialData={data}
    />
  );
}
