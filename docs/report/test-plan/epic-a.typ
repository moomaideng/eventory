// ============================================================
// Owners: Eff (stories 1–2), Pun (stories 3–4). Swap if you like.
// Task:   Black-box test plan for 4 user stories of ONE EPIC
//
// Pick the EPIC with Mavin/Imtaez (epic-b.typ uses a different one).
// Stories + acceptance criteria: Backlog Master (Google Sheet)
//   https://docs.google.com/spreadsheets/d/1KqyqqsrhDQ_XMk7KXhWC7pmIDYf_PSL3XvMez0JwPXg/edit?gid=1930282482#gid=1930282482
//
// Per story:
//   1. ec-table: split each input into equivalence classes
//      (EC = a group of inputs the system should treat the same).
//      Give each input at least 1 valid + 1 invalid class, with a concrete example.
//      EC IDs only need to be unique within the story.
//   2. test-cases: cover EVERY EC at least once. One card per case,
//      field list + screenshot example: see test-cases in appendix/preambles.typ.
//      TC IDs must be unique across the WHOLE document (TC01, TC02, ...).
//      `actual` + `status` ("P"/"F") are filled after running the tests;
//      leave them out until then.
// The coverage matrix at the end is built automatically.
// Where: `backend/tests/<domain>/` (godog `.feature` + steps), `backend/internal/**/*_test.go`
// ============================================================
#import "../appendix/preambles.typ": *

= EPIC 1: Account and Role Access

// ---------- Story 1 (ImtaeZ) ----------
== US 1-1: Account Creation & Onboarding
_As a new user, I want to create a single primary account,
so that I have a central identity with a public handle
to interact across the platform._

=== Pre-condition
- The user has authenticated for the first time
  and has not completed onboarding.
- The onboarding form is displayed.
- Each test case starts independently with a fresh user
  who has not completed onboarding.

=== Input Equivalence Classes
#ec-table(
  ("EC1", "Handle",
    [Valid format, 3–32 characters, and available.], true),

  ("EC2", "Handle",
    [Already taken by another account.], false),

  ("EC3", "Handle",
    [Fewer than 3 characters.], false),

  ("EC4", "Handle",
    [More than 32 characters.], false),

  ("EC5", "Display Name",
    [Non-empty.], true),

  ("EC6", "Display Name",
    [Empty string], false),
)

=== Test Cases
#test-cases(
  story: "US 1-1",
      (
    id: "TC01",
    action: [
      Enter the handle and display name,
      then submit the onboarding form.
    ],
    inputs: (
      Handle: [`imtaez_test01`],
      "Display Name": [`Imtaez Test`],
    ),
    expected: [
      The system creates and activates the primary account
      with the submitted handle and display name,
      and allows the user to enter the platform.

      #image("../img/tc01-1.png", width: 100%)
      #image("../img/tc01-2.png", width: 100%)
    ],
    actual: [
      The onboarding form closed after submission.
      The platform home page was displayed,
      with "Imtaez Test" shown in the navigation bar.
    ],
    ecs: "EC1, EC5",
    status: "P",
  ),
    (
    id: "TC02",
    pre: "Another account already owns the handle `taken_user`.",
    action: [
      Enter an existing handle and a non-empty display name,
      then submit the onboarding form.
    ],
    inputs: (
      Handle: [`taken_user`],
      "Display Name": [`Imtaez Test`],
    ),
    expected: [
      The system prevents onboarding completion
      and displays an error indicating
      that the handle is already taken.

      #image("../img/tc02.png", width: 100%)
    ],
    actual: [
      The system displayed "Handle is already taken. Please choose another one."
      The onboarding form remained open,
      preventing the user from completing onboarding.
    ],
    ecs: "EC2, EC5",
    status: "P",
  ),
    (
    id: "TC03",
    action: [
      Enter a two-character handle and a non-empty
      display name, then attempt to submit the form.
    ],
    inputs: (
      Handle: [`ab`],
      "Display Name": [`Imtaez Test`],
    ),
    expected: [
      The "Confirm & Start Playing" button is disabled
      when the handle contains fewer than 3 characters,
      preventing onboarding completion.

      #image("../img/tc03.png", width: 100%)
    ],
    actual: [
      The "Confirm & Start Playing" button was disabled
      when the handle was set to "ab",
      preventing onboarding completion.
    ],
    ecs: "EC3, EC5",
    status: "P",
  ),
    (
    id: "TC04",
    action: [
      Attempt to enter a 33-character handle
      and submit the form with a non-empty display name.
    ],
    inputs: (
      Handle: [`abcdefghijklmnopqrstuvwxyzabcdefg`],
      "Display Name": [`Imtaez Test`],
    ),
    expected: [
      The handle field accepts a maximum of 32 characters
      and prevents additional characters from being entered.

      #image("../img/tc04.png", width: 100%)
    ],
    actual: [
      The handle field accepted only the first 32 characters
      of the supplied handle. The 33rd character could not be entered.
    ],
    ecs: "EC4, EC5",
    status: "P",
  ),
    (
    id: "TC05",
    action: [
      Enter an available valid handle,
      clear the display name,
      then attempt to submit the form.
    ],
    inputs: (
      Handle: [`imtaez_test01`],
      "Display Name": [Empty string (`""`)],
    ),
    expected: [
      The "Confirm & Start Playing" button is disabled
      when the display name is empty,
      preventing onboarding completion.

      #image("../img/tc05.png", width: 100%)
    ],
    actual: [
      The "Confirm & Start Playing" button was disabled
      when the display name was empty,
      preventing onboarding completion.
    ],
    ecs: "EC1, EC6",
    status: "P",
  ),
)

