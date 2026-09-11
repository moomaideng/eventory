# Eventory Platform — Consolidated Acceptance Criteria

This document serves as the master business specification for the Eventory platform across **EPIC 1 through EPIC 7**. 

All criteria are written in business-readable **Behavior-Driven Development (BDD)** format using **`Scenario Outline`** blocks with **`Examples:`** tables. Duplicate scenarios that vary primarily by input data, user roles, or validation boundaries are consolidated into parameterized rows with measurable, observable outcomes.

---

## EPIC 1: Account and Role Access

### US1-1: Create Single Primary Account
> **User Story:** As a new user, I want to create a single primary account so that I have a central identity with a public handle to interact across the platform.

```gherkin
Scenario Outline: Primary account registration and onboarding validation
  Given a new user has completed email signup
  When the user submits onboarding details with display name "<displayName>" and handle "<handle>"
  Then the system <systemAction>
  And the user sees the message "<userFeedback>"

  Examples:
    | displayName | handle          | systemAction                                              | userFeedback                                     |
    | Alice Gamer | alice_pro       | creates the account and directs the user to the dashboard | Welcome to Eventory!                             |
    | Bob The Pro | existing_handle | keeps the user on onboarding without creating an account  | This handle is already taken                     |
    | Dan Solo    | al              | keeps the user on onboarding without creating an account  | Handle must be between 3 and 32 characters       |
    | Eve Online  | eve!@#$         | keeps the user on onboarding without creating an account  | Handle can only contain letters, numbers, and _  |
    |             | sam_valid       | keeps the user on onboarding without creating an account  | Display name is required                         |
```

---

### US1-2: Create and Switch Between Role Profiles
> **User Story:** As a user, I want to create and switch between Organizer and Sponsor profiles linked to my account so that I can manage events or corporate funds without maintaining separate logins.

```gherkin
Scenario Outline: Creating and switching active role profiles
  Given an authenticated user with established profiles: "<existingProfiles>"
  And the current active role is "<activeRole>"
  When the user requests to "<action>" profile for role "<targetRole>" with organization name "<orgName>"
  Then the system <systemAction>
  And the active workspace shows "<visibleNavigation>"

  Examples:
    | existingProfiles               | activeRole | action | targetRole | orgName         | systemAction                                           | visibleNavigation                 |
    | COMPETITOR                     | COMPETITOR | CREATE | ORGANIZER  | Siam Esports    | creates the organizer profile and links it to account  | Competitor Dashboard              |
    | COMPETITOR, ORGANIZER          | COMPETITOR | SWITCH | ORGANIZER  |                 | switches the active workspace without requiring re-auth| Tournament Hosting & Management   |
    | COMPETITOR, ORGANIZER          | ORGANIZER  | CREATE | SPONSOR    | Apex Brands Co. | creates the sponsor profile and links it to account    | Tournament Hosting & Management   |
    | COMPETITOR, ORGANIZER, SPONSOR | ORGANIZER  | SWITCH | SPONSOR    |                 | switches the active workspace without requiring re-auth| Sponsorship Tiers & Brand Portal  |
    | COMPETITOR                     | COMPETITOR | SWITCH | ORGANIZER  |                 | blocks the switch with message "Profile not found"     | Competitor Dashboard              |
    | COMPETITOR, SPONSOR            | SPONSOR    | SWITCH | ORGANIZER  |                 | blocks the switch with message "Profile not found"     | Sponsorship Tiers & Brand Portal  |
```

---

### US1-3: View and Update Role-Specific Profiles
> **User Story:** As an account owner, I want to view and update my role-specific profile(s) so that my contact and public information stays accurate.

```gherkin
Scenario Outline: Updating role profile contact details and authorization
  Given an authenticated user "<currentUser>"
  And an existing "<roleType>" profile owned by "<ownerUser>"
  When the user submits an update with organization name "<orgName>" and contact email "<email>"
  Then the system <systemAction>
  And the user sees the notice "<feedbackNotice>"

  Examples:
    | currentUser   | roleType  | ownerUser     | orgName         | email               | systemAction                                        | feedbackNotice                     |
    | organizer_bob | ORGANIZER | organizer_bob | Bob Gaming Int. | contact@bobgame.com | saves the new contact details to the profile        | Profile updated successfully       |
    | sponsor_acme  | SPONSOR   | sponsor_acme  | Acme Worldwide  | corp@acme.com       | saves the new contact details to the profile        | Profile updated successfully       |
    | organizer_bob | ORGANIZER | organizer_bob |                 | valid@bobgame.com   | rejects the update and preserves existing data      | Organization name cannot be blank  |
    | organizer_bob | ORGANIZER | organizer_bob | Bob Gaming Int. | not-an-email        | rejects the update and preserves existing data      | Please enter a valid email address |
    | attacker_dan  | ORGANIZER | organizer_bob | Hacked Brand    | evil@attacker.com   | denies the request and records no change            | You do not have permission to edit |
```

---

## EPIC 2: Tournament Discovery

