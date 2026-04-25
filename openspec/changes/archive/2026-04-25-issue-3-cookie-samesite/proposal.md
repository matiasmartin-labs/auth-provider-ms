# Proposal: Apply configured SameSite on Google callback token cookie

## Intent

Ensure the auth token cookie set during Google OAuth callback honors `security.cookie.same-site`, closing a security/configuration gap where SameSite is currently not applied.

## Scope

### In Scope
- Update Google callback cookie-writing logic to map config `same-site` (`Strict`, `Lax`, `None`, case-insensitive) to HTTP cookie SameSite.
- Document and preserve behavior for invalid or empty `same-site` values (default/omitted SameSite attribute).
- Add/adjust tests to verify `Set-Cookie` SameSite behavior for supported modes and invalid/empty inputs.
- Guard against regressions in `Secure`, `HttpOnly`, and `Max-Age` cookie attributes.

### Out of Scope
- Changing sign-out cookie behavior or introducing new sign-out SameSite rules.
- Enforcing extra policy for `SameSite=None` beyond existing `Secure` behavior.
- Refactoring all cookie writes into a shared utility.

## Capabilities

### New Capabilities
None.

### Modified Capabilities
- `google-oauth-callback`: callback token cookie contract now includes config-driven SameSite behavior and explicit fallback semantics for invalid/empty values.

## Approach

Use a handler-local cookie write in `GoogleCallbackHandler` with `http.SetCookie` and explicit SameSite mapping from `CookieConfig.GetSameSite()`. Keep logic simple with early returns, preserve current cookie fields, and verify via header-level tests.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `internal/infrastructure/port/in/google/routes.go` | Modified | Set callback token cookie via explicit `http.Cookie` with mapped SameSite. |
| `internal/infrastructure/port/in/google/routes_test.go` | Modified | Add table-driven assertions for `Set-Cookie` SameSite modes and fallback behavior. |
| `openspec/changes/issue-3-cookie-samesite/specs/google-oauth-callback/spec.md` | New (next phase) | Delta spec for updated cookie SameSite contract. |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Invalid/empty config interpretation mismatch | Med | Specify fallback semantics in specs/tests and assert header output. |
| `SameSite=None` client expectations vary | Low | Document non-goal and keep `Secure` behavior unchanged. |
| Cookie attribute regression | Low | Add explicit regression assertions for `Secure`/`HttpOnly`/`Max-Age`. |

## Rollback Plan

Revert callback cookie write to prior `ctx.SetCookie(...)` path and remove SameSite-specific tests; this restores previous runtime behavior immediately without schema/config changes.

## Dependencies

- Existing `security.cookie.same-site` config ingestion via `pkg/config-security.go`.

## Success Criteria

- [ ] Callback `Set-Cookie` reflects configured SameSite for `Strict`, `Lax`, and `None`.
- [ ] Invalid/empty `same-site` behavior is documented and validated by tests.
- [ ] Callback cookie `Secure`, `HttpOnly`, and `Max-Age` behavior remains unchanged.
