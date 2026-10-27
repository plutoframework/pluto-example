# Pluto purchasing example

A runnable Go ergonomics prototype for an organization purchasing API. Northstar Research orders lab equipment; Harbor Studio buys production supplies. Members draft and submit requests, approvers decide requests from colleagues, and administrators manage existing memberships. Audit history follows the same organization boundary.

This is original synthetic application code, intended for review before extracting a framework. `internal/pluto` holds provisional HTTP helpers. The framework, CLI and RBAC library are developed in separate repositories. There is no frontend, CLI, production identity provider, or framework plugin system.

## Run in two terminals

Requires Go 1.25+; memory mode needs no database, container, or external service.

```sh
git clone https://github.com/plutoframework/pluto-example.git
cd pluto-example
PLUTO_EXAMPLE_DEMO_AUTH=true \
PLUTO_EXAMPLE_FEATURE_PURCHASE_EXPORT=true \
go run ./cmd/api
```

Then run the guided sequence:

```sh
./scripts/demo.sh
```

It creates a draft, submits it as Alice, approves it as Bob, checks Harbor's isolation, reads personal and organization audit history, exports JSON, and changes Alice's role as Carol. Memory state resets when the process restarts. The script prints the generated request ID and curl output, including expected failure cases. Re-running uses new request IDs and restores Alice to member at the end.

For individual requests, use [requests.http](requests.http). The generated [OpenAPI document](openapi/openapi.json) can be opened in an OpenAPI viewer; the server does not host a documentation UI.

Demo authentication is **off by default** and startup fails until explicitly enabled. It only permits a numeric loopback bind. Published tokens select fixed fixture actors; `X-Organization-ID` selects an organization only after membership lookup. Arbitrary identity/role headers and organization fields in purchase bodies do not grant access.

| Actor | Token | Organization and initial role |
| --- | --- | --- |
| Alice | `demo_alice_mJ7Kq9vL2xR8` | Northstar member; Harbor member |
| Bob | `demo_bob_tW4nC6yP1sF3` | Northstar approver |
| Carol | `demo_carol_hG8dA5uZ9eQ2` | Northstar administrator |
| Dana | `demo_dana_bN6rV3kT7jS4` | Harbor administrator |

These are local demo credentials, not a production authentication scheme. Bearer authentication is implemented; cookie sessions and CSRF defenses for cookie authentication are not.

## Read the code in this order

1. [Project manifest](pluto.toml) and [runtime configuration](internal/config/config.go): tooling metadata stays separate from one loaded, validated, injected settings value. Multi-tenancy is an explicit choice for this example. A future framework could default to single tenancy; that second application is not implemented here.
2. [Purchase endpoints](internal/paths/purchases.go): stable endpoint IDs, human names, descriptions, methods, templates, permissions, flags, and concrete wire models. Route strings live only in this catalog. [Roles and permissions](internal/domain/catalog.go) have independent IDs and human descriptions for future administration interfaces.
3. [Composition root](internal/app/app.go) and [route wiring](internal/app/routes.go): concrete adapters and constructor injection; no initialization registration or service locator.
4. [Purchase adapters](internal/handlers/purchases/create.go) and [HTTP boundary](internal/pluto/register.go): exact JSON field names, body limit, verified principal, operation call, consistent response/error envelope. Endpoint permission/feature checks happen before invoking a handler operation.
5. [Purchase operations](internal/operations/purchases.go) and [role/audit operations](internal/operations/administration.go): permissions, ownership, no self-decision, valid transitions, and tenant-scoped repository calls. These packages have no HTTP or database dependency. [Unit tests](internal/operations/purchases_test.go) exercise operations using a small fake.
6. [Feature definitions and evaluator](internal/features/features.go): immutable startup snapshot with absent-versus-explicit-false handling, declared fallbacks, and an internal-only flag proving the client allowlist. `/api/features` and the export gate share this evaluator. A flag enables a capability; it does not grant permissions. This adapter is not an OpenFeature/OFREP implementation.
7. [Transaction adapter](internal/storage/postgres/transactions.go), [SQL queries](db/queries/purchases.sql), and [migration](db/migrations/001_initial.sql): tenant predicates, composite tenant keys, optimistic versions, and purchase/role writes plus audit inserts in one transaction. [Memory adapter](internal/storage/memory/store.go) uses one mutex per write/audit pair for demonstration; it has no durability or PostgreSQL transaction guarantees.
8. [Router tests](internal/app/routes_test.go), [architecture check](internal/architecture/architecture_test.go), and [SQL integration test](internal/storage/postgres/integration_test.go): real middleware/registration with fake operations, tenant and response assertions, dependency rules, and opt-in SQL verification.

## Domain and transport contracts

