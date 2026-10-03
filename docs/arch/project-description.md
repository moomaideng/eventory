# Project Description

## Project Name

Eventory

## Group Members 

- Ashira Aungsumal 6732041921
- Kittichet Arayasujin 6732005321
- Kittichon Chaonawig 6731305421
- Pasin Thanyakasikol 6732025921
- Peeravas Piboolvorakul 6732033921

## Problem Description

TODO

## Target Customers

- TODO1
- TODO2
- TODO3

## Scenario (use-case & description)

TODO, use 3 usecases in project-proposal-draft, but below is just example. no need to follow it

### UC-01: TODO

**Primary Actor**: Customer (college student)

**Goal**: Record an expense with minimal manual typing by automatically picking up new bank receipt images from the device gallery instead of entering each transaction by hand.

**Preconditions**: Customer is signed in. Customer has at least one bank expense hook configured and has granted media library access for that hook.

**Main Flow**:

1. Customer completes a bank payment. The bank app saves a receipt image to the device gallery (where the OS and bank app allow).
2. System detects a newly added receipt image associated with the configured bank hook (media-library listener or polling via the native mobile app).
3. System uploads the image to object storage.
4. System extracts metadata (amount, merchant/counterparty, date) from the stored image.
5. System creates a pending expense record and notifies the customer with a suggested category.
6. Customer confirms the suggested category or picks a different one (`<<include>> Manage Category`).
7. System finalizes the expense record under the confirmed category.

**Postcondition**: A categorized expense record exists and is reflected in the dashboard.

**Outcome**: Their spending gets tracked accurately without manually re-typing every bank transaction.

**Alternate/Exceptional Flow:**

1. **User denies media access**: System explains the requirement and provides a shortcut to OS settings.
2. **Duplicate receipt detected**: System ignores duplicate images by hash to prevent double-counting.

### UC-02: TODO

### UC-03: TODO

### UC-04: TODO

## Functional Requirements


TODO: these are only example sections, don't need to follow it.

### TODO Section

- FR1: TODO
- FR2: TODO
- FR3: TODO

### TODO Section

- FR4: TODO
- FR5: TODO
- FR6: TODO
- FR7: TODO
- FR8: TODO

## Non-functional Requirements

TODO: these are only example sections, don't need to follow it.

### Security & Privacy

- NFR1: TODO

### Usability

- NFR2: TODO

### Reliability

- NFR3: TODO

### Performance

- NFR4: TODO
