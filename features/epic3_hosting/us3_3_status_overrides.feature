@epic3 @hosting @status_override
Feature: Manual Tournament Status Override (US3-3)
  As an organizer
  I want the ability to manually override the tournament status
  So that I can quickly adjust for delays, cancellations, or unexpected schedule changes

  Background:
    Given an organizer owns tournament "Spring Arena" currently in "REGISTRATION_OPEN"

  Scenario: Successfully cancel a tournament with a recorded reason
    When the organizer overrides the tournament status to "CANCELLED" with reason "Venue flooding"
    Then the system updates the tournament status to "CANCELLED"
    And the status transition audit log records the new status, reason "Venue flooding", and timestamp

  Scenario: Transition tournament from registration closed to ongoing
    Given the tournament status is "REGISTRATION_CLOSED"
    When the organizer overrides the tournament status to "ONGOING" with reason "Opening ceremony started"
    Then the tournament status becomes "ONGOING"

  Scenario: Cancelling without providing a reason is rejected
    When the organizer attempts to override the tournament status to "CANCELLED" with an empty reason
    Then the system rejects the status change
    And the tournament status remains "REGISTRATION_OPEN"

  Scenario: Disallowed status transition from completed to open is rejected
    Given the tournament status is "COMPLETED"
    When the organizer attempts to override the tournament status to "REGISTRATION_OPEN"
    Then the system rejects the transition as invalid
    And the tournament status remains "COMPLETED"
