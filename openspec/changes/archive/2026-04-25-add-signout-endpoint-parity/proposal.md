# Proposal: Add Sign-Out Endpoint Parity

## Intent

Close Java/Go parity gap by adding `POST /api/v1/auth/sign-out` so clients can trigger server-driven logout and receive a consistent cookie-clearing response.

## Scope

### In Scope
- Add POST route registration support in the application wrapper used by HTTP adapters.
- Add `POST /api/v1/auth/sign-out` route in server wiring and implement a focused sign-out handler.
- Return `204 No Content` and clear `token` cookie with a consistent contract.
- Add tests for route reachability, status code, and cookie-clearing behavior.

### Out of Scope
- Changes to login/token issuance flow.
- Introducing or changing SameSite/domain cookie behavior beyond current project defaults.
- Frontend/client logout workflow changes.

## Capabilities

### New Capabilities
- `auth-signout`: Server endpoint to invalidate session cookie semantics via response cookie expiration.

### Modified Capabilities
- None.

## Approach

- Extend `pkg/application.go` with POST registration helper(s) consistent with current GET abstractions.
- Wire `POST /api/v1/auth/sign-out` in `internal/infrastructure/port/in/server/server.go`.
- Implement sign-out handler under inbound HTTP adapter layer.
- **Cookie decision (normative):** clear cookie via `SetCookie("token", "", 0, "/", "", secureFromConfig, true)`.
  - `HttpOnly` is always `true` for sign-out response (explicit parity/security contract).
  - `Secure` follows existing config behavior.
  - No new SameSite behavior is introduced in this change.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `pkg/application.go` | Modified | Add POST route helper(s). |
| `internal/infrastructure/port/in/server/server.go` | Modified | Register `/api/v1/auth/sign-out`. |
| `internal/infrastructure/port/in/**/routes*.go` | New/Modified | Add sign-out handler and route wiring. |
| `internal/infrastructure/port/in/**/*_test.go` | New/Modified | Add sign-out contract tests. |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Cookie-attribute mismatch vs expectations | Medium | Lock contract in spec/tests (status + cookie attributes). |
| Wrapper change affects route conventions | Low | Keep helper shape aligned with existing GET helpers and test wiring. |

## Rollback Plan

Revert POST helper and sign-out route/handler files; remove associated tests; redeploy previous build with unchanged auth behavior.

## Dependencies

- Existing Gin context cookie API and current security config loading.

## Success Criteria

- [ ] `POST /api/v1/auth/sign-out` is reachable in router.
- [ ] Endpoint returns `204 No Content`.
- [ ] Response clears `token` cookie (`Max-Age=0`, `Path=/`, `HttpOnly=true`, `Secure` from config).
- [ ] Automated tests verify status and cookie-clearing contract.
