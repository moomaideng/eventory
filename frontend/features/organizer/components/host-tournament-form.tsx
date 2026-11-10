"use client";

import React, { useState, type FormEvent } from "react";
import { useRouter } from "next/navigation";
import { AlertCircle, FileEdit, PlusCircle } from "lucide-react";
import { apiClient } from "@/lib/api/client";
import { useAuth } from "@/context/auth-context";
import { extractProblemMessage } from "@/lib/schemas";
import { Button } from "@/components/ui/button";
import { Spinner } from "@/components/ui/spinner";
import {
  createTournamentFormSchema,
  type CreateTournamentFormInput,
} from "../schemas";
import { HostTournamentGeneralCard } from "./host-tournament-general-card";

export interface HostTournamentFormProps {
  organizerName?: string;
  onValuesChange: (values: CreateTournamentFormInput) => void;
}

type FormErrors = Partial<Record<keyof CreateTournamentFormInput, string>>;

export function HostTournamentForm({
  organizerName,
  onValuesChange,
}: HostTournamentFormProps) {
  const router = useRouter();
  const { authorizationHeader } = useAuth();

  const [name, setName] = useState("");
  const [game, setGame] = useState("");
  const [location, setLocation] = useState("Online");
  const [description, setDescription] = useState("");
  const [detail, setDetail] = useState("");

  const [fieldErrors, setFieldErrors] = useState<FormErrors>({});
  const [apiError, setApiError] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  function updateAndNotify(overrides: Partial<CreateTournamentFormInput> = {}) {
    const updated: CreateTournamentFormInput = {
      name,
      game,
      location,
      description,
      detail,
      ...overrides,
    };
    onValuesChange(updated);
  }

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setApiError(null);

    const payload = {
      name,
      game,
      location,
      description,
      detail,
    };

    const result = createTournamentFormSchema.safeParse(payload);
    if (!result.success) {
      const flattened = result.error.flatten().fieldErrors;
      const mappedErrors: FormErrors = {};
      for (const [key, messages] of Object.entries(flattened)) {
        if (messages?.[0]) {
          mappedErrors[key as keyof CreateTournamentFormInput] = messages[0];
        }
      }
      setFieldErrors(mappedErrors);
      return;
    }

    setFieldErrors({});

    if (!authorizationHeader) {
      setApiError("Authentication token expired. Please sign in again.");
      return;
    }

    setIsSubmitting(true);
    try {
      const { data, error } = await apiClient.POST("/api/v1/tournaments", {
        headers: { Authorization: authorizationHeader },
        body: {
          name: result.data.name,
          game: result.data.game,
          location: result.data.location,
          description: result.data.description,
          detail: result.data.detail || undefined,
        },
      });

      if (error || !data) {
        setApiError(
          extractProblemMessage(error, "Failed to create tournament.")
        );
        return;
      }

      router.push(`/organizer/tournaments/${data.id}`);
    } catch {
      setApiError("Connection failure. Please check your network and retry.");
    } finally {
      setIsSubmitting(false);
    }
  }

  return (
    <form onSubmit={handleSubmit} noValidate className="flex flex-col gap-6">
      {apiError ? (
        <div
          role="alert"
          className="bg-destructive/10 text-destructive border-destructive/20 flex items-center gap-3 rounded-lg border p-3 text-sm"
        >
          <AlertCircle className="size-4 shrink-0" />
          <span>{apiError}</span>
        </div>
      ) : null}

      <div className="bg-muted/40 text-muted-foreground border-border flex items-start gap-3 rounded-lg border p-4 text-xs leading-relaxed sm:text-sm">
        <FileEdit className="text-primary mt-0.5 size-4 shrink-0" />
        <div>
          <span className="text-foreground font-medium">
            Draft State Initialization:
          </span>{" "}
          Your tournament will be created in Draft state. You can configure
          registration rules, entry fees, tournament dates, and crowdfunding
          campaigns directly from your workspace after creation.
        </div>
      </div>

      <HostTournamentGeneralCard
        organizerName={organizerName}
        name={name}
        onNameChange={(val) => {
          setName(val);
          setFieldErrors((prev) => ({ ...prev, name: undefined }));
          updateAndNotify({ name: val });
        }}
        game={game}
        onGameChange={(val) => {
          setGame(val);
          setFieldErrors((prev) => ({ ...prev, game: undefined }));
          updateAndNotify({ game: val });
        }}
        location={location}
        onLocationChange={(val) => {
          setLocation(val);
          setFieldErrors((prev) => ({ ...prev, location: undefined }));
          updateAndNotify({ location: val });
        }}
        description={description}
        onDescriptionChange={(val) => {
          setDescription(val);
          setFieldErrors((prev) => ({ ...prev, description: undefined }));
          updateAndNotify({ description: val });
        }}
        detail={detail}
        onDetailChange={(val) => {
          setDetail(val);
          setFieldErrors((prev) => ({ ...prev, detail: undefined }));
          updateAndNotify({ detail: val });
        }}
        errors={fieldErrors}
      />

      {/* Form Submission Actions */}
      <div className="flex flex-col-reverse gap-3 pt-2 sm:flex-row sm:items-center sm:justify-end">
        <Button
          type="button"
          variant="ghost"
          onClick={() => router.push("/organizer")}
          disabled={isSubmitting}
        >
          Cancel
        </Button>

        <Button type="submit" disabled={isSubmitting} className="min-w-44">
          {isSubmitting ? (
            <>
              <Spinner data-icon="inline-start" />
              Creating Tournament...
            </>
          ) : (
            <>
              <PlusCircle data-icon="inline-start" />
              Create Tournament
            </>
          )}
        </Button>
      </div>
    </form>
  );
}
