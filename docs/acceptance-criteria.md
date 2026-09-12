# Eventory Platform — Acceptance Criteria (E2E Master Specification)

This document serves as the master end-to-end (E2E) acceptance criteria specification for the Eventory platform across **EPIC 1 through EPIC 7**.

Every user story strictly adheres to the **TA Product Backlog Acceptance Criteria standard**, structured with:
1. **User Story Statement** (`As a..., I want..., So that...`)
2. **Core Acceptance Criteria (Given-When-Then)** covering:
   - **Primary Action & Condition** (Happy path flow)
   - **Incomplete / Invalid Input** (Validation error handling)
   - **Success Confirmation** (Visual feedback and indicators)
   - **Business Rules & Permissions** (Access control and edge cases)
3. **Consolidated Gherkin Scenario Outline** with parameterized **`Examples:`** table consolidating input variations and boundary values.

---

## EPIC 1: Account and Role Access

### US1-1: Create Single Primary Account
> **User Story:**
> As a new user,
> I want to register a single primary account as a User
> So that I can have a central identity with a unique public handle to interact across the platform.

#### Acceptance Criteria
- **AC1 (Registration & Handle):**
  - **Given** the user is on the account onboarding page,
  - **When** the user attempts to register an account with a display name and unique handle,
  - **Then** the system should allow the action and grant access to the personal dashboard.
- **AC2 (Validation):**
  - **Given** the user input is incomplete or invalid (e.g. handle already taken, shorter than 3 characters, contains invalid characters, or display name is blank),
  - **When** the user attempts to register an account,
  - **Then** the system should display an appropriate error message and keep the user on the onboarding page.
- **AC3 (Confirmation):**
  - **Given** the system has successfully processed the registration,
  - **When** the user completes onboarding,
  - **Then** the system should confirm success with a visual indicator and redirect to the dashboard.
- **AC4 (Data Security):**
  - **Given** the action involves personal identity information,
  - **When** the user completes account registration,
  - **Then** the system should securely associate the profile with the user's authenticated credentials.

```gherkin
Scenario Outline: Primary account registration and onboarding validation
  Given the user is on the onboarding page
  When the user attempts to register with display name "<displayName>" and handle "<handle>"
  Then the system should <systemAction>
  And the user should see message "<feedbackMessage>"

  Examples:
    | displayName | handle          | systemAction                               | feedbackMessage                                 |
    | Alice Gamer | alice_pro       | grant access and redirect to the dashboard | Welcome to Eventory!                            |
    | Bob The Pro | existing_handle | reject the action and highlight handle     | This handle is already taken                    |
    | Dan Solo    | al              | reject the action and highlight handle     | Handle must be between 3 and 32 characters      |
    | Eve Online  | eve!@#$         | reject the action and highlight handle     | Handle can only contain letters, numbers, and _ |
    |             | sam_valid       | reject the action and highlight name field | Display name is required                        |
```

---

### US1-2: Create and Switch Between Role Profiles
> **User Story:**
> As a user,
> I want to create and switch between Organizer and Sponsor profiles linked to my account
> So that I can manage events or corporate funds without maintaining separate logins.

#### Acceptance Criteria
- **AC1 (Profile Creation):**
  - **Given** the user is authenticated and on the role profile management page,
  - **When** the user attempts to create an Organizer or Sponsor profile with a valid organization name,
  - **Then** the system should allow the action and link the new role profile to the primary account.
- **AC2 (Role Switching):**
  - **Given** an authenticated user with multiple linked role profiles,
  - **When** the user attempts to switch between active roles (Competitor, Organizer, Sponsor),
  - **Then** the system should allow the action without requiring re-authentication and update the active workspace.
- **AC3 (Validation & Access):**
  - **Given** the user attempts to switch to a role profile that has not been created yet,
  - **When** the user selects that unlinked role,
  - **Then** the system should display an appropriate error message and maintain the current active role.
- **AC4 (Confirmation):**
  - **Given** the system has successfully processed the role switch,
  - **When** the workspace reloads,
  - **Then** the system should confirm success with a visual role indicator and role-specific navigation menus.

```gherkin
Scenario Outline: Creating and switching active role profiles
  Given an authenticated user on the dashboard with profiles "<existingProfiles>"
  And the current active role is "<activeRole>"
  When the user attempts to "<action>" profile for role "<targetRole>" with organization name "<orgName>"
  Then the system should <systemAction>
  And the active role indicator should display "<activeRoleBadge>"

  Examples:
    | existingProfiles               | activeRole | action | targetRole | orgName         | systemAction                                | activeRoleBadge |
    | COMPETITOR                     | COMPETITOR | CREATE | ORGANIZER  | Siam Esports    | create organizer profile and link to account| Competitor      |
    | COMPETITOR, ORGANIZER          | COMPETITOR | SWITCH | ORGANIZER  |                 | switch active workspace to Organizer        | Organizer       |
    | COMPETITOR, ORGANIZER          | ORGANIZER  | CREATE | SPONSOR    | Apex Brands Co. | create sponsor profile and link to account  | Organizer       |
    | COMPETITOR, ORGANIZER, SPONSOR | ORGANIZER  | SWITCH | SPONSOR    |                 | switch active workspace to Sponsor          | Sponsor         |
    | COMPETITOR                     | COMPETITOR | SWITCH | ORGANIZER  |                 | display error "Profile not found"           | Competitor      |
    | COMPETITOR, SPONSOR            | SPONSOR    | SWITCH | ORGANIZER  |                 | display error "Profile not found"           | Sponsor         |
```

---

### US1-3: View and Update Role-Specific Profiles
> **User Story:**
> As an account owner,
> I want to view and update my role-specific profile(s)
> So that my contact and public information stays accurate.

#### Acceptance Criteria
- **AC1 (Edit Information):**
  - **Given** the user is on the role profile edit page,
  - **When** the user attempts to edit their profile details (organization name, contact email, phone),
  - **Then** the system should allow the action if the profile belongs to the authenticated user.
- **AC2 (Validation):**
  - **Given** the user input is incomplete or invalid (e.g. blank organization name or malformed email),
  - **When** the user attempts to save profile changes,
  - **Then** the system should display an appropriate error message and prevent saving.
- **AC3 (Confirmation):**
  - **Given** the system has successfully processed the update request,
  - **When** the user saves the changes,
  - **Then** the system should confirm success with a message or visual indicator.

```gherkin
Scenario Outline: Updating role profile contact details
  Given an authenticated user on the "Edit <roleType> Profile" page
  When the user attempts to update organization name to "<orgName>" and contact email to "<email>"
  Then the system should <systemAction>
  And the user should see message "<feedbackNotice>"

  Examples:
    | roleType  | orgName         | email               | systemAction                                | feedbackNotice                     |
    | Organizer | Bob Gaming Int. | contact@bobgame.com | save changes and display updated details    | Profile updated successfully       |
    | Sponsor   | Acme Worldwide  | corp@acme.com       | save changes and display updated details    | Profile updated successfully       |
    | Organizer |                 | valid@bobgame.com   | reject changes and highlight name field     | Organization name cannot be blank  |
    | Organizer | Bob Gaming Int. | not-an-email        | reject changes and highlight email field    | Please enter a valid email address |
```

---

## EPIC 2: Tournament Discovery

### US2-1: Browse and Filter Tournaments
> **User Story:**
> As a user,
> I want to browse and filter tournaments
> So that I can quickly find events that fit my schedule and budget.

