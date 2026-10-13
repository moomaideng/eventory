@domain_application_review @epic3
Feature: Organizer application review
  As an organizer
  I want to review submitted participant applications to approve or reject entries
  So that I maintain full control over who competes in my tournament

  Background:
    Given the Eventory API service is running

  # US3-6
  # As an organizer, I want to review submitted participant applications to
  # approve or reject entries so that I maintain full control over who competes
  # in my tournament.
  #
  # Note: the backlog acceptance criteria only describe approval. Rejecting an
  # application and switching an earlier decision were agreed with the team.

  # =========================================================================
  # 1. VALID (POSITIVE) SCENARIOS
  # =========================================================================

  @US3-6 @valid
  Scenario: Approve a locked application
    Given an organizer manages a tournament that still has open spots
    And a team has locked its application for that tournament
    When the organizer approves the application
    Then the application should be accepted
    And the accepted count on the dashboard should increase

  @US3-6 @valid
  Scenario: Reject a locked application with a reason
    Given an organizer manages a tournament that still has open spots
    And a team has locked its application for that tournament
    When the organizer rejects the application with a reason
    Then the application should be rejected
    And the accepted count on the dashboard should stay the same

  @US3-6 @valid
  Scenario: Change an accepted application to rejected
    Given an organizer manages a tournament that still has open spots
    And a team application has already been accepted
    When the organizer rejects the application
    Then the application should be rejected
    And the accepted count on the dashboard should decrease

  # =========================================================================
  # 2. INVALID (NEGATIVE) SCENARIO
  # =========================================================================

  @US3-6 @invalid
  Scenario Outline: Block an approval that cannot be applied
    Given an organizer reviews an application where <situation>
    When the organizer approves the application
    Then the approval should be blocked
    And the application should keep its previous status

    Examples:
      | situation                                       |
      | the tournament has already reached its capacity |
      | the team has not locked its application yet     |
      | the application has already been accepted       |
      | the tournament is managed by another organizer  |
      