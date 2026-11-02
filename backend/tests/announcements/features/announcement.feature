@domain_announcements @epic3
Feature: Tournament announcements
  As an organizer
  I want to announce further tournament details to the related parties
  So that they can receive updates about the tournament

  Background:
    Given the Eventory API service is running

  # US3-4
  # As an organizer, I want to announce further tournament details (e.g. venue
  # change) to the related parties so that they can receive updates about the
  # tournament.
  #
  # Note: the success confirmation message, the reset input fields and the
  # notification to participants are checked in the UI and in the notification
  # step of this story, not here. Owners and admins may post; a referee may not.

  # =========================================================================
  # 1. VALID (POSITIVE) SCENARIOS
  # =========================================================================

  @US3-4 @valid
  Scenario: Post an announcement to a tournament
    Given an organizer manages a published tournament
    When the organizer posts an announcement with a title and a message
    Then the announcement should appear first in the tournament's announcement list
    And the announcement should show who posted it and when

  @US3-4 @valid
  Scenario: An admin posts an announcement
    Given an organizer manages a published tournament
    And an admin has been added to that tournament's staff
    When the admin posts an announcement with a title and a message
    Then the announcement should appear first in the tournament's announcement list
    And the announcement should show who posted it and when

  @US3-4 @valid
  Scenario: Visitors see the newest announcement first
    Given an organizer manages a published tournament
    And the organizer has posted two announcements, one after the other
    When a visitor who is not signed in opens the tournament's announcement list
    Then the newest announcement should be listed first

  # =========================================================================
  # 2. INVALID (NEGATIVE) SCENARIOS
  # =========================================================================

  @US3-4 @invalid
  Scenario Outline: Block an announcement that cannot be posted
    Given a user prepares an announcement where <situation>
    When the user posts the announcement
    Then the announcement should be blocked
    And the announcement list should stay unchanged

    Examples:
      | situation                                      |
      | the title is empty                             |
      | the title contains only spaces                 |
      | the message is empty                           |
      | the title is longer than 150 characters        |
      | the message is longer than 5000 characters     |
      | the user does not manage the tournament        |
      | the user is only a referee of the tournament   |

  @US3-4 @invalid
  Scenario: Hide the announcements of a tournament that is not published
    Given an organizer manages a tournament that is not published yet
    And the organizer has posted an announcement
    When a visitor who is not signed in opens the tournament's announcement list
    Then the announcement list should not be available