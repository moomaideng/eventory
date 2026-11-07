// Shared workload for the Eventory load tests.
//
// Every request goes through the Traefik gateway (BASE_URL), so the numbers
// include gateway -> service -> database and the Tournament -> Account gRPC hop.
//
// Mix per iteration (weights add up to 100):
//   30 list tournaments      Tournament svc + Postgres, public
//   22 tournament detail     Tournament svc + Postgres, public
//   12 bracket               Tournament svc + Postgres, public
//   12 sponsorship tiers     Tournament svc + Postgres, public
//   22 GET /accounts/me      Account svc + Postgres, authenticated
//    2 create tournament     Tournament svc -> Account (gRPC) -> Postgres, authenticated write
//
// Writes are kept rare on purpose: every created tournament stays in the table and
// makes later list queries slower, which would blur "more load" with "more data".
import http from "k6/http";
import exec from "k6/execution";
import { Counter } from "k6/metrics";
import { check, sleep } from "k6";

export const BASE_URL = __ENV.BASE_URL || "http://localhost:8080";

// Dev-mode bearer tokens (backend/internal/middlewares/auth.go).
const COMPETITOR_HEADERS = { Authorization: "Bearer dev-token" };
const ORGANIZER_HEADERS = {
  Authorization: "Bearer dev-token-organizer",
  "Content-Type": "application/json",
};

// Non-2xx responses counted by exact status so the summary shows *why* requests
// failed. Status 0 is a transport failure (timeout, connection reset).
const FAIL_STATUSES = [0, 429, 500, 502, 503, 504];
const failCounters = {};
for (const s of FAIL_STATUSES)
  failCounters[s] = new Counter(`fail_status_${s}`);
const failOther = new Counter("fail_status_other");

function countFailure(res, expected) {
  if (res.status === expected) return;
  (failCounters[res.status] || failOther).add(1);
}

// Seeded fixtures are read from the live API so the script has no hard-coded IDs.
export function setup() {
  const list = http.get(`${BASE_URL}/api/v1/tournaments?pageSize=50`);
  const me = http.get(`${BASE_URL}/api/v1/accounts/me`, {
    headers: COMPETITOR_HEADERS,
  });
  if (list.status !== 200 || me.status !== 200) {
    throw new Error(
      `Stack not ready at ${BASE_URL}: list=${list.status} me=${me.status}. Run "task dev:backend && task seed".`,
    );
  }
  const ids = list.json("items").map((t) => t.id);
  if (ids.length === 0) {
    throw new Error('No tournaments found. Run "task seed" first.');
  }
  return { ids };
}

// Label for the current moment of the run, derived from elapsed test time.
// `schedule` is [{ until: seconds, name }] in ascending order.
export function phaseAt(schedule) {
  const elapsed = exec.instance.currentTestRunDuration / 1000;
  for (const p of schedule) {
    if (elapsed < p.until) return p.name;
  }
  return schedule[schedule.length - 1].name;
}

function newTournamentBody() {
  const day = 24 * 3600 * 1000;
  const start = Date.now() + 60 * day;
  return JSON.stringify({
    name: `k6-load-${exec.vu.idInTest}-${exec.vu.iterationInInstance}`,
    description: "Created by the k6 load test",
    game: "Valorant",
    location: "Online",
    startAt: new Date(start).toISOString(),
    endAt: new Date(start + 8 * 3600 * 1000).toISOString(),
    registrationDeadline: new Date(start - day).toISOString(),
    entryFee: 0,
    registrationMode: "SOLO",
    minTeamSize: 1,
    maxTeamSize: 1,
    capacity: 16,
  });
}

function pick(ids) {
  return ids[Math.floor(Math.random() * ids.length)];
}

// One user iteration. `phase` tags every request so results can be split by
// stage (stress step, spike/recovery) via threshold sub-metrics.
export function iteration(data, phase) {
  const r = Math.random() * 100;
  let res;
  let name;
  let ok = 200;

  if (r < 30) {
    name = "list_tournaments";
    res = http.get(`${BASE_URL}/api/v1/tournaments?pageSize=12`, {
      tags: { name, phase },
    });
  } else if (r < 52) {
    name = "tournament_detail";
    res = http.get(`${BASE_URL}/api/v1/tournaments/${pick(data.ids)}`, {
      tags: { name, phase },
    });
  } else if (r < 64) {
    name = "bracket";
    res = http.get(`${BASE_URL}/api/v1/tournaments/${pick(data.ids)}/bracket`, {
      tags: { name, phase },
    });
  } else if (r < 76) {
    name = "sponsorship_tiers";
    res = http.get(
      `${BASE_URL}/api/v1/tournaments/${pick(data.ids)}/sponsorship-tiers`,
      {
        tags: { name, phase },
      },
    );
  } else if (r < 98) {
    name = "account_me";
    res = http.get(`${BASE_URL}/api/v1/accounts/me`, {
      headers: COMPETITOR_HEADERS,
      tags: { name, phase },
    });
  } else {
    name = "create_tournament";
    ok = 201;
    res = http.post(`${BASE_URL}/api/v1/tournaments`, newTournamentBody(), {
      headers: ORGANIZER_HEADERS,
      tags: { name, phase },
    });
  }

  check(res, { [`${name} status ${ok}`]: (x) => x.status === ok }, { phase });
  countFailure(res, ok);
  sleep(0.3 + Math.random() * 0.4); // think time 0.3-0.7 s
}

// Keep the full console summary and also persist machine-readable results.
export function summaryHandler(name, textSummary) {
  return (data) => ({
    stdout: textSummary(data, { indent: " ", enableColors: true }),
    [`loadtest/results/${name}-summary.json`]: JSON.stringify(data, null, 2),
  });
}

// Thresholds common to all tests. `abortOnFail` is off so the run always completes
// and the report shows where each limit was crossed.
export const baseThresholds = {
  http_req_failed: ["rate<0.01"],
  http_req_duration: ["p(95)<500"],
};

// Sub-metric thresholds that always pass; they exist only so k6 reports
// per-phase latency and error rate in the end-of-test summary.
export function phaseThresholds(names) {
  const t = {};
  for (const n of names) {
    t[`http_req_duration{phase:${n}}`] = ["max>=0"];
    t[`http_req_failed{phase:${n}}`] = ["rate>=0"];
  }
  return t;
}
