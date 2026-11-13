"use client";

import { useState } from "react";
import { X } from "lucide-react";
import {
  Field,
  FieldDescription,
  FieldError,
  FieldGroup,
  FieldLabel,
} from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
  normalizeSponsorTags,
  parseSponsorTagDraft,
} from "@/features/sponsor/tags";
import { sponsorTagsSchema } from "@/features/sponsor/schemas";

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
  function add() {
    const pending = parseSponsorTagDraft(draft);
    if (!pending.length) return;
    const result = sponsorTagsSchema.safeParse([...selected, ...pending]);
    if (!result.success) {
      setAddError(result.error.issues[0].message);
      return;
    }
    onChange(result.data);
    onDraftChange("");
    setAddError("");
  }
  return (
    <FieldGroup>
      <Field data-invalid={Boolean(addError || error)} data-disabled={disabled}>
        <FieldLabel htmlFor={id}>Tags</FieldLabel>
        <FieldDescription id={id + "-help"}>
          Add hashtags that describe your interests, like #gaming #valorant. Use
          up to 10 tags, with up to 40 characters each.
        </FieldDescription>
        <Input
          id={id}
          value={draft}
          disabled={disabled}
          placeholder="Type a hashtag, e.g. #gaming"
          aria-invalid={Boolean(addError || error)}
          aria-describedby={id + "-help " + id + "-instructions " + id + "-error"}
          onChange={(event) => {
            onDraftChange(event.target.value);
            setAddError("");
          }}
          onKeyDown={(event) => {
            if (event.nativeEvent.isComposing) return;
            if (
              event.key === "Enter" ||
              event.key === " " ||
              event.key === ","
            ) {
              event.preventDefault();
              add();
            }
          }}
          onBlur={add}
        />
        <FieldDescription id={id + "-instructions"}>
          Type a tag, then press Space or Enter to add it below. You can paste
          several hashtags at once. Any tags still in this box are included when
          you save your profile.
        </FieldDescription>
        <FieldError id={id + "-error"}>{addError || error}</FieldError>
      </Field>
      <div className="flex flex-col gap-2" aria-live="polite">
        <p className="text-muted-foreground text-sm">
          Your tags ({selected.length}/10)
        </p>
        <div className="flex flex-wrap gap-2">
          {selected.map((tag) => (
            <Badge
              key={tag}
              variant="secondary"
              className="max-w-full gap-1 break-words whitespace-normal"
            >
              #{tag}
              <Button
                type="button"
                variant="ghost"
                size="icon-xs"
                aria-label={`Remove ${tag}`}
                disabled={disabled}
                onClick={() => {
                  onChange(selected.filter((item) => item !== tag));
                  setAddError("");
                }}
              >
                <X />
              </Button>
            </Badge>
          ))}
          {!selected.length && (
            <p className="text-muted-foreground text-sm">No tags added yet.</p>
          )}
        </div>
      </div>
    </FieldGroup>
  );
}
