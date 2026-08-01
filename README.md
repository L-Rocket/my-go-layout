# my-go-layout

A small demo of a layered, vertical-modular Go layout. Runs locally with
**zero external services** (in-memory storage by default).

## Layout

```
my-go-layout/
├── cmd/server/main.go            # the ONLY assembly point: config → infrastructure → domain services → transports → serve
├── internal/
│   ├── transport/
│   │   ├── http/                 # gin engine, route table, middleware, envelope, single error→status mapping
│   │   └── grpc/                 # grpc.Server + reflection + logging interceptor (no business RPCs yet)
│   ├── session/                  # Session domain module (types, service rules, Store contract, errors)
│   ├── message/                  # Message domain module
│   ├── infrastructure/
│   │   ├── postgrest/            # Store implementations over Supabase PostgREST
│   │   ├── memory/               # in-memory Store implementations (local default)
│   │   └── logger/               # thin log/slog wrapper (JSON)
│   └── config/                   # env vars with defaults — nothing else
├── api/proto/                    # gRPC contracts will live here (README only for now)
├── migrations/                   # SQL migrations (timestamped naming; README only for now)
├── test/                         # whole-service integration tests: assembled like cmd/server, driven over the public API on real ports
└── Makefile / Dockerfile / .gitignore
```

## Layer rules

- **Dependencies point inward.** Transports and infrastructure depend on
  domain modules; domain modules depend on nothing above them.
- **The domain owns its storage contract** (dependency inversion): each
  module declares a `Store` interface; `internal/infrastructure/*`
  implements it. The domain never imports infrastructure.
- **Cross-module rule:** `message.Service` depends on `session.Service`
  (via a narrow `SessionService` interface declared in the message
  package), **not** on `session.Store` — ownership and lifecycle rules
  live in exactly one place.
- **Errors map to HTTP statuses in exactly one place:**
  `internal/transport/http/errors.go`. Handlers only decode → call → encode.
- **`cmd/server` is the only place that knows how pieces are wired.**

## Quickstart

```sh
make run          # serves HTTP on :8080, gRPC (reflection only) on :50051
```

Then, in another terminal:

```sh
# create a session (ownership is just a string — no user module in this demo)
curl -s -X POST localhost:8080/api/sessions -d '{"user_id":"user-ada","title":"demo"}'
# → {"data":{"id":"sess-...","status":"active",...}}

# post + list messages
curl -s -X POST localhost:8080/api/sessions/sess-.../messages -d '{"user_id":"user-ada","content":"hello"}'
curl -s "localhost:8080/api/sessions/sess-.../messages?user_id=user-ada"

# end the session (posting afterwards → 409)
curl -s -X POST localhost:8080/api/sessions/sess-.../end -d '{"user_id":"user-ada"}'

# liveness
curl -s localhost:8080/health
# → {"status":"up","store":"memory"}
```

## How storage is selected

At startup, `cmd/server` checks the env: if **both** `SUPABASE_URL` and
`SUPABASE_SERVICE_ROLE_KEY` are set, the PostgREST implementations are
wired; otherwise the in-memory stores are used. The choice is logged
(`storage selected store=memory|postgrest`) and reported by `/health`.
The PostgREST side assumes `sessions`/`messages` tables exist — schema is
not shipped in this demo.

## Tests & gates

```sh
make build vet test   # GOTOOLCHAIN=local is pinned in the Makefile
```

Stdlib `testing` only; no external services needed — the PostgREST tests
run against an `httptest` fake.
