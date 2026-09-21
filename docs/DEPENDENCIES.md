# Reproducible dependencies

Go 1.27.1 is selected by go.mod; Go's toolchain resolver installs it locally.
Go tool dependencies and runtime dependencies are pinned by go.mod/go.sum.
The API generator is oapi-codegen 2.8.0 (Apache-2.0), runtime 1.6.0
(Apache-2.0), chi 5.3.2 (MIT), kin-openapi 0.142.0 (MIT).

The frontend was created using official `sv` 0.17.1 minimal TypeScript scaffold,
with static adapter, Tailwind, ESLint, Prettier, Vitest and Playwright addons.
Direct versions are exact in package.json and transitive versions/integrity
hashes live in package-lock.json. Use npm ci. Node 24.21.0 is the selected LTS
builder; the host Node installation need not change. Svelte and SvelteKit are
MIT, TypeScript Apache-2.0, Tailwind MIT, Vitest MIT, Playwright Apache-2.0.
Image digests and additional runtime dependencies are recorded with deployment.

Generated API code is excluded from authored-code size and coverage limits;
never hand-edit it. Run make generate; make check-generated requires generated
files to be tracked and free of staged or unstaged drift. No OpenAPI version
conversion was needed for this concrete contract.

Database image: official multi-architecture PostgreSQL 16 bookworm digest
`sha256:efedf3595f1d6f415c08568ba171029bf54052e754cc9f030e3f2412b21f3d67`.
PGDG packages pin PostGIS 3.6.4+dfsg-2.pgdg12+1 and h3-pg
4.2.3-4.pgdg12+1 (includes h3_postgis). h3_postgis requires postgis_raster.
The package route follows https://pgxn.org/dist/h3/; upstream 4.5.0 exists,
but the selected PGDG package is 4.2.3. PostgreSQL uses the PostgreSQL license,
PostGIS GPL-2.0-or-later, h3-pg Apache-2.0. ARM64 image build verified locally;
AMD64 verification belongs to CI. Package repositories must retain pinned
versions; later updates are explicit reviewable changes.

Database clients: pgx/v5 5.11.0 (MIT), goose/v3 3.28.0 (MIT),
sqlc 1.31.1 (MIT); testcontainers-go is test-only (MIT). Exact versions and
transitive checksums are authoritative in go.mod/go.sum.

Security-tool pins: gosec 2.29.0, govulncheck 1.8.0 and goimports are managed
as Go tools. The madmin SDK's Prometheus dependency is explicitly upgraded to
0.311.3 to fix reachable vulnerability advisories found by govulncheck.
Lighthouse 13.5.0 replaces the older Lighthouse CI wrapper, whose transitive
archive/temporary-file dependencies had advisories. npm overrides cookie to
0.7.2, preserving its parse/serialize API while removing the <0.7 advisory.
Regression suites and npm audit validate these changes; no audit-force
downgrade is used. Brotli 1.2.4 (MIT) handles negotiated compression; static
variants are precompressed during the frontend build. Weighted Accept-Encoding
offers are normalized before chi because its matching ignores q-values.

The map module remains a browser-only dynamic import. The map page preloads
its build-manifest URL to avoid a hydration download waterfall; it counts in
the initial transfer budget. Zod validates exploration responses at the network boundary; the legacy
coverage client uses Zod Mini. The methods page does not preload MapLibre.

Fonts are subset at build time with subset-font 2.9.0 (BSD-3-Clause) using
HarfBuzz WASM. The English UI subset preserves ASCII and its punctuation;
other characters use system fallbacks. Public Sans retains weight range
400–700. Font copyright/license name records and complete OFL files remain.

MapLibre 6.10 uses its documented Vite `?worker&url` entry, producing a
self-contained same-origin worker; worker-src needs only 'self'. See the
[official Vite integration](https://maplibre.org/maplibre-gl-js/docs/).

Exploration adds PMTiles 4.5.0 (BSD-3-Clause) for same-origin range access and
`github.com/jonas-p/go-shp` 0.1.1 (MIT) for the pinned USGS shapefile adapter.
Map label PBFs retain Noto's OFL license and checksums in the source ledger.
The populated map currently exceeds the initial JavaScript and map-ready
budgets; see the exploration verification report.
