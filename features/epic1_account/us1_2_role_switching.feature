@epic1 @account @roles
Feature: Role Profile Creation and Switching (US1-2)
  As a user
  I want to create and switch between Organizer and Sponsor profiles linked to my account
  So that I can manage events or corporate funds without maintaining separate logins

  Background:
    Given an authenticated active user "user_pro" exists

  Scenario: Create and link an Organizer profile
    Given the user has no linked organizer profile
    When the user submits an organizer profile with organization name "Arena Kings" and email "contact@arenakings.com"
    Then the system creates the organizer profile linked to the user's primary account
    And the organizer profile shows organization name "Arena Kings"

  Scenario: Create and link a Sponsor profile
    Given the user has no linked sponsor profile
    When the user submits a sponsor profile with company name "Tech Gear Corp" and email "sponsor@techgear.io"
    Then the system creates the sponsor profile linked to the user's primary account
    And the sponsor profile shows company name "Tech Gear Corp"

  Scenario: Switch active context from Competitor to Organizer
    Given the user has both active Competitor and Organizer profiles
    And the current active role context is "COMPETITOR"
    When the user switches their active role to "ORGANIZER"
    Then the active context is updated to "ORGANIZER"
    And the navigation displays tournament hosting and management controls
    And the user is not required to re-authenticate

  Scenario: Switch active context from Organizer to Sponsor
    Given the user has linked Organizer and Sponsor profiles
    And the current active role context is "ORGANIZER"
    When the user switches their active role to "SPONSOR"
    Then the active context is updated to "SPONSOR"
    And the navigation displays sponsorship tiers and brand management controls

  Scenario: Attempting to switch to an uncreated profile fails
    Given the user has not created an organizer profile
    When the user attempts to switch their active role to "ORGANIZER"
    Then the system rejects the role switch with a bad request error
    And the active role remains unchanged
