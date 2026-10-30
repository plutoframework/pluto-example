<p align="center">
  <img src="https://raw.githubusercontent.com/plutoframework/pluto/main/assets/readme/banner.svg" alt="Pluto — a small framework, wide open possibilities" width="100%" />
</p>

# Pluto Example

A working purchasing app, from draft to independent approval.

[Website](https://plutoframework.github.io/pluto-site/) · [Documentation](https://plutoframework.github.io/pluto-site/docs/example/) · [CLI demo](https://github.com/plutoframework/pluto/blob/main/docs/DEMO.md) · [First issues](https://github.com/plutoframework/pluto/blob/main/CONTRIBUTING.md#starter-tasks)

## What you get

- Go API, PostgreSQL/sqlc and a React/HeroUI v3 frontend.
- Verified-email auth, scoped RBAC, feature flags and audit history.
- Jobs, cache, telemetry, typed agent review and responsive action feedback.

## Get started

### 1. Get the project

Requires Go **1.25.5+**, Node **24.22.1**, Python 3 and Docker:

```sh
git clone https://github.com/plutoframework/pluto-example.git
cd pluto-example
mise install
mise exec -- npm --prefix web ci
```

### 2. Start isolated development services

Use a fresh disposable database/Mailpit; the documented loopback ports must be free:

```sh
docker compose -p pluto-example up -d --wait
set -a
. ./.env.example
set +a
PLUTO_EXAMPLE_SERVER_ADDRESS=127.0.0.1:18080 \
  PLUTO_EXAMPLE_FEATURE_PURCHASE_EXPORT=true go run ./cmd/api
```

### 3. Open the full-stack application

In another terminal:

```sh
mise run web
```

Open **https://127.0.0.1:5173**. The local certificate is self-signed; no trust-store
change is made. Enroll/verify a synthetic account through Mailpit, then follow the
[frontend guide](web/README.md) for trusted fixture membership binding. Public
enrollment grants no business access. Do not reuse these settings in production.

For the complete API/worker walkthrough, migrations and checks, use the
[application guide](docs/REFERENCE.md) and [frontend guide](web/README.md).

## See Pluto in motion

[![Pluto CLI scaffolds an API and curl receives HTTP 200 from its public health endpoint](https://raw.githubusercontent.com/plutoframework/pluto/main/assets/readme/demo.gif)](https://github.com/plutoframework/pluto/blob/main/docs/DEMO.md)

Recorded from the real CLI. Compilation waits are shortened. The demo uses the
built-in `/healthz` route; business operations still need implementation and auth.
[Text transcript and static image](https://github.com/plutoframework/pluto/blob/main/docs/DEMO.md).

## Account authentication upgrade

This revision uses Pluto Auth v0.4.0. Account-backed deployments explicitly apply
`db/migrations/006_authentication.sql` after the existing migrations, before
running the upgraded application. Constructors perform no DDL. Existing sessions
keep unknown authentication evidence until the next real sign-in; rotation does
not make them fresh. See [the authentication guide](https://plutoframework.github.io/pluto-site/docs/backend/authentication/).

The library foundation adds account policy, method evidence and freshness guards.
Native social sign-in, TOTP enrollment and their frontend flows remain next work;
no factor policy is enabled automatically by this upgrade.

## Read further

- [API and integration reference](docs/REFERENCE.md)
- [Field guide](https://plutoframework.github.io/pluto-site/docs/example/)
- [Pluto ecosystem](https://github.com/plutoframework/pluto#the-ecosystem)

## Boundaries

This is a synthetic development application. Loopback/demo guards, explicit migrations, real delivery boundaries and tenant checks remain intentional. Optional agent review is off by default; no live model key is required to browse.

## Contribute

Start with [CONTRIBUTING.md](CONTRIBUTING.md), then choose a [good first issue](https://github.com/plutoframework/pluto/blob/main/CONTRIBUTING.md#starter-tasks).