// ---------- Story 2 (ImtaeZ) ----------
== US 1-2: Organizer & Sponsor Role Profiles
_As a user, I want to create and switch between Organizer
and Sponsor profiles linked to my account,
so that I can manage events or corporate funds
without maintaining separate logins._

=== Pre-condition
The user is authenticated. Each test case specifies the required account status,
whether the relevant role profile already exists, and any additional setup.

=== Input Equivalence Classes
#ec-table(
  ("EC1", "Account Status",
    [ACTIVE: the user has completed onboarding.], true),

  ("EC2", "Account Status",
    [ONBOARDING: the user has not completed onboarding.], false),

  ("EC3", "Profile Type",
    [Organizer profile.], true),

  ("EC4", "Profile Type",
    [Sponsor profile.], true),

  ("EC5", "Profile Name",
    [Non-empty.], true),

  ("EC6", "Profile Name",
    [Empty string.], false),

  ("EC7", "Contact Email",
    [Valid email format.], true),

  ("EC8", "Profile Consent",
    [Privacy consent checkbox is checked.], true),

  ("EC9", "Profile Consent",
    [Privacy consent checkbox is not checked.], false),
)

=== Test Cases
#test-cases(
  story: "US 1-2",
    (
    id: "TC06",
    pre: [
      The user is authenticated with an ACTIVE account
      and does not have an Organizer profile.
    ],
    action: [
      Open Organizer profile creation,
      enter valid details, and submit the form.
      Then open the Organizer workspace.
    ],
    inputs: (
      "Profile Type": [Organizer],
      "Organization Name": [`Imtaez Events`],
      "Contact Email": [`organizer@example.com`],
      "Profile Consent": [Checked],
    ),
    expected: [
      The system creates an Organizer profile linked to the current
      primary account using the submitted name and contact email.

      #image("../img/tc06-profile.png", width: 100%)
      #image("../img/tc06-workspace.png", width: 100%)
    ],
    actual: [
      The Organizer profile setup form closed after submission.
      The "Host a Tournament" page became accessible.
    ],
    ecs: "EC1, EC3, EC5, EC7, EC8",
    status: "P",
  ),
    (
    id: "TC07",
    pre: [
      The user is authenticated with an ACTIVE account
      and does not have a Sponsor profile.
    ],
    action: [
      Open Sponsor profile creation,
      enter valid details, and submit the form.
      Then open the Sponsor workspace.
    ],
    inputs: (
      "Profile Type": [Sponsor],
      "Company Name": [`Imtaez Brand`],
      "Contact Email": [`sponsor@example.com`],
      Tags: [`gaming`],
    ),
    expected: [
      The system creates a Sponsor profile linked to the current
      primary account using the submitted name and contact email.

      #image("../img/tc07-profile.png", width: 100%)
      #image("../img/tc07-workspace.png", width: 100%)
    ],
    actual: [
      The Sponsor profile setup form closed after submission.
      The Sponsor dashboard was displayed.
    ],
    ecs: "EC1, EC4, EC5, EC7",
    status: "P",
  ),
    (
    id: "TC08",
    pre: [
      The user is authenticated with an ONBOARDING account
      and does not have an Organizer profile.
      A valid access token for this user is available for API requests.
    ],
    action: [
      Using an HTTP client, send an authenticated PUT request to
      `/api/v1/accounts/me/organizer-profile`
      using the user's Bearer access token and valid profile details.
      Inspect the response status and error message.
      Send a GET request to the same endpoint to verify
      that no Organizer profile was created.
    ],
    inputs: (
      "Profile Type": [Organizer],
      "Organization Name": [`Imtaez Events`],
      "Contact Email": [`organizer@example.com`],
    ),
    expected: [
      Reject profile creation with HTTP 403 Forbidden.
      GET returns HTTP 404 because no Organizer profile exists.

      #image("../img/tc08-1.png", width: 100%)
      #image("../img/tc08-2.png", width: 100%)
    ],
    actual: [
      PUT returned 403: "Onboarding required: please complete your profile first".
      GET returned 404: "Organizer profile not found".
    ],
    ecs: "EC2, EC3, EC5, EC7",
    status: "P",
  ),
    (
    id: "TC09",
    pre: [
      The user is authenticated with an ONBOARDING account
      and does not have a Sponsor profile.
      A valid access token for this user is available for API requests.
    ],
    action: [
      Using an HTTP client, send an authenticated PUT request to
      `/api/v1/accounts/me/sponsor-profile`
      using the user's Bearer access token and valid profile details.
      Inspect the response status and error message.
      Send a GET request to the same endpoint to verify
      that no Sponsor profile was created.
    ],
    inputs: (
      "Profile Type": [Sponsor],
      "Company Name": [`Imtaez Brand`],
      "Contact Email": [`sponsor@example.com`],
      Tags: [`Gaming`],
    ),
    expected: [
      Reject profile creation with HTTP 403 Forbidden.
      GET returns HTTP 404 because no Sponsor profile exists.

      #image("../img/tc09-1.png", width: 100%)
      #image("../img/tc09-2.png", width: 100%)
    ],
    actual: [
      PUT returned 403: "Onboarding required: please complete your profile first".
      GET returned 404: "Sponsor profile not found".
    ],
    ecs: "EC2, EC4, EC5, EC7",
    status: "P",
  ),
    (
    id: "TC10",
    pre: [
      The user is authenticated with an ACTIVE account
      and does not have an Organizer profile.
      The Organizer profile creation form is displayed.
    ],
    action: [
      Leave the organization name empty,
      enter a valid contact email,
      and attempt to submit the form.
    ],
    inputs: (
      "Profile Type": [Organizer],
      "Organization Name": [Empty string (`""`)],
      "Contact Email": [`organizer@example.com`],
    ),
    expected: [
      The system rejects Organizer profile creation when the
      organizer name is empty and displays "Enter an organizer name."


      #image("../img/tc10.png", width: 100%)
    ],
    actual: [
      The form displayed "Enter an organizer name." below the empty Organizer Name field.
    ],
    ecs: "EC1, EC3, EC6, EC7",
    status: "P",
  ),
    (
    id: "TC11",
    pre: [
      The user is authenticated with an ACTIVE account
      and does not have a Sponsor profile.
      The Sponsor profile creation form is displayed.
    ],
    action: [
      Leave the company name empty,
      enter a valid contact email,
      and attempt to submit the form.
    ],
    inputs: (
      "Profile Type": [Sponsor],
      "Sponsor Name": [Empty string (`""`)],
      "Contact Email": [`sponsor@example.com`],
    ),
    expected: [
      The system rejects Sponsor profile creation when the
      sponsor name is empty and displays "Enter a sponsor name."

      #image("../img/tc11.png", width: 100%)
    ],
    actual: [
      The form displayed "Enter a sponsor name." below the empty Sponsor Name field.
    ],
    ecs: "EC1, EC4, EC6, EC7",
    status: "P",
  ),
  (
    id: "TC32",
    pre: [
      The user is authenticated with an ACTIVE account
      and does not have an Organizer profile.
      The Organizer profile creation form is displayed.
    ],
    action: [
      Enter a valid organizer name and contact email.
      Leave the privacy consent checkbox unchecked
      and submit the form.
    ],
    inputs: (
      "Profile Type": [Organizer],
      "Organization Name": [`Imtaez Events`],
      "Contact Email": [`organizer@example.com`],
      "Profile Consent": [Unchecked],
    ),
    expected: [
      A privacy consent error is shown.
      The form remains open and no Organizer profile is created.

      #image("../img/tc32.png", width: 100%)
    ],
    actual: [
      The form displayed "Please confirm consent to store and display your organizer profile."
      and remained open without creating the profile.
    ],
    ecs: "EC1, EC3, EC5, EC7, EC9",
    status: "P",
  ),

)


