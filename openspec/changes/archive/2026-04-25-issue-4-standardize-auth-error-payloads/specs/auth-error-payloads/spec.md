# Auth Error Payloads Specification

## Purpose

Define a canonical auth error payload contract so auth failures are machine-readable and consistent across middleware and auth-scoped endpoints.

## Requirements

### Requirement: Canonical Auth Error Envelope

The system MUST return auth failures as a JSON object containing exactly `code` and `message` fields.

#### Scenario: Middleware auth failure uses canonical envelope

- GIVEN an auth request evaluated by middleware
- WHEN the request is rejected for authentication reasons
- THEN the response body MUST include `code` and `message`
- AND the response body MUST NOT include legacy `error` as the contract field

### Requirement: Auth Error Codes Are Stable and Branch-Specific

The system SHALL use stable, branch-specific auth error codes so clients can reliably branch on failure type across releases.

#### Scenario: Missing token maps to stable code

- GIVEN a protected route request with no token
- WHEN middleware rejects the request
- THEN status MUST be `401`
- AND `code` SHALL be `auth_token_missing`

#### Scenario: Invalid token maps to stable code

- GIVEN a protected route request with an invalid token
- WHEN middleware rejects the request
- THEN status MUST be `401`
- AND `code` SHALL be `auth_token_invalid`

### Requirement: Messages Are Client-Safe and Actionable

The system MUST provide a client-safe `message` suitable for user/client handling and MUST NOT expose internal stack or provider internals in auth error responses.

#### Scenario: Message omits internal details

- GIVEN an auth-scoped endpoint returns an auth failure
- WHEN the response is serialized
- THEN `message` MUST describe the failure class in user-safe terms
- AND `message` MUST NOT include internal stack traces, secrets, or provider internals
