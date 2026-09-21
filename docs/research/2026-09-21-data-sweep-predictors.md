# Corridor data sweep: roads, map context and environmental predictors

Verified 2026-09-21 UTC. This is a research acquisition plan, not production coverage. Read together with `data/SOURCES.md`; source-level reuse evidence does not clear every artifact or downstream combination. No events, bulk road network, raster grid, photo, or production import was acquired here. Saved material is metadata, documentation and one astronomical API example, in the ignored `data/evidence/sweep-2026-09-21/predictors/` directory. The acquisition appendix records exact URLs, UTC retrieval-completion timestamps, bytes and SHA-256. An API error body is evidence of failure, never dataset coverage.

## Recommended acquisition stack

1. **Make the map useful first:** a bounded Protomaps PMTiles extract, OSM PBF road geometry, current NPS unit boundaries, and GNIS populated-place names. Keep the basemap visually distinct from collision/risk overlays and label the latter unavailable until data exist. Preserve OSM and applicable WorldCover/icon notices.
2. **Establish road exposure:** start with Caltrans AADT and its documented public-use rights; match route/postmile and ahead/back sections carefully. WSDOT has a useful 2023 line service, but reuse terms remain unresolved. FHWA's advertised 2018 and indexed 2023 WA endpoints both returned ArcGIS service-not-found errors in this pass; do not make them Phase 2 dependencies without repairing access.
3. **Build stable features for one study area:** 3DEP approximately 10 m elevation, Annual NLCD Collection 1.2 land cover/imperviousness, tree canopy as a separate product, and one pinned hydrography release. Use NHDPlus HR for reproducible historical coverage; test 3DHP coverage and lineage before replacement.
4. **Build historical weather once:** GHCN station QA plus one main grid (gridMET is easiest on explicit rights; Daymet offers 1 km detail but needs current policy-link resolution). PRISM's current custom terms expressly allow redistribution with attribution; it is a viable alternative. Add SNODAS only with its erroneous-zero repair mask and field-specific unit handling.
5. **Add mitigation and changing habitat carefully:** Utah crossing/fence app references are useful leads but the named public services were inaccessible. Seek installation/retirement dates, fence lines and ends, and surveyed coverage. MTBS supports lagged fire features, not real-time active-fire navigation. Add NWS forecast snapshots for future trip conditions; retrospective weather must not masquerade as an operational forecast.

None of these layers supplies collision labels or animal abundance. A plausible habitat map is not a validated animal predictor. Missing AADT, unmapped fence, no station report and absent collision report must each remain unknown with their own coverage flags.

## Road network, exposure and interventions

### P01 — OpenStreetMap through Geofabrik

