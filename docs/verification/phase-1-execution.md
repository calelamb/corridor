# Phase 1 execution record

Approved native execution of the eight-task foundation plan. Base `bc7af3f`; implementation `1e06b83`; review fixes `64f7de1`. All tasks executed inline, followed by one independent whole-branch review and one regression-tested fix pass.

## Rulings, in execution order

- Ruling: execute on feat/foundation worktree — approved plan requires isolation — changes remain on feature branch until integration.
- Ruling: use supported Codex tools for final independent review, implement all tasks inline — native choice overrides per-task agent instructions — no per-task independent reviewer.
- Ruling: application coverage must filter delayed/sensitive observations even at count level — exact raw counts could leak additions — fuller publication policy remains Phase 2.
- Task 2: Ruling: use PGDG h3 4.2.3 with PostGIS raster — packaged h3_postgis requires raster and ARM64 packages are available — larger database image; no application raster endpoint.
- Task 2: Ruling: provenance artifacts reject administrator UPDATE/DELETE as well as public access — originals must remain immutable — corrections require new artifacts.
- Task 1: Ruling: generated source is exempt from authored line-count and coverage limits — regeneration controls it — generator defects need contract/integration checks.
- Task 3: Ruling: executable dependency wiring moves to Task 4 — storage constructor is a Task 4 dependency — Task 3 verifies the HTTP runtime independently, with no fake readiness main.
- Task 4: Ruling: compile the pinned final MinIO release from checksum-verified source — its matching Quay image is unavailable — slower first build; archived lifecycle remains documented.
- Task 6: Ruling: standalone production-binary and browser acceptance runs with Task 7's real dependency stack — main correctly requires a real database and private bucket — no fake readiness path is added to production.
- Task 8: Ruling: use current standalone Lighthouse instead of @lhci/cli — wrapper transitives contain high advisories — equivalent repeatable three-run measurements use a local script.
- Task 8: Ruling: use Zod Mini and preload the browser-only dynamic map module on the map page — measured download waterfall missed the two-second gate — map bytes count toward the initial budget; methods page does not preload them.
- Task 8: Ruling: font/rendering/lifecycle checks use component and cross-browser tests; numerical frontend coverage covers authored TS helpers — compiler-generated Svelte branches are not unit logic — manual screen-reader testing remains a documented limitation.
- Task 8: Ruling: record remote CI and AMD64 as unexecuted — no remote exists and local host is ARM64 — architecture-specific failures remain possible until CI runs.
- Task 8: Ruling: H3-negative integration uses an explicit PostGIS-only build stage — plain PostgreSQL failed before reaching H3 — one extra disposable test image.
- Task 8: Ruling: subset the English UI fonts, preserving OFL/name records and system fallbacks for other glyphs — repeated map timing straddled 2s — additional character sets may need explicit subsets later.
- Task 8: Ruling: explicitly bundle MapLibre's worker using its documented Vite worker URL — v6's automatic sibling URL is unsuitable for hashed bundling — a deferred worker asset remains outside the empty-canvas benchmark.
- Task 8: Ruling: measure gzip budgets from actual deployed response bodies — host and Docker build hashes differ — local output must not be silently used to undercount deployed assets.
- Final: Ruling: public URL publication requires a separate reviewed landing-page field, without automatic acquisition-URL backfill — raw URLs can carry credentials; structural validation cannot recognize arbitrary secrets embedded in path text, so review remains required — existing sources stay absent from /v1/sources until a public URL is explicitly approved.
- Final: Ruling: reviewer did not execute remote CI or AMD64 — local ARM64 checks stand and no remote success is claimed — architecture-specific failures remain possible until CI runs.
- Final: Ruling: reviewer did not judge full WCAG — retain automated axe/keyboard/zoom/cross-browser evidence with manual screen-reader testing explicitly unperformed — untested assistive-technology behavior may need remediation.
- Final: Ruling: reviewer set aside loaded-map performance and future ingestion/auth/privacy products — Phase 1 delivers an empty canvas and restricted foundation only — later phases must remeasure and implement full release controls before public data publication.
- Final: Ruling: reviewer did not rerun destructive acceptance scenarios to preserve running services — author reruns the owned acceptance stack outage test and isolated fresh-initialization test after fixes — evidence is author-run rather than independently repeated.
- Final: Ruling: preserve feat/foundation and its worktree after verified implementation — native execution is approved; no remote is configured and integration/publication has not been requested — main remains the earlier research/docs checkout until the branch is integrated.

