"use client";

import { useMemo, useState } from "react";
import { CheckCircle2, Clock, MapPin, Plus, Trophy } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { cn } from "@/lib/utils";
import type { InflowSlotStatus, MatchResponse } from "../utils";

interface BracketTreeViewProps {
  matches: MatchResponse[];
  canEdit?: boolean;
  nodeSlotStatus?: Record<string, InflowSlotStatus>;
  onSelectMatch?: (match: MatchResponse) => void;
  onAddMatchToStage?: (stageLevel: number) => void;
  onAddStage?: () => void;
  onAddNewMatch?: () => void;
}

interface ConnectorLine {
  id: string;
  d: string;
  isCompleted: boolean;
  advances: boolean;
}

const CARD_HEADER_FOOTER_HEIGHT = 50;
const SLOT_ROW_HEIGHT = 26;
const BASE_GAP = 32;
const MAX_DISPLAY_SLOTS = 4;
const STAGE_WIDTH = 272; // w-68 = 17rem = 272px
const STAGE_GAP = 56; // gap-14 = 3.5rem = 56px
const PADDING_LEFT = 8; // px-2 = 8px
const STAGE_HEADER_OFFSET = 68; // pt-2 (8px) + header (40px) + gap-5 (20px) = 68px

function getCardHeight(teamCount: number): number {
  const totalSlots = Math.max(2, teamCount);
  const rows = totalSlots <= MAX_DISPLAY_SLOTS ? totalSlots : MAX_DISPLAY_SLOTS;
  return CARD_HEADER_FOOTER_HEIGHT + rows * SLOT_ROW_HEIGHT;
}

