# Proposal: Standardize Auth Error Payloads

## Intent

Auth endpoints currently return inconsistent `{"error":"..."}` payloads with no machine-readable contract. This change standardizes auth failures to a shared JSON shape (`code`, `message`) so clients can handle errors predictably across middleware and OAuth callback paths.

## Scope

### In Scope
- Define one shared auth error response contract and helper for HTTP handlers/middleware.
- Apply the contract to `AuthMiddleware` unauthorized branches (missing/invalid token).
- Apply the contract to Google OAuth callback auth-related error branches (invalid request, disallowed email, provider/token failures) and align tests to assert schema + stable codes.

### Out of Scope
- Global non-auth error normalization for all routes.
- Changes to success payloads, cookie/session behavior, or OAuth happy path semantics.

## Capabilities

### New Capabilities
- `auth-error-payloads`: Defines canonical auth error envelope fields (`code`, `message`), code stability expectations, and response behavior for middleware/me-auth failures.

### Modified Capabilities
- `google-oauth-callback`: Tightens callback error payload contract to use standardized auth error envelope for relevant error branches.

## Approach

Implement a small shared auth error helper/type (recommended from exploration) and reuse it in scoped handlers. Keep handlers thin: map branch condition → stable auth code/message → status. Update tests to validate status plus response keys/values, avoiding fragile substring-only checks.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `pkg/security-middleware.go` | Modified | Replace ad-hoc unauthorized payloads with shared auth error envelope |
| `pkg/security-middleware_test.go` | Modified | Assert `code`/`message` contract |
| `internal/infrastructure/port/in/google/routes.go` | Modified | Use shared auth error envelope in callback error branches |
| `internal/infrastructure/port/in/google/routes_test.go` | Modified | Assert standardized callback error payloads |
| `internal/infrastructure/port/in/me/routes.go` | Modified | Align auth-related fallback errors to shared envelope (if branch is in auth scope) |
| `openspec/changes/issue-4-standardize-auth-error-payloads/specs/*` | New | Delta specs for new/modified capabilities |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Auth code taxonomy ambiguity | Med | Freeze code set in spec phase before implementation |
| Scope creep into all API errors | Med | Restrict to auth-related branches only |
| Test churn from stricter assertions | Med | Use table-driven assertions for stable schema contract |

## Rollback Plan

Revert helper usage in scoped files to prior payload shape and restore prior tests in the same commit window; no schema/data migration is required.

## Dependencies

- Exploration artifact: `sdd/issue-4-standardize-auth-error-payloads/explore` and `openspec/changes/issue-4-standardize-auth-error-payloads/exploration.md`.

## Success Criteria

- [ ] Scoped auth error responses return JSON with `code` and `message` only (or documented envelope fields).
- [ ] Middleware + OAuth callback tests assert standardized payload contract and pass.
- [ ] Spec deltas clearly define code/message semantics for affected capabilities.
