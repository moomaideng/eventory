import React from "react";
import { ClipboardList, Lock } from "lucide-react";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { FormBuilderSidebar } from "@/features/registration-form/components/form-builder-sidebar";
import { FormFieldPreviewRow } from "@/features/registration-form/components/form-field-preview-row";
import { formLockReason } from "@/features/registration-form/lock";
import { parseFormFields } from "@/features/registration-form/utils";
import type { components } from "@/lib/api/schema";

type RegistrationFormResponse = components["schemas"]["RegistrationFormResponse"];

/** Read-only view of the saved form, shown once registration has opened. */
export function RegistrationFormPreview({
  data,
  status,
  onViewRegistrations,
}: {
  data?: RegistrationFormResponse;
  status: string;
  onViewRegistrations?: () => void;
}) {
  const fields = parseFormFields(data?.fields);

  const handleScrollToField = (id: string) => {
    document
      .getElementById(`field-${id}`)
      ?.scrollIntoView({ behavior: "smooth", block: "center" });
  };

  return (
    <div className="flex flex-col gap-6 w-full">
      <div className="flex flex-col gap-1 pb-4 border-b border-border/40">
        <h2 className="text-xl font-semibold tracking-tight">Registration Questions</h2>
        <p className="text-muted-foreground text-sm">This is the form players see.</p>
      </div>

      <Alert role="status">
        <Lock />
        <AlertTitle>Questions can no longer be edited</AlertTitle>
        <AlertDescription>
          <p>{formLockReason(status)}</p>
          {onViewRegistrations ? (
            <Button
              type="button"
              variant="outline"
              size="sm"
              className="mt-3"
              onClick={onViewRegistrations}
            >
              <ClipboardList data-icon="inline-start" />
              View registrations
            </Button>
          ) : null}
        </AlertDescription>
      </Alert>

      <div className="grid grid-cols-1 lg:grid-cols-[minmax(0,1fr)_20rem] xl:grid-cols-[minmax(0,1fr)_22rem] gap-8 items-start w-full">
        <Card className="min-w-0 border-border border-t-4 border-t-primary shadow-xs">
          <CardContent className="flex flex-col gap-5 pt-6">
            <div className="flex flex-col gap-1.5">
              <h3 className="font-heading text-xl font-bold leading-normal tracking-tight">
                {data?.title?.trim() || "Tournament Registration Form"}
              </h3>
              {data?.description?.trim() ? (
                <p className="whitespace-pre-wrap text-sm text-muted-foreground">
                  {data.description}
                </p>
              ) : null}
            </div>

            {fields.length === 0 ? (
              <p className="border-t pt-5 text-sm text-muted-foreground">
                No custom questions were added. Players only provided the basic competitor
                details collected by default.
              </p>
            ) : (
              <ul className="flex flex-col divide-y border-t pt-5">
                {fields.map((field, idx) => (
                  <FormFieldPreviewRow key={field.id} field={field} index={idx} />
                ))}
              </ul>
            )}
          </CardContent>
        </Card>

        <FormBuilderSidebar fields={fields} onScrollToField={handleScrollToField} readOnly />
      </div>
    </div>
  );
}
