# ADR-03: Authentication Pattern

## Context

Users need one login, then Organizer / Sponsor / Competitor workspaces. Building and securing email/password (hashing, reset, MFA) is not the product's differentiator. The team already uses Supabase in the SE stack for Google sign-in and JWT verification on the API.

ASRs: no password store in Eventory services; browser client must use a standard OAuth-style flow; APIs must authorize with a short-lived token after identity is established.

Options considered:

- **Supabase Auth (external IdP, Google OAuth)**: managed identity, JWT for APIs, fits current client; depends on Supabase availability.
- **Custom email/password**: full control, but we own credential security and reset flows.
- **Home-grown OAuth against Google only**: possible, but reimplements what Supabase already provides to the project.

## Decision

Use **Supabase Auth** as the external identity provider (Google OAuth). Eventory services validate JWTs (JWKS) and do **not** store passwords. Account Service owns application profiles (organizer/sponsor) after identity exists. Mode switch between Competitor / Organizer / Sponsor is client workspace navigation, not a separate auth provider.

## Status

Accepted

## Consequences

**Positive**

- No password storage or reset flows inside Eventory services.
- Reuses the existing SE auth path; less new security surface for the team.
- Clear boundary: identity external, domain profiles in Account Service.

**Negative**

- Sign-in depends on Supabase (and Google) availability.
- Users without a supported IdP path cannot register until another provider is added.
- Token revoke/refresh edge cases must be handled at the gateway and services.
