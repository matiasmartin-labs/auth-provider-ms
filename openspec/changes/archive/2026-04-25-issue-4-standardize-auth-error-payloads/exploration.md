## Exploration: issue-4-standardize-auth-error-payloads

### Current State
Auth-related responses currently return mixed JSON error payloads that all use a single `error` string field, but with no shared contract type or stable machine-readable code. The two main paths in scope are: (1) `AuthMiddleware` in `pkg/security-middleware.go` (missing/invalid token → `401`), and (2) Google OAuth callback in `internal/infrastructure/port/in/google/routes.go` (invalid state/code `400`, disallowed email `401`, provider/token failures `500`). Tests mostly assert status and substring message, not a shared structured contract.

### Affected Areas
- `pkg/security-middleware.go` — currently emits `{"error": "..."}` for unauthorized responses; must adopt standardized auth error schema.
- `pkg/security-middleware_test.go` — assertions currently check raw message substrings; should assert schema keys and stable code/message semantics.
- `internal/infrastructure/port/in/google/routes.go` — callback error branches currently emit `{"error": "..."}` for bad request/unauthorized/internal errors.
- `internal/infrastructure/port/in/google/routes_test.go` — callback tests must assert the same schema on unauthorized/error branches.
- `internal/infrastructure/port/in/me/routes.go` — contains fallback unauthorized/internal errors and may need alignment to avoid contract drift on auth endpoint behavior.
- `internal/infrastructure/port/in/me/routes_test.go` — if `me` fallback responses are included in scope, tests need schema assertions.
- `openspec/specs/google-oauth-callback/spec.md` — currently specifies payload clarity but not a concrete schema; follow-up spec delta should codify canonical fields.

### Approaches
1. **Inline replacement per handler/middleware** — Replace every `gin.H{"error": ...}` in scoped files with `gin.H{"code": ..., "message": ...}` directly.
   - Pros: Fastest path, low implementation overhead, minimal file churn.
   - Cons: Duplicates mapping logic and string constants; higher risk of future drift and inconsistent codes.
   - Effort: Low.

2. **Shared auth error contract helper (recommended)** — Introduce a small shared helper/type for auth error responses (e.g., `code` + `message`) and use it from middleware + callback branches.
   - Pros: Enforces one schema at compile-time usage points, improves consistency, cleaner tests, easier future expansion.
   - Cons: Slight upfront refactor and naming decisions; requires careful scope control to avoid broad API refactor.
   - Effort: Medium.

3. **Global error middleware normalization** — Centralize all error responses via Gin middleware and map auth failures into unified envelope.
   - Pros: Strong long-term consistency across all endpoints.
   - Cons: Over-scoped for this issue; requires broader handler error propagation changes and larger regression surface.
   - Effort: High.

### Recommendation
Use **Approach 2 (Shared auth error contract helper)** with strict scope to issue #4: middleware + OAuth2 callback unauthorized/error branches (and optionally `me` fallback if considered in-scope during proposal/spec clarification). This best balances contract consistency (Java-aligned `code` + `message`) with low regression risk, while avoiding a larger global error architecture change.

### Risks
- **Contract ambiguity risk**: Issue notes “for example, code + message” but not exact enum values; codes must be agreed (e.g., `AUTH_TOKEN_MISSING`, `AUTH_TOKEN_INVALID`, `AUTH_EMAIL_NOT_ALLOWED`, etc.) before implementation.
- **Scope creep risk**: Expanding to all endpoint errors (e.g., JWKS/non-auth) may delay delivery; keep scope explicit to auth-related branches.
- **Test brittleness risk**: Existing tests rely on substring matching; converting to strict JSON structure assertions may require broader updates than expected.
- **OpenSpec context gap**: `openspec/config.yaml` is currently absent, so phase-specific local OpenSpec rules cannot be validated from repo config.

### Ready for Proposal
Yes — enough code and test context exists to draft proposal/spec/tasks for a focused refactor that introduces one auth error schema and applies it consistently to middleware and OAuth2 callback error paths.