### US2-1: Browse and Filter Tournaments
> **User Story:** As a user, I want to browse and filter tournaments so that I can quickly find events that fit my schedule and budget.

```gherkin
Scenario Outline: Filtering tournaments by keyword, dates, and maximum budget
  Given published tournaments exist across various games, dates, and fees
  When the user enters keyword "<keyword>", From date "<fromDate>", To date "<toDate>", and Maximum fee "<maxFee>"
  And clicks "Apply filters"
  Then the system <systemAction>
  And the catalog displays "<catalogDisplay>"

  Examples:
    | keyword  | fromDate   | toDate     | maxFee | systemAction                                                     | catalogDisplay                     |
    | Valorant |            |            |        | updates the URL and queries matching tournaments                 | Valorant tournaments only          |
    |          | 2026-06-01 | 2026-07-01 |        | updates the URL and queries tournaments starting in June 2026    | June 2026 tournaments              |
    |          |            |            | 500    | updates the URL and queries tournaments with entry fee under 500 | Free and entry fee <= 500 events   |
    |          | 2026-10-01 | 2026-09-01 |        | disables the submit button and shows "To date must be after From"| Current results without re-querying|
    | Unknown  |            |            |        | updates the URL and queries matching tournaments                 | "No tournaments match" empty state |
```

---

### US2-2: View Tournament Details, Teams, and Funding Progress
> **User Story:** As a user (potential competitor or sponsor), I want to view tournament details, teams, and funding progress so that I can decide whether to participate or provide support.

```gherkin
Scenario Outline: Accessing tournament details page by tournament status
  Given a tournament exists in state "<state>"
  When a user navigates to the tournament details page
  Then the system <systemAction>
  And the page displays "<visibleContent>"

  Examples:
    | state             | systemAction                                               | visibleContent                                                |
    | PUBLISHED         | renders the full tournament overview                       | Schedule, rules, approved teams, and crowdfunding progress bar|
    | DRAFT_UNPUBLISHED | blocks viewing with message "Tournament is not published"  | Access denied notice                                          |
    | NON_EXISTENT      | redirects to error page with message "Tournament not found"| Not Found message                                             |
```

---

### US2-3: View Organizer Credibility and Public Sponsor List
> **User Story:** As a user, I want to see an organizer's credibility and public sponsor list so that I can judge whether a tournament is trustworthy.

```gherkin
Scenario Outline: Viewing organizer credibility metrics and sponsor privacy consent
  Given an organizer profile with verification status "<verificationStatus>" and "<completedCount>" completed tournaments
  And the organizer has a linked sponsor with public listing consent set to "<consentGranted>"
  When a user views the organizer public credibility profile
  Then the profile displays a verification badge reading "<badgeText>"
  And the completed tournament counter displays "<completedCount>"
  And the sponsor logo display status is "<sponsorDisplay>"

  Examples:
    | verificationStatus | completedCount | consentGranted | badgeText     | sponsorDisplay |
    | VERIFIED           | 15             | TRUE           | Verified Host | Visible        |
    | UNVERIFIED         | 1              | TRUE           | Unverified    | Visible        |
    | VERIFIED           | 8              | FALSE          | Verified Host | Hidden         |
```

---

## EPIC 3: Tournament Hosting and Lifecycle

### US3-1: Create Tournament
> **User Story:** As an organizer, I want to create a tournament so that I can define the event, funding needs, rules, solo/team size restrictions, tournament format, and rewards.

```gherkin
Scenario Outline: Tournament configuration boundaries and validation
  Given an authenticated organizer on the tournament creation form
  When the organizer submits start date "<startDate>", end date "<endDate>", team size min "<minTeam>" to max "<maxTeam>", and fee "<fee>"
  Then the system <systemAction>
  And the organizer sees the message "<feedbackMessage>"

  Examples:
    | startDate  | endDate    | minTeam | maxTeam | fee  | systemAction                                             | feedbackMessage                       |
    | 2026-10-01 | 2026-10-05 | 5       | 5       | 500  | saves the tournament as a new draft                      | Tournament created successfully       |
    | 2026-10-10 | 2026-10-05 | 5       | 5       | 500  | rejects the form and highlights the end date field       | End date cannot be before start date  |
    | 2026-10-01 | 2026-10-05 | 6       | 4       | 500  | rejects the form and highlights team size fields         | Min team size cannot exceed max size  |
    | 2026-10-01 | 2026-10-05 | 0       | 0       | 500  | rejects the form and highlights the team size field      | Team size must be at least 1 player   |
    | 2026-10-01 | 2026-10-05 | 1       | 1       | -100 | rejects the form and highlights the entry fee field      | Entry fee cannot be negative          |
```

---

### US3-2: Tournament Dashboard
> **User Story:** As an organizer, I want to see tournament's dashboard (e.g. status, participant, etc.) so that I can view overall status of the tournament.

