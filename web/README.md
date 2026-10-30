# Pluto purchasing frontend

A React/TypeScript application for the real Go purchasing API, rebuilt with
HeroUI v3. The UI uses the same curated library profile as Pluto CLI v0.5.0.
It includes sign-in, enrollment, email confirmation/resend, recovery/reset,
purchasing, draft submission, independent review, role management and audit views.

## Run locally

From the repository root, install the pinned Node 24 toolchain and dependencies:

```sh
mise install
mise exec -- npm --prefix web ci
# Fresh, isolated development Postgres and Mailpit:
docker compose -p pluto-frontend-review up -d --wait
set -a; . ./.env.example; set +a
PLUTO_EXAMPLE_SERVER_ADDRESS=127.0.0.1:18080 \
  PLUTO_EXAMPLE_FEATURE_PURCHASE_EXPORT=true go run ./cmd/api
```

In a second terminal, run `mise run web`, then open
**https://127.0.0.1:5173**. Accept the self-signed certificate warning for this
loopback development site. The certificate is generated locally by OpenSSL,
expires after 30 days, and its ignored private key has mode 0600. Nothing is added
to the OS trust store. Never reuse it for production.

The HTTPS Vite proxy forwards `/api` and `/auth` to numeric loopback port 18080,
preserves Origin and Secure/HttpOnly cookie attributes, and keeps the browser on
one origin. The Go allowlist explicitly includes https://127.0.0.1:5173. No
wildcard CORS, cookie rewriting or bearer tokens in browser storage. Port 8080
remains available for the original API walkthrough. For another loopback API,
set PLUTO_EXAMPLE_WEB_API_TARGET in Vite's process and configure Go separately.

Use an `example.test` address and Mailpit at http://127.0.0.1:8025. Enrollment and
verification grant no memberships. A trusted developer must explicitly bind the
four verified synthetic fixture account IDs using the root README's command.
An account without membership shows its ID and an access explanation. Choosing
an organization without membership does not discard the valid account session.
The organization picker lists the two known example fixtures; it is not a
production organization-discovery API. The original repository fixtures also
retain unbound synthetic memberships for business tests, so display names can
repeat; membership identity always uses the server account ID.

The worker is optional for browsing, but run `go run ./cmd/worker` with the same
exported Go environment to deliver the existing transactional outbox to its local
archive sink. There is no external SIEM or email provider configured.

## Where things belong

```text
src/
  app/                         router, providers, shell, visual tokens
  api/
    client.ts                  the one HTTP boundary
    generated/                 OpenAPI-owned types, paths and operations
  features/
    session/                   auth forms, session query and mutations
    workspace/                 current scope, access and feature visibility
    purchases/                 pages, components, forms, queries, mutations
    memberships/               member listing and explicit role changes
    audit/                     read-only activity history
  shared/
    ui/                        props/callbacks and semantic native controls
    lib/                       formatting, downloads and URL cursor history
```

API data lives in TanStack Query. Navigation, search/status and cursor history
live in TanStack Router's URL state. Forms own input/field feedback; menus and
other visual details use local state. Session and workspace providers contain
only their shared application context. No Redux/store duplication or persisted
auth/tenant state.

Pages compose feature hooks and presentational components. Components/forms never
import transport or query-cache internals. Cross-feature imports use public.ts;
shared UI cannot depend on feature code. The AST checker enforces imports,
re-exports, literal dynamic imports, direct network access, leaf-component
transitive data dependencies and browser-storage restrictions. It rejects data
effects in pages/components/forms; the reviewed useScopeGuard integration hook
uses cleanup only to suppress late UI effects after a scope changes. HeroUI
manages dialog/select focus, keyboard interaction and overlay portals. These checks
are technical boundaries, not a proof that every function has the right semantics.

## HeroUI profile

`pluto.toml` saves `ui = "heroui-v3"` and `template_version = 1`. Dependencies
match the CLI profile: @heroui/react and @heroui/styles 3.2.6, Tailwind CSS and its
Vite plugin 4.3.3, with the resolved graph in package-lock.json.

HeroUI provides buttons, text/search fields, selects/listboxes, purchase/member
tables, status chips, avatars, cards, alerts, spinners and the create-request modal.
HeroUI supplies its default dark palette, rounded controls and blue accent.
The theme mode and font are selected in the parent pluto.toml and applied to the
document root, including portalled overlays.
Application CSS handles layout and spacing; src/app/heroui.css does not reskin
HeroUI controls. Components import HeroUI directly; feature data/policy logic
stays in its existing modules. There is no second generic Pluto control API.