- **Access:** [Utah extract directory](https://download.geofabrik.de/north-america/us/utah.html), [direct PBF](https://download.geofabrik.de/north-america/us/utah-latest.osm.pbf); equivalent state directories cover WA, CA, MT and others. Global coverage, frequently updated snapshots, with dated older files in the directory. Pin a dated file and OSM replication timestamp rather than retaining a mutable `latest` URL alone.
- **Schema/use:** PBF nodes/ways/relations and tags supply road centerlines, class, `maxspeed`, `lanes`, `oneway`, bridges, tunnels, access and surface; geometric curvature and junction density can be derived. Speed values may carry units/conditional clauses; tags are not uniformly present or verified. Free shapefiles omit some original tags, so PBF is preferable for analysis.
- **Rights/cost:** [Geofabrik terms](https://www.geofabrik.de/data/download.html) explicitly identify ODbL 1.0, attribution and derived-database obligations. Downloads are free; storage/hosting are Corridor costs. State files exceed this sweep's 50 MB cap; none downloaded. Public extracts omit contributor personal metadata.
- **Readiness:** source-level ready. Download bounded states in Phase 2, assess road continuity/topology, attach vintage and missing-tag indicators, and separate OSM-derived database obligations from independently licensed event records. Today's geometry can leak future construction into historical evaluations.

### P02 — Protomaps vector basemap / PMTiles

- **Access:** [daily builds](https://maps.protomaps.com/builds/) and [download/extract documentation](https://docs.protomaps.com/basemaps/downloads). Global zooms 0–15; approximately 120 GB for a planet archive. Use the documented Go PMTiles CLI to extract the study bbox/zoom range, then self-host; vendor discourages hotlinking. Daily retention is short, so preserve selected bytes and build metadata.
- **Schema/use:** MVT layers in a PMTiles archive provide labels, road drawing, land/water and context. This is a display product; tile simplification/clipping is unsuitable for road segmentation or exposure joins.
- **Rights:** publisher calls the basemap an ODbL Produced Work. The saved [data-license file](https://github.com/protomaps/basemaps/blob/main/LICENSE_DATA.md) also identifies OSM coastlines, public-domain Natural Earth, CC BY 4.0 WorldCover-derived landcover when displayed, and MIT Mapzen icons in default styles. Display all applicable notices, not just OSM attribution.
- **Readiness:** source-level ready; no map downloaded. Measure actual regional size/render performance after extraction; no performance claim here.

### P03 — FHWA HPMS / ARNOLD

- **Access:** [official public release](https://www.fhwa.dot.gov/policyinformation/hpms/shapefiles.cfm) is explicitly **2018**, despite its undated URL. It lists state services including [Washington 2018](https://geo.dot.gov/server/rest/services/Hosted/Washington_2018_PR/FeatureServer). Search also indexed [HPMS_FULL_WA_2023](https://geo.dot.gov/server/rest/services/Hosted/HPMS_FULL_WA_2023/FeatureServer/layers). Both direct JSON checks returned `404 Service not found` inside HTTP-success responses.
- **Schema/use:** documented fields include `YEAR_RECORD`, `ROUTE_ID`, `BEGIN_POINT`, `END_POINT`, `AADT`, truck AADT, `F_SYSTEM`, `SPEED_LIMIT`, `THROUGH_LANES`, access control and structure type. Full-extent coverage is the defined federal-aid/NHS system, not every local road. Useful exposure and road design inputs.
- **Rights/limits:** [federal catalog](https://catalog.data.gov/dataset/highway-performance-monitoring-system-hpms) is the existing government-works reuse basis; retain state contributor and edition metadata. No fee documented. Dataset cautions include imperfect spatial joins and unevaluated topology; not navigation-grade.
- **Readiness:** **access_failed for checked WA endpoints**. Resolve a current authoritative release/download first. Annual AADT is a two-way annual estimate, not hourly traffic, patrol effort or causal exposure. Use the release actually available at prediction time.

### P04 — Caltrans Annual Average Daily Traffic

- **Access:** [catalog API](https://data.ca.gov/api/3/action/package_show?id=annual-average-daily-traffic), [live layer](https://caltrans-gis.dot.ca.gov/arcgis/rest/services/CHhighway/Traffic_AADT/FeatureServer/0), [CSV download](https://gis.data.ca.gov/api/download/v1/items/d8833219913c44358f2a9a71bda57f76/csv?layers=0). [Traffic Census](https://dot.ca.gov/programs/traffic-operations/census) supplies year-specific reports through 2024 in the inspected page.
- **Coverage/schema:** California highway count **points**, not traffic-segment lines. Verified fields: `DISTRICT`, `RTE`, `RTE_SFX`, `CNTY`, `PM_PFX`, `PM`, `PM_SFX`, `DESCRIPTION`, `BACK_AADT`, `AHEAD_AADT`, peak-hour and peak-month values. Live schema has no year column. Catalog metadata modification in 2026 does not establish a 2026 traffic year.
- **Rights/access:** saved catalog says `license_id=cc-by`, Creative Commons Attribution, and `rights=No restrictions on public use`; exact CC version is not supplied. Layer credits State of California. Anonymous queries, 2,000-row cap; page and verify counts. Bulk cost/size not stated; no bulk download made.
- **Readiness:** strong acquisition candidate after binding the layer to a specific report year and license version. Preserve ahead/back semantics; never assign a count to both sides indiscriminately. Publisher's count year is October–September with sampling/seasonal adjustment.

### P05 — WSDOT annual traffic sections and counts

- **Access:** [TrafficData service](https://data.wsdot.wa.gov/arcgis/rest/services/Shared/TrafficData/FeatureServer); layer 0 points, [layer 1 lines](https://data.wsdot.wa.gov/arcgis/rest/services/Shared/TrafficData/FeatureServer/1). Saved layer description explicitly says **2023** estimated AADT for state highways, both directions combined.
- **Schema/use:** `RouteIdentifier`, `StateRouteNumber`, `AADT`, `ReportingYear`, `BeginAccumulatedRouteMile`, `EndAccumulatedRouteMile`, `LRSDate`. Linear referencing is valuable for event and treatment joins; lines are easier than point counts for exposure assignment.
- **Rights/limits:** service copyright is WSDOT; no express redistribution license found here, so **license_unresolved**. Public JSON/GeoJSON/PBF, pagination, 2,000-row maximum. No account or price advertised. Metadata-only acquisition.
- **Next:** establish reuse permission and historical editions, validate actual `ReportingYear` distribution, and document spatial matching uncertainty. Avoid representing 2023 AADT as current measured traffic or all-road coverage.

### P06 — Utah wildlife crossings and fence survey

- **Authority/access:** [official Utah Wildlife Migration Initiative crossing page](https://wildlifemigration.utah.gov/land-animals/crossing/) embeds [this ArcGIS experience](https://experience.arcgis.com/experience/5d80fdda0d7d47a5b94071377d43a9aa). Its [app configuration](https://www.arcgis.com/sharing/rest/content/items/5d80fdda0d7d47a5b94071377d43a9aa/data?f=pjson) references [crossing points](https://services.arcgis.com/ZzrwjTRez6FJiOq4/arcgis/rest/services/Wildlife_Crossing_public/FeatureServer/0) and [fence lines](https://services.arcgis.com/ZzrwjTRez6FJiOq4/arcgis/rest/services/Fence_Survey_PublicView/FeatureServer/0).
- **Result:** crossing service returned `499 Token Required`; its item returned 403. Fence endpoint returned `400 Invalid URL`; item was absent/inaccessible. A public app configuration and a name containing `public` do not establish data access. No features, attachments or photos retrieved; underlying field schema/year completeness remains unverified.
- **Rights/readiness:** **access_restricted / license_unresolved**. No bypass, request, signup or agreement attempted. Ask publisher later for a releasable inventory with use terms, structure type, commissioned date, active status, fence geometry/endpoints and survey extent.
- **Contribution/caution:** mitigation proximity, distance to fence ends and time since installation; absent feature cannot mean no fence. Treatment placement is nonrandom and often follows historical collisions. Do not infer treatment dates from record creation or make causal savings claims from proximity alone.

## Park/place context and habitat

### P07 — NPS administrative unit boundaries and API context

- **Exact release:** [federal catalog](https://catalog.data.gov/dataset/administrative-boundaries-of-national-park-system-units-national-geospatial-data-asset-ngd) identifies **2026-07-22**, IRMA reference **2319366**, [GDB ZIP](https://irma.nps.gov/DataStore/DownloadFile/761368?Reference=2319366) and [metadata XML](https://irma.nps.gov/DataStore/DownloadFile/761365?Reference=2319366). XML acquired successfully; an older 2224545 landing-page attempt failed 403. National unit polygons provide park filters and context, not mortality coverage or land ownership proof.
- **Rights/schema:** catalog explicitly labels public domain. XML says `GIS_NOTES` beginning `Lands` means completed QA; `Preliminary` has not completed all QA. Separate park/preserve designations can be separate polygons; quarterly updates. Not engineering/legal boundary precision. GDB file size not checked.
- **Operational supplement:** [NPS API](https://www.nps.gov/subjects/developer/api-documentation.htm) supplies parks/alerts, needs an API key and defaults to 1,000 requests/hour. No account/key created. Confirm individual photo/content rights under NPS disclaimer; boundary rights do not license media. Alerts are not a road routing graph.
- **Next:** acquire/pin GDB, inspect fields/CRS and QA flags, preserve unit IDs and release dates. Historic park boundaries need historic editions.

### P08 — GNIS names and Census place boundaries

- **Access:** [GNIS download instructions](https://www.usgs.gov/us-board-on-geographic-names/download-gnis-data), [staged names products](https://prd-tnm.s3.amazonaws.com/index.html?prefix=StagedProducts/GeographicNames/), and [TIGER/Line](https://www.census.gov/cgi-bin/geo/shapefiles/). GNIS national/state pipe-delimited TXT, GDB/GPKG; names downloads refresh every other month, with separate static legacy archives. Search app CSV export cap is 2,000 records.
- **Schema/use:** feature ID, official/variant names, class, state/county and geographic coordinates support offline place search and distance-to-town features. TIGER place polygons provide FIPS/GEOID boundaries for joins; demographic attributes require separate Census tables.
- **Rights/readiness:** USGS public-domain policy is source-level basis, subject to item notices; Census technical-documentation legal disclaimer should be recorded with the selected vintage. No bundle downloaded. Free bulk access; national size unmeasured.
- **Caution:** GNIS removed many administrative classes in 2021, including parks, bridges and trails ([official explanation](https://www.usgs.gov/us-board-on-geographic-names/domestic-names)); do not use it as a complete park inventory. Settlement proximity is a reporting-effort proxy, not measured effort. Pin historic geography for evaluations.

### P09 — PAD-US 4.1

- **Access:** [official downloads](https://www.usgs.gov/programs/gap-analysis-project/science/pad-us-data-download), [release DOI](https://doi.org/10.5066/P96WBCHS), [state GDB/KMZ collection](https://www.sciencebase.gov/catalog/item/6759abcfd34edfeb8710a004). National and 56 state/territory selections; current 4.1 inventory and distinct overlap-resolved analysis products.
- **Schema/use:** ownership, manager, designation, GAP protection, access and establishment attributes support jurisdiction/context and protected-land overlap. Choose analysis layer for area percentages and full inventory for overlapping designations.
- **Rights/access:** USGS public-domain basis from existing ledger, retain contributor metadata and use constraints; no selected artifact cleared here. Free catalog route, but ScienceBase state page returned 403 to browser fetch; size not verified.
- **Readiness:** catalog verified, acquisition pending. Publisher warns against comparing editions as true acquisition/change histories; many differences are GIS improvements. `Date_Est` is incomplete. Protection is not habitat quality or an animal-presence label.

### P10 — Annual NLCD Collection 1.2

- **Version/access:** [current MRLC catalog](https://www.mrlc.gov/data) and [USGS product overview](https://www.usgs.gov/centers/eros/science/about-annual-nlcd) verify **1985–2025 CONUS**, released June 2026. Older access FAQ still says through 2024. [ScienceBase collection](https://www.sciencebase.gov/catalog/item/655ceb8ad34ee4b6e05cc51a), MRLC direct downloads and [OGC services](https://www.mrlc.gov/data-services-page) are discovery routes; do not invent a file URL from an old collection layout.
- **Schema/use:** 30 m annual land cover, change, confidence, fractional impervious surface, impervious descriptor and spectral-change day. Compute 250 m/1 km class shares, forest edge, wetland and development context. Class codes are categorical; canopy percentage is a different product.
- **Rights/limits:** USGS public-domain source policy; preserve release metadata. MRLC direct access free; USGS AWS route is requester-pays. Viewer subsets may require email delivery (not used). CONUS mosaics are large; acquire tiles/bbox and selected years, not all products.
- **Readiness:** prioritize historical annual layers. Final annual imagery/reprocessing can include information unavailable at a forecast date; lag features and document retrospective vs operational evaluation. Alaska/Hawaii are not included in this CONUS product.

### P11 — USDA Forest Service Tree Canopy Cover v2025.6

- **Access/version:** [publisher page](https://data.fs.usda.gov/geodata/rastergateway/treecanopycover/) and [MRLC science product](https://www.mrlc.gov/data/type/science-tree-canopy-cover) verify 30 m **CONUS 1985–2025**, released 2026. Direct example linked by publisher: [1985 Science TCC ZIP](https://data.fs.usda.gov/geodata/rastergateway/treecanopycover/docs/v2025-6/science_tcc_conus_1985_v2025-6_wgs84.zip). Select target years from live dropdown URLs; no ZIP fetched.
- **Schema:** Science TCC includes modeled percentage and standard error/uncertainty; NLCD TCC is separately masked, filtered and temporally processed. Science output can assign canopy to water/non-tree crops; no-data/background sentinels must be decoded from selected metadata.
- **Rights/access:** publisher supplies enterprise-data disclaimer, but an exact new-release license/use-constraints document was not inspected. **Selected-artifact rights pending**, not a CC license claim. Public free downloads; large national rasters, size unmeasured.
- **Contribution/next:** canopy density and edge structure, kept separate from NLCD forest class. Compare science vs processed layer for model input. Temporal smoothing across years can leak future observations into a historical forecasting test; pin processing version and lag strategy.

### P12 — ESA WorldCover 2020 v100 / 2021 v200

- **Access/rights:** [publisher download and license page](https://esa-worldcover.org/en/data-access), [2021 DOI](https://doi.org/10.5281/zenodo.7254221), [2020 DOI](https://doi.org/10.5281/zenodo.5571936). Explicit **CC BY 4.0**, free use with WorldCover/Copernicus acknowledgement, including maps and model/data products.
- **Coverage/schema:** global 10 m land-cover maps, fixed 2020/2021 vintages; gridded tiles and GeoTIFF/COG access described by product manuals. Higher-resolution global context/extension beyond CONUS; 11 broad land-cover categories, not species habitat probability.
- **Limits/next:** download a selected tile after size inspection, preserve version and class legend. Do not use difference between these two algorithm versions as a clean land-cover-change measure, or represent 2021 land cover as current 2026 conditions. Secondary to NLCD's historical time series for western-US evaluation.

## Terrain and hydrology

### P13 — USGS 3DEP DEM

- **Access:** [TNM API documentation](https://www.usgs.gov/faqs/there-api-accessing-national-map-data). A real [bounded query](https://tnmaccess.nationalmap.gov/api/v1/products?datasets=National%20Elevation%20Dataset%20(NED)%201%2F3%20arc-second&bbox=-111.2,45.5,-111.0,45.7&max=1&outputFormat=JSON) was saved. It returned historical tile `USGS 1/3 Arc Second n46w111 20121001`, **416,332,989 bytes**, [exact TIFF](https://prd-tnm.s3.amazonaws.com/StagedProducts/Elevation/13/TIFF/historical/n46w111/USGS_13_n46w111_20121001.tif). TIFF not downloaded.
- **Schema/rights:** about 10 m, 1-degree GeoTIFF; returned metadata says NAD83 geographic coordinates, elevation meters/NAVD88 in CONUS, all 3DEP products public domain. [Publisher](https://www.usgs.gov/3d-elevation-program) confirms no fee/use restrictions. Other resolutions/areas differ.
- **Contribution:** slope, aspect, ruggedness, valley/drainage position and road grade. Reproject appropriately before distances/slope; maintain datum and acquisition dates.
- **Readiness:** strong candidate. The first API result is **historical**, despite 2026 catalog modification; select intentionally by product/date and actual road-buffer coverage. Large files motivate regional processing and immutable caching.

### P14 — NHDPlus HR / NHD / WBD and successor 3DHP

- **Access:** [official transition/download page](https://www.usgs.gov/3d-hydrography-program/access-3dhp-data-products), [NHDPlus HR staged regional products](https://prd-tnm.s3.amazonaws.com/index.html?prefix=StagedProducts/Hydrography/NHDPlusHR/), [live 3DHP FeatureServer](https://3dhp.nationalmap.gov/arcgis/rest/services/usgs_3dhp_all/FeatureServer). Saved service metadata says **refreshed 2026-09-04**, newer than landing-page July date.
- **Schema:** legacy regional GDB flowlines, waterbodies, catchments and HUCs; 3DHP verified layer IDs 50 flowline, 60 waterbody, 80 catchment, 90 EDH workunit, 92 elevation workunit, 94 DEM limitation area. Live query cap 2,500; JSON/feature-service access.
- **Rights/readiness:** USGS public-domain source basis with per-product contributor review. No raster/vector features downloaded. Legacy products no longer maintained following transition in 2024; new service combines coverage needing workunit/lineage inspection. Pin annual DOI release when reproducibility matters.
- **Contribution/caution:** distance to water, drainage crossings, riparian context, catchment grouping. Hydrographic lines do not prove year-round water or wildlife passage. NHD/NHDPlus/WBD are related, not independent corroboration. Avoid duplicate overlapping editions and any assumption national 3DHP coverage is uniformly newly remapped.

## Weather, snow, fire and time

### P15 — NOAA GHCN-Daily v3

- **Access:** [collection and constraints](https://www.ncei.noaa.gov/access/metadata/landing-page/bin/iso?id=gov.noaa.ncdc:C00861), [schema README](https://www.ncei.noaa.gov/pub/data/ghcn/daily/readme.txt), [station files](https://www.ncei.noaa.gov/pub/data/ghcn/daily/all/). Global station observations with uneven historical periods. Existing ledger already has the station inventory and Bozeman `USW00024132.dly`; not redownloaded here.
- **Schema/use:** fixed-width station/month/element with daily slots; TMIN/TMAX, PRCP, SNOW and SNWD as available, measurement/quality/source flags and element-specific scales. Useful local weather validation and snow-depth observations; preserve flags and missing sentinels.
- **Rights/cost:** NOAA metadata specifies citation and liability constraints; source-level reuse is documented, not an invented SPDX license. Direct HTTPS free/no key; selected station files modest, all-station archive large.
- **Next/caution:** inventory target-area station/element/year completeness. Stations can move; nearest station may be at a very different elevation. Observation-day convention and late QC revisions must be respected. No station report is not zero precipitation/snow.

### P16 — Daymet Version 4 R1 (collection 2129)

- **Current metadata:** saved [CMR search](https://cmr.earthdata.nasa.gov/search/collections.umm_json?keyword=Daymet%20Daily%20V4&page_size=10) identifies `Daymet_Daily_V4R1_2129`, version 4.1 and [DOI](https://doi.org/10.3334/ORNLDAAC/2129), replacing the older ledger's 1840 candidate. [Current guide PDF](https://data.ornldaac.earthdata.nasa.gov/public/daymet/Daymet_Daily_V4R1/comp/Daymet_Daily_V4R1.pdf) acquired as documentation. [Web services](https://daymet.ornl.gov/web_services) provide subset discovery; Earthdata/OPeNDAP bulk paths may need free Earthdata login.
- **Coverage/schema:** North America and Hawaii **1980–2025**; Puerto Rico begins 1950. CMR's collection start 1950 is therefore not western-US history. 1 km CF NetCDF, daily Tmin/Tmax, precipitation, shortwave radiation, vapor pressure, SWE and daylength. Leap years omit December 31 in the 365-day convention. R1 corrected 2020/2021 input problems; outside those years files are unchanged from v4.
- **Rights/readiness:** CMR explicitly points to NASA data-information policy, but its exact legacy URL returned 404 in this pass; keep **policy-link resolution pending** rather than asserting blanket public domain. No climate data downloaded. Size depends on subset; annual continent grids can be large.
- **Use/caution:** detailed historical weather; SWE is not snow depth. Full-year retrospective processing is not a forecast input. Preserve domain, collection and revision identity.

### P17 — PRISM climate grids

- **Access:** [data](https://prism.oregonstate.edu/data/), [bulk instructions](https://prism.oregonstate.edu/downloads/), [data directory](https://data.prism.oregonstate.edu/). CONUS daily/monthly/time-series and normals; choose exact resolution, product, dates and stability status from current catalog. Historic products and normals are not interchangeable; selected raster schema/vintage not sampled here.
- **Rights:** saved [current terms](https://prism.oregonstate.edu/terms/) expressly allow data reproduction and distribution; require prominent PRISM Group/OSU name, URL and access date. Map graphics require copyright/URL/map-creation date. Ownership remains PRISM/OSU; **custom terms**, not CC0. This is stronger evidence than the earlier unresolved ledger entry.
- **Contribution/limits:** terrain-informed precipitation/temperature and climate normals; public bulk routes, no purchase made and size/rate cap unverified. Do not assume every premium/fine-resolution product follows the same download conditions.
- **Next/caution:** pin one free product and inspect headers/units/metadata. Provisional and stable revisions differ. Publisher cautions against very long-term trend use; final retrospective fields are unavailable at the original forecast date.

### P18 — gridMET

- **Access/rights:** [publisher documentation](https://www.climatologylab.org/gridmet.html), [annual NetCDF directory](https://www.northwestknowledge.net/metdata/data/). Explicit author copyright/related-rights waiver and statement free of known restrictions; cite Abatzoglou and retain metadata. Free public download; directory has `early`, `provisional`, `permanent` branches and annual variable files, often tens/hundreds of MB. No grid downloaded.
- **Coverage/schema:** CONUS 1979 onward, daily ~4 km (1/24 degree); some real-time southern British Columbia coverage. Temperature, precipitation, humidity, radiation, wind, evapotranspiration, drought and fire-weather indices. NETCDF4 scales/offsets require decoding.
- **Use/limits:** broad weather/drought predictors; not road-level microclimates. Radiation is planar, uncorrected for terrain. Nominal day is midnight Mountain Standard Time, not local DST. Most recent 60 days preliminary and subject to replacement; upstream NLDAS revisions affect older periods too.
- **Readiness:** high-priority historical grid with explicit rights. Save version/hash and query window; choose a stable vintage for model evaluation and an as-of snapshot policy for operational features.

### P19 — SNODAS G02158 v1

- **Access:** [NSIDC dataset](https://nsidc.org/data/g02158/versions/1), [anonymous archive](https://noaadata.apps.nsidc.org/NOAA/G02158/) with masked/unmasked products; index and [guide](https://nsidc.org/sites/default/files/g02158-v001-userguide_2_1.pdf) saved. US domain daily ~1 km from late September/October 2003 onward, bounds differ by masked/unmasked product. Binary grids with headers; no daily archive fetched.
- **Schema:** 16-bit big-endian signed grids; snow depth, SWE, precipitation and snowpack/flux variables. Units/scales and valid period are variable-specific. This is model/assimilation output, not direct observation everywhere.
- **Rights/access:** provider requires citation and subset/access-date acknowledgement; [citation policy](https://nsidc.org/about/data-use-and-copyright). No SPDX grant established from inspected pages; preserve exact use constraints before redistribution. Anonymous HTTPS is documented, no fee, large daily CONUS grids.
- **Critical QA:** guide documents erroneous zero SWE/depth in affected cells from **2014-10-09 to 2019-10-10**; apply `SNODAS_Zero_Repair_Mask.tif` as no-data. Fetch mask before historical features. Missing archive dates also have a published list. Best snow-specific candidate after these checks; do not substitute zeros.

### P20 — Monitoring Trends in Burn Severity (MTBS)

- **Access:** [program](https://www.mtbs.gov/), [current Burn Severity direct download](https://burnseverity.cr.usgs.gov/direct-download), [product FAQ](https://www.mtbs.gov/faqs). Historical fire perimeter vectors and 30 m severity GeoTIFFs, US including Alaska/Hawaii/Puerto Rico, 1984 onward. Covers qualifying fires ≥1,000 acres in western US and ≥500 acres in eastern US, including prescribed fire.
- **Use:** years since fire, fraction of habitat burned, severity and recovery context. Burn polygons are not road closures; active fire extent must come from a separate operational source.
- **Rights/access:** federal USGS/USFS program, but selected release license/use constraints still need inspection; no assumed blanket clearance. Free portal downloads by fire/state/year; no bulk file or size inspected. The obsolete guessed `/pages/direct-download` failed; use current verified link above.
- **Caution/next:** small fires are omitted by design. Extended assessments may use imagery from the following growing season, causing temporal leakage. Use acquisition/publication time plus fire date. Choose a pinned vector release first, then limited severity rasters only if useful.

### P21 — Sun, twilight and moon via USNO; computed seasonal calendar

- **Access:** [official API](https://aa.usno.navy.mil/data/api), [one-day reference endpoint](https://aa.usno.navy.mil/api/rstt/oneday?date=2025-09-21&coords=45.68,-111.04&tz=-7&dst=true) saved for Bozeman. Global positions, dates 1700–2100; one-day GeoJSON supplies sunrise/set, civil twilight, moonrise/set, phase and noon illumination fraction.
- **Contribution:** deterministic dark/light intervals overlapping commute periods, seasonal sine/cosine terms, moon phase; use a vetted Go implementation for scalable computation with USNO as reference validation. Observed weather/clouds and terrain affect actual illumination.
- **Rights/limits:** US federal astronomical service; no dataset-specific reuse/rate-limit statement verified here, so API output redistribution remains a metadata check, not an asserted CC license. No key/payment requested; do not fan out millions of API calls.
- **Next/caution:** verify time-zone/DST and null polar-day/night values. Noon moon fraction is not illumination at crash time. Species-specific rut/hunt seasons need state, species, zone and year sources; month alone is a proxy, not verified biology or current regulation.

### P22 — NWS forecasts and alerts for operational route context

- **Access/rights:** [official API documentation](https://www.weather.gov/documentation/services-web-api) expressly says free/open for any purpose, no fee, unpublished reasonable rate limits. User-Agent required. [point discovery example](https://api.weather.gov/points/45.68,-111.04) resolves current forecast office/grid; do not hardcode office mapping.
- **Coverage/schema:** US WFO forecast grid ~2.5 km; GeoJSON/JSON-LD forecast, hourly and raw grid data for the next seven days; active alerts and recent alert history. Preserve `generatedAt`/issuance, valid times and units. Endpoint not sampled in this pass; documentation verified.
- **Contribution:** snow/precipitation/visibility-related trip context and forecast covariates, separate from baseline seasonal collision risk. Alerts feed only retains recent history and cannot supply years of training.
- **Readiness:** source-level ready for a bounded future adapter. Snapshot forecasts before outcomes and evaluate issued forecasts, not later observed weather. Cache respecting expiry; handle rate limits, stale forecasts, missing cells and changes in grid mapping. An alert-free response is not a safe-driving guarantee.

## Implementation acceptance gates

Keep source observation date, valid date range, provider publication time, retrieval time and feature-build time separate. Pin edition, checksum, CRS, datum, units and no-data semantics. Detect semantic API error objects even with HTTP 200. For ArcGIS, inspect capability/page limits, retrieve IDs/counts, page explicitly and reject truncation. None of this sweep's metadata requests establishes actual feature counts.

Model features should use data available when a prediction would have been issued. For a first retrospective experiment, state that limitation explicitly: matching year alone does not eliminate hindsight from land-cover smoothing, final AADT, weather reanalysis or later mitigation inventories. Historical geometry and treatment changes also need validity intervals. Exposures require segment length, time coverage and traffic-vintage uncertainty; reporting effort is a separate problem.

## Acquisition appendix

The following table is generated from the saved raw files. Retrieval timestamp means completion of this research download (filesystem write time in UTC), not data vintage. Schemas/coverage/rights are described in P01–P22 above. Error/empty responses are retained to substantiate limitations. These evidence files are excluded from Git; their hashes must not be presented as hashes of the underlying national datasets.

| Evidence file | Retrieval completion UTC | Bytes | SHA-256 | Exact URL | Result |
| --- | --- | ---: | --- | --- | --- |
| `wsdot-traffic.json` | 2026-09-21T21:03:55Z | 15440 | `717db9dc7f3a43031120b6450f72fe1c1f0c0e18f13d9c2d382d9916c2908740` | [source](https://data.wsdot.wa.gov/arcgis/rest/services/Shared/TrafficData/FeatureServer/1?f=pjson) | metadata_or_documentation_acquired |
| `hpms-wa-2023.json` | 2026-09-21T21:04:14Z | 83 | `c17a60336f4f80d284ce3065665b79c54480b2a19506e0b71d0ee6fe8edd3092` | [source](https://geo.dot.gov/server/rest/services/Hosted/HPMS_FULL_WA_2023/FeatureServer/layers?f=pjson) | API error: {"code":404,"message":"Service not found","details":[]} |
| `prism-terms.html` | 2026-09-21T21:03:54Z | 16331 | `6391f53b5a8286eb4298bf78657af7b0e303c6b833acf3fcae88b460a2690e63` | [source](https://prism.oregonstate.edu/terms/) | metadata_or_documentation_acquired |
| `gridmet.html` | 2026-09-21T21:03:55Z | 65161 | `6b1366c1fd279628a150d7c3d1a5e7456550ed56dd998423d6a6023b3a03e9d2` | [source](https://www.climatologylab.org/gridmet.html) | metadata_or_documentation_acquired |
| `hydro-products.html` | not acquired | 0 | null | [source](https://www.usgs.gov/3d-hydrography-program/access-3dhp-data-products) | HTTP fetch failed; no artifact saved |
| `caltrans-aadt.json` | 2026-09-21T21:03:54Z | 10061 | `ede68ea50744fcc9c21390d59530bfb6da1b4721193e2b617f9af27dbfde4350` | [source](https://data.ca.gov/api/3/action/package_show?id=annual-average-daily-traffic) | metadata_or_documentation_acquired |
| `geofabrik-utah.html` | 2026-09-21T21:03:54Z | 22354 | `0bdd2e13f55cf2e78bc52899c0f73e8518f4cac3be54311c54d22a1992fa364e` | [source](https://download.geofabrik.de/north-america/us/utah.html) | metadata_or_documentation_acquired |
| `protomaps-license.md` | 2026-09-21T21:03:54Z | 3369 | `9d9064bc0004a507ce105b92552538dd1484cd704afe42da02a9a74c3bd4a293` | [source](https://raw.githubusercontent.com/protomaps/basemaps/main/LICENSE_DATA.md) | metadata_or_documentation_acquired |
| `daymet-metadata.json` | 2026-09-21T21:03:54Z | 30 | `13b1d60fadb5a82ecc0761894c1969aa5850602e06242dab61c698281ad9f601` | [source](https://cmr.earthdata.nasa.gov/search/collections.umm_json?short_name=Daymet_Daily_V4&provider=ORNL_CLOUD) | empty_search_result |
| `nps-boundaries.html` | not acquired | 0 | null | [source](https://irma.nps.gov/DataStore/Reference/Profile/2224545) | HTTP fetch failed; no artifact saved |
| `snodas-index.html` | 2026-09-21T21:03:54Z | 629 | `73fe12de69dc0b2e3cbe98eaea508073fe7b28dcfbabe28ac626f285beb03854` | [source](https://noaadata.apps.nsidc.org/NOAA/G02158/) | metadata_or_documentation_acquired |
| `caltrans-aadt-layer.json` | 2026-09-21T21:04:41Z | 9630 | `a11f89a0be6cab2340d09964f830b01da3c18e852190e00b903e331cdc901e64` | [source](https://caltrans-gis.dot.ca.gov/arcgis/rest/services/CHhighway/Traffic_AADT/FeatureServer/0?f=pjson) | metadata_or_documentation_acquired |
| `utah-crossings.html` | 2026-09-21T21:04:42Z | 60620 | `ecc967625ba97e030973ff10dc9dda7e9907c002f246928d6005811266b64116` | [source](https://wildlifemigration.utah.gov/land-animals/crossing/) | metadata_or_documentation_acquired |
| `daymet-search.json` | 2026-09-21T21:04:41Z | 60392 | `fdfda90a93fdce1737ace6ec83f4ed56dba8b7d3a65f4c359d3fb0ddb9f26eb8` | [source](https://cmr.earthdata.nasa.gov/search/collections.umm_json?keyword=Daymet%20Daily%20V4&page_size=10) | metadata_or_documentation_acquired |
| `tnm-dem-sample.json` | 2026-09-21T21:04:41Z | 4547 | `0dfafa338d1f0621880c76a26bf35d8af5f65fe218bb4998b411a2d733336074` | [source](https://tnmaccess.nationalmap.gov/api/v1/products?datasets=National%20Elevation%20Dataset%20(NED)%201%2F3%20arc-second&bbox=-111.2,45.5,-111.0,45.7&max=1&outputFormat=JSON) | metadata_or_documentation_acquired |
| `tree-canopy.html` | 2026-09-21T21:04:41Z | 197066 | `02b5be283233e3f9ae7fc3566ddf76af33a46b7d791acc4f92add03051007f26` | [source](https://data.fs.usda.gov/geodata/rastergateway/treecanopycover/) | metadata_or_documentation_acquired |
| `usno-day.json` | 2026-09-21T21:04:41Z | 1290 | `a2964b48698edb2f34dd7a2e8d51bfebe06b90c4c18ae2b347d94b078a8b9af3` | [source](https://aa.usno.navy.mil/api/rstt/oneday?date=2025-09-21&coords=45.68,-111.04&tz=-7&dst=true) | metadata_or_documentation_acquired |
| `utah-crossings-app.json` | 2026-09-21T21:05:25Z | 67583 | `dea68f9a62e54c3ab96b90d09221a66f006f9be546a59e2d317468e17318c0b9` | [source](https://www.arcgis.com/sharing/rest/content/items/5d80fdda0d7d47a5b94071377d43a9aa/data?f=pjson) | metadata_or_documentation_acquired |
| `3dhp-service.json` | 2026-09-21T21:05:25Z | 13722 | `dd4d434cc313285e259f7eb33b682b9992ab3fd768cb8831895c425ee5c31eed` | [source](https://3dhp.nationalmap.gov/arcgis/rest/services/usgs_3dhp_all/FeatureServer?f=pjson) | metadata_or_documentation_acquired |
| `nps-boundary-metadata.xml` | 2026-09-21T21:05:25Z | 1031407 | `0653e1bf56d3785bca24524862ad56cad5b1f3b2bf4916af02e4827199b72db5` | [source](https://irma.nps.gov/DataStore/DownloadFile/761365?Reference=2319366) | metadata_or_documentation_acquired |
| `daymet-r1-guide.pdf` | 2026-09-21T21:05:26Z | 2947884 | `57672bf26ebc6382823749154fbea3540c25c85e852e4fae812e5d4eae0c87b4` | [source](https://data.ornldaac.earthdata.nasa.gov/public/daymet/Daymet_Daily_V4R1/comp/Daymet_Daily_V4R1.pdf) | metadata_or_documentation_acquired |
| `snodas-guide.pdf` | 2026-09-21T21:05:25Z | 756907 | `0ed2430b555bf463b1aba20f7a9c47ba927743a1931f161a3e32e4319825cf5e` | [source](https://nsidc.org/sites/default/files/g02158-v001-userguide_2_1.pdf) | metadata_or_documentation_acquired |
| `mtbs-download.html` | not acquired | 0 | null | [source](https://burnseverity.cr.usgs.gov/pages/direct-download) | HTTP fetch failed; no artifact saved |
| `hpms-wa-2018.json` | 2026-09-21T21:05:55Z | 83 | `c17a60336f4f80d284ce3065665b79c54480b2a19506e0b71d0ee6fe8edd3092` | [source](https://geo.dot.gov/server/rest/services/Hosted/Washington_2018_PR/FeatureServer?f=pjson) | API error: {"code":404,"message":"Service not found","details":[]} |
| `utah-crossing-layer.json` | 2026-09-21T21:06:12Z | 170 | `06e467aec7ba27d5107a5e00d79c4b11221d46836625a2b30a62c8dea0e63623` | [source](https://services.arcgis.com/ZzrwjTRez6FJiOq4/arcgis/rest/services/Wildlife_Crossing_public/FeatureServer/0?f=pjson) | API error: {"code":499,"message":"Token Required","messageCode":"GWM_0003","details":["Token Required"]} |
| `utah-fence-layer.json` | 2026-09-21T21:06:12Z | 130 | `0b679644c877cfdcd4b90912e34575cd2ad3cba69eeab90392eaeb50b3ae3382` | [source](https://services.arcgis.com/ZzrwjTRez6FJiOq4/arcgis/rest/services/Fence_Survey_PublicView/FeatureServer/0?f=pjson) | API error: {"code":400,"message":"Invalid URL","details":["Invalid URL"]} |
| `utah-crossing-item.json` | 2026-09-21T21:06:11Z | 165 | `875a610ad073508137968cff34f828e6a5a3cb11390d92d0cab65e22cf558d3f` | [source](https://www.arcgis.com/sharing/rest/content/items/440acd2ef0ee492286377f8b50800914?f=pjson) | API error: {"code":403,"messageCode":"GWM_0003","message":"You do not have permissions to access this resource or perform this operation.","details":[]} |
| `utah-fence-item.json` | 2026-09-21T21:06:11Z | 127 | `9ef7b281588428373c764295a351a81daff8cef8301ebec1961231b205d1d4f3` | [source](https://www.arcgis.com/sharing/rest/content/items/f2012876a77e4b7fb76e9f6a973b55fa?f=pjson) | API error: {"code":400,"messageCode":"CONT_0001","message":"Item does not exist or is inaccessible.","details":[]} |

Machine-readable manifest: `data/evidence/sweep-2026-09-21/predictors/manifest.json`.
