"use client";

import { useMemo, useState } from "react";
import { Check, Settings2, Trash2, Trophy } from "lucide-react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs";
import { Spinner } from "@/components/ui/spinner";
import type { MatchResponse } from "../utils";
import type { BracketMatchNode, TeamPlacementInput } from "../schemas";

interface MatchDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  match: MatchResponse | null;
  defaultStageLevel?: number;
  allMatches: BracketMatchNode[];
  onSaveNode: (node: BracketMatchNode) => void;
  onDeleteNode?: (tempId: string) => void;
  onRecordScores?: (matchId: string, placements: TeamPlacementInput[]) => Promise<void>;
  isRecording?: boolean;
}

export function MatchDialog({
  open,
  onOpenChange,
  match,
  defaultStageLevel = 1,
  allMatches,
  onSaveNode,
  onDeleteNode,
  onRecordScores,
  isRecording = false,
}: MatchDialogProps) {
  if (!open) return null;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md max-h-[90vh] overflow-y-auto">
        <MatchDialogContent
          key={match ? match.id : `new-stage-${defaultStageLevel}`}
          match={match}
          defaultStageLevel={defaultStageLevel}
          allMatches={allMatches}
          onClose={() => onOpenChange(false)}
          onSaveNode={onSaveNode}
          onDeleteNode={onDeleteNode}
          onRecordScores={onRecordScores}
          isRecording={isRecording}
        />
      </DialogContent>
    </Dialog>
  );
}

interface MatchDialogContentProps {
  match: MatchResponse | null;
  defaultStageLevel: number;
  allMatches: BracketMatchNode[];
  onClose: () => void;
  onSaveNode: (node: BracketMatchNode) => void;
  onDeleteNode?: (tempId: string) => void;
  onRecordScores?: (matchId: string, placements: TeamPlacementInput[]) => Promise<void>;
  isRecording: boolean;
}

