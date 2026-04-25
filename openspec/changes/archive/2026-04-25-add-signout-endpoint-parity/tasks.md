# Tasks: Add Sign-Out Endpoint Parity

## Phase 1: Foundation

- [x] 1.1 Update `pkg/application.go` to add `RegisterPOST(path string, handler gin.HandlerFunc)` mirroring existing `RegisterGET` style.
- [x] 1.2 Extend `pkg/application_test.go` with `TestApplication_RegisterPOST` that registers a POST route and asserts handler execution (`200` + expected response body).

## Phase 2: Sign-out Endpoint Implementation

- [x] 2.1 Create `internal/infrastructure/port/in/signout/routes.go` with `SignOutHandler(*gin.Context)` that always sets `token` cookie clear semantics (`""`, `Max-Age=0`, `Path=/`, `HttpOnly=true`, `Secure` from config) and returns `204` with empty body.
- [x] 2.2 Keep sign-out handler adapter-only (no domain call, no request payload parsing) so repeated calls stay idempotent and contract-consistent.

## Phase 3: HTTP Wiring

- [x] 3.1 Modify `internal/infrastructure/port/in/server/server.go` to register `POST /api/v1/auth/sign-out` using `app.RegisterPOST` and `signout.SignOutHandler`.
- [x] 3.2 Add `internal/infrastructure/port/in/server/server_test.go` route wiring coverage that executes `Routes(app)` and verifies `POST /api/v1/auth/sign-out` is reachable (not `404`).

## Phase 4: Testing and Verification

- [x] 4.1 Create `internal/infrastructure/port/in/signout/routes_test.go` table-driven tests for cookie `Secure` enabled/disabled; assert `204`, empty body, and `token` cookie attributes (`Value=""`, `Path=/`, `MaxAge=0`, `HttpOnly=true`, `Secure` matches config).
- [x] 4.2 In `internal/infrastructure/port/in/signout/routes_test.go`, add repeated-call test (same client sends sign-out twice) asserting both responses are `204` and both include cookie-clear contract.
- [x] 4.3 Run `go test ./pkg ./internal/infrastructure/port/in/signout ./internal/infrastructure/port/in/server` and then `go test ./...` to validate parity behavior and avoid regressions.
