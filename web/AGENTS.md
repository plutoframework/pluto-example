# Working on the purchasing frontend

Read README.md and the root API contracts first. This is a reviewable full-stack
example, not a frontend generator or a general component library.

- HeroUI v3 is the saved component library. Import its controls directly, use its compound components and theme CSS variables, and preserve the pinned profile. Do not restore parallel handwritten button, dialog or select implementations.
- Preserve HeroUI's defaults unless pluto.toml explicitly overrides them. Application CSS arranges content; do not reskin library controls with the earlier paper/navy/copper palette. Use secondary field variants inside cards and modals for visible contrast.
- React + TypeScript + Vite; TanStack Router owns URL state and TanStack Query owns server state. Keep transient visual state local and form state inside forms. Do not copy query records into global stores or persist auth/tenant state in browser storage.
- app/ composes the router/providers/shell. features/<name>/ owns pages, components, forms, queries, mutations and types. shared/ui receives props/callbacks. Cross-feature imports must go through that feature's public.ts; shared cannot depend on features.
- Pages assemble screens and call feature hooks. Components/forms never import the HTTP client or generated bindings. Only feature queries.ts/mutations.ts import api/client.ts. Never fetch data in component effects or add another HTTP library.
- Generated types come from local Go OpenAPI via npm run api:generate. Never hand-edit generated/ or rewrite authored files. Every generated module and source file stays below 1,000 lines; prefer components below 250 lines. The standard package-lock.json resolution graph is the sole line-count exception; inspect with npm tooling, not whole-file reads. Ignore node_modules, dist, .cert and binary resources when inspecting source.
- Go owns identity, permissions, resource policies and validation. access.permissions controls presentation only. Do not recreate RBAC role inheritance in JavaScript or treat a hidden button as authorization.
- Scoped queries include account, session and organization. Switching identity/organization cancels and removes old reads and resets authored forms. A 401 from a tenant route is not proof the account session is invalid; do not unconditionally log out on every 401.
- Cookies stay Secure, HttpOnly and host-only. The loopback HTTPS Vite proxy preserves Origin and cookie attributes. Do not disable Go Origin guards, rewrite cookies, add wildcard CORS, read cookies in JS, or persist tokens. Local cert keys are ignored and mode 0600. Never commit runtime credentials.
- Writes use explicit stable action keys only on endpoints declaring idempotency. Keep the same key for a retry of the same logical action. Auth writes never enter replay storage. Automatic mutation retries are disabled.
- Use semantic HeroUI controls, labels, dialog focus behavior, visible focus rings and reduced-motion CSS. Handle loading, empty, error and success states. Financial inputs use integer cents with exact decimal parsing; Go remains authoritative.
- Run npm run fmt, npm run check and the documented real API/browser walkthrough. Pure backend package checks target ./cmd/... ./internal/...; do not execute Go files bundled by npm dependencies.
- Native browser integration uses Playwright against loopback services only. The product-native preview cannot currently accept this self-signed local HTTPS certificate; use the documented isolated browser test context with ignoreHTTPSErrors enabled for this loopback example only. Keep traces/video/storage-state off for credential tests and do not print passwords, cookies or challenges.

Theme configuration lives in the parent pluto.toml. Never hand-edit
src/generated/theme*.css or public/pluto-theme.js. Custom selectors belong in
src/styles/overrides.css; regenerate with npm run theme:generate and verify with
npm run theme:check. Do not import the TOML document into browser code.

Use `npm run fmt` for the canonical expanded style, then `npm run lint` to verify.
Prettier alone is not the final formatter for TS/TSX. scripts/readability.mjs owns
mandatory braces, multiline blocks, object/type member expansion and logical
blank lines. Nested ternaries require explicit branches or named mappings.
Do not collapse code to fit the line ceiling or disable a rule to avoid splitting
responsibilities. Formatter regression tests run with the normal test command.
