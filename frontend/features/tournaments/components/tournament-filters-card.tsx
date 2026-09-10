"use client";

import React, { useRef, useState } from "react";
import { Search } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Field,
  FieldError,
  FieldGroup,
  FieldLabel,
} from "@/components/ui/field";
import { Input } from "@/components/ui/input";

interface TournamentFiltersCardProps {
  searchKey: string;
  defaultQ: string;
  defaultStartFrom: string;
  defaultStartTo: string;
  defaultMaxEntryFee: string;
  onApplyFilters: (event: React.FormEvent<HTMLFormElement>) => void;
  onClearFilters: () => void;
}

export function TournamentFiltersCard({
  searchKey,
  defaultQ,
  defaultStartFrom,
  defaultStartTo,
  defaultMaxEntryFee,
  onApplyFilters,
  onClearFilters,
}: TournamentFiltersCardProps) {
  const [startFrom, setStartFrom] = useState(defaultStartFrom);
  const [startTo, setStartTo] = useState(defaultStartTo);
  const [prevSearchKey, setPrevSearchKey] = useState(searchKey);
  const formRef = useRef<HTMLFormElement>(null);

  if (searchKey !== prevSearchKey) {
    setPrevSearchKey(searchKey);
    setStartFrom(defaultStartFrom);
    setStartTo(defaultStartTo);
  }

  const isInvalidDateRange = Boolean(
    startFrom && startTo && startFrom > startTo
  );

  function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    if (isInvalidDateRange) {
      event.preventDefault();
      return;
    }
    onApplyFilters(event);
  }

  function handleClear() {
    setStartFrom("");
    setStartTo("");
    formRef.current?.reset();
    onClearFilters();
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Search and filters</CardTitle>
        <CardDescription>
          Dates and times use Bangkok time. Entry fees are currently listed in
          Thai baht.
        </CardDescription>
      </CardHeader>
      <CardContent>
        <form
          ref={formRef}
          id="tournament-filters"
          key={searchKey}
          onSubmit={handleSubmit}
        >
          <FieldGroup className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
            <Field>
              <FieldLabel htmlFor="tournament-search">Search</FieldLabel>
              <Input
                id="tournament-search"
                name="q"
                defaultValue={defaultQ}
                placeholder="Name, game, or keyword"
                maxLength={100}
              />
            </Field>
            <Field data-invalid={isInvalidDateRange || undefined}>
              <FieldLabel htmlFor="start-from">From</FieldLabel>
              <Input
                id="start-from"
                name="startFrom"
                type="date"
                value={startFrom}
                onChange={(e) => setStartFrom(e.target.value)}
                max={startTo || undefined}
                aria-invalid={isInvalidDateRange || undefined}
              />
            </Field>
            <Field data-invalid={isInvalidDateRange || undefined}>
              <FieldLabel htmlFor="start-to">To</FieldLabel>
              <Input
                id="start-to"
                name="startTo"
                type="date"
                value={startTo}
                onChange={(e) => setStartTo(e.target.value)}
                min={startFrom || undefined}
                aria-invalid={isInvalidDateRange || undefined}
              />
              {isInvalidDateRange ? (
                <FieldError>To date must be on or after From date.</FieldError>
              ) : null}
            </Field>
            <Field>
              <FieldLabel htmlFor="maximum-fee">Maximum fee (THB)</FieldLabel>
              <Input
                id="maximum-fee"
                name="maxEntryFee"
                type="number"
                min={0}
                step={1}
                defaultValue={defaultMaxEntryFee}
                placeholder="Any budget"
              />
            </Field>
          </FieldGroup>
        </form>
      </CardContent>
      <CardFooter className="flex flex-wrap justify-end gap-2">
        <Button type="button" variant="ghost" onClick={handleClear}>
          Clear filters
        </Button>
        <Button
          type="submit"
          form="tournament-filters"
          disabled={isInvalidDateRange}
        >
          <Search data-icon="inline-start" />
          Apply filters
        </Button>
      </CardFooter>
    </Card>
  );
}
