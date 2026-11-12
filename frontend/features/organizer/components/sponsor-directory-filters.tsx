"use client";

import { Hash, Plus, X } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
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
}) {
  const term = search.trim().replace(/^#+/, "").toLowerCase();
  const matches = (tags ?? []).filter(
    (tag) => tag.includes(term) && !selected.includes(tag)
  );
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
      {selected.length > 0 && (
        <Card size="sm">
          <CardHeader>
            <CardTitle>Selected tags ({selected.length}/10)</CardTitle>
            <CardDescription>
              Sponsors must have every tag below. Click × to remove a filter.
            </CardDescription>
          </CardHeader>
          <CardContent className="flex-row flex-wrap gap-2">
            {selected.map((tag) => (
              <Badge key={tag}>
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
          </CardContent>
        </Card>
      )}
      {term && (
        <div className="flex flex-col gap-3" aria-live="polite">
          {loading ? (
            <p className="text-muted-foreground text-sm">Searching…</p>
          ) : error ? null : (
            <Card size="sm">
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <Hash className="size-5" aria-hidden="true" />
                  Suggested tags
                </CardTitle>
                <CardDescription>
                  Add a tag to narrow the results.
                </CardDescription>
              </CardHeader>
              <CardContent className="flex-row flex-wrap gap-2">
                {matches.slice(0, 8).map((tag) => (
                  <Button
                    key={tag}
                    size="sm"
                    variant="outline"
                    disabled={selected.length >= 10}
                    onClick={() => onAdd(tag)}
                  >
                    <Plus data-icon="inline-start" />#{tag}
                  </Button>
                ))}
                {!matches.length && (
                  <p className="text-muted-foreground text-sm">
                    No matching tags.
                  </p>
                )}
              </CardContent>
            </Card>
          )}
        </div>
      )}
      {(selected.length > 0 || search) && (
        <Button className="self-start" variant="ghost" onClick={onClear}>
          Clear filters
        </Button>
      )}
    </div>
  );
}