```gherkin
Scenario Outline: Tournament dashboard access permissions
  Given an existing tournament owned by "Organizer Alpha"
  When user "<requesterRole>" opens the tournament dashboard URL
  Then the system <systemAction>
  And private participant and financial metrics are "<metricVisibility>"

  Examples:
    | requesterRole     | systemAction                                                  | metricVisibility |
    | TOURNAMENT_OWNER  | displays live participant counts, pending teams, and funding  | Visible          |
    | ASSIGNED_STAFF    | displays live participant counts and pending team approvals   | Visible          |
    | REGISTERED_PLAYER | redirects with error "Access restricted to organizers"        | Hidden           |
    | ANONYMOUS_GUEST   | redirects to login page                                       | Hidden           |
```

---

### US3-3: Manual Tournament Status Override
> **User Story:** As an organizer, I want the ability to manually override the tournament status, so that I can quickly adjust for delays, cancellations, or unexpected schedule changes.

```gherkin
Scenario Outline: Manual tournament status transitions and mandatory reason checks
  Given a tournament currently in status "<currentStatus>"
  When the organizer requests to transition status to "<targetStatus>" with reason "<reason>"
  Then the system <systemAction>
  And the resulting tournament status is "<finalStatus>"

  Examples:
    | currentStatus       | targetStatus        | reason             | systemAction                                                    | finalStatus         |
    | REGISTRATION_OPEN   | CANCELLED           | Severe weather     | updates status, logs the reason, and alerts registered teams    | CANCELLED           |
    | REGISTRATION_CLOSED | ONGOING             | Matches starting   | updates status and opens match bracket for live scores          | ONGOING             |
    | ONGOING             | COMPLETED           | Finals concluded   | archives standings and locks match score edits                  | COMPLETED           |
    | REGISTRATION_OPEN   | CANCELLED           |                    | blocks the transition with error "Reason is required to cancel" | REGISTRATION_OPEN   |
    | COMPLETED           | REGISTRATION_OPEN   | Reopening test     | blocks the transition with error "Cannot reopen finalized event"| COMPLETED           |
```

---

### US3-4: Announce Tournament Details
> **User Story:** As an organizer, I want to announce further tournament details (e.g. venue change) to the related parties so that they can receive updates about the tournament.

```gherkin
Scenario Outline: Publishing tournament announcements and message limits
  Given an authorized tournament organizer
  When the organizer posts an announcement with title "<title>" and message body length "<bodyLength>"
  Then the system <systemAction>
  And the announcement delivery status is "<deliveryStatus>"

  Examples:
    | title           | bodyLength | systemAction                                                     | deliveryStatus |
    | Venue Update    | 150 chars  | publishes post to tournament feed and notifies registered players| Delivered      |
    | Schedule Change | 50 chars   | publishes post to tournament feed and notifies registered players| Delivered      |
    |                 | 150 chars  | blocks posting with error "Announcement title is required"       | Not Sent       |
    | Urgent Notice   | 0 chars    | blocks posting with error "Message body cannot be empty"         | Not Sent       |
    | Rules Revision  | 6000 chars | blocks posting with error "Message exceeds limit of 2000 chars"  | Not Sent       |
```

---

### US3-5: Configure Custom Registration Form
> **User Story:** As an organizer, I want to configure a custom registration form with custom field types (text, dropdown, file upload) so that I can collect tournament-specific information.

```gherkin
Scenario Outline: Custom registration form field configuration
  Given an organizer customizing a tournament registration form
  When the organizer adds a field of type "<fieldType>" with label "<label>" and options "<options>"
  Then the system <systemAction>
  And the field availability in applicant form is "<formStatus>"

  Examples:
    | fieldType   | label        | options           | systemAction                                               | formStatus |
    | TEXT        | In-Game ID   |                   | adds text input field to registration form                 | Enabled    |
    | DROPDOWN    | T-Shirt Size | S, M, L, XL, XXL  | adds selectable options dropdown to registration form      | Enabled    |
    | DROPDOWN    | Rank Tier    |                   | rejects field with error "Dropdown must have options"      | Excluded   |
    | FILE_UPLOAD | Student ID   | max_size_mb: 5    | adds file upload field with 5MB restriction                | Enabled    |
    | FILE_UPLOAD | Proof Photo  | max_size_mb: 100  | rejects field with error "File size limit cannot exceed 10MB"| Excluded |
```

---

### US3-6: Review Participant Applications
> **User Story:** As an organizer, I want to review submitted participant applications to approve or reject entries so that I maintain full control over who competes in my tournament.

```gherkin
Scenario Outline: Reviewing team and solo applicant entries
  Given an applicant entry with status "<initialStatus>"
  When the organizer reviews the entry and selects action "<action>" with optional reason "<reason>"
  Then the entry status becomes "<finalStatus>"
  And the tournament participant roster slot count is "<rosterAction>"

  Examples:
    | initialStatus | action  | reason           | finalStatus | rosterAction        |
    | PENDING       | APPROVE |                  | APPROVED    | Slot reserved (+1)  |
    | PENDING       | REJECT  | Rank ineligible  | REJECTED    | Slot released (0)   |
    | APPROVED      | REJECT  | Disqualified     | REJECTED    | Slot reopened (-1)  |
    | REJECTED      | APPROVE | Re-evaluating    | REJECTED    | Action blocked      |
```