export function BracketTreeView({
  matches,
  canEdit = true,
  nodeSlotStatus = {},
  onSelectMatch,
  onAddMatchToStage,
  onAddStage,
  onAddNewMatch,
}: BracketTreeViewProps) {
  const [viewingMatch, setViewingMatch] = useState<MatchResponse | null>(null);

  function handleCardClick(m: MatchResponse) {
    if (onSelectMatch) {
      onSelectMatch(m);
    } else {
      setViewingMatch(m);
    }
  }

  // Group matches by StageLevel
  const stages = useMemo(() => {
    const stageMap = new Map<number, MatchResponse[]>();
    for (const m of matches) {
      const list = stageMap.get(m.stageLevel) ?? [];
      list.push(m);
      stageMap.set(m.stageLevel, list);
    }
    return Array.from(stageMap.entries()).sort(([a], [b]) => a - b);
  }, [matches]);

  const maxStageLevel = useMemo(() => {
    if (stages.length === 0) return 0;
    return Math.max(...stages.map(([lvl]) => lvl));
  }, [stages]);

  const stageIndexMap = useMemo(() => {
    const map = new Map<number, number>();
    stages.forEach(([lvl], idx) => {
      map.set(lvl, idx);
    });
    return map;
  }, [stages]);

  // 1. Order matches backwards from the highest stage down to Stage 1.
  // Feeder matches in Stage S are sorted based on the position of their target in Stage S+N.
  // This automatically clusters feeders opposite their targets, completely eliminating criss-crossing lines.
  const orderedStages = useMemo(() => {
    const matchOrderIndex = new Map<string, number>();

    // Process from highest stage down to Stage 1
    const reversedStages = [...stages].sort(([a], [b]) => b - a);
    const resultStagesMap = new Map<number, MatchResponse[]>();

    for (const [lvl, stageMatches] of reversedStages) {
      const sorted = [...stageMatches].sort((a, b) => {
        const targetA = a.targetMatchId ? matchOrderIndex.get(a.targetMatchId) : undefined;
        const targetB = b.targetMatchId ? matchOrderIndex.get(b.targetMatchId) : undefined;

        if (targetA !== undefined && targetB !== undefined) {
          if (targetA !== targetB) return targetA - targetB;
          return stageMatches.indexOf(a) - stageMatches.indexOf(b);
        }
        if (targetA !== undefined) return -1;
        if (targetB !== undefined) return 1;

        return stageMatches.indexOf(a) - stageMatches.indexOf(b);
      });

      sorted.forEach((m, idx) => {
        matchOrderIndex.set(m.id, idx);
      });

      resultStagesMap.set(lvl, sorted);
    }

    return stages.map(([lvl]) => [lvl, resultStagesMap.get(lvl) ?? []] as const);
  }, [stages]);

  // Symmetrical bracket alignment: center Stage N cards relative to incoming feeder matches
  // Symmetrical bracket alignment: center Stage N cards relative to incoming feeder matches
  const matchPositions = useMemo(() => {
    const positions: Record<string, number> = {};

    // Map: targetMatchId -> feeder matches
    const incomingMap = new Map<string, MatchResponse[]>();
    for (const m of matches) {
      if (m.targetMatchId) {
        const list = incomingMap.get(m.targetMatchId) ?? [];
        list.push(m);
        incomingMap.set(m.targetMatchId, list);
      }
    }

    // Traverse stages in ascending order (Stage 1 -> Stage 2 -> Stage 3...)
    for (const [stageLevel, stageMatches] of orderedStages) {
      let prevBottom = 0;

      stageMatches.forEach((m) => {
        const cardH = getCardHeight(m.teamCount);

        if (stageLevel === 1) {
          // Stage 1: stacked according to orderedStages with dynamic heights
          positions[m.id] = prevBottom;
          prevBottom = prevBottom + cardH + BASE_GAP;
        } else {
          // Stage > 1: center dynamically between incoming feeders
          const feeders = incomingMap.get(m.id) ?? [];
          if (feeders.length > 0) {
            const sumCenter = feeders.reduce((acc, f) => {
              const fY = positions[f.id] ?? 0;
              const fH = getCardHeight(f.teamCount);
              return acc + fY + fH / 2;
            }, 0);
            const idealCenter = sumCenter / feeders.length;
            const idealTop = idealCenter - cardH / 2;

            // Ensure no overlap with previous card in this stage
            const y = Math.max(idealTop, prevBottom);
            positions[m.id] = y;
            prevBottom = y + cardH + BASE_GAP;
          } else {
            // Unconnected card: place below previous card
            positions[m.id] = prevBottom;
            prevBottom = prevBottom + cardH + BASE_GAP;
          }
        }
      });
    }

    return positions;
  }, [orderedStages, matches]);

  const maxTrackHeight = useMemo(() => {
    const matchMap = new Map(matches.map((m) => [m.id, m]));
    let maxBottom = 420;
    for (const [id, y] of Object.entries(matchPositions)) {
      const match = matchMap.get(id);
      const h = match ? getCardHeight(match.teamCount) : 104;
      if (y + h + 80 > maxBottom) {
        maxBottom = y + h + 80;
      }
    }
    return maxBottom;
  }, [matchPositions, matches]);

  // Pure mathematical SVG bracket connector lines (instant & glitch-free)
  const lines = useMemo(() => {
    const matchMap = new Map(matches.map((m) => [m.id, m]));
    const result: ConnectorLine[] = [];

    for (const m of matches) {
      if (!m.targetMatchId) continue;
      const target = matchMap.get(m.targetMatchId);
      if (!target) continue;

      const s1 = stageIndexMap.get(m.stageLevel);
      const s2 = stageIndexMap.get(target.stageLevel);
      if (s1 === undefined || s2 === undefined) continue;

      const y1Pos = matchPositions[m.id];
      const y2Pos = matchPositions[target.id];
      if (y1Pos === undefined || y2Pos === undefined) continue;

      const h1 = getCardHeight(m.teamCount);
      const h2 = getCardHeight(target.teamCount);

      const x1 = PADDING_LEFT + s1 * (STAGE_WIDTH + STAGE_GAP) + STAGE_WIDTH;
      const y1 = STAGE_HEADER_OFFSET + y1Pos + h1 / 2;

      const x2 = PADDING_LEFT + s2 * (STAGE_WIDTH + STAGE_GAP);
      const y2 = STAGE_HEADER_OFFSET + y2Pos + h2 / 2;

      const midX = x1 + (x2 - x1) / 2;
      const d = `M ${x1} ${y1} H ${midX} V ${y2} H ${x2}`;

      result.push({
        id: `${m.id}-${m.targetMatchId}`,
        d,
        isCompleted: Boolean(m.isCompleted),
        advances: (m.entries ?? []).some((e) => e.advances),
      });
    }

    return result;
  }, [matches, matchPositions, stageIndexMap]);

  function getStageTitle(stageLevel: number) {
    return `Stage ${stageLevel}`;
  }

  if (matches.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center rounded-xl border border-dashed border-border py-16 px-4 text-center bg-card">
        <div className="size-12 rounded-full bg-primary/10 text-primary flex items-center justify-center mb-3">
          <Trophy className="size-6" />
        </div>
        <h3 className="text-lg font-bold">
          {canEdit ? "No Bracket Tree Configured" : "No Bracket Available"}
        </h3>
        <p className="text-sm text-muted-foreground max-w-sm mt-1 mb-4">
          {canEdit
            ? "Start building your bracket tree by adding matches to Stage 1."
            : "The tournament bracket has not been generated yet. It will appear here once the organizer starts the tournament."}
        </p>

        {canEdit && onAddNewMatch ? (
          <div className="flex items-center gap-3">
            <Button onClick={onAddNewMatch} className="gap-2">
              <Plus className="size-4" />
              Add Match to Stage 1
            </Button>
          </div>
        ) : null}
      </div>
    );
  }

  const svgWidth = Math.max(100, PADDING_LEFT + stages.length * (STAGE_WIDTH + STAGE_GAP) + 300);
  const svgHeight = maxTrackHeight + STAGE_HEADER_OFFSET + 40;

  return (
    <div className="relative w-full overflow-x-auto pb-10 pt-2">
      {/* SVG Connector Lines Overlay */}
      <svg
        style={{ width: `${svgWidth}px`, height: `${svgHeight}px` }}
        className="pointer-events-none absolute top-0 left-0 overflow-visible z-0"
        aria-hidden="true"
      >
        {lines.map((line) => (
          <path
            key={line.id}
            d={line.d}
            fill="none"
            stroke={line.isCompleted ? "var(--primary)" : "currentColor"}
            strokeWidth="1.5"
            strokeOpacity={line.isCompleted ? 0.9 : 0.3}
            strokeLinejoin="round"
            strokeDasharray={line.isCompleted ? undefined : "3 3"}
            className="transition-all duration-200"
          />
        ))}
      </svg>

      {/* Horizontal Stages Track */}
      <div className="relative z-10 flex items-start gap-14 min-w-max px-2">
        {orderedStages.map(([stageLevel, stageMatches]) => {
          const stageTitle = getStageTitle(stageLevel);

          return (
            <div key={stageLevel} className="flex flex-col gap-5 w-68 shrink-0">
              {/* Stage Pill Header with inline + Match button */}
              <div className="flex items-center justify-between h-9 px-3 rounded-lg bg-muted/60 border border-border/80 text-foreground text-xs font-semibold tracking-wide uppercase shadow-2xs">
                <span>{stageTitle}</span>
                {canEdit && onAddMatchToStage ? (
                  <Button
                    variant="ghost"
                    size="sm"
                    className="h-6 px-2 text-[11px] hover:bg-background hover:text-foreground text-muted-foreground gap-1 border border-border/50 shadow-2xs"
                    onClick={() => onAddMatchToStage(stageLevel)}
                    title={`Add match to ${stageTitle}`}
                  >
                    <Plus className="size-3" />
                    Match
                  </Button>
                ) : null}
              </div>

              {/* Symmetrically aligned Match Cards track */}
              <div
                className="relative w-full"
                style={{ height: `${maxTrackHeight}px` }}
              >
                {stageMatches.map((m) => {
                  const topPos = matchPositions[m.id] ?? 0;
                  const cardH = getCardHeight(m.teamCount);
                  const slot = nodeSlotStatus[m.id];
                  const entries = m.entries ?? [];
                  const totalSlots = Math.max(2, m.teamCount);

                  // Prioritize showing advancing teams, winners, or ranked teams first
                  const sortedEntries = [...entries].sort((a, b) => {
                    if (a.advances !== b.advances) return a.advances ? -1 : 1;
                    if (a.finalRank && b.finalRank) return a.finalRank - b.finalRank;
                    if (a.finalRank) return -1;
                    if (b.finalRank) return 1;
                    return 0;
                  });

                  const displaySlots: Array<
                    | { type: "slot"; entry: (typeof entries)[0] | undefined; slotIdx: number }
                    | { type: "more"; count: number }
                  > =
                    totalSlots <= MAX_DISPLAY_SLOTS
                      ? Array.from({ length: totalSlots }, (_, i) => ({
                          type: "slot" as const,
                          entry: sortedEntries[i],
                          slotIdx: i,
                        }))
                      : [
                          ...Array.from({ length: 3 }, (_, i) => ({
                            type: "slot" as const,
                            entry: sortedEntries[i],
                            slotIdx: i,
                          })),
                          { type: "more" as const, count: totalSlots - 3 },
                        ];

                  return (
                    <div
                      key={m.id}
                      data-match-id={m.id}
                      onClick={() => handleCardClick(m)}
                      style={{
                        position: "absolute",
                        top: `${topPos}px`,
                        left: 0,
                        right: 0,
                        height: `${cardH}px`,
                      }}
                      className={cn(
                        "group flex flex-col justify-between rounded-lg border bg-card/95 transition-colors duration-150 shadow-sm overflow-hidden cursor-pointer hover:border-primary/80 hover:shadow-md",
                        m.isCompleted ? "border-primary/50 shadow-xs" : "border-border/80"
                      )}
                    >
                      {/* Top bar: Venue / Status */}
                      <div className="h-[26px] flex items-center justify-between px-2.5 text-[11px] text-muted-foreground border-b border-border/40 shrink-0">
                        <span className="truncate font-medium flex items-center gap-1">
                          <MapPin className="size-2.5 shrink-0" />
                          {m.venue || "Venue"}
                        </span>
                        {m.isCompleted ? (
                          <span className="text-[10px] text-emerald-500 font-semibold flex items-center gap-0.5">
                            <CheckCircle2 className="size-2.5" /> Done
                          </span>
                        ) : (
                          <span className="text-[10px] text-muted-foreground flex items-center gap-0.5">
                            <Clock className="size-2.5" /> Pending
                          </span>
                        )}
                      </div>

                      {/* Team slots */}
                      <div className="flex flex-col divide-y divide-border/40 text-xs flex-1 justify-center">
                        {displaySlots.map((item) => {
                          if (item.type === "more") {
                            return (
                              <div
                                key="more-slots"
                                className="h-[26px] flex items-center justify-center px-2.5 text-[10px] text-muted-foreground bg-muted/20 font-medium group-hover:text-primary transition-colors cursor-pointer"
                              >
                                +{item.count} more teams (click to view)
                              </div>
                            );
                          }

                          const entry = item.entry;
                          const slotIdx = item.slotIdx;
                          const isWinner = Boolean(m.isCompleted && entry && entry.finalRank === 1);
                          const hasAdvanced = Boolean(m.isCompleted && entry && entry.advances);

                          if (!entry) {
                            return (
                              <div
                                key={`slot-${slotIdx}`}
                                className="h-[26px] flex items-center justify-between px-2.5 text-muted-foreground/50 text-xs"
                              >
                                <span className="truncate">TBD</span>
                                <span className="font-mono text-xs text-muted-foreground/40">-</span>
                              </div>
                            );
                          }

                          const displayScore = entry.score
                            ? entry.score
                            : m.isCompleted && entry.finalRank
                              ? `#${entry.finalRank}`
                              : "-";

                          return (
                            <div
                              key={entry.id || `entry-${slotIdx}`}
                              className={cn(
                                "h-[26px] flex items-center justify-between px-2.5 transition-colors",
                                isWinner
                                  ? "bg-primary/15 font-bold text-foreground"
                                  : hasAdvanced
                                    ? "bg-emerald-500/10 font-medium text-foreground"
                                    : "text-muted-foreground"
                              )}
                            >
                              <span className="truncate max-w-[140px] text-xs">
                                {entry.teamName || `Team ${slotIdx + 1}`}
                              </span>
                              <div className="flex items-center gap-1.5 shrink-0 pl-1">
                                {hasAdvanced ? (
                                  <span className="text-[8px] uppercase font-bold text-emerald-600 dark:text-emerald-400 bg-emerald-500/15 px-1 rounded">
                                    ADV
                                  </span>
                                ) : null}
                                <span
                                  className={cn(
                                    "font-mono text-xs px-1.5 py-0.5 rounded",
                                    isWinner
                                      ? "bg-primary text-primary-foreground font-bold"
                                      : "bg-muted text-muted-foreground font-medium"
                                  )}
                                >
                                  {displayScore}
                                </span>
                              </div>
                            </div>
                          );
                        })}
                      </div>

                      {/* Bottom bar: Name & Inflow status */}
                      <div className="h-[24px] flex items-center justify-between px-2.5 bg-muted/20 border-t border-border/40 text-[10px] text-muted-foreground shrink-0">
                        <span className="font-semibold truncate text-foreground/80">{m.name}</span>

                        {canEdit && stageLevel > 1 && slot ? (
                          <span
                            className={cn(
                              "text-[10px] font-bold",
                              slot.status === "satisfied"
                                ? "text-emerald-600 dark:text-emerald-400"
                                : slot.status === "underflow"
                                  ? "text-amber-600 dark:text-amber-400"
                                  : "text-destructive"
                            )}
                          >
                            {slot.currentIncoming}/{slot.requiredCount} Inflow
                          </span>
                        ) : null}
                      </div>
                    </div>
                  );
                })}
              </div>
            </div>
          );
        })}

        {/* Rightmost: + Add Next Stage Button */}
        {canEdit && onAddStage ? (
          <div className="flex flex-col gap-5 w-68 shrink-0">
            <Button
              variant="outline"
              size="sm"
              className="h-9 px-3 rounded-lg border border-dashed border-border hover:border-primary/60 hover:bg-muted/60 text-muted-foreground hover:text-foreground text-xs font-semibold tracking-wide uppercase shadow-2xs gap-1.5 w-full transition-colors"
              onClick={onAddStage}
            >
              <Plus className="size-3.5" />
              Add Stage {maxStageLevel + 1}
            </Button>
          </div>
        ) : null}
      </div>

      {/* Read-Only Match Details Dialog (when viewing in public/spectator mode) */}
      {viewingMatch ? (
        <Dialog open={Boolean(viewingMatch)} onOpenChange={(open) => !open && setViewingMatch(null)}>
          <DialogContent className="sm:max-w-md max-h-[85vh] overflow-y-auto">
            <DialogHeader>
              <div className="flex items-center gap-2 text-primary">
                <Trophy className="size-5" />
                <DialogTitle className="text-lg font-bold">{viewingMatch.name}</DialogTitle>
              </div>
              <DialogDescription>
                Stage {viewingMatch.stageLevel} • {viewingMatch.venue || "Main Arena"}
              </DialogDescription>
            </DialogHeader>

            <div className="flex flex-col gap-4 py-2">
              <div className="flex items-center justify-between text-xs text-muted-foreground border-b border-border pb-2">
                <span>
                  Status:{" "}
                  <strong className="text-foreground font-semibold">
                    {viewingMatch.isCompleted ? "Completed" : "Pending / In Progress"}
                  </strong>
                </span>
                <span>
                  Total Slots:{" "}
                  <strong className="text-foreground font-semibold">{viewingMatch.teamCount}</strong>
                </span>
              </div>

              <div>
                <h4 className="text-xs font-bold uppercase tracking-wider text-muted-foreground mb-2">
                  Participating Teams & Slots ({viewingMatch.teamCount} Total)
                </h4>
                <div className="flex flex-col gap-2 max-h-64 overflow-y-auto pr-1">
                  {Array.from({ length: Math.max(2, viewingMatch.teamCount) }).map((_, slotIdx) => {
                    const entry = viewingMatch.entries?.[slotIdx];
                    if (!entry) {
                      return (
                        <div
                          key={`view-slot-${slotIdx}`}
                          className="flex items-center justify-between px-3 py-2 rounded-lg border border-border/60 bg-muted/15 text-xs text-muted-foreground/70"
                        >
                          <span className="font-mono text-[11px]">Slot {slotIdx + 1}</span>
                          <span className="italic">TBD (Waiting for qualifier / seed)</span>
                        </div>
                      );
                    }
                    const isWinner = Boolean(viewingMatch.isCompleted && entry.finalRank === 1);
                    const hasAdvanced = Boolean(viewingMatch.isCompleted && entry.advances);
                    const displayScore = entry.score
                      ? entry.score
                      : viewingMatch.isCompleted && entry.finalRank
                        ? `#${entry.finalRank}`
                        : "-";

                    return (
                      <div
                        key={entry.id || `view-entry-${slotIdx}`}
                        className={cn(
                          "flex items-center justify-between px-3 py-2 rounded-lg border text-xs transition-colors",
                          isWinner
                            ? "border-primary/40 bg-primary/10 font-bold"
                            : hasAdvanced
                              ? "border-emerald-500/30 bg-emerald-500/10"
                              : "border-border/60 bg-muted/20"
                        )}
                      >
                        <div className="flex items-center gap-2 truncate">
                          <span className="font-mono text-muted-foreground text-[11px]">
                            {viewingMatch.isCompleted && entry.finalRank ? `#${entry.finalRank}` : `Slot ${slotIdx + 1}`}
                          </span>
                          <span className="truncate font-semibold text-foreground">
                            {entry.teamName || `Team ${slotIdx + 1}`}
                          </span>
                        </div>
                        <div className="flex items-center gap-2 shrink-0">
                          {hasAdvanced ? (
                            <span className="text-[9px] uppercase font-bold text-emerald-600 dark:text-emerald-400 bg-emerald-500/15 px-1.5 py-0.5 rounded">
                              Advances
                            </span>
                          ) : null}
                          <span className="font-mono font-bold text-xs px-2 py-0.5 rounded bg-muted">
                            {displayScore}
                          </span>
                        </div>
                      </div>
                    );
                  })}
                </div>
              </div>

              {viewingMatch.targetMatchId ? (
                <div className="rounded-lg border border-border/60 bg-muted/10 p-3 text-xs text-muted-foreground">
                  Winner advances to the next stage match.
                </div>
              ) : (
                <div className="rounded-lg border border-primary/30 bg-primary/5 p-3 text-xs text-primary font-medium">
                  Championship / Final match.
                </div>
              )}
            </div>

            <DialogFooter>
              <Button size="sm" variant="outline" onClick={() => setViewingMatch(null)}>
                Close
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      ) : null}
    </div>
  );
}
