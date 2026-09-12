@domain_team_lobby @epic4
Feature: Team Lobby Management
  As a competitor
  I want to create team lobbies and join existing lobbies
  So that my team can assemble and register for tournaments

  Background:
    Given the Eventory API service is running

  Scenario Outline: Create team lobby with custom capacity
    Given an authenticated competitor
    When the competitor creates a team lobby for tournament "<TournamentID>" with team name "<TeamName>" and max members <MaxMembers>
    Then the response status code should be 201
    And the lobby team name should be "<TeamName>"
    And the creator should be the lobby captain

    Examples:
      | TournamentID                          | TeamName     | MaxMembers |
      | 11111111-1111-1111-1111-111111111111 | Cyber Wolves | 5          |
      | 22222222-2222-2222-2222-222222222222 | Duo Phantoms | 2          |

  Scenario: Join an open team lobby with an invite code
    Given an open team lobby exists with invite code "INV-7788"
    And another authenticated competitor
    When the competitor submits a request to join lobby with invite code "INV-7788"
    Then the response status code should be 200
    And the competitor is listed in the team members

  Scenario: Reject joining a full team lobby
    Given a team lobby at maximum capacity with invite code "INV-FULL"
    And another authenticated competitor
    When the competitor submits a request to join lobby with invite code "INV-FULL"
    Then the response status code should be 400