A request starts `draft`. Only its creator can submit it. A submitted request can become `approved` or `rejected` exactly once through the API, by another actor holding the review permission. Administrators follow the same no-self-decision policy. Amounts are integer cents from 1 through 100,000,000; titles must be nonblank, valid UTF-8 without NUL, and at most 160 bytes before trimming.

Members read their own requests and their own audit actions. Approvers/administrators read organization requests and organization audit history. Tenant-mismatched IDs return not-found; a token without membership in the selected organization returns unauthenticated. Administrators may list existing members and change another member's role. Invitations/removals are deliberately outside this example. Self role changes are forbidden; this is not a transactional guarantee that concurrent administrators cannot demote one another.

Successful responses use `{"data": ...}`. Errors use `{"error":{"code":"...","message":"...","request_id":"..."}}`. Statuses: 400 invalid input, 401 unauthenticated, 403 permission/policy or disabled feature, 404 hidden/missing resource, 405 method, 409 stale version/invalid transition, 413 oversized body, 500 unexpected failure. Error text does not reveal database internals.

The server has read/header/write/idle timeouts, a 16 KiB JSON request limit, a 16 KiB header limit, generated request IDs, panic recovery, and global [security headers](internal/pluto/security.go). Gzip is optional and negotiated through Chi's existing compression middleware for JSON. There is no CORS middleware because no cross-origin browser app is implemented. No request body or bearer token is logged. Diagnostic JSON slog records include status and verified endpoint/actor/organization enrichment, including denied feature checks. Log output can fan out to stdout and a file.

The security policy matches the CLI generator: `nosniff`, `X-Frame-Options: DENY`, `Referrer-Policy: no-referrer`, `no-store`, a restrictive JSON-API CSP (`default-src 'none'; base-uri 'none'; form-action 'none'; frame-ancestors 'none'`), disabled camera/microphone/geolocation permissions, and Helmet-style cross-origin and legacy browser policies. It runs before authentication/recovery so success, denied, malformed, missing-route, method-error and recovered-panic responses retain the headers. [Router tests](internal/app/security_test.go) check that wiring. The authored middleware is the customization point for HTML/docs or embeddable resources. CSP applies to the API document, not a separately hosted UI; CORP does not configure CORS. Set HSTS at the HTTPS terminator after choosing a domain policy; the local HTTP example neither trusts forwarded protocol headers nor enables HSTS/subdomains/preload automatically. Headers do not replace authorization, validation or CSRF defenses for cookie authentication.

Audit events are an explicit business API containing event identity, actor, organization, action, resource, outcome, and time. Only successful writes are persisted in this prototype. Denied requests appear in diagnostic logs, not durable audit records. PostgreSQL records business mutation and audit atomically. There is no external SIEM connection or durable delivery; an outbox could be added at the transaction boundary later.

## Configuration

[.env.example](.env.example) documents every supported environment override. The application **does not load `.env` automatically**. For a local shell, explicitly export its values (for example `set -a; . ./.env.example; set +a`) or use the inline commands above. No secret belongs in `pluto.toml`.

Only uppercase `PLUTO_EXAMPLE_` names are used. Defaults plus environment are supported; no runtime config-file loader is implied. Unknown variables are ignored; invalid values for recognized booleans, addresses, modes, and log levels fail startup. Feature settings preserve unset versus false with `*bool`. Export defaults to false and requires restart after a setting change. Server timeout/body/header defaults are set explicitly in configuration/code; only address and gzip have environment overrides in this sample.

## Optional PostgreSQL path

Requires a running Docker daemon. The compose project uses only a dedicated local container, loopback port 55432, and no persistent volume. It applies the migration and synthetic fixture file on container initialization.

```sh
docker compose -p pluto-example up -d --wait
PLUTO_EXAMPLE_DEMO_AUTH=true \
PLUTO_EXAMPLE_DATABASE_MODE=postgres \
PLUTO_EXAMPLE_DATABASE_URL='postgres://pluto_demo:local_only@127.0.0.1:55432/pluto_demo?sslmode=disable' \
PLUTO_EXAMPLE_FEATURE_PURCHASE_EXPORT=true \
go run ./cmd/api
# When finished:
docker compose -p pluto-example down
```

Alternatively, apply [001_initial.sql](db/migrations/001_initial.sql) and [fixtures.sql](db/fixtures.sql) once to a new dedicated PostgreSQL database. There is no production migration runner or generic interchangeable database abstraction.

For the SQL integration check, use the disposable-container script. It allocates a dynamic loopback port and deletes its own container on exit:

```sh
./scripts/postgres-check.sh
```

With an existing **dedicated test database**, an alternative is:

