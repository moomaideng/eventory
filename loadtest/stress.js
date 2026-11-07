// Stress test: step the load up beyond normal to find where the system degrades.
// Each step is a 10 s ramp (tagged `ramp`) + 30 s hold (tagged `vu<N>`), so the
// summary shows latency and error rate per steady level without ramp blur.
// A previous run was healthy at 100 VUs and broken at 200, hence the fine steps there.
import { textSummary } from 'https://jslib.k6.io/k6-summary/0.1.0/index.js';
import { setup, iteration, phaseAt, summaryHandler, baseThresholds, phaseThresholds } from './lib.js';

export { setup };

const levels = [50, 100, 150, 200, 300];
const RAMP = 10;
const HOLD = 30;

const stages = [];
const schedule = [];
let t = 0;
for (const target of levels) {
  stages.push({ duration: `${RAMP}s`, target }, { duration: `${HOLD}s`, target });
  schedule.push({ until: (t += RAMP), name: 'ramp' }, { until: (t += HOLD), name: `vu${target}` });
}
stages.push({ duration: '5s', target: 0 });

export const options = {
  scenarios: {
    stress: { executor: 'ramping-vus', stages, gracefulRampDown: '10s' },
  },
  thresholds: {
    ...baseThresholds,
    ...phaseThresholds(['ramp', ...levels.map((l) => `vu${l}`)]),
  },
};

export default function (data) {
  iteration(data, phaseAt(schedule));
}

export const handleSummary = summaryHandler('stress', textSummary);