// ---------- Story 3 (Mavin) ----------
== US 1-3: Organizer & Sponsor Profile Viewing & Updates
_As an account owner, I want to view and update my role-specific profile(s), so that my contact and public information stays accurate._

=== Pre-condition
- The tester is signed in as an active account owner and has opened the Profile settings page.
- The account already has an Organizer profile and a Sponsor profile.

=== Input Equivalence Classes
#ec-table(
  ("EC1", "Name", 
    [Name is 1-150 characters for an Organizer, or 1-120 for a Sponsor, 
    and may contain Thai or special characters (e.g. `Mavin`, `Dec 48`, `อิมเต้`, `!@#$`)], true),
  ("EC2", "Name",
    [Name with leading or trailing spaces (e.g. `  Imtaez  `); the spaces are trimmed on save], true),
  ("EC3", "Name",
    [Empty name], false),
  ("EC4", "Name",
    [Name that contains only spaces (e.g. `"   "`)], false),
  ("EC5", "Contact Email",
    [A valid email address (e.g. `mavin@example.com`, `imtaez@example.com`)], true),
  ("EC6", "Contact Email",
    [A text that is not a valid email address (e.g. `dec48.com`)], false),
  ("EC7", "Contact Email",
    [Empty email on an Organizer profile (optional field; the account email is used instead)], true),
  ("EC8", "Contact Email",
    [Empty email on a Sponsor profile (required field)], false),
  ("EC9", "Phone",
    [A valid phone number, optionally starting with `+` (e.g. `+66 89 999 9999`)], true),
  ("EC10", "Phone",
    [Empty phone (optional field)], true),
  ("EC11", "Phone",
    [Phone shorter than 6 characters (e.g. `999`)], false),
  ("EC12", "Phone",
    [Phone that contains letters (e.g. `abc999`)], false),
)

