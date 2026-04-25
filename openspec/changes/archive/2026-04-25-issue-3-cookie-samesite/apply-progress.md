# Apply Progress: issue-3-cookie-samesite

## Mode

Strict TDD

## Completed Tasks

- [x] 1.1 Add SameSite parsing helper in `internal/infrastructure/port/in/google/routes.go` using trimmed, case-insensitive mapping for `Strict|Lax|None` with zero-value fallback.
- [x] 1.2 Refactor callback cookie emission in `GoogleCallbackHandler` to use `http.SetCookie` with explicit `http.Cookie` fields and parsed SameSite while preserving name/path/value/MaxAge/HttpOnly/Secure behavior.
- [x] 2.1 Ensure callback success flow still follows existing redirect/JSON behavior after cookie write changes.
- [x] 2.2 Keep sign-out and non-callback cookie code paths unchanged (no shared abstraction scope creep).
- [x] 3.1 Add table-driven subtests for callback success asserting `SameSite=Strict|Lax|None` from mixed-cased config values.
- [x] 3.2 Add table-driven subtests for empty and invalid `security.cookie.same-site` asserting `Set-Cookie` omits `SameSite`.
- [x] 3.3 Add regression assertions for `Secure`, `HttpOnly`, and `Max-Age` for valid and fallback SameSite paths.
- [x] 3.4 Re-run existing callback branch tests (invalid state, missing code, provider error, token error, email not allowed, redirect/no-redirect success).
- [x] 4.1 Run `go test ./internal/infrastructure/port/in/google -v`.
- [x] 4.2 Run `go test ./...`.

## TDD Cycle Evidence

| Task | Test File | Layer | Safety Net | RED | GREEN | TRIANGULATE | REFACTOR |
|------|-----------|-------|------------|-----|-------|-------------|----------|
| 1.1 | `internal/infrastructure/port/in/google/routes_test.go` | Unit | ✅ `go test ./internal/infrastructure/port/in/google -count=1` baseline passing | ✅ `TestParseSameSite` written first; failed with `undefined: parseSameSite` | ✅ `go test ./internal/infrastructure/port/in/google -run TestParseSameSite -count=1` | ✅ 5 cases (`Strict`,`Lax`,`None`, empty, invalid) | ✅ `gofmt` after implementation |
| 1.2, 2.1, 2.2, 3.1, 3.2, 3.3 | `internal/infrastructure/port/in/google/routes_test.go` | Unit/Handler | ✅ Baseline already captured on same package | ✅ `TestGoogleCallbackHandler_Success_SameSiteMapping` added before cookie refactor; initial RED gated by missing parser/behavior | ✅ `go test ./internal/infrastructure/port/in/google -run TestGoogleCallbackHandler_Success_SameSiteMapping -count=1` | ✅ Table-driven with 5 scenarios incl. fallback paths; assertions include cookie security attrs | ✅ kept patch handler-local and formatted |
| 3.4 | Existing tests in `routes_test.go` | Unit/Handler regression | ✅ Existing tests already green before edits | ✅ N/A (verification task) | ✅ `go test ./internal/infrastructure/port/in/google -run 'TestGoogleCallbackHandler' -count=1` | ➖ Verification-only task | ➖ None needed |
| 4.1 | Package suite | Verification | N/A | ➖ Command-only task | ✅ `go test ./internal/infrastructure/port/in/google -v -count=1` | ➖ Command-only task | ➖ None needed |
| 4.2 | Repo suite | Verification | N/A | ➖ Command-only task | ✅ `go test ./... -count=1` | ➖ Command-only task | ➖ None needed |

## Test Summary

- **Total tests written**: 2 new tests (`TestParseSameSite`, `TestGoogleCallbackHandler_Success_SameSiteMapping`)
- **Total tests passing**: all tests in modified package and full repository suite
- **Layers used**: Unit/Handler
- **Approval tests**: None — no refactor-only behavior-preservation task
- **Pure functions created**: 1 (`parseSameSite`)

## Files Changed

- `internal/infrastructure/port/in/google/routes.go` — added `parseSameSite` and switched callback cookie write to explicit `http.SetCookie` with SameSite mapping.
- `internal/infrastructure/port/in/google/routes_test.go` — added table-driven SameSite parser and callback header tests; added configurable cookie setup helper to support scenarios.
- `openspec/changes/issue-3-cookie-samesite/tasks.md` — marked all tasks complete.

## Deviations

None — implementation matches design/spec and remains handler-local.

## Issues

None.
