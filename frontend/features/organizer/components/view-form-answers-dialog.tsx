"use client";

import { useState } from "react";
import { FileText } from "lucide-react";
import { useAuth } from "@/context/auth-context";
import { $api } from "@/lib/api/client";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Spinner } from "@/components/ui/spinner";

interface ViewFormAnswersDialogProps {
  teamId?: string;
  tournamentId: string;
  accountId: string;
  applicantName: string;
}

export function ViewFormAnswersDialog({
  teamId,
  tournamentId,
  accountId,
  applicantName,
}: ViewFormAnswersDialogProps) {
  const [open, setOpen] = useState(false);
  const { authorizationHeader } = useAuth();

  const formQuery = $api.useQuery(
    "get",
    "/api/v1/tournaments/{tournamentId}/form",
    {
      params: { path: { tournamentId } },
      headers: { Authorization: authorizationHeader ?? "" },
    },
    {
      enabled: open && Boolean(authorizationHeader),
      retry: false,
    }
  );

  const responsesQuery = $api.useQuery(
    "get",
    "/api/v1/tournaments/{tournamentId}/form/responses",
    {
      params: { path: { tournamentId } },
      headers: { Authorization: authorizationHeader ?? "" },
    },
    {
      enabled: open && Boolean(authorizationHeader),
      retry: false,
    }
  );

  const isLoading = formQuery.isLoading || responsesQuery.isLoading;
  const form = formQuery.data;
  const responses = responsesQuery.data?.items ?? [];
  const participantResponse =
    responses.find(
      (item) => item.accountId === accountId && item.teamId === teamId
    ) ?? responses.find((item) => item.accountId === accountId && !item.teamId);

  const answerMap = new Map<string, unknown>();
  if (participantResponse?.answers) {
    for (const ans of participantResponse.answers) {
      answerMap.set(ans.fieldId, ans.value);
    }
  }

  function formatValue(value: unknown): string {
    if (value === null || value === undefined || value === "") {
      return "—";
    }
    if (Array.isArray(value)) {
      return value.join(", ");
    }
    return String(value);
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger
        render={
          <Button
            variant="ghost"
            size="sm"
            className="text-xs"
            aria-label={`View answers for ${applicantName}`}
          />
        }
      >
        <FileText data-icon="inline-start" />
        Answers
      </DialogTrigger>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle>Registration Answers</DialogTitle>
          <DialogDescription>
            Submitted responses by{" "}
            <span className="font-medium text-foreground">{applicantName}</span>
          </DialogDescription>
        </DialogHeader>

        {isLoading ? (
          <div className="flex items-center justify-center py-8">
            <Spinner className="size-6 text-muted-foreground" />
          </div>
        ) : !form?.fields?.length ? (
          <p className="py-4 text-center text-sm text-muted-foreground">
            No registration questions configured for this tournament.
          </p>
        ) : !participantResponse ? (
          <p className="py-4 text-center text-sm text-muted-foreground">
            No registration answers submitted by this participant.
          </p>
        ) : (
          <div className="flex max-h-96 flex-col gap-4 overflow-y-auto pr-1">
            {form.fields.map((field) => {
              const val = answerMap.get(field.id);
              const formatted = formatValue(val);
              return (
                <div
                  key={field.id}
                  className="rounded-lg border bg-muted/30 p-3 text-sm"
                >
                  <div className="flex items-center justify-between gap-2">
                    <span className="font-medium">{field.label}</span>
                    {field.required && (
                      <Badge variant="outline" className="text-[10px]">
                        Required
                      </Badge>
                    )}
                  </div>
                  {field.description ? (
                    <p className="mt-0.5 text-xs text-muted-foreground">
                      {field.description}
                    </p>
                  ) : null}
                  <p className="mt-2 text-sm whitespace-pre-wrap font-normal text-foreground">
                    {formatted}
                  </p>
                </div>
              );
            })}
          </div>
        )}

        <DialogFooter>
          <DialogClose render={<Button variant="outline" />}>
            Close
          </DialogClose>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