#### Acceptance Criteria
- **AC1 (Filter Tournaments):**
  - **Given** the user is on the tournament catalog page,
  - **When** the user attempts to filter tournaments by keyword, date range, or maximum fee,
  - **Then** the system should display all published tournaments matching the applied criteria.
- **AC2 (Date Validation):**
  - **Given** the user inputs an inverted date range (To date is earlier than From date),
  - **When** the user attempts to filter,
  - **Then** the system should display an appropriate error message and disable filter submission.
- **AC3 (Empty State):**
  - **Given** no tournaments match the applied filter criteria,
  - **When** the system completes the search,
  - **Then** the system should display an appropriate empty state indicator.

```gherkin
Scenario Outline: Filtering tournaments by keyword, dates, and maximum budget
  Given the user is on the tournament catalog page
  When the user filters by keyword "<keyword>", From date "<fromDate>", To date "<toDate>", and max fee "<maxFee>"
  Then the system should <systemAction>
  And the catalog should display "<catalogDisplay>"

  Examples:
    | keyword  | fromDate   | toDate     | maxFee | systemAction                                       | catalogDisplay                      |
    | Valorant |            |            |        | update list with matching game title               | Valorant tournaments only           |
    |          | 2026-06-01 | 2026-07-01 |        | update list with matching tournament start dates   | June 2026 tournaments               |
    |          |            |            | 500    | update list with entry fees up to 500 THB          | Events with entry fee <= 500 THB    |
    |          | 2026-10-01 | 2026-09-01 |        | display error "To date must be after From"         | Filter button disabled              |
    | Unknown  |            |            |        | display empty state                                | Empty state: "No tournaments found" |
```

---

### US2-2: View Tournament Details, Teams, and Funding Progress
> **User Story:**
> As a user (potential competitor or sponsor),
> I want to view tournament details, teams, and funding progress
> So that I can decide whether to participate or provide support.

#### Acceptance Criteria
- **AC1 (View Details):**
  - **Given** the user navigates to the tournament details page,
  - **When** the tournament is in published status,
  - **Then** the system should display tournament rules, schedule, participating teams, and crowdfunding progress.
- **AC2 (Unpublished Access):**
  - **Given** the requested tournament is in draft or unpublished status,
  - **When** an unauthorized user attempts to view the page,
  - **Then** the system should display an appropriate access restricted message.
- **AC3 (Non-existent Tournament):**
  - **Given** the requested tournament does not exist,
  - **When** the user attempts to access the URL,
  - **Then** the system should display an appropriate "Tournament not found" error message.

```gherkin
Scenario Outline: Accessing tournament details page by tournament status
  Given the user navigates to the tournament details page for a tournament in state "<state>"
  When the page finishes loading
  Then the system should <systemAction>
  And the page should display "<visibleContent>"

  Examples:
    | state             | systemAction                                             | visibleContent                                              |
    | PUBLISHED         | render full tournament overview and registration options | Schedule, rules, approved teams, and funding progress bar   |
    | DRAFT_UNPUBLISHED | block access with a restricted notice                    | Access denied notice: "This tournament is not published"    |
    | NON_EXISTENT      | display not found page                                   | Error message: "Tournament not found"                       |
```

---

### US2-3: View Organizer Credibility and Public Sponsor List
> **User Story:**
> As a user,
> I want to see an organizer's credibility and public sponsor list
> So that I can judge whether a tournament is trustworthy.

#### Acceptance Criteria
- **AC1 (Credibility Metrics):**
  - **Given** the user is on the organizer credibility profile page,
  - **When** the user views the profile,
  - **Then** the system should display the organizer verification status badge and completed tournament count.
- **AC2 (Sponsor Privacy Consent):**
  - **Given** a linked sponsor has granted or withheld public listing consent,
  - **When** the user views the public sponsor section,
  - **Then** the system should display brand logos only for sponsors who granted public consent.

```gherkin
Scenario Outline: Viewing organizer credibility metrics and sponsor privacy consent
  Given the user views an organizer profile with verification "<verificationStatus>" and "<completedCount>" completed tournaments
  And the organizer has a linked sponsor with public listing consent "<consentGranted>"
  When the credibility page is displayed
  Then the verification badge should read "<badgeText>"
  And the completed tournament counter should display "<completedCount>"
  And the sponsor logo section should be "<sponsorDisplay>"

  Examples:
    | verificationStatus | completedCount | consentGranted | badgeText     | sponsorDisplay |
    | VERIFIED           | 15             | TRUE           | Verified Host | Visible        |
    | UNVERIFIED         | 1              | TRUE           | Unverified    | Visible        |
    | VERIFIED           | 8              | FALSE          | Verified Host | Hidden         |
```

---

## EPIC 3: Tournament Hosting and Lifecycle

### US3-1: Create Tournament
> **User Story:**
> As an organizer,
> I want to create a tournament
> So that I can define the event, funding needs, rules, solo/team size restrictions, tournament format, and rewards.

#### Acceptance Criteria
- **AC1 (Tournament Creation):**
  - **Given** an authenticated organizer is on the tournament creation page,
  - **When** the organizer attempts to create a tournament with valid dates, team size restrictions, and fee,
  - **Then** the system should allow the action and save the tournament in draft status.
- **AC2 (Validation):**
  - **Given** the user input is incomplete or invalid (e.g. end date before start date, min team size exceeds max, negative fee),
  - **When** the organizer attempts to submit the form,
  - **Then** the system should display an appropriate error message and prevent creation.
- **AC3 (Confirmation):**
  - **Given** the system has successfully created the tournament,
  - **When** creation completes,
  - **Then** the system should confirm success with a message and redirect to the tournament dashboard.

```gherkin
Scenario Outline: Tournament configuration boundaries and validation
  Given an authenticated organizer is on the tournament creation page
  When the organizer attempts to create a tournament with dates from "<startDate>" to "<endDate>", team size "<minTeam>" to "<maxTeam>", and fee "<fee>"
  Then the system should <systemAction>
  And the user should see message "<feedbackMessage>"

  Examples:
    | startDate  | endDate    | minTeam | maxTeam | fee  | systemAction                                 | feedbackMessage                      |
    | 2026-10-01 | 2026-10-05 | 5       | 5       | 500  | create tournament draft and redirect         | Tournament created successfully      |
    | 2026-10-10 | 2026-10-05 | 5       | 5       | 500  | reject submission and highlight date field   | End date cannot be before start date |
    | 2026-10-01 | 2026-10-05 | 6       | 4       | 500  | reject submission and highlight team sizes   | Min team size cannot exceed max size |
    | 2026-10-01 | 2026-10-05 | 0       | 0       | 500  | reject submission and highlight team sizes   | Team size must be at least 1 player  |
    | 2026-10-01 | 2026-10-05 | 1       | 1       | -100 | reject submission and highlight fee field    | Entry fee cannot be negative         |
```

---

### US3-2: Tournament Dashboard
> **User Story:**
> As an organizer,
> I want to see tournament's dashboard (e.g. status, participant, etc.)
> So that I can view overall status of the tournament.

#### Acceptance Criteria
- **AC1 (Dashboard View):**
  - **Given** an authorized organizer or assigned staff is on the tournament dashboard,
  - **When** the user accesses the page,
  - **Then** the system should display live participant counts, pending teams, and funding metrics.
