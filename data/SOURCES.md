# Source and artifact ledger

Checked: **2026-09-21 UTC** (2026-09-20 America/Denver).

This is a research inventory, not a claim of ingested coverage. `public_catalog` means the publisher documents access and a source-level reuse policy; it does not mean a particular file was downloaded or validated. `per_record` requires checking each observation/study and its media separately. `requires_request` is either publisher-confirmed or explicitly identified as a conservative project classification. `license_unresolved`, `license_conflict`, and `access_failed` prevent production ingestion/publication until resolved.

For every source below, **data retrieval date and data SHA-256 are `null` unless an artifact is explicitly listed in the acquisition table**. A page visit is not data retrieval, and an evidence checksum is not a dataset checksum. Exact files, versions, schemas, and third-party notices must be checked again at ingestion. Sources retain their own terms; Apache-2.0 will apply to Corridor code only.

## Expanded multi-agent sweep — 2026-09-21

The [full sweep and acquisition priorities](../docs/research/2026-09-21-data-sweep.md)
adds concrete collision, movement, corridor, predictor and image datasets beyond
this original 30-entry inventory. Domain reports distinguish verified source
policies, downloaded artifacts, failed endpoints and uncleared publication.
See the [sweep manifest](sweeps/2026-09-21/artifacts.json) for exact artifact URLs,
retrieval times, byte counts and SHA256 values. Downloaded originals remain in
ignored research storage; **none were ingested or published by the sweep**.

The original S01–S30 entries below are retained as the earlier research snapshot.
For current access/rights findings, consult the sweep, especially CROS, Montana,
Utah derived model inputs, PRISM, Daymet and updated migration releases. A new
permissive dataset does not clear another artifact from the same provider.

## Sources with confirmed public access or conditional reuse policies

Sixteen source families below have publisher documentation for access and reuse policy. Conditional platforms are not counted as cleared individual datasets. NHD, NHDPlus, and WBD are related products and must not be treated as independent ecological evidence.

