@epic1 @account
Feature: Primary Account Registration & Onboarding (US1-1)
  As a new user
  I want to create a single primary account
  So that I have a central identity with a public handle to interact across the platform

  Background:
    Given the authentication service is available

  Scenario: Successful onboarding with a valid unique handle and display name
    Given an authenticated user with email "newuser@example.com" has not completed onboarding
    When the user submits onboarding details with display name "Alice Gamer" and handle "alice_01"
    Then the system creates one primary account record linked to the user
    And the account status is set to "ACTIVE"
    And the profile displays handle "alice_01" and display name "Alice Gamer"

  Scenario: Onboarding fails when public handle is already claimed
    Given an account already exists with public handle "legendary_player"
    And an authenticated new user is completing onboarding
    When the user submits onboarding details with handle "legendary_player"
    Then the system rejects the account creation with a conflict error
    And the system informs the user that the handle is already taken

  Scenario: Onboarding fails with handle shorter than minimum length
    Given an authenticated new user is completing onboarding
    When the user submits a handle "al" with length 2
    Then the system rejects the request with a validation error
    And the system indicates that handle length must be at least 3 characters

  Scenario: Onboarding fails with handle containing disallowed characters
    Given an authenticated new user is completing onboarding
    When the user submits a handle "alice@#$!"
    Then the system rejects the request with a validation error
    And the system indicates that handle can only contain letters, numbers, and underscores

  Scenario: Onboarding fails when display name is blank or whitespace
    Given an authenticated new user is completing onboarding
    When the user submits a display name "   "
    Then the system rejects the request with a validation error
    And the account is not initialized
