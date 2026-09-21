# Contributing

Read AGENTS.md, docs/BRIEF.md and docs/PROGRESS.md. Keep original observations
out of Git and public test fixtures. Label synthetic test data explicitly.
Use test-first changes, small focused Go packages, validated boundaries and
conventional commits without attribution trailers.

## Reproduce checks

Use Go 1.27.1, Node 24.21.0+, npm and Docker. Versions and checksums are pinned.

```sh
npm --prefix web ci
make images
make verify
cd web && npx playwright install --with-deps && cd ..
go run ./cmd/corridorctl init-dev
docker compose -p corridor-foundation-acceptance up --build --wait
go test -tags=acceptance ./tests/acceptance -count=1
PLAYWRIGHT_BASE_URL=http://127.0.0.1:8080 npm --prefix web run test:e2e
npm --prefix web run performance
```

Acceptance tests intentionally stop and restart the database of the isolated
`corridor-foundation-acceptance` project. Never point these tests at production
or a personal project with the same name. `make integration` creates disposable
Testcontainers. Missing Docker is a failure, not a silently skipped check.

`make verify` validates generated contract/query code; checks format, Go vet,
frontend types/lint, race tests, per-package authored Go coverage >=80%, unit
logic coverage >=80%, gosec, govulncheck and npm audit. Build frontend assets
before standalone Go commands because the embed tree is generated and ignored.
Generated API/sqlc adapters are excluded from coverage, not from contract and
integration tests. Thin main wrappers receive executable acceptance checks.
Frontend TS helpers receive coverage gates; Svelte rendering and lifecycle are
verified with component and cross-browser tests rather than represented as
unit coverage of generated compiler output.

Generation must produce a clean diff from tracked files. Run `make generate`
and commit generated code with its source schema. Do not hand-edit generated
files. Performance measures a fresh empty canvas; do not extrapolate to loaded
collision layers or 100,000-point rendering.

## Review and security

Fix critical/high findings before merging. Keep administrator credentials out
of the HTTP service. Do not weaken CSP to permit inline scripts. Never import
a source until its exact artifact rights and sensitive-location release policy
are verified. There is no public upload or authenticated session in this phase,
so state-changing browser endpoints and CSRF-bearing flows remain disabled.

Report vulnerabilities privately to the repository owner through the hosting
provider's private reporting facility when one is configured. Do not attach
credentials, private locations or original photos to public issues.
