import * as React from "react";
import ReactMarkdown from "react-markdown";
import { cn } from "@/lib/utils";
import {
  defaultMarkdownComponents,
  defaultMarkdownPlugins,
  defaultRehypePlugins,
} from "@/lib/markdown";

interface MarkdownContentProps {
  content?: string | null;
  className?: string;
  fallback?: React.ReactNode;
}

export function MarkdownContent({
  content,
  className,
  fallback = null,
}: MarkdownContentProps) {
  const trimmed = content?.trim();

  if (!trimmed) {
    return fallback ? <>{fallback}</> : null;
  }

  return (
    <div
      data-slot="markdown-content"
      className={cn("text-foreground break-words text-sm", className)}
    >
      <ReactMarkdown
        remarkPlugins={defaultMarkdownPlugins}
        rehypePlugins={defaultRehypePlugins}
        components={defaultMarkdownComponents}
      >
        {trimmed}
      </ReactMarkdown>
    </div>
  );
}
