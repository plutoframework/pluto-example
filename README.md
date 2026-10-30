# Pluto purchasing example

A runnable Go ergonomics prototype for an organization purchasing API. Northstar Research orders lab equipment; Harbor Studio buys production supplies. Members draft and submit requests, approvers decide requests from colleagues, and administrators manage existing memberships. Audit history follows the same organization boundary.

This is original synthetic application code, intended for review before extracting a framework. `internal/pluto` holds provisional HTTP helpers. The framework, CLI, auth and RBAC libraries are developed in separate repositories. A React purchasing frontend now lives in [web/](web/README.md). There is no production delivery provider or framework plugin system.

## Review the full-stack application

See [web/README.md](web/README.md) for the HTTPS React application, feature layout,
contract generation, architecture checks and real API/browser walkthrough. The
frontend is a convention prototype for review before frontend CLI generators.
It uses React/TypeScript, Vite, TanStack Router/Query and generated OpenAPI bindings.
`mise run web` starts its local HTTPS server after `npm --prefix web ci`.

## Run the local credential walkthrough

Requires Go 1.25.5+, Python 3 and Docker. This example is explicitly development
only: startup requires DEMO_AUTH=true, a numeric loopback listener and PostgreSQL.
There is no credential memory fallback or startup DDL. Memory repositories remain
available for isolated business tests.

```sh
docker compose -p pluto-example up -d --wait
set -a; . ./.env.example; set +a
PLUTO_EXAMPLE_FEATURE_PURCHASE_EXPORT=true go run ./cmd/api
# Another terminal, with the same exported .env.example settings:
go run ./cmd/worker
# A third terminal, with the same environment:
./scripts/demo.sh
# Stop the API and worker, then discard database and Mailpit:
docker compose -p pluto-example down
```

Use a **fresh disposable database and Mailpit** for each walkthrough. The script
generates random passwords in memory, enrolls synthetic alice/bob/carol/dana at
example.test, denies unverified sign-in, consumes emailed verification challenges,
signs in and proves that all four initially lack business access. It obtains
account IDs from authenticated session metadata, then explicitly invokes
`go run ./cmd/dev-fixtures --alice-id ... --bob-id ... --carol-id ... --dana-id ...`.
This trusted developer command requires all four verified/enabled synthetic account
IDs, inserts fixture memberships plus audit records in one transaction, and never
runs from HTTP, enrollment, email matching or startup. It grants Alice member in
Northstar/Harbor, Bob Northstar approver, Carol Northstar administrator and Dana
Harbor administrator. Do not use this fixture administration in production.

The walkthrough submits/approves a purchase, checks isolation, audit, flags and
role policy, then exercises recovery/reset, old-session/password rejection,
reauthentication and logout. Passwords, cookies and challenges are not printed or
written to disk. Only the disposable Mailpit capture holds raw message contents;
it retains at most 100 messages in process memory, with no mounted storage.
Discard it afterwards. Mailpit has no relay configuration. Its UI is loopback
http://127.0.0.1:8025. No real recipients or external delivery are used.

[requests.http](requests.http) shows individual requests. The
[OpenAPI document](openapi/openapi.json) references [auth paths](openapi/auth.json)
and [wire schemas](openapi/schemas.json). Named endpoint metadata remains in
`internal/paths`; local credentials use the published `passwordhttp` adapter.

POST /auth/enroll, /verification, /verify, /signin, /recovery, /reset and /logout
use exact trusted Origin validation and bounded 8 KiB JSON. Responses contain only
an opaque code. Enrollment/verification-resend/recovery accept with 202 without
revealing eligibility; delivery failures also accept and emit a redacted warning.
Request an explicit resend to replace the challenge; duplicate enrollment never
resends or changes a password. GET never consumes a challenge. Sign-in sets a
Secure, HttpOnly, SameSite=Lax, host-only __Host-pluto_session cookie; browsers need
HTTPS. The Python walkthrough explicitly sends the cookie for local HTTP. Reset
requires a new sign-in and revokes existing sessions. No token appears in JSON.

