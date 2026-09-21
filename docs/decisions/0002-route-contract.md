# ADR 0002: Public foundation contract

Accepted 2026-09-21. OpenAPI 3.1 in `api/openapi.yaml` is authoritative.
The foundation serves four read-only GET operations: process health, dependency
readiness, publication-safe coverage and ingested public sources. Other methods
return 405 and unknown API paths return 404, both in the error envelope. HEAD is
not promised. Dependencies failing return 503, never an empty success.

All responses contain status, data, error and meta. Sources use bounded limit and
offset pagination. Every endpoint is rate limited and documents 429/Retry-After.

Future contract (not registered): Phase 8 route scoring will use POST
`/v1/risk/route` with a validated geometry body, avoiding URL-length constraints
and sensitive route coordinates in ordinary URL logs. Risk, tiles, observations,
exports, ingestion and analyst operations in the brief remain future work; this
contract makes no claim that those services exist.
