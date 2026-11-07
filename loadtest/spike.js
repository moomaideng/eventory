// Spike test: sudden 30x surge from a light steady load, then back down.
// Phases: warmup (10 VUs) -> spike (300 VUs, incl. 5 s ramps) -> recovery (10 VUs).
// Comparing `recovery` with `warmup` shows whether the system returns to normal.
import { textSummary } from 'https://jslib.k6.io/k6-summary/0.1.0/index.js';
import { setup, iteration, phaseAt, summaryHandler, baseThresholds, phaseThresholds } from './lib.js';

export { setup };

// Boundaries in seconds from test start; must match `stages` below.
const schedule = [
  { until: 30, name: 'warmup' }, // 10 s ramp + 20 s hold
  { until: 70, name: 'spike' }, // 5 s up + 30 s hold + 5 s down
  { until: 130, name: 'recovery' }, // 60 s at 10 VUs, then 5 s ramp-down
];

export const options = {
  scenarios: {
    spike: {
      executor: 'ramping-vus',
      gracefulRampDown: '10s',
      stages: [
        { duration: '10s', target: 10 },
        { duration: '20s', target: 10 },
        { duration: '5s', target: 300 },
        { duration: '30s', target: 300 },
        { duration: '5s', target: 10 },
        { duration: '60s', target: 10 },
        { duration: '5s', target: 0 },
      ],
    },
  },
  thresholds: {
    ...baseThresholds,
    ...phaseThresholds(schedule.map((s) => s.name)),
  },
};

export default function (data) {
  iteration(data, phaseAt(schedule));
}

export const handleSummary = summaryHandler('spike', textSummary);
