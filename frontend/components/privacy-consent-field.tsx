"use client";

import { Checkbox } from "@/components/ui/checkbox";
import {
  Field,
  FieldContent,
  FieldDescription,
  FieldError,
  FieldLabel,
} from "@/components/ui/field";

interface PrivacyConsentFieldProps {
  id: string;
  checked: boolean;
  onCheckedChange: (checked: boolean) => void;
  description: string;
  label?: string;
  error?: string;
  disabled?: boolean;
}

/**
 * Explicit data-privacy consent checkbox rendered before saving new
 * organizer or sponsor profile information to the database.
 */
export function PrivacyConsentField({
  id,
  checked,
  onCheckedChange,
  description,
  label = "I consent to saving and displaying this profile information",
  error,
  disabled = false,
}: PrivacyConsentFieldProps) {
  const descriptionId = `${id}-description`;
  const errorId = `${id}-error`;
  const describedBy = error ? `${descriptionId} ${errorId}` : descriptionId;

  return (
    <div className="border-border/70 bg-muted/20 rounded-lg border p-3">
      <Field
        orientation="horizontal"
        data-invalid={Boolean(error)}
        data-disabled={disabled || undefined}
      >
        <Checkbox
          id={id}
          checked={checked}
          onCheckedChange={(nextChecked) => onCheckedChange(Boolean(nextChecked))}
          disabled={disabled}
          aria-invalid={Boolean(error)}
          aria-describedby={describedBy}
          required
        />
        <FieldContent>
          <FieldLabel htmlFor={id} className="cursor-pointer font-medium">
            {label}
          </FieldLabel>
          <FieldDescription id={descriptionId}>{description}</FieldDescription>
          {error ? <FieldError id={errorId}>{error}</FieldError> : null}
        </FieldContent>
      </Field>
    </div>
  );
}
