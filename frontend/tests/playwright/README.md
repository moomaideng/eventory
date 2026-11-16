# Playwright browser tests

Each user story has two corresponding files:

```text
epic-b/
  initial-state/us3-4.setup.ts  # Common login and starting page
  test/us3-4.spec.ts           # One test() per TC
```

The spec imports `test` and `expect` from its story setup:

```ts
import { test, expect } from '@/tests/playwright/epic-b/initial-state/us3-4.setup';
```

US3-4 logs in as Alex once per worker, keeps login state in memory, and opens
SEA Valorant's Announcements tab before every test. Each test gets a fresh
browser context. No saved session JSON is required or shared between machines.
Keep the TC actions and assertions in the spec; keep common preparation in the
corresponding setup file. Other stories should use their own setup files.
Existing unit tests remain beside their feature code.

## Prepare

From the repository root, run `task dev`. In another terminal, run `task seed`
if needed. Each machine needs the seeded SEA Valorant tournament owned by Alex.
Dev Quick Login requires the development server.

Install Chromium once from `frontend/`:

```bash
npx playwright install chromium
```

## Run and record

From `frontend/`:

```bash
npm run test:e2e -- tests/playwright/epic-b/test/us3-4.spec.ts
npm run test:codegen -- tests/playwright/epic-b/test/us3-4.spec.ts
npm run test:e2e:ui
npx playwright show-report
```

`test:codegen` runs the story headed with no test timeout. Login and navigation
run automatically, then `page.pause()` opens the recorder on Announcements.
Wait for the terminal's "US3-4 ready" message. Use Record to capture the TC actions and assertions, then copy them into
an individual `test('TC51: ...', ...)` in the spec. Keep the setup import.
Select one test with `-g 'TC51'` once the spec contains multiple cases.
The browser stays open while recording even if you click Resume. Close the
browser when finished; the recording run is marked skipped instead of replaying
the TC against the page you just changed. Ctrl+C also ends the recording.

This command uses the corresponding TypeScript setup and does not accept
codegen's `--output` or `--load-storage` options. It does not overwrite the spec.
Each TC re-runs the story's page preparation; no manual login is needed.

Only Epic B is configured. Epic A's onboarding and account-status cases need
separate authentication and data setup before enabling that project. The current
US3-4 spec checks the starting state; actual report TCs still need to be recorded.

Tests run with one worker and no retries. Browser isolation does not reset the
database: approvals, rejected applications, and created accounts need deliberate
data setup before replaying cases. Failure screenshots and traces go to
`test-results/`; the HTML report goes to `playwright-report/`. Both are ignored.
