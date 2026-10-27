# Working in this example

This directory is the application prototype. Do not change sibling `pluto` or `pluto-cli` repositories as part of routine example work.

- Read README's code tour first. Keep technical responsibilities separate, with feature groups inside `internal/handlers`.
- Keep endpoint templates solely in `internal/paths`. Each endpoint uses the common expanded `Endpoint` definition with stable ID, display text, and authorization metadata. Paths never import handlers.
- Use explicit constructor injection and route registration. No magic `init`, service locator, ORM, hidden globals, or generic database framework.
- Handlers decode, extract verified context, call operations, and map responses/errors. Business policies live in `internal/operations`; core business code must not import HTTP or storage implementations. Never expose sqlc models directly as response models.
- Keep permissions, resource policies, feature flags, diagnostic logging, and business audit events distinct. Every repository call must retain the verified tenant. Every PostgreSQL write must include its business audit event in the same transaction.
- Runtime configuration is loaded and validated once in `internal/config`. `pluto.toml` is project metadata, not runtime config. Use the fixed uppercase `PLUTO_EXAMPLE_` prefix. Preserve explicit false versus unset feature values.
- Prefer descriptive names, expanded structs/composite literals, one statement per line, and small files. No file, including generated artifacts, may exceed 1,000 lines. Do not turn the sample into a framework or giant linter.
- Run `mise run fmt` (or `./scripts/fmt.sh`) for the canonical style, not just gofmt. The pinned golangci-lint configuration wraps breakable code at 80 columns, groups imports, and separates unrelated statements/completed blocks. Keep immediate error checks next to the call that produced the error; keep related registration calls grouped. Use additional blank lines where semantic phases need separation; a formatter cannot infer every logical boundary. Do not compress code to evade file limits.
- `mise run lint` checks without writing; `mise run check` runs the complete check script. Apply the same style to tests. sqlc-generated files are excluded from style rewriting to keep regeneration reproducible, but still pass tests, vet, and the file-length checks.
- Run `./scripts/check.sh` after changes. For policy/decoding changes, also run the bounded fuzz and mutation commands in README. SQL integration requires an isolated dedicated test database or the disposable Docker script; do not represent a skip as a pass.
- Regenerate sqlc and OpenAPI artifacts with the pinned commands in README. Inspect generated diffs. Add focused behavior tests with independent expectations; avoid test placeholders or tests that merely mirror implementation.
- Do not log credentials or bodies. Keep all fixtures original/synthetic. Demo authentication must remain explicitly opt-in and bound to loopback.
