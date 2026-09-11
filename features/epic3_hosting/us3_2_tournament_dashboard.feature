@epic3 @hosting @dashboard
Feature: Tournament Organizer Dashboard (US3-2)
  As an organizer
  I want to see tournament's dashboard (e.g. status, participant, etc.)
  So that I can view overall status of the tournament

  Background:
    Given an organizer "Alice" owns tournament "Summer Championship"

  Scenario: Owner views tournament dashboard metrics
    Given "Summer Championship" has 12 registered teams, 5 pending applications, and 15000 THB raised
    When "Alice" opens the tournament dashboard for "Summer Championship"
    Then the system displays the tournament status, registered participant count, and funding total
    And the system displays quick actions for registration management and team approvals

  Scenario: Non-organizer is denied access to private dashboard
    Given an authenticated user "Bob" who is neither owner nor staff of "Summer Championship"
    When "Bob" attempts to open the dashboard for "Summer Championship"
    Then the system blocks access with an "Access restricted to tournament organizers" message
    And no private participant, team, or financial metrics are shown
