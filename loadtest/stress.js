// Stress test: step the load up beyond normal to find where the system degrades.
// The steps were chosen around the old limit (healthy at 100 VUs, broken at 200).
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

const { stages, schedule, phases } = steppedStages([50, 100, 150, 200, 300]);

export const options = {
  scenarios: {
    stress: { executor: "ramping-vus", stages, gracefulRampDown: "10s" },
  },
  thresholds: { ...baseThresholds, ...phaseThresholds(phases) },
};

export default function (data) {
  iteration(data, phaseAt(schedule));
}

export const handleSummary = summaryHandler("stress", textSummary);