- **AC2 (Access Control):**
  - **Given** an unauthorized user (player or anonymous guest) attempts to access the dashboard,
  - **When** the user opens the URL,
  - **Then** the system should deny access and display an appropriate error message or sign-in prompt.

```gherkin
Scenario Outline: Tournament dashboard access permissions
  Given a user with role "<userRole>" navigates to the tournament dashboard page
  When the page loads
  Then the system should <systemAction>
  And private participant and financial metrics should be "<metricVisibility>"

  Examples:
    | userRole          | systemAction                                               | metricVisibility |
    | TOURNAMENT_OWNER  | grant access to live participant counts and funding metrics| Visible          |
    | ASSIGNED_STAFF    | grant access to participant counts and pending team rosters| Visible          |
    | REGISTERED_PLAYER | deny access with error "Organizer access only"             | Hidden           |
    | ANONYMOUS_GUEST   | redirect to login page with sign-in prompt                 | Hidden           |
```

---

### US3-3: Manual Tournament Status Override
> **User Story:**
> As an organizer,
> I want the ability to manually override the tournament status
> So that I can quickly adjust for delays, cancellations, or unexpected schedule changes.

#### Acceptance Criteria
- **AC1 (Status Override):**
  - **Given** an authorized organizer is on the tournament settings page,
  - **When** the organizer selects a valid target status and provides an explanation reason,
  - **Then** the system should update the tournament status and notify registered participants.
- **AC2 (Mandatory Reason):**
  - **Given** the organizer attempts to cancel or override status without a reason,
  - **When** the submission is attempted,
  - **Then** the system should display an appropriate error message stating a reason is required.
- **AC3 (Locked State):**
  - **Given** a tournament is finalized or completed,
  - **When** the organizer attempts to reopen it,
  - **Then** the system should block the transition and display an appropriate error message.

```gherkin
Scenario Outline: Manual tournament status transitions and mandatory reason checks
  Given an organizer is viewing settings of a tournament in "<currentStatus>" status
  When the organizer overrides status to "<targetStatus>" with explanation reason "<reason>"
  Then the system should <systemAction>
  And the tournament status badge should display "<finalStatus>"

  Examples:
    | currentStatus       | targetStatus        | reason            | systemAction                                    | finalStatus         |
    | REGISTRATION_OPEN   | CANCELLED           | Severe weather    | cancel tournament and alert registered teams    | CANCELLED           |
    | REGISTRATION_CLOSED | ONGOING             | Matches starting  | start tournament and open match bracket         | ONGOING             |
    | ONGOING             | COMPLETED           | Finals concluded  | archive tournament standings and lock brackets  | COMPLETED           |
    | REGISTRATION_OPEN   | CANCELLED           |                   | block action with error "Reason is required"    | REGISTRATION_OPEN   |
    | COMPLETED           | REGISTRATION_OPEN   | Reopening test    | block action with error "Cannot reopen event"   | COMPLETED           |
```

---

### US3-4: Announce Tournament Details
> **User Story:**
> As an organizer,
> I want to announce further tournament details (e.g. venue change) to the related parties
> So that they can receive updates about the tournament.

#### Acceptance Criteria
- **AC1 (Publish Announcement):**
  - **Given** an authorized organizer is on the announcements page,
  - **When** the organizer submits an announcement with a title and message body,
  - **Then** the system should publish the announcement to the public feed and notify participants.
- **AC2 (Validation):**
  - **Given** the announcement title is missing, body is blank, or exceeds maximum length,
  - **When** the organizer attempts to publish,
  - **Then** the system should display an appropriate error message.

```gherkin
Scenario Outline: Publishing tournament announcements and message limits
  Given an organizer is on the "Tournament Announcements" tab
  When the organizer submits an announcement with title "<title>" and body length "<bodyLength>"
  Then the system should <systemAction>
  And the public tournament feed should display "<feedStatus>"

  Examples:
    | title           | bodyLength | systemAction                                           | feedStatus |
    | Venue Update    | 150 chars  | publish announcement and alert registered participants | Visible    |
    | Schedule Change | 50 chars   | publish announcement and alert registered participants | Visible    |
    |                 | 150 chars  | block posting with error "Title is required"           | Not Posted |
    | Urgent Notice   | 0 chars    | block posting with error "Message cannot be empty"     | Not Posted |
    | Rules Revision  | 6000 chars | block posting with error "Exceeds 2000 character limit"| Not Posted |
```

---

### US3-5: Configure Custom Registration Form
> **User Story:**
> As an organizer,
> I want to configure a custom registration form with custom field types (text, dropdown, file upload)
> So that I can collect tournament-specific information.

#### Acceptance Criteria
- **AC1 (Form Field Configuration):**
  - **Given** an organizer is on the registration form builder page,
  - **When** the organizer adds custom fields (text, dropdown with options, file upload with size bounds),
  - **Then** the system should allow the action and update the applicant form preview.
- **AC2 (Validation):**
  - **Given** a dropdown field has no options or file upload size exceeds platform limits,
  - **When** the organizer attempts to save the field,
  - **Then** the system should display an appropriate error message.

```gherkin
Scenario Outline: Custom registration form field configuration
  Given an organizer is on the "Registration Form Builder" page
  When the organizer adds a field of type "<fieldType>" with label "<label>" and options "<options>"
  Then the system should <systemAction>
  And the applicant form preview should show field "<label>" as "<formStatus>"

  Examples:
    | fieldType   | label        | options          | systemAction                                            | formStatus |
    | TEXT        | In-Game ID   |                  | add text input field to the form                        | Enabled    |
    | DROPDOWN    | T-Shirt Size | S, M, L, XL, XXL | add selectable dropdown field to the form               | Enabled    |
    | DROPDOWN    | Rank Tier    |                  | reject field with error "Dropdown must have options"    | Excluded   |
    | FILE_UPLOAD | Student ID   | max_size_mb: 5   | add file upload field with 5MB limit                    | Enabled    |
    | FILE_UPLOAD | Proof Photo  | max_size_mb: 100 | reject field with error "File size cannot exceed 10MB"  | Excluded   |
```

---

### US3-6: Review Participant Applications
> **User Story:**
> As an organizer,
> I want to review submitted participant applications to approve or reject entries
> So that I maintain full control over who competes in my tournament.

#### Acceptance Criteria
- **AC1 (Approve / Reject):**
  - **Given** an organizer is reviewing participant applications,
  - **When** the organizer approves or rejects an application entry with an optional reason,
  - **Then** the system should update the entry status and adjust the confirmed tournament roster count.
- **AC2 (Invalid Transitions):**
  - **Given** an entry is already rejected or disqualified,
  - **When** the organizer attempts an invalid state transition,
  - **Then** the system should display an appropriate error message and prevent the action.

```gherkin
Scenario Outline: Reviewing team and solo applicant entries
  Given an organizer is reviewing an application entry in "<initialStatus>" status
  When the organizer clicks action "<action>" with reason "<reason>"
  Then the system should <systemAction>
  And the entry status badge should update to "<finalStatus>"

  Examples:
    | initialStatus | action     | reason          | systemAction                                | finalStatus |
    | PENDING       | Approve    |                 | approve entry and reserve confirmed slot    | APPROVED    |
    | PENDING       | Reject     | Rank ineligible | reject entry and release reserved slot      | REJECTED    |
    | APPROVED      | Disqualify | Disqualified    | disqualify entry and reopen tournament slot | REJECTED    |
    | REJECTED      | Approve    | Re-evaluating   | block action and maintain rejected state    | REJECTED    |
```

