"use client";

import React, { useEffect, useMemo, useState } from "react";
import { Plus, Save } from "lucide-react";
import { useQueryClient } from "@tanstack/react-query";
import { useAuth } from "@/context/auth-context";
import { $api } from "@/lib/api/client";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Spinner } from "@/components/ui/spinner";
import { FormBuilderSidebar } from "@/features/registration-form/components/form-builder-sidebar";
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
  const queryClient = useQueryClient();
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

  const isDirty = useMemo(() => {
    const origTitle = initialData?.title ?? "Tournament Registration Form";
    const origDesc = initialData?.description ?? "";
    if (title.trim() !== origTitle.trim() || description.trim() !== origDesc.trim()) {
      return true;
    }
    const origFields = initialData?.fields ?? [];
    if (fields.length !== origFields.length) return true;

    for (let i = 0; i < fields.length; i++) {
      const f = fields[i];
      const of = origFields[i];
      if (
        f.id !== of.id ||
        f.type !== of.type ||
        f.label !== of.label ||
        (f.description ?? "") !== (of.description ?? "") ||
        f.required !== (of.required ?? false) ||
        JSON.stringify(f.config) !== JSON.stringify(of.config)
      ) {
        return true;
      }
    }
    return false;
  }, [title, description, fields, initialData]);

  useEffect(() => {
    if (!isDirty) return;
    const handleBeforeUnload = (e: BeforeUnloadEvent) => {
      e.preventDefault();
      e.returnValue = "";
    };
    window.addEventListener("beforeunload", handleBeforeUnload);
    return () => window.removeEventListener("beforeunload", handleBeforeUnload);
  }, [isDirty]);

  const handleScrollToField = (id: string) => {
    const el = document.getElementById(`field-${id}`);
    if (el) {
      el.scrollIntoView({ behavior: "smooth", block: "center" });
    }
  };

  const saveMutation = $api.useMutation(
    "put",
    "/api/v1/tournaments/{tournamentId}/form"
  );

  const handleAddField = () => {
    setSaveSuccess(false);
    const id =
      typeof crypto !== "undefined" && typeof crypto.randomUUID === "function"
        ? `fld_${crypto.randomUUID()}`
        : `fld_${Date.now()}_${Math.random().toString(36).slice(2, 7)}`;
    const newField: FormField = {
      id,
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
    setSaveSuccess(false);
    setFields((prev) => {
      const next = [...prev];
      next[index] = updated;
      return next;
    });
  };

  const handleDeleteField = (index: number) => {
    setSaveSuccess(false);
    setFields((prev) => prev.filter((_, i) => i !== index));
  };

  const handleMoveField = (index: number, direction: -1 | 1) => {
    const targetIndex = index + direction;
    if (targetIndex < 0 || targetIndex >= fields.length) return;
    setSaveSuccess(false);
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
      void queryClient.invalidateQueries();
    } catch (err: unknown) {
      const message =
        err instanceof Error ? err.message : "Failed to save registration form.";
      setErrorMessage(message);
    }
  };

  return (
    <div className="flex flex-col gap-6 w-full">
      {/* Top Action / Status Bar */}
      <div className="flex flex-wrap items-center justify-between gap-4 pb-4 border-b border-border/40">
        <div className="flex flex-col gap-1">
          <div className="flex items-center gap-2.5">
            <h2 className="text-xl font-semibold tracking-tight">Registration Questions</h2>
            {isDirty ? (
              <Badge variant="outline" className="text-amber-500 border-amber-500/30 bg-amber-500/10 text-xs">
                Unsaved changes
              </Badge>
            ) : saveSuccess ? (
              <Badge variant="outline" className="text-primary border-primary/30 bg-primary/10 text-xs">
                Saved
              </Badge>
            ) : null}
          </div>
          <p className="text-muted-foreground text-sm">
            Customize the questionnaire players must fill out when registering for this tournament.
          </p>
        </div>

        <div className="flex items-center gap-3">
          {errorMessage && (
            <span className="text-xs font-medium text-destructive">
              {errorMessage}
            </span>
          )}

          <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={handleAddField}
          >
            <Plus data-icon="inline-start" />
            Add Question
          </Button>

          <Button
            type="button"
            size="sm"
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

      {/* 2-Column Layout */}
      <div className="grid grid-cols-1 lg:grid-cols-[minmax(0,1fr)_20rem] xl:grid-cols-[minmax(0,1fr)_22rem] gap-8 items-start w-full">
        {/* Main Form Canvas */}
        <div className="flex flex-col gap-6 min-w-0">
          <FormHeaderCard
            title={title}
            description={description}
            onTitleChange={(val) => {
              setTitle(val);
              setSaveSuccess(false);
            }}
            onDescriptionChange={(val) => {
              setDescription(val);
              setSaveSuccess(false);
            }}
          />

          {fields.length === 0 ? (
            <div className="flex flex-col items-center justify-center p-8 sm:p-12 rounded-xl border border-dashed border-border/70 bg-card/40 text-center gap-3">
              <div className="size-10 rounded-full bg-primary/10 text-primary flex items-center justify-center">
                <Plus className="size-5" />
              </div>
              <div className="flex flex-col gap-1 max-w-sm">
                <h3 className="font-semibold text-sm">No custom questions added</h3>
                <p className="text-xs text-muted-foreground leading-relaxed">
                  Basic competitor details are already collected by default. Add custom fields if your tournament requires rank proof, Discord handles, or jerseys.
                </p>
              </div>
              <Button type="button" variant="outline" size="sm" onClick={handleAddField}>
                <Plus data-icon="inline-start" />
                Add Question
              </Button>
            </div>
          ) : (
            <div className="flex flex-col gap-6">
              {fields.map((field, idx) => (
                <div key={field.id} id={`field-${field.id}`} className="scroll-mt-6">
                  <FormFieldCard
                    field={field}
                    index={idx}
                    total={fields.length}
                    onUpdate={(updated) => handleUpdateField(idx, updated)}
                    onDelete={() => handleDeleteField(idx)}
                    onMoveUp={() => handleMoveField(idx, -1)}
                    onMoveDown={() => handleMoveField(idx, 1)}
                  />
                </div>
              ))}
            </div>
          )}

          {/* Bottom Actions */}
          {fields.length > 0 && (
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
                {errorMessage && (
                  <span className="text-xs font-medium text-destructive">
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
          )}
        </div>

        {/* Companion Sidebar */}
        <FormBuilderSidebar
          fields={fields}
          onScrollToField={handleScrollToField}
        />
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