This example retains its original hand-authored schema-1 prototype manifest and
Go composition. This was an explicit UI migration, using the CLI profile's pins
and auth component as a base. The CLI's schema-7 generator does not overwrite or
implicitly migrate this existing application. For a new CLI-generated API use
`pluto frontend init --ui heroui-v3 --auth`; add supported controls afterward with
`pluto frontend add component purchase-form --kind form`.

## Contracts, access and consistency

`npm run api:generate` uses openapi-typescript against local Go OpenAPI shards.
It does no network schema resolution. The generator splits the resulting AST into
small modules by declaration/operation group. `npm run api:check` fails on drift.
Generated bindings are compile-time contracts, not runtime response validation.

GET /api/access describes this request's verified role, effective permission IDs,
and administrator role choices using the compiled Go RBAC policy. Frontend checks
only control presentation. Go still checks current membership, permissions,
resource ownership, feature gates and valid transitions on every business request.
No JavaScript copy of the role hierarchy.

Each data key includes account, session and organization. Account rotation,
session expiry, logout and organization switching cancel/remove private queries.
Switching organization remounts forms and resets URL paging/filter state. Late
navigation/download effects are suppressed after their original scope unmounts.
Previously authorized writes can still finish; UI cancellation is not a database
rollback or transactional revocation guarantee.

Writes have automatic retries disabled. Caller-owned keys are stable for a retry
of the same logical action; changed draft payloads receive a new key. Only the
API's declared idempotent business writes use them. Auth writes never use replay
keys. Auth inputs are transient form/mutation values; mutation cache retention is set
to zero after observers detach. This is not secure memory erasure. Session cookies
are never readable or persisted by application JavaScript.

Amounts are exact integer cents. Client form checks provide feedback; Go remains
authoritative. Summary cards, search and status filters describe the currently
loaded page, not organization-wide totals. Cursor pagination follows server
boundaries. Successful writes invalidate the captured scope. Export is available
only with the server permission and visible purchase_export flag.

## Checks and browser walkthrough

```sh
mise exec -- npm --prefix web run fmt
mise run web:check
# Backend and frontend checks together:
./scripts/check.sh
# Backend race tests without executing Go sources bundled inside npm packages:
go test -race ./cmd/... ./internal/...
```

Source/generated API files have a 1,000-line ceiling and 80-column formatter
preference; components should stay substantially smaller. CSS is split by visual
responsibility. The sole line-count exception is package-lock.json, a machine-owned
resolution graph retained in npm's standard format. Dependencies, compiled bundles,
local keys and binary resources are outside source limits. Exact package versions,
lockfile and ignored lifecycle scripts keep installs reproducible.

With the isolated Go API, HTTPS frontend and Mailpit above running:

```sh
mise exec -- npm --prefix web run test:e2e
```

The serial real-browser walkthrough enrolls/verifies the fixed synthetic
alice/bob/carol/dana@example.test accounts and uses normal recovery to choose
fresh random passwords for repeat runs. It revokes their old sessions. It then
calls the explicit trusted fixture command; enrollment never grants access.
Run only against a disposable local fixture environment. Passwords/challenges
stay in process memory and Mailpit; no storage-state, traces, screenshots or
video of credential steps are saved. Chrome is used through Playwright's installed
Chrome channel. The test contexts allow untrusted TLS for their hardcoded loopback URLs;
production browser verification stays enabled.

The walkthrough exercises real cookie sign-in/logout, member creation/submission,
independent approval, tenant isolation, URL filters, administrator role changes,
self-change restrictions, audit history and a mobile viewport. Pure tests cover
money parsing, both error envelopes, action-key headers, session refresh/cleanup,
late scope effects and architecture guard behavior. Go's disposable Postgres tests
also prove access descriptions update after membership role changes.

## Build and deployment boundary

`npm run build` creates static assets in dist/. `npm run preview` serves that build
on the same local HTTPS port with the same API proxy. It is a local verification
server, not a production deployment. Production should serve assets and proxy Go
behind one HTTPS origin, use explicit trusted origins, configure SPA fallback and
an HTML-appropriate CSP, and keep API routes separate from fallback. The Go example
retains its loopback-only development restriction. No production host, SSR runtime,
frontend generator changes are included in this example migration.

Manrope and Bricolage Grotesque are self-hosted under their included OFL licenses.

## Visual review

`npm run capture` uses the same isolated loopback fixture environment to capture
real API data at desktop, 375px and 320px. It recovers the synthetic Bob/Carol
accounts, seeds a few original purchasing requests through the API and saves
non-credential screenshots in evidence/. It also verifies dialog focus restoration
and viewport bounds. Like the browser walkthrough, it revokes old fixture sessions.
`npm run capture -- --open` leaves a nonpersistent Chrome review window open with
a synthetic administrator session; closing it releases that browser context.

## Central theme configuration

