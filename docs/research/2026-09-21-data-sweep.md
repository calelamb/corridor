# Corridor multi-agent data sweep

Checked 2026-09-21 UTC. Four parallel research lanes covered US collision records,
animal movement/corridors, map/prediction covariates, and global roadkill/citizen
science/image data. The user requested a full search for data Corridor can obtain
and include in the ranger workbench. This is a broad, evidence-backed sweep, not
an assertion that every possible dataset was discovered or that all candidates
are cleared for publication.

## Sweep size and verified acquisitions

- **77 catalog entries:** 24 collision, 18 movement/corridor, 22 predictor/context,
  and 13 global/citizen-science/image entries. These include overlapping releases,
  mirrors, families and restricted leads, not 77 independent cleared datasets.
- **Eight raw dataset files**, 142,199,108 bytes total (~135.6 MiB), privately cached.
- **125 hash-verified artifacts** across raw data, documentation, error-response
  evidence and derived inspection summaries; 141 manifest records include failed
  retrieval attempts. An error-response hash is not a successful data download.
- **Zero production imports, accounts, purchases, custodian messages or accepted
  access agreements.**

## What the sweep changes

Corridor now has concrete raw research material for highway observations and
historical animal movement. The strongest route forward is to combine a cleared
collision pilot, a separately labeled movement study, authoritative migration
corridors and local road/environment context. **No new data were ingested into
the application, no public track product was approved, and no prediction model
was trained in this sweep.** Raw files remain private and Git-ignored.

### Most useful findings

| Feature | Concrete source | What we actually verified | What remains before use |
| --- | --- | --- | --- |
| Covered-highway evidence | Global Roadkill Figshare v5, G01 | CC BY 4.0; CSV downloaded and publisher checksum verified. 5,211 US rows; 2,200 rows name I-84 or I-84/86. | Row QA, aggregate multiplicities, precise highway footprint, study lineage and public release policy. Dominated by barn owls in that highway subset; not representative deer exposure. |
| Historical movement playback | Yellowstone Cusack archive, M08 | CC0; ZIP acquired/integrity checked. 167,721 elk fixes from 55 IDs, 81,712 wolf fixes from 59 IDs. | Confirm projected CRS/timezone and sampling gaps. Start with anonymous/aggregated elk. Keep wolf tracks and predation sites private; neither supplies collision labels. |
| Seasonal migration overlay | USGS western ungulate releases Volumes 1–6, M01–M06 | Release-specific CC0, child metadata and actual download routes. Volume 6 exists as of March 2026. | Select small herd packages, inspect child terms/CRS and deduplicate overlapping releases. Released corridors are derived geometry, not timestamped fixes. |
| Targeted Wyoming playback | McKee Wyoming/Jackson study, M07 | CC0; metadata/file hashes and 2h/12h sequence schema verified. | Direct files returned 403/401 here. Acquire through normal public workflow; no bypass or fabricated playback. |
| Mortality model experiments | Utah seasonal derived inputs, C11 | CC BY 4.0 model tables acquired. | Read exact response/exposure/control definitions; do not convert model support rows into individual observed collisions or assume raw collar-track rights. |
| Human-fatality crash context | NHTSA FARS 2023, C01 | Actual national CSV archive acquired and inspected. | Filter correct live-animal coding, retain fatal-crash selection and unknown species; do not treat as a carcass census. |
| Ranger park/road context | NPS boundaries + GNIS + OSM/Protomaps, P01/P02/P07/P08 | Source rights/access paths and current metadata checked. | Regional extracts and names/boundary ingestion; road analysis geometry is separate from simplified basemap tiles. |
| Historical prediction covariates | GHCN/gridMET, NLCD, 3DEP, hydrography, P09–P19 | Primary documentation, schemas, vintages and relevant reuse conditions checked. | Download bounded study-area products, handle no-data/units and align as-of dates to prevent hindsight leakage. |
| Species recognition | NACTI/Caltech Camera Traps, G11/G12 | Publisher-linked permissive data licenses, concrete image/annotation downloads and splits. | Select small licensed subsets and location-separated validation. These are image labels, not animal movement paths or collision ground truth. |
| International comparison | Scotland deer collisions, G09 | GeoJSON ZIP acquired: 16,383 points, 2008–2018. OGL v3 verified inside archive. | Accuracy/source QA and public generalization; label as Scottish evidence, not western-US coverage. |

## Recommended acquisition and integration order

1. **Build an auditable collision pilot from G01.** Preserve every occurrence ID,
   source/year/survey stratum and `numberOfRoadkill`. Distinguish discrete records
   from aggregates; retain accepted/rejected/duplicate/unsnapped counts. Use
   source-based road footprint evidence rather than tinting an entire state.
2. **Build a separate Yellowstone movement pilot from M08.** Parse the bare-CR
   CSVs safely, resolve CRS/timezone from study documentation, validate time
   ordering/gaps, and generate anonymous public products. This supports real
   historical movement exploration; it is not a live wildlife-location service.
3. **Add released USGS corridor packages and road/park context.** A small Pequop
   Nevada package is a concrete corridor starting point. Regional OSM geometry,
   Protomaps basemap, NPS boundaries and GNIS can supply the search/map experience.