---

### US3-7: Invite and Assign Staff Roles
> **User Story:** As an organizer, I want to be able to invite and assign roles (staff, referee) to other accounts so that I don't have to share my credentials to others to have them help manage the tournament.

```gherkin
Scenario Outline: Assigning tournament management roles via invite link
  Given an organizer generates an invite link for role "<assignedRole>"
  When an invited user accepts the link under link status "<linkStatus>"
  Then the system <systemAction>
  And the invitee permissions granted are "<permissions>"

  Examples:
    | assignedRole | linkStatus | systemAction                                                   | permissions                |
    | REFEREE      | VALID      | assigns user as Referee and logs the assignment timestamp      | Enter & verify match scores|
    | STAFF        | VALID      | assigns user as Staff and logs the assignment timestamp        | Manage check-in & rosters  |
    | REFEREE      | EXPIRED    | blocks acceptance with error "This invitation link has expired"| No staff access            |
    | REFEREE      | REVOKED    | blocks acceptance with error "This invitation has been revoked"| No staff access            |
    | OWNER        | VALID      | blocks action with error "Owner role cannot be delegated"      | No staff access            |
```

---

## EPIC 4: Team Formation and Registration

### US4-1: Team Lobby Creation & Invite Code
> **User Story:** As a team captain, I want to create a team lobby for a specific tournament and generate an invite link/code so that I can gather my teammates into a unified roster before registering.

```gherkin
Scenario Outline: Team lobby creation rules by tournament mode and captain status
  Given a tournament with registration mode "<regMode>" and status "<tourneyStatus>"
  And a competitor who has "<existingTeamsCount>" team entries in this tournament
  When the competitor creates a team lobby named "<teamName>"
  Then the system <systemAction>
  And the invite code generation result is "<codeStatus>"

  Examples:
    | regMode | tourneyStatus     | existingTeamsCount | teamName        | systemAction                                              | codeStatus |
    | TEAM    | REGISTRATION_OPEN | 0                  | Phoenix Squad   | creates lobby in 'Forming' state and assigns user captain | Generated  |
    | BOTH    | REGISTRATION_OPEN | 0                  | Duo Champions   | creates lobby in 'Forming' state and assigns user captain | Generated  |
    | SOLO    | REGISTRATION_OPEN | 0                  | Solo Squad      | rejects creation with error "Tournament is for solo only" | None       |
    | TEAM    | REGISTRATION_OPEN | 1                  | Second Squad    | rejects creation with error "You already captain a team"  | None       |
    | TEAM    | REGISTRATION_CLOSED| 0                 | Late Squad      | rejects creation with error "Registration has closed"     | None       |
```

---

### US4-2: Join Team Lobby via Invite Link/Code
> **User Story:** As a team member, I want to join a captain's team lobby using an invite link and complete my individual registration questions so that my profile and consent details are attached to the team's tournament entry.

```gherkin
Scenario Outline: Joining a team lobby with invite code and consent
  Given a team lobby with maximum capacity "<capacity>" and current member count "<currentMembers>"
  And the lobby status is "<lobbyStatus>"
  When an applicant enters invite code "<inviteCode>" and participant consent "<consentGiven>"
  Then the system <systemAction>
  And the team member count becomes "<newCount>"

  Examples:
    | capacity | currentMembers | lobbyStatus | inviteCode | consentGiven | systemAction                                              | newCount |
    | 5        | 3              | FORMING     | VALID_CODE | ACCEPTED     | adds applicant to roster as Member                        | 4        |
    | 5        | 5              | FORMING     | VALID_CODE | ACCEPTED     | rejects join with error "Team roster is full"             | 5        |
    | 5        | 2              | LOCKED      | VALID_CODE | ACCEPTED     | rejects join with error "Roster is locked for submission" | 2        |
    | 5        | 2              | FORMING     | WRONG_CODE | ACCEPTED     | rejects join with error "Invalid team invite code"        | 2        |
    | 5        | 2              | FORMING     | VALID_CODE | DECLINED     | rejects join with error "Consent is required to register" | 2        |
```

---

### US4-3: Lock Roster and Pay Team Registration Fee
> **User Story:** As a team captain, I want to lock our completed roster and pay the full team registration fee in one transaction so that our complete team application is submitted to the organizer for official entry.

```gherkin
Scenario Outline: Locking roster and processing team registration checkout
  Given a team lobby with "<memberCount>" members where the tournament requires exactly "<requiredCount>"
  When the captain locks the roster and submits payment with card outcome "<cardOutcome>"
  Then the system <systemAction>
  And the lobby state is updated to "<lobbyState>"

  Examples:
    | memberCount | requiredCount | cardOutcome | systemAction                                                      | lobbyState |
    | 5           | 5             | APPROVED    | locks roster, marks fee paid, and submits application to organizer| Submitted  |
    | 3           | 5             | APPROVED    | blocks submission with error "Team needs 5 players to lock roster"| Forming    |
    | 5           | 5             | DECLINED    | keeps roster unlocked with error "Payment declined by bank"       | Forming    |
```

