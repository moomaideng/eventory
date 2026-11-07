"use client";

import * as React from "react";
import {
  Bold,
  Italic,
  Heading3,
  Quote,
  Code,
  List,
  ListOrdered,
  Link as LinkIcon,
  Eye,
  PenLine,
} from "lucide-react";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { MarkdownContent } from "@/components/ui/markdown-content";

export interface MarkdownEditorProps {
  id?: string;
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
  maxLength?: number;
  rows?: number;
  disabled?: boolean;
  className?: string;
  "aria-invalid"?: boolean;
}

export function MarkdownEditor({
  id,
  value,
  onChange,
  placeholder = "Write tournament details and rules in Markdown...",
  maxLength = 10000,
  rows = 10,
  disabled = false,
  className,
  "aria-invalid": ariaInvalid,
}: MarkdownEditorProps) {
  const textareaRef = React.useRef<HTMLTextAreaElement>(null);
  const [activeTab, setActiveTab] = React.useState<string>("write");

  const applyFormat = (prefix: string, suffix = "", defaultText = "") => {
    const textarea = textareaRef.current;
    if (!textarea) return;

    const start = textarea.selectionStart;
    const end = textarea.selectionEnd;
    const selected = value.slice(start, end);
    const textToInsert = selected || defaultText;
    const replacement = `${prefix}${textToInsert}${suffix}`;

    const newValue = value.slice(0, start) + replacement + value.slice(end);
    if (maxLength && newValue.length > maxLength) return;

    onChange(newValue);

    // Reposition cursor and refocus
    setTimeout(() => {
      textarea.focus();
      const cursorStart = start + prefix.length;
      const cursorEnd = cursorStart + textToInsert.length;
      textarea.setSelectionRange(cursorStart, cursorEnd);
    }, 0);
  };

  const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    // Support Ctrl+B / Cmd+B for Bold
    if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === "b") {
      e.preventDefault();
      applyFormat("**", "**", "bold text");
    }
    // Support Ctrl+I / Cmd+I for Italic
    if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === "i") {
      e.preventDefault();
      applyFormat("*", "*", "italic text");
    }
    // Support Ctrl+K / Cmd+K for Link
    if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === "k") {
      e.preventDefault();
      applyFormat("[", "](https://example.com)", "link text");
    }
  };

  return (
    <div
      data-slot="markdown-editor"
      className={cn(
        "flex flex-col rounded-lg border border-input bg-card shadow-xs transition-colors focus-within:border-ring focus-within:ring-ring/50 focus-within:ring-3",
        ariaInvalid && "border-destructive ring-destructive/20 ring-3",
        className
      )}
    >
      <Tabs
        value={activeTab}
        onValueChange={setActiveTab}
        className="flex flex-col gap-0"
      >
        {/* Header bar: Tabs on left, Quick Action Toolbar on right */}
        <div className="flex flex-wrap items-center justify-between gap-2 border-b border-border bg-muted/30 px-3 py-1.5">
          <TabsList className="h-8 bg-muted/60 p-0.5">
            <TabsTrigger
              value="write"
              className="h-7 gap-1.5 px-3 text-xs data-active:bg-background data-active:shadow-xs"
            >
              <PenLine className="size-3.5" />
              Write
            </TabsTrigger>
            <TabsTrigger
              value="preview"
              className="h-7 gap-1.5 px-3 text-xs data-active:bg-background data-active:shadow-xs"
            >
              <Eye className="size-3.5" />
              Preview
            </TabsTrigger>
          </TabsList>

          {activeTab === "write" ? (
            <div className="flex items-center gap-0.5 text-muted-foreground">
              <Button
                type="button"
                variant="ghost"
                size="icon-xs"
                title="Bold (Ctrl+B)"
                disabled={disabled}
                onClick={() => applyFormat("**", "**", "bold text")}
              >
                <Bold className="size-3.5" />
              </Button>
              <Button
                type="button"
                variant="ghost"
                size="icon-xs"
                title="Italic (Ctrl+I)"
                disabled={disabled}
                onClick={() => applyFormat("*", "*", "italic text")}
              >
                <Italic className="size-3.5" />
              </Button>
              <Button
                type="button"
                variant="ghost"
                size="icon-xs"
                title="Heading (H3)"
                disabled={disabled}
                onClick={() => applyFormat("\n### ", "\n", "Heading")}
              >
                <Heading3 className="size-3.5" />
              </Button>
              <div className="mx-1 h-3.5 w-px bg-border" />
              <Button
                type="button"
                variant="ghost"
                size="icon-xs"
                title="Bullet List"
                disabled={disabled}
                onClick={() => applyFormat("\n- ", "", "List item")}
              >
                <List className="size-3.5" />
              </Button>
              <Button
                type="button"
                variant="ghost"
                size="icon-xs"
                title="Numbered List"
                disabled={disabled}
                onClick={() => applyFormat("\n1. ", "", "First item")}
              >
                <ListOrdered className="size-3.5" />
              </Button>
              <Button
                type="button"
                variant="ghost"
                size="icon-xs"
                title="Blockquote"
                disabled={disabled}
                onClick={() => applyFormat("\n> ", "\n", "Quote")}
              >
                <Quote className="size-3.5" />
              </Button>
              <Button
                type="button"
                variant="ghost"
                size="icon-xs"
                title="Code"
                disabled={disabled}
                onClick={() => applyFormat("`", "`", "code")}
              >
                <Code className="size-3.5" />
              </Button>
              <Button
                type="button"
                variant="ghost"
                size="icon-xs"
                title="Insert Link (Ctrl+K)"
                disabled={disabled}
                onClick={() => applyFormat("[", "](https://example.com)", "link text")}
              >
                <LinkIcon className="size-3.5" />
              </Button>
            </div>
          ) : null}
        </div>

        {/* Tab 1: Write (Textarea) */}
        <TabsContent value="write" className="flex flex-col gap-0 p-0 outline-none">
          <Textarea
            ref={textareaRef}
            id={id}
            value={value}
            onChange={(e) => onChange(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder={placeholder}
            rows={rows}
            maxLength={maxLength}
            disabled={disabled}
            className="border-0 shadow-none focus-visible:ring-0 rounded-none bg-transparent px-3 py-2.5 text-sm resize-y"
          />
        </TabsContent>

        {/* Tab 2: Preview (Rendered Markdown) */}
        <TabsContent value="preview" className="flex flex-col gap-0 p-0 outline-none">
          <div
            className="min-h-[220px] max-h-[500px] overflow-y-auto px-4 py-3 bg-background/50"
            style={{ minHeight: `${rows * 24}px` }}
          >
            <MarkdownContent
              content={value}
              fallback={
                <p className="py-8 text-center text-xs italic text-muted-foreground">
                  Nothing to preview. Switch to Write to author details & rules.
                </p>
              }
            />
          </div>
        </TabsContent>
      </Tabs>

      {/* Footer bar: Markdown hint & Character counter */}
      <div className="flex items-center justify-between border-t border-border/60 bg-muted/20 px-3 py-1.5 text-[11px] text-muted-foreground">
        <span>Markdown supported: headings, lists, bold, italics, tables, links.</span>
        <span>
          {value.length.toLocaleString()} / {maxLength.toLocaleString()}
        </span>
      </div>
    </div>
  );
}
