import type { BracketMatchNode } from "./schemas";
import type { components } from "@/lib/api/schema";

export type MatchResponse = components["schemas"]["MatchResponse"];
export type MatchEntryResponse = components["schemas"]["MatchEntryResponse"];

export interface InflowSlotStatus {
  currentIncoming: number;
  requiredCount: number;
  isSatisfied: boolean;
  status: "satisfied" | "underflow" | "overflow";
}

export interface InvariantValidationResult {
  isValid: boolean;
  errors: string[];
  nodeSlotStatus: Record<string, InflowSlotStatus>;
  stage1SlotsSum: number;
}

/**
 * Validates bracket match graph against business invariants:
 * 1. Must contain at least one match.
 * 2. Must contain at least one Stage 1 match for initial entry.
 * 3. Must contain at least one Final / Championship match (without target destination).
 * 4. Target destinations must exist, point to a higher stage level, and not target self.
 * 5. Advancing count cannot exceed team count. If a match has a target, it must advance >= 1 team.
 * 6. Inflow Invariant: For every match at Stage > 1, sum(incoming advancingCount) == target.teamCount.
 * 7. Graph must be an acyclic converging DAG (no directed cycles).
 * 8. Stage 1 slots sum must equal accepted entries count (when acceptedCount > 0).
 */
export function validateBracketInvariants(
  matches: BracketMatchNode[],
  acceptedCount?: number
): InvariantValidationResult {
  const errors: string[] = [];
  const nodeSlotStatus: Record<string, InflowSlotStatus> = {};

  if (!matches || matches.length === 0) {
    return {
      isValid: false,
      errors: ["Bracket must contain at least one match."],
      nodeSlotStatus,
      stage1SlotsSum: 0,
    };
  }

  const map = new Map<string, BracketMatchNode>();
  let hasStage1 = false;
  let hasChampionship = false;
  let stage1SlotsSum = 0;

  for (const m of matches) {
    const id = m.tempId?.trim();
    if (!id) {
      errors.push("Each match must have a valid identifier.");
      continue;
    }
    if (map.has(id)) {
      errors.push(`Duplicate match ID "${id}".`);
    }
    map.set(id, m);

    if (m.stageLevel === 1) {
      hasStage1 = true;
      stage1SlotsSum += m.teamCount;
    }
    if (!m.targetTempId || m.targetTempId.trim() === "") {
      hasChampionship = true;
    }

    if (m.stageLevel > 1) {
      nodeSlotStatus[id] = {
        currentIncoming: 0,
        requiredCount: m.teamCount,
        isSatisfied: false,
        status: "underflow",
      };
    }

    if (m.teamCount < 1) {
      errors.push(`Match "${m.name || id}" must have a team count of at least 1.`);
    }

    if (m.advancingCount > m.teamCount) {
      errors.push(
        `Match "${m.name || id}" advancing count (${m.advancingCount}) cannot exceed team count (${m.teamCount}).`
      );
    }
  }

  if (!hasStage1) {
    errors.push("Bracket must contain at least one Stage 1 match for starting teams.");
  }
  if (!hasChampionship) {
    errors.push("Bracket must contain at least one Final / Championship match without a target destination.");
  }

  if (acceptedCount !== undefined && acceptedCount > 0) {
    if (stage1SlotsSum !== acceptedCount) {
      errors.push(
        `Total Stage 1 team slots (${stage1SlotsSum}) must equal the number of accepted entries (${acceptedCount}).`
      );
    }
  }

  const inDegree: Record<string, number> = {};
  const adj: Record<string, string[]> = {};
  for (const [id] of map) {
    inDegree[id] = 0;
    adj[id] = [];
  }

  for (const m of matches) {
    const id = m.tempId.trim();
    const targetId = m.targetTempId?.trim();
    if (!targetId) continue;

    if (targetId === id) {
      errors.push(`Match "${m.name || id}" cannot target itself.`);
      continue;
    }

    const target = map.get(targetId);
    if (!target) {
      errors.push(`Match "${m.name || id}" targets an unknown match "${targetId}".`);
      continue;
    }

    if (target.stageLevel <= m.stageLevel) {
      errors.push(
        `Match "${m.name || id}" (Stage ${m.stageLevel}) must target a higher stage level (Target "${target.name || targetId}" is Stage ${target.stageLevel}).`
      );
      continue;
    }

    if (m.advancingCount < 1) {
      errors.push(`Match "${m.name || id}" targets "${target.name || targetId}" but advances 0 teams.`);
    }

    if (nodeSlotStatus[targetId]) {
      nodeSlotStatus[targetId].currentIncoming += m.advancingCount;
    }

    adj[id].push(targetId);
    inDegree[targetId] = (inDegree[targetId] || 0) + 1;
  }

  // Check Inflow Invariant for all Stage > 1 matches
  for (const [id, slot] of Object.entries(nodeSlotStatus)) {
    const targetNode = map.get(id);
    const name = targetNode?.name || id;

    if (slot.currentIncoming === slot.requiredCount) {
      slot.isSatisfied = true;
      slot.status = "satisfied";
    } else if (slot.currentIncoming < slot.requiredCount) {
      slot.isSatisfied = false;
      slot.status = "underflow";
      const missing = slot.requiredCount - slot.currentIncoming;
      errors.push(
        `Match "${name}" expects ${slot.requiredCount} teams, but incoming matches only advance ${slot.currentIncoming} (${missing} missing).`
      );
    } else {
      slot.isSatisfied = false;
      slot.status = "overflow";
      const extra = slot.currentIncoming - slot.requiredCount;
      errors.push(
        `Match "${name}" capacity is ${slot.requiredCount} teams, but incoming matches advance ${slot.currentIncoming} (${extra} too many).`
      );
    }
  }

  // Cycle detection via Kahn's algorithm
  const queue: string[] = [];
  for (const [id, deg] of Object.entries(inDegree)) {
    if (deg === 0) queue.push(id);
  }

  let visitedCount = 0;
  while (queue.length > 0) {
    const current = queue.shift()!;
    visitedCount++;
    for (const neighbor of adj[current] || []) {
      inDegree[neighbor]--;
      if (inDegree[neighbor] === 0) {
        queue.push(neighbor);
      }
    }
  }

  if (visitedCount < map.size) {
    errors.push("Graph contains a cycle or circular loop. All matches must progress forward.");
  }

  return {
    isValid: errors.length === 0,
    errors,
    nodeSlotStatus,
    stage1SlotsSum,
  };
}

/**
 * Converts a MatchResponse from backend into a BracketMatchNode for editing.
 */
export function matchResponseToNode(m: MatchResponse): BracketMatchNode {
  return {
    tempId: m.id,
    name: m.name,
    venue: m.venue,
    stageLevel: m.stageLevel,
    teamCount: m.teamCount,
    advancingCount: m.advancingCount,
    targetTempId: m.targetMatchId ?? null,
  };
}