=== Test Cases
#test-cases(
  story: "US 1-3",
  (
    id: "TC22",
    action: [Enter the Organizer details, save the profile, then reload the page.],
    inputs: (Name: [`Mavin`], "Contact Email": [`mavin@example.com`], Phone: [`+66 89 999 9999`]),
    expected: [
      The system saves the profile and shows the same details after the reload.

      #image("../img/tc22.png", width: 100%)
    ],
    actual: [
      After saving and reloading, the Organizer profile showed `Mavin`,
      `mavin@example.com` and `+66 89 999 9999`.
    ],
    ecs: "EC1, EC5, EC9",
    status: "P",
  ),
  (
    id: "TC23",
    action: [Enter the Sponsor details with a different phone, save the profile, then reload the page.],
    inputs: (Name: [`อิมเต้`], "Contact Email": [`imtaez@example.com`], Phone: [`+66 81 111 1111`]),
    expected: [
      The system saves the Sponsor profile and shows it after the reload.
      The Organizer phone stays `+66 89 999 9999`.

      #image("../img/tc23.png", width: 100%)
    ],
    actual: [
      After saving and reloading, the Sponsor profile showed `อิมเต้`,
      `imtaez@example.com` and `+66 81 111 1111`. The Organizer phone stayed
      `+66 89 999 9999`.
    ],
    ecs: "EC1, EC5, EC9",
    status: "P",
  ),
  (
    id: "TC24",
    action: [Enter the Organizer name with spaces, clear the email and phone, then save the profile.],
    inputs: (Name: [`  Imtaez  `], "Contact Email": [empty], Phone: [empty]),
    expected: [
      The system saves the profile and shows `Imtaez` (spaces removed) and the account email as the contact email, with no phone.

      #image("../img/tc24-1.png", width: 100%)
      #image("../img/tc24-2.png", width: 100%)
    ],
    actual: [
      After saving, the Organizer profile showed `Imtaez` (spaces removed),
      the account email `dev@eventory.gg` as the contact email, and the phone
      as "Not set".
    ],
    ecs: "EC2, EC7, EC10",
    status: "P",
  ),
  (
    id: "TC25",
    action: [Clear the Organizer name, then save the profile.],
    inputs: (Name: [empty], "Contact Email": [`mavin@example.com`], Phone: [empty]),
    expected: [
      The system shows an error under Name and does not save the profile.

      #image("../img/tc25.png", width: 100%)
    ],
    actual: [
      After pressing Save profile, the message "Enter an organizer name."
      appeared under the field and the form stayed open.
    ],
    ecs: "EC3",
    status: "P",
  ),
  (
    id: "TC26",
    action: [Replace the Sponsor name with spaces, then save the profile.],
    inputs: (Name: [`"   "`], "Contact Email": [`imtaez@example.com`], Phone: [empty]),
    expected: [
      The system shows an error under Name and does not save the profile.

      #image("../img/tc26.png", width: 100%)
    ],
    actual: [
      After pressing Save profile, the message "Enter a sponsor name."
      appeared under the field and the form stayed open.
    ],
    ecs: "EC4",
    status: "P",
  ),
  (
    id: "TC27",
    action: [Enter an invalid email in the Organizer profile, then save the profile.],
    inputs: (Name: [`Dec 48`], "Contact Email": [`dec48.com`], Phone: [empty]),
    expected: [
      The system shows an error under Contact Email and does not save the profile.

      #image("../img/tc27.png", width: 100%)
    ],
    actual: [
      After pressing Save profile, the message "Enter a valid email address."
      appeared under the field and the form stayed open.
    ],
    ecs: "EC6",
    status: "P",
  ),
  (
    id: "TC28",
    action: [Clear the email in the Sponsor profile, then save the profile.],
    inputs: (Name: [`Dec 48`], "Contact Email": [empty], Phone: [empty]),
    expected: [
      The system shows an error under Contact Email and does not save the profile.

      #image("../img/tc28.png", width: 100%)
    ],
    actual: [
      After pressing Save profile, the message "Enter a valid contact email."
      appeared under the field and the form stayed open.
    ],
    ecs: "EC8",
    status: "P",
  ),
  (
    id: "TC29",
    action: [Enter a short phone in the Organizer profile, then save the profile.],
    inputs: (Name: [`Mavin`], "Contact Email": [`mavin@example.com`], Phone: [`999`]),
    expected: [
      The system shows an error under Phone and does not save the profile.

      #image("../img/tc29.png", width: 100%)
    ],
    actual: [
      After pressing Save profile, the message "Enter a valid phone number,
      e.g. +66 81 234 5678." appeared under the field and the form stayed open.
    ],
    ecs: "EC11",
    status: "P",
  ),
  (
    id: "TC30",
    action: [Enter a phone with letters in the Sponsor profile, then save the profile.],
    inputs: (Name: [`Imtaez`], "Contact Email": [`imtaez@example.com`], Phone: [`abc999`]),
    expected: [
      The system shows an error under Phone and does not save the profile.

      #image("../img/tc30.png", width: 100%)
    ],
    actual: [
      After pressing Save profile, the message "Enter a valid phone number,
      e.g. +66 81 234 5678." appeared under the field and the form stayed open.
    ],
    ecs: "EC12",
    status: "P",
  ),
  (
    id: "TC31",
    action: [Enter special characters as the Sponsor name, clear the phone, then save the profile.],
    inputs: (Name: [`!@#$`], "Contact Email": [`imtaez@example.com`], Phone: [empty]),
    expected: [
      The system saves the profile and shows `!@#$` as the name.

      #image("../img/tc31.png", width: 100%)
    ],
    actual: [
      After saving, the Sponsor profile showed `!@#$` as the name, with
      `imtaez@example.com` and the phone as "Not set".
    ],
    ecs: "EC1, EC5, EC10",
    status: "P",
  ),
)

