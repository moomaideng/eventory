import assert from "node:assert/strict";
import test from "node:test";
import { normalizeSponsorTags, parseSponsorTagDraft } from "./tags.ts";

test("pasted hashtags are separated and deduplicated, including Thai", () => {
  assert.deepEqual(parseSponsorTagDraft("#Gaming #valorant#กีฬา #GAMING"), [
    "gaming",
    "valorant",
    "กีฬา",
  ]);
});
test("plain text and commas work without requiring punctuation", () => {
  assert.deepEqual(parseSponsorTagDraft("gaming, chess volleyball"), [
    "gaming",
    "chess",
    "volleyball",
  ]);
});
test("legacy tags with spaces survive saving and hashtag prefixes are removed", () => {
  assert.deepEqual(
    normalizeSponsorTags(["Gaming & Esports", "#VALORANT", "valorant"]),
    ["gaming & esports", "valorant"]
  );
});
test("empty hashtag input does not create empty tags", () => {
  assert.deepEqual(parseSponsorTagDraft(" # ## , "), []);
});
