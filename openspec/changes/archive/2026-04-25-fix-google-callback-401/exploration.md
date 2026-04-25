## Exploration: fix-google-callback-401

### Current State
The Google OAuth callback handler currently validates `state` and `code`, fetches user info from the provider, validates the user email against configured allowlist, and then issues a token/cookie plus optional redirect. In the disallowed-email branch (`!userInfo.IsEmailAllowed()`), it returns `403 Forbidden` with payload `{ "error": "email is not allowed" }`. Existing unit tests in the Google callback suite assert `403` for that path.

### Affected Areas
- `internal/infrastructure/port/in/google/routes.go` — contains the callback branch that currently returns `http.StatusForbidden` for disallowed email.
- `internal/infrastructure/port/in/google/routes_test.go` — `TestGoogleCallbackHandler_EmailNotAllowed` currently expects `http.StatusForbidden`; must assert `http.StatusUnauthorized`.
- `internal/domain/model/user_info.go` — provides `IsEmailAllowed()` behavior used by callback flow; no behavior change expected, but relevant to branch coverage.
- `internal/infrastructure/port/in/server/server.go` — route registration for callback endpoint; no code change expected, but confirms impacted endpoint (`/login/oauth2/code/google`).

### Approaches
1. **Targeted status-code correction** — Change only the disallowed-email HTTP status from 403 to 401 and update related tests.
   - Pros: Minimal, low-risk, directly satisfies acceptance criteria, preserves existing payload/message and success flow behavior.
   - Cons: Keeps authorization semantics embedded in handler (not centralized).
   - Effort: Low.

2. **Centralized auth error mapping** — Introduce shared auth error types/mapping and make callback return mapped status for disallowed identity.
   - Pros: Better long-term consistency across handlers/middleware.
   - Cons: Larger scope for a small bugfix, higher regression risk, unnecessary for current issue.
   - Effort: Medium.

### Recommendation
Use **Approach 1 (Targeted status-code correction)**. It is the smallest possible change to meet issue #2 and all acceptance criteria: disallowed email returns 401, payload stays actionable, tests align with 401, and successful callback/login flow remains untouched.

### Risks
- Semantic/API compatibility risk: clients that previously treated this specific branch as 403 may need to handle 401 (expected per issue alignment with Java service).
- Test coverage risk: if only one test is updated, hidden assumptions in other integration paths could remain unverified; run callback test suite to confirm no success-flow regression.

### Ready for Proposal
Yes — exploration is sufficient to move to proposal/spec/tasks for a focused bugfix that updates callback unauthorized behavior and corresponding tests.