// ---------- Story 4 (Mavin) ----------
== US 1-4: Primary Account Profile Viewing & Updates
_As an account owner, I want to view and update my primary account profile, so that my information stays accurate._

=== Pre-condition
- The tester is signed in as an account owner with an active account.
- The Profile settings page is open and the card with the Display name and Handle fields is shown.
- Another account with the handle `player_one` already exists.
- Each test case starts from the saved profile shown on the page.

=== Input Equivalence Classes
#ec-table(
  ("EC1", "Display Name",
    [1–64 characters after trimming, e.g. `Imtaez`, `Dec48` or `!@#$`.], true),
  ("EC2", "Display Name",
    [Valid text with leading or trailing spaces, e.g. `  Dec48  `
    (the spaces are removed when saved).], true),
  ("EC3", "Display Name",
    [Empty, e.g. `""`.], false),
  ("EC4", "Display Name",
    [Only spaces, e.g. `"   "`.], false),
  ("EC5", "Display Name",
    [More than 64 characters, e.g. `a` repeated 65 times.], false),
  ("EC6", "Handle",
    [3–24 characters using only lowercase letters, numbers and
    underscores, and not used by another account, e.g. `imtaez_01` or `dec_48`.
    Uppercase letters are converted to lowercase.], true),
  ("EC7", "Handle",
    [Contains a character that is not allowed, e.g. `john-doe`
     or `john doe`.], false),
  ("EC8", "Handle",
    [Fewer than 3 characters, e.g. `aa`.], false),
  ("EC9", "Handle",
    [More than 32 characters, e.g. `a` repeated 33 times.], false),
  ("EC10", "Handle",
    [Already used by another account, e.g. `player_one`.], false),
  ("EC11", "Handle",
    [25–32 characters using only allowed characters, e.g.
    `aaaaaaaaaabbbbbbbbbbcccccccccc` (30 characters).
    The page states that 3–32 characters are allowed,
    but the system keeps only the first 24.], true),
)

