# Corridor progress

Updated: 2026-09-21 UTC (2026-09-20 America/Denver).

## Current state

Research checkpoint committed separately from product implementation. The repository began empty. The original user brief is preserved in [BRIEF.md](BRIEF.md). **Corridor is not a running application yet.** No model, collision ingestion, deployment, or UX performance result is claimed.

| Phase | State | Evidence / remaining gate |
| --- | --- | --- |
| 0 Research | Research inventory and literature review written; source-level verification complete for 16 public/conditional source families | [RESEARCH.md](RESEARCH.md), [SOURCES.md](../data/SOURCES.md). Thirty ledger entries including restricted/unresolved candidates; only three raw artifacts acquired. Per-artifact schema/license clearance remains necessary before ingestion. This is not 16 production-ready datasets. |
| 1 Foundation | Design approved; implementation plan ready for review | [Foundation design](superpowers/specs/2026-09-20-foundation-design.md), [eight-task implementation plan](superpowers/plans/2026-09-21-foundation.md). Plan review and execution-method selection remain; product implementation has not started. |
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

No product unit, integration, E2E, coverage, Lighthouse, or model parity tests are applicable yet because Phase 0 contains no product code. Do not describe these research checks as application tests. Research uses shell retrieval/inspection commands; it adds no non-Go production implementation.

## Material findings

- WSDOT raw data requires a request. Utah bulk reuse remains unconfirmed; classify conservatively as request-dependent.
- Montana DOI resolves, but the hosting repository failed over both HTTP/2 and HTTP/1.1. Its schema/license were not verified from raw bytes.
- California Figshare's catalog and archive licenses disagree. The archive lacks record-level rights/IDs, includes private-location columns, and has strongly concentrated coverage. Keep it quarantined.
- Oregon provides a real public collision service, but publication terms need clearance. Its metadata alone does not count as downloaded events.
- WSDOT public rankings are aggregates, suitable to investigate for comparison rather than fabricate individual training events.
- GHCN acquisition supplies real weather research data only. No park mortality coverage has been established.

## Next work

Review the Phase 1 implementation plan and select native or subagent-driven execution. The user approved the written design. Planning checked official Go, npm, generator and database-extension documentation: Go 1.27.1 is the current stable target; oapi-codegen 2.8.0 adds initial OpenAPI 3.1 support, removing the anticipated conversion requirement. Docker is available locally. No product dependencies were installed.

MinIO's upstream repository is archived. The plan preserves the requested isolated local development companion and requires a verified image/build pin plus an explicit lifecycle ADR. Production storage needs a maintained service decision before deployment; no replacement was silently selected.

During Phase 2, verify each exact artifact's rights and schema; promote source states only on evidence. No source requests have been sent, no accounts created, no paid services purchased, and no remote repository configured. Do not silently count unavailable data toward the required WA/CA/MT/UT and national-park coverage.
