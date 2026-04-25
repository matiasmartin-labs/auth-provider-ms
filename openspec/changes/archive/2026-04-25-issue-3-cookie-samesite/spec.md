# Delta for google-oauth-callback

## ADDED Requirements

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