=== Test Cases
#test-cases(
  story: "US 1-4",
  (
    id: "TC12",
    action: [Enter the display name and handle, then save the profile.],
    inputs: ("Display Name": [`Imtaez`], Handle: [`imtaez_01`]),
    expected: [
      The system saves the profile and shows `Imtaez` and `imtaez_01` on the page.

      #image("../img/tc12.png", width: 100%)
    ],
    actual: [
      After saving, the page showed "Profile updated". The Account card showed
      `Imtaez` and `imtaez_01`, and the navigation bar showed "Imtaez".
    ],
    ecs: "EC1, EC6",
    status: "P",
  ),
  (
    id: "TC13",
    action: [Enter the display name and handle, then save the profile.],
    inputs: ("Display Name": [`  Dec48  `], Handle: [`Dec_48`]),
    expected: [
      The system saves the profile and shows `Dec48` (spaces removed)
      and `dec_48` (converted to lowercase).

      #image("../img/tc13-1.png", width: 100%)
      #image("../img/tc13-2.png", width: 100%)
    ],
    actual: [
      The handle field converted `D` to lowercase while typing.
      After saving, the page showed "Profile updated", the display name was
      `Dec48` (spaces removed) and the handle was `dec_48`.
      The navigation bar showed "Dec48".
    ],
    ecs: "EC2, EC6",
    status: "P",
  ),
  (
    id: "TC14",
    action: [Clear the display name, then save the profile.],
    inputs: ("Display Name": [`""` (empty)], Handle: [`dec_48`]),
    expected: [
      The system shows an error under Display Name and does not save the profile.

      #image("../img/tc14.png", width: 100%)
    ],
    actual: [
      After pressing Save, the message "Display name cannot be empty" appeared
      under the field. The navigation bar still showed "Dec48".
    ],
    ecs: "EC3",
    status: "P",
  ),
  (
    id: "TC15",
    action: [Enter spaces only as the display name, then save the profile.],
    inputs: ("Display Name": [`"   "` (spaces only)], Handle: [`dec_48`]),
    expected: [
      The system shows an error under Display Name and does not save the profile.

      #image("../img/tc15-1.png", width: 100%)
      #image("../img/tc15-2.png", width: 100%)
    ],
    actual: [
      After pressing Save, the field was emptied and the message
      "Display name cannot be empty" appeared. The navigation bar
      still showed "Dec48".
    ],
    ecs: "EC4",
    status: "P",
  ),
  (
    id: "TC16",
    action: [Try to type a 65-character display name.],
    inputs: ("Display Name": [`a` repeated 65 times], Handle: [`dec_48`]),
    expected: [
      The system does not accept more than 64 characters:
      the extra character cannot be entered, or an error is shown.

      #image("../img/tc16-1.png", width: 100%)
      #image("../img/tc16-2.png", width: 100%)
    ],
    actual: [
      The field stopped accepting characters after the 64th.
      The 65th `a` could not be typed.
      The browser console showed `maxLength = 64` for the display name input.
    ],
    ecs: "EC5",
    status: "P",
  ),
  (
    id: "TC17",
    action: [Try to type a handle that contains a hyphen.],
    inputs: ("Display Name": [`Imtaez`], Handle: [`imtaez-01`]),
    expected: [
      The system does not accept the hyphen:
      the character cannot be entered, or an error is shown.

      #image("../img/tc17.png", width: 100%)
    ],
    actual: [
      The hyphen could not be typed. The field showed `imtaez01`.
    ],
    ecs: "EC7",
    status: "P",
  ),
  (
    id: "TC18",
    action: [Enter a 2-character handle, then save the profile.],
    inputs: ("Display Name": [`Dec48`], Handle: [`de`]),
    expected: [
      The system shows an error under Handle and does not save the profile.

      #image("../img/tc18.png", width: 100%)
    ],
    actual: [
      After pressing Save, the message "Handle must be at least 3 characters"
      appeared under the field. The navigation bar still showed "Dec48".
    ],
    ecs: "EC8",
    status: "P",
  ),
  (
    id: "TC19",
    action: [Try to type a 33-character handle.],
    inputs: ("Display Name": [`Imtaez`], Handle: [`a` repeated 33 times]),
    expected: [
      The system does not accept more than 32 characters:
      the extra character cannot be entered, or an error is shown.

      #image("../img/tc19-1.png", width: 100%)
      #image("../img/tc19-2.png", width: 100%)
    ],
    actual: [
      The field stopped accepting characters after the 32nd.
      The 33rd `a` could not be typed.
      The browser console showed `maxLength = 32` for the handle input.
    ],
    ecs: "EC9",
    status: "P",
  ),
  (
    id: "TC20",
    action: [Enter a handle that another account already uses, then save the profile.],
    inputs: ("Display Name": [`Dec48`], Handle: [`player_one`]),
    expected: [
      The system shows an error that the handle is already taken
      and does not save the profile.

      #image("../img/tc20.png", width: 100%)
    ],
    actual: [
      After pressing Save, the message "Handle is already taken. Please choose
      another one." appeared under the field. The navigation bar still showed "Dec48".
    ],
    ecs: "EC10",
    status: "P",
  ),
  (
    id: "TC21",
    action: [Enter a 30-character handle, then save the profile.],
    inputs: ("Display Name": [`Dec48`], Handle: [`aaaaaaaaaabbbbbbbbbbcccccccccc`]),
    expected: [
      The system saves the handle as entered (30 characters), because the page
      states that 3–32 characters are allowed.

      #image("../img/tc21-1.png", width: 100%)
      #image("../img/tc21-2.png", width: 100%)
    ],
    actual: [
      The page showed "Profile updated" without any warning, but the handle
      was shortened to `aaaaaaaaaabbbbbbbbbbcccc` (24 characters).
      The last 6 characters were removed.
    ],
    ecs: "EC11",
    status: "F",
  ),
)
