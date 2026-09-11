@epic1 @account @profile
Feature: View and Update Role-Specific Profiles (US1-3)
  As an account owner
  I want to view and update my role-specific profile(s)
  So that my contact and public information stays accurate

  Background:
    Given an authenticated user "organizer_bob" with an established organizer profile "Bob Esports"

  Scenario: View own organizer profile
    When the user requests their organizer profile
    Then the system returns the organizer profile details including organization name and contact email

  Scenario: Update organizer profile with valid information
    When the user updates the organization name to "Bob Global Gaming" and contact email to "info@bobgaming.com"
    Then the system persists the updated organization details
    And subsequent requests for the organizer profile reflect "Bob Global Gaming"

  Scenario: Update organizer profile with invalid email format fails
    When the user updates the organizer contact email to "not-an-email"
    Then the system returns a validation error
    And the original organizer email remains unchanged

  Scenario: Update organizer profile with empty organization name fails
    When the user updates the organization name to ""
    Then the system returns a validation error requiring a non-empty name
    And the original organization name is preserved

  Scenario: Unauthorized user cannot modify another user's profile
    Given an authenticated user "attacker_dan"
    When "attacker_dan" attempts to update the profile belonging to "organizer_bob"
    Then the system denies the request with an authorization error
    And no modifications are recorded
