# Contributing

Fork [calelamb/corridor](https://github.com/calelamb/corridor), create a focused
branch, and open a pull request against `main`. For substantial changes, open an
issue first to describe the user need and available evidence. Read AGENTS.md,
docs/BRIEF.md and docs/PROGRESS.md. Keep original observations
out of Git and public test fixtures. Label synthetic test data explicitly.
Use test-first changes, small focused Go packages, validated boundaries and
conventional commits without attribution trailers.

## Reproduce checks

Use Go 1.27.1, Node 24.21.0+, npm, Docker and golangci-lint 2.13.2.
Install golangci-lint from its [official release](https://github.com/golangci/golangci-lint/releases/tag/v2.13.2)
and verify the published checksum. Its build must support Go 1.27; older Homebrew
binaries can fail even when your installed Go is current. CI installs the pinned
version through the official action. `make lint` prefers a project-local
`bin/golangci-lint` when present, then falls back to `PATH`; binaries stay ignored.

```sh
npm --prefix web ci
make images
make verify
cd web && npx playwright install --with-deps && cd ..
npm --prefix web run test:e2e -- --workers=2
```

For executable acceptance checks, use a **separate disposable clone** with an
empty database, no personal `.env`, and port 8080 available:

```sh
go run ./cmd/corridorctl init-dev
docker compose -p corridor-foundation-acceptance up --build --wait
go test -tags=acceptance ./tests/acceptance -count=1
```

Acceptance tests intentionally stop and restart the database of the isolated
`corridor-foundation-acceptance` project. Never point these tests at production
or a personal project with the same name. `make integration` creates disposable
Testcontainers. Missing Docker is a failure, not a silently skipped check.

`make lint` builds the embedded frontend, validates `.golangci.yml`, and runs the
standard Go linters and gofmt/goimports checks over `./...`, including integration
and acceptance test files. It reports all findings, not just new code. Strictly
marked generated files and third-party npm packages are excluded from diagnostics;
authored tests are not excluded. The standard suite includes errcheck, govet,
ineffassign, staticcheck and unused. Override the executable when needed:
`make lint GOLANGCI_LINT=/path/to/golangci-lint`.

`make verify` includes this lint gate and validates generated contract/query code; checks format, Go vet,
frontend types/lint, race tests, per-package authored Go coverage >=80%, unit
logic coverage >=80%, gosec, govulncheck and npm audit. Build frontend assets
before standalone Go commands because the embed tree is generated and ignored.
Generated API/sqlc adapters are excluded from coverage, not from contract and
integration tests. Thin main wrappers receive executable acceptance checks.
Frontend TS helpers receive coverage gates; Svelte rendering and lifecycle are
verified with component and cross-browser tests rather than represented as
unit coverage of generated compiler output.

## Populated pilot and performance

Normal CI runs synthetic browser fixtures and empty-stack acceptance checks. It
never acquires wildlife datasets. After separately acquiring and verifying the
licensed pilot, use `compose.maps.yaml` as described in README.md, then run:

```sh
PLAYWRIGHT_BASE_URL=http://127.0.0.1:8080 npm --prefix web run test:e2e -- --project=chromium --workers=1 live-map
npm --prefix web run performance
```

These live checks require the exact populated pilot. Performance budgets currently
fail for map-ready time and JavaScript size; they remain visible development gates,
not an optional failure disguised as a green CI job. Keep the measurements and
limitations in `docs/verification/highway-exploration.md` current. A clean CI run
alone does not establish loaded-map performance, production readiness, or model
quality on new regions.

Generation must produce a clean diff from tracked files. Run `make generate`
and commit generated code with its source schema. Do not hand-edit generated
files. State which data and environment were measured; do not extrapolate to other
regions or 100,000-point rendering.

## Review and security

Fix critical/high findings before merging. Keep administrator credentials out
of the HTTP service. Do not weaken CSP to permit inline scripts. Never import
a source until its exact artifact rights and sensitive-location release policy
are verified. There is no public upload or authenticated session in this phase,
so state-changing browser endpoints and CSRF-bearing flows remain disabled.

Follow [SECURITY.md](SECURITY.md) to report vulnerabilities through GitHub private
vulnerability reporting. Do not attach
credentials, private locations or original photos to public issues.
