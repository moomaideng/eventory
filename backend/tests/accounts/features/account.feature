@domain_account @epic1
Feature: Account Management and Profiles
  As an Eventory platform user
  I want to register an account, manage my personal profile, and update organizer/sponsor profiles
  So that I can participate, organize tournaments, and sponsor events seamlessly

  Background:
    Given the Eventory API service is running

  Scenario Outline: Successful account onboarding
    Given a new authenticated user with email "<Email>"
    When the user submits onboarding request with display name "<DisplayName>" and handle "<Handle>"
    Then the response status code should be 201
    And the created account should have display name "<DisplayName>"
    And the account handle should be valid

    Examples:
      | Email                    | DisplayName    | Handle          |
      | competitor_01@eventory.gg| Pro Competitor | pro_competitor  |
      | organizer_01@eventory.gg | Main Organizer | main_organizer  |
      | sponsor_01@eventory.gg   | Prime Sponsor  | prime_sponsor   |

  Scenario: Onboarding with empty handle falls back to a generated handle
    Given a new authenticated user
    When the user submits onboarding request with display name "No Handle User" and empty handle
    Then the response status code should be 201
    And the created account should have a non-empty handle
    And the created account should have display name "No Handle User"

  Scenario: Reject profile update when handle is already taken by another user
    Given an existing user onboarded with handle "unique_champion"
    And another user onboarded with handle "challenger_two"
    When the second user updates their profile with handle "unique_champion"
    Then the response status code should be 409

  Scenario: Retrieve current authenticated account profile
    Given an existing user onboarded with display name "Alice Gamer" and handle "alice_gamer"
    When the user requests their current account details via "GET /me"
    Then the response status code should be 200
    And the returned account email should match the user's registered email
    And the returned account display name should be "Alice Gamer"

  Scenario: Update personal profile details
    Given an existing user onboarded with display name "Bob Original" and handle "bob_original"
    When the user updates their profile with display name "Bob Updated" and phone "+66812345678"
    Then the response status code should be 200
    And the updated account display name should be "Bob Updated"
    And the updated account phone should be "+66812345678"

  Scenario: Reject profile update with blank display name
    Given an existing user onboarded with display name "Valid User" and handle "valid_user"
    When the user updates their profile with display name ""
    Then the response status code should be 422

  Scenario: Upsert and retrieve linked organizer profile
    Given an existing user onboarded with display name "Org Lead" and handle "org_lead"
    When the user upserts their organizer profile with organization "Esports Global" and email "contact@esportsglobal.gg"
    Then the response status code should be 200
    And the organizer profile name should be "Esports Global"
    When the user retrieves their organizer profile
    Then the response status code should be 200
    And the organizer profile name should be "Esports Global"
    And the organizer contact email should be "contact@esportsglobal.gg"

  Scenario: Upsert and retrieve linked sponsor profile
    Given an existing user onboarded with display name "Sponsor Rep" and handle "sponsor_rep"
    When the user upserts their sponsor profile with company "Red Bull Thailand" and email "sponsor@redbull.co.th"
    Then the response status code should be 200
    And the sponsor profile name should be "Red Bull Thailand"
    When the user retrieves their sponsor profile
    Then the response status code should be 200
    And the sponsor profile name should be "Red Bull Thailand"
    And the sponsor contact email should be "sponsor@redbull.co.th"

  Scenario: Reject unauthenticated access to account endpoints
    When an unauthenticated request is sent to "GET /me"
    Then the response status code should be 401
