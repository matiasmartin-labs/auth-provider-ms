# Design: Standardize Auth Error Payloads

## Technical Approach

Introduce a small shared auth-error contract in `pkg` and route all in-scope auth failure branches through that contract. The implementation maps each known auth failure condition to a stable `code` and client-facing `message`, then emits JSON via a helper so middleware and handlers stay thin. This follows the proposal scope: `AuthMiddleware` unauthorized responses plus Google callback auth-related error branches, with optional `me` alignment only for auth-scoped failures.

## Architecture Decisions

### Decision: Centralize auth error envelope

| Option | Tradeoff | Decision |
|---|---|---|
| Inline `gin.H{"code","message"}` at each call site | Fast, but duplicates constants and risks drift | ❌ |
| Shared contract type + helper in `pkg` | Small refactor, but single source of truth and reusable tests | ✅ |
| Global error middleware for all routes | Strong consistency, but out of scope and high regression risk | ❌ |

**Choice**: Shared contract helper in `pkg`.
**Rationale**: Matches current architecture (handlers call package-level helpers), minimizes scope, and reduces future divergence.

### Decision: Stabilize auth codes with scoped taxonomy

| Option | Tradeoff | Decision |
|---|---|---|
| Keep free-form messages only | Human-readable, but not machine-actionable | ❌ |
| Add stable auth `code` + user-safe `message` | Requires taxonomy governance, but enables deterministic clients/tests | ✅ |

**Choice**: Emit `code` and `message` for in-scope auth failures.
**Rationale**: Satisfies proposal/spec intent for predictable client behavior without changing success payloads.

## Data Flow

Auth failures converge to one response path:

    Request → Middleware/Handler branch check
                     │
                     ├─ missing/invalid token (middleware)
                     ├─ invalid callback params
                     ├─ disallowed email
                     └─ provider/token auth failure
                             │
                             v
                 pkg.WriteAuthError(ctx, status, code, message)
                             │
                             v
                 HTTP JSON {"code": "...", "message": "..."}

Success branches (token set, redirect/200) remain unchanged.

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `pkg/auth_error.go` | Create | Defines shared auth error payload type, constants, and writer helper used by middleware/handlers. |
| `pkg/security-middleware.go` | Modify | Replace `{"error": ...}` unauthorized payloads with shared helper and stable codes for missing/invalid token. |
| `internal/infrastructure/port/in/google/routes.go` | Modify | Map callback auth-related branches to shared helper (`400/401/500` with stable auth codes/messages). |
| `internal/infrastructure/port/in/me/routes.go` | Modify (conditional) | Align auth-related `claims` missing/invalid branches to shared contract if confirmed in scope. |
| `pkg/security-middleware_test.go` | Modify | Assert status and exact `code`/`message` contract instead of message-only substring checks. |
| `internal/infrastructure/port/in/google/routes_test.go` | Modify | Assert standardized error envelope for invalid state/code, disallowed email, and provider/token failures. |
| `internal/infrastructure/port/in/me/routes_test.go` | Modify (conditional) | Assert standardized envelope for auth-scoped failure branches if `me` remains in scope. |

## Interfaces / Contracts

```go
// pkg/auth_error.go
type AuthErrorResponse struct {
    Code    string `json:"code"`
    Message string `json:"message"`
}

func WriteAuthError(ctx *gin.Context, status int, code, message string)

const (
    AuthCodeTokenMissing   = "AUTH_TOKEN_MISSING"
    AuthCodeTokenInvalid   = "AUTH_TOKEN_INVALID"
    AuthCodeCallbackState  = "AUTH_CALLBACK_STATE_INVALID"
    AuthCodeCallbackCode   = "AUTH_CALLBACK_CODE_MISSING"
    AuthCodeEmailDenied    = "AUTH_EMAIL_NOT_ALLOWED"
    AuthCodeProviderFailed = "AUTH_PROVIDER_FAILURE"
    AuthCodeTokenIssue     = "AUTH_TOKEN_GENERATION_FAILED"
)
```

Contract constraints:
- Envelope fields for in-scope auth failures are `code` and `message`.
- `message` stays safe/actionable (no internals); internal errors are not leaked.
- Codes are stable identifiers for clients and tests.

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit | Helper emits exact JSON shape and status | Table-driven tests for `WriteAuthError` (status, code, message keys only). |
| Integration | Middleware + callback branches return mapped codes | Existing Gin route tests updated to decode JSON and assert exact contract per branch. |
| E2E | Auth login/callback happy path unchanged | Run existing callback success tests (redirect/no redirect, cookie behavior) as regression guard. |

## Migration / Rollout

No migration required. Rollout is code-only and backward-compatible for successful responses; only in-scope auth error payloads change shape.

## Open Questions

- [ ] Confirm final code taxonomy in spec delta (exact enum set and naming).
- [ ] Confirm whether `internal/infrastructure/port/in/me/routes.go` fallback errors are explicitly in-scope for this issue.