| ID / source | URL and license evidence | Access status | Coverage / documented schema | Bias and integration notes |
| --- | --- | --- | --- | --- |
| S01 OpenStreetMap | [Data / ODbL 1.0](https://www.openstreetmap.org/copyright) | `public_catalog` | Global road ways, nodes, tags; regional extract still to select | Variable tagging/completeness; preserve attribution and assess derived-database obligations. Public tile servers are not free application infrastructure. |
| S02 NHTSA FARS | [Access](https://www.nhtsa.gov/research-data/fatality-analysis-reporting-system-fars); [copy/distribution permitted](https://www.nhtsa.gov/about-nhtsa/terms-use) | `public_catalog` | US fatal traffic crashes since 1975; annual crash/vehicle/person files and year-specific codebooks | Human-fatal crashes only; animal codes and unknown coordinate sentinels require edition-specific validation. No species-level inference from “animal.” |
| S03 FHWA HPMS | [Geospatial releases](https://www.fhwa.dot.gov/policyinformation/hpms/shapefiles.cfm); [catalog government-works license](https://catalog.data.gov/dataset/highway-performance-monitoring-system-hpms) | `public_catalog` | US annual road/traffic information; linear features, route/location measures, traffic attributes varying by release | Select exact vintage and state; traffic estimates and sampling are not uniform patrol exposure. |
| S04 USGS 3DEP | [Download](https://www.usgs.gov/the-national-map-data-delivery/gis-data-download); [free without use restrictions](https://pubs.usgs.gov/publication/fs20243056/full) | `public_catalog` | US elevation; DEM raster and lidar products at differing resolutions | Check vertical datum, units, acquisition year and no-data; terrain derivation remains Go/PostGIS. |
| S05 Annual NLCD | [Access](https://www.usgs.gov/centers/eros/science/annual-nlcd-data-access); [USGS public-domain policy](https://www.usgs.gov/faqs/are-usgs-reportspublications-copyrighted) | `public_catalog` | Annual conterminous-US land-cover science products beginning in 1985; classified raster | Verify collection/version and class dictionary. Land-cover classes are not percent canopy. AWS route may be requester-pays; choose a suitable documented route. |
| S06 NHDPlus HR | [Product and access](https://www.usgs.gov/national-hydrography/nhdplus-high-resolution); [USGS policy](https://www.usgs.gov/faqs/are-usgs-reportspublications-copyrighted) | `public_catalog` | US hydrologic network/catchments; feature geodatabases with identifiers and flow relationships | Match release, region and CRS; do not combine overlapping editions. |
| S07 NHD legacy | [Access and retirement notice](https://www.usgs.gov/national-hydrography/national-hydrography-dataset); [USGS policy](https://www.usgs.gov/faqs/are-usgs-reportspublications-copyrighted) | `public_catalog` | US streams/water bodies; vector flowlines/polygons and feature codes | Retired October 2023; historical reproducibility option, not the maintained future source. Evaluate 3DHP separately. |
| S08 Watershed Boundary Dataset | [Product/data model](https://www.usgs.gov/national-hydrography/watershed-boundary-dataset); [USGS policy](https://www.usgs.gov/faqs/are-usgs-reportspublications-copyrighted) | `public_catalog` | Hydrologic-unit polygons and hierarchical HUC identifiers | Useful stratification/context; not animal movement or stream centerlines. |
| S09 USGS National Transportation Dataset | [Access](https://www.usgs.gov/the-national-map-data-delivery/gis-data-download); [formats](https://www.usgs.gov/faqs/what-download-formats-are-available-boundaries-structures-and-transportation-data-products); [USGS policy](https://www.usgs.gov/faqs/are-usgs-reportspublications-copyrighted) | `public_catalog` | US road/transport vector products | Alternative geometry/reference candidate; inspect original contributor notices and data model before use. Does not replace AADT. |
| S10 PAD-US | [Product overview](https://www.usgs.gov/programs/gap-analysis-project/science/pad-us-data-overview); [USGS policy](https://www.usgs.gov/faqs/are-usgs-reportspublications-copyrighted) | `public_catalog` | US protected-area polygons with ownership/management/protection attributes | Overlapping designations need explicit dissolve/join rules; protection status is not habitat quality. |
| S11 USGS ungulate migration, Volume 1 | [Specific release and public-domain label](https://data.usgs.gov/datacatalog/data/USGS%3A5f80c88d82cebef40f0fefc5) | `public_catalog` | Western-US selected herds; migration corridors, stopovers, seasonal ranges | Selected collared herds, not complete species distribution. Release-level license confirmed; each child artifact and sensitive geometry still need review. Later volumes are separate acquisition candidates. |
| S12 NPS administrative boundaries | [Catalog and public-domain label](https://catalog.data.gov/dataset/administrative-boundaries-of-national-park-system-units-national-geospatial-data-asset-ngd); [official access guidance](https://www.nps.gov/subjects/gisandmapping/tools-and-data.htm) | `public_catalog` | US park-unit polygon boundaries and identifiers | Boundary overlap does not establish NPS ownership of an event or a park mortality dataset. Not survey-grade legal boundaries. |
| S13 NOAA GHCN-Daily v3 | [Access/use constraints](https://www.ncei.noaa.gov/access/metadata/landing-page/bin/iso?id=gov.noaa.ncdc:C00861); [schema](https://www.ncei.noaa.gov/pub/data/ghcn/daily/readme.txt) | `downloaded_subset` | Global station weather; a Bozeman station file and station inventory acquired below | NOAA distribution/citation terms apply; metadata lists citation and liability constraints rather than an SPDX license. Station siting and missing observations matter; retain QC flags and source codes. |
| S14 GBIF occurrences | [Licenses/access policy](https://www.gbif.org/terms); [Darwin Core publication formats](https://www.gbif.org/publishing-data) | `per_record` | Global occurrence records; Darwin Core archives/API, dataset and occurrence provenance | CC0, CC BY, or CC BY-NC per dataset. Use compatible records only; presence does not imply roadkill. Media have separate terms. |
| S15 iNaturalist observations and media | [Publisher license policy](https://help.inaturalist.org/en/support/solutions/articles/151000175695) | `per_record` | Global observations; taxon, date, public coordinates/uncertainty, record/media licenses | Default CC BY-NC is not an unrestricted commercial-use license. Filter record and photo rights independently; preserve IDs and attribution. Require positive roadkill evidence; do not infer death from road proximity. |
| S16 Movebank | [Study-specific policy](https://www.movebank.org/cms/movebank-content/data-policy); [terms](https://www.movebank.org/cms/movebank-content/general-movebank-terms-of-use) | `per_record` | Global animal tracking; study/individual/timestamp/location | CC0, CC BY, or CC BY-NC for public/repository studies, plus study restrictions. No selected study downloaded or approved. Do not expose precise collar tracks. |

USGS policy is a source-level basis, not permission to ignore third-party notices. The provider explicitly identifies exceptions: [access controls and copyrights](https://www.usgs.gov/data-management/access-controls-and-copyrights).

## Restricted, unresolved, or additional candidates

| ID / source | Verified evidence | Status / license | Coverage and next action |
| --- | --- | --- | --- |
| S17 WSDOT raw carcasses | [Publisher states request requirement](https://wsdot.wa.gov/construction-planning/protecting-environment/wildlife-habitat-connectivity) | `requires_request`; terms supplied with request, not an open license | WA state highways since 1973; maintenance and salvage sources. Schema not received. Request desired road ranges/years and redistribution/precision terms. No request sent. |
| S18 WSDOT WAHCAP 2025 safety rankings | [Live layer](https://data.wsdot.wa.gov/arcgis/rest/services/Shared/WAHabitatConnectivityActionPlan/FeatureServer/3) | `license_unresolved`; copyright WSDOT | Metadata acquired. Polygon features with `RouteID`, `BARM`, `EARM`, `AADT_1`, species counts, crash counts, rankings. Derived 2019–2023 inputs, not dated individual events. Review public disclosure risk for sensitive-species counts. |
| S19 CROS | [Public portal](https://wildlifecrossing.net/california/); [terms discussion](https://wildlifecrossing.net/california/support) | `license_unresolved`; page references Creative Commons without establishing a bulk dataset's exact terms | California volunteer/agency records. Downloading one's own records is documented; general export scope and exact license must be established. |
| S20 CHP SWITRS / CCRS | [CHP SWITRS page](https://www.chp.ca.gov/Programs-Services/Services-Information/SWITRS-Internet-Statewide-Integrated-Traffic-Records-System) | `requires_request` (project classification); raw export and reuse terms unconfirmed | California police crashes. Public statistical reports are documented. Confirm current request workflow and version-specific animal-related fields; do not assume old ISWITRS links work. |
| S21 Gallatin County, Montana | [Supporting dataset catalog](https://rosap.ntl.bts.gov/view/dot/78594); [DOI](https://doi.org/10.15788/1727734814) | `access_failed`; license unverified | Catalog describes CSV/TXT ZIP supporting 2008–2022 and 2018–2022 analyses. DOI redirects to ScholarWorks handle 1/18869; HTTP/2 protocol error and HTTP/1.1 empty response observed. Do not assert row schema, per-year coverage, or permission from the report alone. |
| S22 Utah Roadkill Reporter | [App](https://wildlifecollisions.utah.gov/); [agency explanation](https://wildlifemigration.utah.gov/stories/dwr-udot-release-new-app-to-report-roadkill/) | `requires_request` (conservative project classification); license unknown | Utah DWR/UDOT reports. No public raw export/license confirmed; no request sent. |
| S23 CA iNaturalist derivative, Figshare v4 | [Item](https://figshare.com/articles/dataset/CA_State_Roadkill_Observation_Data/29123012); [API](https://api.figshare.com/v2/articles/29123012) | `license_conflict`; catalog CC BY 4.0 versus archive CC BY-SA 4.0 | Archive acquired for inspection only. Publisher describes 2018–April 2025, over 33,000 records, strong Santa Clara/newt concentration. These counts are publisher claims, not an independent data-quality result. Quarantine pending license chain, record identifiers, and privacy review. |
| S24 Oregon wildlife collisions | [Public service](https://gis.odot.state.or.us/arcgis1006/rest/services/data_layers/wildlife_collisions/MapServer); [distribution terms](https://www.oregon.gov/odot/Data/Pages/GIS%20Data.aspx) | `license_unresolved`; informational use/distribution disclaimer, no explicit open license found in layer metadata | Deer/elk, Aug 2009–Dec 2019. Metadata confirmed `HWYNUMB`, `MP`, `ID`, `DATE`, `ANIMAL`, `LAT`, `LONGTD`, `YEAR`, `SEASON`; service uses EPSG:3857. Page size cap 1,000; future adapter must page and check truncation. |
| S25 Daymet v4 | [Product guide](https://daac.ornl.gov/DAYMET/guides/Daymet_Daily_V4.html); [services](https://daymet.ornl.gov/web_services); [NASA policy](https://www.earthdata.nasa.gov/engage/open-data-services-software/data-use-policy) | `license_unresolved` at selected-artifact level | North America daily 1 km weather, 1980 onward; CF NetCDF. NASA default policy has exceptions; inspect exact collection use constraints before clearance. SWE is not snow depth. Daymet's leap-year calendar needs explicit handling. |
| S26 PRISM | [Data](https://prism.oregonstate.edu/data/); [terms](https://prism.oregonstate.edu/terms/) | `license_unresolved`; custom terms must be reviewed for selected product/use | US gridded climate alternative. No download or blanket open-license claim. |
| S27 Colorado crashes / traffic | [Crash access](https://www.codot.gov/safety/traffic-safety/data-analysis/crash-data); [OTIS](https://dtdapps.codot.gov/otis) | `license_unresolved` | Inspect export terms, animal coding, coordinate quality, and traffic vintage. Portal existence is not adapter readiness. |
| S28 Idaho wildlife / crashes | [Wildlife program](https://itd.idaho.gov/environmental/wildlife/); [crash access](https://itd.idaho.gov/service/order-a-crash-report/) | `requires_request` (for a suitable bulk extract); license unknown | Separate public dashboards, individual paid reports, and research data. No purchase/request made. |
| S29 NPS mortality | [NPS data entry point](https://www.nps.gov/subjects/gisandmapping/tools-and-data.htm) | `requires_request` (fallback); no suitable park-specific mortality dataset verified | Search IRMA per park and event type. Boundaries/API alerts cannot substitute for mortality records. |
| S30 Wyoming / Nevada carcasses | No suitable licensed raw dataset verified in this pass | `unverified` | Continue state portal searches; do not claim unavailable solely because this pass found none. |

## What data we actually have

| Material | Local state | Permitted use in Corridor now |
| --- | --- | --- |
| GHCN station inventory | Untouched text, 132,503 physical lines | Research / future weather adapter input; no production ingestion yet |
| Bozeman GHCN station USW00024132 | Untouched `.dly`, 11,891 station-month-element lines | Weather schema verification; these are not collision events or daily row counts |
| California derivative archive | Untouched ZIP, private local research cache | License/schema investigation only; excluded from publication and training |
| WSDOT and Oregon layer metadata | JSON, no queried event features | Adapter planning and schema evidence only |
| Montana / WA raw / Utah / park mortality | Not acquired | No coverage claims |

## Acquisition record

All artifacts retrieved **2026-09-21 UTC**. Raw bytes remain in the ignored local research cache. They must be uploaded to private object storage with immutable keys during Phase 2 before they can become production inputs. The Phase 1 private storage foundation now exists, but these research artifacts have not been imported into it. Metadata/README downloads preserve evidence only and do not inherit the code license.

| Local artifact | Exact acquisition URL / origin | SHA-256 |
| --- | --- | --- |
| `data/raw/ghcnd-stations.txt` | https://www.ncei.noaa.gov/pub/data/ghcn/daily/ghcnd-stations.txt | `14c6467fcce6cce8dc0786eba8bbceed51a2f1c1056b0e221b573157559fc824` |
| `data/raw/USW00024132.dly` | https://www.ncei.noaa.gov/pub/data/ghcn/daily/all/USW00024132.dly | `ccb68fae80fcb5ee110f19937930ec32808a735e3a6db7168111613705cafd84` |
| `data/raw/CA_State_Roadkill_2018_2025.zip` | https://ndownloader.figshare.com/files/62181815 | `0ac0fcc9bf39cf58b5e4c4ff29a35bc993d484d56573ca0522dbf83cdefc1797` |
| `data/evidence/figshare-29123012.json` | https://api.figshare.com/v2/articles/29123012 | `f74d7c6665e0a89d63f4d6c1d04d63d0f33108942be6f9693250dc16e3776b6c` |
| `data/evidence/ghcnd-readme.txt` | https://www.ncei.noaa.gov/pub/data/ghcn/daily/readme.txt | `3a71ae355252a6693a56a473ed7182b500ae7816c2cb2b2ce708c1e9ea9dcfde` |
| `data/evidence/odot-wildlife-layer.json` | https://gis.odot.state.or.us/arcgis1006/rest/services/data_layers/wildlife_collisions/MapServer/0?f=pjson | `fca6e050410e859539d3ca3090f63e544b868dc35e7476e946a721f967780605` |
| `data/evidence/wsdot-safety-layer.json` | https://data.wsdot.wa.gov/arcgis/rest/services/Shared/WAHabitatConnectivityActionPlan/FeatureServer/3?f=pjson | `fdd693d46a26190567ad5f2144281b343325b6443b68593646cb883c450c4c5e` |
| `data/evidence/siriema-readme.md` | https://raw.githubusercontent.com/nerf-ufrgs/siriema/main/README.md | `e4b02d7840c0c5ba70380cda4f8d272f38c0d3fc36ed457e94223c39df0c76aa` |

Two additional evidence files are exact extractions of the California archive's `LICENSE` and `README.md`, not independent downloads. Their checksums, together with all above artifacts, are in [SHA256SUMS](SHA256SUMS). Verify from the repository root using `shasum -a 256 -c data/SHA256SUMS`. A fresh clone will lack ignored artifacts; absence must be reported, never silently treated as a successful verification.

The California CSV header is `observed_on,time_observed_at,quality_grade,latitude,longitude,positional_accuracy,private_latitude,private_longitude,public_positional_accuracy,scientific_name,common_name`. Header inspection does not establish whether private fields are populated. Never expose them through logs or public views.

## Ingestion acceptance policy

Each future adapter must pin the artifact/version, record source and retrieval timestamps separately, verify its hash, validate schema/CRS/units/sentinels, respect licenses and sensitive data, preserve raw bytes and row provenance, and produce accepted/rejected/duplicate/unsnapped counts. License-conflicted or unapproved sources are skipped with an explicit reason. A successful `ingest all` must not imply requested or inaccessible sources were imported.

## Approved exploration pilot — 2026-09-21

The local application now imports two artifact-level-cleared products through
`corridorctl ingest all`. The other research candidates are not implicitly
approved. Repeated import preserved 5,211 US records / 8,854 reported animals;
218 Pequop route geometries generate 163 public H3 areas. No trained model.

| Artifact | Exact acquisition URL and retrieval | License / coverage / schema | SHA-256 |
| --- | --- | --- | --- |
| Global Roadkill v5 CSV | https://ndownloader.figshare.com/files/53393273 · 2026-09-21T21:04:30Z | Grilo et al., CC BY 4.0; https://doi.org/10.6084/m9.figshare.25714233.v5. US subset 1983–2023, WGS84, occurrence IDs, date intervals, reported multiplicity and uncertainty. Full raw CSV preserved privately. | `ba005179e7f02bae33ba4b9374b6957cdf795f04a8ada773ef4887fe61a46c16` |
| Pequop migration SHP | https://www.sciencebase.gov/catalog/file/get/5f8db5c282ce32418791d554?f=__disk__e0%2F23%2F33%2Fe02333b80946b422fe4fae052e5a3ce19ac9d404 · 2026-09-21T21:38Z | USGS / Nevada Department of Wildlife, CC0, Volume 1 DOI 10.5066/P9O2YM6I. 218 polylines, 2011–2017 study period. NAD83 / Conus Albers EPSG:5070 confirmed by published XML and PRJ. Go reads shapes; PostGIS transforms CRS. No timestamps or animal IDs are published. | `07e52bd6218e59f4fccae43af48d2c24e052995a0827327a5fc8c0ed9aaa18bd` |
| Pequop PRJ evidence | https://www.sciencebase.gov/catalog/file/get/5f8db5c282ce32418791d554?f=__disk__55%2F2f%2F8d%2F552f8dc509e09e20768885f97391b0dfd70dfabf · 2026-09-21T21:38Z | Same CC0 release. NAD83/GRS1980; parallels 29.5/45.5, origin 23, meridian −96, metres. Local evidence retained. | `baa291f16322a5a674a2a202b8565016be7c5414ebc9b134d665957e608a3aa0` |
| Regional basemap extract | https://build.protomaps.com/20260921.pmtiles · 2026-09-21T21:28Z | Protomaps 4.15.2, ODbL Produced Work; OSM, Natural Earth and WorldCover attribution retained. Go PMTiles CLI 1.31.2, bbox −118,40,−109,46, zooms 0–10; ~11 MB, archive verification passed. Regional context, not analytical road segments. | `ca6ff0ff628140a8ae01992a47babdcde81e6a0e9c9835eaed1e5e8a1b51d9bc` |

Public release `h3-r6-month-v1` generalizes **all** collision locations to H3
resolution 6 and strips exact dates, raw IDs, row references and raw coordinates.
Intervals must end at least 30 days ago; uncertainty over 1,000 m is withheld
until a suitable coarser product exists. All current US records pass these gates.
A fixed cell/species/road/source/year/month product underlies counts, filters,
details and vector tiles; source withdrawal is checked on every request and
public evidence responses use `no-store`. Quantity >1 is explicitly stored as an
aggregate observation, not duplicated event points. All 5,211 records are currently
unsnapped: the map does not claim surveyed or precisely matched road segments.

Migration display is a separate static product: route geometries are transformed,
densified at 500 m, and generalized to H3 resolution 6. Study routes are not GPS
fixes; no animation or collision labels are inferred. Yellowstone raw telemetry
remains withheld pending coordinate/time validation and a track release review.

### Reproduce the local pilot

Preserve the pinned original files at these ignored paths:

- `data/raw/sweep-2026-09-21/global-roadkill-v5.csv`
- `data/raw/exploration/pequop.shp`
- `data/raw/exploration/pequop.prj` (CRS evidence)
- `data/raw/exploration/region.pmtiles` (explicitly public basemap product)

Acquire the first three from the exact URLs above using normal TLS-verified
HTTP downloads. Use the official PMTiles CLI for the bounded extract:

```sh
pmtiles extract https://build.protomaps.com/20260921.pmtiles data/raw/exploration/region.pmtiles --bbox=-118,40,-109,46 --maxzoom=10
pmtiles verify data/raw/exploration/region.pmtiles
docker compose build corridor
docker compose up -d
docker compose --profile tools run --rm ingest
```

Verify the pilot with `shasum -a 256 -c data/EXPLORATION_SHA256SUMS`. Daily provider URLs have limited retention;
if the pinned build expires, use a reviewed replacement acquisition with a new
checksum rather than silently substituting another build. Collision/migration
originals are copied to content-addressed private MinIO keys before database
import. The public basemap extract remains a separate read-only mounted asset,
never a general raw-directory HTTP route. Per-source database transactions are
idempotent; an error in a later source leaves earlier completed imports intact,
so the command exits nonzero and can safely be rerun.

Map labels use the Noto Sans Regular 0–255 PBF from
https://protomaps.github.io/basemaps-assets/fonts/Noto%20Sans%20Regular/0-255.pbf,
retrieved 2026-09-21. Font license is retained at
`web/static/licenses/noto-OFL.txt`; this is a UI font asset, not a wildlife dataset.

Label glyph artifacts, retrieved 2026-09-21 from the same Noto Sans Regular
endpoint (replace the final filename), are tracked as licensed UI assets:

| Filename | SHA-256 |
| --- | --- |
| `0-255.pbf` | `62c6d49b15fa836eb6aa45e259c7ca6762f44b011b09e47776efbe4a6db1b397` |
| `256-511.pbf` | `2eca7561f9f566bcacfda5dd04fb5880baec1328ec0f5484678289a13994de8a` |
| `8192-8447.pbf` | `8ea977a587352fe31b4159ffdbc9a40be79056f2472017c742ea1e4a931864b9` |
