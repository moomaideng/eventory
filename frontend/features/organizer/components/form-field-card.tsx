"use client";

import React from "react";
import { ChevronDown, ChevronUp, Plus, Trash2, X } from "lucide-react";
import { Badge } from "@/components/ui/badge";
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
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import type { FieldType, FormField, FormFieldOption } from "@/features/organizer/schemas";

const FIELD_TYPE_OPTIONS = [
  { value: "SHORT_TEXT", label: "Short text" },
  { value: "LONG_TEXT", label: "Long text" },
  { value: "NUMBER", label: "Number" },
  { value: "SELECT_ONE", label: "Select one" },
  { value: "SELECT_MULTIPLE", label: "Select multiple" },
];

export interface FormFieldCardProps {
  field: FormField;
  index: number;
  total: number;
  onUpdate: (updated: FormField) => void;
  onDelete: () => void;
  onMoveUp: () => void;
  onMoveDown: () => void;
}

export function FormFieldCard({
  field,
  index,
  total,
  onUpdate,
  onDelete,
  onMoveUp,
  onMoveDown,
}: FormFieldCardProps) {
  const handleTypeChange = (type: FieldType) => {
    const nextConfig = { ...field.config };
    if ((type === "SELECT_ONE" || type === "SELECT_MULTIPLE") && (!nextConfig.options || nextConfig.options.length === 0)) {
      nextConfig.options = [
        { id: `opt_${Date.now()}_1`, label: "Option 1" },
        { id: `opt_${Date.now()}_2`, label: "Option 2" },
      ];
    }
    onUpdate({ ...field, type, config: nextConfig });
  };

  const handleAddOption = () => {
    const currentOptions = field.config?.options ?? [];
    const newOption: FormFieldOption = {
      id: `opt_${Date.now()}`,
      label: `Option ${currentOptions.length + 1}`,
    };
    onUpdate({
      ...field,
      config: { ...field.config, options: [...currentOptions, newOption] },
    });
  };

  const handleUpdateOption = (optIndex: number, label: string) => {
    const currentOptions = [...(field.config?.options ?? [])];
    if (currentOptions[optIndex]) {
      currentOptions[optIndex] = { ...currentOptions[optIndex], label };
      onUpdate({
        ...field,
        config: { ...field.config, options: currentOptions },
      });
    }
  };

  const handleDeleteOption = (optIndex: number) => {
    const currentOptions = (field.config?.options ?? []).filter((_, i) => i !== optIndex);
    onUpdate({
      ...field,
      config: { ...field.config, options: currentOptions },
    });
  };

  return (
    <Card className="border-border">
      <CardHeader className="flex flex-row items-center justify-between pb-2">
        <div className="flex items-center gap-2">
          <Badge variant="outline" className="font-mono">
            #{index + 1}
          </Badge>
          <CardTitle className="text-base">Field Configuration</CardTitle>
        </div>
        <div className="flex items-center gap-1">
          <Button
            type="button"
            variant="ghost"
            size="sm"
            onClick={onMoveUp}
            disabled={index === 0}
            aria-label="Move field up"
          >
            <ChevronUp data-icon="inline-start" />
          </Button>
          <Button
            type="button"
            variant="ghost"
            size="sm"
            onClick={onMoveDown}
            disabled={index === total - 1}
            aria-label="Move field down"
          >
            <ChevronDown data-icon="inline-start" />
          </Button>
          <Button
            type="button"
            variant="ghost"
            size="sm"
            onClick={onDelete}
            aria-label="Delete field"
            className="text-destructive hover:bg-destructive/10"
          >
            <Trash2 data-icon="inline-start" />
          </Button>
        </div>
      </CardHeader>
      <CardContent>
        <FieldGroup>
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
            <div className="sm:col-span-2">
              <Field>
                <FieldLabel htmlFor={`label-${field.id}`}>Field Label</FieldLabel>
                <Input
                  id={`label-${field.id}`}
                  value={field.label}
                  onChange={(e) => onUpdate({ ...field, label: e.target.value })}
                  placeholder="e.g. In-Game Name, Discord ID"
                />
              </Field>
            </div>
            <div>
              <Field>
                <FieldLabel htmlFor={`type-${field.id}`}>Type</FieldLabel>
                <Select
                  items={FIELD_TYPE_OPTIONS}
                  value={field.type}
                  onValueChange={(val) => val && handleTypeChange(val as FieldType)}
                >
                  <SelectTrigger id={`type-${field.id}`}>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectGroup>
                      {FIELD_TYPE_OPTIONS.map((opt) => (
                        <SelectItem key={opt.value} value={opt.value}>
                          {opt.label}
                        </SelectItem>
                      ))}
                    </SelectGroup>
                  </SelectContent>
                </Select>
              </Field>
            </div>
          </div>

          <Field>
            <FieldLabel htmlFor={`desc-${field.id}`}>Help Text / Description</FieldLabel>
            <Input
              id={`desc-${field.id}`}
              value={field.description ?? ""}
              onChange={(e) => onUpdate({ ...field, description: e.target.value })}
              placeholder="Instructions or hint for applicant (optional)"
            />
          </Field>

          <div className="flex items-center gap-2 pt-1">
            <input
              type="checkbox"
              id={`req-${field.id}`}
              checked={field.required}
              onChange={(e) => onUpdate({ ...field, required: e.target.checked })}
              className="size-4 rounded border-input text-primary focus:ring-ring"
            />
            <label htmlFor={`req-${field.id}`} className="text-sm font-medium cursor-pointer">
              Required field
            </label>
          </div>

          {/* Type-Specific Options */}
          {(field.type === "SELECT_ONE" || field.type === "SELECT_MULTIPLE") && (
            <div className="flex flex-col gap-3 rounded-lg border border-border/70 bg-muted/20 p-4">
              <div className="text-sm font-medium">Choices</div>
              <div className="flex flex-col gap-2">
                {(field.config?.options ?? []).map((opt, optIdx) => (
                  <div key={opt.id} className="flex items-center gap-2">
                    <Input
                      value={opt.label}
                      onChange={(e) => handleUpdateOption(optIdx, e.target.value)}
                      placeholder={`Option ${optIdx + 1}`}
                      className="flex-1"
                    />
                    <Button
                      type="button"
                      variant="ghost"
                      size="sm"
                      onClick={() => handleDeleteOption(optIdx)}
                      disabled={(field.config?.options?.length ?? 0) <= 1}
                      aria-label="Remove option"
                    >
                      <X data-icon="inline-start" />
                    </Button>
                  </div>
                ))}
              </div>
              <div>
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  onClick={handleAddOption}
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
                      value={field.config?.minSelections ?? ""}
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
                      value={field.config?.maxSelections ?? ""}
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
          )}

          {field.type === "NUMBER" && (
            <div className="grid grid-cols-1 gap-3 sm:grid-cols-3 rounded-lg border border-border/70 bg-muted/20 p-4">
              <Field>
                <FieldLabel htmlFor={`min-${field.id}`}>Min Value</FieldLabel>
                <Input
                  id={`min-${field.id}`}
                  type="number"
                  value={field.config?.min ?? ""}
                  onChange={(e) =>
                    onUpdate({
                      ...field,
                      config: {
                        ...field.config,
                        min: e.target.value === "" ? undefined : Number(e.target.value),
                      },
                    })
                  }
                  placeholder="Optional"
                />
              </Field>
              <Field>
                <FieldLabel htmlFor={`max-${field.id}`}>Max Value</FieldLabel>
                <Input
                  id={`max-${field.id}`}
                  type="number"
                  value={field.config?.max ?? ""}
                  onChange={(e) =>
                    onUpdate({
                      ...field,
                      config: {
                        ...field.config,
                        max: e.target.value === "" ? undefined : Number(e.target.value),
                      },
                    })
                  }
                  placeholder="Optional"
                />
              </Field>
              <div className="flex items-center gap-2 pt-6">
                <input
                  type="checkbox"
                  id={`dec-${field.id}`}
                  checked={field.config?.allowDecimals ?? false}
                  onChange={(e) =>
                    onUpdate({
                      ...field,
                      config: { ...field.config, allowDecimals: e.target.checked },
                    })
                  }
                  className="size-4 rounded border-input text-primary focus:ring-ring"
                />
                <label htmlFor={`dec-${field.id}`} className="text-sm font-medium cursor-pointer">
                  Allow decimals
                </label>
              </div>
            </div>
          )}

          {(field.type === "SHORT_TEXT" || field.type === "LONG_TEXT") && (
            <div className="grid grid-cols-2 gap-3 rounded-lg border border-border/70 bg-muted/20 p-4">
              <Field>
                <FieldLabel htmlFor={`min-len-${field.id}`}>Min Length</FieldLabel>
                <Input
                  id={`min-len-${field.id}`}
                  type="number"
                  min={0}
                  value={field.config?.minLength ?? ""}
                  onChange={(e) =>
                    onUpdate({
                      ...field,
                      config: {
                        ...field.config,
                        minLength: e.target.value === "" ? undefined : Number(e.target.value),
                      },
                    })
                  }
                  placeholder="Optional"
                />
              </Field>
              <Field>
                <FieldLabel htmlFor={`max-len-${field.id}`}>Max Length</FieldLabel>
                <Input
                  id={`max-len-${field.id}`}
                  type="number"
                  min={0}
                  value={field.config?.maxLength ?? ""}
                  onChange={(e) =>
                    onUpdate({
                      ...field,
                      config: {
                        ...field.config,
                        maxLength: e.target.value === "" ? undefined : Number(e.target.value),
                      },
                    })
                  }
                  placeholder="Optional"
                />
              </Field>
            </div>
          )}
        </FieldGroup>
      </CardContent>
    </Card>
  );
}
