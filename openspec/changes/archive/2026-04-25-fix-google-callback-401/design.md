# Design: Align Google Callback Disallowed Email to 401

## Technical Approach

Implement a minimal, handler-local behavior change in `GoogleCallbackHandler`: when `userInfo.IsEmailAllowed()` is false, return `401 Unauthorized` instead of `403 Forbidden`, while preserving the same JSON payload (`{"error":"email is not allowed"}`).

This matches the approved proposal and avoids refactoring route wiring, provider adapters, or domain allowlist logic. The design relies on existing callback flow and updates tests to assert the new status without altering successful authentication behavior.

## Architecture Decisions

| Option | Tradeoff | Decision |
|---|---|---|
| Change status code inline in callback handler | Small targeted diff; keeps current error handling style; no shared abstraction | ✅ Chosen |
| Introduce centralized auth error mapper/middleware | Cleaner long-term semantics, but out of scope and higher blast radius | ❌ Rejected |
| Change allowlist semantics in domain layer | Could mix authorization semantics with domain logic; unnecessary for this issue | ❌ Rejected |

## Data Flow

No flow shape changes; only one response code changes in an existing branch.

`GET /login/oauth2/code/google`

1. Validate `state` and `code` query params.
2. Fetch user info via `ProviderRepository.GetUserInfo`.
3. Check `userInfo.IsEmailAllowed()`.
4. If false: return `401` + `{"error":"email is not allowed"}`.
5. If true: generate token, set cookie, then redirect or return `200` token JSON.

```
Callback Route -> ProviderRepository -> UserInfo
                                  -> IsEmailAllowed?
                                        ├─ no  -> 401 JSON error
                                        └─ yes -> TokenGenerator -> cookie -> redirect|200
```

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `internal/infrastructure/port/in/google/routes.go` | Modify | Replace `http.StatusForbidden` with `http.StatusUnauthorized` in disallowed-email branch; keep error body unchanged. |
| `internal/infrastructure/port/in/google/routes_test.go` | Modify | Update disallowed-email assertion to `401`; preserve coverage for success and error branches. |

## Interfaces / Contracts

External HTTP contract update (Google callback disallowed-email case only):

```http
GET /login/oauth2/code/google?state=...&code=...

Before: 403 {"error":"email is not allowed"}
After:  401 {"error":"email is not allowed"}
```

No interface/type changes for:
- `ports.ProviderRepository`
- `ports.TokenGenerator`
- `model.UserInfo.IsEmailAllowed()`

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit/Handler | Disallowed email returns `401` and same message | Update `TestGoogleCallbackHandler_EmailNotAllowed` assertion. |
| Unit/Handler | Success paths remain stable | Keep `TestGoogleCallbackHandler_Success_NoRedirect` and `_WithRedirect` unchanged and passing. |
| Unit/Handler | Error branches unaffected (`invalid state`, `missing code`, provider/token failures) | Re-run existing tests in `routes_test.go`; maintain explicit branch assertions. |

Note: Existing suite is mostly per-case tests; if touched broadly, prefer table-driven consolidation per project standard, but not required for this narrow status-code change.

## Backward Compatibility

This is a deliberate API behavior change for one failure branch. Clients that currently interpret disallowed-email as `403` must update to handle `401`. Error payload text remains unchanged to reduce migration friction.

## Migration / Rollout

No data migration required. Rollout is immediate with deployment; no feature flag needed given limited scope.

## Rollback

Revert the status constant in `internal/infrastructure/port/in/google/routes.go` from `http.StatusUnauthorized` back to `http.StatusForbidden`, restore corresponding test assertion(s) in `routes_test.go`, and rerun callback handler tests.

## Open Questions

- [ ] Should we document auth status-code semantics (`401` vs `403`) in a shared API contract to avoid divergence across handlers?
