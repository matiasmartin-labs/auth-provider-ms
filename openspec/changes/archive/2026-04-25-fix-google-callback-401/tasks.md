# Tasks: Align Google Callback Disallowed Email to 401

## Phase 1: Foundation

- [x] 1.1 Inspect `internal/infrastructure/port/in/google/routes.go` and confirm the disallowed-email branch currently returns `http.StatusForbidden` with `{"error":"email is not allowed"}`.
- [x] 1.2 Review `internal/infrastructure/port/in/google/routes_test.go` callback coverage and identify cases to keep explicit for success + error branches (per project testing standard).

## Phase 2: Core Implementation

- [x] 2.1 Update `internal/infrastructure/port/in/google/routes.go` to return `http.StatusUnauthorized` in the `!userInfo.IsEmailAllowed()` branch.
- [x] 2.2 Preserve existing unauthorized payload semantics in `internal/infrastructure/port/in/google/routes.go` (`{"error":"email is not allowed"}`) and keep early-return behavior so token/cookie issuance does not execute.

## Phase 3: Test Updates

- [x] 3.1 Update disallowed-email callback assertion(s) in `internal/infrastructure/port/in/google/routes_test.go` to expect `401` and unchanged error payload.
- [x] 3.2 Keep or refactor callback tests in `internal/infrastructure/port/in/google/routes_test.go` into table-driven form if multiple cases are touched, ensuring explicit coverage for allowed/disallowed outcomes.
- [x] 3.3 Re-assert existing error branches in `internal/infrastructure/port/in/google/routes_test.go` (invalid state, missing code, provider/token failures) remain unchanged.

## Phase 4: Verification

- [x] 4.1 Run focused callback tests for `internal/infrastructure/port/in/google/routes_test.go` and verify disallowed-email case now returns `401`.
- [x] 4.2 Run package-level regression tests for `internal/infrastructure/port/in/google` to validate success paths (`TestGoogleCallbackHandler_Success_NoRedirect`, `TestGoogleCallbackHandler_Success_WithRedirect`) still pass.
- [x] 4.3 Perform a manual contract check against `openspec/changes/fix-google-callback-401/spec.md` and `openspec/changes/fix-google-callback-401/design.md` to confirm implementation matches required status and payload semantics.
