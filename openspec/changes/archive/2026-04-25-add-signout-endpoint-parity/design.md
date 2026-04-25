# Design: Add Sign-Out Endpoint Parity

## Technical Approach

Implement a minimal inbound HTTP addition that follows existing Gin wiring: extend the application wrapper with POST registration, add `POST /api/v1/auth/sign-out` in server route wiring, and implement a small sign-out handler that always responds `204` while clearing the `token` cookie. This keeps handlers thin and matches the current layered flow (`cmd` → `server.Routes` → inbound adapter handler).

## Architecture Decisions

| Option | Tradeoff | Decision |
|---|---|---|
| Add `RegisterPOST` to `pkg.Application` | Small wrapper growth, but preserves current abstraction used by route modules | ✅ Chosen |
| Register POST directly on Gin engine from server module | Less wrapper code, but breaks encapsulation/convention | Rejected |

| Option | Tradeoff | Decision |
|---|---|---|
| Keep sign-out endpoint unprotected and always clear cookie | Does not validate token presence, but allows idempotent logout and simpler client behavior | ✅ Chosen |
| Protect endpoint with auth middleware | Stronger gate, but fails logout when cookie is expired/invalid | Rejected |

| Option | Tradeoff | Decision |
|---|---|---|
| Clear cookie via `SetCookie("token", "", 0, "/", "", secureFromConfig, true)` | Keeps parity contract, but hard-codes `HttpOnly=true` at sign-out | ✅ Chosen |
| Reuse `http-only` config value for clearing | More configurable, but diverges from proposal’s explicit parity/security requirement | Rejected |

## Data Flow

Client `POST /api/v1/auth/sign-out`  
→ Gin route registered by `server.Routes`  
→ `signout.SignOutHandler` reads cookie security config (`Secure`)  
→ sets clearing cookie (`token` empty, `Max-Age=0`, `Path=/`, `HttpOnly=true`)  
→ returns `204 No Content`.

```
HTTP Client -> server.Routes -> app.RegisterPOST
                                |
                                v
                       signout.SignOutHandler
                                |
                                v
                       Set-Cookie(token=; Max-Age=0)
                                + 204
```

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `pkg/application.go` | Modify | Add `RegisterPOST(path, handler)` helper mirroring existing GET helper style. |
| `pkg/application_test.go` | Modify | Add coverage that POST registration routes and executes handler correctly. |
| `internal/infrastructure/port/in/server/server.go` | Modify | Wire `POST /api/v1/auth/sign-out` to sign-out handler. |
| `internal/infrastructure/port/in/signout/routes.go` | Create | Implement thin sign-out handler with cookie clearing + `204` response. |
| `internal/infrastructure/port/in/signout/routes_test.go` | Create | Validate status and cookie-clearing contract. |
| `internal/infrastructure/port/in/server/server_test.go` | Create (optional/minimal) | Verify route is reachable through server wiring (`POST` returns non-404 and expected status). |

## Interfaces / Contracts

```http
POST /api/v1/auth/sign-out
Response: 204 No Content
Headers:
  Set-Cookie: token=; Max-Age=0; Path=/; HttpOnly; [Secure if configured]
Body: empty
```

Handler contract (inbound adapter): no request body, no domain call, deterministic response.

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit | Application POST helper works | In `pkg/application_test.go`, register POST test endpoint and assert `200` from recorder. |
| Unit | Sign-out handler contract | In `signout/routes_test.go`, invoke POST route and assert `204`, empty body, and `token` cookie attributes (`Value=""`, `MaxAge=0`, `Path=/`, `HttpOnly=true`, `Secure` matches config). |
| Integration | Route wiring parity | In server routing test, execute `server.Routes(app)` and assert `POST /api/v1/auth/sign-out` is registered and returns sign-out response. |

## Migration / Rollout

No migration required. Safe additive API change.

## Open Questions

- [ ] Should server-level route reachability be covered only via sign-out handler test or also by a dedicated `server.Routes` integration test in this repo?
