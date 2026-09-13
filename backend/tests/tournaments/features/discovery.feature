@domain_tournament @epic2
Feature: Tournament discovery
  As a potential competitor or sponsor
  I want to find suitable tournaments and see their public details
  So that I can decide whether to participate or support an event

  Background:
    Given tournaments with different schedules, fees and publication states exist

  @us2_1
  Scenario Outline: Browse published tournaments matching schedule and budget
    When a visitor searches tournaments with "<query>"
    Then the response status should be 200
    And the catalog should contain exactly "<names>" with total <total>

    Examples:
      | query                                                                            | names                         | total |
      |                                                                                  | Free Cup, Morning Cup, Late Cup, Next Day Cup, Premium Cup | 5 |
      | startFrom=2026-10-01&startTo=2026-10-01&minEntryFee=100&maxEntryFee=500              | Morning Cup, Late Cup         | 2     |
      | startFrom=2026-10-01&startTo=2026-10-01&minEntryFee=500&maxEntryFee=500              | Late Cup                      | 1     |
      | minEntryFee=0&maxEntryFee=0                                                       | Free Cup                      | 1     |
      | startFrom=2026-11-01&startTo=2026-11-30                                            |                               | 0     |
      | q=late&status=REGISTRATION_OPEN                                                  | Late Cup                      | 1     |
      | sort=fee_desc&page=2&pageSize=2                                                   | Next Day Cup, Morning Cup     | 5     |

  @us2_1
  Scenario Outline: Reject invalid filter ranges
    When a visitor searches tournaments with "<query>"
    Then the response status should be 400
    And the error detail should be "Invalid tournament filters"
    And no tournament data should be exposed

    Examples:
      | query                                                |
      | startFrom=2026-10-02&startTo=2026-10-01                 |
      | minEntryFee=501&maxEntryFee=500                        |
      | startFrom=2026-10-02&startTo=2026-10-01&minEntryFee=2&maxEntryFee=1 |
      | startFrom=not-a-date                                  |

  @us2_2
  Scenario Outline: Read current details, accepted teams and funding progress
    Given "Morning Cup" has teams in every registration status
    And "Morning Cup" has a funding goal of <goal> with <raised> raised and 4 supporters
    When a visitor opens the details for "Morning Cup"
    Then the response status should be 200
    And the public details should match "Morning Cup"
    And only the accepted team and its 2 members should be counted
    And funding should show goal <goal>, raised <raised>, remaining <remaining> and percentage <percentage>
    And team invite codes and member account data should be private
    When a visitor searches tournaments with "q=Morning"
    Then the response status should be 200
    And the catalog registration count should be 1

    Examples:
      | goal  | raised | remaining | percentage |
      | 10000 | 6250   | 3750      | 62.5       |
      | 10000 | 10000  | 0         | 100        |
      | 10000 | 12500  | 0         | 125        |
      | 0     | 0      | 0         | 0          |

  @us2_2
  Scenario: Read a tournament without funding or teams
    When a visitor opens the details for "Morning Cup"
    Then the response status should be 200
    And the details should contain empty teams and zero funding

  @us2_2
  Scenario Outline: Hide missing tournaments and another organizer's draft
    Given "Secret Draft" has teams in every registration status
    And "Secret Draft" has a funding goal of 99999 with 12345 raised and 4 supporters
    When a visitor opens the details for "<name>"
    Then the response status should be 404
    And the error detail should be "Tournament not found"
    And no tournament data should be exposed

    Examples:
      | name         |
      | Missing Cup  |
      | Secret Draft |