GET /auth/session and DELETE /auth/session retain session-only metadata/revocation
compatibility without requiring a tenant. Opaque bearer sessions remain opt-in in
transport wiring; mixed cookie/Authorization credentials fail. Anonymous lifecycle
routes mount before business tenancy/RBAC and never enter idempotency capture.
All local sessions use `service.Sessions()` against the SAME `postgres.NewLocal`
backend; do not issue/rotate local accounts with the session-only store or Start.
Auth accounts own no application profile, roles or memberships. Fresh unknown
accounts receive no business privileges, even with a matching display name/email.

The development sender uses stdlib SMTP only to a numeric loopback sink and only
example.test recipients. TLS/AUTH are intentionally absent on this disposable
connection; its dial/I/O/cancellation budget is three seconds. This is not a
production SMTP adapter. `auth.LocalDependencies` exposes a real context-aware
Sender, AttemptLimiter, HTTP Ingress, optional PasswordPolicy and redacted failure
hook. Production requires explicit delivery with fixed trusted HTTPS confirmation
URLs (prefer token fragments), TLS/auth/timeouts, a maintained password blocklist
when required, shared identifier/client limits, TLS/database privileges, account
retention and scheduled session/challenge pruning. Never use request Host/return
URLs or persist raw tokens in an outbox. No production integration is claimed.

The local ingress ignores X-Forwarded-For, admits only the direct loopback socket
peer, bounds 120 requests/minute per peer and four concurrent requests with no
queue, and supplies a ten-second context. Account limits are bounded, process-local
20/minute per operation/identifier. Neither is suitable for a proxy deployment or
multiple replicas; implement trusted/shared enforcement there. Active Argon2 work
cannot be interrupted; queueing and downstream work honor cancellation.

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

The server has read/header/write/idle timeouts, a 16 KiB JSON request limit, a 16 KiB header limit, generated request IDs, panic recovery, and global [security headers](internal/pluto/security.go). Gzip is optional and negotiated through Chi's existing compression middleware for JSON. The HTTPS frontend uses a same-origin development proxy; there is no wildcard CORS middleware. No request body or bearer token is logged. Diagnostic JSON slog records include status and verified endpoint/actor/organization enrichment, including denied feature checks. Log output can fan out to stdout and a file.

The security policy matches the CLI generator: `nosniff`, `X-Frame-Options: DENY`, `Referrer-Policy: no-referrer`, `no-store`, a restrictive JSON-API CSP (`default-src 'none'; base-uri 'none'; form-action 'none'; frame-ancestors 'none'`), disabled camera/microphone/geolocation permissions, and Helmet-style cross-origin and legacy browser policies. It runs before authentication/recovery so success, denied, malformed, missing-route, method-error and recovered-panic responses retain the headers. [Router tests](internal/app/security_test.go) check that wiring. The authored middleware is the customization point for HTML/docs or embeddable resources. CSP applies to the API document, not a separately hosted UI; CORP does not configure CORS. Set HSTS at the HTTPS terminator after choosing a domain policy; the local HTTP example neither trusts forwarded protocol headers nor enables HSTS/subdomains/preload automatically. Headers do not replace authorization, validation or CSRF defenses for cookie authentication.

Audit events are an explicit business API containing event identity, actor, organization, action, resource, outcome, and time. Only successful writes are persisted in this prototype. Denied requests appear in diagnostic logs, not durable audit records. PostgreSQL records business mutation and audit atomically. Each successful API business mutation also enqueues a tenant-scoped audit-delivery job in that same transaction. The separate worker archives the canonical event in an idempotent local demonstration sink; no external SIEM connection is configured.

## Authorization

