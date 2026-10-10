"use client";

import { useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { $api, apiClient } from "@/lib/api/client";
import { extractProblemMessage } from "@/lib/schemas";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Skeleton } from "@/components/ui/skeleton";
import { FormEditorDialog } from "@/features/registration/components/form-editor-dialog";
import {
  registrationConfigSchema,
  type RegistrationConfig,
} from "@/features/registration/schemas";

export function RegistrationFormManager({
  tournamentId,
  authorizationHeader,
}: {
  tournamentId: string;
  authorizationHeader: string;
}) {
  const [open, setOpen] = useState(false);
  const [saved, setSaved] = useState(false);
  const queryClient = useQueryClient();
  const query = $api.useQuery(
    "get",
    "/api/v1/tournaments/{tournamentId}/registration-form",
    {
      params: { path: { tournamentId } },
      headers: { Authorization: authorizationHeader },
    },
    { staleTime: 0, retry: false }
  );
  const config = registrationConfigSchema.safeParse(
    query.data
      ? {
          questions: (query.data.questions ?? []).map((q) => ({
            ...q,
            options: q.options ?? [],
            allowedTypes: q.allowedTypes ?? [],
          })),
          consentNotice: query.data.consentNotice,
        }
      : null
  );
  async function save(value: RegistrationConfig) {
    const { error } = await apiClient.PUT(
      "/api/v1/tournaments/{tournamentId}/registration-form",
      {
        params: { path: { tournamentId } },
        headers: { Authorization: authorizationHeader },
        body: value,
      }
    );
    if (error)
      throw new Error(extractProblemMessage(error, "Could not save the form."));
    await queryClient.invalidateQueries({
      queryKey: ["get", "/api/v1/tournaments/{tournamentId}/registration-form"],
    });
    setSaved(true);
  }
  return (
    <>
      <Card>
        <CardHeader>
          <CardTitle>Registration form</CardTitle>
          <CardDescription>
            Edits apply to new applicants. Existing participants do not need to
            fill in the form again.
          </CardDescription>
        </CardHeader>
        <CardContent>
          {query.isLoading ? (
            <Skeleton className="h-5 w-48" />
          ) : query.error || !config.success ? (
            <Alert variant="destructive">
              <AlertDescription>
                Could not load the registration form.{" "}
                <Button
                  type="button"
                  variant="outline"
                  onClick={() => void query.refetch()}
                >
                  Try again
                </Button>
              </AlertDescription>
            </Alert>
          ) : (
            <p className="text-muted-foreground text-sm">
              {config.data.questions.length} questions · Version{" "}
              {query.data?.version}
              {saved ? " · Saved" : ""}
            </p>
          )}
        </CardContent>
        <CardFooter>
          <Button
            type="button"
            variant="outline"
            disabled={
              query.isLoading || Boolean(query.error) || !config.success
            }
            onClick={() => {
              setSaved(false);
              setOpen(true);
            }}
          >
            {query.data?.version
              ? "Edit registration form"
              : "Add registration form"}
          </Button>
        </CardFooter>
      </Card>
      {config.success ? (
        <FormEditorDialog
          open={open}
          onOpenChange={setOpen}
          initial={config.data}
          onSave={save}
        />
      ) : null}
    </>
  );
}