---

### US4-4: Solo Competitor Registration and Payment
> **User Story:** As a solo competitor, I want to fill out required registration questions and pay the individual entry fee directly so that I can register for individual tournaments without creating a team lobby.

```gherkin
Scenario Outline: Solo tournament competitor registration and payment
  Given a tournament with registration mode "<regMode>" and entry fee "<entryFee>"
  When an individual competitor submits required answers with payment outcome "<paymentOutcome>"
  Then the system <systemAction>

  Examples:
    | regMode | entryFee | paymentOutcome | systemAction                                                     |
    | SOLO    | 300 THB  | APPROVED       | submits individual entry and displays confirmation receipt       |
    | BOTH    | 0 THB    | FREE_ENTRY     | submits individual entry immediately with no payment required    |
    | SOLO    | 300 THB  | DECLINED       | shows payment failure error and does not create an entry         |
    | TEAM    | 300 THB  | APPROVED       | blocks registration with error "Team lobby required for this event"|
```

---

### US4-5: Kick Member from Lobby
> **User Story:** As a team captain, I want to remove (kick) a member from my team lobby before the roster is locked and paid, so that I can replace inactive players or manage roster changes.

```gherkin
Scenario Outline: Removing members from an unlocked team lobby
  Given an unlocked team lobby
  When user with role "<requesterRole>" attempts to remove member with role "<targetRole>"
  Then the system <systemAction>
  And the team member slot is "<slotStatus>"

  Examples:
    | requesterRole | targetRole | systemAction                                                    | slotStatus |
    | CAPTAIN       | MEMBER     | removes member from lobby and notifies them of removal          | Vacated    |
    | CAPTAIN       | CAPTAIN    | blocks action with error "Captains cannot remove themselves"    | Preserved  |
    | MEMBER        | MEMBER     | blocks action with error "Only the team captain can kick players| Preserved  |
```

---

### US4-6: Regenerate Lobby Invite Code
> **User Story:** As a team captain, I want to regenerate the lobby invite link/code and automatically revoke the previous one, so that unauthorized players cannot join if the original link is leaked.

```gherkin
Scenario Outline: Regenerating team lobby invite code and revoking prior code
  Given a team lobby with lock status "<lockStatus>" and active code "CODE_ALPHA"
  When the team captain requests a new invite code
  Then the system <systemAction>
  And any player attempting to use "CODE_ALPHA" sees "<oldCodeResult>"

  Examples:
    | lockStatus | systemAction                                                | oldCodeResult                     |
    | UNLOCKED   | issues a new unique 6-character code and revokes CODE_ALPHA | "This invite code has expired"    |
    | LOCKED     | blocks action with error "Cannot regenerate code for locked"| "Team roster is locked"           |
```

---

### US4-7: Disband Team Lobby
> **User Story:** As a team captain, I want to disband my team lobby before registration and payment are finalized, so that all members are released and the lobby is permanently removed.

```gherkin
Scenario Outline: Disbanding team lobby before registration finalization
  Given a team lobby with finalization status "<isFinalized>"
  When user with role "<actorRole>" confirms lobby disbandment
  Then the system <systemAction>
  And former teammates <teammateStatus>

  Examples:
    | isFinalized | actorRole | systemAction                                                     | teammateStatus                   |
    | FALSE       | CAPTAIN   | deletes the lobby and revokes all invite codes                   | are free to join another lobby   |
    | TRUE        | CAPTAIN   | blocks disbandment with error "Cannot disband paid/locked entry" | remain locked in the application |
    | FALSE       | MEMBER    | blocks action with error "Only captains can disband a lobby"     | remain in the lobby              |
```

---

## EPIC 5: Sponsorship Management

### US5-1: Crowdfunding-Only Stage Launch
> **User Story:** As an organizer, I want to launch a tournament in a crowdfunding-only stage with a target funding goal and deadline, while keeping tournament operational details in draft status.

```gherkin
Scenario Outline: Launching tournament in crowdfunding-only stage
  Given an organizer configuring a crowdfunding campaign
  When the organizer sets funding goal "<goalAmount>" and deadline "<deadline>"
  Then the system <systemAction>
  And competitor registration status is "<registrationStatus>"

  Examples:
    | goalAmount | deadline       | systemAction                                                         | registrationStatus |
    | 50000 THB  | 30 Days Future | publishes campaign page with funding progress meter and pledge button| Closed             |
    | 0 THB      | 30 Days Future | rejects launch with error "Funding goal must be greater than zero"   | Draft (Unpublished)|
    | 50000 THB  | 5 Days Past    | rejects launch with error "Crowdfunding deadline must be in future"  | Draft (Unpublished)|
```

---

### US5-2: Crowdfunded Event Finalization or Early Transition
> **User Story:** As an organizer of a crowdfunded event, I want to finalize event details and open participant registrations once the funding goal is met--or manually force-start/cancel the event.

