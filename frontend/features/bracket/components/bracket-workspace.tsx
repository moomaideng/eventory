"use client";

import { useMemo, useState } from "react";
import {
  AlertTriangle,
  ChevronDown,
  ChevronUp,
  RefreshCw,
  RotateCcw,
  Save,
  Shuffle,
  Trophy,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Spinner } from "@/components/ui/spinner";
import { useBracket } from "../hooks/use-bracket";
import { BracketTreeView } from "./bracket-tree-view";
import { MatchDialog } from "./match-dialog";
import { ResetBracketDialog } from "./reset-bracket-dialog";
import {
  matchResponseToNode,
  validateBracketInvariants,
  type MatchResponse,
} from "../utils";
import type { BracketMatchNode } from "../schemas";

interface BracketWorkspaceProps {
  tournamentId: string;
}

export function BracketWorkspace({ tournamentId }: BracketWorkspaceProps) {
  const [selectedMatch, setSelectedMatch] = useState<MatchResponse | null>(null);
  const [newMatchStageLevel, setNewMatchStageLevel] = useState(1);
  const [isDialogOpen, setIsDialogOpen] = useState(false);
  const [isResetDialogOpen, setIsResetDialogOpen] = useState(false);
  const [showErrorDetails, setShowErrorDetails] = useState(true);
  const [actionError, setActionError] = useState<string | null>(null);

  const {
    matches,
    acceptedCount,
    isLoading,
    isFetching,
    refetch,
    defineBracket,
    isDefining,
    seedStage1,
    isSeeding,
    recordResults,
    isRecording,
  } = useBracket({ tournamentId });

  // Local draft overrides (null means directly in sync with server data)
  const [customNodes, setCustomNodes] = useState<BracketMatchNode[] | null>(null);

  const draftNodes = useMemo(() => {
    if (customNodes !== null) return customNodes;
    return matches.map(matchResponseToNode);
  }, [customNodes, matches]);

  const isDirty = customNodes !== null;

  // Client-side invariant validation (including Stage 1 capacity == accepted teams)
  const validation = useMemo(
    () => validateBracketInvariants(draftNodes, acceptedCount),
    [draftNodes, acceptedCount]
  );

  // Merged matches for visual display (preserving entries & live score results)
  const displayMatches: MatchResponse[] = useMemo(() => {
    const matchById = new Map(matches.map((m) => [m.id, m]));
    return draftNodes.map((node) => {
      const existing = matchById.get(node.tempId);
      return {
        id: node.tempId,
        tournamentId,
        name: node.name,
        venue: node.venue,
        stageLevel: node.stageLevel,
        teamCount: node.teamCount,
        advancingCount: node.advancingCount,
        targetMatchId: node.targetTempId ?? undefined,
        isCompleted: existing?.isCompleted ?? false,
        entries: existing?.entries ?? [],
        createdAt: existing?.createdAt ?? new Date().toISOString(),
        updatedAt: existing?.updatedAt ?? new Date().toISOString(),
      };
    });
  }, [draftNodes, matches, tournamentId]);

  const hasCompletedMatches = useMemo(
    () => matches.some((m) => m.isCompleted),
    [matches]
  );

  const stage1Matches = useMemo(
    () => displayMatches.filter((m) => m.stageLevel === 1),
    [displayMatches]
  );

  const maxStageLevel = useMemo(() => {
    if (draftNodes.length === 0) return 0;
    return Math.max(...draftNodes.map((n) => n.stageLevel));
  }, [draftNodes]);

  function handleOpenNewMatch(stageLevel = 1) {
    setSelectedMatch(null);
    setNewMatchStageLevel(stageLevel);
    setIsDialogOpen(true);
  }

  function handleAddNewStage() {
    const nextLevel = maxStageLevel + 1;
    handleOpenNewMatch(nextLevel);
  }

  function handleSelectMatch(match: MatchResponse) {
    setSelectedMatch(match);
    setIsDialogOpen(true);
  }

  function handleSaveNode(node: BracketMatchNode) {
    const prev = draftNodes;
    const index = prev.findIndex((m) => m.tempId === node.tempId);
    if (index >= 0) {
      const next = [...prev];
      next[index] = node;
      setCustomNodes(next);
    } else {
      setCustomNodes([...prev, node]);
    }
  }

  function handleDeleteNode(tempId: string) {
    const next = draftNodes
      .filter((m) => m.tempId !== tempId)
      .map((m) => (m.targetTempId === tempId ? { ...m, targetTempId: null } : m));
    setCustomNodes(next);
  }

  async function handleSaveLayout() {
    setActionError(null);
    if (!validation.isValid) {
      setActionError("Cannot save: bracket contains invariant validation errors.");
      return;
    }
    try {
      await defineBracket(draftNodes);
      setCustomNodes(null);
    } catch (err: unknown) {
      setActionError(err instanceof Error ? err.message : "Failed to save bracket layout.");
    }
  }

  async function handleSeedTeams() {
    setActionError(null);
    try {
      await seedStage1();
    } catch (err: unknown) {
      setActionError(err instanceof Error ? err.message : "Failed to seed teams into Stage 1.");
    }
  }

  async function handleResetStage1Teams() {
    setActionError(null);
    try {
      // Re-saving current layout recreates matches in DB with empty entries
      await defineBracket(draftNodes);
      setCustomNodes(null);
      await refetch();
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : "Failed to reset Stage 1 teams.";
      setActionError(msg);
      throw err;
    }
  }

  async function handleResetAll() {
    setActionError(null);
    try {
      await defineBracket([]);
      setCustomNodes(null);
      await refetch();
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : "Failed to reset bracket layout.";
      setActionError(msg);
      throw err;
    }
  }

  return (
    <div className="flex flex-col gap-6 w-full">
      {/* Top Header & Actions Bar */}
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4 pb-4 border-b border-border">
        <div className="flex items-center gap-3">
          <div className="size-10 rounded-lg bg-primary/10 text-primary flex items-center justify-center shrink-0">
            <Trophy className="size-5" />
          </div>
          <div>
            <div className="flex items-center gap-2">
              <h2 className="text-xl font-bold tracking-tight">Bracket & Matches</h2>
              {acceptedCount > 0 ? (
                <span className="text-[11px] font-semibold px-2 py-0.5 rounded-full bg-primary/10 text-primary border border-primary/20">
                  {acceptedCount} Accepted Teams
                </span>
              ) : null}
            </div>
            <p className="text-xs text-muted-foreground">
              Add stages, add matches, link winners to the next round, and record scores directly on the tree.
            </p>
          </div>
        </div>

        <div className="flex flex-wrap items-center gap-2">
          {/* Save Layout button (when changed) */}
          {isDirty && !hasCompletedMatches ? (
            <Button
              size="sm"
              variant="default"
              className="text-xs h-8 gap-1.5"
              onClick={handleSaveLayout}
              disabled={isDefining || !validation.isValid}
            >
              {isDefining ? <Spinner /> : <Save className="size-3.5" />}
              Save Layout
            </Button>
          ) : null}

          {/* Seed Stage 1 button */}
          {!isDirty && stage1Matches.length > 0 && !hasCompletedMatches ? (
            <Button
              size="sm"
              variant="outline"
              className="text-xs h-8 gap-1.5"
              onClick={handleSeedTeams}
              disabled={
                isSeeding || (acceptedCount > 0 && validation.stage1SlotsSum !== acceptedCount)
              }
              title={
                acceptedCount > 0 && validation.stage1SlotsSum !== acceptedCount
                  ? `Stage 1 slots (${validation.stage1SlotsSum}) must match accepted teams (${acceptedCount})`
                  : undefined
              }
            >
              {isSeeding ? <Spinner /> : <Shuffle className="size-3.5" />}
              Seed Accepted Teams
            </Button>
          ) : null}

          {/* Reset button (opens Reset Dialog) */}
          {draftNodes.length > 0 && !hasCompletedMatches ? (
            <Button
              variant="outline"
              size="sm"
              className="text-xs h-8 text-muted-foreground hover:text-destructive hover:border-destructive/40 gap-1.5"
              onClick={() => setIsResetDialogOpen(true)}
              title="Reset Stage 1 teams or entire bracket"
            >
              <RotateCcw className="size-3.5" />
              Reset
            </Button>
          ) : null}

          {/* Refresh button */}
          <Button
            variant="outline"
            size="icon"
            className="size-8"
            onClick={() => void refetch()}
            disabled={isFetching}
            title="Refresh bracket from server"
          >
            <RefreshCw className={isFetching ? "size-3.5 animate-spin" : "size-3.5"} />
          </Button>
        </div>
      </div>

      {/* Validation Issues Alert Banner (Expandable) */}
      {!validation.isValid && draftNodes.length > 0 ? (
        <div className="rounded-lg border border-amber-500/30 bg-amber-500/10 p-3.5 text-xs text-amber-900 dark:text-amber-200">
          <div
            className="flex items-center justify-between cursor-pointer font-semibold"
            onClick={() => setShowErrorDetails(!showErrorDetails)}
          >
            <div className="flex items-center gap-2">
              <AlertTriangle className="size-4 text-amber-600 dark:text-amber-400" />
              <span>
                Bracket Invariant Validation: {validation.errors.length} issue(s) need attention before saving
              </span>
            </div>
            {showErrorDetails ? <ChevronUp className="size-4" /> : <ChevronDown className="size-4" />}
          </div>

          {showErrorDetails ? (
            <ul className="list-disc pl-6 pt-2 space-y-1 text-[11px] text-amber-800 dark:text-amber-300">
              {validation.errors.map((err, idx) => (
                <li key={idx}>{err}</li>
              ))}
            </ul>
          ) : null}
        </div>
      ) : null}

      {/* Action Error Banner */}
      {actionError ? (
        <div className="rounded-lg bg-destructive/15 border border-destructive/20 p-3 text-sm text-destructive font-medium flex items-center justify-between">
          <span>{actionError}</span>
          <Button
            variant="ghost"
            size="sm"
            className="h-6 px-2 text-xs text-destructive hover:bg-destructive/10"
            onClick={() => setActionError(null)}
          >
            Dismiss
          </Button>
        </div>
      ) : null}

      {/* Main Single Page: Interactive Symmetrical Bracket Tree */}
      {isLoading ? (
        <div className="flex flex-col items-center justify-center py-24 gap-3 text-muted-foreground text-sm">
          <Spinner className="size-6 text-primary" />
          <span>Loading bracket data...</span>
        </div>
      ) : (
        <BracketTreeView
          matches={displayMatches}
          canEdit={!hasCompletedMatches}
          nodeSlotStatus={validation.nodeSlotStatus}
          onSelectMatch={handleSelectMatch}
          onAddMatchToStage={(stageLevel) => handleOpenNewMatch(stageLevel)}
          onAddStage={handleAddNewStage}
          onAddNewMatch={() => handleOpenNewMatch(1)}
        />
      )}

      {/* Unified Match Modal (Edit details, delete, record scores) */}
      <MatchDialog
        open={isDialogOpen}
        onOpenChange={setIsDialogOpen}
        match={selectedMatch}
        defaultStageLevel={newMatchStageLevel}
        allMatches={draftNodes}
        onSaveNode={handleSaveNode}
        onDeleteNode={handleDeleteNode}
        onRecordScores={recordResults}
        isRecording={isRecording}
      />

      {/* Reset Modal (Reset Stage 1 teams or Reset Entire Bracket) */}
      <ResetBracketDialog
        open={isResetDialogOpen}
        onOpenChange={setIsResetDialogOpen}
        onResetStage1Teams={handleResetStage1Teams}
        onResetAll={handleResetAll}
      />
    </div>
  );
}
