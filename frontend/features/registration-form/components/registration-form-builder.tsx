"use client";

import React, { useState } from "react";
import { Plus, Save } from "lucide-react";
import { useAuth } from "@/context/auth-context";
import { $api } from "@/lib/api/client";
import { Button } from "@/components/ui/button";
import { Spinner } from "@/components/ui/spinner";
import { FormFieldCard } from "@/features/registration-form/components/form-field-card";
import { FormHeaderCard } from "@/features/registration-form/components/form-header-card";
import { formFieldSchema, type FormField } from "@/features/registration-form/schemas";
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
      return initialData.fields.map((f, i) => {
        const parsed = formFieldSchema.safeParse({
          id: f.id,
          type: f.type,
          label: f.label,
          description: f.description ?? "",
          required: f.required ?? false,
          order: f.order ?? i + 1,
          config: typeof f.config === "object" && f.config !== null ? f.config : {},
        });
        if (parsed.success) {
          return parsed.data;
        }
        return {
          id: f.id,
          type: "SHORT_TEXT",
          label: f.label,
          description: f.description ?? "",
          required: f.required ?? false,
          order: f.order ?? i + 1,
          config: {},
        } satisfies FormField;
      });
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
      required: true,
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
      const field = fields[i];
      if (!field.label.trim()) {
        setErrorMessage(`Field #${i + 1} must have a label.`);
        return;
      }
      if (
        (field.type === "SELECT_ONE" || field.type === "SELECT_MULTIPLE") &&
        field.config.options.length === 0
      ) {
        setErrorMessage(`Field #${i + 1} (${field.label}) must have at least one choice.`);
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
      <FormHeaderCard
        title={title}
        description={description}
        onTitleChange={setTitle}
        onDescriptionChange={setDescription}
      />

      {/* Field Cards List */}
      <div className="flex flex-col gap-6">
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
      <div className="flex flex-wrap items-center justify-between gap-3 pt-2">
        <Button
          type="button"
          variant="outline"
          onClick={handleAddField}
        >
          <Plus data-icon="inline-start" />
          Add Question
        </Button>

        <div className="flex items-center gap-3">
          {saveSuccess && (
            <span className="text-sm font-medium text-primary">
              Form saved successfully.
            </span>
          )}

          {errorMessage && (
            <span className="text-sm font-medium text-destructive">
              {errorMessage}
            </span>
          )}

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
