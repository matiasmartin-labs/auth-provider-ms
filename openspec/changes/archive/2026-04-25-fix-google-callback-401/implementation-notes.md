# Implementation Notes: fix-google-callback-401

## Summary

Aligned Google callback disallowed-email handling from `403 Forbidden` to `401 Unauthorized` while preserving the existing payload contract (`{"error":"email is not allowed"}`) and keeping early-return behavior.

## Code Changes

- `internal/infrastructure/port/in/google/routes.go`
  - Changed disallowed-email branch status code: `http.StatusForbidden` → `http.StatusUnauthorized`.
- `internal/infrastructure/port/in/google/routes_test.go`
  - Updated disallowed-email assertions to `401`.
  - Refactored disallowed-email test into table-driven subtests (`without redirect`, `with redirect configured`).
  - Added explicit assertions that no cookie is issued and no redirect location is set for disallowed-email responses.

## Verification Executed

- Focused disallowed-email test:
  - `go test ./internal/infrastructure/port/in/google -run TestGoogleCallbackHandler_EmailNotAllowed`
- Error-branch regression checks:
  - `go test ./internal/infrastructure/port/in/google -run "TestGoogleCallbackHandler_(InvalidState|MissingCode|ProviderError|TokenGenerationError)"`
- Success-flow regression checks:
  - `go test ./internal/infrastructure/port/in/google -run TestGoogleCallbackHandler_Success_NoRedirect`
  - `go test ./internal/infrastructure/port/in/google -run TestGoogleCallbackHandler_Success_WithRedirect`
- Package regression:
  - `go test ./internal/infrastructure/port/in/google`
- Full suite:
  - `go test ./...`

All commands passed after implementation.

## Contract Check

Manual check against:
- `openspec/changes/fix-google-callback-401/spec.md`
- `openspec/changes/fix-google-callback-401/design.md`

Result: implementation matches required behavior (`401` + unchanged payload) and preserves successful callback flows.

## Post-Verify Blocker Remediation

Verifier reported a project-level strict gate failure in `go vet ./...`:

- `pkg/application_test.go:105:19: call of assert.NotNil copies lock value`

Minimal, non-behavioral fix applied in tests only:

- `pkg/application_test.go`
  - Replaced `assert.NotNil(t, app.Server)` with `assert.NotNil(t, app.Server.Handler)` to avoid copying `http.Server` (which embeds noCopy locks).
  - Kept all original semantic checks (`Addr`, timeouts, header bytes) unchanged.

Validation after fix:

- `go test ./pkg -run "TestApplication_(UseServer|ChainedMethodCalls)"` ✅
- `go vet ./...` ✅
- `go test ./...` ✅
