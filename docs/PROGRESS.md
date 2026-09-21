# Corridor progress

Updated: 2026-09-21.

## Current state

The Phase 1 foundation now runs locally as a Go application with embedded SvelteKit, PostgreSQL/PostGIS/H3 and private storage. Independent review findings were fixed and local verification passed. The original brief is preserved in [BRIEF.md](BRIEF.md). No collision ingestion, trained model, hosted deployment or populated risk-map result is claimed.

| Phase | State | Evidence / remaining gate |
| --- | --- | --- |
| 0 Research | Research inventory and literature review written; source-level verification complete for 16 public/conditional source families | [RESEARCH.md](RESEARCH.md), [SOURCES.md](../data/SOURCES.md). Thirty ledger entries including restricted/unresolved candidates; only three raw artifacts acquired. Per-artifact schema/license clearance remains necessary before ingestion. This is not 16 production-ready datasets. |
| 1 Foundation | Complete for the approved Phase 1 scope; review fixes and local verification passed | [Verification evidence](verification/phase-1.md), [foundation design](superpowers/specs/2026-09-20-foundation-design.md), [implementation plan](superpowers/plans/2026-09-21-foundation.md). |
| 2 Ingestion | Not started | Need exact public-source artifacts, verified rights, Go adapters, object storage, privacy, idempotency and QA tests. |
| 3 Exploration | Not started | Depends on ingestion and shared public-release policy. |
| 4 Hotspots | Not started | Define/test all four methods and obtain a suitable WA comparison product with cleared terms. |
| 5 Features and ML | Not started | No licensed multi-state longitudinal training corpus yet. No evaluation scores or trained weights. |
| 6 Prioritization | Not started | Primary economic inputs, treatment-specific effectiveness and uncertainty still required. |
| 7 Field PWA | Not started | Offline reporting, auth, upload privacy and classifier selection/evaluation still required. |
| 8 Route risk/polish | Not started | Routing, public sharing, measured accessibility/performance and deployment still required. |

## Research verification

Checks executed successfully on the local research artifacts:

- `shasum -a 256 -c data/SHA256SUMS`: all 10 raw/evidence artifacts matched.
- `unzip -t data/raw/CA_State_Roadkill_2018_2025.zip`: archive integrity passed.
- JSON assertions confirmed Figshare v4 / catalog CC BY 4.0 and Oregon point geometry / `ANIMAL` field.
- Header-only California CSV inspection confirmed the documented fields; no private coordinate values were printed.
- Git exclusions keep `data/raw/` and `data/evidence/` out of tracked history.

These Phase 0 checks are research-artifact checks, separate from the Phase 1 application test evidence. No model parity test is applicable because no inference model exists.

## Material findings

- WSDOT raw data requires a request. Utah bulk reuse remains unconfirmed; classify conservatively as request-dependent.
- Montana DOI resolves, but the hosting repository failed over both HTTP/2 and HTTP/1.1. Its schema/license were not verified from raw bytes.
- California Figshare's catalog and archive licenses disagree. The archive lacks record-level rights/IDs, includes private-location columns, and has strongly concentrated coverage. Keep it quarantined.
- Oregon provides a real public collision service, but publication terms need clearance. Its metadata alone does not count as downloaded events.
- WSDOT public rankings are aggregates, suitable to investigate for comparison rather than fabricate individual training events.
- GHCN acquisition supplies real weather research data only. No park mortality coverage has been established.

## Next work

Finish independent foundation review, then scope Phase 2 against exact cleared artifacts. The public map currently has no basemap, collision records or risk model. MinIO remains an archived, isolated development companion; production storage requires a maintained service decision. No remote CI run is claimed.

During Phase 2, verify each exact artifact's rights and schema; promote source states only on evidence. No source requests have been sent, no accounts created, no paid services purchased, and no remote repository configured. Do not silently count unavailable data toward the required WA/CA/MT/UT and national-park coverage.
