# AGENTS.md — my-go-layout

A demo of one architecture: **vertical slices inside a hexagonal core**.
This file is the law for how the repo grows. If you change the
architecture, change this file in the same commit.

## The architecture

```
        HTTP / gRPC
            │
            ▼
     transport/            driving adapters — protocols only, no rules
            │
            ▼
     session/  message/    the hexagon — vertical slices, business rules
            │
        interface          ports (store.go) — OWNED BY THE DOMAIN
            │
            ▼
     infrastructure/       driven adapters — postgrest, memory, logger
            │
     Postgres / Redis / external APIs
```

Two independent rules make up the design:

1. **Hexagonal = which way arrows point.** Dependencies only point
   inward: transport → domain ← infrastructure. The domain imports
   nothing above itself. Ports (`Store` interfaces) are declared in the
   domain package that consumes them — never next to their
   implementation. `cmd/server/main.go` is the composition root and the
   only place adapters are wired to ports.
2. **Vertical slice = how the core is cut.** The domain is organized by
   aggregate (`session/`, `message/`), not by technical role. A slice
   owns everything about its feature in one package: types, rules,
   port, errors. There is no shared `service/` blob and no `dao/`
   package collecting interfaces away from their owners.

## File placement

```
cmd/server/main.go      # assembly ONLY: config → adapters → services → transports → serve
internal/
  transport/
    http/               # gin engine, route table, middleware, envelope, errors.go
    grpc/               # grpc.Server, interceptors, reflection
  session/              # slice: session.go (types), service.go (rules),
                        #        store.go (port), errors.go (domain errors)
  message/              # slice: same four files
  infrastructure/
    postgrest/          # port implementations over Supabase PostgREST
    memory/             # port implementations in-memory (local default)
    logger/             # slog setup
  config/               # env vars with defaults — nothing else
```

Slices live **flat** under `internal/` while there are few of them — a
wrapper folder adds a path segment with no meaning. Once the domain
exceeds **five** slices (or domain-adjacent packages appear: shared
types, domain events), group them under `internal/domain/`
(`internal/domain/session`, …) in one move. The name is `domain/`, not
`core/` — it says what the packages are, not where they sit in a
metaphor.

## Rules for future development

**Adding an endpoint**
1. Add/extend the domain method on the slice's `Service`.
2. Register the route in `transport/http/routes.go`: decode → call one
   service method → encode. Handlers contain **no business logic**.
3. Map any new domain error to a status in `transport/http/errors.go` —
   the ONLY place that mapping exists.

**Adding a domain module (e.g. `project/`)**
1. Create the four files: `<name>.go` (types), `service.go` (rules),
   `store.go` (port interface), `errors.go` (domain errors).
2. Implement its port in `infrastructure/memory/` (required — the binary
   must always run with zero external services) and, when a real backend
   exists, in `infrastructure/postgrest/`.
3. Wire it in `main.go`. Nothing else should know how it is constructed.

**Cross-module dependencies**
- A slice needing another slice's rules declares a **narrow interface in
  its own package** (see `message.SessionService`) and receives the
  concrete service from main. Never import the other slice's `store.go`,
  never re-implement its rules, and keep the interface to the methods
  actually used.

**Persistence**
- Port methods are domain-shaped (`ListByUser`), not SQL-shaped.
- Both backends must behave identically (IDs, ordering, errors) — if
  PostgREST uses a `bigserial`, memory uses one global counter, not a
  per-entity sequence.
- Empty results are `nil, nil`; missing rows are `ErrNotFound` from the
  slice's `errors.go`, never a raw PostgREST error leaking upward.

**Utilities**
- `internal/` root holds only layers and slices. A technical capability
  with no domain meaning (crypto, logging, metrics) goes to
  `infrastructure/`; protocol-specific helpers stay in `transport/`.
- **No orphans**: every package must be wired into the running binary.
  Dead code is deleted, not kept "for later".

**Testing** (stdlib only, no testify)
- Slices: unit-test `Service` against the memory store — no servers.
- Transport: full-stack `httptest` with real services + memory stores.
- Infrastructure/postgrest: `httptest`-faked Supabase asserting paths,
  headers, and error mapping.
- `test/`: whole-service integration tests (package `test`). Assemble the
  app exactly as `cmd/server/main.go` wires it (memory stores → domain
  services → transports) and drive the PUBLIC API over real ports. Unit
  tests stay next to the code; `test/` is the only place cross-layer
  end-to-end tests live. It may import `internal/` (same module) but must
  use exported APIs only, and fakes belong only at true external
  boundaries — there are none by default, since memory is the local
  backend.
- Gates before every push: `gofmt -l .`, `GOTOOLCHAIN=local go vet ./...`,
  `GOTOOLCHAIN=local go test ./...` (Makefile pins `GOTOOLCHAIN=local`;
  the `go` directive stays on the 1.25 line).

## Deliberate non-goals (demo scope)

- No authentication — ownership is a plain `user_id` string in requests.
- PostgREST schema is not shipped (`migrations/` is convention notes
  only); the memory backend is the reference implementation.
- gRPC serves reflection only, as a second-transport demonstration.
