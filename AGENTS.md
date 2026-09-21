# Corridor engineering contract

Read `docs/BRIEF.md` for the complete user requirements and `docs/PROGRESS.md` before resuming work.

- Work through phases 0–8 in order. End each completed phase with passing applicable checks, an updated progress record, and a conventional commit. Do not add AI attribution to commits.
- Go implements all production services, ingestion, spatial orchestration, jobs, inference, tiles, and exports. TypeScript/Svelte is permitted for the frontend; Python only for offline training in `ml/`. Record any proposed exception in an ADR before implementation.
- Never invent production data, model quality, source licenses, or performance results. Synthetic data belongs only in explicitly labeled tests.
- Every acquired artifact needs its exact source URL, retrieval timestamp, license status, SHA-256, coverage, and schema notes in `data/SOURCES.md`. Preserve raw bytes. Public accessibility does not establish redistribution permission.
- Keep raw records, evidence downloads, photos, credentials, and model binaries out of Git. Ingestion will put originals in private object storage; the Phase 0 local research cache is not an ingestion system.
- Sensitive species require public location generalization to H3 resolution 6 or coarser and a 30-day delay. Apply privacy consistently to API, tiles, exports, SSE, photos, caches, and derived products. Only authorized analysts may access full precision.
- UX completion requires responsive layouts, keyboard and screen-reader access, WCAG 2.2 AA, light/dark modes, reduced motion, and measured performance. An empty or unavailable dataset must never appear as zero risk.
- Use test-first development, validate boundaries, preserve provenance, and keep functions/files small. Follow the user's global engineering rules and relevant language rules under `~/.claude/rules/`.
- Do not write, edit, run, or submit code in zyBooks activities that track coding trails; uncertain tracking means read-only.