```sh
PLUTO_EXAMPLE_TEST_DATABASE_URL='postgres://user:password@127.0.0.1:5432/test_database?sslmode=disable' \
go test -race -count=1 ./internal/storage/postgres -run TestPostgres -v
```

The test creates a random isolated schema, applies migration/fixtures there, and drops only that schema. It exercises real SQL tenant predicates, failed-audit rollback of creates/transitions/roles, and concurrent optimistic writes. Without the test URL, it reports `SKIP`; in-memory tests are not evidence for SQL isolation.

## Repeatable checks

The example pins **golangci-lint 2.15.0** in [mise.toml](mise.toml). After reviewing the project configuration, set up the tool and use these entry points:

```sh
mise trust
mise install
mise run fmt
mise run lint
mise run check
```

The scripts also work directly (`./scripts/fmt.sh`, `./scripts/golangci.sh run ./...`, `./scripts/check.sh`). They use the pinned version through mise, or an already installed matching golangci-lint binary. Python 3.11+ is required by the scripts; Go itself remains sufficient for `go test ./...`.

[.golangci.yml](.golangci.yml) defines a focused readability policy:

- `golines` wraps long signatures and calls at an 80-column default target. Projects can change `formatters.settings.golines.max-len` in `.golangci.yml`. Long indivisible strings/comments may remain longer; this is a formatting target, not a hard byte limit.
- `wsl_v5` separates unrelated assignments, completed blocks, and returns while keeping immediate error checks next to their calls and related calls together.
- `gofumpt` applies consistent Go formatting, and `gci` groups standard-library, application, and external imports.
- `govet`, `staticcheck`, `ineffassign`, and `unused` provide static correctness checks.

`fmt.sh` applies formatters, then the whitespace analyzer's automatic fixes, then formatting again. `lint` and `check` never rewrite source and fail if these rules are violated. Running `golangci-lint fmt` alone does **not** apply `wsl_v5` fixes. The policy applies to handwritten tests as well as application code; sqlc output keeps its generator's format and remains subject to compilation, tests, vet, and file limits. Developers still separate semantic phases where tools cannot infer them; no rule here guarantees every struct literal is multiline or proves the absence of business logic in handlers.

Additional verification commands:

```sh
go test ./...
go vet ./...
go test -race ./...
go test ./internal/operations -run '^$' -fuzz '^FuzzDecisionPolicy$' -fuzztime=3s -parallel=2
go test ./internal/pluto -run '^$' -fuzz '^FuzzDecodePurchase$' -fuzztime=3s -parallel=2
python3 scripts/mutation-smoke.py
go run ./cmd/openapi -check
```

Native fuzz tests assert policy/no-write invariants and bounded exact-key JSON decoding, rather than only checking for panics. [Seeded mutation smoke](scripts/mutation-smoke.py) removes six meaningful guards in temporary copies and independently runs the relevant behavior tests. It reports killed, surviving, invalid/build-failed, and timed-out mutations separately and fails unless all six are killed. This is a small reproducible smoke check, not comprehensive mutation coverage or a general mutation engine. Expand real mutation tooling once the framework surface stabilizes; CRAP metrics are deferred.

Regenerate the two committed artifacts reproducibly:

```sh
go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.29.0 generate
go run ./cmd/openapi
```

sqlc is a pinned development generator, not a runtime dependency. OpenAPI derives route metadata and concrete JSON model structure from shared Go definitions, with reusable schema references. It does not infer business policies, amount/title constraints, enums, or every runtime validation rule. Errors use a shared default response; status mapping is documented above. It is intentionally a sample generator, not a general schema engine.

The architecture test walks source imports transitively, rejects handler access to database adapters/drivers and business access to HTTP/storage, checks paths do not reach handlers, and enforces a 1,000-line Go source limit including generated code. [check.sh](scripts/check.sh) runs the non-mutating golangci-lint checks, checks every non-Git project file's line count, runs tests/vet, and verifies OpenAPI freshness. It can be invoked unchanged by a future Git hook or CI job; this example does not install a hook or remote CI workflow. These checks enforce dependency boundaries; they do not semantically prove that handlers contain no business logic.

## Prototype limits and review questions

Lists/exports/audit history are unpaginated and not streaming. There is no production authentication, browser UI, cookie session implementation, idempotency key, monetary currency model, attachment workflow, budget/approval threshold, or distributed feature service. Authorization uses membership at request start; concurrent role revocation may race an already authorized write. SQL tenant isolation uses explicit predicates/composite keys rather than PostgreSQL row-level security. Production deployment needs its own identity, migration, concurrency, and operational decisions.

The useful review questions are whether endpoint/schema declarations remain pleasant to author, whether these narrow operation/repository boundaries are the right extraction point, and whether role changes should eventually enforce an organization-wide last-administrator invariant. No framework API is declared stable by this example.
