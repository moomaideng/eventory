# ADR-03: Authentication Pattern

## Context

Users need one login, then Organizer / Sponsor / Competitor workspaces. Eventory's value is tournament hosting, sponsorship, and registration, not building a custom login system. Implementing email/password ourselves would mean owning password hashing, reset flows, and multi-factor auth, which adds security work without improving the core product. Eventory services should not store passwords. The browser should use a standard OAuth-style flow, and APIs should authorize with a short-lived token after identity is established.

Options considered:

- **Supabase Auth (external IdP, Google OAuth)**: managed identity and JWT for APIs; depends on Supabase availability.
- **Custom email/password**: full control, but we own credential security and reset flows.
- **Custom OAuth against Google only**: possible, but we would build and maintain the OAuth integration ourselves instead of using a managed IdP.

## Decision

Use **Supabase Auth** as the external identity provider (Google OAuth). Eventory services validate JWTs (JWKS) and do **not** store passwords. Account Service owns application profiles (organizer/sponsor) after identity exists. Mode switch between Competitor / Organizer / Sponsor is client workspace navigation, not a separate auth provider.

## Status

Accepted

## Consequences

**Positive**

- No password storage or reset flows inside Eventory services.
- Less auth security surface for the team to build and maintain.
- Clear boundary: identity external, domain profiles in Account Service.

**Negative**

- Sign-in depends on Supabase (and Google) availability.
- Users without a Google account cannot sign in unless we add another login option later.
