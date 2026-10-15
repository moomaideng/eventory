"use client";

import React from "react";
import {
  FileText,
  HelpCircle,
  ListOrdered,
  Sparkles,
} from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import type { FormField } from "@/features/registration-form/schemas";

const TYPE_LABELS: Record<string, string> = {
  SHORT_TEXT: "Short text",
  LONG_TEXT: "Long text",
  NUMBER: "Number",
  SELECT_ONE: "Single choice",
  SELECT_MULTIPLE: "Multiple choice",
};

export interface FormBuilderSidebarProps {
  fields: FormField[];
  onScrollToField?: (id: string) => void;
}

export function FormBuilderSidebar({
  fields,
  onScrollToField,
}: FormBuilderSidebarProps) {
  const requiredCount = fields.filter((f) => f.required).length;
  const optionalCount = fields.length - requiredCount;

  const typeCounts = fields.reduce<Record<string, number>>((acc, f) => {
    acc[f.type] = (acc[f.type] || 0) + 1;
    return acc;
  }, {});

  return (
    <aside className="flex flex-col gap-6 w-full lg:sticky lg:top-6">
      {/* Overview Card */}
      <Card className="border-border/60 bg-card/60">
        <CardHeader className="pb-3">
          <div className="flex items-center justify-between gap-2">
            <CardTitle className="text-base font-semibold flex items-center gap-2">
              <FileText className="size-4 text-primary" aria-hidden="true" />
              Form summary
            </CardTitle>
            <Badge variant="outline" className="tabular-nums">
              {fields.length} {fields.length === 1 ? "field" : "fields"}
            </Badge>
          </div>
        </CardHeader>
        <CardContent className="flex flex-col gap-4 text-sm">
          <div className="grid grid-cols-2 gap-2 text-xs">
            <div className="flex flex-col gap-0.5 rounded-lg border border-border/40 bg-muted/20 p-2.5">
              <span className="text-muted-foreground">Required</span>
              <span className="text-base font-bold tabular-nums text-foreground">
                {requiredCount}
              </span>
            </div>
            <div className="flex flex-col gap-0.5 rounded-lg border border-border/40 bg-muted/20 p-2.5">
              <span className="text-muted-foreground">Optional</span>
              <span className="text-base font-bold tabular-nums text-foreground">
                {optionalCount}
              </span>
            </div>
          </div>

          {Object.keys(typeCounts).length > 0 && (
            <div className="flex flex-col gap-2 border-t border-border/40 pt-3">
              <span className="text-xs font-medium text-muted-foreground">
                Field types breakdown
              </span>
              <div className="flex flex-wrap gap-1.5">
                {Object.entries(typeCounts).map(([type, count]) => (
                  <Badge
                    key={type}
                    variant="secondary"
                    className="text-xs font-normal"
                  >
                    {TYPE_LABELS[type] ?? type}: {count}
                  </Badge>
                ))}
              </div>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Question Outline Card */}
      <Card className="border-border/60 bg-card/60">
        <CardHeader className="pb-3">
          <CardTitle className="text-base font-semibold flex items-center gap-2">
            <ListOrdered className="size-4 text-primary" aria-hidden="true" />
            Questions outline
          </CardTitle>
        </CardHeader>
        <CardContent className="text-sm">
          {fields.length === 0 ? (
            <p className="text-xs text-muted-foreground italic">
              No questions added yet. Click &quot;Add Question&quot; to begin.
            </p>
          ) : (
            <ol className="flex flex-col gap-2 max-h-72 overflow-y-auto pr-1">
              {fields.map((f, i) => (
                <li key={f.id}>
                  <button
                    type="button"
                    onClick={() => onScrollToField?.(f.id)}
                    className="flex w-full items-center justify-between gap-2 rounded-md p-2 text-left text-xs transition-colors hover:bg-muted/50 focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
                  >
                    <span className="truncate flex-1 font-medium text-foreground">
                      <span className="text-muted-foreground mr-1">#{i + 1}</span>
                      {f.label.trim() ? f.label : <span className="italic text-muted-foreground">Untitled question</span>}
                    </span>
                    <Badge variant="outline" className="text-[10px] shrink-0 font-normal">
                      {f.required ? "Required" : "Optional"}
                    </Badge>
                  </button>
                </li>
              ))}
            </ol>
          )}
        </CardContent>
      </Card>

      {/* Participant Guidelines Card */}
      <Card className="border-border/40 bg-muted/10">
        <CardHeader className="pb-2">
          <CardTitle className="text-xs font-semibold text-muted-foreground uppercase tracking-wider flex items-center gap-1.5">
            <HelpCircle className="size-3.5 text-muted-foreground" aria-hidden="true" />
            Organizer notes
          </CardTitle>
        </CardHeader>
        <CardContent className="text-xs text-muted-foreground flex flex-col gap-2 leading-relaxed">
          <p>
            Questions configured here will be presented to participants during tournament registration.
          </p>
          <p className="flex items-center gap-1 text-primary">
            <Sparkles className="size-3.5 shrink-0" aria-hidden="true" />
            Switching tabs preserves your edits automatically.
          </p>
        </CardContent>
      </Card>
    </aside>
  );
}
