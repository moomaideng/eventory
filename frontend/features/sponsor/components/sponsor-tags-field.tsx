"use client";

import { useState } from "react";
import { Plus, X } from "lucide-react";
import {
  Field,
  FieldDescription,
  FieldError,
  FieldGroup,
  FieldLabel,
  FieldLegend,
  FieldSet,
} from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
  normalizeSponsorTags,
  suggestedSponsorTags,
} from "@/features/sponsor/tags";

export function SponsorTagsField({
  value,
  onChange,
  draft,
  onDraftChange,
  error,
  disabled,
  id = "sponsor-tags",
}: {
  value: string[];
  onChange: (value: string[]) => void;
  draft: string;
  onDraftChange: (value: string) => void;
  error?: string;
  disabled?: boolean;
  id?: string;
}) {
  const [addError, setAddError] = useState("");
  const selected = normalizeSponsorTags(value);
  function remove(tag: string) {
    onChange(selected.filter((item) => item !== tag));
    setAddError("");
  }
  function addCustomTag() {
    const tag = draft.trim().toLowerCase();
    if (!tag) return;
    if (tag.length > 40) {
      setAddError("Each tag must be 40 characters or fewer.");
      return;
    }
    if (selected.includes(tag)) {
      setAddError("This tag is already selected.");
      return;
    }
    if (selected.length >= 10) {
      setAddError("Use up to 10 tags. Remove a tag to add another.");
      return;
    }
    onChange([...selected, tag]);
    onDraftChange("");
    setAddError("");
  }
  return (
    <FieldSet disabled={disabled} aria-describedby={`${id}-help`}>
      <FieldLegend>Tags</FieldLegend>
      <FieldDescription id={`${id}-help`}>
        Choose any that fit, or add your own. You can select up to 10 tags in
        total.
      </FieldDescription>
      <FieldGroup className="grid grid-cols-1 gap-2 sm:grid-cols-2">
        {suggestedSponsorTags.map((label, index) => {
          const tag = label.toLowerCase();
          const checked = selected.includes(tag);
          return (
            <Field
              key={tag}
              orientation="horizontal"
              data-disabled={disabled || (!checked && selected.length >= 10)}
            >
              <input
                id={`${id}-option-${index}`}
                type="checkbox"
                className="accent-primary size-4 shrink-0"
                checked={checked}
                disabled={disabled || (!checked && selected.length >= 10)}
                onChange={(event) =>
                  event.target.checked
                    ? onChange([...selected, tag])
                    : remove(tag)
                }
              />
              <FieldLabel htmlFor={`${id}-option-${index}`}>{label}</FieldLabel>
            </Field>
          );
        })}
      </FieldGroup>
      <Field data-invalid={Boolean(addError || error)}>
        <FieldLabel htmlFor={`${id}-custom`}>Add a custom tag</FieldLabel>
        <div className="flex gap-2">
          <Input
            id={`${id}-custom`}
            value={draft}
            disabled={disabled}
            placeholder="e.g. valorant"
            aria-invalid={Boolean(addError)}
            aria-describedby={`${id}-custom-help`}
            onChange={(event) => {
              onDraftChange(event.target.value);
              setAddError("");
            }}
            onKeyDown={(event) => {
              if (event.key === "Enter" && !event.nativeEvent.isComposing) {
                event.preventDefault();
                addCustomTag();
              }
            }}
          />
          <Button
            type="button"
            variant="outline"
            disabled={disabled || !draft.trim()}
            onClick={addCustomTag}
          >
            <Plus data-icon="inline-start" />
            Add tag
          </Button>
        </div>
        <FieldDescription id={`${id}-custom-help`}>
          Type one tag, then click Add tag or press Enter. You can add more than
          one. Up to 40 characters per tag.
        </FieldDescription>
        {addError || error ? (
          <FieldError>{addError || error}</FieldError>
        ) : null}
      </Field>
      <div className="flex flex-col gap-2" aria-live="polite">
        <p className="text-muted-foreground text-sm">
          Selected tags ({selected.length}/10)
        </p>
        <div className="flex flex-wrap gap-2">
          {selected.map((tag) => (
            <Badge
              key={tag}
              variant="secondary"
              className="max-w-full gap-1 break-words whitespace-normal"
            >
              {suggestedSponsorTags.find(
                (label) => label.toLowerCase() === tag
              ) ?? tag}
              <Button
                type="button"
                variant="ghost"
                size="icon-xs"
                aria-label={`Remove ${tag}`}
                disabled={disabled}
                onClick={() => remove(tag)}
              >
                <X />
              </Button>
            </Badge>
          ))}
          {!selected.length ? (
            <p className="text-muted-foreground text-sm">
              No tags selected yet.
            </p>
          ) : null}
        </div>
      </div>
    </FieldSet>
  );
}
