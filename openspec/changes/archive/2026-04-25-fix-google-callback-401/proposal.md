# Proposal: Align Google Callback Disallowed Email to 401

## Intent

Issue #2 requires parity with upstream auth semantics: when Google OAuth callback receives a disallowed email, return `401 Unauthorized` (not `403`) while keeping the response payload clear and preserving successful login behavior.

## Scope

### In Scope
- Update Google callback disallowed-email response status from `403` to `401`.
- Preserve existing error payload/message clarity for this branch.
- Update callback tests to assert `401` and verify no regression in success path.

### Out of Scope
- Refactoring callback error handling into shared/global error mappers.
- Changing allowlist logic, user model behavior, or OAuth route wiring.

## Capabilities

### New Capabilities
- `google-oauth-callback`: Defines callback response semantics for allowed and disallowed emails, including unauthorized handling.

### Modified Capabilities
- None.

## Approach

Apply a targeted handler change in the disallowed-email branch (`!userInfo.IsEmailAllowed()`), replacing `http.StatusForbidden` with `http.StatusUnauthorized`. Keep payload unchanged. Update unit tests in the Google callback suite (table-driven where applicable) to cover success and disallowed-email error branches explicitly.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `internal/infrastructure/port/in/google/routes.go` | Modified | Disallowed-email callback status becomes `401`. |
| `internal/infrastructure/port/in/google/routes_test.go` | Modified | Assertions updated to `401`; success + error branch coverage preserved. |
| `openspec/changes/fix-google-callback-401/specs/` | New | Delta spec(s) for callback unauthorized behavior. |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Clients expecting `403` may need adaptation | Medium | Communicate API contract update in issue/PR notes. |
| Hidden assumptions in callback tests | Low | Run callback-focused tests covering success and disallowed-email paths. |

## Rollback Plan

Revert the status constant change in `routes.go` and restore previous test expectations in `routes_test.go`; rerun callback test suite to confirm behavior returns to pre-change baseline.

## Dependencies

- Existing Google OAuth callback handler and test harness in current codebase.

## Success Criteria

- [ ] Disallowed-email callback response returns HTTP `401 Unauthorized`.
- [ ] Error payload remains explicit and unchanged in meaning.
- [ ] Callback success path tests still pass with no behavioral regression.