The parent pluto.toml now owns `[frontend.theme]`: mode, optional radius,
font_family and the project-relative overrides CSS path. Optional light/dark
subtables set semantic colors using #RRGGBB or #RRGGBBAA. The example keeps
mode="dark" and HeroUI's default colors; commented examples show where to start.
Font selection does not fetch assets; the existing Manrope font is self-hosted.

Run `npm run theme:generate`, or let Vite regenerate on start, build and TOML
changes. `npm run theme:check` detects stale outputs without writes and is included
in npm check. The compiler is the same implementation scaffolded by Pluto CLI;
it parses TOML at build time with the pinned smol-toml dev dependency. The example
still has its schema-1 prototype manifest and is not silently migrated by CLI.
Use the npm commands in this example; CLI theme commands support CLI manifests.

Generated files are src/generated/theme.css, src/generated/theme-overrides.css
and public/pluto-theme.js. Never hand-edit these. Library defaults precede the
generated tokens; web/src/styles/overrides.css loads last and belongs to the
application. The compiler never rewrites that file. Set a new overrides path only
after creating the file under web/src, outside generated directories. Unsafe
paths, symlinks, unknown tokens and malformed CSS values fail before writes.
Advanced selectors/effects/OKLCH colors remain available in authored CSS.

The external bootstrap selects light/dark before React mounts and follows OS
changes for mode="system". It needs no local storage, inline script or weaker
CSP. Browser assets receive only the theme mode and configured CSS tokens, never
the full project manifest or backend runtime settings.

## Published frontend state primitives

Purchase approval and decline now import `@plutoframework/frontend` **0.1.1** from
npm. [The flow](src/features/purchases/flows/decision.ts) defines viewing,
confirmation, submission, success and failure states. HeroUI controls subscribe
through the published React adapter; HTTP work still uses the existing feature
mutation layer and refreshes canonical TanStack Query data after settlement.

Selecting a decision opens a confirmation dialog. Cancellation sends no request;
duplicate confirmation cannot start a second write. A retry of the same verdict
retains its business idempotency key while each UI attempt has a separate
correlation ID. Request IDs do not provide backend idempotency or permission.

Controllers are created in committed integration-hook setup and disposed on
cleanup. StrictMode receives a fresh controller after replayed setup, and changing
purchase/account/session/organization fences and aborts old work. Cancellation
cannot undo a backend commit; refresh the purchase after an uncertain failure.
The backend remains authoritative for membership, permissions and business state.

Unit tests cover real HeroUI confirmation/cancellation, retries and duplicate
writes. Hook tests exercise every scope dimension under StrictMode, including
late responses from a transport that ignores cancellation. The real API/browser
walkthrough confirms approval through the new dialog. Bounded mutation smoke:

```sh
mise exec -- python3 scripts/frontend-decision-mutation.py
```

## Expanded source style

`npm run fmt` is the canonical formatter: it combines the 80-column Prettier
configuration with the ESLint fixes in `scripts/readability.mjs`. `npm run lint`
checks that same result without writing. ESLint requires braces, expanded control
blocks, one statement per line, blank lines around logical phases, and separate
lines for multiple object/type members. Nested ternaries must be rewritten as
named values or explicit branches; that semantic refactor is intentionally manual.

The formatter handles TypeScript-specific cases where plain Prettier would
compact a type again. Avoid running plain Prettier over TS/TSX as the final step;
use `npm run fmt`, then `npm run lint`. Editor integrations should apply ESLint
fixes as well. Edit `scripts/readability.mjs` and `.prettierrc.json` to change the
project's conventions. Generated API/theme artifacts retain their generators'
format and are type-checked. The 1,000-line ceiling still applies to source.

Standalone comment blocks have one blank line above them, except at line 1.
Consecutive comment lines stay together, and the comment remains attached to the
code below. JavaScript and TypeScript lint/format checks enforce this convention,
including the start of a function or branch. Inline trailing comments are kept
in place, and next-line compiler/linter directives retain their target. The
cross-language convention is also recorded in AGENTS.md for future changes.

## Immediate action feedback

Purchases use the published `/query` action helper with pending confirmation.
Writes show progress immediately, block conflicting resource actions and apply
confirmed response data to scoped detail/existing list caches before refreshing
in the background. Creating a draft navigates to its confirmed cached detail
without waiting for another round trip. Pagination boundaries are preserved.
A refresh failure is displayed separately from the write's outcome.

Uncertain results stay locked; refresh to establish authoritative state before
retrying. Stable backend idempotency keys remain owned by the feature. The
library's explicit optimistic projection policy is available for predictable
interactions, but this financial approval flow deliberately waits for confirmation.
Unit tests and the browser walkthrough hold writes open to assert immediate
feedback; scope, concurrency and rollback behavior are also tested in the library.