---

### US3-7: Invite and Assign Staff Roles
> **User Story:**
> As an organizer,
> I want to be able to invite and assign roles (staff, referee) to other accounts
> So that I don't have to share my credentials to others to have them help manage the tournament.

#### Acceptance Criteria
- **AC1 (Staff Invitation):**
  - **Given** an authorized organizer is on the staff management page,
  - **When** the organizer sends an invitation to an email with role Referee or Staff,
  - **Then** the system should create the invitation and grant appropriate permissions upon acceptance.
- **AC2 (Link Validation):**
  - **Given** an invitation link is expired or revoked,
  - **When** the invitee attempts to accept the link,
  - **Then** the system should display an appropriate error message and deny staff access.

```gherkin
Scenario Outline: Assigning tournament management roles via invite link
  Given an invited user opens a staff invitation link for role "<assignedRole>" with status "<linkStatus>"
  When the user attempts to accept the invitation
  Then the system should <systemAction>
  And the user's workspace should provide "<permissions>"

  Examples:
    | assignedRole | linkStatus | systemAction                                           | permissions                 |
    | REFEREE      | VALID      | assign user as Referee and confirm role assignment     | Enter & verify match scores |
    | STAFF        | VALID      | assign user as Staff and confirm role assignment       | Manage check-in and rosters |
    | REFEREE      | EXPIRED    | display error "This invitation link has expired"       | No management access        |
    | REFEREE      | REVOKED    | display error "This invitation has been revoked"       | No management access        |
    | OWNER        | VALID      | display error "Owner role cannot be delegated"         | No management access        |
```

---

## EPIC 4: Team Formation and Registration

### US4-1: Team Lobby Creation & Invite Code
> **User Story:**
> As a team captain,
> I want to create a team lobby for a specific tournament and generate an invite link/code
> So that I can gather my teammates into a unified roster before registering.

#### Acceptance Criteria
- **AC1 (Lobby Creation):**
  - **Given** a competitor is on the tournament lobby page for an open team tournament,
  - **When** the competitor attempts to create a team lobby with a valid team name,
  - **Then** the system should create the lobby, assign the user as Captain, and generate a unique invite code.
- **AC2 (Validation & Restrictions):**
  - **Given** the tournament is solo-only, closed, or the user already captains a team in the event,
  - **When** the competitor attempts to create a lobby,
  - **Then** the system should display an appropriate error message and prevent lobby creation.

```gherkin
Scenario Outline: Team lobby creation rules by tournament mode and captain status
  Given a competitor is on the lobby creation page for tournament mode "<regMode>" and status "<tourneyStatus>"
  And the competitor currently captains "<existingTeamsCount>" teams in this tournament
  When the competitor attempts to create a team lobby named "<teamName>"
  Then the system should <systemAction>
  And the lobby invite code card should be "<inviteCodeDisplay>"

  Examples:
    | regMode | tourneyStatus      | existingTeamsCount | teamName      | systemAction                                    | inviteCodeDisplay |
    | TEAM    | REGISTRATION_OPEN  | 0                  | Phoenix Squad | create lobby in Forming status and issue code   | Visible           |
    | BOTH    | REGISTRATION_OPEN  | 0                  | Duo Champions | create lobby in Forming status and issue code   | Visible           |
    | SOLO    | REGISTRATION_OPEN  | 0                  | Solo Squad    | reject creation with error "Solo-only event"    | Hidden            |
    | TEAM    | REGISTRATION_OPEN  | 1                  | Second Squad  | reject creation with error "Already a captain"  | Hidden            |
    | TEAM    | REGISTRATION_CLOSED| 0                  | Late Squad    | reject creation with error "Registration closed"| Hidden            |
```

---

### US4-2: Join Team Lobby via Invite Link/Code
> **User Story:**
> As a team member,
> I want to join a captain's team lobby using an invite link and complete my individual registration questions
> So that my profile and consent details are attached to the team's tournament entry.

#### Acceptance Criteria
- **AC1 (Join Lobby):**
  - **Given** an applicant has a valid invite code and gives participant consent,
  - **When** the applicant attempts to join an open team lobby with available slots,
  - **Then** the system should add the applicant to the team roster and update the member count.
- **AC2 (Validation):**
  - **Given** the lobby is full, locked, the code is invalid, or consent is withheld,
  - **When** the applicant attempts to join,
  - **Then** the system should display an appropriate error message.

```gherkin
Scenario Outline: Joining a team lobby with invite code and consent
  Given a competitor is on the join lobby screen for a team with capacity "<capacity>" and members "<currentMembers>"
  And the lobby status is "<lobbyStatus>"
  When the competitor enters invite code "<inviteCode>" with consent checkbox "<consentGiven>"
  Then the system should <systemAction>
  And the team member count should display "<newCount>"

  Examples:
    | capacity | currentMembers | lobbyStatus | inviteCode | consentGiven | systemAction                                    | newCount |
    | 5        | 3              | FORMING     | VALID_CODE | CHECKED      | add competitor to roster and confirm join       | 4        |
    | 5        | 5              | FORMING     | VALID_CODE | CHECKED      | display error "Team roster is full"             | 5        |
    | 5        | 2              | LOCKED      | VALID_CODE | CHECKED      | display error "Roster is locked for submission" | 2        |
    | 5        | 2              | FORMING     | WRONG_CODE | CHECKED      | display error "Invalid team invite code"        | 2        |
    | 5        | 2              | FORMING     | VALID_CODE | UNCHECKED    | display error "Consent is required to register" | 2        |
```

---

### US4-3: Lock Roster and Pay Team Registration Fee
> **User Story:**
> As a team captain,
> I want to lock our completed roster and pay the full team registration fee in one transaction
> So that our complete team application is submitted to the organizer for official entry.

#### Acceptance Criteria
- **AC1 (Lock & Pay):**
  - **Given** a team captain has a complete team roster meeting tournament team size requirements,
  - **When** the captain locks the roster and submits payment,
  - **Then** the system should lock the roster, mark the fee paid, and submit the entry to the organizer.
- **AC2 (Incomplete Roster / Declined Payment):**
  - **Given** the roster has fewer players than required or payment is declined,
  - **When** the captain attempts to lock and pay,
  - **Then** the system should display an appropriate error message and keep the roster unlocked.

```gherkin
Scenario Outline: Locking roster and processing team registration checkout
  Given a team captain in a lobby with "<memberCount>" members where the tournament requires "<requiredCount>"
  When the captain locks the roster and submits payment with card result "<cardOutcome>"
  Then the system should <systemAction>
  And the lobby status badge should display "<lobbyBadge>"

  Examples:
    | memberCount | requiredCount | cardOutcome | systemAction                                           | lobbyBadge |
    | 5           | 5             | APPROVED    | lock roster, record payment, and submit entry          | Submitted  |
    | 3           | 5             | APPROVED    | block submission with error "Roster requires 5 players"| Forming    |
    | 5           | 5             | DECLINED    | display error "Payment declined. Please retry"         | Forming    |
```

---

### US4-4: Solo Competitor Registration and Payment
> **User Story:**
> As a solo competitor,
> I want to fill out required registration questions and pay the individual entry fee directly
> So that I can register for individual tournaments without creating a team lobby.

