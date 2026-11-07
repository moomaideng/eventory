"use client";

import { useState } from "react";
import { ArrowLeftRight } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Field, FieldDescription, FieldGroup, FieldLabel } from "@/components/ui/field";
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Spinner } from "@/components/ui/spinner";
import type { SwapSlot } from "../hooks/use-bracket";
import type { MatchResponse } from "../utils";

interface MatchSwapFormProps {
  match: MatchResponse;
  bracketMatches: MatchResponse[];
  onSwap: (first: SwapSlot, second: SwapSlot) => Promise<void>;
  isSwapping: boolean;
  onClose: () => void;
}

const teamLabel = (name: string | undefined, teamId: string) => name || `Team (${teamId.slice(0, 6)})`;

export function MatchSwapForm({ match, bracketMatches, onSwap, isSwapping, onClose }: MatchSwapFormProps) {
  const ownOptions = (match.entries ?? []).map((e) => ({ value: e.teamId, label: teamLabel(e.teamName, e.teamId) }));
  // Only teams in other unfinished matches of the same stage can be swapped in.
  const otherOptions = bracketMatches
    .filter((m) => m.id !== match.id && !m.isCompleted && m.stageLevel === match.stageLevel)
    .flatMap((m) =>
      (m.entries ?? []).map((e) => ({
        value: `${m.id}:${e.teamId}`,
        label: `${teamLabel(e.teamName, e.teamId)} (${m.name})`,
      }))
    );

  const [ownTeam, setOwnTeam] = useState("");
  const [otherSlot, setOtherSlot] = useState("");
  const [error, setError] = useState<string | null>(null);

  async function handleSwap() {
    const [otherMatchId, otherTeamId] = otherSlot.split(":");
    try {
      setError(null);
      await onSwap({ matchId: match.id, teamId: ownTeam }, { matchId: otherMatchId, teamId: otherTeamId });
      onClose();
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : "Failed to swap teams.");
    }
  }

  return (
    <div className="flex flex-col gap-4">
      {error ? (
        <div className="rounded-md bg-destructive/15 p-3 text-sm text-destructive font-medium">{error}</div>
      ) : null}

      <FieldGroup>
        <Field>
          <FieldLabel htmlFor="swap-own-team">Team in this match</FieldLabel>
          <Select items={ownOptions} value={ownTeam} onValueChange={(v) => setOwnTeam((v as string) ?? "")}>
            <SelectTrigger id="swap-own-team" size="sm" className="w-full">
              <SelectValue placeholder="Select team" />
            </SelectTrigger>
            <SelectContent className="z-60">
              <SelectGroup>
                {ownOptions.map((o) => (
                  <SelectItem key={o.value} value={o.value}>
                    {o.label}
                  </SelectItem>
                ))}
              </SelectGroup>
            </SelectContent>
          </Select>
        </Field>

        <Field>
          <FieldLabel htmlFor="swap-other-team">Swap with</FieldLabel>
          <Select items={otherOptions} value={otherSlot} onValueChange={(v) => setOtherSlot((v as string) ?? "")}>
            <SelectTrigger id="swap-other-team" size="sm" className="w-full">
              <SelectValue placeholder="Select team" />
            </SelectTrigger>
            <SelectContent className="z-60">
              <SelectGroup>
                {otherOptions.map((o) => (
                  <SelectItem key={o.value} value={o.value}>
                    {o.label}
                  </SelectItem>
                ))}
              </SelectGroup>
            </SelectContent>
          </Select>
          <FieldDescription>Only teams in unfinished Stage {match.stageLevel} matches.</FieldDescription>
        </Field>
      </FieldGroup>

      <div className="flex justify-end gap-2">
        <Button type="button" variant="outline" size="sm" onClick={onClose}>
          Close
        </Button>
        <Button size="sm" onClick={handleSwap} disabled={!ownTeam || !otherSlot || isSwapping}>
          {isSwapping ? <Spinner /> : <ArrowLeftRight data-icon="inline-start" />}
          Swap Teams
        </Button>
      </div>
    </div>
  );
}
