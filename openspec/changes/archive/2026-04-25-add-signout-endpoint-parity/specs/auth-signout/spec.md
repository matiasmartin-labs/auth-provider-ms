# auth-signout Specification

## Purpose

Define server sign-out behavior parity for `POST /api/v1/auth/sign-out`, including the normative cookie-clearing contract for the auth token cookie.

## Requirements

### Requirement: Sign-out endpoint availability and response

The system MUST expose `POST /api/v1/auth/sign-out` and MUST return `204 No Content` when invoked.

The system MUST treat sign-out as safe to repeat and MUST preserve the same response contract on repeated calls.

#### Scenario: Sign-out endpoint is reachable

- GIVEN the HTTP server is running
- WHEN a client sends `POST /api/v1/auth/sign-out`
- THEN the response status is `204 No Content`
- AND the response body is empty

#### Scenario: Repeated sign-out remains contract-consistent

- GIVEN a client has already called sign-out
- WHEN the same client sends `POST /api/v1/auth/sign-out` again
- THEN the response status is `204 No Content`
- AND the cookie-clearing contract is still applied

### Requirement: Token cookie-clearing contract

The system MUST clear the `token` cookie in the sign-out response by setting a cookie with:
- `Name=token`
- `Value=""`
- `Path=/`
- `Max-Age=0` (delete semantics)
- `HttpOnly=true`

The system MUST set the `Secure` attribute according to current runtime security configuration.

The system SHALL NOT introduce new SameSite or Domain behavior as part of this capability.

#### Scenario: Sign-out clears token cookie with delete semantics

- GIVEN a client sends `POST /api/v1/auth/sign-out`
- WHEN the server builds the response headers
- THEN `Set-Cookie` includes `token=`, `Path=/`, `Max-Age=0`, and `HttpOnly`
- AND the cookie represents deletion semantics for `token`

#### Scenario: Secure attribute follows configuration

- GIVEN runtime cookie security configuration is enabled or disabled
- WHEN a client sends `POST /api/v1/auth/sign-out`
- THEN the sign-out `Set-Cookie` includes `Secure` when enabled
- AND omits `Secure` when disabled
