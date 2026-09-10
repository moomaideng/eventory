"use client";

import React from "react";
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
import { Field, FieldGroup, FieldLabel } from "@/components/ui/field";
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
          id="tournament-filters"
          key={searchKey}
          onSubmit={onApplyFilters}
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
            <Field>
              <FieldLabel htmlFor="start-from">Starts after</FieldLabel>
              <Input
                id="start-from"
                name="startFrom"
                type="date"
                defaultValue={defaultStartFrom}
              />
            </Field>
            <Field>
              <FieldLabel htmlFor="start-to">Starts before</FieldLabel>
              <Input
                id="start-to"
                name="startTo"
                type="date"
                defaultValue={defaultStartTo}
              />
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
        <Button type="button" variant="ghost" onClick={onClearFilters}>
          Clear filters
        </Button>
        <Button type="submit" form="tournament-filters">
          <Search data-icon="inline-start" />
          Apply filters
        </Button>
      </CardFooter>
    </Card>
  );
}
