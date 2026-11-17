# Epic B Playwright tests

The browser suite covers all 21 test-plan cases: US3-4 (TC51–55), US3-5
(TC56–58), US3-6 (TC59–61), and US3-1 (TC62–71). Each story has a setup file
under `epic-b/initial-state/` and a spec under `epic-b/test/`.

## Run

Start the local backend and frontend with development authentication enabled.
The account seeds and the published SEA Valorant tournament are required for
Alex Dev login and US3-4. Install dependencies and Chromium once:

```bash
cd frontend
npm ci
npx playwright install chromium
npm run test:e2e
```

The suite expects the frontend at `http://localhost:3000` and the API at
`http://localhost:8080`. If the API uses another port, set `PLAYWRIGHT_API_URL`,
for example `PLAYWRIGHT_API_URL=http://localhost:18080 npm run test:e2e`.
Run one case with `npm run test:e2e -- -g TC56`; use `npm run test:e2e:ui` to inspect
steps and `npx playwright show-report` for the HTML report.

Each browser test starts at the home page and navigates by clicking visible
controls; TC56 and TC57 refresh the page to verify saved state. Tests use fresh
browser contexts. TC56 creates a Draft and opens Registration
to check the player form. TC59–61 create fresh free-entry tournaments and
applications through the API, then review them in the browser. TC62 creates a
Draft, and TC51 posts an announcement to the seeded tournament. These records
remain in the disposable local database; their tournament names begin with
`Playwright Number Form`, `Playwright Review`, or `Test Cup`, and the announcement
title begins with `Venue update`. Do not run this suite against shared or
production data.