The example imports the published `github.com/plutoframework/pluto-rbac` module,
pinned to `v0.1.0` in go.mod/go.sum. There is no local
replacement or dependency on a sibling checkout. [The application catalog](internal/domain/catalog.go)
uses the library's permission/role types: members have base capabilities,
approvers inherit members, and administrators inherit approvers.
`domain.NewPolicy()` compiles and validates the catalog once at startup.
The composition root injects the same immutable policy into route registration
and business operations; `Principal.Can(policy, permission)` delegates grant
evaluation to the package rather than a second local role/permission algorithm.
Unknown roles/permissions and missing policies deny capability checks.
Catalog roles contain direct grants and inheritance; use the library's
`EffectivePermissions` to enumerate inherited capabilities for an assignment UI.

The session adapter resolves membership for the exact actor/organization
pair and supplies the trusted role in that principal. The application consumes
that request-start snapshot rather than performing another membership lookup for
every permission check. Tenant-scoped SQL, ownership checks, no self-approval,
feature gates and transactional audits remain application responsibilities.
Concurrent revocation can still race an already authorized operation; this
integration does not imply transactional authorization/revocation guarantees.

## Configuration

[.env.example](.env.example) documents every supported environment override. The application **does not load `.env` automatically**. For a local shell, explicitly export its values (for example `set -a; . ./.env.example; set +a`) or use the inline commands above. No secret belongs in `pluto.toml`.

Only uppercase `PLUTO_EXAMPLE_` names are used. Defaults plus environment are supported; no runtime config-file loader is implied. Unknown variables are ignored; invalid values for recognized booleans, addresses, modes, and log levels fail startup. Feature settings preserve unset versus false with `*bool`. Export defaults to false and requires restart after a setting change. Server timeout/body/header defaults are set explicitly in configuration/code; only address and gzip have environment overrides in this sample.

## PostgreSQL migrations and persisted tests

Compose initializes a new disposable database with migrations 001 through 004 and
synthetic domain fixtures. Existing databases must explicitly apply
[004_local_credentials.sql](db/migrations/004_local_credentials.sql) **after**
001_initial, 002_auth_sessions and 003_idempotency. It adds accounts, digest-only
challenges and versioned session bindings without changing existing session-only
rows. No constructor performs DDL. Do not apply fixtures to a production database.
Runtime users should have DML privileges separately from migration owners.

```sh
./scripts/postgres-check.sh
```

This script creates PostgreSQL on a dynamic loopback port and removes its container
on exit. Tests create/drop only random isolated schemas. It runs persisted domain,
audit/rollback/concurrency, session compatibility and complete application HTTP
credential tests under the race detector. Without a dedicated
PLUTO_EXAMPLE_TEST_DATABASE_URL those tests report SKIP, not SQL evidence.

## Repeatable checks

The example pins **golangci-lint 2.15.0** in [mise.toml](mise.toml). After reviewing the project configuration, set up the tool and use these entry points:

```sh
mise trust
mise install
mise run fmt
mise run lint
mise run check
```

The scripts also work directly (`./scripts/fmt.sh`, `./scripts/golangci.sh run ./...`, `./scripts/check.sh`). They use the pinned version through mise, or an already installed matching golangci-lint binary. Python 3.11+ is required by the scripts; Go itself remains sufficient for `go test ./cmd/... ./internal/...`.

[.golangci.yml](.golangci.yml) defines a focused readability policy:

- `golines` wraps long signatures and calls at an 80-column default target. Projects can change `formatters.settings.golines.max-len` in `.golangci.yml`. Long indivisible strings/comments may remain longer; this is a formatting target, not a hard byte limit.
- `wsl_v5` separates unrelated assignments, completed blocks, and returns while keeping immediate error checks next to their calls and related calls together.
- `gofumpt` applies consistent Go formatting, and `gci` groups standard-library, application, and external imports.
- `govet`, `staticcheck`, `ineffassign`, and `unused` provide static correctness checks.