4. **Join exposure and environmental features only in supported areas.** Caltrans
   AADT is a strong rights-documented lead after tying it to a year; use suitable
   local traffic releases elsewhere. Add stable land cover/elevation, one hydro
   edition and one historical weather grid first. Missing coverage remains null.
5. **Evaluate forecast feasibility on the actual overlap.** The initial collision
   and movement pilots are not automatically colocated, contemporaneous or the
   same species. Do not fuse them into a pretend training set. Define a specific
   target/region, check sample/effort support, and run spatial and temporal
   holdouts against a simple density baseline before enabling predictions.
6. **Expand through small well-documented releases, then permission-dependent
   sources.** Utah derived model tables, New Mexico crossing detections and
   additional USGS herd packages are concrete candidates. Request-only sources
   remain a backlog; no agencies were contacted by this sweep.

The original phase order remains ingestion → exploration → hotspots → models.
The requested ranger workflow remains the eventual acceptance target, including
real model execution. Research and working map controls alone cannot satisfy it.

## Corrections to the earlier inventory

- **Montana:** direct statewide 2016–2025 carcass spreadsheets are now discoverable;
  availability is materially better than the earlier failed university endpoint.
  Rights/redistribution and accuracy checks remain separate. Gallatin's ROSAP
  metadata also improves reuse evidence even where direct download still fails.
- **CROS:** its linked Creative Commons instrument can be identified more exactly;
  bulk export and applicability to a particular extract still need confirmation.
- **Utah:** published derived mortality/crossing model inputs can be reusable even
  while raw collar data remain permission-controlled. Neither grants blanket
  access to Utah Roadkill Reporter or the crossing/fence app's underlying layers.
- **PRISM:** current inspected terms permit redistribution with attribution under
  custom terms; the earlier generic unresolved label was too coarse.
- **Daymet:** current R1 collection 2129 covers North America from 1980; the 1950
  collection start belongs to Puerto Rico. Do not manufacture western-US history.
- **HPMS:** indexed 2018/2023 Washington endpoints returned semantic ArcGIS errors
  despite HTTP-success responses. A catalog hit is not a functioning dataset.
- **SNODAS:** documented erroneous-zero cells in 2014–2019 need the repair mask;
  otherwise “no snow” features would be wrong.
- **BC WARS:** reviewed 2026 terms prohibit redistribution and derived distributed
  products, even free/noncommercial ones. Excluded from Corridor distribution
  pending specific permission. No WARS data were downloaded.
- **Global Roadkill:** row count, summed animal count, version and GBIF survey
  classification differ. Figshare/GBIF mirrors are not independent extra data.

## Rights and scientific interpretation

Use the per-source status, not a blanket “open data” label. A data file may be
publicly downloadable while its redistribution is limited; a platform may host
both permissive and noncommercial records. A paper's open-access license does
not automatically license its linked data. Agency hosting alone is insufficient
where an individual release has third-party or custom restrictions.

Observed tracks, migration polygons, camera detections, collision events, aggregate
mortality and modeled resistance are different evidence types. Preserve their
units and provenance in separate products. An observed highway crossing is not
a collision; a predator kill site is not roadkill; unreported does not mean zero.

The first inferred prediction may be reporting intensity rather than true
collision probability if no reliable traffic/survey denominator exists. The UI
must use the model's actual target and support boundaries. No available source
alone establishes a reliable next-week crossing or collision forecast.

## Reports and reproducibility

- [US collision and agency sweep](2026-09-21-data-sweep-collisions.md)
- [Movement, migration and crossing sweep](2026-09-21-data-sweep-movement.md)
- [Roads, habitat, weather and predictor sweep](2026-09-21-data-sweep-predictors.md)
- [Global, citizen science and imagery sweep](2026-09-21-data-sweep-global.md)
- [Machine-readable artifact manifest](../../data/sweeps/2026-09-21/artifacts.json)
- [SHA256 verification list](../../data/sweeps/2026-09-21/SHA256SUMS)
- [Prioritized integration backlog](../../data/sweeps/2026-09-21/priorities.json)

Manifests distinguish raw datasets, documentation, derived inspection outputs and
failed attempts. Counts of release entries are not counts of independent datasets.
Exact public acquisition URLs are retained; short-lived signed redirect URLs are
excluded from committed records. Metadata hashes do not validate unseen data.
Downloaded artifacts are intentionally absent from fresh clones; checksum
verification must report missing files there rather than silently pass.

The movement report documents one inspection handling error: bare-CR line endings
caused coordinate-bearing content to enter a tool log. Subsequent inspection used
validated headers and aggregates. Those coordinates are not included in reports,
Git or the application. Treat that inspection log as sensitive.

## Verification

All 125 retained artifact checksums passed on the research workstation. All
manifest/report JSON parsed, all 77 source identifiers were accounted for, and
all retained raw/evidence files were confirmed Git-ignored. Committed exports
were checked for signed redirect credentials; summary links resolve locally.
`git diff --check` passed. This research-only change does not alter application
code or import data, so application tests were not rerun.
