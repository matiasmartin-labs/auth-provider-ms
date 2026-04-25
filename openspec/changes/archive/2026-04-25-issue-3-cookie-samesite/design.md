# Design: Apply SameSite on Google Callback Token Cookie

## Technical Approach

Implement a minimal-risk, handler-local change in `GoogleCallbackHandler` to set the token cookie with `http.SetCookie` and explicit `SameSite` mapping from `security.cookie.same-site`.

This preserves current callback flow, keeps existing cookie attributes (`Secure`, `HttpOnly`, `Max-Age`, path/name/value) unchanged, and avoids introducing shared cookie abstractions for this targeted fix.

## Architecture Decisions

| Option | Tradeoff | Decision |
|---|---|---|
| Map SameSite in `google/routes.go` and emit cookie with `http.SetCookie` | Smallest blast radius; one handler diverges from `ctx.SetCookie` style | ✅ Chosen |
| Add shared cookie utility used by callback + signout | Better reuse, but larger refactor and out-of-scope risk | ❌ Rejected |
| Global Set-Cookie middleware mutation | Centralized policy, but opaque behavior and high regression risk | ❌ Rejected |

| Mapping fallback strategy | Tradeoff | Decision |
|---|---|---|
| Invalid/empty config maps to zero-value `http.SameSite` (omit attribute) | Keeps backward-compatible behavior and avoids ambiguous `SameSite` token serialization | ✅ Chosen |
| Invalid/empty maps to `http.SameSiteDefaultMode` | May serialize as bare `SameSite` token depending on Go behavior; can alter existing header contract | ❌ Rejected |

## Data Flow

```
GET /login/oauth2/code/google
  -> validate state/code
  -> ProviderRepository.GetUserInfo
  -> UserInfo.IsEmailAllowed
  -> TokenGenerator.GenerateToken
  -> cookieCfg.GetSameSite() --(normalize/validate)-> http.SameSite
  -> http.SetCookie(writer, token cookie)
  -> redirect (307) OR JSON 200
```

SameSite normalization rules (case-insensitive, trimmed):
- `Strict` -> `http.SameSiteStrictMode`
- `Lax` -> `http.SameSiteLaxMode`
- `None` -> `http.SameSiteNoneMode`
- Any other value / empty -> omit SameSite attribute

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `internal/infrastructure/port/in/google/routes.go` | Modify | Replace `ctx.SetCookie` call with explicit `http.Cookie` + SameSite mapping helper local to package/file. |
| `internal/infrastructure/port/in/google/routes_test.go` | Modify | Add table-driven callback success tests asserting SameSite mapping and fallback behavior from `Set-Cookie`; keep existing branch tests passing. |

## Interfaces / Contracts

No exported interface changes (`pkg.CookieConfig` already exposes `GetSameSite() string`).

Internal helper contract (non-exported):

```go
func parseSameSite(raw string) http.SameSite
```

HTTP contract update for callback cookie:
- `Set-Cookie: token=...` now includes `SameSite=Strict|Lax|None` when configured accordingly.
- Invalid/empty config keeps prior behavior by not emitting a SameSite attribute.

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit/Handler | SameSite mapping (`Strict`, `Lax`, `None`, invalid, empty) | Table-driven subtests in `routes_test.go`; inspect `w.Result().Cookies()` and raw `Set-Cookie` header for attribute presence/absence. |
| Unit/Handler | Non-regression for existing cookie attrs | Assert token cookie still sets `HttpOnly`, configured `Secure`, and expected `MaxAge`; keep name/path/value unchanged. |
| Unit/Handler | Existing callback behavior unaffected | Re-run existing tests for invalid state, missing code, provider/token errors, email-not-allowed, redirect/no-redirect success. |

## Migration / Rollout

No migration required. Rollout is immediate on deploy. No feature flag required due to localized change.

## Open Questions

- [ ] Should signout cookie emission also adopt explicit SameSite handling in a separate follow-up change for symmetry?
