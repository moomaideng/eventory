# Project Description

## Project Name

Eventory

## Group Members 

- Ashira Aungsumal 6732041921
- Kittichet Arayasujin 6732005321
- Kittichon Chaonawig 6731305421
- Pasin Thanyakasikol 6732025921
- Peeravas Piboolvorakul 6732033921

## Problem Description

Running a tournament today splits work across many tools: organizers post rules and brackets in chat apps or social media, collect entries with forms and spreadsheets, chase payments over bank transfer, and update results by hand. Competitors hunt for events, join teams, and submit forms in different places. Sponsors have no shared place to pick a package, pay, and see their support shown on the event.

Eventory puts those flows on one platform. An organizer creates a tournament (with optional crowdfunding before open registration) and a registration form. A sponsor pledges a package, pays, and appears on the tournament page as funding progresses. A team captain locks the roster, pays any entry fee, and submits the registration. Match results are recorded on the same tournament so standings stay visible in one place.

## Target Customers

- **Organizer**: hosts and runs tournaments
- **Sponsor**: funds tournaments via packages/pledges
- **Competitor**: joins tournaments

## Scenario (use-case & description)

### UC-01: Create a Tournament

**Primary Actor**: Organizer

**Goal**: Set up a tournament on the platform (create a draft, configure crowdfunding and the registration form as needed, then open for participants).

**Preconditions**: Organizer is signed in and has an active Organizer profile.

**Main Flow**:

1. Organizer opens tournament creation and enters event details (title, dates, rules, solo/team format, entry fee, rewards).
2. System creates the tournament as a draft and shows the organizer dashboard.
3. Organizer configures the draft before opening:
   - Optionally enables crowdfunding (funding goal, deadline, sponsorship packages/tiers).
   - Configures the custom registration form (fields such as text, dropdown, file upload).
   - May update crowdfunding and the form again while the tournament is still a draft.
4. When ready, organizer opens the tournament for sponsorship and/or competitor registration.

**Postcondition**: A tournament exists with its registration form configured and is open according to the organizer’s chosen path.

**Outcome**: The organizer has one place to define the event, fund it if needed, and accept registrations, without chat apps, forms, and spreadsheets.

**Alternate/Exceptional Flow:**

1. **Invalid or incomplete details**: System blocks create/save, highlights missing or invalid fields (e.g. end before start, empty title), and no tournament is published.
2. **Invalid crowdfunding setup**: System rejects a zero/negative goal or a past deadline; tournament stays unpublished for that campaign.
3. **Incomplete registration form**: System blocks saving a field with no label or a dropdown without enough options; form stays as last valid version.

### UC-02: Sponsor a Tournament

**Primary Actor**: Sponsor

**Goal**: Support a tournament by choosing a package/pledge, paying, and appearing on the tournament page as funding progress updates.

**Preconditions**: Sponsor is signed in and has an active Sponsor profile. Tournament is open for sponsorship (including crowdfunding stage when used).

**Main Flow**:

1. Sponsor opens the tournament page and reviews available sponsorship packages/tiers and funding progress.
2. Sponsor selects a package, uploads brand logo (and link if required), and starts checkout.
3. System processes payment via the payment gateway.
4. On success, system records the pledge, updates funding progress, and shows the sponsor (e.g. logo) on the tournament page.

**Postcondition**: A confirmed sponsorship exists; funding progress and public sponsor display are updated.

**Outcome**: The sponsor funds the event in one flow and is visibly recognized on the tournament.

**Alternate/Exceptional Flow:**

1. **Payment declined or fails**: System informs the sponsor, no pledge is confirmed, funding and public display stay unchanged; sponsor can retry.
2. **Package sold out or unavailable**: System blocks checkout and explains that the tier has no remaining slots.

### UC-03: Submit Registration

**Primary Actor**: Competitor (solo player, or Team Captain for team events)

**Goal**: Register for a tournament as a solo entry, or as a team after locking the roster; pay any entry fee, and submit the registration form to the organizer.

**Preconditions**: Competitor is signed in with a completed profile. Tournament registration is open. For team events, a team lobby exists with members joined via invite before lock.

**Main Flow**:

1. Competitor opens the tournament and starts registration.
2. Registration follows one of these branches:
   - **2a. Solo**: Competitor fills the tournament registration form and proceeds to pay entry fee if any.
   - **2b. Team**: Team members join the captain’s lobby via invite. When the roster meets tournament size rules, the Team Captain locks the roster, completes the registration form, and pays the team entry fee if any.
3. System processes payment when required (payment gateway).
4. On success, system submits the registration to the organizer for review/acceptance and confirms to the competitor.

**Postcondition**: A registration entry exists and is visible to the organizer for review.

**Outcome**: The competitor (or team) is officially entered without separate forms and bank-transfer chasing.

**Alternate/Exceptional Flow:**

