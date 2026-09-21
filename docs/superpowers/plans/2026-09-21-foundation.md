# Corridor Foundation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a self-hosted, tested Go binary serving an accessible SvelteKit empty map with real dependency health and source status.

**Architecture:** Static SvelteKit assets are embedded in Go and served on the same origin as a generated chi API. PostgreSQL/PostGIS/h3-pg stores provenance and spatial records; private S3-compatible storage holds originals. Dependency failures remain distinct from an empty database, and no raw research cache becomes public application data.

**Tech Stack:** Go 1.27.1; chi, pgx, sqlc, goose, oapi-codegen 2.8.0 with runtime >=1.6.0; PostgreSQL 16+ and PostGIS 3.4+; h3-pg; MinIO development companion; Svelte 5.57.1, SvelteKit 2.70.3, adapter-static 3.0.10, MapLibre GL 6.10.0; Tailwind; Vitest, Playwright, axe, testcontainers-go.

**Spec:** [Approved foundation design](../specs/2026-09-20-foundation-design.md). Read it together with [the user brief](../../BRIEF.md), [source ledger](../../../data/SOURCES.md), and [progress](../../PROGRESS.md).

## Global Constraints

- “PostgreSQL 16+ / PostGIS 3.4+ in Compose.” Verify h3-pg explicitly.
- “No Python/Node service in production.” TypeScript belongs to the frontend; all production server and CLI logic is Go.
- “API and tile paths never fall back to HTML.”
- “Until then, privileged endpoints remain disabled.” No analyst impersonation switch.
- “No collision data loaded” is the empty state; unknown coverage is distinct from zero risk.
- “Controls meet the brief's 48 px field target.” Support keyboard, screen reader, reduced motion, and light/dark palettes.
- “Go package coverage target at least 80% for implemented application logic.” Do not lower this target to pass a gate.
- “A fresh `docker compose up --build` must show the styled empty map at localhost.”
- Preserve raw data and its original licenses. No fabricated production observations, model scores, populated rankings, or benchmark results.
- Functions stay under 50 lines and files under 800 lines; immutable domain updates, validated boundaries, contextual errors, conventional commits without AI attribution.
- Work on an isolated feature branch/worktree at execution time. Never delete the ignored research downloads when making that worktree.

## Review Focus

1. A database outage must show “Data unavailable,” never an empty/zero-risk success (Tasks 3, 5).
2. Encoded traversal, missing JS assets, and unknown API routes must not receive the application HTML (Task 6).
3. Browser storage denial, WebGL failure, and reduced motion must leave navigation/coverage text usable (Task 5).
4. A clean ARM64 machine and an existing database volume must both start safely; missing h3 must fail migration/readiness clearly (Tasks 2, 7).
5. An unauthenticated client must not access raw event tables, object storage, source credentials, or privileged endpoints (Tasks 2–4, 7).

## Planning evidence and risks

Checked 2026-09-21: the workspace has docs only; no remote is configured. Installed Go is 1.26.6, Node 25.9.0, Docker daemon 28.3.3. The official Go download endpoint reports stable 1.27.1; build with that version without upgrading the host globally. npm registry reported the frontend versions above. Lock exact transitive dependencies during Task 1.

