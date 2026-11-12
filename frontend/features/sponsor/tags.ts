export function normalizeSponsorTags(tags: string[]) {
  return [
    ...new Set(
      tags
        .map((tag) => tag.trim().replace(/^#+/, "").trim().toLowerCase())
        .filter(Boolean)
    ),
  ];
}

// Preserve spaces in saved legacy tags; tokenize only newly typed hashtags.
export function parseSponsorTagDraft(draft: string) {
  return normalizeSponsorTags(draft.split(/[\s#,]+/u));
}
