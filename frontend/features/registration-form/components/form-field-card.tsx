"use client";

import React, { useState } from "react";
import { Check, ChevronDown, ChevronUp, Pencil, Trash2 } from "lucide-react";
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
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { FormFieldOptions } from "@/features/registration-form/components/form-field-options";
import type { FieldType, FormField } from "@/features/registration-form/schemas";

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
  const [isEditingLabel, setIsEditingLabel] = useState(!field.label);

  const handleTypeChange = (type: FieldType) => {
    const nextConfig = { ...field.config };
    if (
      (type === "SELECT_ONE" || type === "SELECT_MULTIPLE") &&
      (!nextConfig.options || nextConfig.options.length === 0)
    ) {
      nextConfig.options = [
        { id: `opt_${Date.now()}_1`, label: "Option 1" },
        { id: `opt_${Date.now()}_2`, label: "Option 2" },
      ];
    }
    onUpdate({ ...field, type, config: nextConfig });
  };

  return (
    <Card className="border-border">
      <CardHeader className="flex flex-row items-center justify-between pb-3 gap-3">
        {/* Field Label Heading / Editable Underscored Input */}
        <div className="flex items-center gap-2 min-w-0 flex-1">
          <Badge variant="outline" className="font-mono shrink-0">
            #{index + 1}
          </Badge>

          {isEditingLabel ? (
            <div className="flex items-center gap-1.5 min-w-0 flex-1 max-w-sm sm:max-w-md">
              <Input
                autoFocus
                value={field.label}
                onChange={(e) => onUpdate({ ...field, label: e.target.value })}
                onBlur={() => setIsEditingLabel(false)}
                onKeyDown={(e) => {
                  if (e.key === "Enter") {
                    setIsEditingLabel(false);
                  }
                }}
                placeholder="Enter question title..."
                className="h-auto py-0 px-0 rounded-none border-x-0 border-t-0 border-b border-muted-foreground/30 font-heading text-base md:text-base font-medium leading-normal text-card-foreground shadow-none focus-visible:border-b-2 focus-visible:border-primary focus-visible:ring-0 bg-transparent"
              />
              <Button
                type="button"
                variant="ghost"
                size="sm"
                onClick={() => setIsEditingLabel(false)}
                aria-label="Confirm label"
                className="size-7 p-0 shrink-0 text-muted-foreground hover:text-foreground"
              >
                <Check className="size-3.5" />
              </Button>
            </div>
          ) : (
            <div className="flex items-center gap-1.5 min-w-0">
              <CardTitle
                className="font-heading text-base md:text-base font-medium leading-normal text-card-foreground cursor-pointer truncate hover:text-primary transition-colors"
                onClick={() => setIsEditingLabel(true)}
              >
                {field.label.trim() ? (
                  field.label
                ) : (
                  <span className="text-muted-foreground italic">Untitled question</span>
                )}
              </CardTitle>
              <Button
                type="button"
                variant="ghost"
                size="sm"
                onClick={() => setIsEditingLabel(true)}
                aria-label="Edit question label"
                className="size-7 p-0 shrink-0 text-muted-foreground hover:text-foreground"
              >
                <Pencil className="size-3.5" />
              </Button>
            </div>
          )}
        </div>

        {/* Top Right: Required Checkbox & Card Actions */}
        <div className="flex items-center gap-2 sm:gap-3 shrink-0">
          <div className="flex items-center gap-1.5">
            <input
              type="checkbox"
              id={`req-${field.id}`}
              checked={field.required}
              onChange={(e) => onUpdate({ ...field, required: e.target.checked })}
              className="size-4 rounded border-input text-primary focus:ring-ring cursor-pointer"
            />
            <Label
              htmlFor={`req-${field.id}`}
              className="text-sm font-medium cursor-pointer select-none text-muted-foreground hover:text-foreground"
            >
              Required
            </Label>
          </div>

          <div className="h-4 w-px bg-border" aria-hidden="true" />

          {/* Card Actions: Move Up, Move Down, Delete */}
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
        </div>
      </CardHeader>

      <CardContent>
        <FieldGroup>
          {/* Top Row: Type and Help text */}
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <Field>
              <FieldLabel htmlFor={`type-${field.id}`}>Field Type</FieldLabel>
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

            <Field>
              <FieldLabel htmlFor={`desc-${field.id}`}>Help Text / Description</FieldLabel>
              <Input
                id={`desc-${field.id}`}
                value={field.description ?? ""}
                onChange={(e) => onUpdate({ ...field, description: e.target.value })}
                placeholder="Guidelines or instructions (optional)"
              />
            </Field>
          </div>

          {/* Type-Specific Options for SELECT_ONE & SELECT_MULTIPLE */}
          {(field.type === "SELECT_ONE" || field.type === "SELECT_MULTIPLE") && (
            <FormFieldOptions field={field} onUpdate={onUpdate} />
          )}

          {/* Type-Specific Options for NUMBER */}
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

          {/* Type-Specific Options for SHORT_TEXT & LONG_TEXT */}
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