1. **Payment fails**: System informs the competitor; solo registration is not submitted; for teams the roster stays unlocked and the team is not sent to the organizer; retry is allowed.
2. **Incomplete team roster**: System blocks lock/submit until the roster meets required size; lobby stays forming.
3. **Invalid or incomplete registration form**: System rejects submit, shows which required fields/files are missing or invalid, and does not create a submitted entry.

### UC-04: Record Match Results

**Primary Actor**: Organizer or Tournament Staff

**Goal**: Record the result of a match so standings/bracket on the tournament page stay up to date.

**Preconditions**: Actor is authorized for that tournament (owner organizer or assigned staff). Match exists and is ready for a result.

**Main Flow**:

1. Actor opens the match (from bracket/schedule) for the tournament.
2. Actor enters the final result (e.g. scores or placement as the format requires).
3. System validates and saves the official result.
4. System updates the tournament view (bracket progression and/or standings) so competitors and others see the outcome.

**Postcondition**: The match has an official recorded result; public tournament results reflect it.

**Outcome**: Results live on the same platform as the event, with no manual spreadsheet or chat updates.

**Alternate/Exceptional Flow:**

1. **Unauthorized actor**: System rejects the update; match result unchanged.
2. **Invalid result**: System rejects (e.g. disallowed tie for an elimination match) and does not advance or update standings.

## Functional Requirements

*Format: `[Condition] [Subject] [Action] [Object] [Constraint]` (condition and constraint optional).*

### Authentication

- FR1: The system shall authenticate users with Google OAuth.
- FR2: On a user's first successful sign-in, the system shall create a new account.
- FR3: When a signed-in user signs out, the system shall end the session.
- FR4: When an active account submits valid Organizer or Sponsor profile details, the system shall create the corresponding profile for that account.
- FR5: When an active account requests a mode change, the system shall switch that account among Competitor, Organizer, and Sponsor modes under one login.

### Tournament Hosting

- FR6: When an organizer submits required event details (title, dates, rules, solo/team format, entry fee, rewards), the system shall create a tournament draft.
- FR7: When an organizer sets a funding goal, deadline, and sponsorship packages/tiers on a draft tournament, the system shall configure crowdfunding for that tournament.
- FR8: When an organizer defines form fields on a draft tournament, the system shall configure the custom registration form for that tournament.
- FR9: When an organizer changes crowdfunding or registration form settings on a draft tournament, the system shall update those draft settings.
- FR10: When an organizer chooses to open a configured tournament, the system shall open the tournament for sponsorship and/or competitor registration.

### Sponsorship

- FR11: When a tournament is open for sponsorship, the system shall display available sponsorship packages/tiers and funding progress.
- FR12: When a sponsor submits a package selection and brand logo for checkout, the system shall process the checkout through the payment gateway.
- FR13: On successful payment, the system shall record the confirmed sponsorship with updated funding progress and public sponsor display on the tournament page.
- FR14: When payment fails or the selected package has no remaining slots, the system shall keep the sponsorship unconfirmed with a failure notice to the sponsor.

### Registration

FR15 is for Solo, FR16-19 for Team, and FR20-21 are shared.

- FR15: When a competitor completes the registration form for a solo-format tournament, the system shall accept the form and process any entry fee through the payment gateway (or continue if free).
- FR16: When a competitor creates a team for a team-format tournament, the system shall create a team lobby with an invite.
- FR17: When a competitor joins with a valid invite, the system shall add the competitor to the team lobby before registration is submitted.
- FR18: When a team captain submits a roster that meets tournament size rules, the system shall lock the team roster for registration submission.
- FR19: When a team captain submits the registration form for a locked roster, the system shall process any team entry fee through the payment gateway.
- FR20: On successful payment when required (or immediately when free), the system shall submit the registration to the organizer for review with confirmation to the competitor.
- FR21: When payment fails or a team roster fails required size checks, the system shall keep the registration unsubmitted with the team roster unlocked.

### Match Results

- FR22: When an authorized organizer or tournament staff submits a final match result, the system shall record the result for that match.
- FR23: After a valid result is saved, the system shall update the tournament bracket and/or standings on the tournament page.
- FR24: When an unauthorized actor submits a result or the result is invalid (e.g. a disallowed tie in an elimination match), the system shall keep the match unchanged with an error notice to the actor.

## Non-functional Requirements

### Security & Privacy

- NFR1: The system shall authenticate users with Google OAuth tokens only.
- NFR2: The system shall process payments through an external payment gateway that handles card data.
- NFR3: The system shall accept match result recording from the tournament owner organizer or assigned tournament staff only.

### Usability

- NFR4: The system shall present Competitor, Organizer, and Sponsor mode switching in the web UI under one login.

### Reliability

- NFR5: When a fee or pledge applies and payment succeeds, the system shall confirm the related registration or sponsorship.
- NFR6: When a fee or pledge applies and payment fails, the system shall keep the previous registration or sponsorship state (team roster unlocked; sponsorship and registration unconfirmed).

### Performance

- NFR7: Under normal course-demo load, the system shall complete competitor registration confirmation with p95 latency under 3 seconds, excluding external payment-gateway processing time.
