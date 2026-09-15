@domain_tournament @epic3
Feature: Manual tournament status override
  As an organizer
  I want to move my tournament to another lifecycle status with a recorded reason
  So that I can adjust for delays, cancellations, or unexpected schedule changes

  @us3_3
  Scenario Outline: Move a tournament to a permitted status and record the change
    Given an organizer owns a tournament in "<currentStatus>" status
    When the organizer overrides the status to "<targetStatus>" with reason "Venue double-booked, shifting the schedule"
    Then the response status should be 200
    And the tournament status should be "<targetStatus>"
    And the audit trail should record "<currentStatus>" to "<targetStatus>" with the acting organizer and a timestamp

    Examples:
      | currentStatus       | targetStatus        |
      | REGISTRATION_OPEN   | REGISTRATION_CLOSED |
      | REGISTRATION_OPEN   | ONGOING             |
      | REGISTRATION_CLOSED | REGISTRATION_OPEN   |
      | ONGOING             | REGISTRATION_CLOSED |
      | ONGOING             | COMPLETED           |
      | ONGOING             | CANCELLED           |
      | DRAFT               | REGISTRATION_OPEN   |
      | CROWDFUNDING        | REGISTRATION_OPEN   |

  @us3_3
  Scenario Outline: Reject transitions that are not permitted from the current status
    Given an organizer owns a tournament in "<currentStatus>" status
    When the organizer overrides the status to "<targetStatus>" with reason "Venue double-booked, shifting the schedule"
    Then the response status should be 422
    And the error detail should be "That status change is not allowed from the tournament's current status"
    And the tournament status should be "<currentStatus>"
    And no status change should be recorded

    Examples:
      | currentStatus     | targetStatus      |
      | COMPLETED         | ONGOING           |
      | COMPLETED         | CANCELLED         |
      | CANCELLED         | REGISTRATION_OPEN |
      | REGISTRATION_OPEN | DRAFT             |
      | ONGOING           | CROWDFUNDING      |

  @us3_3
  Scenario: Reject a change that leaves the status untouched
    Given an organizer owns a tournament in "ONGOING" status
    When the organizer overrides the status to "ONGOING" with reason "Venue double-booked, shifting the schedule"
    Then the response status should be 422
    And the error detail should be "Tournament is already in that status"
    And no status change should be recorded

  @us3_3
  Scenario Outline: Require an explanation before changing the status
    Given an organizer owns a tournament in "REGISTRATION_OPEN" status
    When the organizer overrides the status to "CANCELLED" with a reason of <length> characters
    Then the response status should be 422
    And the tournament status should be "REGISTRATION_OPEN"
    And no status change should be recorded

    Examples:
      | length |
      | 0      |
      | 9      |
      | 501    |

  @us3_3
  Scenario: Reject an override from a signed-out visitor
    Given an organizer owns a tournament in "REGISTRATION_OPEN" status
    When an anonymous visitor overrides the status to "CANCELLED" with reason "Venue double-booked, shifting the schedule"
    Then the response status should be 401
    And the tournament status should be "REGISTRATION_OPEN"
    And no status change should be recorded

  @us3_3
  Scenario: Hide another organizer's tournament instead of refusing access
    Given an organizer owns a tournament in "REGISTRATION_OPEN" status
    And another organizer exists
    When the other organizer overrides the status to "CANCELLED" with reason "Venue double-booked, shifting the schedule"
    Then the response status should be 404
    And the error detail should be "Tournament not found or not owned by your account"
    And the tournament status should be "REGISTRATION_OPEN"
    And no status change should be recorded

  @us3_3
  Scenario: Read the change history and the statuses still reachable
    Given an organizer owns a tournament in "REGISTRATION_OPEN" status
    When the organizer overrides the status to "ONGOING" with reason "Venue confirmed early, starting ahead of schedule"
    And the organizer overrides the status to "COMPLETED" with reason "All matches finished and results recorded"
    And the organizer reads the status history
    Then the response status should be 200
    And the current status should be "COMPLETED"
    And the history should list 2 changes, newest first
    And the history should not expose the organizer's email

  @us3_3
  Scenario Outline: Offer only the statuses reachable from the current one
    Given an organizer owns a tournament in "<currentStatus>" status
    When the organizer reads the status history
    Then the response status should be 200
    And the allowed transitions should be "<allowed>"

    Examples:
      | currentStatus     | allowed                                                  |
      | REGISTRATION_OPEN | REGISTRATION_CLOSED, ONGOING, COMPLETED, CANCELLED       |
      | ONGOING           | REGISTRATION_CLOSED, COMPLETED, CANCELLED                |
      | COMPLETED         |                                                          |
      | CANCELLED         |                                                          |

  @us3_3
  Scenario: Hide the change history from another organizer
    Given an organizer owns a tournament in "REGISTRATION_OPEN" status
    And another organizer exists
    When the other organizer reads the status history
    Then the response status should be 404
    And the error detail should be "Tournament not found or not owned by your account"
