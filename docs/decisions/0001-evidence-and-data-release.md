# ADR 0001: Evidence and data-release states

Date: 2026-09-21. Status: accepted for research; production enforcement pending.

## Context

The brief requires real, licensed, traceable datasets and protection of sensitive locations. Research found request-only sources, an inaccessible repository, a conflict between a catalog license and archive license, and aggregate layers that cannot support event-level inference.

## Decision

Track source discovery, artifact acquisition, license clearance, schema validation, ingestion, and public-release approval as separate states. A public URL is not an approval. Unknown or conflicting rights prevent publication and training. Downloaded research files remain unmodified and excluded from Git; they are not yet production ingestion artifacts. Future Go ingestion will copy approved originals to private object storage, record immutable hashes, and keep row-level provenance.

Record aggregated counts at their actual spatial/temporal grain. Do not synthesize dates, species, coordinates, or individual events from aggregates. Do not equate fatal crashes, observed carcasses, and all animal deaths.

## Consequences

Phase 1 can run with an honest empty map. Source access failures need visible statuses and must not cause fake coverage. The first trained model depends on obtaining licensed, sufficiently detailed data. The public release boundary must cover derived results, exports, tiles, photos, and live streams as well as event JSON.

## Alternatives

Automatically trusting portal license labels would be simpler but fails on the observed California archive conflict. Treating every source as raw events would simplify schema code but destroy provenance and validity. Both are rejected.
