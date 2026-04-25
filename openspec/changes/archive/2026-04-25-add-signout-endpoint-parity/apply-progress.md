# Apply Progress: add-signout-endpoint-parity

## Mode

Strict TDD

## Completed Tasks

- [x] 1.1 Update `pkg/application.go` to add `RegisterPOST(path string, handler gin.HandlerFunc)` mirroring existing `RegisterGET` style.
- [x] 1.2 Extend `pkg/application_test.go` with `TestApplication_RegisterPOST` that registers a POST route and asserts handler execution (`200` + expected response body).
- [x] 2.1 Create `internal/infrastructure/port/in/signout/routes.go` with `SignOutHandler(*gin.Context)` that always sets `token` cookie clear semantics (`""`, `Max-Age=0`, `Path=/`, `HttpOnly=true`, `Secure` from config) and returns `204` with empty body.
- [x] 2.2 Keep sign-out handler adapter-only (no domain call, no request payload parsing) so repeated calls stay idempotent and contract-consistent.
- [x] 3.1 Modify `internal/infrastructure/port/in/server/server.go` to register `POST /api/v1/auth/sign-out` using `app.RegisterPOST` and `signout.SignOutHandler`.
- [x] 3.2 Add `internal/infrastructure/port/in/server/server_test.go` route wiring coverage that executes `Routes(app)` and verifies `POST /api/v1/auth/sign-out` is reachable (not `404`).
- [x] 4.1 Create `internal/infrastructure/port/in/signout/routes_test.go` table-driven tests for cookie `Secure` enabled/disabled; assert `204`, empty body, and `token` cookie attributes (`Value=""`, `Path=/`, `MaxAge=0`, `HttpOnly=true`, `Secure` matches config).
- [x] 4.2 In `internal/infrastructure/port/in/signout/routes_test.go`, add repeated-call test (same client sends sign-out twice) asserting both responses are `204` and both include cookie-clear contract.
- [x] 4.3 Run `go test ./pkg ./internal/infrastructure/port/in/signout ./internal/infrastructure/port/in/server` and then `go test ./...` to validate parity behavior and avoid regressions.

## TDD Cycle Evidence

| Task | Test File | Layer | Safety Net | RED | GREEN | TRIANGULATE | REFACTOR |
|------|-----------|-------|------------|-----|-------|-------------|----------|
| 1.1-1.2 | `pkg/application_test.go` | Unit | ✅ `go test ./pkg -run 'TestApplication_RegisterGET\|TestApplication_RegisterProtectedGET'` | ✅ `TestApplication_RegisterPOST` added first (failed: missing `RegisterPOST`) | ✅ `go test ./pkg -run TestApplication_RegisterPOST` | ✅ Added GET-not-registered case to force POST-only behavior path | ✅ `gofmt` and focused helper-style parity |
| 2.1-2.2, 4.1-4.2 | `internal/infrastructure/port/in/signout/routes_test.go` | Unit | N/A (new package) | ✅ Added signout tests first (failed: undefined `SignOutHandler`) | ✅ `go test ./internal/infrastructure/port/in/signout -run 'TestSignOutHandler_ClearsTokenCookieAndReturnsNoContent\|TestSignOutHandler_IsIdempotentAcrossRepeatedCalls'` | ✅ Table-driven `Secure` on/off + repeated-call idempotency scenario | ✅ Extracted `findTokenCookie` helper and formatted |
| 3.1-3.2 | `internal/infrastructure/port/in/server/server_test.go` | Integration | ✅ `go test ./internal/infrastructure/port/in/server` (baseline: no tests, no failures) | ✅ Added route wiring test first (failed 404 before route registration) | ✅ `go test ./internal/infrastructure/port/in/server -run TestRoutes_RegistersSignOutEndpoint` | ➖ Single behavior scenario in spec for wiring reachability | ✅ `gofmt` and minimal fixture setup |
| 4.3 | N/A (verification command task) | Verification | N/A | ➖ Command-only task | ✅ `go test ./pkg ./internal/infrastructure/port/in/signout ./internal/infrastructure/port/in/server` and `go test ./...` | ➖ Command-only task | ➖ None needed |

## Test Summary

- **Total tests written**: 4
- **Total tests passing**: 4 new tests + full suite passing
- **Layers used**: Unit (3 tests), Integration (1 test), E2E (0)
- **Approval tests**: None — no refactor-only task
- **Pure functions created**: 1 (`resolveCookieSecure`)

## Files Changed

- `pkg/application.go` — added `RegisterPOST` helper.
- `pkg/application_test.go` — added TDD coverage for POST registration including method-path differentiation.
- `internal/infrastructure/port/in/signout/routes.go` — added adapter-only sign-out handler returning `204` and clearing cookie.
- `internal/infrastructure/port/in/signout/routes_test.go` — added table-driven cookie contract tests and repeated-call idempotency test.
- `internal/infrastructure/port/in/server/server.go` — wired `POST /api/v1/auth/sign-out` route.
- `internal/infrastructure/port/in/server/server_test.go` — added route reachability integration test.
- `openspec/changes/add-signout-endpoint-parity/tasks.md` — marked completed tasks `[x]`.

## Deviations

None — implementation matches spec/design.

## Issues

None.