## Task and review evidence

- User approved design, plan and native execution. Base bc7af3f.
- Baseline: docs-only repository, clean; no executable tests existed.
- Pre-flight 1→3→5: envelope/coverage/source contracts consistent.
- Pre-flight 2→3: store methods must use public projection and bounded pagination.
- Pre-flight 3→4→7: ready check closes over DB and bucket; CLI provisioning distinct from HTTP credentials.
- Pre-flight 5→6→7→8: build Svelte assets before Go embedding and tests in fresh clone.
- Task 1: complete (commits bc7af3f..fec7867, tests: make test check-generated web-check → All matched files use Prettier code style!)
- Task 2: complete (commits fec7867..f503357, tests: go test -race -tags=integration ./internal/db -count=1 → ok  	corridor/internal/db	4.311s)
- Task 3: complete (commits f503357..037deb2, tests: go test -race ./internal/config ./internal/api ./internal/server → ok  	corridor/internal/server	1.514s)
- Task 4: private setup integration caught actual MinIO repeat-policy code; corrected to XMinioAdminPolicyChangeAlreadyApplied and reran real repeat setup.
- Task 4: complete (commits 037deb2..1cd9b4c, tests: go test -race -tags=integration ./internal/storage ./internal/cli ./internal/db -count=1 → ok  	corridor/internal/db	12.835s)
- Task 5: complete (commits 1cd9b4c..5073fbf, tests: npm --prefix web run test:e2e →   48 passed (14.3s))
- Task 6: complete (commits 5073fbf..74c9e75, tests: go test -race ./internal/web → ok  	corridor/internal/web	(cached))
- Task 7: complete (commits 74c9e75..66254a0, tests: go test -tags=acceptance ./tests/acceptance -count=1 → ok  	corridor/tests/acceptance	4.451s)
- Task 8: source-artifact integrity regression failed, then migration 00002 enforced source-matching foreign keys without rewriting the existing migration.
- Task 8: generation-drift and TypeScript-failure probes rejected controlled invalid changes; originals restored.
- Task 8: complete (commits 66254a0..1e06b83, tests: make verify → found 0 vulnerabilities)
- Final review: independent fresh-context reviewer found two Important issues; no Critical or Minor findings. Both severities stand because credential exposure and unreliable first startup affect real users.
- Final: fixed acquisition-URL exposure — TestAcquisitionCredentialsStayPrivate reproduced userinfo/signed-query leakage then passed; migration 00003 adds separately constrained public_url, keeps exact acquisition URL private and declines unsafe rollback.
- Final: fixed premature database readiness — TestFreshPostgresWaitsForTCP reproduced healthy socket-only initialization then passed after TCP probe; isolated project/volume is removed by test cleanup.
- Final verification after fix commit 64f7de1: make verify passed; 15 frontend unit/component tests; all authored Go packages >=80%; 3/3 acceptance tests; 60/60 browser tests against rebuilt Go; existing-volume migration 3 and healthy application confirmed. Security: gosec zero findings, govulncheck zero reachable, npm audit zero advisories.
- Deferred minors: none reported.

## Independent reviewer scope

The reviewer inspected the branch, plan, design and rulings; ran focused Go tests; and checked the actual PostgreSQL image entrypoint. Two Important findings were reported: public acquisition URL credentials and socket-only initialization readiness. Both were reproduced by tests and fixed in the single fix pass. There were no Critical or Minor findings. No re-review was commissioned.

Remote CI/AMD64, native screen-reader interaction, loaded maps and future products were not independently judged. The reviewer preserved the running application by leaving destructive acceptance tests to the author. Each scope decision is explicitly ruled on above.

## Retained workspace

Branch `feat/foundation` remains at `/Users/calelamb/Desktop/corridor/.worktrees/foundation`. No merge, push or hosted deployment was performed. The local application runs in the owned `corridor-foundation-acceptance` Compose project; its volumes are preserved. Only disposable integration/startup-test projects were removed.
