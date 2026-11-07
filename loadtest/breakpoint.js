// Breakpoint test: keep stepping the load up until the pass criteria fail, to find
// how many concurrent users the system handles. Run on its own (not in load:all).
// The first level that crosses p95 < 500 ms or errors < 1% is the limit.
import { textSummary } from "https://jslib.k6.io/k6-summary/0.1.0/index.js";
import {
  setup,
  iteration,
  phaseAt,
  summaryHandler,
  baseThresholds,
  phaseThresholds,
  steppedStages,
} from "./lib.js";

export { setup };

const { stages, schedule, phases } = steppedStages([300, 500, 800, 1200, 1600]);

export const options = {
  scenarios: {
    breakpoint: { executor: "ramping-vus", stages, gracefulRampDown: "10s" },
  },
  thresholds: { ...baseThresholds, ...phaseThresholds(phases) },
};

export default function (data) {
  iteration(data, phaseAt(schedule));
}

export const handleSummary = summaryHandler("breakpoint", textSummary);
