# Movement, migration, connectivity and crossing data sweep

Verified 2026-09-21 UTC. This inventory identifies **18 release/study entries**, not 18 independent populations or production-ready datasets. M01–M06 overlap as a series; mirrors are not extra sources. **One raw research archive acquired; zero production imports or approved public track products.** Source metadata and downloads are distinguished below. Publication year is not observation year.

## Best acquisitions and honest visual behavior

1. **M08, northern Yellowstone winter elk:** acquired CC0 archive, 167,721 elk fixes from 55 unique source IDs. It offers actual historical playback after CRS/timezone, QA and privacy validation. Start with anonymous, aggregated elk movement, not wolves or persistent individual identities. The record also contains sensitive wolf tracks and predator kill sites; keep the whole archive private. These kills are not vehicle collisions.
2. **M01–M06, USGS migration releases, with a small M01 Pequop subset first:** actual mapped migration routes/corridors, seasonal ranges and stopovers are the strongest broad western overlay. Clear release-level CC0 evidence; review child metadata before acquisition. These support seasonal highlights and corridor/road intersection features, **not dots moving at observed times**. There is no GPS feed hidden behind the map. Volume 6 is available, published in 2026.
3. **M07, McKee Wyoming migration sequences:** best targeted future playback candidate: actual 2-hour fixes and a matched 12-hour subsample, including Jackson elk. Explicit CC0 and exact archive hashes are verified in the API. Website downloads returned 403 and anonymous API downloads returned 401 in this environment. Use the documented download workflow when available; do not call it acquired. M17's small CC BY crossing workbook is the alternative next acquisition when direct download is the priority.

Animation labels must distinguish **observed historical fixes**, **interpolated motion**, **seasonal habitat**, **camera detections**, and **modeled connectivity**. Break tracks across missing fixes; never animate a linearly interpolated highway crossing as an observation. Do not show control/available steps from selection models as real animal movements. No listed study is a live current-position feed or a validated collision predictor. Ranger filters can use species, herd/study, observation period, season, evidence type, precision and coverage; “not sampled” is different from zero animals or zero risk. The brief excludes individual tracking, so the public product should use herd-level/anonymous aggregation and remove persistent animal IDs.

## Release inventory

### M01–M06: USGS Ungulate Migrations of the Western United States

