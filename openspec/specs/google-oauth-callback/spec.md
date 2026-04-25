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

The system MUST return the standardized auth error envelope for auth-related callback failures, with clear and actionable client-safe semantics.
(Previously: Payload meaning stayed implicit and tied to legacy error-string semantics.)

#### Scenario: Disallowed email returns canonical auth envelope

- GIVEN a callback request that reaches the disallowed-email branch
- WHEN the handler returns unauthorized
- THEN the response status SHALL be `401`
- AND the payload MUST include `code=auth_email_not_allowed` and a client-safe `message`

#### Scenario: Callback auth failure payload remains unambiguous

- GIVEN a callback request that reaches an auth-related failure branch
- WHEN the handler returns an auth failure response
- THEN the payload MUST use `code` and `message` fields
- AND clients SHALL identify authorization failure type from `code` without parsing free text

### Requirement: Callback Test Suite Reflects Unauthorized Contract

The callback test suite MUST assert standardized auth error envelope behavior for auth-related callback failures while continuing to cover success and error branches.
(Previously: Tests asserted unauthorized status and payload clarity, but not stable envelope fields/codes.)

#### Scenario: Tests assert standardized callback auth error contract

- GIVEN callback tests for disallowed-email and other auth-related failure handling
- WHEN tests execute against current callback behavior
- THEN expected payload assertions MUST include `code` and `message`
- AND expected status/code pairs SHALL be validated for each covered branch

#### Scenario: Success flow regression guard remains intact

- GIVEN callback tests for allowed-email login flow
- WHEN tests execute after unauthorized-status alignment
- THEN successful callback/login behavior MUST remain unchanged
- AND success-path assertions SHALL continue to pass

### Requirement: Callback Token Cookie Applies Configured SameSite Modes

The system MUST map `security.cookie.same-site` to the callback token cookie `SameSite` attribute using case-insensitive values.

#### Scenario: SameSite Strict is applied

- GIVEN Google callback succeeds and `security.cookie.same-site` is `Strict` (any casing)
- WHEN the callback response sets the auth token cookie
- THEN the `Set-Cookie` header MUST include `SameSite=Strict`

#### Scenario: SameSite Lax is applied

- GIVEN Google callback succeeds and `security.cookie.same-site` is `Lax` (any casing)
- WHEN the callback response sets the auth token cookie
- THEN the `Set-Cookie` header MUST include `SameSite=Lax`

#### Scenario: SameSite None is applied

- GIVEN Google callback succeeds and `security.cookie.same-site` is `None` (any casing)
- WHEN the callback response sets the auth token cookie
- THEN the `Set-Cookie` header MUST include `SameSite=None`

### Requirement: Callback Token Cookie Has Deterministic SameSite Fallback

The system MUST omit the `SameSite` cookie attribute when `security.cookie.same-site` is empty or invalid.

#### Scenario: Empty SameSite config omits attribute

- GIVEN Google callback succeeds and `security.cookie.same-site` is empty
- WHEN the callback response sets the auth token cookie
- THEN the `Set-Cookie` header MUST NOT include any `SameSite` attribute

#### Scenario: Invalid SameSite config omits attribute

- GIVEN Google callback succeeds and `security.cookie.same-site` is an unsupported value
- WHEN the callback response sets the auth token cookie
- THEN the `Set-Cookie` header MUST NOT include any `SameSite` attribute

### Requirement: Existing Callback Cookie Security Attributes Are Preserved

The system SHALL keep callback token cookie `Secure`, `HttpOnly`, and `Max-Age` behavior unchanged while applying SameSite mapping and fallback behavior.

#### Scenario: Security attributes remain unchanged for valid SameSite

- GIVEN Google callback succeeds with `security.cookie.same-site` set to `Strict`, `Lax`, or `None`
- WHEN the callback response sets the auth token cookie
- THEN `Secure`, `HttpOnly`, and `Max-Age` MUST match their pre-change callback behavior

#### Scenario: Security attributes remain unchanged for fallback path

- GIVEN Google callback succeeds with empty or invalid `security.cookie.same-site`
- WHEN the callback response sets the auth token cookie without `SameSite`
- THEN `Secure`, `HttpOnly`, and `Max-Age` MUST match their pre-change callback behavior
