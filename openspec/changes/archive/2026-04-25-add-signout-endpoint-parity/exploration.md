## Exploration: add-signout-endpoint-parity

### Current State
The HTTP server currently registers only GET endpoints via `Application.RegisterGET` and `Application.RegisterProtectedGET`; there is no POST registration helper and no existing sign-out route. Authentication cookie issuance happens in the Google callback handler (`/login/oauth2/code/google`) using `ctx.SetCookie("token", token, maxAge, "/", "", secure, httpOnly)`, where `secure` and `httpOnly` are read from `security.cookie` config. Protected auth functionality currently exposes only `GET /api/v1/auth/me` behind middleware.

### Affected Areas
- `pkg/application.go` — currently supports only GET route helpers; adding `POST /api/v1/auth/sign-out` likely requires a `RegisterPOST` helper (and optionally protected variant).
- `internal/infrastructure/port/in/server/server.go` — central route registration point where `/api/v1/auth/sign-out` must be wired.
- `internal/infrastructure/port/in/google/routes.go` — existing source of truth for token cookie creation behavior; useful to mirror security/cookie behavior expectations.
- `internal/infrastructure/port/in/google/routes_test.go` — demonstrates current cookie assertions pattern (find `token` cookie and assert attributes/value).
- `cmd/provider-auth-ms/config.yaml` and `pkg/config-security.go` — define cookie security knobs (`secure`, `http-only`, `max-age`, `same-site`) that sign-out behavior must preserve where applicable.
- `internal/infrastructure/port/in/.../*_test.go` (new sign-out test file expected) — tests should cover happy path (`204`) and cookie-clearing contract.

### Approaches
1. **Dedicated sign-out handler + POST route support** — add a focused sign-out inbound handler and register it on `POST /api/v1/auth/sign-out`.
   - Pros: Clean parity with required endpoint/method, explicit intent, easiest to test at handler level, minimal blast radius.
   - Cons: Requires small framework extension (`RegisterPOST`) since app wrapper is GET-only today.
   - Effort: Low.

2. **Register route directly on Gin engine (bypass app wrapper)** — expose/use raw router internals to add POST without extending `Application`.
   - Pros: No new wrapper method.
   - Cons: Breaks current abstraction pattern, leaks infrastructure details, creates inconsistency for future endpoints.
   - Effort: Low/Medium.

3. **Model sign-out as GET (keep existing helpers)** — add sign-out on GET for minimal plumbing.
   - Pros: No wrapper changes.
   - Cons: Violates explicit requirement (`POST /api/v1/auth/sign-out`), weaker semantic parity with Java service.
   - Effort: Low, but not acceptable.

### Recommendation
Use **Approach 1**: introduce a dedicated sign-out handler and add explicit POST registration support in the `Application` wrapper, then wire `POST /api/v1/auth/sign-out` from `server.Routes`. Implement cookie clearing with `token` cookie, `Max-Age=0`, `Path=/`, and `HttpOnly=true`, while keeping existing security behavior alignment by reusing config-driven fields where expected (notably `Secure`). Add tests validating `204 No Content` and that the response includes a clearing `token` cookie with required attributes.

### Risks
- **Config vs hard requirement ambiguity**: requirement says `HttpOnly=true`, while config exposes `http-only`; forcing true may diverge from configurable behavior if currently set false in some envs.
- **Cookie attribute coverage gap**: `same-site` exists in config but is not currently applied through `ctx.SetCookie`; parity expectations should avoid silently introducing behavior not used elsewhere unless explicitly requested.
- **Routing abstraction change**: adding POST helpers in `pkg.Application` can affect conventions/tests; existing tests currently cover only GET and protected GET registration.

### Ready for Proposal
Yes — enough context exists to draft proposal/spec/tasks for a small feature change: add POST sign-out endpoint, clear auth cookie with required attributes, and add focused tests for status and cookie clearing behavior.