All six official release pages explicitly mark the work **CC0 1.0 Universal**, linking to [CC0 terms](https://creativecommons.org/publicdomain/zero/1.0/). This is release-specific evidence, not an inference from federal hosting. [USGS explains](https://pubs.usgs.gov/publication/sir20265123/full) that providers retain GPS data and supply derived layers; partner restrictions mean some report herds have no released data. Shapefile line/polygon products include corridors, routes, stopovers and seasonal ranges; classifications, resolution, survey dates and CRS vary by herd. Use released child layers only. Do not deduplicate different volumes solely by species name or add their herd counts together. Metadata API URLs below expose exact file URLs, sizes and child relationships.

| ID | Specific release and authoritative rights | Access/API | Geography/time and next step |
|---|---|---|---|
| M01 | [Volume 1, 2020; DOI 10.5066/P9O2YM6I](https://www.usgs.gov/data/ungulate-migrations-western-united-states-volume-1) | [ScienceBase JSON](https://www.sciencebase.gov/catalog/item/5f80c88d82cebef40f0fefc5?format=json) | Selected western herds; observation years vary. Full ZIP catalog size 32,766,735 bytes. Start with Pequop routes below. Metadata acquired; geometry not acquired. |
| M02 | [Volume 2, 2022; DOI 10.5066/P9TKA3L8](https://www.usgs.gov/data/ungulate-migrations-western-united-states-volume-2) | [ScienceBase JSON](https://www.sciencebase.gov/catalog/item/61fd7f6ed34e622189cf3fb9?format=json) | Selected western herds, separate release. Resolve exact child vintage/coverage and overlapping updates. DOI metadata acquired; file listing not acquired. |
| M03 | [Volume 3, 2023 data release; DOI 10.5066/P9LSKEZQ](https://www.usgs.gov/data/ungulate-migrations-western-united-states-volume-3) | [ScienceBase JSON](https://www.sciencebase.gov/catalog/item/63598d30d34ebe442503eb3f?format=json) | Western herds; report is dated 2022, data page 2023. Keep publication/observation timestamps distinct. DOI metadata acquired; child files to inspect. |
| M04 | [Volume 4, 2024; DOI 10.5066/P9SS9GD9](https://www.usgs.gov/data/ungulate-migrations-western-united-states-volume-4) | [ScienceBase JSON](https://www.sciencebase.gov/catalog/item/651f150bd34e44db0e2dd484?format=json) | Selected western/Tribal herds. Preserve local agency method differences and source attribution. DOI metadata acquired; select child files. |
| M05 | [Volume 5, 2025; DOI 10.5066/P1YJCCQA](https://www.usgs.gov/data/ungulate-migrations-western-united-states-volume-5) | [ScienceBase JSON](https://www.sciencebase.gov/catalog/item/6729962bd34e338a476a38ef?format=json) | Western herds, including cross-state analyses. DOI metadata acquired; inspect child notices and superseded herds before combining. |
| M06 | [Volume 6, March 2026; DOI 10.5066/P1ETBSYE](https://www.usgs.gov/data/ungulate-migrations-western-united-states-volume-6) | [ScienceBase JSON](https://www.sciencebase.gov/catalog/item/68dbf660d4be0204610b5ed7?format=json) | Selected western/Tribal herds, not universal coverage. Full ZIP 160,106,900 bytes: not downloaded. Metadata acquired; prefer small individual-herd packages. |

**Concrete M01 subset:** [Pequop mule-deer migration routes, Nevada](https://catalog.data.gov/dataset/migration-routes-of-mule-deer-in-the-pequop-mountains-nevada), [child API](https://www.sciencebase.gov/catalog/item/5f8db5c282ce32418791d554?format=json), [FGDC XML](https://data.usgs.gov/datacatalog/metadata/USGS.5f8db5c282ce32418791d554.xml). API identifies `NV_MD_Area7_Pequop_Routes_Ver1_2019`, MultiLineString, EPSG:5070, and component download links. Publisher describes 218 sequences from 79 collared animals, collected at variable 1–25-hour intervals. The released line geometry is not those timestamped fixes. Useful along I-80/US93 for corridor intersections and crossing context; do not invent event times. XML and JSON acquired. This is a subset of M01, not entry 19.

[Western Migrations](https://westernmigrations.net/) is a viewer for this family, not an additional dataset. [WAFWA SO3362](https://wafwa.org/so3362/) and its [Wildlife Movement and Connectivity Initiative](https://wafwa.org/initiatives/wmci/) provide state plans and program context; no separate open telemetry package was verified from those program pages. “SOFWA” did not resolve to a relevant primary dataset; WAFWA/SO3362 were the relevant verified sources.

### M07: McKee et al. — Estimating ungulate migration corridors from sparse movement data

[Dryad study](https://datadryad.org/dataset/doi:10.5061/dryad.15dv41p51); [metadata API](https://datadryad.org/api/v2/datasets/doi%3A10.5061%2Fdryad.15dv41p51); [file manifest](https://datadryad.org/api/v2/versions/305836/files). **CC0-1.0**, explicit API license; [Dryad reuse terms](https://datadryad.org/help/guides/reuse).

Wyoming Atlantic Rim, Clarks Fork and Dubois mule deer; Jackson elk. Published July 2024; actual year range not inspected. ZIP shapefiles provide spring/fall migration sequences at 2-hour fixes and 12-hour subsampling. README fields: State, Herd, species common/scientific names, AID, SeasonYr, Season, Year, Timestamp, UTM_E/UTM_N; NAD83(HARN)/UTM12N EPSG:3742; timestamps labeled Mountain Time. DST interpretation requires validation.

[2-hour download](https://datadryad.org/api/v2/files/3334126/download): 8,938,302 bytes; catalog SHA256 `2ea7316f75030d5ecef6f4dd2baebaabf66ac4576052383838c5754791201351`. [12-hour download](https://datadryad.org/api/v2/files/3334125/download): 1,539,661 bytes. Access failed here as above; metadata only. **Playback candidate**, not complete annual tracks. Cody elk and Idaho Tex Creek/Lowman/Northfork data are explicitly request-dependent; the public package must not imply their coverage. Next: lawful public-file acquisition, coordinate/date checks, anonymous aggregate playback.

### M08: Cusack et al. — Weak spatiotemporal response of prey to predation risk

[Dryad](https://datadryad.org/dataset/doi:10.5061/dryad.tp546d7); [official Dryad-community Zenodo mirror](https://zenodo.org/records/4984691); [API and explicit CC0](https://zenodo.org/api/records/4984691); [archive download](https://zenodo.org/api/records/4984691/files/Cusack_et_al_JAE_Data.zip/content).

Northern Yellowstone winter GPS data: elk 2012–2016, wolves 2004–2016, plus predator-kill sites and vegetation openness raster. **Acquired research-only**, not approved for public precision. CSV inspection found 167,721 elk rows (55 source IDs), 81,712 wolf rows (59 IDs), 1,805 kill-site rows. Fields: elk `Elk_ID,Winter,Easting,Northing,DateTime`; wolf adds `Pack`; kills `Year,Date,Easting,Northing`. Bare CR endings require care. Dates interpreted DD/MM/YY span November 2012–April 2016 for elk and January 2004–April 2016 for wolves. CRS, timezone, missing-fix intervals and duplication need validation.

Supports actual historical time playback after QA, with interpolations visibly distinguished. Predation sites must never become WVC labels. Wolves require the project’s H3≤6 public generalization and delay, including derived products; archive is old but age does not remove precision requirements. Start with elk-only aggregated output.

### M09: Gigliotti et al. — Elk avoidance of residential/agricultural land use

[Dryad DOI 10.5061/dryad.cvdncjt7x](https://datadryad.org/dataset/doi:10.5061/dryad.cvdncjt7x); [API](https://datadryad.org/api/v2/datasets/doi%3A10.5061%2Fdryad.cvdncjt7x); [README download](https://datadryad.org/api/v2/files/2185888/download); [fourth-order table](https://datadryad.org/api/v2/files/2185879/download). **CC0-1.0** explicit API license and [Dryad terms](https://datadryad.org/help/guides/reuse).

Greater Yellowstone, Idaho/Montana/Wyoming; 2023 publication; underlying years not verified. Study analyzed 765 elk from 21 herds. Released CSVs cover second/third/fourth-order selection and functional-response estimates. Main tables total about 2.36GB; no raw download. The existence of movement-path analysis does **not** establish unfiltered timestamped GPS in the archive. Metadata acquired; direct README returned 403. Useful research on development covariates and spatial scales. **Playback not established.** Next: inspect README via normal authorized download, distinguish used versus available observations, then consider selected small coefficient tables; do not ingest multi-GB files by default.

### M10: Xu et al. — Fencing amplifies individual differences

[Dryad DOI 10.6078/D1Q71K](https://datadryad.org/dataset/doi:10.6078/D1Q71K); [API](https://datadryad.org/api/v2/datasets/doi%3A10.6078%2FD1Q71K); [deer monthly CSV](https://datadryad.org/api/v2/files/2000731/download); [README](https://datadryad.org/api/v2/files/2000738/download). **CC0-1.0**, explicit API; [terms](https://datadryad.org/help/guides/reuse).

Wyoming 61 pronghorn and 96 mule deer; published December 2022; observation span unresolved. Four small monthly/survival CSVs plus README total 332,259 bytes. File manifest confirms summary products, **not a verified GPS-fix archive**. Useful for fence-exposure hypotheses, individual heterogeneity and survival research. No demonstrated map playback, no raw fence geometry confirmed, no causal effectiveness transferable directly to wildlife highway fencing. Metadata/file manifest acquired; raw not acquired. Next: read dictionary and survey period, then use only supported summarized covariates.

### M11: Poulin et al. — Sociality and elk crossing a major highway

[Dryad DOI 10.5061/dryad.wh70rxwzs](https://datadryad.org/dataset/doi:10.5061/dryad.wh70rxwzs); [API](https://datadryad.org/api/v2/datasets/doi%3A10.5061%2Fdryad.wh70rxwzs); [RDS download](https://datadryad.org/api/v2/files/3923918/download); [README](https://datadryad.org/api/v2/files/3923922/download). **CC0-1.0** explicit API; [terms](https://datadryad.org/help/guides/reuse).

Yoho National Park, BC, Canada; winter-range periods 2019–2020 and 2020–2021. RDS 1,387,471 bytes contains selected hourly traveling steps with at least two collared elk together; separate interaction CSV. README documents elkID, bioYear, Timestamp_MT, UTM12N X/Y, hwyCros, trafficVol, season, group/social metrics. **Observed-step playback candidate with intentional selection gaps**, not complete tracks. Strong crossing/traffic research comparison; crossing is not collision. Metadata only. Next: acquire and validate RDS conversion under Go-first contract, exact datum/timezone and omitted steps. Do not imply Canadian findings are calibrated US collision risk.

### M12: Prokopenko et al. — Behavioral responses to roads with step selection

[Dryad DOI 10.5061/dryad.t8v81](https://datadryad.org/dataset/doi:10.5061/dryad.t8v81); [API](https://datadryad.org/api/v2/datasets/doi%3A10.5061%2Fdryad.t8v81); [archive](https://datadryad.org/api/v2/files/3471/download); [README DOCX](https://datadryad.org/api/v2/files/3472/download). **CC0-1.0** explicit API; [terms](https://datadryad.org/help/guides/reuse).

Southwestern Alberta winter elk, 2007–2013; 175 elk-years analyzed. `ElkYrRelocations.zip`, 416,789,142 bytes, contains relocations and step/habitat covariates split by elk-year. Full field dictionary, coordinate reference and sampling interval not inspected. Metadata only; large archive not downloaded. **Potential historical playback**, conditional on preserving observed/control distinctions and time fields; otherwise methods/road-response research. Next: small README first, confirm actual fix schema before bulk acquisition. Canadian demonstration/training research only until local transfer is evaluated.

### M13: Movebank — Forage maturation in partially migratory elk

[DOI 10.5441/001/1.k8s2g5v7](https://doi.org/10.5441/001/1.k8s2g5v7); [repository landing](https://www.datarepository.movebank.org/handle/10255/move.489); [registered metadata API](https://api.datacite.org/dois/10.5441/001/1.k8s2g5v7). **CC0-1.0 registered explicitly**, [legal terms](https://creativecommons.org/publicdomain/zero/1.0/legalcode); preserve scholarly citation and [Movebank policy](https://www.movebank.org/cms/movebank-content/data-policy).

Canadian Rocky Mountains, growing seasons 2002–2004; study used VHF/GPS telemetry from 119 female elk. DOI metadata acquired; repository landing did not return a usable file catalog (753-byte response). **Archive filenames, byte sizes, downloadable file URLs and GPS time schema unverified.** Do not equate VHF and GPS frequency. Good historical seasonal demonstration candidate after access validation; not US park coverage or a live feed. Next: inspect the actual repository file manifest through its normal interface, verify archived subset and sample frequency. Status `access_failed` for files; not cleared on platform policy alone.

### M14: Movebank — Long-term Ya Ha Tinda elk migration variability

[DOI 10.5441/001/1.5g4h5t6c](https://doi.org/10.5441/001/1.5g4h5t6c); [repository](https://www.datarepository.movebank.org/handle/10255/move.1129); [metadata API](https://api.datacite.org/dois/10.5441/001/1.5g4h5t6c). **CC0-1.0 in DOI rights**, [legal terms](https://creativecommons.org/publicdomain/zero/1.0/legalcode); [Movebank study policy](https://www.movebank.org/cms/movebank-content/data-policy).

The registered archive title is `Data from: Study "Ya Ha Tinda elk project, Banff National Park, 2001-2020 (females)"`, Alberta. Its associated analysis describes ten years, 223 adult female elk and 630 elk-years; that ten-year analysis is not the whole archive’s stated 2001–2020 coverage. Actual archived fix-year bounds, interval and downloadable files remain unverified. DOI metadata acquired; file catalog access unresolved. **Potential long-term playback and migration classification research**, subject to actual schema; separate Canadian evidence, not Yellowstone. Next: acquire authoritative README and file list, establish overlap with M13 before combining. The same research program may reuse individuals/years; do not treat archives as independent test cohorts automatically.

### M15: Bohemian Forest multispecies road response

[Zenodo/Dryad archive](https://zenodo.org/records/10811816), DOI 10.5061/dryad.3xsj3txp9; [API](https://zenodo.org/api/records/10811816); [README](https://zenodo.org/records/10811816/files/README.md?download=1); [example roe-deer file](https://zenodo.org/records/10811816/files/roe_deer_females.rds?download=1). **CC0-1.0 explicitly on record**, [terms](https://creativecommons.org/publicdomain/zero/1.0/).

Germany/Czech Republic; roe deer, red deer, wild boar, lynx; 2024 publication, observation years not confirmed. RDS tables total 279.2MB; metadata and README acquired only. Fields include animal ID, used/available case flag, step length, angle, start/end timestamps, season, time-of-day and scaled road/habitat covariates. README does **not list endpoint coordinates**: no verified geographic playback. Strong global methods demonstration; show step/road-response charts, not invented tracks. Only case=true represents used steps. Lynx remains sensitive. Next: small file schema review for location fields if playback is needed; avoid importing random available steps as detections.

### M16: Theodore Roosevelt National Park bison density regions

[USGS release DOI 10.5066/P13M5MHT](https://www.usgs.gov/data/greatest-density-regions-gps-locations-bison-bos-bison-north-unit-theodore-roosevelt-national); [catalog](https://data.usgs.gov/datacatalog/data/USGS%3A66f1b5c6d34e0606a9dc845a); [ScienceBase API](https://www.sciencebase.gov/catalog/item/66f1b5c6d34e0606a9dc845a?format=json). **CC0-1.0** explicit release rights; catalog also records US public-domain label.

North Unit, North Dakota; observations September 2017–October 2020. Shapefiles `density_contours` and child `quantile_bands`: 25/50/75/99% kernel-density regions and nonoverlapping bands. Metadata acquired; geometry not acquired. **Static observed-use summary**, not timestamps or current bison whereabouts. Appropriate park habitat/use shading; percentile denotes nominal proportion of source locations, not chance of seeing a bison or collision probability. Next: inspect child files, projection and aggregation methods; acquire bounded park package and retain date badge.

### M17: US64 New Mexico wildlife-crossing detections

[USDOT supporting-dataset record](https://rosap.ntl.bts.gov/view/dot/61857); [Zenodo DOI 10.5281/zenodo.4273209](https://zenodo.org/records/4273209); [API](https://zenodo.org/api/records/4273209); [XLSX download](https://zenodo.org/api/records/4273209/files/19SAUNM03_Data.xlsx/content). **CC BY 4.0**, explicit API and [USDOT metadata document](https://rosap.ntl.bts.gov/view/dot/61857/dot_61857_DS1.pdf); [license terms](https://creativecommons.org/licenses/by/4.0/).

Two US64 under-bridge crossings near Lumberton, New Mexico; seven months of monitoring, published October 2020 (exact observation dates need workbook). `19SAUNM03_Data.xlsx`, 41,469 bytes. Publisher describes wildlife approach/passage monitoring supplemented with WVC counts. Only metadata acquired: no guarantee workbook contains every timestamp or photos. **Detection/activity charts candidate**, not continuous GPS movement. Next: inspect workbook sheets, dates, species, effort and counts; use crossing-specific temporal activity if available. Preserve attribution; do not reuse photo rights or infer causal reduction from narrative alone.

### M18: Florida Wildlife Corridor modeled connectivity

[USGS release DOI 10.5066/P132GNBL](https://www.usgs.gov/data/data-and-code-a-multi-species-framework-corridor-evaluation-species-different-dispersal); [metadata API](https://api.datacite.org/dois/10.5066/P132GNBL). **CC0-1.0** explicit release rights; inspect separately attributed inputs before reuse.

Florida, published June 2026; fox squirrel, Florida black bear, white-tailed deer and bobcat. Release contains inputs/code for species-distribution and spatial absorbing Markov-chain corridor evaluation. Underlying observation years, raster resolutions and exact files not inspected. DOI metadata acquired. **Modeled connectivity, never observed tracks**; relevant as a separately labeled methods example, not western coverage. GBIF input licenses/provenance still need inspection even if release-level output rights are CC0. Next: follow DOI landing and enumerate exact input/output artifacts, distinguish resistance/connectivity from probability of vehicle collision. Public sensitive-species derivatives require privacy review.

## Acquisition and verification record

Evidence root: `data/evidence/sweep-2026-09-21/movement/`. Raw root: `data/raw/sweep-2026-09-21/movement/`. Both are Git-ignored. [Acquisition manifest](../../data/evidence/sweep-2026-09-21/movement/acquisitions.json) records exact URLs, UTC retrieval/attempt timestamps, byte counts, SHA-256, kind and access status. Metadata bytes are original responses; generated aggregate inspection is labeled derived. Failed HTTP attempts are recorded and must not count as acquired files. Evidence licensing describes source documentation, not permission to relicense page HTML as application content.

| Acquired raw artifact | Source / retrieval | Bytes / SHA-256 | Scope and status |
|---|---|---|---|
| `data/raw/sweep-2026-09-21/movement/Cusack_et_al_JAE_Data.zip` | [Zenodo content URL](https://zenodo.org/api/records/4984691/files/Cusack_et_al_JAE_Data.zip/content), 2026-09-21T21:06:49.207Z | 12,210,769; `3396250548995e579d60baf6738f2be2df4b5f12eaa9e776abf9b1bf2d53a32c` | CC0; private local research only. ZIP integrity and publisher MD5 both passed. Contains sensitive tracks. |

The archive expands to about 80.6MB; the largest member is a 68.2MB vegetation raster. It was integrity-tested without extracting that raster. Only in-memory CSV aggregate inspection was performed; untouched ZIP remains the raw original. [Inspection summary](../../data/evidence/sweep-2026-09-21/movement/cusack-inspection.json) contains field names/counts and interpreted date bounds, not coordinates.

**Handling incident:** the first attempted header-only inspection assumed LF line endings. The files use bare CR; that mistake emitted coordinate-bearing content into tool output. It was corrected with CR/LF-aware parsing and a bounded alphabetic header check before output. No coordinates are reproduced in this report. The tool log should be treated as sensitive; it is inaccurate to claim coordinates never appeared in logs. No raw data were committed, published or ingested.

Production acceptance remains separate: source/child terms, CRS/datum/timezone, observation versus derived status, absent effort, spatial/temporal sampling bias, date and coordinate QA, project privacy across all public derivatives, immutable private storage and row provenance. Every file eventually used must enter `data/SOURCES.md`; the parent research task owns that shared ledger. These findings do not amend the project phase order or authorize a live tracking product.
