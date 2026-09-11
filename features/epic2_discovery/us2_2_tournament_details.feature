@epic2 @tournaments @details
Feature: Tournament Details, Rosters, and Crowdfunding (US2-2)
  As a user (potential competitor or sponsor)
  I want to view tournament details, teams, and funding progress
  So that I can decide whether to participate or provide support

  Background:
    Given a published tournament "Apex Invitational" exists with:
      | game             | Apex Legends |
      | entryFee         | 300          |
      | currency         | THB          |
      | registrationMode | TEAM         |
      | minTeamSize      | 3            |
      | maxTeamSize      | 3            |
      | capacity         | 20           |
      | fundingGoal      | 50000        |
      | raisedAmount     | 25000        |

  Scenario: View public tournament overview and schedule
    When a user opens the tournament detail page for "Apex Invitational"
    Then the system displays the event game, schedule, entry fee, and registration deadline
    And the system displays the funding progress as 50 percent of the 50000 THB goal

  Scenario: View approved teams roster
    Given 2 teams "Red Dragons" and "Blue Wolves" are approved for "Apex Invitational"
    When a user views the teams section on the tournament page
    Then the system displays "Red Dragons" and "Blue Wolves" with member counts

  Scenario: Requesting a non-existent tournament displays not found page
    When a user requests tournament details for an unknown tournament
    Then the system displays a "Tournament not found" notice
    And no tournament information is rendered
