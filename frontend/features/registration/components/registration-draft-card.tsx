"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { FormEditorDialog } from "@/features/registration/components/form-editor-dialog";
import {
  defaultConsentNotice,
  type RegistrationConfig,
} from "@/features/registration/schemas";

export function RegistrationDraftCard({
  value,
  onChange,
  disabled,
}: {
  value: RegistrationConfig | null;
  onChange: (value: RegistrationConfig | null) => void;
  disabled: boolean;
}) {
  const [open, setOpen] = useState(false);
  return (
    <>
      <Card>
        <CardHeader>
          <CardTitle>Registration form</CardTitle>
          <CardDescription>
            Ask each participant for the information you need.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <p className="text-muted-foreground text-sm">
            {value
              ? `${value.questions.length} questions ready. You can edit the form later without affecting existing participants.`
              : "Optional. Without a form, participants only need to give consent."}
          </p>
        </CardContent>
        <CardFooter className="gap-2">
          <Button
            type="button"
            variant="outline"
            disabled={disabled}
            onClick={() => setOpen(true)}
          >
            {value ? "Edit registration form" : "Add registration form"}
          </Button>
          {value ? (
            <Button
              type="button"
              variant="ghost"
              disabled={disabled}
              onClick={() => onChange(null)}
            >
              Remove form
            </Button>
          ) : null}
        </CardFooter>
      </Card>
      <FormEditorDialog
        open={open}
        onOpenChange={setOpen}
        draft
        initial={
          value ?? { questions: [], consentNotice: defaultConsentNotice }
        }
        onSave={async (config) => {
          onChange(config);
        }}
      />
    </>
  );
}
