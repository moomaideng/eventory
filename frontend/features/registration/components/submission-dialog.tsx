"use client";

import { useState } from "react";
import { Download } from "lucide-react";
import { $api } from "@/lib/api/client";
import { extractProblemMessage } from "@/lib/schemas";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyTitle,
} from "@/components/ui/empty";
import type { components } from "@/lib/api/schema";

export type SubmissionParticipant = { id: string; name: string };

export function SubmissionDialog({
  participant,
  authorizationHeader,
  onClose,
}: {
  participant: SubmissionParticipant | null;
  authorizationHeader: string;
  onClose: () => void;
}) {
  return (
    <Dialog
      open={Boolean(participant)}
      onOpenChange={(open) => {
        if (!open) onClose();
      }}
    >
      <DialogContent className="max-h-[85dvh] overflow-y-auto sm:max-w-xl">
        <DialogHeader>
          <DialogTitle>Submitted registration</DialogTitle>
          <DialogDescription>
            {participant?.name} · Original questions and answers at the time of
            submission.
          </DialogDescription>
        </DialogHeader>
        {participant ? (
          <SubmissionDetails
            key={participant.id}
            memberId={participant.id}
            authorizationHeader={authorizationHeader}
          />
        ) : null}
        <DialogFooter>
          <Button type="button" variant="outline" onClick={onClose}>
            Close
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

function SubmissionDetails({
  memberId,
  authorizationHeader,
}: {
  memberId: string;
  authorizationHeader: string;
}) {
  const { data, error, isLoading, refetch } = $api.useQuery(
    "get",
    "/api/v1/lobbies/members/{memberId}/registration",
    {
      params: { path: { memberId } },
      headers: { Authorization: authorizationHeader },
    },
    {
      enabled: Boolean(authorizationHeader),
      staleTime: 0,
      gcTime: 0,
      retry: false,
    }
  );
  if (isLoading) return <Skeleton className="h-48 w-full" />;
  if (error || !data)
    return (
      <Alert variant="destructive">
        <AlertDescription>
          {extractProblemMessage(error, "Could not load this submission.")}
          <Button
            type="button"
            variant="outline"
            onClick={() => void refetch()}
          >
            Try again
          </Button>
        </AlertDescription>
      </Alert>
    );
  const questions = data.questions ?? [];
  const answers = new Map(
    (data.answers ?? []).map((answer) => [answer.questionId, answer])
  );
  return (
    <div className="flex flex-col gap-6">
      <Badge variant="outline" className="self-start">
        Form version {data.formVersion}
      </Badge>
      {questions.length ? (
        <dl className="flex flex-col gap-5">
          {questions.map((q) => {
            const answer = answers.get(q.id);
            return (
              <div key={q.id} className="flex flex-col gap-2">
                <dt className="font-medium wrap-anywhere">{q.label}</dt>
                <dd className="text-muted-foreground wrap-anywhere whitespace-pre-wrap">
                  {q.type === "FILE" && answer?.file ? (
                    <SubmissionFile file={answer.file} />
                  ) : (
                    answer?.value || "Not provided"
                  )}
                </dd>
              </div>
            );
          })}
        </dl>
      ) : (
        <Empty>
          <EmptyHeader>
            <EmptyTitle>No additional questions</EmptyTitle>
            <EmptyDescription>
              No question answers were collected with this registration.
            </EmptyDescription>
          </EmptyHeader>
        </Empty>
      )}
      {data.consentNotice ? (
        <div className="flex flex-col gap-2">
          <h3 className="font-medium">Agreed consent notice</h3>
          <p className="text-muted-foreground wrap-anywhere whitespace-pre-wrap">
            {data.consentNotice}
          </p>
        </div>
      ) : null}
    </div>
  );
}

function SubmissionFile({
  file,
}: {
  file: components["schemas"]["RegistrationFile"];
}) {
  const [failure, setFailure] = useState("");
  function download() {
    setFailure("");
    try {
      const bytes = Uint8Array.from(atob(file.data), (character) =>
        character.charCodeAt(0)
      );
      // Force a download instead of rendering uploaded content as a web page.
      const url = URL.createObjectURL(
        new Blob([bytes.buffer], { type: "application/octet-stream" })
      );
      const link = document.createElement("a");
      link.href = url;
      link.download = file.name;
      document.body.appendChild(link);
      link.click();
      link.remove();
      window.setTimeout(() => URL.revokeObjectURL(url), 1000);
    } catch {
      setFailure("Could not download this file. Please try again.");
    }
  }
  return (
    <div className="flex flex-col gap-2">
      <span>{file.name}</span>
      <Button
        type="button"
        variant="outline"
        size="sm"
        onClick={download}
        className="self-start"
      >
        <Download data-icon="inline-start" />
        Download file
      </Button>
      {failure ? (
        <Alert variant="destructive">
          <AlertDescription>{failure}</AlertDescription>
        </Alert>
      ) : null}
    </div>
  );
}
