@epic4 @teams @lobby_creation
Feature: Team Lobby Creation (US4-1)
  As a team captain
  I want to create a team lobby for a specific tournament and generate an invite link/code
  So that I can gather my teammates into a unified roster before registering

  Background:
    Given a team tournament "Valorant Masters" is accepting registrations
    And the tournament has maximum team size 5
    And competitor "Captain Dave" has no existing team in this tournament

  Scenario: Successfully create a team lobby
    When "Captain Dave" creates a team lobby named "Shadow Strikers" for "Valorant Masters"
    Then the system creates the lobby with status "FORMING"
    And assigns "Captain Dave" as the captain and first roster member
    And generates a unique 6-character alphanumeric invite code
    And provides a shareable join link

  Scenario: Attempting to create team lobby for a SOLO tournament fails
    Given a tournament "Chess Blitz" configured with registration mode "SOLO"
    When "Captain Dave" attempts to create a team lobby for "Chess Blitz"
    Then the system rejects the request with an error indicating only solo registration is supported

  Scenario: Captain cannot create multiple team lobbies in the same tournament
    Given "Captain Dave" already created a lobby "Shadow Strikers" in "Valorant Masters"
    When "Captain Dave" attempts to create another lobby "Second Strikers" in "Valorant Masters"
    Then the system rejects the duplicate lobby creation