#### Acceptance Criteria
- **AC1 (Solo Registration):**
  - **Given** a solo competitor is registering for a solo tournament,
  - **When** the competitor fills out required answers and submits payment (or joins a free tournament),
  - **Then** the system should confirm the individual registration and display a receipt.
- **AC2 (Validation):**
  - **Given** the payment is declined or the tournament requires team lobbies,
  - **When** the competitor attempts to register,
  - **Then** the system should display an appropriate error message and prevent entry.

```gherkin
Scenario Outline: Solo tournament competitor registration and payment
  Given a solo competitor is on the checkout page for tournament mode "<regMode>" and fee "<entryFee>"
  When the competitor submits answers with payment result "<paymentOutcome>"
  Then the system should <systemAction>
  And the user's registration status should display "<registrationBadge>"

  Examples:
    | regMode | entryFee | paymentOutcome | systemAction                                              | registrationBadge |
    | SOLO    | 300 THB  | APPROVED       | confirm solo registration and generate receipt            | Confirmed         |
    | BOTH    | 0 THB    | FREE_ENTRY     | confirm free solo registration immediately                | Confirmed         |
    | SOLO    | 300 THB  | DECLINED       | display error "Payment failed. Card was not charged"      | Unregistered      |
    | TEAM    | 300 THB  | APPROVED       | block registration with error "Team lobby is required"    | Unregistered      |
```

---

### US4-5: Kick Member from Lobby
> **User Story:**
> As a team captain,
> I want to remove (kick) a member from my team lobby before the roster is locked and paid
> So that I can replace inactive players or manage roster changes.

#### Acceptance Criteria
- **AC1 (Kick Teammate):**
  - **Given** an unlocked team lobby,
  - **When** the team captain removes a member from the roster and confirms,
  - **Then** the system should remove the member, free the slot, and notify the removed player.
- **AC2 (Permissions):**
  - **Given** a non-captain attempts to remove a player or captain attempts to kick themselves,
  - **When** the action is attempted,
  - **Then** the system should display an appropriate error message and preserve the roster.

```gherkin
Scenario Outline: Removing members from an unlocked team lobby
  Given a team lobby viewed by user with role "<requesterRole>"
  When the user attempts to remove a member with role "<targetRole>"
  Then the system should <systemAction>
  And the roster slot status should be "<slotStatus>"

  Examples:
    | requesterRole | targetRole | systemAction                                                 | slotStatus |
    | CAPTAIN       | MEMBER     | remove member from lobby and confirm removal                 | Empty Slot |
    | CAPTAIN       | CAPTAIN    | block action with error "Captains cannot remove themselves"  | Preserved  |
    | MEMBER        | MEMBER     | block action with error "Only team captains can kick members"| Preserved  |
```

---

### US4-6: Regenerate Lobby Invite Code
> **User Story:**
> As a team captain,
> I want to regenerate the lobby invite link/code and automatically revoke the previous one
> So that unauthorized players cannot join if the original link is leaked.

#### Acceptance Criteria
- **AC1 (Regenerate Code):**
  - **Given** an unlocked team lobby,
  - **When** the captain requests a new invite code,
  - **Then** the system should generate a new unique code and immediately revoke the previous code.
- **AC2 (Revoked Code Validation):**
  - **Given** a previous code has been revoked,
  - **When** any player attempts to join using the old code,
  - **Then** the system should display an appropriate error message stating the code has expired.

```gherkin
Scenario Outline: Regenerating team lobby invite code and revoking prior code
  Given a team captain in a lobby with lock status "<lockStatus>" and active code "INV-1111"
  When the captain clicks the "Regenerate Invite Code" button
  Then the system should <systemAction>
  And any player attempting to use "INV-1111" sees "<oldCodeResult>"

  Examples:
    | lockStatus | systemAction                                            | oldCodeResult                     |
    | UNLOCKED   | issue a new unique code and revoke INV-1111             | This invitation code has expired  |
    | LOCKED     | block action with error "Cannot regenerate locked lobby"| Team roster is locked             |
```

---

### US4-7: Disband Team Lobby
> **User Story:**
> As a team captain,
> I want to disband my team lobby before registration and payment are finalized
> So that all members are released and the lobby is permanently removed.

#### Acceptance Criteria
- **AC1 (Disband Lobby):**
  - **Given** an unpaid and unlocked team lobby,
  - **When** the team captain confirms lobby disbandment,
  - **Then** the system should delete the lobby, revoke all codes, and release all members.
- **AC2 (Lock Protection):**
  - **Given** the team application is locked and paid,
  - **When** the captain attempts to disband the lobby,
  - **Then** the system should display an appropriate error message and prevent deletion.

```gherkin
Scenario Outline: Disbanding team lobby before registration finalization
  Given a user with role "<actorRole>" in a lobby with finalization status "<isFinalized>"
  When the user confirms lobby disbandment
  Then the system should <systemAction>
  And former teammates should <teammateStatus>

  Examples:
    | isFinalized | actorRole | systemAction                                                   | teammateStatus                   |
    | FALSE       | CAPTAIN   | delete lobby, revoke codes, and redirect to catalog            | be free to join another lobby    |
    | TRUE        | CAPTAIN   | block disbandment with error "Cannot disband paid application" | remain locked in the application |
    | FALSE       | MEMBER    | block action with error "Only captains can disband a lobby"    | remain in the lobby              |
```

---

## EPIC 5: Sponsorship Management

### US5-1: Crowdfunding-Only Stage Launch
> **User Story:**
> As an organizer,
> I want to launch a tournament in a crowdfunding-only stage with a target funding goal and deadline
> So that I can secure corporate sponsorship before opening registrations.

#### Acceptance Criteria
- **AC1 (Campaign Launch):**
  - **Given** an organizer is configuring a crowdfunding campaign,
  - **When** the organizer sets a positive funding goal and future deadline,
  - **Then** the system should publish the crowdfunding campaign page with a funding progress bar and pledge options.
- **AC2 (Validation):**
  - **Given** the funding goal is zero or deadline is in the past,
  - **When** the organizer attempts to launch the campaign,
  - **Then** the system should display an appropriate error message.

```gherkin
Scenario Outline: Launching tournament in crowdfunding-only stage
  Given an organizer is configuring a crowdfunding campaign
  When the organizer launches campaign with goal "<goalAmount>" and deadline "<deadline>"
  Then the system should <systemAction>
  And the tournament public banner should display "<publicBanner>"

  Examples:
    | goalAmount | deadline       | systemAction                                             | publicBanner         |
    | 50000 THB  | 30 Days Future | publish campaign page with funding meter and pledge card | Crowdfunding Active  |
    | 0 THB      | 30 Days Future | reject launch with error "Goal must be greater than 0"   | Draft (Unpublished)  |
    | 50000 THB  | 5 Days Past    | reject launch with error "Deadline must be in future"    | Draft (Unpublished)  |
```

---

### US5-2: Crowdfunded Event Finalization or Early Transition
> **User Story:**
> As an organizer of a crowdfunded event,
> I want to finalize event details and open participant registrations once the funding goal is met--or manually force-start/cancel the event.

#### Acceptance Criteria
- **AC1 (Goal Met / Force Start):**
  - **Given** a crowdfunded tournament has met its funding goal (or organizer force-starts with valid reason),
  - **When** the organizer transitions the event,
  - **Then** the system should finalize operational details and open participant registration.
- **AC2 (Cancellation & Refund):**
  - **Given** the funding goal is unmet and the campaign is cancelled with a reason,
  - **When** the organizer confirms cancellation,
  - **Then** the system should initiate sponsor refunds and mark the tournament cancelled.

