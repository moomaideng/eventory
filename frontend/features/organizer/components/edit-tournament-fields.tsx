"use client";

import type { ReactNode } from "react";
import { Globe } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  Field,
  FieldDescription,
  FieldError,
  FieldGroup,
  FieldLabel,
} from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { MarkdownEditor } from "@/components/ui/markdown-editor";
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Textarea } from "@/components/ui/textarea";
import type {
  DetailsErrors,
  DetailsField,
  DetailsFormValues,
} from "@/features/organizer/tournament-details-form";

export type SectionFieldsProps = {
  values: DetailsFormValues;
  errors: DetailsErrors;
  disabled: boolean;
  onChange: (patch: Partial<DetailsFormValues>) => void;
};

function FormField({
  field,
  label,
  error,
  action,
  description,
  children,
}: {
  field: DetailsField;
  label: string;
  error?: string;
  action?: ReactNode;
  description?: ReactNode;
  children: ReactNode;
}) {
  return (
    <Field data-invalid={Boolean(error)}>
      <div className="flex items-center justify-between gap-2">
        <FieldLabel htmlFor={`edit-${field}`}>{label}</FieldLabel>
        {action}
      </div>
      {children}
      {description ? <FieldDescription>{description}</FieldDescription> : null}
      {error ? <FieldError>{error}</FieldError> : null}
    </Field>
  );
}

function SuffixInput({
  suffix,
  ...props
}: React.ComponentProps<typeof Input> & { suffix: string }) {
  return (
    <div className="relative">
      <Input {...props} className="pr-14" />
      <span className="text-muted-foreground pointer-events-none absolute inset-y-0 right-3 flex items-center text-sm">
        {suffix}
      </span>
    </div>
  );
}

function textProps(
  field: DetailsField,
  { values, errors, disabled, onChange }: SectionFieldsProps
) {
  return {
    id: `edit-${field}`,
    value: values[field],
    disabled,
    "aria-invalid": Boolean(errors[field]),
    onChange: (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) =>
      onChange({ [field]: e.target.value }),
  };
}

export function GeneralFields(props: SectionFieldsProps) {
  const { values, errors, disabled, onChange } = props;
  return (
    <FieldGroup>
      <FormField field="name" label="Tournament name" error={errors.name}>
        <Input {...textProps("name", props)} maxLength={160} />
      </FormField>
      <FormField field="game" label="Game" error={errors.game}>
        <Input {...textProps("game", props)} maxLength={80} />
      </FormField>
      <FormField
        field="location"
        label="Location"
        error={errors.location}
        action={
          <Button
            type="button"
            variant="outline"
            size="xs"
            disabled={disabled}
            onClick={() => onChange({ location: "Online" })}
          >
            <Globe data-icon="inline-start" />
            Online
          </Button>
        }
      >
        <Input {...textProps("location", props)} maxLength={160} />
      </FormField>
      <FormField
        field="description"
        label="Short description"
        error={errors.description}
        description="Shown on tournament cards and at the top of the public page."
      >
        <Textarea
          {...textProps("description", props)}
          maxLength={500}
          rows={3}
        />
      </FormField>
      <FormField
        field="detail"
        label="Rules & details"
        error={errors.detail}
        description="Markdown supported."
      >
        <MarkdownEditor
          id="edit-detail"
          value={values.detail}
          onChange={(detail) => onChange({ detail })}
          rows={8}
          maxLength={10000}
          disabled={disabled}
          aria-invalid={Boolean(errors.detail)}
        />
      </FormField>
    </FieldGroup>
  );
}

const MODE_ITEMS = [
  { label: "Team", value: "TEAM" },
  { label: "Solo", value: "SOLO" },
];

export function RegistrationFields(
  props: SectionFieldsProps & { currency: string }
) {
  const { values, errors, disabled, onChange, currency } = props;
  const isTeam = values.registrationMode === "TEAM";
  return (
    <FieldGroup>
      <FormField field="registrationMode" label="Format">
        <Select
          items={MODE_ITEMS}
          value={values.registrationMode}
          disabled={disabled}
          onValueChange={(mode) => {
            if (mode === "SOLO" || mode === "TEAM") {
              onChange({ registrationMode: mode });
            }
          }}
        >
          <SelectTrigger id="edit-registrationMode" className="w-full sm:w-60">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectGroup>
              {MODE_ITEMS.map((item) => (
                <SelectItem key={item.value} value={item.value}>
                  {item.label}
                </SelectItem>
              ))}
            </SelectGroup>
          </SelectContent>
        </Select>
      </FormField>
      {isTeam ? (
        <div className="grid gap-4 sm:grid-cols-2">
          <FormField
            field="minTeamSize"
            label="Min players per team"
            error={errors.minTeamSize}
          >
            <Input {...textProps("minTeamSize", props)} type="number" min={1} />
          </FormField>
          <FormField
            field="maxTeamSize"
            label="Max players per team"
            error={errors.maxTeamSize}
          >
            <Input {...textProps("maxTeamSize", props)} type="number" min={1} />
          </FormField>
        </div>
      ) : null}
      <div className="grid gap-4 sm:grid-cols-2">
        <FormField
          field="capacity"
          label={isTeam ? "Capacity (teams)" : "Capacity (players)"}
          error={errors.capacity}
        >
          <Input {...textProps("capacity", props)} type="number" min={1} />
        </FormField>
        <FormField
          field="entryFee"
          label="Entry fee"
          error={errors.entryFee}
          description="Use 0 for free entry."
        >
          <SuffixInput
            {...textProps("entryFee", props)}
            type="number"
            min={0}
            suffix={currency}
          />
        </FormField>
      </div>
    </FieldGroup>
  );
}

export function ScheduleFields(props: SectionFieldsProps) {
  const { errors } = props;
  return (
    <FieldGroup>
      <FormField
        field="registrationDeadline"
        label="Registration closes"
        error={errors.registrationDeadline}
      >
        <Input
          {...textProps("registrationDeadline", props)}
          type="datetime-local"
        />
      </FormField>
      <FormField field="startAt" label="Starts" error={errors.startAt}>
        <Input {...textProps("startAt", props)} type="datetime-local" />
      </FormField>
      <FormField field="endAt" label="Ends" error={errors.endAt}>
        <Input {...textProps("endAt", props)} type="datetime-local" />
      </FormField>
    </FieldGroup>
  );
}

export function CrowdfundingFields(
  props: SectionFieldsProps & { currency: string }
) {
  const { errors, currency } = props;
  return (
    <FieldGroup>
      <FormField
        field="fundingGoal"
        label="Funding goal"
        error={errors.fundingGoal}
      >
        <SuffixInput
          {...textProps("fundingGoal", props)}
          type="number"
          min={1}
          suffix={currency}
        />
      </FormField>
      <FormField
        field="fundingDeadline"
        label="Funding deadline"
        error={errors.fundingDeadline}
        description="Must be in the future and before registration closes."
      >
        <Input {...textProps("fundingDeadline", props)} type="datetime-local" />
      </FormField>
    </FieldGroup>
  );
}
