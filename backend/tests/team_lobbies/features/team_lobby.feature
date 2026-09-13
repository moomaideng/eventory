@domain_team_lobby @epic4
Feature: Tournament team lobby creation and joining
  As a competitor
  I want to create or join a tournament team lobby
  So that a team can be formed for tournament registration

  Background:
    Given the Eventory API service is running

  # US4-1
  # As a team captain, I want to create a team lobby for a specific tournament
  # and generate an invite link/code so that I can gather my teammates into a
  # unified roster before registering.

  # =========================================================================
  # 1. VALID (POSITIVE) SCENARIO
  # =========================================================================

  @US4-1 @valid
  Scenario: Create a lobby for an open team tournament
    Given a published team tournament exists with registration open
    And the competitor has signed in and completed onboarding
    When the competitor creates a team lobby named "Cyber Wolves"
    Then the team lobby should be created successfully
    And the lobby should be named "Cyber Wolves"
    And the lobby status should be "FORMING"
    And the competitor should be the lobby captain
    And the lobby should have a valid six-character invite code

  # =========================================================================
  # 2. INVALID (NEGATIVE) SCENARIO
  # =========================================================================

  @US4-1 @invalid
  Scenario: Reject lobby creation without a team name
    Given a published team tournament exists with registration open
    And the competitor has signed in and completed onboarding
    When the competitor tries to create a team lobby without entering a team name
    Then the team lobby creation should fail
    And the competitor should not be attached to a tournament team

  # US4-2
  # As a team member, I want to join a captain's team lobby using an invite link
  # so that my profile is attached to the team's tournament entry.

  # =========================================================================
  # 1. VALID (POSITIVE) SCENARIO
  # =========================================================================

  @US4-2 @valid
  Scenario: Join a captain's team lobby using its invite code
    Given a captain has created a forming team lobby
    And another competitor has signed in and completed onboarding
    When the other competitor joins the lobby using its current invite code
    Then the other competitor should join the team lobby successfully
    And the other competitor's profile should be attached to the team entry
    And the lobby should contain 2 members

  # =========================================================================
  # 2. INVALID (NEGATIVE) SCENARIO
  # =========================================================================

  @US4-2 @invalid
  Scenario: Reject joining with an unknown invite code
    Given the competitor has signed in and completed onboarding
    When the competitor tries to join with unknown invite code "ZZZZZZ"
    Then joining the team lobby should fail
    And the competitor should not be attached to a tournament team
