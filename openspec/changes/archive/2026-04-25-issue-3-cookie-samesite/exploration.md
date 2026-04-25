## Exploration: apply configured cookie SameSite policy when setting auth token

### Current State
`same-site` is already present in runtime config (`security.cookie.same-site`) and exposed through `CookieConfig.GetSameSite()`, but the auth token cookie path (`GoogleCallbackHandler`) currently uses `gin.Context.SetCookie(...)`, which does not apply SameSite from config in the current implementation. As a result, issued auth token cookies include max-age/secure/http-only but not config-driven SameSite.

### Affected Areas
- `internal/infrastructure/port/in/google/routes.go` — token cookie is set here; current `ctx.SetCookie(...)` call is the behavior gap.
- `internal/infrastructure/port/in/google/routes_test.go` — tests currently validate token presence/redirect flow but do not assert Set-Cookie SameSite variants.
- `pkg/config-security.go` — `CookieConfig` already exposes `GetSameSite()` and mapstructure binding for `same-site`.
- `cmd/provider-auth-ms/config.yaml` — default config includes `same-site: Strict` and serves as documented runtime input.
- `pkg/config_test.go`, `pkg/application_test.go` — fixture configs already include `same-site`, confirming config ingestion path.

### Approaches
1. **Handler-local explicit cookie write (recommended)** — build an `http.Cookie` in `GoogleCallbackHandler` and call `http.SetCookie(ctx.Writer, cookie)` with mapped SameSite.
   - Pros: Minimal surface area; preserves existing control flow; straightforward to test via `Set-Cookie` header; lowest regression risk.
   - Cons: Cookie write path in this handler diverges from `ctx.SetCookie` style.
   - Effort: Low.

2. **Shared cookie utility/helper** — introduce a helper that sets token cookie with mapped SameSite and call it from handler(s).
   - Pros: Reusable and centralizes parsing/mapping rules.
   - Cons: Slightly larger refactor for a small fix; potentially touches signout/other routes unnecessarily.
   - Effort: Medium.

3. **Global/middleware Set-Cookie mutation** — post-process headers to inject SameSite globally.
   - Pros: Centralized policy.
   - Cons: High risk, opaque behavior, can affect unrelated cookies and ordering; hardest to reason about.
   - Effort: High.

### Recommendation
Use **Approach 1**: apply SameSite directly in `GoogleCallbackHandler` by switching only this token cookie write to `http.SetCookie` and mapping config string to `http.SameSite` (`Strict`, `Lax`, `None`, case-insensitive). For invalid/empty values, set `SameSiteDefaultMode`/omit explicit SameSite attribute (documented behavior) so existing behavior remains stable and predictable.

### Risks
- **Behavioral ambiguity for invalid/empty values**: must be explicitly documented and tested; otherwise behavior can differ across clients.
- **`SameSite=None` browser expectations**: many browsers require `Secure` for `None`; current acceptance does not require enforcement, but this should be called out.
- **Regression on existing cookie attributes**: secure/httpOnly/max-age must remain byte-for-byte compatible in header assertions.

### Ready for Proposal
Yes — scope is clear, impact is localized, and tests can be added to prove supported SameSite modes plus no regression in secure/httpOnly/max-age behavior.
