export const suggestedSponsorTags = [
  "Technology",
  "Gaming & Esports",
  "Food & Beverage",
  "Fashion & Apparel",
  "Health & Wellness",
  "Education",
  "Finance",
  "Travel & Hospitality",
  "Media & Entertainment",
] as const;

export function normalizeSponsorTags(tags: string[]) {
  return [
    ...new Set(tags.map((tag) => tag.trim().toLowerCase()).filter(Boolean)),
  ];
}