```gherkin
Scenario Outline: Transitioning crowdfunded tournament to operational state
  Given a crowdfunded tournament with funding ratio "<fundingRatio>"
  When the organizer triggers action "<action>" with explanation "<reason>"
  Then the system <systemAction>
  And the tournament state becomes "<finalState>"

  Examples:
    | fundingRatio | action       | reason              | systemAction                                                   | finalState         |
    | 100%         | OPEN_REG     | Goal met on time    | publishes venue and schedule and opens player registration     | REGISTRATION_OPEN  |
    | 60%          | FORCE_START  | Private funding add | publishes operational details and opens player registration    | REGISTRATION_OPEN  |
    | 30%          | CANCEL       | Goal unmet          | initiates sponsor refund process and marks campaign cancelled  | CANCELLED          |
    | 30%          | CANCEL       |                     | blocks action with error "Reason required to cancel tournament"| CROWDFUNDING       |
```

---

### US5-3: Sponsor Directory and Public Contact Directory
> **User Story:** As an organizer, I want to browse a directory of verified sponsor profiles with public contact information so that I can find and reach out to relevant brands.

```gherkin
Scenario Outline: Sponsor directory visibility filters
  Given a sponsor profile with verification status "<isVerified>" and public directory consent "<hasConsent>"
  When an organizer searches the sponsor directory
  Then the sponsor profile appearance in search results is "<appearance>"
  And public contact information is "<contactVisibility>"

  Examples:
    | isVerified | hasConsent | appearance | contactVisibility |
    | VERIFIED   | TRUE       | Listed     | Visible           |
    | UNVERIFIED | TRUE       | Excluded   | Hidden            |
    | VERIFIED   | FALSE      | Excluded   | Hidden            |
```

---

### US5-4: Structured Sponsorship Tier Packages
> **User Story:** As an organizer, I want to create structured sponsorship tier packages (e.g., Gold, Silver) with defined perk descriptions and a funding goal.

```gherkin
Scenario Outline: Configuring sponsorship tier packages
  Given an organizer creating a sponsorship tier
  When the organizer sets tier name "<name>", price "<price>", and available slots "<slots>"
  Then the system <systemAction>

  Examples:
    | name   | price     | slots | systemAction                                                     |
    | Gold   | 10000 THB | 2     | saves tier and displays self-service pledge card on campaign page|
    | Silver | 5000 THB  | 5     | saves tier and displays self-service pledge card on campaign page|
    | Bronze | 0 THB     | 10    | rejects tier with error "Tier amount must be greater than zero"  |
    | Custom | 15000 THB | 0     | rejects tier with error "Available slots must be at least 1"     |
```

---

### US5-5: Sponsor Checkout and Logo Showcase
> **User Story:** As a sponsor, I want to select a sponsorship package, upload my brand logo/link, and complete payment so that my brand is officially recognized.

```gherkin
Scenario Outline: Sponsor checkout, asset validation, and public recognition
  Given an available sponsorship tier with available slots count "<availableSlots>"
  When a sponsor uploads logo format "<fileFormat>" and payment outcome is "<paymentOutcome>"
  Then the system <systemAction>
  And the brand logo display on tournament page is "<logoDisplay>"

  Examples:
    | availableSlots | fileFormat | paymentOutcome | systemAction                                                    | logoDisplay |
    | 2              | PNG        | APPROVED       | records pledge, decreases slots by 1, and shows brand in sponsors| Visible    |
    | 0              | PNG        | APPROVED       | blocks checkout with error "This tier is sold out"              | Hidden      |
    | 2              | EXE        | APPROVED       | rejects file with error "Logo must be a PNG, JPG, or SVG image" | Hidden      |
    | 2              | PNG        | DECLINED       | shows error "Payment was declined" and reserves no slot         | Hidden      |
```

---

### US5-6: Automated Refund on Campaign Expiry
> **User Story:** As a sponsor, I want the system to automatically refund my pledge if a crowdfunding campaign expires without meeting its goal.

```gherkin
Scenario Outline: Automated campaign expiry refund execution
  Given a crowdfunding tournament has reached its deadline
  And final funding goal achievement ratio is "<goalRatio>"
  When the campaign deadline scheduler executes
  Then the system <systemAction>
  And sponsors receive notification "<notificationMessage>"

  Examples:
    | goalRatio | systemAction                                                      | notificationMessage                       |
    | 70%       | initiates automatic payment refund to all sponsors for this event | "Campaign did not meet goal. Refund sent."|
    | 100%      | locks funds for event disbursement and issues no refunds          | "Campaign successfully funded!"           |
    | 125%      | locks funds for event disbursement and issues no refunds          | "Campaign successfully funded!"           |
```

---

## EPIC 6: Match Scheduling and Scoring

### US6-1: Tournament Bracket & Progression Graph Generation
> **User Story:** As an organizer, I want to customize the tournament format as a graph so that the platform automatically generates the match bracket or scorecard lobby.

