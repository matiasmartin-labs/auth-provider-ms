# Tasks: Apply configured SameSite on Google callback token cookie

## Phase 1: Foundation

- [x] 1.1 Add SameSite parsing helper in `internal/infrastructure/port/in/google/routes.go` (e.g., `parseSameSite(raw string) http.SameSite`) using trimmed, case-insensitive mapping for `Strict|Lax|None`; default to zero value for empty/invalid.
- [x] 1.2 Refactor callback cookie emission in `GoogleCallbackHandler` (`internal/infrastructure/port/in/google/routes.go`) to use `http.SetCookie` with explicit `http.Cookie` fields and parsed SameSite, preserving existing name/path/value/MaxAge/HttpOnly/Secure behavior.

## Phase 2: Core implementation wiring

- [x] 2.1 Ensure callback success flow still follows existing redirect/JSON behavior after cookie write changes in `internal/infrastructure/port/in/google/routes.go`.
- [x] 2.2 Keep sign-out and non-callback cookie code paths unchanged in `internal/infrastructure/port/in/google/routes.go` (no scope creep to shared cookie abstractions).

## Phase 3: Verification tests

- [x] 3.1 Add table-driven subtests in `internal/infrastructure/port/in/google/routes_test.go` for callback success asserting `Set-Cookie` contains `SameSite=Strict`, `SameSite=Lax`, `SameSite=None` when config values use mixed casing.
- [x] 3.2 Add table-driven subtests in `internal/infrastructure/port/in/google/routes_test.go` for empty and invalid `security.cookie.same-site` asserting `Set-Cookie` omits any `SameSite` attribute.
- [x] 3.3 Add regression assertions in callback success tests for `Secure`, `HttpOnly`, and `Max-Age` to match pre-change behavior for both valid and fallback SameSite paths.
- [x] 3.4 Re-run existing callback branch tests in `internal/infrastructure/port/in/google/routes_test.go` (invalid state, missing code, provider error, token error, email not allowed, redirect/no-redirect success) to confirm no functional regressions.

## Phase 4: Final validation

- [x] 4.1 Run `go test ./internal/infrastructure/port/in/google -v` and ensure all new/existing route tests pass.
- [x] 4.2 Run `go test ./...` to verify repo-wide behavior remains stable after cookie handling change.
