"use client";

import { useState } from "react";
import { AlertCircle, RotateCcw, Trash2, Users } from "lucide-react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Spinner } from "@/components/ui/spinner";

interface ResetBracketDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onResetStage1Teams: () => Promise<void>;
  onResetAll: () => Promise<void>;
}

export function ResetBracketDialog({
  open,
  onOpenChange,
  onResetStage1Teams,
  onResetAll,
}: ResetBracketDialogProps) {
  const [loadingAction, setLoadingAction] = useState<"teams" | "all" | null>(null);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);

  async function handleResetTeams() {
    setErrorMsg(null);
    setLoadingAction("teams");
    try {
      await onResetStage1Teams();
      onOpenChange(false);
    } catch (err: unknown) {
      setErrorMsg(err instanceof Error ? err.message : "Failed to reset Stage 1 teams.");
    } finally {
      setLoadingAction(null);
    }
  }

  async function handleResetAll() {
    setErrorMsg(null);
    setLoadingAction("all");
    try {
      await onResetAll();
      onOpenChange(false);
    } catch (err: unknown) {
      setErrorMsg(err instanceof Error ? err.message : "Failed to reset bracket layout.");
    } finally {
      setLoadingAction(null);
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <div className="flex items-center gap-2 text-primary">
            <RotateCcw className="size-5" />
            <DialogTitle className="text-lg font-bold">Reset Bracket</DialogTitle>
          </div>
          <DialogDescription>
            Choose how you would like to reset the tournament bracket.
          </DialogDescription>
        </DialogHeader>

        {errorMsg ? (
          <div className="rounded-md bg-destructive/15 p-3 text-sm text-destructive font-medium flex items-center gap-2">
            <AlertCircle className="size-4 shrink-0" />
            <span>{errorMsg}</span>
          </div>
        ) : null}

        <div className="flex flex-col gap-4 py-2">
          {/* Option 1: Reset Stage 1 Teams */}
          <div className="flex flex-col gap-2 rounded-xl border border-border p-4 bg-muted/20 hover:border-primary/40 transition-colors">
            <div className="flex items-center gap-2.5">
              <div className="size-8 rounded-lg bg-primary/10 text-primary flex items-center justify-center shrink-0">
                <Users className="size-4" />
              </div>
              <div>
                <h4 className="text-sm font-bold text-foreground">Reset Stage 1 Teams</h4>
                <p className="text-xs text-muted-foreground mt-0.5">
                  Clear all seeded teams and scores from Stage 1 matches. The bracket structure and match connections will remain intact.
                </p>
              </div>
            </div>
            <div className="flex justify-end pt-2">
              <Button
                variant="outline"
                size="sm"
                className="text-xs h-8"
                onClick={handleResetTeams}
                disabled={loadingAction !== null}
              >
                {loadingAction === "teams" ? <Spinner /> : <RotateCcw className="size-3.5 mr-1" />}
                Reset Teams Only
              </Button>
            </div>
          </div>

          {/* Option 2: Reset Everything */}
          <div className="flex flex-col gap-2 rounded-xl border border-destructive/20 p-4 bg-destructive/5 hover:border-destructive/40 transition-colors">
            <div className="flex items-center gap-2.5">
              <div className="size-8 rounded-lg bg-destructive/10 text-destructive flex items-center justify-center shrink-0">
                <Trash2 className="size-4" />
              </div>
              <div>
                <h4 className="text-sm font-bold text-destructive">Reset Entire Bracket</h4>
                <p className="text-xs text-muted-foreground mt-0.5">
                  Delete all matches, rounds, connections, and scores. This will return the tournament to a blank canvas.
                </p>
              </div>
            </div>
            <div className="flex justify-end pt-2">
              <Button
                variant="destructive"
                size="sm"
                className="text-xs h-8"
                onClick={handleResetAll}
                disabled={loadingAction !== null}
              >
                {loadingAction === "all" ? <Spinner /> : <Trash2 className="size-3.5 mr-1" />}
                Reset Everything
              </Button>
            </div>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}