```gherkin
Scenario Outline: Transitioning crowdfunded tournament to operational state
  Given an organizer on the dashboard of a crowdfunded event with funding goal ratio "<fundingRatio>"
  When the organizer triggers action "<action>" with explanation reason "<reason>"
  Then the system should <systemAction>
  And the tournament state badge should update to "<finalState>"

  Examples:
    | fundingRatio | action       | reason              | systemAction                                          | finalState         |
    | 100%         | OPEN_REG     | Goal met on time    | publish event schedule and open player registration   | REGISTRATION_OPEN  |
    | 60%          | FORCE_START  | Private funding add | publish event schedule and open player registration   | REGISTRATION_OPEN  |
    | 30%          | CANCEL       | Goal unmet          | initiate sponsor refunds and mark campaign cancelled  | CANCELLED          |
    | 30%          | CANCEL       |                     | block action with error "Reason is required to cancel"| CROWDFUNDING       |
```

---

### US5-3: Sponsor Directory and Public Contact Directory
> **User Story:**
> As an organizer,
> I want to browse a directory of verified sponsor profiles with public contact information
> So that I can find and reach out to relevant brands.

#### Acceptance Criteria
- **AC1 (Verified Directory):**
  - **Given** an organizer is on the sponsor directory page,
  - **When** the organizer searches for brands,
  - **Then** the system should display verified sponsor profiles who have granted public directory listing consent.
- **AC2 (Privacy & Verification):**
  - **Given** a sponsor profile is unverified or has withheld directory listing consent,
  - **When** search is performed,
  - **Then** the system should exclude the sponsor from public directory results.

```gherkin
Scenario Outline: Sponsor directory visibility filters
  Given an organizer is browsing the "Sponsor Directory" page
  When the organizer searches for a brand with verification "<isVerified>" and directory consent "<hasConsent>"
  Then the sponsor brand card should be "<cardAppearance>" in search results
  And public contact information should be "<contactVisibility>"

  Examples:
    | isVerified | hasConsent | cardAppearance | contactVisibility |
    | VERIFIED   | TRUE       | Displayed      | Visible           |
    | UNVERIFIED | TRUE       | Excluded       | Hidden            |
    | VERIFIED   | FALSE      | Excluded       | Hidden            |
```

---

### US5-4: Structured Sponsorship Tier Packages
> **User Story:**
> As an organizer,
> I want to create structured sponsorship tier packages (e.g., Gold, Silver) with defined perk descriptions and a funding goal.

#### Acceptance Criteria
- **AC1 (Create Tier):**
  - **Given** an organizer is on the sponsorship tier setup page,
  - **When** the organizer defines tier name, positive pledge amount, and available slots count,
  - **Then** the system should save the tier and display it on the campaign pledge card.
- **AC2 (Validation):**
  - **Given** pledge amount is zero or slots count is invalid,
  - **When** the organizer attempts to save,
  - **Then** the system should display an appropriate error message.

```gherkin
Scenario Outline: Configuring sponsorship tier packages
  Given an organizer is on the "Sponsorship Tier Packages" tab
  When the organizer creates tier with name "<name>", pledge amount "<price>", and slots "<slots>"
  Then the system should <systemAction>
  And the campaign pledge card should be "<tierStatus>"

  Examples:
    | name   | price     | slots | systemAction                                           | tierStatus |
    | Gold   | 10000 THB | 2     | save tier and display self-service pledge card         | Displayed  |
    | Silver | 5000 THB  | 5     | save tier and display self-service pledge card         | Displayed  |
    | Bronze | 0 THB     | 10    | reject tier with error "Amount must be greater than 0" | Hidden     |
    | Custom | 15000 THB | 0     | reject tier with error "Slots must be at least 1"      | Hidden     |
```

---

### US5-5: Sponsor Checkout and Logo Showcase
> **User Story:**
> As a sponsor,
> I want to select a sponsorship package, upload my brand logo/link, and complete payment
> So that my brand is officially recognized.

#### Acceptance Criteria
- **AC1 (Checkout & Logo Display):**
  - **Given** an available sponsorship tier with remaining slots,
  - **When** a sponsor uploads a valid image logo (PNG, JPG, SVG) and completes payment,
  - **Then** the system should confirm the pledge, decrease available slots, and display the logo on the tournament page.
- **AC2 (Validation):**
  - **Given** the tier is sold out, logo format is invalid, or payment is declined,
  - **When** checkout is attempted,
  - **Then** the system should display an appropriate error message.

```gherkin
Scenario Outline: Sponsor checkout, asset validation, and public recognition
  Given a sponsor on the checkout page for a tier with "<availableSlots>" available slots
  When the sponsor uploads logo format "<fileFormat>" and payment outcome is "<paymentOutcome>"
  Then the system should <systemAction>
  And the brand logo in the tournament sponsors section should be "<logoDisplay>"

  Examples:
    | availableSlots | fileFormat | paymentOutcome | systemAction                                          | logoDisplay |
    | 2              | PNG        | APPROVED       | record pledge, reserve slot, and confirm sponsorship  | Visible     |
    | 0              | PNG        | APPROVED       | block checkout with error "Tier is sold out"          | Hidden      |
    | 2              | EXE        | APPROVED       | reject file with error "Logo must be PNG, JPG, or SVG"| Hidden      |
    | 2              | PNG        | DECLINED       | display error "Payment declined. Please retry"        | Hidden      |
```

---

### US5-6: Automated Refund on Campaign Expiry
> **User Story:**
> As a sponsor,
> I want the system to automatically refund my pledge if a crowdfunding campaign expires without meeting its goal.

#### Acceptance Criteria
- **AC1 (Automatic Refund on Expiry):**
  - **Given** a crowdfunding campaign has expired without meeting its funding goal,
  - **When** the campaign deadline scheduler executes,
  - **Then** the system should automatically initiate refunds to all pledged sponsors and notify them.
- **AC2 (Goal Met):**
  - **Given** the campaign has met or exceeded its goal,
  - **When** the deadline scheduler executes,
  - **Then** the system should lock funds for event delivery without issuing refunds.

```gherkin
Scenario Outline: Automated campaign expiry refund execution
  Given a crowdfunding campaign has reached its deadline with funding goal ratio "<goalRatio>"
  When the deadline scheduler executes
  Then the system should <systemAction>
  And the sponsor's pledge card should display "<refundNotice>"

  Examples:
    | goalRatio | systemAction                                                     | refundNotice                                        |
    | 70%       | initiate automatic refunds to all sponsors for this event        | Campaign expired. Automatic refund issued to card.  |
    | 100%      | lock funds for event disbursement and issue no refunds           | Campaign successful! Funds locked for event delivery|
    | 125%      | lock funds for event disbursement and issue no refunds           | Campaign successful! Funds locked for event delivery|
```

---

## EPIC 6: Match Scheduling and Scoring

### US6-1: Tournament Bracket & Progression Graph Generation
> **User Story:**
> As an organizer,
> I want to customize the tournament format as a graph
> So that the platform automatically generates the match bracket or scorecard lobby.

#### Acceptance Criteria
- **AC1 (Bracket Generation):**
  - **Given** an organizer has configured a tournament format graph with confirmed participants,
  - **When** registration closes and the organizer generates the bracket,
  - **Then** the system should automatically create match pairings according to the format.
