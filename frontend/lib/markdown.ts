import * as React from "react";
import type { Components } from "react-markdown";
import type { PluggableList } from "unified";
import remarkGfm from "remark-gfm";
import rehypeSanitize, { defaultSchema } from "rehype-sanitize";
import { cn } from "cn";

/**
 * Custom rehype-sanitize schema ensuring safe links, tables, and no raw scripts/iframes.
 */
export const safeSanitizeSchema = {
  ...defaultSchema,
  protocols: {
    ...defaultSchema.protocols,
    href: ["http", "https", "mailto"],
  },
};

export const defaultMarkdownPlugins: PluggableList = [remarkGfm];
export const defaultRehypePlugins: PluggableList = [[rehypeSanitize, safeSanitizeSchema]];

/**
 * Strips common markdown formatting syntax for clean plain-text excerpts.
 */
export function stripMarkdown(markdown?: string | null): string {
  if (!markdown) return "";
  return markdown
    .replace(/^#{1,6}\s+/gm, "") // headers
    .replace(/\*\*([^*]+)\*\*/g, "$1") // bold
    .replace(/\*([^*]+)\*/g, "$1") // italic
    .replace(/~~([^~]+)~~/g, "$1") // strikethrough
    .replace(/`{1,3}[^`]*`{1,3}/g, "") // inline & block code
    .replace(/\[([^\]]+)\]\([^)]+\)/g, "$1") // links
    .replace(/^>\s+/gm, "") // blockquotes
    .replace(/^[-*+]\s+/gm, "") // unordered lists
    .replace(/^\d+\.\s+/gm, "") // ordered lists
    .replace(/\|/g, " ") // table borders
    .replace(/\n{2,}/g, " ") // excessive newlines
    .replace(/\s+/g, " ") // whitespace collapse
    .trim();
}

/**
 * Safe component overrides for ReactMarkdown matching Eventory's styling tokens.
 */
export const defaultMarkdownComponents: Components = {
  h1: ({ children }) =>
    React.createElement(
      "h2",
      { className: "mt-4 mb-2 text-lg font-bold tracking-tight text-foreground first:mt-0" },
      children
    ),
  h2: ({ children }) =>
    React.createElement(
      "h3",
      { className: "mt-4 mb-2 text-base font-semibold tracking-tight text-foreground first:mt-0" },
      children
    ),
  h3: ({ children }) =>
    React.createElement(
      "h4",
      { className: "mt-3 mb-1.5 text-sm font-semibold text-foreground first:mt-0" },
      children
    ),
  h4: ({ children }) =>
    React.createElement(
      "h5",
      { className: "mt-2 mb-1 text-sm font-medium text-foreground first:mt-0" },
      children
    ),
  p: ({ children }) =>
    React.createElement(
      "p",
      { className: "my-2 leading-relaxed text-foreground/90 first:mt-0 last:mb-0" },
      children
    ),
  a: ({ href, children }) => {
    const isExternal = href?.startsWith("http://") || href?.startsWith("https://");
    return React.createElement(
      "a",
      {
        href,
        target: isExternal ? "_blank" : undefined,
        rel: isExternal ? "noopener noreferrer" : undefined,
        className: "text-primary underline underline-offset-4 transition-opacity hover:opacity-80",
      },
      children
    );
  },
  ul: ({ children }) =>
    React.createElement(
      "ul",
      { className: "my-2 ml-5 list-disc space-y-1 text-foreground/90" },
      children
    ),
  ol: ({ children }) =>
    React.createElement(
      "ol",
      { className: "my-2 ml-5 list-decimal space-y-1 text-foreground/90" },
      children
    ),
  li: ({ children }) =>
    React.createElement("li", { className: "leading-relaxed" }, children),
  blockquote: ({ children }) =>
    React.createElement(
      "blockquote",
      { className: "my-2 border-l-2 border-border pl-4 italic text-muted-foreground" },
      children
    ),
  code: ({ className: codeClassName, children, ...props }) => {
    const isInline = !codeClassName && typeof children === "string" && !children.includes("\n");
    if (isInline) {
      return React.createElement(
        "code",
        {
          className: "rounded bg-muted px-1.5 py-0.5 font-mono text-xs text-foreground",
          ...props,
        },
        children
      );
    }
    return React.createElement(
      "code",
      { className: cn("font-mono text-xs", codeClassName), ...props },
      children
    );
  },
  pre: ({ children }) =>
    React.createElement(
      "pre",
      { className: "my-3 overflow-x-auto rounded-lg bg-muted p-3 font-mono text-xs text-foreground" },
      children
    ),
  table: ({ children }) =>
    React.createElement(
      "div",
      { className: "my-3 overflow-x-auto" },
      React.createElement(
        "table",
        { className: "w-full border-collapse border border-border text-sm" },
        children
      )
    ),
  thead: ({ children }) =>
    React.createElement("thead", { className: "bg-muted/60" }, children),
  th: ({ children }) =>
    React.createElement(
      "th",
      { className: "border border-border px-3 py-1.5 text-left font-medium text-foreground" },
      children
    ),
  td: ({ children }) =>
    React.createElement(
      "td",
      { className: "border border-border px-3 py-1.5 text-foreground/90" },
      children
    ),
  hr: () => React.createElement("hr", { className: "my-4 border-border" }),
};
