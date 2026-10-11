"use client";

import React, { useState } from "react";
import { Check, Pencil } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";

export interface FormHeaderCardProps {
  title: string;
  description: string;
  onTitleChange: (val: string) => void;
  onDescriptionChange: (val: string) => void;
}

export function FormHeaderCard({
  title,
  description,
  onTitleChange,
  onDescriptionChange,
}: FormHeaderCardProps) {
  const [isEditingTitle, setIsEditingTitle] = useState(false);

  return (
    <Card className="border-border border-t-4 border-t-primary shadow-xs">
      <CardContent className="flex flex-col gap-4 pt-6">
        {/* Form Title */}
        <div className="flex items-center justify-between gap-3 min-w-0">
          {isEditingTitle ? (
            <div className="flex items-center gap-2 flex-1 min-w-0">
              <Input
                autoFocus
                value={title}
                onChange={(e) => onTitleChange(e.target.value)}
                onBlur={() => setIsEditingTitle(false)}
                onKeyDown={(e) => {
                  if (e.key === "Enter") {
                    setIsEditingTitle(false);
                  }
                }}
                placeholder="Enter form title..."
                className="h-auto py-0 px-0 rounded-none border-x-0 border-t-0 border-b border-muted-foreground/30 font-heading text-2xl md:text-2xl font-bold leading-normal tracking-tight text-foreground shadow-none focus-visible:border-b-2 focus-visible:border-primary focus-visible:ring-0 bg-transparent"
              />
              <Button
                type="button"
                variant="ghost"
                size="sm"
                onClick={() => setIsEditingTitle(false)}
                aria-label="Confirm title"
                className="size-8 p-0 shrink-0 text-muted-foreground hover:text-foreground"
              >
                <Check className="size-4" />
              </Button>
            </div>
          ) : (
            <div className="flex items-center gap-2 flex-1 min-w-0">
              <h1
                onClick={() => setIsEditingTitle(true)}
                className="font-heading text-2xl md:text-2xl font-bold leading-normal tracking-tight cursor-pointer truncate hover:text-primary transition-colors text-foreground"
              >
                {title.trim() ? (
                  title
                ) : (
                  <span className="text-muted-foreground italic">Untitled form</span>
                )}
              </h1>
              <Button
                type="button"
                variant="ghost"
                size="sm"
                onClick={() => setIsEditingTitle(true)}
                aria-label="Edit form title"
                className="size-8 p-0 shrink-0 text-muted-foreground hover:text-foreground"
              >
                <Pencil className="size-4" />
              </Button>
            </div>
          )}
        </div>

        {/* Form Description */}
        <Textarea
          value={description}
          onChange={(e) => onDescriptionChange(e.target.value)}
          placeholder="Form description..."
          rows={3}
        />
      </CardContent>
    </Card>
  );
}
