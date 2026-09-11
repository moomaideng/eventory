@epic3 @hosting @staff_roles
Feature: Invite and Assign Tournament Staff Roles (US3-7)
  As an organizer
  I want to be able to invite and assign roles (staff, referee) to other accounts
  So that I don't have to share my credentials to others to have them help manage the tournament

  Background:
    Given organizer "Alice" owns tournament "Mega Tourney"
    And user "Charlie" has an active primary account

  Scenario: Organizer generates a staff invitation link for a referee
    When "Alice" generates an invitation for tournament "Mega Tourney" with role "REFEREE"
    Then the system produces a unique invitation token
    And sets an expiration time for the invitation

  Scenario: Invitee accepts valid staff invitation
    Given an active invitation token exists for role "REFEREE" on "Mega Tourney"
    When "Charlie" accepts the invitation token using their account
    Then the system assigns "Charlie" the role "REFEREE" for "Mega Tourney"
    And "Charlie" gains permissions to record match scores
    And "Charlie" cannot modify tournament core settings or delete the tournament

  Scenario: Accepting an expired invitation token fails
    Given an invitation token for "Mega Tourney" has expired
    When "Charlie" attempts to accept the expired token
    Then the system rejects the acceptance with an expired token error
    And no staff permissions are granted

  Scenario: Non-organizer cannot invite staff
    Given a user "Mallory" who is not an organizer of "Mega Tourney"
    When "Mallory" attempts to generate a staff invitation for "Mega Tourney"
    Then the system denies the action with an authorization error
