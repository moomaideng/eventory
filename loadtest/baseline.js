// Average-load baseline: flat 20 VUs for ~1 min. Reference for stress and spike.
import { textSummary } from 'https://jslib.k6.io/k6-summary/0.1.0/index.js';
import { setup, iteration, summaryHandler, baseThresholds } from './lib.js';

export { setup };

export const options = {
  scenarios: {
    baseline: {
      executor: 'ramping-vus',
      stages: [
        { duration: '10s', target: 20 },
        { duration: '40s', target: 20 },
        { duration: '5s', target: 0 },
      ],
    },
  },
  thresholds: baseThresholds,
};

export default function (data) {
  iteration(data, 'baseline');
}

export const handleSummary = summaryHandler('baseline', textSummary);
