@domain_tournament @epic2 @epic3
Feature: Tournament Discovery and Management
  As an Eventory competitor or organizer
  I want to search, view details, create, and manage tournaments
  So that competitive events can be successfully published, discovered, and operated

  Background:
    Given the Eventory API service is running

  Scenario Outline: Filter tournaments by criteria
    Given the tournament catalog contains published tournaments
    When the user searches tournaments with query "<Query>" and max fee <MaxFee>
    Then the search results count should be <ExpectedCount>

    Examples:
      | Query    | MaxFee | ExpectedCount |
      | Valorant | 1000   | 1             |
      | CS2      | 500    | 1             |
      | All      | 2000   | 3             |

  Scenario: View published tournament details
    Given a published tournament "Siam Clash" exists with prize pool 50000
    When the user requests tournament details by ID
    Then the response status code should be 200
    And the tournament name should be "Siam Clash"
    And the prize pool should be 50000

  Scenario: Create a new tournament as organizer
    Given an authenticated organizer
    When the organizer creates a tournament with name "Bangkok Masters 2026", game "Valorant", and fee 0
    Then the response status code should be 201
    And the tournament status should be "DRAFT"

  Scenario: Organizer views their tournament dashboard
    Given an organizer with created tournaments
    When the organizer views their tournament dashboard
    Then the organizer sees their created tournaments listed with accurate status badges

  Scenario: Update tournament status overrides
    Given a tournament in "REGISTRATION_OPEN" status
    When the organizer overrides the status to "IN_PROGRESS"
    Then the response status code should be 200
    And the updated tournament status should be "IN_PROGRESS"

  Scenario: Invite staff or referee by email
    Given an organizer of tournament "Bangkok Masters 2026"
    When the organizer invites "referee_dan@example.com" with role "REFEREE"
    Then the response status code should be 201
    And an invitation record is created for "referee_dan@example.com"
