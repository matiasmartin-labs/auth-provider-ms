# Delta for Google OAuth Callback

## MODIFIED Requirements

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
- WHEN tests execute after auth error payload standardization
- THEN successful callback/login behavior MUST remain unchanged
- AND success-path assertions SHALL continue to pass
