# Corridor Phase 1 foundation design

Status: approved by the user on 2026-09-21 UTC; implementation has not started.
Authority: [user brief](../../BRIEF.md), [research](../../RESEARCH.md), [source ledger](../../../data/SOURCES.md).

## Intended outcome

An agency planner, researcher, field reporter, or driver can open a fast, accessible Corridor shell whose map and source status clearly distinguish missing data from low risk. Phase 1 establishes the requested single Go server, static SvelteKit frontend, spatial database, private object storage, spec-first API contract, and automated checks. Later phases add verified data and analysis in the user's order.

Your brief supplies the stack, audiences, phase order, and quality criteria. The approved defaults below concern build mechanics, empty-state behavior, and API details. The implementation plan is tracked separately.

## Options and recommendation

1. **Recommended: the specified Go binary with embedded static SvelteKit and PostGIS.** One application origin, simple self-hosting, and a clear public/private boundary. Build frontend assets before the Go embed step. Native inference dependencies arrive in Phase 5 with a packaging ADR if needed.
2. Separate Go API and frontend services. Independent releases are easier, but deployment and authentication become more complex and conflict with the desired single-image experience.
3. Start with a GIS dashboard platform. Faster generic map scaffolding, but the required Go inference, reporting flow, and consumer-quality UX would require substantial adaptation.

## Runtime and repository

- Root Go module with `cmd/corridor` and `cmd/corridorctl`; small packages under `internal/` for configuration, API, database, and embedded web assets. No invented GitHub module owner: use a local module name until a real remote is selected, or document that choice in the implementation plan.
- PostgreSQL 16+ / PostGIS 3.4+ in Compose. Pin compatible image versions. h3-pg support must be verified in the image rather than assumed; build a documented database image if necessary.
- Private MinIO buckets and explicit development credentials from ignored environment configuration. No sensitive host ports exposed beyond loopback in the development default. Production startup fails on missing credentials.
- Separate migration invocation, bounded connection attempts, health/readiness probes, graceful shutdown, context deadlines, JSON `slog`, and bounded request bodies.
- Frontend build uses SvelteKit's static adapter. Go serves embedded assets with correct content types, cache policies, and safe routing fallback. API and tile paths never fall back to HTML.
- Docker multi-stage build produces one application image. PostgreSQL and MinIO remain companion services. No Python/Node service in production.

## API and data foundations

Keep `api/openapi.yaml` authoritative at OpenAPI 3.1. Verify the selected oapi-codegen version's 3.1 support before generation; if unsupported, record an ADR for a validated generation conversion without silently weakening the public contract.

Phase 1 implements liveness, readiness, and read-only source/coverage status. Reserved domain endpoints should not return pretend successful data. JSON uses the requested common envelope: status, nullable data/error, pagination metadata where relevant; tiles and streaming formats retain their native bodies.

Propose **POST `/v1/risk/route`** for a GeoJSON request body. The brief's GET-with-body design is unreliable across clients/caches. Persisted share links can use GET by opaque identifier later. This deviation needs an accepted ADR before implementation.

Migrations establish spatial types, source/artifact provenance, species, events and road segments with appropriate constraints and indexes. Represent date precision, coordinate uncertainty, and aggregated observations separately; do not force aggregates into events. Future sensitive tables are inaccessible to the public database role. Do not claim a full schema/security implementation merely from migration success.

Auth integration arrives with role-dependent features. Until then, privileged endpoints remain disabled. Phase 1 does not provide a fake analyst toggle. Plan OIDC state/nonce/PKCE validation, single-use magic-link tokens, hashed API keys, secure cookies/CSRF protection, and role-scoped cache behavior before implementing those flows.

## Map and interaction design

Full-height map canvas with a compact wordmark, floating search/control panel, season selection, accessible legend, and coverage summary. Use sage, sandstone, dusk blue and forest tokens; pale sand-to-ember risk colors paired with text labels. Define light and dark palettes and measurable contrast pairs. Display/UI fonts must have verified redistribution licenses; subset and self-host them with system fallbacks.

On an empty database, show “No collision data loaded” and the next available data action, with no glowing risk roads, invented counts, or working-looking controls for unimplemented features. A help panel explains what the map will show. Unknown coverage is visually distinct from a zero score.

MapLibre supplies the canvas. A configurable PMTiles basemap can be added from a verified licensed source; the initial empty map must render without a proprietary key or external tiles. A geography-free local style is an explicit empty state, not a claim to have downloaded a road network. Plan a list summary and WebGL-unavailable fallback. Attribution remains visible when a basemap is loaded.

Mobile panels use a bottom-sheet pattern without covering primary navigation. Controls meet the brief's 48 px field target, visible focus, keyboard access, reduced-motion behavior, and screen-reader names. No first-use geolocation prompt on the public map; later reporter mode requests location deliberately.

## Validation and completion gates

- RED/GREEN tests for configuration validation, routing/errors, method handling, cache headers, shutdown, and database readiness; Go package coverage target at least 80% for implemented application logic.
- PostGIS integration tests use testcontainers-go to execute migrations and verify spatial constraints. Separate tests check migration upgrade behavior and missing-extension failure.
- Frontend component tests exercise theme and empty/unavailable states. Playwright checks desktop/mobile navigation, keyboard focus, theme persistence, axe findings, and screenshots.
- CI runs Go tests/lint/build, OpenAPI validation and generation-drift checks, Svelte checking/lint/component tests, production embedding, and browser tests.
- A fresh `docker compose up --build` must show the styled empty map at localhost; readiness becomes healthy only after required dependencies are ready. Record the exact command, result and screenshot.
- Performance/accessibility measurements are recorded with device/network settings. Automated axe/Lighthouse results supplement manual keyboard, screen-reader and contrast checks; they do not alone prove WCAG conformance.

Phase 1 ends only after these applicable checks pass, `docs/PROGRESS.md` records evidence and limitations, and the changes are committed. Subsequent phases retain their original goals; this foundation is not a completed v1.

## Dependencies still outside Phase 1

Raw WA/UT requests, Montana availability, California license reconciliation, exact model training data, model weights licenses, discount/cost evidence, deployment account/domain, and any paid storage/hosting commitment remain explicit later dependencies. No messages, purchases, or external deployments have been authorized by this draft.
