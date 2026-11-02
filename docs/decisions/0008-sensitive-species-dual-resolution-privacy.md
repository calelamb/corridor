# ADR 0008: Sensitive Species Dual-Resolution Privacy Architecture and Role-Based Access Control

Date: 2026-10-12. Status: Accepted.

## Context

Corridor ingests, models, and visualizes wildlife observations and wildlife-vehicle collision events across North America. Several recorded taxa—including endangered, threatened, or harvest-targeted species such as bighorn sheep (*Ovis canadensis*), grizzly bears (*Ursus arctos horribilis*), gray wolves (*Canis lupus*), Canada lynx (*Lynx canadensis*), and specific raptors—face severe poaching, trophy hunting pressure, harassment, and illegal commercial exploitation when precise GPS coordinates are publicly disclosed.

Simultaneously, state wildlife biologists, park rangers, tribal natural resource managers, and highway maintenance dispatchers require exact GPS coordinates, high-resolution incident maps, and real-time operational feeds to perform carcass salvage, deploy temporary variable warning signage, conduct field forensics, and prioritize infrastructure investments.

A monolithic data-release model cannot reconcile these competing mandates. Releasing unredacted data imperils vulnerable wildlife, while coarsening all internal records hobbles park ranger operations and DOT safety response.

## Decision

We implement a dual-resolution privacy gating architecture enforced through Role-Based Access Control (RBAC) in `internal/auth/` and API middleware in `internal/api/`:

1. **Role-Based Access Control (RBAC)**:
   - Define explicit user roles in `internal/auth/rbac.go`:
     - `RolePublic`: Default unauthenticated public consumers, citizen researchers, and general web visitors.
     - `RoleReporter`: Field staff and authorized contributors submitting roadkill observations.
     - `RoleAnalyst`: Verified park rangers, wildlife biologists, DOT safety engineers, and research partners.
     - `RoleAdmin`: System and security administrators.
   - Support authentication via OpenID Connect (OIDC) JWT bearer tokens (with signature verification, issuer, audience, and expiry validation) or cryptographic agency API keys (`X-API-Key`).
2. **Dual-Resolution Privacy Policies**:
   - Gated via `SensitiveFilterForRole(role)` and `PrivacyLevelForRole(role)`:
     - **Public Access (`RolePublic`, `PrivacyGeneralized`)**:
       - **Spatial Generalization**: Precise coordinates of sensitive species are generalized to Uber H3 resolution 6 hexagons (~36 km² cell area). Points and lines snap to hexagonal cell centroids; raw coordinates are never transmitted.
       - **Temporal Quarantine**: Sensitive species incident records are embargoed behind a mandatory 30-day delay window. Incidents younger than 30 days are omitted from public query results.
       - **Media Sanitization**: Uploaded photo attachments have all EXIF metadata (GPS latitude, longitude, altitude, timestamp, and device serials) permanently stripped upon ingest.
     - **Authorized Analyst Access (`RoleAnalyst`, `RoleAdmin`, `PrivacyFull`)**:
       - Unredacted GPS coordinates at full precision (H3 resolution up to 15 / meter-level geometry).
       - Zero temporal delay: immediate, real-time access to current and historical incidents.
       - Live Server-Sent Events (SSE) dispatch streams broadcasting incidents with raw coordinates as events occur.
3. **Comprehensive Defense-in-Depth Enforcement**:
   - The privacy filter boundary is strictly enforced across all outbound data surfaces:
     - REST API JSON endpoints (`/v1/segments`, `/v1/observations`, `/v1/hotspots`, `/v1/reports`).
     - Map tile services (`/tiles/` vector and raster layers).
     - Live SSE telemetry stream (`/v1/reports/stream`).
     - Data exports (CSV and PDF briefs).
     - Cached query layers and HTTP response caches.

## Consequences

- **Species Protection**: Eliminates poaching and harassment vectors by preventing public discovery of high-resolution locations of vulnerable species.
- **Full Operational Utility**: Equips authorized park rangers and DOT dispatchers with real-time, high-precision tools required for wildlife hazard response and corridor planning.
- **Consistent Governance**: Privacy rules are evaluated systematically at the middleware and query layer, eliminating accidental data leakage through secondary formats (tiles, exports, SSE).
- **Taxonomy Maintenance**: Requires maintaining an accurate, authoritative list of sensitive species flags in the database schema, with audit trails for any taxonomic classification changes.

## Alternatives

- **Complete Public Redaction**: Excluding sensitive species from all public views would create deceptive "zero-risk" blind spots on highway maps, misleading drivers and policymakers regarding collision frequency in known wildlife migration corridors.
- **Client-Side Blurring**: Sending raw coordinates to web browsers and obscuring them in frontend JavaScript is rejected; all spatial generalization and temporal filtering must occur server-side before serialization.
- **Fixed Bounding Box / Grid Blurring**: Rejected in favor of the Uber H3 discrete global grid system, which provides uniform hexagonal cell areas and avoids polar/latitudinal distortion.
