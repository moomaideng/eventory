"use client";

import { useState } from "react";
import { X } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import {
  Field,
  FieldGroup,
  FieldLabel,
  FieldDescription,
} from "@/components/ui/field";

export function SponsorDirectoryFilters({
  search,
  onSearch,
  selected,
  onAdd,
  onRemove,
  onClear,
  tags,
  loading,
  error,
  onRetry,
  sponsors,
}: {
  search: string;
  onSearch: (value: string) => void;
  selected: string[];
  onAdd: (tag: string) => void;
  onRemove: (tag: string) => void;
  onClear: () => void;
  tags?: string[];
  loading: boolean;
  error: boolean;
  onRetry: () => void;
  sponsors: { accountId: string; sponsorName: string }[];
}) {
  const [showTags, setShowTags] = useState(false);
  const term = search.trim().replace(/^#+/, "").toLowerCase();
  const matches = (tags ?? []).filter(
    (tag) => tag.includes(term) && !selected.includes(tag)
  );
  const names = sponsors
    .filter((sponsor) => sponsor.sponsorName.toLowerCase().includes(term))
    .slice(0, 5);
  return (
    <div className="flex flex-col gap-4">
      <FieldGroup>
        <Field>
          <FieldLabel htmlFor="sponsor-search">
            Search sponsors or hashtags
          </FieldLabel>
          <Input
            id="sponsor-search"
            value={search}
            maxLength={100}
            placeholder="Try a sponsor name or #gaming"
            onChange={(event) => onSearch(event.target.value)}
            aria-describedby="sponsor-search-help"
          />
          <FieldDescription id="sponsor-search-help">
            Results update as you type. Select tags to narrow your search.
          </FieldDescription>
        </Field>
      </FieldGroup>
      {term && (
        <div className="flex flex-col gap-3" aria-live="polite">
          {loading ? (
            <p className="text-muted-foreground text-sm">Searching…</p>
          ) : error ? null : (
            <>
              <p className="text-sm font-medium">Suggested tags</p>
              <div className="flex flex-wrap gap-2">
                {matches.slice(0, 8).map((tag) => (
                  <Button
                    key={tag}
                    size="sm"
                    variant="outline"
                    disabled={selected.length >= 10}
                    onClick={() => onAdd(tag)}
                  >
                    #{tag}
                  </Button>
                ))}
                {!matches.length && (
                  <p className="text-muted-foreground text-sm">
                    No matching tags.
                  </p>
                )}
              </div>
              <p className="text-sm font-medium">Matching sponsor names</p>
              {names.map((sponsor) => (
                <a
                  key={sponsor.accountId}
                  className="text-primary text-sm hover:underline"
                  href={`#sponsor-${sponsor.accountId}`}
                >
                  {sponsor.sponsorName}
                </a>
              ))}
              {!names.length && (
                <p className="text-muted-foreground text-sm">
                  No matching names. Check the results below for sponsors with
                  matching tags.
                </p>
              )}
            </>
          )}
        </div>
      )}
      <div className="flex flex-wrap items-center gap-2">
        <Button
          variant="outline"
          aria-expanded={showTags}
          aria-controls="all-sponsor-tags"
          onClick={() => setShowTags(!showTags)}
        >
          All tags
        </Button>
        {(selected.length > 0 || search) && (
          <Button variant="ghost" onClick={onClear}>
            Clear filters
          </Button>
        )}
      </div>
      {showTags && (
        <div
          id="all-sponsor-tags"
          className="flex max-h-64 flex-wrap gap-2 overflow-y-auto"
          aria-live="polite"
        >
          {loading ? (
            <p className="text-muted-foreground text-sm">Loading tags…</p>
          ) : error ? (
            <>
              <p className="text-muted-foreground text-sm">
                Could not load tags.
              </p>
              <Button variant="outline" onClick={onRetry}>
                Retry tags
              </Button>
            </>
          ) : !tags?.length ? (
            <p className="text-muted-foreground text-sm">
              No tags available yet. Tags appear here when sponsors add them to
              their profiles.
            </p>
          ) : (
            tags.map((tag) => (
              <Button
                key={tag}
                size="sm"
                variant="outline"
                disabled={selected.includes(tag) || selected.length >= 10}
                onClick={() => onAdd(tag)}
              >
                #{tag}
              </Button>
            ))
          )}
        </div>
      )}
      {selected.length > 0 && (
        <div className="flex flex-col gap-2">
          <p className="text-muted-foreground text-sm">
            Must match all selected tags ({selected.length}/10)
          </p>
          <div className="flex flex-wrap gap-2">
            {selected.map((tag) => (
              <Badge key={tag} variant="secondary">
                #{tag}
                <Button
                  variant="ghost"
                  size="icon-xs"
                  aria-label={`Remove filter ${tag}`}
                  onClick={() => onRemove(tag)}
                >
                  <X />
                </Button>
              </Badge>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
