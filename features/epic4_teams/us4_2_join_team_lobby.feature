@epic4 @teams @lobby_join
Feature: Join Team Lobby via Invite Link/Code (US4-2)
  As a team member
  I want to join a captain's team lobby using an invite link and complete my individual registration questions
  So that my profile and consent details are attached to the team's tournament entry

  Background:
    Given a team lobby "Shadow Strikers" exists with invite code "A9B2C3" for a tournament with max team size 3
    And "Captain Dave" is the captain in the lobby

  Scenario: Successfully join lobby using valid invite code
    Given competitor "Elena" is authenticated and not in any team for this tournament
    When "Elena" joins team lobby using invite code "A9B2C3" and completes required consent
    Then "Elena" is added to the roster of "Shadow Strikers" with role "MEMBER"
    And the roster count becomes 2 out of 3

  Scenario: Joining fails when team lobby roster is full
    Given the lobby "Shadow Strikers" already has 3 members
    When competitor "Frank" attempts to join using invite code "A9B2C3"
    Then the system rejects the join request with a "team roster full" error
    And "Frank" is not added to the roster

  Scenario: Joining fails with invalid or non-existent invite code
    When competitor "Elena" attempts to join a lobby with invite code "ZZZZZZ"
    Then the system returns an invite code not found error

  Scenario: Joining fails when team lobby is already locked
    Given the captain has locked the roster of "Shadow Strikers" with status "LOCKED"
    When competitor "Elena" attempts to join using invite code "A9B2C3"
    Then the system rejects the request because the roster is locked