[oapi-codegen 2.8.0](https://github.com/oapi-codegen/oapi-codegen/releases/tag/v2.8.0) adds initial OpenAPI 3.1 support and requires runtime 1.6.0+. Validate the actual contract; do not silently down-convert it. [Static adapter](https://svelte.dev/docs/kit/adapter-static) supports prerendering: preserve server-rendered HTML and defer MapLibre to the browser. [Svelte CSP](https://svelte.dev/docs/kit/configuration#csp) supports static-page hashes; the nonce rule needs a documented static-build exception rather than `unsafe-inline` scripts. [h3-pg](https://pgxn.org/dist/h3/) must be installed with matching PostgreSQL development headers.

**Material dependency change:** [MinIO's repository](https://github.com/minio/minio) is archived; its latest release API reports `RELEASE.2025-10-15T17-29-55Z`, with AGPL-3.0 licensing. Preserve the requested MinIO development companion, isolated on the Compose network, with a verified build/image digest. Keep the Go client S3-compatible so production hosting can use a maintained service. Do not call the archived development dependency production-ready or change providers silently. Record this in ADR 0004 and the final phase limitations.

GitHub repository/code searches were repeated for SvelteKit/Go embedding and Tegola ST_AsMVT; those queries returned no matches. No suitable whole-project skeleton was established. Reuse official SvelteKit scaffolding and vetted libraries, with project-specific glue. Fontsource reports OFL-1.1 for Fraunces and Public Sans; retain each font's actual license on packaging.

## File map

| Area | Files and responsibility |
| --- | --- |
| Reproducible tools | `go.mod`, `go.sum`, `Makefile`, `web/package.json`, `web/package-lock.json`, `docs/DEPENDENCIES.md`: pinned versions and commands |
| Configuration/runtime | `internal/config/config.go`, `config_test.go`; `internal/server/run.go`, `run_test.go`; `cmd/corridor/main.go`: validation, lifecycle, executable wiring |
| Contracts | `api/openapi.yaml`, `api/oapi-codegen.yaml`, `internal/api/generated.go`, `contract_test.go`: authoritative schemas and generated interfaces |
| Database | `internal/db/migrations/00001_foundation.sql`, `migrate.go`, `store.go`, `queries/coverage.sql`, `sqlc.yaml`, `generated/`, `integration_test.go`: schema, queries, migrations |
| Public HTTP | `internal/api/router.go`, `health.go`, `coverage.go`, `middleware.go`, associated `_test.go` files: read-only API, protections, errors |
| Storage | `internal/storage/s3.go`, `s3_test.go`: private bucket availability, bounded SDK calls |
| CLI | `cmd/corridorctl/main.go`, `internal/cli/run.go`, `run_test.go`: migrations and local development environment initialization |
| Frontend | `web/src/routes/+layout.ts`, `+layout.svelte`, `+page.svelte`, `data/+page.svelte`; `web/src/lib/{map,coverage,theme,ui}/`; `web/src/styles/{tokens,global}.css`: accessible shell and source page |
| Embedding | `internal/web/embed.go`, `handler.go`, `handler_test.go`, generated `dist/`: production assets and safe serving |
| Deployment | `compose.yaml`, `deploy/Dockerfile`, `deploy/postgres.Dockerfile`, `deploy/minio.Dockerfile` if needed, `.dockerignore`, `.env.example`: local stack |
| Validation | `web/vite.config.ts`, `playwright.config.ts`, `tests/*.spec.ts`, `.github/workflows/ci.yml`, `.golangci.yml`, `web/eslint.config.js`, `web/.prettierrc`, `web/lighthouserc.json`: tests and reproducibility |
| Documentation | `docs/decisions/0002-route-contract.md`, `0003-static-csp.md`, `0004-storage-lifecycle.md`, `README.md`, `CONTRIBUTING.md`, `CODE_OF_CONDUCT.md`, `LICENSE`, `docs/PROGRESS.md`, `docs/verification/phase-1.md`: decisions, onboarding, measured evidence |

Dependency order: 1 → 2 → 3 → 4 → 5 → 6 → 7 → 8. Every task ends with its stated checks and a reviewable conventional commit. The final task produces the required phase commit and progress update.

## Task 1: Pin tools and establish the API contract

**Interfaces:** Public GET `/healthz`, `/readyz`, `/v1/coverage`, `/v1/sources`. JSON envelope contains `status: success|error`, `data`, `error` (string or null), and `meta` (pagination object or null). Coverage state is `empty` when count is zero and `unmodeled` when count is positive; `ingested_events` is a nonnegative integer and `model_available` is false in this phase. Fresh installation returns `{state: "empty", ingested_events: 0, model_available: false}`. Source list starts empty because no sources have been ingested. The research inventory remains documentation.

- [ ] Create the tool manifests and official SvelteKit static scaffold. Use local module name `corridor` until a real hosting path is chosen. Add `go 1.27.1`; record dependency versions/checksums and licenses, resolve Node's supported LTS image with registry evidence, and pin image digests during execution. Keep Go-only tooling commands in Go's tool module support and frontend tooling in npm scripts.

```sh
GOTOOLCHAIN=go1.27.1 go version
npm view @sveltejs/kit@2.70.3 engines --json
npm view vite engines --json
go get -tool github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.8.0
```

- [ ] Write the first contract test before the spec. In `internal/api/contract_test.go`, parse `../../api/openapi.yaml` with the pinned kin-openapi loader, validate it, require version `3.1.0`, and assert all four paths and response codes exist. Test rejects a schema whose required `status` property is removed. RED command: `go test ./internal/api -run TestContract -v`; expected missing-spec failure.

```yaml
openapi: 3.1.0
info: { title: Corridor API, version: 0.1.0 }
# Each operation references explicitly named envelope/data schemas.
# Null branches use type: 'null'; avoid permissive untyped data blobs.
```

- [ ] Define 200/503/429 responses, GET method rules, optional HEAD health semantics, and 404/405 envelopes. Write ADR 0002 accepting POST `/v1/risk/route` in Phase 8; do not register a pretend route implementation now. Keep the remaining brief endpoints in a clearly labeled future-contract section in the ADR, outside the active generated server contract.

```yaml
package: api
output: internal/api/generated.go
generate:
  models: true
  chi-server: true
  strict-server: true
  embedded-spec: true
```

- [ ] Generate with `go tool oapi-codegen -config api/oapi-codegen.yaml api/openapi.yaml`; compile and rerun contract tests. Add `make generate` and `make check-generated` (regenerate, then `git diff --exit-code` for generated paths). A missing/untracked generated file must also fail the check. Record initial OpenAPI 3.1 limitations only if this concrete contract encounters them.
- [ ] Review and commit: `chore: establish reproducible foundation contracts`.

## Task 2: Spatial schema and least-privilege queries

**Interfaces:** `db.Open(ctx context.Context, dsn string) (*pgxpool.Pool, error)`; `db.Migrate(ctx context.Context, dsn string) error`; `db.Store.EventCount(ctx context.Context) (int64, error)`; `db.Store.Sources(ctx context.Context, limit, offset int32) ([]db.SourceSummary, int64, error)`. `SourceSummary` has `ID`, `Name`, `URL`, `License`, `Status` strings. Queries read an approved public-source projection only.

- [ ] Build a PostgreSQL 16+/PostGIS 3.4+ development image with pinned h3 and h3_postgis packages/source. Record versions and architecture support. Do not assume the base PostGIS image contains H3. Use testcontainers-go with this same image and fail clearly if Docker is unavailable in the integration job.
- [ ] Write integration tests that fail before migration: required extensions, idempotent second migration, constraints, empty count, and rejected direct SELECT on events by the public role. Start a container without h3 and assert migration fails without partially applying the schema. RED: `go test -tags=integration ./internal/db -count=1 -v`.

```sql
SELECT extname FROM pg_extension
WHERE extname IN ('postgis', 'h3', 'h3_postgis');
SELECT has_table_privilege('corridor_public', 'events', 'SELECT');
-- Assert three extension names and false, respectively.
```

- [ ] Write a transactional goose migration and sqlc queries. Schema includes sources; immutable source_artifacts with source ID, object key, retrieval timestamp and SHA-256; species; road_segments `geometry(LineString,4326)`; events `geometry(Point,4326)` with time precision and coordinate uncertainty; and aggregate_observations with actual period/geometry/count. Use unique source-native record identifiers, nonnegative counts/uncertainty, positive segment length, foreign keys and GiST indexes. Keep raw JSON and precise locations private. No synthetic seed rows.

```sql
-- Public coverage is a scalar projection, not access to precise events.
CREATE VIEW public_coverage AS SELECT count(*)::bigint AS ingested_events FROM events;
REVOKE ALL ON events, source_artifacts, aggregate_observations FROM PUBLIC;
-- Create corridor_public as NOLOGIN; grant only the required public projections.
-- Application login is provisioned separately and assigned this role.
```

- [ ] Generate parameterized pgx queries with sqlc; enforce limits 1–100 and offsets >=0 at the API boundary. Return generic public errors while logging sanitized operation names and request IDs. Migration administrator DSN must not be supplied to the HTTP process.
- [ ] GREEN: rerun integration suite including migration on an existing volume and public-role privilege tests; review and commit `feat: add provenance-aware spatial foundation`.

## Task 3: Configuration, public HTTP, and process lifecycle

**Interfaces:** `config.Load(lookup func(string) (string, bool)) (config.Config, error)` returns a validated value; `api.NewRouter(deps api.Dependencies) http.Handler`; `Dependencies` contains `Store` (Task 2 interface), `Ready func(context.Context) error`, `Logger *slog.Logger`, and a shared rate limiter. `server.Run(ctx context.Context, listener net.Listener, handler http.Handler, logger *slog.Logger) error` owns serving/shutdown and returns fatal startup/serve errors.

- [ ] Write table-driven configuration tests: missing DB URL, invalid port, credential-bearing URL redaction, invalid timeout, and invalid environment. Require `CORRIDOR_DATABASE_URL`, `CORRIDOR_S3_ENDPOINT`, `CORRIDOR_S3_ACCESS_KEY`, `CORRIDOR_S3_SECRET_KEY`, `CORRIDOR_S3_BUCKET`; default HTTP address `:8080`. Reject embedded URL credentials in S3 endpoints. Bound timeouts to positive durations; never log config dumps.
- [ ] Add HTTP tests before handlers. RED: `go test ./internal/config ./internal/api ./internal/server -count=1`.

```go
func TestReadinessUnavailable(t *testing.T) {
    handler := NewRouter(Dependencies{
        Ready: func(context.Context) error { return errors.New("private-dsn") },
        Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
    })
    recorder := httptest.NewRecorder()
    handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/readyz", nil))
    if recorder.Code != http.StatusServiceUnavailable { t.Fatal(recorder.Code) }
    if strings.Contains(recorder.Body.String(), "private-dsn") { t.Fatal("secret leak") }
}
```

- [ ] Implement the generated strict interface and explicit nil-dependency errors; coverage/store errors produce 503, not zero. Test count 0 as empty and count 1 as unmodeled using labeled synthetic test fixtures only. `/healthz` checks the process; `/readyz` checks DB schema/extensions and private bucket under a bounded context. `/v1/sources` uses pagination and public projection only. Reject malformed pagination with 400. Add deadlines, request IDs, panic recovery, 1 MiB request-body cap and method checks. No trusted forwarded IP headers by default.
- [ ] Use a bounded process-level limiter (no unbounded per-IP map) on every endpoint; start with configurable 100 requests/sec and burst 200 for development. Test with burst 1 and immediate repeated requests: second request gets 429 and Retry-After. Distinguish listener readiness from dependency readiness; protect health endpoints without sharing user identity or credentials.
- [ ] Test shutdown using a canceled parent context and a real loopback listener, no sleep-based synchronization. Assert shutdown returns within its five-second deadline; bind failures propagate; a dependency check canceled by the caller exits. HTTP defaults: header timeout 5 s, read 10 s, write 15 s, idle 60 s, max headers 16 KiB. Use signal.NotifyContext in the small executable.
- [ ] GREEN/race/coverage: `go test -race -coverprofile=coverage.out ./internal/config ./internal/api ./internal/server`. Review and commit `feat: serve bounded public health and coverage API`.

## Task 4: Private storage and safe CLI lifecycle

**Interfaces:** `storage.New(cfg config.Config) (*storage.Client, error)` and `(*Client).Check(ctx context.Context) error` use a pinned S3 SDK and report bucket existence/access without returning keys. `cli.Run(ctx context.Context, args []string, stdout, stderr io.Writer) error` handles `migrate`, `init-dev`, and `help`. Future ingestion commands return a clear unsupported-command error.

- [ ] Write tests for canceled S3 calls, denied bucket access, unavailable service, unknown CLI command, and an already existing `.env`. `init-dev` must refuse to overwrite it. RED: `go test ./internal/storage ./internal/cli`.
- [ ] Implement credentials through environment configuration; retain TLS verification. Do not create/publicize buckets from HTTP handlers. `corridorctl migrate` uses the explicitly supplied migration DSN. A one-shot private setup command creates the bucket and least-privilege application credentials; verify rerunning setup is safe.
- [ ] `corridorctl init-dev` creates a local ignored `.env` with cryptographically generated credentials, file mode 0600 and exclusive creation. `.env.example` contains variable names and blank secret values. Never print secrets; report only the created filename. Test entropy source failure via injection, path exclusivity, and file mode.

```go
// Exclusive creation is required even if another process wins the race.
f, err := os.OpenFile(".env", os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
if err != nil { return fmt.Errorf("create development environment: %w", err) }
// Write generated values, check write/close errors; never use a fallback password.
```

- [ ] Integration tests run against the pinned MinIO companion and prove anonymous object access fails. Record AGPL notices/build provenance in ADR 0004. GREEN: `go test -race ./internal/storage ./internal/cli` and the tagged storage integration tests. Review and commit `feat: add private storage checks and development CLI`.

## Task 5: Responsive empty map and data-status UI

**Interfaces:** Zod-validated `CoverageResponse` matches Task 1. `loadCoverage(signal: AbortSignal): Promise<CoverageResponse>` fetches the same-origin API; malformed bodies and failures become unavailable UI. `readTheme(storage: Pick<Storage, 'getItem'> | null, systemDark: boolean): 'light'|'dark'` tolerates storage denial; `saveTheme` catches denied writes and leaves the active theme usable. `createMap(container: HTMLElement, theme: 'light'|'dark'): Promise<{destroy(): void}>` imports MapLibre in the browser and releases it on unmount.

- [ ] Add Vitest component tests for empty versus failed coverage, denied storage and keyboard-accessible theme control. Add initial Playwright tests for copy and dark-theme persistence. RED: `npm --prefix web run test:unit -- --run` and `npm --prefix web run test:e2e`.

```ts
test('an outage is not presented as zero collisions', async ({ page }) => {
  await page.route('**/v1/coverage', route => route.fulfill({ status: 503, body: '{}' }));
  await page.goto('/');
  await expect(page.getByRole('status')).toContainText('Data unavailable');
  await expect(page.getByText('No collision data loaded', { exact: true })).toHaveCount(0);
});
```

- [ ] Use static prerendering for `/` and `/data/`; fetch live status only after hydration. SvelteKit output goes directly to `../internal/web/dist`. Use Tailwind plus central CSS tokens. Configure static CSP hashes in ADR 0003, maintain strict script policy, and test MapLibre worker compatibility without broad origin wildcards.

```ts
// web/src/routes/+layout.ts
export const prerender = true;
export const trailingSlash = 'always';
```

```css
:root { --surface: #f6f3ea; --ink: #183c32; --muted: #526258; --accent: #315f50; }
:root[data-theme='dark'] { --surface: #122b25; --ink: #f6f3ea; --muted: #bdcbbf; --accent: #9ac3a5; }
:focus-visible { outline: 3px solid currentColor; outline-offset: 4px; }
@media (prefers-reduced-motion: reduce) { *, *::before, *::after { scroll-behavior: auto; animation: none; transition: none; } }
```

- [ ] Pair self-hosted Fraunces display with Public Sans UI, tabular numbers, and license notices. Measure contrast; change any failing token pair. Build map-first layered panels with desktop controls and a mobile bottom sheet. Use native buttons/details or tested focus management. Provide skip link, semantic headings, visible focus, readable coverage summary, and 48 px interactive targets.
- [ ] The local MapLibre style has a background layer and no invented geography, risk lines, or records. Show a visible empty-state explanation; no geolocation prompt. Search, season and drive planning are labeled unavailable until their phases, not deceptively active controls. The working “Data & methods” link opens the real source/methods page. No camera or GPS permissions are requested.
- [ ] Implement map error/fallback messaging, cleanup after navigation, theme synchronization and reduced-motion behavior. Test WebGL unavailable, storage getter throwing, 200% zoom, viewport widths 320/375/768/1024/1440/1920, and both themes. Assert no horizontal overflow. Run Chromium, Firefox and WebKit tests. Validate tab order and focus restoration manually.
- [ ] GREEN: unit/component tests, `svelte-check`, lint/format, build and browser tests. Review and commit `feat: add accessible map-first foundation shell`.

## Task 6: Embed assets and enforce safe HTTP serving

**Interfaces:** `web.Handler(assets fs.FS) http.Handler` serves only approved static paths. `web.Assets() (fs.FS, error)` returns the embedded dist subdirectory; `api.NewRouter` from Task 3 is mounted before the static handler by the executable. Use a pure fs.FS fixture in handler unit tests, never commit fabricated domain data.

- [ ] Write table tests for root/index, data route, valid JS, missing JS, GET/HEAD, POST 405, traversal, directory listing, unknown API and missing tiles. RED: `go test ./internal/web` after building static assets.

```go
for _, requestPath := range []string{"/v1/missing", "/tiles/events/0/0/0.mvt", "/_app/missing.js"} {
    rr := httptest.NewRecorder()
    h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, requestPath, nil))
    if rr.Code != http.StatusNotFound { t.Errorf("%s: %d", requestPath, rr.Code) }
    if strings.Contains(rr.Body.String(), "<html") { t.Errorf("HTML fallback: %s", requestPath) }
}
```

- [ ] Embed via `//go:embed all:dist` and `fs.Sub`. Serve only actual files/prerendered directories; no blanket SPA fallback is necessary for these two static routes. Do not list directories. `index.html` gets no-cache; hashed `/_app/immutable/` assets get a year plus immutable; other assets receive short explicit caching. Ensure HEAD is body-free and content types are correct.
- [ ] Set nosniff, frame denial, referrer policy, CSP and phase-appropriate permissions policy. Add HSTS only when deployed behind explicitly configured HTTPS; localhost remains functional. Static CSP hash generation must match the shipped HTML; never add script `unsafe-inline` to silence failures. Bound map worker/connect origins to actual assets/configuration.
- [ ] Build the binary, move/copy it to a temporary directory without `web/`, and test both pages and JS requests. Confirm it needs no Node process or source directory at runtime. GREEN: `go test -race ./internal/web` and browser suite against this binary. Review and commit `feat: embed and safely serve the static frontend`.

## Task 7: Compose deployment and fresh-install acceptance

**Interfaces:** root `compose.yaml` defines `postgres`, `minio`, one-shot `migrate`/storage setup, and `corridor`; public endpoint `http://localhost:8080`. Dependencies remain internal; only optional admin ports bind to 127.0.0.1. Application listens on container `:8080`, published as `127.0.0.1:8080:8080`.

- [ ] Write a Go integration acceptance test that invokes the built binary health probe and requests the root/coverage endpoints against the Compose project URL. Record RED before deployment definitions exist. Test success and dependency failure. Use an isolated project name/temporary volumes, never `down -v` against the user's existing project.
- [ ] Write multi-stage Dockerfile: pinned Node builder → Go 1.27.1 builder → non-root runtime with CA certificates and only the two binaries (`corridor` and `corridorctl`) plus embedded assets. Commit image digests resolved from actual registries. Exclude `.git`, `.env`, research caches, tests' reports and node_modules from build context. Include the PostGIS+h3 image from Task 2. Use secret interpolation `${VAR:?required}` in Compose.
- [ ] Start migration only after DB readiness, application only after successful migration and private bucket setup. Health probes use the Go binary's `healthcheck` CLI mode or another explicitly included tool, not a curl executable missing from the runtime image. Extend `cli.Run` tests to cover probe failure/nonzero exit. Use named volumes and bounded startup retries.

```sh
# After generating .env through corridorctl init-dev:
docker compose config --quiet
docker compose up --build --wait
curl --fail http://localhost:8080/readyz
curl --fail http://localhost:8080/v1/coverage
```

- [ ] Test restart with existing volumes, missing credentials, read-only filesystem where feasible, anonymous object denial, and application DB role denial. Verify ARM64 locally; build/test AMD64 in CI. Document any image platform constraint rather than marking unsupported platforms passed.
- [ ] Capture the actual light/dark desktop/mobile pages through Playwright against Compose. Save visual evidence under `docs/verification/` with no sensitive content. GREEN: Compose checks and integration acceptance tests. Review and commit `feat: package the self-hosted foundation stack`.

## Task 8: CI, measured quality, and phase completion

**Interfaces:** `make verify` runs the same generation/lint/test/build checks documented in CONTRIBUTING; `make integration` uses the tagged dependency tests; `npm --prefix web run test:e2e` accepts a base URL for the embedded server. Evidence report records commands, versions, dates and limitations.

- [ ] Write CI with least permissions (`contents: read`), pinned action commit SHAs and exact tools. Jobs validate contracts/generation, Go formatting/lint/vet/race/coverage/security, frontend lint/type/unit coverage, PostGIS/S3 integration, Docker build, embedded-server browser tests, and uploaded failure artifacts. No production credentials or automatic deployments.

```sh
make check-generated
npm --prefix web ci
npm --prefix web run check
npm --prefix web run lint
npm --prefix web run test:unit -- --run --coverage
npm --prefix web run build
go test -race -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
npm --prefix web run test:e2e
```

- [ ] Enforce >=80% coverage for authored Go application logic and frontend unit-testable logic. Generated API/sqlc code has a documented exclusion; thin main wrappers still receive subprocess acceptance tests. Add gosec and govulncheck plus frontend dependency audit; investigate findings rather than blindly upgrade/downgrade packages. Confirm type errors or generation drift actually fail CI with controlled temporary changes, then restore them.
- [ ] Add axe scans on `/` and `/data/` for desktop/mobile and both themes. Save Playwright snapshots at 320/768/1024/1440 widths; inspect them before acceptance. Test real keyboard navigation and screen-reader labels; record manual checks separately from automated pass counts.
- [ ] Run Lighthouse against production embedded assets with repeatable mobile/4G emulation and record median of three runs: public-page performance/accessibility >=90. Measure actual map-ready event separately (<2 s target); empty canvas timing cannot establish future loaded-map performance or 100k-point rendering. Record initial critical JS/CSS and lazy map chunk sizes; investigate exceeding app budgets (<300 KiB JS gzipped, <50 KiB CSS) without hiding deferred bytes.
- [ ] Add Apache-2.0 LICENSE for original code, font/dependency notices, CONTRIBUTING, CODE_OF_CONDUCT, tested 60-second quickstart (after prerequisites/image downloads), screenshots and troubleshooting. Preserve real-data limitations and archived MinIO development warning in the README. Do not claim CI ran remotely when no GitHub remote exists; run all jobs' commands locally and state remote CI is unexecuted.
- [ ] Review the whole branch and security boundaries after task-level checks. Fix all critical/high findings, rerun affected checks, record coverage and exact verification output in `docs/verification/phase-1.md`, and update `docs/PROGRESS.md` only to the proven state.
- [ ] Commit the phase result as `feat: complete verified Corridor foundation`. Report commit, launch command, screenshots, checks and unresolved dependencies. Do not begin Phase 2 by auto-importing quarantined or request-only sources.

## Self-review and handoff

Spec coverage: runtime/config/health (Tasks 1–4), private spatial/provenance foundation (2), accessible map shell and typography (5), static embedding/routing/CSP (6), reproducible local deployment and H3 (7), all named checks/documentation (8). Every Review Focus item has an explicit owning test. All public status copy preserves empty/unavailable distinctions. Production auth, ingestion, scoring and deployment to an external host are outside this phase.

Recommend native execution in this task with task-level review and a final independent branch review: the eight deliverables share API/schema/build interfaces, so keeping implementation context together should reduce coordination overhead. Use subagent-driven execution if separate implementation/review contexts per task are preferred. The user's global agent roles name models unavailable in this Codex toolset; use supported roles/models only, and do not invoke nonexistent model aliases.

Plan status: all eight tasks executed natively after user approval. Independent review findings were fixed; see [verification](../../verification/phase-1.md) and [execution record](../../verification/phase-1-execution.md) for exact evidence, deviations and remaining limits.