```gherkin
Scenario Outline: Generating tournament match bracket upon registration close
  Given an organizer configured format graph "<formatType>"
  When registration closes with "<participantCount>" approved participants
  Then the system <systemAction>
  And the bracket status becomes "<bracketStatus>"

  Examples:
    | formatType         | participantCount | systemAction                                                     | bracketStatus |
    | SINGLE_ELIMINATION | 8                | generates 7 bracket matches pairing seeded participants          | Ready         |
    | DOUBLE_ELIMINATION | 16               | generates upper and lower bracket match trees                    | Ready         |
    | SINGLE_ELIMINATION | 3                | rejects generation with error "Need minimum 4 players for bracket"| Draft         |
    | INVALID_CYCLIC     | 8                | rejects generation with error "Format progression graph is invalid"| Draft       |
```

---

### US6-2: Manual Reseeding and Matchup Swapping
> **User Story:** As an organizer or staff member, I want to manually reseed matchups or swap participants in the bracket before matches start.

```gherkin
Scenario Outline: Swapping participants in tournament bracket
  Given a bracket match in state "<matchState>"
  When an authorized user swaps two participants with explanation "<reason>"
  Then the system <systemAction>
  And match schedule updates are sent to "<notificationRecipients>"

  Examples:
    | matchState  | reason            | systemAction                                                    | notificationRecipients |
    | SCHEDULED   | Schedule conflict | swaps participants in matchup and records change audit log      | Both affected teams    |
    | IN_PROGRESS | Late swap request | blocks swap with error "Cannot swap participants in active match"| No one                 |
    | SCHEDULED   |                   | blocks swap with error "Reason is required to reseed matchup"   | No one                 |
```

---

### US6-3: Match Times and Venue Assignment
> **User Story:** As an organizer or staff member, I want to assign times and venue locations to scheduled matches so that competitors know exactly where and when to compete.

```gherkin
Scenario Outline: Match scheduling and resource conflict detection
  Given a tournament match to be scheduled
  When the scheduler assigns match time offset "<timeOffset>" and venue conflict status is "<hasConflict>"
  Then the system <systemAction>
  And the public calendar updates are "<calendarStatus>"

  Examples:
    | timeOffset | hasConflict | systemAction                                                        | calendarStatus |
    | +2 Days    | FALSE       | assigns time and venue and sends alerts to teams and officials       | Published      |
    | -1 Days    | FALSE       | rejects schedule with error "Match time cannot be in the past"      | Unchanged      |
    | +2 Days    | TRUE        | rejects schedule with error "Venue or official has a time conflict"  | Unchanged      |
```

---

### US6-4: Match Score Recording and Leaderboard Advancement
> **User Story:** As an organizer, staff member, or a referee, I want to record match scores or FFA placement rankings so that the winning teams advance automatically.

```gherkin
Scenario Outline: Recording match scores and automatic winner advancement
  Given a scheduled match between Team A and Team B
  When a user with role "<reporterRole>" submits final score Team A "<scoreA>" to Team B "<scoreB>"
  Then the system <systemAction>
  And the advanced participant is "<advancingTeam>"

  Examples:
    | reporterRole | scoreA | scoreB | systemAction                                                      | advancingTeam |
    | REFEREE      | 2      | 1      | records official result and advances winner to next bracket round | Team A        |
    | ORGANIZER    | 0      | 2      | records official result and advances winner to next bracket round | Team B        |
    | REFEREE      | 1      | 1      | rejects score with error "Elimination matches cannot end in a draw"| None         |
    | COMPETITOR   | 2      | 0      | blocks submission with error "Only referees or organizers record" | None          |
```

---

### US6-5: Public Live Bracket Tree and Standings
> **User Story:** As a competitor or spectator, I want to view the live bracket tree or standings table so that I can track event progression in real time.

```gherkin
Scenario Outline: Public live bracket view by tournament publication status
  Given a tournament in state "<tournamentState>"
  When a spectator views the live bracket URL
  Then the system <systemAction>
  And the displayed bracket information is "<bracketVisibility>"

  Examples:
    | tournamentState | systemAction                                                    | bracketVisibility   |
    | PUBLISHED       | renders live bracket tree with match scores and participant names| Real-time displayed |
    | DRAFT           | denies access with error "Tournament is not published"          | Hidden              |
    | DELETED         | displays error "Tournament not found"                           | Hidden              |
```

---

### US6-6: Tournament Finalization and Standings Archival
> **User Story:** As an organizer, I want to finalize the tournament once all rounds are complete so that placement standings are permanently archived on the public tournament page.

```gherkin
Scenario Outline: Finalizing tournament and locking standings
  Given all matches completed and unresolved score disputes count is "<disputeCount>"
  When the organizer clicks "Finalize Tournament"
  Then the system <systemAction>
  And the tournament state becomes "<finalState>"

  Examples:
    | disputeCount | systemAction                                                       | finalState |
    | 0            | archives final standings on public page and locks further editing  | Finalized  |
    | 2            | blocks finalization with error "Resolve open score disputes first" | Ongoing    |
```