- **AC2 (Validation):**
  - **Given** confirmed participant count is below minimum or format graph is invalid,
  - **When** the organizer attempts generation,
  - **Then** the system should display an appropriate error message.

```gherkin
Scenario Outline: Generating tournament match bracket upon registration close
  Given an organizer is on the bracket generator page for format "<formatType>" with "<participantCount>" approved teams
  When the organizer attempts to generate the bracket
  Then the system should <systemAction>
  And the bracket canvas status should be "<bracketStatus>"

  Examples:
    | formatType         | participantCount | systemAction                                            | bracketStatus |
    | SINGLE_ELIMINATION | 8                | generate 7 bracket matches pairing seeded participants  | Rendered      |
    | DOUBLE_ELIMINATION | 16               | generate upper and lower bracket match trees            | Rendered      |
    | SINGLE_ELIMINATION | 3                | reject generation with error "Minimum 4 teams required" | Empty Canvas  |
    | INVALID_CYCLIC     | 8                | reject generation with error "Progression graph invalid"| Empty Canvas  |
```

---

### US6-2: Manual Reseeding and Matchup Swapping
> **User Story:**
> As an organizer or staff member,
> I want to manually reseed matchups or swap participants in the bracket before matches start.

#### Acceptance Criteria
- **AC1 (Reseed Matchup):**
  - **Given** scheduled matches in an unstarted round,
  - **When** an authorized user swaps participants with an explanation reason,
  - **Then** the system should update the match pairings and notify affected participants.
- **AC2 (Validation):**
  - **Given** a match is currently in progress or reason is missing,
  - **When** swapping is attempted,
  - **Then** the system should display an appropriate error message and prevent changes.

```gherkin
Scenario Outline: Swapping participants in tournament bracket
  Given an organizer is viewing a bracket match card in status "<matchState>"
  When the organizer swaps participants with explanation reason "<reason>"
  Then the system should <systemAction>
  And the match card slot should display "<newMatchup>"

  Examples:
    | matchState  | reason            | systemAction                                            | newMatchup   |
    | SCHEDULED   | Schedule conflict | swap participants in matchup and record audit log       | Swapped Team |
    | IN_PROGRESS | Late swap request | block swap with error "Cannot swap during active match" | Unchanged    |
    | SCHEDULED   |                   | block swap with error "Reason is required to reseed"    | Unchanged    |
```

---

### US6-3: Match Times and Venue Assignment
> **User Story:**
> As an organizer or staff member,
> I want to assign times and venue locations to scheduled matches
> So that competitors know exactly where and when to compete.

#### Acceptance Criteria
- **AC1 (Schedule Assignment):**
  - **Given** scheduled matches in a tournament,
  - **When** an authorized user assigns start times and venue locations without conflicts,
  - **Then** the system should publish the schedule to the public calendar and alert competitors.
- **AC2 (Conflict / Past Date):**
  - **Given** a selected venue has a time conflict or the time is in the past,
  - **When** scheduling is attempted,
  - **Then** the system should display an appropriate error message.

```gherkin
Scenario Outline: Match scheduling and resource conflict detection
  Given an organizer is scheduling a tournament match
  When the organizer assigns match time "<timeOffset>" and venue conflict status is "<hasConflict>"
  Then the system should <systemAction>
  And the match card schedule tag should display "<calendarDisplay>"

  Examples:
    | timeOffset | hasConflict | systemAction                                            | calendarDisplay  |
    | +2 Days    | FALSE       | assign time and venue and notify teams and officials    | In 2 Days        |
    | -1 Days    | FALSE       | reject schedule with error "Time cannot be in the past" | Unscheduled      |
    | +2 Days    | TRUE        | reject schedule with error "Venue has a time conflict"  | Unscheduled      |
```

---

### US6-4: Match Score Recording and Leaderboard Advancement
> **User Story:**
> As an organizer, staff member, or a referee,
> I want to record match scores or FFA placement rankings
> So that the winning teams advance automatically.

#### Acceptance Criteria
- **AC1 (Record Score & Advance):**
  - **Given** a scheduled match in progress,
  - **When** an authorized referee or organizer enters the final match score,
  - **Then** the system should record the official result and automatically advance the winner in the bracket.
- **AC2 (Validation & Authorization):**
  - **Given** the score results in an invalid tie for an elimination match or entered by an unauthorized user,
  - **When** submitted,
  - **Then** the system should display an appropriate error message and prevent advancement.

```gherkin
Scenario Outline: Recording match scores and automatic winner advancement
  Given a scheduled match between Team A and Team B
  When a user with role "<reporterRole>" submits final score Team A "<scoreA>" to Team B "<scoreB>"
  Then the system should <systemAction>
  And the winner slot in the next bracket round should display "<advancingTeam>"

  Examples:
    | reporterRole | scoreA | scoreB | systemAction                                              | advancingTeam |
    | REFEREE      | 2      | 1      | record official result and advance winner to next round   | Team A        |
    | ORGANIZER    | 0      | 2      | record official result and advance winner to next round   | Team B        |
    | REFEREE      | 1      | 1      | reject score with error "Elimination match cannot tie"    | None          |
    | COMPETITOR   | 2      | 0      | block submission with error "Only officials record score" | None          |
```

---

### US6-5: Public Live Bracket Tree and Standings
> **User Story:**
> As a competitor or spectator,
> I want to view the live bracket tree or standings table
> So that I can track event progression in real time.

#### Acceptance Criteria
- **AC1 (Live View):**
  - **Given** a published tournament with active matches,
  - **When** spectators or competitors view the live bracket page,
  - **Then** the system should display the interactive bracket tree, real-time scores, and standings.
- **AC2 (Unpublished Access):**
  - **Given** the tournament is in draft or unpublished status,
  - **When** access is attempted,
  - **Then** the system should display an appropriate access restricted message.

```gherkin
Scenario Outline: Public live bracket view by tournament publication status
  Given a spectator navigates to the public bracket URL for a tournament in status "<tournamentState>"
  When the page finishes loading
  Then the system should <systemAction>
  And the live bracket viewer should be "<bracketVisibility>"

  Examples:
    | tournamentState | systemAction                                                  | bracketVisibility |
    | PUBLISHED       | render live bracket tree with match scores and standings      | Visible           |
    | DRAFT           | deny access with message "Tournament is not published"        | Hidden            |
    | DELETED         | display error page "Tournament not found"                     | Hidden            |
```

---

### US6-6: Tournament Finalization and Standings Archival
> **User Story:**
> As an organizer,
> I want to finalize the tournament once all rounds are complete
> So that placement standings are permanently archived on the public tournament page.

#### Acceptance Criteria
- **AC1 (Finalize Event):**
  - **Given** all tournament matches are completed and all score disputes resolved,
  - **When** the organizer clicks the "Finalize Tournament" button,
  - **Then** the system should lock the bracket and permanently archive final placement standings on the public page.
- **AC2 (Open Disputes):**
  - **Given** there are unresolved match score disputes,
  - **When** finalization is attempted,
  - **Then** the system should display an appropriate error message and block finalization.

```gherkin
Scenario Outline: Finalizing tournament and locking standings
  Given all tournament matches are completed and unresolved dispute count is "<disputeCount>"
  When the organizer attempts to finalize the tournament
  Then the system should <systemAction>
  And the tournament state badge should update to "<finalState>"

  Examples:
    | disputeCount | systemAction                                               | finalState |
    | 0            | archive final standings on public page and lock bracket    | Finalized  |
    | 2            | block finalization with error "Resolve open score disputes"| Ongoing    |
```

