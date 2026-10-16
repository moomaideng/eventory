"use client";

import React from "react";
import { ListOrdered } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import type { FormField } from "@/features/registration-form/schemas";

export interface FormBuilderSidebarProps {
  fields: FormField[];
  onScrollToField?: (id: string) => void;
}

export function FormBuilderSidebar({
  fields,
  onScrollToField,
}: FormBuilderSidebarProps) {
  return (
    <aside className="flex flex-col gap-6 w-full lg:sticky lg:top-6">
      {/* Question Outline Card */}
      <Card className="border-border/60 bg-card/60">
        <CardHeader className="pb-3">
          <div className="flex items-center justify-between gap-2">
            <CardTitle className="text-base font-semibold flex items-center gap-2">
              <ListOrdered className="size-4 text-primary" aria-hidden="true" />
              Questions outline
            </CardTitle>
            <Badge variant="outline" className="tabular-nums">
              {fields.length} {fields.length === 1 ? "question" : "questions"}
            </Badge>
          </div>
        </CardHeader>
        <CardContent className="text-sm">
          {fields.length === 0 ? (
            <p className="text-xs text-muted-foreground italic">
              No questions added yet. Click &quot;Add Question&quot; to begin.
            </p>
          ) : (
            <ol className="flex flex-col gap-2 max-h-[32rem] overflow-y-auto pr-1">
              {fields.map((f, i) => (
                <li key={f.id}>
                  <button
                    type="button"
                    onClick={() => onScrollToField?.(f.id)}
                    className="flex w-full items-center justify-between gap-2 rounded-md p-2 text-left text-xs transition-colors hover:bg-muted/50 focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
                  >
                    <span className="truncate flex-1 font-medium text-foreground">
                      <span className="text-muted-foreground mr-1">#{i + 1}</span>
                      {f.label.trim() ? (
                        f.label
                      ) : (
                        <span className="italic text-muted-foreground">
                          Untitled question
                        </span>
                      )}
                    </span>
                    <Badge
                      variant="outline"
                      className="text-[10px] shrink-0 font-normal"
                    >
                      {f.required ? "Required" : "Optional"}
                    </Badge>
                  </button>
                </li>
              ))}
            </ol>
          )}
        </CardContent>
      </Card>
    </aside>
  );
}
