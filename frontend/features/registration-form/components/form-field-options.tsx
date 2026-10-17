"use client";

import React, { useState } from "react";
import { Check, Pencil, Plus, X } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Field, FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import type { FormField, FormFieldOption } from "@/features/registration-form/schemas";

export interface ChoiceRowProps {
  option: FormFieldOption;
  index: number;
  fieldType: "SELECT_ONE" | "SELECT_MULTIPLE";
  canDelete: boolean;
  isLast: boolean;
  initialEditing?: boolean;
  onUpdate: (label: string) => void;
  onDelete: () => void;
  onEnterLast?: () => void;
}

export function ChoiceRow({
  option,
  index,
  fieldType,
  canDelete,
  isLast,
  initialEditing = false,
  onUpdate,
  onDelete,
  onEnterLast,
}: ChoiceRowProps) {
  const [isEditing, setIsEditing] = useState(initialEditing);

  return (
    <div className="flex items-center justify-between gap-2 min-h-8">
      <div className="flex items-center gap-2 flex-1 min-w-0">
        {/* Choice indicator bullet */}
        {fieldType === "SELECT_ONE" ? (
          <div
            className="size-3.5 rounded-full border border-muted-foreground/60 shrink-0"
            aria-hidden="true"
          />
        ) : (
          <div
            className="size-3.5 rounded-[3px] border border-muted-foreground/60 shrink-0"
            aria-hidden="true"
          />
        )}

        {isEditing ? (
          <div className="flex items-center gap-1.5 flex-1 min-w-0">
            <Input
              autoFocus
              value={option.label}
              onFocus={(e) => e.target.select()}
              onChange={(e) => onUpdate(e.target.value)}
              onBlur={() => setIsEditing(false)}
              onKeyDown={(e) => {
                if (e.key === "Enter") {
                  e.preventDefault();
                  setIsEditing(false);
                  if (isLast && onEnterLast) {
                    onEnterLast();
                  }
                }
              }}
              placeholder={`Option ${index + 1}`}
              className="h-auto py-0 px-0 rounded-none border-x-0 border-t-0 border-b border-muted-foreground/30 text-sm md:text-sm font-normal leading-normal text-foreground shadow-none focus-visible:border-b-2 focus-visible:border-primary focus-visible:ring-0 bg-transparent flex-1"
            />
            <Button
              type="button"
              variant="ghost"
              size="sm"
              onMouseDown={(e) => e.preventDefault()}
              onClick={() => setIsEditing(false)}
              aria-label="Confirm choice"
              className="size-6 p-0 shrink-0 text-muted-foreground hover:text-foreground"
            >
              <Check className="size-3.5" />
            </Button>
          </div>
        ) : (
          <div className="flex items-center gap-1.5 flex-1 min-w-0">
            <span
              onClick={() => setIsEditing(true)}
              className="text-sm md:text-sm font-normal leading-normal text-foreground cursor-pointer truncate hover:text-primary transition-colors flex-1"
            >
              {option.label.trim() ? (
                option.label
              ) : (
                <span className="text-muted-foreground italic">Option {index + 1}</span>
              )}
            </span>
            <Button
              type="button"
              variant="ghost"
              size="sm"
              onClick={() => setIsEditing(true)}
              aria-label={`Edit option ${index + 1}`}
              className="size-6 p-0 shrink-0 text-muted-foreground hover:text-foreground"
            >
              <Pencil className="size-3" />
            </Button>
          </div>
        )}
      </div>

      <Button
        type="button"
        variant="ghost"
        size="sm"
        onClick={onDelete}
        disabled={!canDelete}
        aria-label={`Remove option ${index + 1}`}
        className="size-7 p-0 shrink-0 text-muted-foreground hover:text-destructive"
      >
        <X data-icon="inline-start" />
      </Button>
    </div>
  );
}

export interface FormFieldOptionsProps {
  field: Extract<FormField, { type: "SELECT_ONE" | "SELECT_MULTIPLE" }>;
  onUpdate: (updated: FormField) => void;
}

export function FormFieldOptions({ field, onUpdate }: FormFieldOptionsProps) {
  const [autoEditId, setAutoEditId] = useState<string | null>(null);

  const handleAddOption = (autoEdit = false) => {
    const currentOptions = field.config.options;
    const newId =
      typeof crypto !== "undefined" && typeof crypto.randomUUID === "function"
        ? `opt_${crypto.randomUUID()}`
        : `opt_${Date.now()}_${Math.random().toString(36).slice(2, 7)}`;
    const newOption: FormFieldOption = {
      id: newId,
      label: `Option ${currentOptions.length + 1}`,
    };
    if (autoEdit) {
      setAutoEditId(newId);
    }
    onUpdate({
      ...field,
      config: { ...field.config, options: [...currentOptions, newOption] },
    });
  };

  const handleUpdateOption = (optIndex: number, label: string) => {
    const currentOptions = [...field.config.options];
    if (currentOptions[optIndex]) {
      currentOptions[optIndex] = { ...currentOptions[optIndex], label };
      onUpdate({
        ...field,
        config: { ...field.config, options: currentOptions },
      });
    }
  };

  const handleDeleteOption = (optIndex: number) => {
    const currentOptions = field.config.options.filter((_, i) => i !== optIndex);
    onUpdate({
      ...field,
      config: { ...field.config, options: currentOptions },
    });
  };

  const options = field.config.options;

  return (
    <div className="flex flex-col gap-3 rounded-lg border border-border/70 bg-muted/20 p-4">
      <div className="text-sm font-medium">Choices</div>
      <div className="flex flex-col gap-2">
        {options.map((opt, optIdx) => {
          const isLast = optIdx === options.length - 1;
          return (
            <ChoiceRow
              key={opt.id}
              option={opt}
              index={optIdx}
              fieldType={field.type}
              canDelete={options.length > 1}
              isLast={isLast}
              initialEditing={autoEditId === opt.id}
              onUpdate={(label) => handleUpdateOption(optIdx, label)}
              onDelete={() => handleDeleteOption(optIdx)}
              onEnterLast={() => handleAddOption(true)}
            />
          );
        })}
      </div>
      <div>
        <Button
          type="button"
          variant="outline"
          size="sm"
          onClick={() => handleAddOption(false)}
        >
          <Plus data-icon="inline-start" />
          Add Choice
        </Button>
      </div>

      {field.type === "SELECT_MULTIPLE" && (
        <div className="grid grid-cols-2 gap-3 pt-2">
          <Field>
            <FieldLabel htmlFor={`min-sel-${field.id}`}>Min Selections</FieldLabel>
            <Input
              id={`min-sel-${field.id}`}
              type="number"
              min={0}
              value={field.config.minSelections ?? ""}
              onChange={(e) =>
                onUpdate({
                  ...field,
                  config: {
                    ...field.config,
                    minSelections: e.target.value === "" ? undefined : Number(e.target.value),
                  },
                })
              }
              placeholder="Optional"
            />
          </Field>
          <Field>
            <FieldLabel htmlFor={`max-sel-${field.id}`}>Max Selections</FieldLabel>
            <Input
              id={`max-sel-${field.id}`}
              type="number"
              min={0}
              value={field.config.maxSelections ?? ""}
              onChange={(e) =>
                onUpdate({
                  ...field,
                  config: {
                    ...field.config,
                    maxSelections: e.target.value === "" ? undefined : Number(e.target.value),
                  },
                })
              }
              placeholder="Optional"
            />
          </Field>
        </div>
      )}
    </div>
  );
}