---

## EPIC 7: Transaction Management

### US7-1: Secure External Payment Gateway Processing
> **User Story:**
> As a competitor or sponsor,
> I want to pay registration fees or sponsorship tier pledges through a secure external payment gateway
> So that my payment is securely processed.

#### Acceptance Criteria
- **AC1 (Secure Payment Settlement):**
  - **Given** a user is on the checkout page for an entry fee or sponsorship pledge,
  - **When** the user completes payment through the external payment gateway,
  - **Then** the system should confirm the payment, generate a receipt, and update registration/sponsorship status without storing raw card numbers.
- **AC2 (Declined Payment):**
  - **Given** the payment gateway declines the transaction,
  - **When** the callback is received,
  - **Then** the system should display an appropriate error message and prompt the user to retry payment.

```gherkin
Scenario Outline: External payment checkout settlement
  Given a user is at checkout for item "<itemType>" with fee "<amount>"
  When the external payment gateway returns status "<callbackStatus>"
  Then the system should <systemAction>
  And raw credit card numbers stored in platform database should be "<cardStorage>"

  Examples:
    | itemType     | amount    | callbackStatus | systemAction                                              | cardStorage |
    | REGISTRATION | 500 THB   | APPROVED       | mark registration fee as paid and generate receipt        | None        |
    | SPONSORSHIP  | 10000 THB | APPROVED       | confirm sponsorship tier purchase and activate perks      | None        |
    | REGISTRATION | 500 THB   | DECLINED       | display error "Card declined" and prompt user to retry    | None        |
```

---

### US7-2: Automatic Refund on Application Rejection or Event Cancellation
> **User Story:**
> As a competitor or sponsor,
> I want to receive an automatic refund to my original payment method if my team registration is rejected or the tournament is cancelled.

#### Acceptance Criteria
- **AC1 (Automatic Refund Trigger):**
  - **Given** a participant entry fee or sponsorship pledge was captured,
  - **When** the application is rejected or the tournament is cancelled,
  - **Then** the system should automatically issue a refund to the original payment method and record it in the ledger.
- **AC2 (Idempotent Refund):**
  - **Given** a transaction has already been refunded,
  - **When** a cancellation trigger occurs,
  - **Then** the system should perform no duplicate refund.

```gherkin
Scenario Outline: Triggering automated refunds for rejected or cancelled entries
  Given a captured payment with refund status "<initialRefundStatus>"
  When event trigger "<triggerEvent>" occurs
  Then the system should <systemAction>
  And the user billing history should record status "<refundRecordStatus>"

  Examples:
    | initialRefundStatus | triggerEvent         | systemAction                                         | refundRecordStatus |
    | NOT_REFUNDED        | APPLICATION_REJECTED | issue automatic refund to original payment method    | Refund Issued      |
    | NOT_REFUNDED        | TOURNAMENT_CANCELLED | issue automatic refund to original payment method    | Refund Issued      |
    | ALREADY_REFUNDED    | TOURNAMENT_CANCELLED | take no duplicate action and maintain existing state | Refund Issued      |
```

---

### US7-3: Centralized Transaction Ledger and Receipt Summaries
> **User Story:**
> As an account owner,
> I want to view a centralized ledger of all my incoming and outgoing transactions with downloadable receipt summaries.

#### Acceptance Criteria
- **AC1 (Ledger View & Filter):**
  - **Given** an authenticated user is on the transaction ledger page,
  - **When** the user filters by transaction type (all, payment, refund) and date range,
  - **Then** the system should display matching transaction records with downloadable PDF receipts.
- **AC2 (Date Validation):**
  - **Given** the start date is after the end date,
  - **When** the user attempts to filter,
  - **Then** the system should display an appropriate error message.

```gherkin
Scenario Outline: Viewing and filtering personal transaction ledger
  Given an account owner is on the "Billing & Transactions" page
  When the owner filters by type "<filterType>" and date range from "<fromDate>" to "<toDate>"
  Then the system should <systemAction>
  And the download receipt button should be "<downloadButtonState>"

  Examples:
    | filterType | fromDate   | toDate     | systemAction                                            | downloadButtonState |
    | ALL        | 2026-01-01 | 2026-06-01 | display matching payment, refund, and payout records    | Enabled             |
    | REFUND     | 2026-01-01 | 2026-06-01 | display only refund transaction records                 | Enabled             |
    | PAYMENT    | 2026-06-01 | 2026-01-01 | display error "Start date must precede End date"        | Disabled            |
```

---

### US7-4: Tournament Financial Breakdown
> **User Story:**
> As an organizer,
> I want to view a financial breakdown of my tournament
> So that I have an accurate accounting summary before funds are released.

#### Acceptance Criteria
- **AC1 (Financial Summary View):**
  - **Given** an authorized organizer or finance administrator,
  - **When** the user views the tournament financial breakdown page,
  - **Then** the system should display gross revenue, platform commission cuts, and net payout balances.
- **AC2 (Access Control):**
  - **Given** an unauthorized competitor or spectator attempts to access the page,
  - **When** the user opens the URL,
  - **Then** the system should deny access and display an appropriate error message.

```gherkin
Scenario Outline: Tournament financial accounting breakdown access
  Given a user with role "<viewerRole>" accesses the tournament financial breakdown page
  When the page loads
  Then the system should <systemAction>
  And gross revenue, platform cut, and net payout cards should be "<financialCardsVisibility>"

  Examples:
    | viewerRole       | systemAction                                                    | financialCardsVisibility |
    | TOURNAMENT_OWNER | display reconciled gross revenue, platform cuts, and net payout | Visible                  |
    | FINANCE_ADMIN    | display reconciled gross revenue, platform cuts, and net payout | Visible                  |
    | COMPETITOR       | deny access with error "You do not have permission to view"     | Hidden                   |
```

---

### US7-5: Payout Details Registration and Earnings Release
> **User Story:**
> As an organizer or winning competitor,
> I want to register my external payout details to receive net event earnings or prize money once the tournament is over.

#### Acceptance Criteria
- **AC1 (Payout Destination Registration):**
  - **Given** an eligible user with tournament earnings is on the payout settings page,
  - **When** the user submits valid payout credentials (PromptPay or Stripe Connect),
  - **Then** the system should verify and save the payout destination.
- **AC2 (Validation):**
  - **Given** payout credentials are blank or fail provider verification,
  - **When** the user attempts to save,
  - **Then** the system should display an appropriate error message.

```gherkin
Scenario Outline: Registering external payout details for earnings release
  Given an eligible user with tournament earnings is on the "Payout Settings" page
  When the user submits payout provider "<payoutProvider>" and account identifier "<accountId>"
  Then the system should <systemAction>
  And the payout destination status card should display "<releaseEligibility>"

  Examples:
    | payoutProvider | accountId    | systemAction                                                    | releaseEligibility |
    | PROMPTPAY      | 0812345678   | validate format and save verified payout destination            | Ready for Payout   |
    | STRIPE_CONNECT | acct_valid12 | verify Stripe onboarding and link payout destination            | Ready for Payout   |
    | PROMPTPAY      |              | reject with error "Account identifier cannot be blank"          | Ineligible         |
    | STRIPE_CONNECT | acct_invalid | reject with error "Stripe account onboarding is incomplete"     | Ineligible         |
