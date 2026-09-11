@epic3 @hosting @tournament_creation
Feature: Tournament Creation (US3-1)
  As an organizer
  I want to create a tournament
  So that I can define the event, funding needs, rules, solo/team size restrictions, tournament format, and rewards

  Background:
    Given an authenticated user with an active "ORGANIZER" profile

  Scenario: Successfully create a tournament draft with valid configuration
    When the organizer creates a tournament with:
      | name                 | Thailand Major 2026  |
      | game                 | Dota 2               |
      | startAt              | 2026-10-01T10:00:00Z |
      | endAt                | 2026-10-05T18:00:00Z |
      | registrationDeadline | 2026-09-25T23:59:59Z |
      | registrationMode     | TEAM                 |
      | minTeamSize          | 5                    |
      | maxTeamSize          | 5                    |
      | capacity             | 16                   |
      | entryFee             | 1000                 |
    Then the system creates the tournament in draft status
    And the tournament is owned by the organizer

  Scenario: Reject tournament creation when end date precedes start date
    When the organizer creates a tournament with start date "2026-10-10" and end date "2026-10-05"
    Then the system rejects tournament creation with a date validation error

  Scenario: Reject tournament creation when min team size exceeds max team size
    When the organizer creates a team tournament with min team size 6 and max team size 4
    Then the system rejects tournament creation with an invalid team size error

  Scenario: Reject tournament creation when registration deadline is after start date
    When the organizer sets registration deadline "2026-10-05" after tournament start date "2026-10-01"
    Then the system rejects tournament creation with a deadline validation error
