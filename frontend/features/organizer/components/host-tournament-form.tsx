"use client";

import React, { useState, type FormEvent } from "react";
import { useRouter } from "next/navigation";
import { AlertCircle, PlusCircle } from "lucide-react";
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
import { HostTournamentFormatCard } from "./host-tournament-format-card";
import { HostTournamentScheduleCard } from "./host-tournament-schedule-card";

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
  const [registrationMode, setRegistrationMode] = useState<"SOLO" | "TEAM">(
    "TEAM"
  );
  const [minTeamSize, setMinTeamSize] = useState(5);
  const [maxTeamSize, setMaxTeamSize] = useState(5);
  const [capacity, setCapacity] = useState(16);
  const [entryFee, setEntryFee] = useState(0);
  const [registrationDeadline, setRegistrationDeadline] = useState("");
  const [startAt, setStartAt] = useState("");
  const [endAt, setEndAt] = useState("");

  const [fieldErrors, setFieldErrors] = useState<FormErrors>({});
  const [apiError, setApiError] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  // Sync draft values up to parent workspace for the live preview card
  function updateAndNotify(
    overrides: Partial<CreateTournamentFormInput> = {}
  ) {
    const updated: CreateTournamentFormInput = {
      name,
      game,
      location,
      description,
      registrationMode,
      minTeamSize: registrationMode === "SOLO" ? 1 : minTeamSize,
      maxTeamSize: registrationMode === "SOLO" ? 1 : maxTeamSize,
      capacity,
      entryFee,
      registrationDeadline,
      startAt,
      endAt,
      ...overrides,
    };
    onValuesChange(updated);
  }

  function handleModeChange(nextMode: "SOLO" | "TEAM") {
    setRegistrationMode(nextMode);
    const newMin = nextMode === "SOLO" ? 1 : 5;
    const newMax = nextMode === "SOLO" ? 1 : 5;
    if (nextMode === "SOLO") {
      setMinTeamSize(1);
      setMaxTeamSize(1);
    }
    updateAndNotify({
      registrationMode: nextMode,
      minTeamSize: newMin,
      maxTeamSize: newMax,
    });
  }

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setApiError(null);

    const payload = {
      name,
      game,
      location,
      description,
      registrationMode,
      minTeamSize: registrationMode === "SOLO" ? 1 : minTeamSize,
      maxTeamSize: registrationMode === "SOLO" ? 1 : maxTeamSize,
      capacity,
      entryFee,
      registrationDeadline,
      startAt,
      endAt,
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
          registrationMode: result.data.registrationMode,
          minTeamSize: result.data.minTeamSize,
          maxTeamSize: result.data.maxTeamSize,
          capacity: result.data.capacity,
          entryFee: result.data.entryFee,
          registrationDeadline: new Date(
            result.data.registrationDeadline
          ).toISOString(),
          startAt: new Date(result.data.startAt).toISOString(),
          endAt: new Date(result.data.endAt).toISOString(),
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
          className="bg-destructive/10 text-destructive flex items-center gap-3 rounded-lg border border-destructive/20 p-3 text-sm"
        >
          <AlertCircle className="size-4 shrink-0" />
          <span>{apiError}</span>
        </div>
      ) : null}

      {/* 1. General Information Card (Basic Info + Description & Rules) */}
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
        errors={fieldErrors}
      />

      {/* 2. Format & Capacity Card */}
      <HostTournamentFormatCard
        registrationMode={registrationMode}
        onModeChange={handleModeChange}
        minTeamSize={minTeamSize}
        onMinTeamSizeChange={(val) => {
          setMinTeamSize(val);
          setFieldErrors((prev) => ({ ...prev, minTeamSize: undefined }));
          updateAndNotify({ minTeamSize: val });
        }}
        maxTeamSize={maxTeamSize}
        onMaxTeamSizeChange={(val) => {
          setMaxTeamSize(val);
          setFieldErrors((prev) => ({ ...prev, maxTeamSize: undefined }));
          updateAndNotify({ maxTeamSize: val });
        }}
        capacity={capacity}
        onCapacityChange={(val) => {
          setCapacity(val);
          setFieldErrors((prev) => ({ ...prev, capacity: undefined }));
          updateAndNotify({ capacity: val });
        }}
        entryFee={entryFee}
        onEntryFeeChange={(val) => {
          setEntryFee(val);
          setFieldErrors((prev) => ({ ...prev, entryFee: undefined }));
          updateAndNotify({ entryFee: val });
        }}
        errors={fieldErrors}
      />

      {/* 3. Tournament Schedule Card */}
      <HostTournamentScheduleCard
        registrationDeadline={registrationDeadline}
        onRegistrationDeadlineChange={(val) => {
          setRegistrationDeadline(val);
          setFieldErrors((prev) => ({
            ...prev,
            registrationDeadline: undefined,
          }));
          updateAndNotify({ registrationDeadline: val });
        }}
        startAt={startAt}
        onStartAtChange={(val) => {
          setStartAt(val);
          setFieldErrors((prev) => ({ ...prev, startAt: undefined }));
          updateAndNotify({ startAt: val });
        }}
        endAt={endAt}
        onEndAtChange={(val) => {
          setEndAt(val);
          setFieldErrors((prev) => ({ ...prev, endAt: undefined }));
          updateAndNotify({ endAt: val });
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
              Publishing Tournament...
            </>
          ) : (
            <>
              <PlusCircle data-icon="inline-start" />
              Publish Tournament
            </>
          )}
        </Button>
      </div>
    </form>
  );
}