`fmt.sh` applies formatters, then the whitespace analyzer's automatic fixes, then formatting again. `lint` and `check` never rewrite source and fail if these rules are violated. Running `golangci-lint fmt` alone does **not** apply `wsl_v5` fixes. The policy applies to handwritten tests as well as application code; sqlc output keeps its generator's format and remains subject to compilation, tests, vet, and file limits. Developers still separate semantic phases where tools cannot infer them; no rule here guarantees every struct literal is multiline or proves the absence of business logic in handlers.

Additional verification commands:

```sh
go test ./cmd/... ./internal/...
go vet ./...
go test -race ./cmd/... ./internal/...
go test ./internal/operations -run '^$' -fuzz '^FuzzDecisionPolicy$' -fuzztime=3s -parallel=2
go test ./internal/pluto -run '^$' -fuzz '^FuzzDecodePurchase$' -fuzztime=3s -parallel=2
python3 scripts/mutation-smoke.py
go run ./cmd/openapi -check
```

Native fuzz tests assert policy/no-write invariants and bounded exact-key JSON decoding, rather than only checking for panics. [Seeded mutation smoke](scripts/mutation-smoke.py) removes eleven meaningful guards in temporary copies and independently runs the relevant behavior tests. It reports killed, surviving, invalid/build-failed, and timed-out mutations separately and fails unless all eleven are killed. This is a small reproducible smoke check, not comprehensive mutation coverage or a general mutation engine. Expand real mutation tooling once the framework surface stabilizes; CRAP metrics are deferred.

Regenerate the two committed artifacts reproducibly:

```sh
go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.31.1 generate
go run ./cmd/openapi
```

sqlc is a pinned development generator, not a runtime dependency. sqlc 1.31.1
uses a Go 1.26+ toolchain for generation (Go can download it automatically); the
application continues to target Go 1.25.5+. OpenAPI derives route metadata and concrete JSON model structure from shared Go definitions, with reusable schema references. It does not infer business policies, amount/title constraints, enums, or every runtime validation rule. Errors use a shared default response; status mapping is documented above. It is intentionally a sample generator, not a general schema engine.

The architecture test walks source imports transitively, rejects handler access to database adapters/drivers and business access to HTTP/storage, checks paths do not reach handlers, and enforces a 1,000-line Go source limit including generated code. [check.sh](scripts/check.sh) runs the non-mutating golangci-lint checks, checks every non-Git project file's line count, runs tests/vet, and verifies OpenAPI freshness. It can be invoked unchanged by a future Git hook or CI job; this example does not install a hook or remote CI workflow. These checks enforce dependency boundaries; they do not semantically prove that handlers contain no business logic.

## Prototype limits and review questions

Collections use bounded cursor pages; the separate export remains unpaginated and is not streaming. There is no production delivery/shared-limit integration, browser UI, monetary currency model, attachment workflow, budget/approval threshold, or distributed feature service. Authorization uses membership at request start; concurrent role revocation may race an already authorized write. SQL tenant isolation uses explicit predicates/composite keys rather than PostgreSQL row-level security. Production deployment needs its own identity, migration, concurrency, and operational decisions.

The useful review questions are whether endpoint/schema declarations remain pleasant to author, whether these narrow operation/repository boundaries are the right extraction point, and whether role changes should eventually enforce an organization-wide last-administrator invariant. No framework API is declared stable by this example.

## Shared pagination and idempotency

`github.com/plutoframework/pluto-core` owns the reusable runtime. Thin adapters in
internal/pluto wire shared types; domain policy and tenant predicates stay here.
Purchases/memberships order by unique immutable ID; audit orders by occurred_at
and ID, preserving same-timestamp boundaries. The list route catalog explicitly
marks Collection; GET details/features/session/export remain separate. List
responses retain data plus page: {has_more,next_cursor}. Default limit 25, maximum
100. Use ?limit=2, then ?limit=2&cursor=<next_cursor>. Filters and verified
actor/organization/role/endpoint bind cursors. Cursors are unsigned opaque
boundaries, never authorization. Repositories enforce tenant/owner predicates
before LIMIT limit+1; memory retains only bounded next-page rows while scanning.
OpenAPI now splits readable schemas into openapi/schemas.json (external refs).

