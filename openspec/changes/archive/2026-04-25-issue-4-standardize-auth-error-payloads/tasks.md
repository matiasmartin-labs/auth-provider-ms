# Tasks: Standardize Auth Error Payloads

## Phase 1: Contract Foundation

- [x] 1.1 Create `pkg/auth_error.go` with `AuthErrorResponse` and `WriteAuthError(ctx, status, code, message)` that always serializes `code` and `message` only.
- [x] 1.2 Define and export stable auth code constants in `pkg/auth_error.go` using spec-aligned snake_case values (`auth_token_missing`, `auth_token_invalid`, callback/auth branch codes).
- [x] 1.3 Add `pkg/auth_error_test.go` table-driven tests validating status, exact JSON keys (`code`,`message`), and absence of legacy `error` field.

## Phase 2: Middleware and Me Endpoint Alignment

- [x] 2.1 Refactor `pkg/security-middleware.go` unauthorized branches (missing token, invalid/expired token) to call `WriteAuthError` with stable codes and client-safe messages.
- [x] 2.2 Update `pkg/security-middleware_test.go` to decode JSON and assert exact status/code/message contract for rejected auth requests.
- [x] 2.3 Confirm `internal/infrastructure/port/in/me/routes.go` scope from issue/design open question; if in scope, replace auth-related failure payloads with shared auth envelope and stable codes.
- [x] 2.4 If 2.3 is in scope, update `internal/infrastructure/port/in/me/routes_test.go` with contract assertions for claims-missing/invalid branches.

## Phase 3: Google Callback Standardization

- [x] 3.1 Refactor `internal/infrastructure/port/in/google/routes.go` invalid callback input branches (state mismatch, missing code) to emit standardized auth envelope via helper.
- [x] 3.2 Refactor callback auth failure branches in `internal/infrastructure/port/in/google/routes.go` (disallowed email, provider failure, token generation failure) to mapped stable codes and client-safe messages.
- [x] 3.3 Preserve success behavior in `internal/infrastructure/port/in/google/routes.go` (cookie set, redirect/no-redirect, token success payload) unchanged.

## Phase 4: Integration Verification

- [x] 4.1 Update `internal/infrastructure/port/in/google/routes_test.go` using named table-driven subtests to assert status + exact `code`/`message` across auth failure branches.
- [x] 4.2 Add/adjust assertions in middleware and callback tests that legacy `error` contract field is not returned for standardized auth failures.
- [x] 4.3 Run targeted regression tests: `go test ./pkg ./internal/infrastructure/port/in/google ./internal/infrastructure/port/in/me` and fix any contract mismatches.

## Phase 5: Final Consistency Pass

- [x] 5.1 Review all in-scope auth failure responses to ensure internal/provider details are not leaked in `message` values.
- [x] 5.2 Confirm implementation matches spec scenarios in `openspec/changes/issue-4-standardize-auth-error-payloads/specs/**` before handing off to `sdd-apply`.
