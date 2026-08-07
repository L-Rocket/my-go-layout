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

## Testing

Stdlib `testing` only; no testify, no external services needed.

```sh
make build vet test   # standard Makefile; `make gates` before every push
```

The repo has four test layers, each with a distinct job:

- **Domain slices** (next to the code): unit-test `Service` against the
  memory store — no servers, no network.
- **Transport**: full-stack `httptest` with real services + memory stores.
- **Infrastructure/postgrest**: `httptest`-faked Supabase asserting paths,
  headers, and error mapping.
- **`test/` — whole-service integration tests** (package `test`). The app
  is assembled here **exactly as `cmd/server/main.go` wires it** (memory
  stores → domain services → transports) and driven over its **public
  API on real loopback ports** — no `httptest`, no reaching into
  internals. Rules: exported APIs only (it may import `internal/` as the
  same module), and **fakes belong only at true external boundaries** —
  there are none here, since memory is the default backend.

Current integration coverage: `TestHTTPSessionLifecycle` (create → post
message → list → end → posting after end returns 409) and
`TestGRPCReflection` (the gRPC server comes up on a real port and
answers a reflection ListServices call).

Unit tests stay next to the code they cover; `test/` is the only place
cross-layer end-to-end tests live.