Business writes (create, submit, decide, change role) require a client-supplied
Idempotency-Key, 1–128 ASCII letters/digits/-_.:. Pick a new key for a distinct
action and reuse it for retries. The request demo supplies action keys on the
client. Same subject/organization/operation/key and exact fingerprint replays
status/body; changed request or pending outcome returns 409. Authentication,
current membership/RBAC and feature gates run before every replay. Sessions and
credential lifecycle routes are exempt and never cache cookies or credentials.
Only Content-Type is retained with bounded JSON bytes/status; global security
headers remain current. Body limit stays 16 KiB; recorded response limit 256 KiB.

Memory test adapters are process-local and cannot deduplicate across instances.
The runnable example uses PostgreSQL with the separate
published Core adapter and explicit db/migrations/003_idempotency.sql. Apply this
to existing databases; no startup DDL. Completed retention is 24h from claim;
expired completed keys can execute again. Schedule adapter.PruneCompleted from
an application-owned maintenance path; it removes at most 1,000 rows per call.
Pending claims never expire automatically or get stolen after a crash. Monitor
and reconcile them against business/audit records before explicitly resolving or
removing them. Memory capacity and storage failures fail closed.

This example's generic JSON capture records completion after the domain write:
mutation+audit remain atomic with each other, but a crash before completion leaves
the retry blocked for reconciliation. It does not promise exactly-once effects.
For atomic mutation+audit+completion use the lower-level Core claim primitives
and `postgres.New(tx).Complete` in the SAME transaction as typed app queries; the
[core transaction walkthrough](https://github.com/plutoframework/pluto-core/blob/main/docs/transactions.md)
and real PostgreSQL rollback/commit tests demonstrate that pattern. Arbitrary
business transactions and external effects are never automatically retried.

## Durable background delivery

The example imports published Pluto Jobs v0.1.1 and its Postgres adapter v0.1.0.
Apply `db/migrations/005_jobs.sql` to existing databases before starting the updated
API or worker. Fresh Compose databases receive it before development fixtures;
existing databases are never migrated implicitly. Auth schema 004 remains separate.

Purchase creation/transitions and role changes write domain state, business audit
and an `audit.deliver` outbox request in ONE pgx transaction. If any insert fails,
all three roll back. The request contains only an event ID and explicit verified
tenant metadata, never a raw credential/challenge or audit body. HTTP handlers
remain thin; enqueue happens in the repository transaction, not after HTTP success.
The in-memory business repository is for isolated unit tests and has no durable
outbox; the runnable app requires PostgreSQL.

`go run ./cmd/worker` explicitly registers the audit handler using the published
worker implementation. Nested `config.Config.Jobs` holds its bounded defaults.
There are no worker goroutines in the API. The handler loads the canonical audit
with BOTH tenant and event predicates and copies it into `audit_archive`. A unique
(tenant_id,event_id) key makes redelivery after acknowledgement loss safe. This
archive is a real, local example effect, not a claim of external SIEM/email/webhook
delivery. Replace the narrow AuditSink implementation to integrate a real provider.
Trusted development membership fixtures do not produce external delivery jobs.

Inspect delivery without exposing payloads:

```sql
SELECT state, attempts, diagnostic, count(*) FROM pluto_jobs
GROUP BY state, attempts, diagnostic;
SELECT tenant_id, event_id, delivered_at FROM audit_archive
ORDER BY delivered_at;
```

Delivery is at least once with finite retries and no heartbeat. Use stable job.ID
or event IDs for provider idempotency, honor context deadlines and cooperative
shutdown, and deploy compatible handlers before producers. Applications own
retention, monitoring, failed-job inspection and reconciliation. The worker logs
no payloads/provider errors. Auth notification delivery remains synchronous; raw
challenge tokens are deliberately never put in this outbox.

`./scripts/postgres-check.sh` tests real business/audit/outbox rollback, uncommitted
visibility, worker retries, tenant isolation and duplicate-safe archival, along
with the credential and pagination scenarios. It creates/removes its own database.

Jobs integration validation passed: repository check/lint/vet/OpenAPI freshness,
race tests, real disposable PostgreSQL scenarios above, nine killed mutations
(no survivors/invalid builds/timeouts), and a three-second handler fuzz sample
with 145,905 executions. A live API + separate worker + loopback Mailpit walkthrough
produced five jobs, archived all five, and left zero unsettled jobs. Disposable
services were removed. These are integration checks, not an independent audit.


## Data cache

The example uses published Pluto Core v0.2.1 and its optional Valkey adapter
v0.1.0. `GET /api/organization` demonstrates cache-aside reads of organization
display metadata for one minute. Live sessions, membership and `organization:read`
permissions are checked on every request, including hits. Enrollment grants none.
The operation also validates the cached organization ID against the verified tenant.

The default `PLUTO_EXAMPLE_CACHE_BACKEND=memory` uses a bounded, expiring LRU.
Set it to `none` to bypass storage, or run the optional development service:

```sh
docker compose --profile cache up -d valkey
export PLUTO_EXAMPLE_CACHE_BACKEND=valkey
export PLUTO_EXAMPLE_CACHE_URL='redis://:local_only@127.0.0.1:6379/0'
```

These are synthetic local credentials. Supply real credentials through your secret
manager. External connections use `rediss://` with verified TLS; plaintext beyond
loopback requires explicit `PLUTO_EXAMPLE_CACHE_ALLOW_INSECURE=true`. A process
inside Docker must use `valkey:6379` and explicitly opt into local plaintext.
Set `PLUTO_EXAMPLE_CACHE_ENVIRONMENT` separately for each deployment. Scoped hashed
keys include the application, environment, namespace, schema version and tenant.

The cache is an optimization: backend failures fall back to Postgres, and identity,
roles, auth challenges, idempotency and audit delivery never depend on it. Display
metadata can remain stale until expiry. A future organization write must invalidate
the same key after commit; deletion alone does not eliminate concurrent refill
races. There is no organization update API or transactional cache guarantee here.

The application owns cache construction and cleanup in `internal/caching` and
`internal/app`; operations consume the shared Core interface. Unit tests prove
hits, tenant/environment isolation, fallback, invalidation and live permissions.
`./scripts/postgres-check.sh` also starts authenticated Valkey and verifies real HTTP
membership revocation after a warm cache. Core separately tests the same adapter
against Redis and Valkey. No additional SQL migration is needed for this cache.

## Optional OpenTelemetry

The API and worker explicitly construct their own optional Core telemetry runtime,
with service names `pluto-example-api` and `pluto-example-worker`, version
`development`, and the configured cache environment. Set
`PLUTO_EXAMPLE_OTEL_ENABLED=true` and optionally
`PLUTO_EXAMPLE_OTEL_ENDPOINT=http://127.0.0.1:4318` for an OTLP/HTTP collector.
`PLUTO_EXAMPLE_OTEL_SAMPLE_RATIO` defaults to 1. Telemetry is disabled by default.
Configuration is read once; runtime providers and propagation remain local to the
application. Both commands flush telemetry on exit with a separate bounded
shutdown context, after application resources close.

Chi middleware records route templates, and context-aware JSON slog calls correlate
request logs. PostgreSQL pools receive the runtime query tracer before connecting;
shared auth/idempotency queries therefore use the same tracer. Cache calls and
worker audit delivery use explicit runtime wrappers. Worker execution starts a new
trace: no trace headers or baggage are persisted in the job payload or schema.
Transactional audit enqueueing retains `EnqueueTx` and its atomicity; it is observed
through PostgreSQL instrumentation rather than the nontransactional enqueuer
adapter. There are no application-owned outbound HTTP clients in this example.
Instrumentation excludes raw request paths, SQL, query parameters, cache keys,
job payloads and credential data. Business audit records remain independent.

## Read-only agent purchase review

Purchase review is off by default. With the normal local Postgres/Mailpit setup
and HTTPS frontend, start the API with:

```sh
PLUTO_EXAMPLE_AGENT_MODE=demo \
PLUTO_EXAMPLE_FEATURE_PURCHASE_REVIEW=true \
PLUTO_EXAMPLE_SERVER_ADDRESS=127.0.0.1:18080 go run ./cmd/api
```

Use the ordinary authentication configuration described above. Open a visible
purchase and choose **Review purchase**. Demo mode invokes the actual scoped
purchase operation with a deterministic scripted provider and labels its results.
It makes no model/network request. `POST /api/purchases/{purchaseID}/review`
returns `{data:{summary,next_steps,demo}}`. It cannot approve, pay or change data.

Real mode requires explicit `PLUTO_EXAMPLE_AGENT_MODE=openai`, the review feature
flag, `PLUTO_EXAMPLE_AGENT_MODEL=gpt-5.4` (or its supported dated snapshot), and
`PLUTO_EXAMPLE_AGENT_OPENAI_API_KEY` supplied privately through the environment.
Enabled invalid real configuration fails startup. Never store credentials in
pluto.toml. No live credentials are needed for tests: the real adapter is tested
against a local HTTP fake. The endpoint is fixed by the adapter in real mode;
model arguments cannot choose URLs, purchase IDs, identity, tenant or roles.

Current permission and resource checks run before a provider is constructed and
again in the tool. The only registered/run-allowed tool reads the server-selected
purchase through existing operations. Advice is untrusted text for human review;
it grants no authority. Optional application telemetry uses the published agent
OTEL observer without prompts, tool arguments, results or credentials in spans.

Each request starts a fresh run, even if an Idempotency-Key is supplied. There is
no response replay cache, credential cache, durable run store or exactly-once
claim. Client retries are disabled; deliberately triggering another review can
incur another provider charge. Limits are two model calls, one tool call, no
repairs, 512 output tokens per call, 8,000 accounted tokens and a ten-second run
budget (four seconds/model, one second/tool). Token accounting is post-hoc, not a
hard spending cap. Cancellation is cooperative and cannot undo provider billing.
No shared rate limiter or durable agent audit trail is claimed.

To exercise review in the serial browser walkthrough set
`PLUTO_EXAMPLE_E2E_AGENT_REVIEW=true` with demo mode and its flag enabled. Set
`PLUTO_EXAMPLE_E2E_MAILPIT` to a disposable loopback Mailpit API when using dynamic
ports. Fixture setup respects explicitly supplied environment configuration.

## Purchase lifecycle rules

[Purchase lifecycle](internal/domain/purchase_lifecycle.go) uses the published
`github.com/plutoframework/pluto-core/state` package (Core v0.4.0). Draft requests
can be submitted by their creator; submitted requests can be approved or rejected
by another actor. Approved/rejected are terminal. Guard facts come from the loaded
purchase and verified principal, never directly from request bodies or AI output.

The service compiles its static definition once. Operations still enforce RBAC
and tenant visibility, then ask `Next` for a decision. The existing repository
atomically checks the expected version and writes the purchase, audit event and
outbox delivery. All updates to guard-relevant data must increment that version.
A database concurrency test forces competing approval/rejection operations to
read the same snapshot and proves only one commits, with no losing audit/job.
No schema migration or frontend contract change is required.
