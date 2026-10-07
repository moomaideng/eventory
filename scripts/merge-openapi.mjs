#!/usr/bin/env node
/**
 * Merge Account and Tournament OpenAPI docs into one file for openapi-typescript.
 * Services must be reachable on the ports published by Compose (8081 and 8082).
 */
import { writeFileSync } from "node:fs";

const sources = [
  { name: "account", url: process.env.ACCOUNT_OPENAPI_URL || "http://127.0.0.1:8081/openapi.json" },
  { name: "tournament", url: process.env.TOURNAMENT_OPENAPI_URL || "http://127.0.0.1:8082/openapi.json" },
];

const merged = {
  openapi: "3.1.0",
  info: { title: "Eventory API", version: "1.0.0" },
  paths: {},
  components: { schemas: {}, securitySchemes: {} },
};

for (const source of sources) {
  const response = await fetch(source.url);
  if (!response.ok) {
    throw new Error(`Failed to fetch ${source.name} OpenAPI from ${source.url}: ${response.status}`);
  }
  const doc = await response.json();
  Object.assign(merged.paths, doc.paths || {});
  Object.assign(merged.components.schemas, doc.components?.schemas || {});
  Object.assign(merged.components.securitySchemes, doc.components?.securitySchemes || {});
}

const out = new URL("../frontend/.openapi.merged.json", import.meta.url);
writeFileSync(out, JSON.stringify(merged, null, 2));
console.log(`Wrote ${out.pathname}`);