---

## EPIC 7: Transaction Management

### US7-1: Secure External Payment Gateway Processing
> **User Story:** As a competitor or sponsor, I want to pay registration fees or sponsorship tier pledges through a secure external payment gateway so that my payment is securely processed.

```gherkin
Scenario Outline: External payment checkout settlement
  Given a user initiating checkout for item "<itemType>" with fee "<amount>"
  When the external payment gateway callback returns status "<callbackStatus>"
  Then the system <systemAction>
  And raw credit card numbers stored in platform database are "<cardStorage>"

  Examples:
    | itemType     | amount    | callbackStatus | systemAction                                                    | cardStorage |
    | REGISTRATION | 500 THB   | APPROVED       | marks registration fee as paid and generates payment receipt     | None        |
    | SPONSORSHIP  | 10000 THB | APPROVED       | confirms sponsorship tier purchase and triggers sponsor perks   | None        |
    | REGISTRATION | 500 THB   | DECLINED       | keeps registration unpaid and prompts user to retry payment     | None        |
```

---

### US7-2: Automatic Refund on Application Rejection or Event Cancellation
> **User Story:** As a competitor or sponsor, I want to receive an automatic refund to my original payment method if my team registration is rejected or the tournament is cancelled.

```gherkin
Scenario Outline: Triggering automated refunds for rejected or cancelled entries
  Given a captured entry fee payment with refund status "<initialRefundStatus>"
  When event trigger "<triggerEvent>" occurs
  Then the system <systemAction>
  And the payer personal ledger records an entry of type "<ledgerEntryType>"

  Examples:
    | initialRefundStatus | triggerEvent          | systemAction                                                    | ledgerEntryType |
    | NOT_REFUNDED        | APPLICATION_REJECTED  | issues automatic refund to payer's original payment method       | Refund Issued   |
    | NOT_REFUNDED        | TOURNAMENT_CANCELLED  | issues automatic refund to payer's original payment method       | Refund Issued   |
    | ALREADY_REFUNDED    | TOURNAMENT_CANCELLED  | performs no duplicate action                                    | None            |
```

---

### US7-3: Centralized Transaction Ledger and Receipt Summaries
> **User Story:** As an account owner, I want to view a centralized ledger of all my incoming and outgoing transactions with downloadable receipt summaries.

```gherkin
Scenario Outline: Viewing and filtering personal transaction ledger
  Given an account owner viewing their billing transaction history
  When the owner filters by transaction type "<txType>" and date range from "<fromDate>" to "<toDate>"
  Then the system <systemAction>
  And receipt PDF download capability is "<downloadCapability>"

  Examples:
    | txType  | fromDate   | toDate     | systemAction                                                     | downloadCapability |
    | ALL     | 2026-01-01 | 2026-06-01 | displays matching payment, refund, and payout records            | Enabled            |
    | REFUND  | 2026-01-01 | 2026-06-01 | displays only refund transaction records                         | Enabled            |
    | PAYMENT | 2026-06-01 | 2026-01-01 | shows error "Start date cannot be after end date" without query  | Disabled           |
```

---

### US7-4: Tournament Financial Breakdown
> **User Story:** As an organizer, I want to view a financial breakdown of my tournament so that I have an accurate accounting summary before funds are released.

```gherkin
Scenario Outline: Tournament financial accounting breakdown access
  Given a tournament owned by "Organizer Prime"
  When user with role "<viewerRole>" accesses the tournament financial breakdown page
  Then the system <systemAction>
  And total entry fees, sponsorships, and platform fees are "<financialVisibility>"

  Examples:
    | viewerRole      | systemAction                                                      | financialVisibility |
    | TOURNAMENT_OWNER| displays reconciled gross revenue, platform cuts, and net payout  | Visible             |
    | FINANCE_ADMIN   | displays reconciled gross revenue, platform cuts, and net payout  | Visible             |
    | COMPETITOR      | denies access with error "You do not have permission"             | Hidden              |
```

---

### US7-5: Payout Details Registration and Earnings Release
> **User Story:** As an organizer or winning competitor, I want to register my external payout details to receive net event earnings or prize money once the tournament is over.

```gherkin
Scenario Outline: Registering external payout details for earnings release
  Given an eligible user with tournament earnings
  When the user submits payout provider "<payoutProvider>" and account identifier "<accountId>"
  Then the system <systemAction>
  And payout release eligibility is "<releaseEligibility>"

  Examples:
    | payoutProvider | accountId    | systemAction                                                    | releaseEligibility |
    | PROMPTPAY      | 0812345678   | validates phone/ID format and saves verified payout destination | Ready for Payout   |
    | STRIPE_CONNECT | acct_valid12 | verifies Stripe account onboarding and links payout destination | Ready for Payout   |
    | PROMPTPAY      |              | rejects with error "Account identifier cannot be blank"         | Ineligible         |
    | STRIPE_CONNECT | acct_invalid | rejects with error "Stripe account verification failed"         | Ineligible         |
