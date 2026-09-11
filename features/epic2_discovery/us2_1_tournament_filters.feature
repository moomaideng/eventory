@epic2 @tournaments @discovery
Feature: Tournament Browsing and Filtering (US2-1)
  As a user
  I want to browse and filter tournaments
  So that I can quickly find events that fit my schedule and budget

  Background:
    Given the tournament catalog contains the following published tournaments:
      | name               | game      | startAt    | entryFee | status            |
      | Bangkok Open 2026  | Valorant  | 2026-05-10 | 0        | REGISTRATION_OPEN |
      | Siam Clash         | CS2       | 2026-06-01 | 500      | REGISTRATION_OPEN |
      | Chiang Mai Masters | Valorant  | 2026-07-20 | 1200     | REGISTRATION_OPEN |

  Scenario: Filter tournaments by search keyword
    When the user enters search query "Valorant"
    And clicks "Apply filters"
    Then the results contain "Bangkok Open 2026" and "Chiang Mai Masters"
    And the results do not contain "Siam Clash"

  Scenario: Filter tournaments by start date range
    When the user selects "From" date "2026-05-01"
    And the user selects "To" date "2026-06-15"
    And clicks "Apply filters"
    Then the results contain "Bangkok Open 2026" and "Siam Clash"
    And the results do not contain "Chiang Mai Masters"

  Scenario: Filter tournaments by Maximum Fee
    When the user enters "Maximum fee (THB)" as 500
    And clicks "Apply filters"
    Then the results contain "Bangkok Open 2026" and "Siam Clash"
    And the results do not contain "Chiang Mai Masters"

  Scenario: Dynamic datepicker boundary constraint
    When the user selects "From" date "2026-06-01"
    Then the "To" input enforces a minimum date of "2026-06-01"
    When the user selects "To" date "2026-06-20"
    Then the "From" input enforces a maximum date of "2026-06-20"

  Scenario: Inverted date range triggers validation and disables submission
    When the user manually enters "From" date "2026-08-01" and "To" date "2026-07-01"
    Then the "To" input displays the error message "To date must be on or after From date."
    And the "Apply filters" button is disabled
    And submitting the form is prevented

  Scenario: Clear filters restores all published tournaments
    Given the user has applied search query "Valorant"
    When the user clicks "Clear filters"
    Then all filter inputs are cleared
    And the URL search parameters are removed
    And all 3 published tournaments are displayed
