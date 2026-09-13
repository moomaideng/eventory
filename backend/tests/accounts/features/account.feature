@domain_account @epic1
Feature: Account Management and Profiles
  As an Eventory platform user
  I want to register an account, manage my personal profile, and update organizer/sponsor profiles
  So that I can participate, organize tournaments, and sponsor events seamlessly

  Background:
    Given the Eventory API service is running

  # =========================================================================
  # 1. VALID (POSITIVE) SCENARIOS - Happy Path & Business Workflows
  # =========================================================================

  # [VALID] Step 1: Initial account provisioning (JIT upon authentication)
  Scenario Outline: Initial account provisioning upon authentication
    Given a new authenticated user with email "<Email>"
    When the user provisions their initial account
    Then the response status code should be 201
    And the created account status should be "ONBOARDING"
    And the created account should have a valid default handle
    And the returned account email should match the user's registered email

    Examples:
      | Email                    |
      | competitor_01@eventory.gg|
      | organizer_01@eventory.gg |
      | sponsor_01@eventory.gg   |

  # [VALID] Step 1: Idempotent account provisioning
  Scenario: Idempotent account provisioning returns existing account
    Given an existing provisioned user with email "idempotent@eventory.gg"
    When the user provisions their initial account again
    Then the response status code should be 201
    And the returned account id should match the existing account id

  # [VALID] Step 2: Complete profile onboarding and activate account
  Scenario Outline: Complete profile onboarding and activate account
    Given a newly provisioned user in "ONBOARDING" status
    When the user completes onboarding with display name "<DisplayName>" and handle "<Handle>"
    Then the response status code should be 200
    And the account status should be "ACTIVE"
    And the updated account display name should be "<DisplayName>"
    And the updated account handle should be "<Handle>"

    Examples:
      | DisplayName    | Handle          |
      | Pro Competitor | pro_competitor  |
      | Main Organizer | main_organizer  |
      | Prime Sponsor  | prime_sponsor   |

  # [VALID] Read personal account details
  Scenario: Retrieve current authenticated account profile
    Given an active onboarded user with display name "Alice Gamer" and handle "alice_gamer"
    When the user requests their current account details via "GET /me"
    Then the response status code should be 200
    And the returned account email should match the user's registered email
    And the returned account display name should be "Alice Gamer"
    And the account status should be "ACTIVE"

  # [VALID] Update personal profile details
  Scenario: Update personal profile details
    Given an active onboarded user with display name "Bob Original" and handle "bob_original"
    When the user updates their profile with display name "Bob Updated" and phone "+66812345678"
    Then the response status code should be 200
    And the updated account display name should be "Bob Updated"
    And the updated account phone should be "+66812345678"

  # [VALID] Link and read Organizer role profile
  Scenario: Upsert and retrieve linked organizer profile
    Given an active onboarded user with display name "Org Lead" and handle "org_lead"
    When the user upserts their organizer profile with organization "Esports Global" and email "contact@esportsglobal.gg"
    Then the response status code should be 200
    And the organizer profile name should be "Esports Global"
    When the user retrieves their organizer profile
    Then the response status code should be 200
    And the organizer profile name should be "Esports Global"
    And the organizer contact email should be "contact@esportsglobal.gg"

  # [VALID] Link and read Sponsor role profile
  Scenario: Upsert and retrieve linked sponsor profile
    Given an active onboarded user with display name "Sponsor Rep" and handle "sponsor_rep"
    When the user upserts their sponsor profile with company "Red Bull Thailand" and email "sponsor@redbull.co.th"
    Then the response status code should be 200
    And the sponsor profile name should be "Red Bull Thailand"
    When the user retrieves their sponsor profile
    Then the response status code should be 200
    And the sponsor profile name should be "Red Bull Thailand"
    And the sponsor contact email should be "sponsor@redbull.co.th"

  # =========================================================================
  # 2. INVALID (NEGATIVE) SCENARIOS - Validation, Conflict & Security Guards
  # =========================================================================

  # [INVALID] Guard: Reject organizer profile creation if still in ONBOARDING (Expect 403 Forbidden)
  Scenario: Reject organizer profile creation when account is still in ONBOARDING
    Given a newly provisioned user in "ONBOARDING" status
    When the user upserts their organizer profile with organization "Illegal Org" and email "illegal@org.gg"
    Then the response status code should be 403

  # [INVALID] Guard: Reject sponsor profile creation if still in ONBOARDING (Expect 403 Forbidden)
  Scenario: Reject sponsor profile creation when account is still in ONBOARDING
    Given a newly provisioned user in "ONBOARDING" status
    When the user upserts their sponsor profile with company "Illegal Sponsor" and email "illegal@sponsor.gg"
    Then the response status code should be 403

  # [INVALID] Duplicate handle conflict check (Expect 409 Conflict)
  Scenario: Reject profile update when handle is already taken by another user
    Given an active onboarded user with handle "unique_champion"
    And another active onboarded user with handle "challenger_two"
    When the second user updates their profile with handle "unique_champion"
    Then the response status code should be 409

  # [INVALID] Blank display name validation failure (Expect 422 Unprocessable Entity)
  Scenario: Reject profile update with blank display name
    Given an active onboarded user with display name "Valid User" and handle "valid_user"
    When the user updates their profile with display name ""
    Then the response status code should be 422

  # [INVALID] Unauthenticated access rejection (Expect 401 Unauthorized)
  Scenario: Reject unauthenticated access to account endpoints
    When an unauthenticated request is sent to "GET /me"
    Then the response status code should be 401
