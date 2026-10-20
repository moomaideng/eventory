@domain_sponsorship @epic5
Feature: Sponsorship tier packages
  As an organizer
  I want to create structured sponsorship tiers with perks
  So that brands can sponsor my tournament through self-serve checkout

  Background:
    Given an organizer owns a published tournament open for sponsorship

  @us5_4
  Scenario Outline: Create a sponsorship tier
    When the organizer creates a tier named "<name>" costing <price> with <slots> slots and perks "<perks>"
    Then the response status should be <code>
    And sponsors should see exactly "<visible>" on the tournament

    Examples:
      | name   | price | slots | perks                      | code | visible |
      | Gold   | 10000 | 2     | Logo on stream; Banner     | 201  | Gold    |
      | Bronze | 0     | 10    | Shout-out                  | 422  |         |
      | Custom | 15000 | 0     | Custom package             | 422  |         |
      |        | 3000  | 3     | Logo on page               | 422  |         |
      | Silver | 5000  | 5     |                            | 422  |         |

  @us5_4
  Scenario: Tier names are unique within a tournament
    Given the organizer has a tier named "Gold"
    When the organizer creates a tier named "Gold" costing 8000 with 1 slots and perks "Logo"
    Then the response status should be 409
    And sponsors should see exactly "Gold" on the tournament

  @us5_4
  Scenario: Tiers are listed in display order
    Given the organizer has tiers "Silver" with order 2 and "Gold" with order 1
    Then sponsors should see exactly "Gold, Silver" on the tournament

  @us5_4
  Scenario: The organizer edits and deletes tiers
    Given the organizer has a tier named "Gold"
    When the organizer renames tier "Gold" to "Platinum"
    Then the response status should be 200
    And sponsors should see exactly "Platinum" on the tournament
    When the organizer deletes tier "Platinum"
    Then the response status should be 204
    And sponsors should see exactly "" on the tournament

  @us5_4
  Scenario: Only the owner can manage tiers
    Given the organizer has a tier named "Gold"
    And another organizer exists
    When the other organizer creates a tier named "Rival" costing 1000 with 1 slots and perks "Logo"
    Then the response status should be 404
    When the other organizer deletes tier "Gold"
    Then the response status should be 404
    And sponsors should see exactly "Gold" on the tournament

  @us5_4
  Scenario Outline: Tiers are locked once the tournament is over
    Given the tournament status is "<status>"
    When the organizer creates a tier named "Late" costing 1000 with 1 slots and perks "Logo"
    Then the response status should be 409

    Examples:
      | status    |
      | COMPLETED |
      | CANCELLED |

  @us5_4
  Scenario: Unpublished tournaments do not show tiers
    Given the organizer has a tier named "Gold"
    And the tournament is unpublished
    When a visitor lists the tournament's tiers
    Then the response status should be 404

  @us5_4
  Scenario: The owner manages tiers before publishing
    Given the organizer has a tier named "Gold"
    And the tournament is unpublished
    And another organizer exists
    When the organizer lists tiers to manage
    Then the response status should be 200
    And the managed tiers should be exactly "Gold"
    When the other organizer lists tiers to manage
    Then the response status should be 404