function MatchDialogContent({
  match,
  defaultStageLevel,
  allMatches,
  onClose,
  onSaveNode,
  onDeleteNode,
  onRecordScores,
  isRecording,
}: MatchDialogContentProps) {
  // Other matches in bracket (excluding current one being edited)
  const otherMatches = allMatches.filter((m) => m.tempId !== (match?.id ?? ""));
  const maxExistingStage = useMemo(() => {
    if (otherMatches.length === 0) return 0;
    return Math.max(...otherMatches.map((m) => m.stageLevel));
  }, [otherMatches]);

  const maxAllowedStage = Math.max(maxExistingStage + 1, match?.stageLevel ?? 1, 1);
  const allowedStages = useMemo(() => {
    return Array.from({ length: maxAllowedStage }, (_, i) => i + 1);
  }, [maxAllowedStage]);

  const initialStageLevel = match
    ? match.stageLevel
    : Math.min(defaultStageLevel, maxAllowedStage);

  // Initialize state directly from props (no useEffect needed)
  const [name, setName] = useState(
    match ? match.name : `Match ${allMatches.length + 1}`
  );
  const [venue, setVenue] = useState(match ? match.venue : "Main Arena");
  const [stageLevel, setStageLevel] = useState(initialStageLevel);
  const [teamCount, setTeamCount] = useState(match ? match.teamCount : 2);
  const [advancingCount, setAdvancingCount] = useState(
    match ? (match.advancingCount || 1) : 1
  );
  const [targetTempId, setTargetTempId] = useState(match?.targetMatchId ?? "");

  const stageOptions = useMemo(
    () =>
      allowedStages.map((lvl) => ({
        value: String(lvl),
        label: `Stage ${lvl}${lvl === maxExistingStage + 1 && maxExistingStage > 0 ? " (+ New Stage)" : ""}`,
      })),
    [allowedStages, maxExistingStage]
  );

  const entries = match?.entries ?? [];
  const hasEntries = entries.length > 0;
  const isCompleted = match?.isCompleted ?? false;

  const [placements, setPlacements] = useState<Record<string, { rank: number; score: string }>>(() => {
    const initial: Record<string, { rank: number; score: string }> = {};
    entries.forEach((e, idx) => {
      initial[e.teamId] = {
        rank: e.finalRank ?? idx + 1,
        score: e.score ?? "",
      };
    });
    return initial;
  });

  const [activeTab, setActiveTab] = useState<string>(hasEntries ? "scores" : "settings");
  const [errorMsg, setErrorMsg] = useState<string | null>(null);

  // Available target matches at higher stages (sorted by stage level, then name)
  const availableTargets = allMatches
    .filter((m) => m.tempId !== (match?.id ?? "") && m.stageLevel > stageLevel)
    .sort((a, b) => a.stageLevel - b.stageLevel || a.name.localeCompare(b.name));

  const targetOptions = useMemo(
    () => [
      { value: "none", label: "None" },
      ...availableTargets.map((t) => ({
        value: t.tempId,
        label: `Stage ${t.stageLevel}: ${t.name}`,
      })),
    ],
    [availableTargets]
  );

  function handleScoreChange(teamId: string, val: string) {
    setPlacements((prev) => ({
      ...prev,
      [teamId]: { ...prev[teamId], score: val },
    }));
  }

  function handleRankChange(teamId: string, val: number) {
    setPlacements((prev) => ({
      ...prev,
      [teamId]: { ...prev[teamId], rank: val },
    }));
  }

  function handleSaveMatchConfig(e: React.FormEvent) {
    e.preventDefault();
    if (!name.trim()) {
      setErrorMsg("Match name cannot be empty.");
      return;
    }
    if (!venue.trim()) {
      setErrorMsg("Venue cannot be empty.");
      return;
    }
    if (teamCount < 1) {
      setErrorMsg("Team count must be at least 1.");
      return;
    }
    if (advancingCount > teamCount) {
      setErrorMsg("Advancing count cannot exceed team count.");
      return;
    }
    if (targetTempId && advancingCount < 1) {
      setErrorMsg("A match targeting another match must advance at least 1 team.");
      return;
    }

    const node: BracketMatchNode = {
      tempId: match ? match.id : `match-${Date.now()}`,
      name: name.trim(),
      venue: venue.trim(),
      stageLevel: Number(stageLevel),
      teamCount: Number(teamCount),
      advancingCount: Number(advancingCount),
      targetTempId: targetTempId.trim() || null,
    };

    onSaveNode(node);
    onClose();
  }

  async function handleRecordScoresSubmit() {
    if (!match || !onRecordScores) return;

    const items: TeamPlacementInput[] = entries.map((e) => ({
      teamId: e.teamId,
      finalRank: placements[e.teamId]?.rank ?? 1,
      score: placements[e.teamId]?.score?.trim() || undefined,
    }));

    const ranks = items.map((x) => x.finalRank);
    if (new Set(ranks).size !== ranks.length) {
      setErrorMsg("Each participating team must have a unique final rank.");
      return;
    }

    try {
      setErrorMsg(null);
      await onRecordScores(match.id, items);
      onClose();
    } catch (err: unknown) {
      setErrorMsg(err instanceof Error ? err.message : "Failed to record match scores.");
    }
  }

  // Brand new match being created: show dedicated format form
  if (!match) {
    return (
      <>
        <DialogHeader>
          <div className="flex items-center gap-2 text-primary">
            <Settings2 className="size-5" />
            <DialogTitle className="text-lg font-bold">Add New Match</DialogTitle>
          </div>
          <DialogDescription>
            Configure stage level and advancement rules for the new match.
          </DialogDescription>
        </DialogHeader>

        {errorMsg ? (
          <div className="rounded-md bg-destructive/15 p-3 text-sm text-destructive font-medium">
            {errorMsg}
          </div>
        ) : null}

        <form onSubmit={handleSaveMatchConfig} className="flex flex-col gap-4 pt-1">
          <div className="grid grid-cols-2 gap-3">
            <div className="col-span-2">
              <Label htmlFor="match-name" className="text-xs">
                Match Name
              </Label>
              <Input
                id="match-name"
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="e.g. Match 1, Final"
                className="h-8 text-sm mt-1"
              />
            </div>

            <div>
              <Label htmlFor="match-venue" className="text-xs">
                Venue / Station
              </Label>
              <Input
                id="match-venue"
                value={venue}
                onChange={(e) => setVenue(e.target.value)}
                placeholder="e.g. Arena A, Court 1"
                className="h-8 text-sm mt-1"
              />
            </div>

            <div>
              <Label htmlFor="match-stage" className="text-xs">
                Stage Level
              </Label>
              <Select
                items={stageOptions}
                value={String(stageLevel)}
                onValueChange={(val) => {
                  if (!val) return;
                  setStageLevel(Number(val));
                  setTargetTempId("");
                }}
              >
                <SelectTrigger id="match-stage" size="sm" className="w-full mt-1">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent className="z-60">
                  <SelectGroup>
                    {stageOptions.map((opt) => (
                      <SelectItem key={opt.value} value={opt.value}>
                        {opt.label}
                      </SelectItem>
                    ))}
                  </SelectGroup>
                </SelectContent>
              </Select>
            </div>

            <div>
              <Label htmlFor="match-team-count" className="text-xs">
                Team Count (Hold)
              </Label>
              <Input
                id="match-team-count"
                type="number"
                min={1}
                value={teamCount}
                onChange={(e) => setTeamCount(Number(e.target.value))}
                className="h-8 text-sm mt-1"
              />
            </div>

            <div>
              <Label htmlFor="match-advancing-count" className="text-xs">
                Advancing Count (Top N)
              </Label>
              <Input
                id="match-advancing-count"
                type="number"
                min={0}
                max={teamCount}
                value={advancingCount}
                onChange={(e) => setAdvancingCount(Number(e.target.value))}
                className="h-8 text-sm mt-1"
              />
            </div>

            <div className="col-span-2">
              <Label htmlFor="match-target" className="text-xs">
                Target Match (Advances to)
              </Label>
              <Select
                items={targetOptions}
                value={targetTempId || "none"}
                onValueChange={(val) =>
                  setTargetTempId(val === "none" || !val ? "" : (val as string))
                }
              >
                <SelectTrigger id="match-target" size="sm" className="w-full mt-1">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent className="z-60">
                  <SelectGroup>
                    {targetOptions.map((opt) => (
                      <SelectItem key={opt.value} value={opt.value}>
                        {opt.label}
                      </SelectItem>
                    ))}
                  </SelectGroup>
                </SelectContent>
              </Select>
            </div>
          </div>

          <DialogFooter className="mt-4 flex sm:justify-between items-center gap-2">
            <Button type="button" variant="outline" size="sm" onClick={onClose}>
              Cancel
            </Button>
            <Button type="submit" size="sm">
              Create Match
            </Button>
          </DialogFooter>
        </form>
      </>
    );
  }

  // Existing match being edited: 2 tabs (Scores & Results vs Format Settings)
  return (
    <>
      <DialogHeader>
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2 text-primary">
            <Trophy className="size-5" />
            <DialogTitle className="text-lg font-bold">{match.name}</DialogTitle>
          </div>
          {isCompleted ? (
            <span className="text-[10px] uppercase font-bold text-emerald-600 dark:text-emerald-400 bg-emerald-500/15 px-2 py-0.5 rounded">
              Completed
            </span>
          ) : (
            <span className="text-[10px] uppercase font-bold text-muted-foreground bg-muted px-2 py-0.5 rounded">
              Pending
            </span>
          )}
        </div>
        <DialogDescription>
          Stage {stageLevel} • {venue || "Main Arena"} • {entries.length}/{teamCount} Teams
        </DialogDescription>
      </DialogHeader>

      {errorMsg ? (
        <div className="rounded-md bg-destructive/15 p-3 text-sm text-destructive font-medium">
          {errorMsg}
        </div>
      ) : null}

      <Tabs value={activeTab} onValueChange={(val) => setActiveTab(val as string)} className="gap-3">
        <TabsList className="grid w-full grid-cols-2">
          <TabsTrigger value="scores" className="gap-1.5 text-xs">
            <Trophy className="size-3.5" />
            Scores & Results
          </TabsTrigger>
          <TabsTrigger value="settings" className="gap-1.5 text-xs">
            <Settings2 className="size-3.5" />
            Format Settings
          </TabsTrigger>
        </TabsList>

        {/* Tab 1: Scores & Results */}
        <TabsContent value="scores" className="flex flex-col gap-3 pt-1">
          {hasEntries ? (
            <>
              <div className="flex items-center justify-between text-xs text-muted-foreground">
                <span>Participating Teams ({entries.length}/{teamCount})</span>
                <span>{advancingCount > 0 ? `Top ${advancingCount} advance` : "Championship"}</span>
              </div>

              <div className="flex flex-col gap-2 max-h-60 overflow-y-auto pr-1">
                {entries.map((entry) => {
                  const teamName = entry.teamName || `Team (${entry.teamId.slice(0, 6)})`;
                  const current = placements[entry.teamId] || { rank: 1, score: "" };

                  return (
                    <div
                      key={entry.teamId}
                      className="flex flex-col gap-2 rounded-lg border border-border p-3 bg-muted/30"
                    >
                      <div className="flex items-center justify-between">
                        <span className="font-semibold text-sm truncate">{teamName}</span>
                        {isCompleted && entry.advances ? (
                          <span className="text-[9px] uppercase font-bold text-emerald-600 dark:text-emerald-400 bg-emerald-500/15 px-1.5 py-0.5 rounded">
                            Advanced
                          </span>
                        ) : null}
                      </div>
                      <div className="grid grid-cols-2 gap-3">
                        <div>
                          <Label className="text-xs">Score / Result</Label>
                          <Input
                            placeholder="e.g. 2-1"
                            value={current.score}
                            onChange={(e) => handleScoreChange(entry.teamId, e.target.value)}
                            className="h-8 text-sm mt-1"
                          />
                        </div>
                        <div>
                          <Label className="text-xs">Final Rank (1 = Winner)</Label>
                          <Input
                            type="number"
                            min={1}
                            max={entries.length}
                            value={current.rank}
                            onChange={(e) => handleRankChange(entry.teamId, Number(e.target.value))}
                            className="h-8 text-sm mt-1"
                          />
                        </div>
                      </div>
                    </div>
                  );
                })}

                {/* Remaining TBD Slots */}
                {Array.from({ length: Math.max(0, teamCount - entries.length) }).map((_, idx) => (
                  <div
                    key={`tbd-slot-${idx}`}
                    className="flex items-center justify-between px-3 py-2 rounded-lg border border-border/50 bg-muted/20 text-xs text-muted-foreground"
                  >
                    <span className="font-mono text-xs">Slot {entries.length + idx + 1}</span>
                    <span className="text-xs text-muted-foreground/60">TBD (Waiting for qualifier)</span>
                  </div>
                ))}
              </div>

              <DialogFooter className="mt-4 flex sm:justify-between items-center gap-2">
                <Button type="button" variant="outline" size="sm" onClick={onClose}>
                  Close
                </Button>
                <Button
                  size="sm"
                  onClick={handleRecordScoresSubmit}
                  disabled={isRecording}
                  className="gap-1.5"
                >
                  {isRecording ? <Spinner /> : <Check className="size-4" />}
                  {isCompleted ? "Update Scores & Placements" : "Confirm Scores & Advance Winner"}
                </Button>
              </DialogFooter>
            </>
          ) : (
            <div className="flex flex-col gap-3 py-6 text-center items-center justify-center border border-dashed border-border rounded-lg p-6 bg-muted/10">
              <div className="p-3 rounded-full bg-muted">
                <Trophy className="size-6 text-muted-foreground" />
              </div>
              <div className="space-y-1">
                <p className="text-sm font-semibold text-foreground">No Teams Seeded Yet</p>
                <p className="text-xs text-muted-foreground max-w-xs">
                  Teams will enter this match as previous stages finish or when seeded.
                </p>
              </div>
              <Button
                variant="secondary"
                size="sm"
                className="text-xs mt-2"
                onClick={() => setActiveTab("settings")}
              >
                Edit Format Settings
              </Button>
              <DialogFooter className="mt-4 w-full">
                <Button type="button" variant="outline" size="sm" onClick={onClose} className="w-full">
                  Close
                </Button>
              </DialogFooter>
            </div>
          )}
        </TabsContent>

        {/* Tab 2: Format Settings */}
        <TabsContent value="settings" className="pt-1">
          <form onSubmit={handleSaveMatchConfig} className="flex flex-col gap-4">
            <div className="grid grid-cols-2 gap-3">
              <div className="col-span-2">
                <Label htmlFor="match-name" className="text-xs">
                  Match Name
                </Label>
                <Input
                  id="match-name"
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  placeholder="e.g. Match 1, Final"
                  className="h-8 text-sm mt-1"
                  disabled={isCompleted}
                />
              </div>

              <div>
                <Label htmlFor="match-venue" className="text-xs">
                  Venue / Station
                </Label>
                <Input
                  id="match-venue"
                  value={venue}
                  onChange={(e) => setVenue(e.target.value)}
                  placeholder="e.g. Arena A, Court 1"
                  className="h-8 text-sm mt-1"
                  disabled={isCompleted}
                />
              </div>

              <div>
                <Label htmlFor="match-stage" className="text-xs">
                  Stage Level
                </Label>
                <Select
                  items={stageOptions}
                  value={String(stageLevel)}
                  onValueChange={(val) => {
                    if (!val) return;
                    setStageLevel(Number(val));
                    setTargetTempId("");
                  }}
                  disabled={isCompleted}
                >
                  <SelectTrigger id="match-stage" size="sm" className="w-full mt-1">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent className="z-60">
                    <SelectGroup>
                      {stageOptions.map((opt) => (
                        <SelectItem key={opt.value} value={opt.value}>
                          {opt.label}
                        </SelectItem>
                      ))}
                    </SelectGroup>
                  </SelectContent>
                </Select>
              </div>

              <div>
                <Label htmlFor="match-team-count" className="text-xs">
                  Team Count (Hold)
                </Label>
                <Input
                  id="match-team-count"
                  type="number"
                  min={1}
                  value={teamCount}
                  onChange={(e) => setTeamCount(Number(e.target.value))}
                  className="h-8 text-sm mt-1"
                  disabled={isCompleted}
                />
              </div>

              <div>
                <Label htmlFor="match-advancing-count" className="text-xs">
                  Advancing Count (Top N)
                </Label>
                <Input
                  id="match-advancing-count"
                  type="number"
                  min={0}
                  max={teamCount}
                  value={advancingCount}
                  onChange={(e) => setAdvancingCount(Number(e.target.value))}
                  className="h-8 text-sm mt-1"
                  disabled={isCompleted}
                />
              </div>

              <div className="col-span-2">
                <Label htmlFor="match-target" className="text-xs">
                  Target Match (Advances to)
                </Label>
                <Select
                  items={targetOptions}
                  value={targetTempId || "none"}
                  onValueChange={(val) =>
                    setTargetTempId(val === "none" || !val ? "" : (val as string))
                  }
                  disabled={isCompleted}
                >
                  <SelectTrigger id="match-target" size="sm" className="w-full mt-1">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent className="z-60">
                    <SelectGroup>
                      {targetOptions.map((opt) => (
                        <SelectItem key={opt.value} value={opt.value}>
                          {opt.label}
                        </SelectItem>
                      ))}
                    </SelectGroup>
                  </SelectContent>
                </Select>
              </div>
            </div>

            <DialogFooter className="mt-4 flex sm:justify-between items-center gap-2">
              {match && !hasEntries && onDeleteNode ? (
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  className="text-xs text-destructive hover:bg-destructive/10 h-8 px-2 gap-1"
                  onClick={() => {
                    onDeleteNode(match.id);
                    onClose();
                  }}
                >
                  <Trash2 className="size-3.5" />
                  Delete Match
                </Button>
              ) : (
                <div />
              )}
              <div className="flex items-center gap-2">
                <Button type="button" variant="outline" size="sm" onClick={onClose}>
                  Close
                </Button>
                {!isCompleted ? (
                  <Button type="submit" size="sm">
                    Save Settings
                  </Button>
                ) : null}
              </div>
            </DialogFooter>
          </form>
        </TabsContent>
      </Tabs>
    </>
  );
}
