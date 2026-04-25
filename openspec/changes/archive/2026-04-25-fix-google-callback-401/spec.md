# Google OAuth Callback Specification

## Purpose

Define callback response behavior for allowed and disallowed Google-authenticated emails, aligning disallowed-email handling to `401 Unauthorized` while preserving successful login behavior.

## Requirements

### Requirement: Disallowed Email Callback Returns Unauthorized

The system MUST return HTTP `401 Unauthorized` when the Google OAuth callback resolves a user whose email is not allowed by configured policy.

#### Scenario: Disallowed email in callback

- GIVEN a valid callback request (`state` and `code`) and provider user info with a disallowed email
- WHEN the callback handler evaluates allowlist policy
- THEN the response status SHALL be `401`
- AND the login session/token issuance MUST NOT proceed

### Requirement: Unauthorized Payload Remains Clear and Actionable

The system MUST preserve a clear, actionable response payload for disallowed-email callback responses; the payload semantics SHOULD remain unchanged except for status alignment to `401`.

#### Scenario: Payload clarity for disallowed email

- GIVEN a callback request that reaches the disallowed-email branch
- WHEN the handler returns unauthorized
- THEN the response payload MUST include the same explicit disallowed-email meaning used previously
- AND clients SHALL be able to identify the response as an authorization failure without ambiguity

### Requirement: Callback Test Suite Reflects Unauthorized Contract

The callback test suite MUST assert `401` for disallowed-email outcomes and MUST continue to cover both success and error branches.

#### Scenario: Test expectation updated for disallowed email

- GIVEN callback tests for disallowed-email handling
- WHEN tests execute against current callback behavior
- THEN expected status MUST be `401`
- AND assertions for payload clarity SHALL remain validated

#### Scenario: Success flow regression guard remains intact

- GIVEN callback tests for allowed-email login flow
- WHEN tests execute after unauthorized-status alignment
- THEN successful callback/login behavior MUST remain unchanged
- AND success-path assertions SHALL continue to pass
