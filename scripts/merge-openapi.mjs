#!/usr/bin/env node
/**
 * Merge public service contracts for openapi-typescript (ports 8081–8084).
 */
import { writeFileSync } from "node:fs";

const sources = [
  { name: "account", url: process.env.ACCOUNT_OPENAPI_URL || "http://127.0.0.1:8081/openapi.json" },
  { name: "tournament", url: process.env.TOURNAMENT_OPENAPI_URL || "http://127.0.0.1:8082/openapi.json" },
  { name: "payment", url: process.env.PAYMENT_OPENAPI_URL || "http://127.0.0.1:8083/openapi.json" },
  { name: "notification", url: process.env.NOTIFICATION_OPENAPI_URL || "http://127.0.0.1:8084/openapi.json" },
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
  for (const [path, operation] of Object.entries(doc.paths || {})) {
    if (path.startsWith("/internal/")) continue;
    if (source.name === "payment" && !path.startsWith("/api/v1/payments")) continue;
    if (source.name === "notification" && !path.startsWith("/api/v1/notifications")) continue;
    merged.paths[path] = operation;
  }
  Object.assign(merged.components.schemas, doc.components?.schemas || {});
  Object.assign(merged.components.securitySchemes, doc.components?.securitySchemes || {});
}

const out = new URL("../frontend/.openapi.merged.json", import.meta.url);
writeFileSync(out, JSON.stringify(merged, null, 2));
console.log(`Wrote ${out.pathname}`);
