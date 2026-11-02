"use client";

import { useState, type FormEvent } from "react";
import { CircleCheck } from "lucide-react";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Field,
  FieldDescription,
  FieldError,
  FieldGroup,
  FieldLabel,
} from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Spinner } from "@/components/ui/spinner";
import { Textarea } from "@/components/ui/textarea";
import { usePostAnnouncement } from "@/features/announcements/hooks/use-announcements";
import {
  ANNOUNCEMENT_BODY_MAX,
  ANNOUNCEMENT_TITLE_MAX,
  announcementSchema,
} from "@/features/announcements/schemas";

type FieldErrors = { title?: string; body?: string };

export function AnnouncementForm({ tournamentId }: { tournamentId: string }) {
  const [title, setTitle] = useState("");
  const [body, setBody] = useState("");
  const [fieldErrors, setFieldErrors] = useState<FieldErrors>({});
  const [showSuccess, setShowSuccess] = useState(false);
  const { post, isPending, errorMessage, reset } =
    usePostAnnouncement(tournamentId);

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setShowSuccess(false);
    reset();

    const result = announcementSchema.safeParse({ title, body });
    if (!result.success) {
      const flat = result.error.flatten().fieldErrors;
      setFieldErrors({ title: flat.title?.[0], body: flat.body?.[0] });
      return;
    }

    setFieldErrors({});
    post(result.data, {
      onSuccess: () => {
        setTitle("");
        setBody("");
        setShowSuccess(true);
      },
    });
  }

  return (
    <form onSubmit={handleSubmit} noValidate>
      <Card>
        <CardHeader>
          <CardTitle>Post an announcement</CardTitle>
          <CardDescription>
            Shown to everyone on the tournament page. Members of accepted teams
            are also notified by email and in the app.
          </CardDescription>
        </CardHeader>

        <CardContent>
          <FieldGroup>
            <Field data-invalid={Boolean(fieldErrors.title)}>
              <FieldLabel htmlFor="announcement-title">Title</FieldLabel>
              <Input
                id="announcement-title"
                value={title}
                onChange={(event) => {
                  setTitle(event.target.value);
                  setFieldErrors((prev) => ({ ...prev, title: undefined }));
                  setShowSuccess(false);
                }}
                placeholder="e.g. Venue change for the final"
                maxLength={ANNOUNCEMENT_TITLE_MAX}
                disabled={isPending}
                aria-invalid={Boolean(fieldErrors.title)}
              />
              <FieldDescription>
                {title.length}/{ANNOUNCEMENT_TITLE_MAX} characters.
              </FieldDescription>
              {fieldErrors.title ? (
                <FieldError>{fieldErrors.title}</FieldError>
              ) : null}
            </Field>

            <Field data-invalid={Boolean(fieldErrors.body)}>
              <FieldLabel htmlFor="announcement-body">Message</FieldLabel>
              <Textarea
                id="announcement-body"
                rows={5}
                value={body}
                onChange={(event) => {
                  setBody(event.target.value);
                  setFieldErrors((prev) => ({ ...prev, body: undefined }));
                  setShowSuccess(false);
                }}
                placeholder="What do players need to know?"
                maxLength={ANNOUNCEMENT_BODY_MAX}
                disabled={isPending}
                aria-invalid={Boolean(fieldErrors.body)}
              />
              <FieldDescription>
                {body.length}/{ANNOUNCEMENT_BODY_MAX} characters. Plain text.
              </FieldDescription>
              {fieldErrors.body ? (
                <FieldError>{fieldErrors.body}</FieldError>
              ) : null}
            </Field>
          </FieldGroup>

          {errorMessage ? (
            <Alert variant="destructive">
              <AlertTitle>Could not post the announcement</AlertTitle>
              <AlertDescription>{errorMessage}</AlertDescription>
            </Alert>
          ) : null}

          {showSuccess ? (
            <Alert>
              <CircleCheck />
              <AlertTitle>Announcement posted</AlertTitle>
              <AlertDescription>
                It now appears at the top of the list below.
              </AlertDescription>
            </Alert>
          ) : null}
        </CardContent>

        <CardFooter className="justify-end">
          <Button type="submit" disabled={isPending}>
            {isPending ? <Spinner data-icon="inline-start" /> : null}
            {isPending ? "Posting…" : "Post announcement"}
          </Button>
        </CardFooter>
      </Card>
    </form>
  );
}
