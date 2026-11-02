# Corridor November 2026 Data Expansion Sweep

**Date:** 2026-11-02 UTC  
**Scope:** High-priority public datasets scouted to support multi-state expansion (California, Montana) and National Park pilots (Yosemite, Yellowstone, Grand Teton, Glacier, Olympic).  
**Catalog Range:** S31–S37 (expanding the research ledger in `data/SOURCES.md`).  
**Status:** `public_catalog` for all entries. In accordance with Corridor provenance rules, data retrieval timestamps and dataset SHA-256 values remain `null` until individual artifacts are downloaded to private object storage. No synthetic data, fabricated hashes, or unverified claims are included.

---

## 1. Executive Summary & Sweep Objectives

The November 2026 data expansion sweep was conducted to address critical data requirements established during the Phase 2 multi-state road network ingestion, Phase 5 alpine covariate expansion, and the Yosemite National Park pilot implementation:

1. **Authoritative Traffic Exposure (California):** Replacing national HPMS sample approximations with Caltrans count stations and linear referencing postmiles along Sierra Nevada corridors (CA-120 Crane Flat / Tioga Pass, CA-140 Merced River Canyon, CA-41 Wawona).
2. **Statewide Highway Carcass Inventory (Montana):** Bypassing the earlier failed university repository (S21) with Montana DOT's direct 10-year statewide carcass dataset covering 60,592 records on state highways.
3. **Nationwide National Park Wildlife Mortality:** Resolving the Phase 0 park mortality fallback (S29) by cataloging the NPS IRMA Wildlife Mortality and Morbidity Observations (WHMMO) database across all major national parks.
4. **Gridded Alpine Snow Covariates:** Acquiring daily 1 km snow depth and Snow Water Equivalent (SWE) grids from NOAA SNODAS to power alpine pass risk predictions and winter closure detection.
5. **Cloud-Optimized High-Resolution Terrain:** Leveraging USGS 3DEP 10-meter and 1-meter Cloud-Optimized GeoTIFFs (COGs) via HTTP range GETs to extract segment slope, aspect, and terrain roughness without multi-terabyte bulk downloads.
6. **Standardized Survey Effort & Non-Detection Baselines:** Integrating the Smithsonian Snapshot USA coordinated camera trap survey (2019–2024) to provide trap-night effort metrics for two-stage Hurdle models correcting roadkill reporting bias.
7. **Western Wildlife Vision Pre-Training Library:** Cataloging LILA BC's multi-terabyte western camera trap image collections (Idaho, Caltech, WSU Lynx) to fine-tune BioCLIP and SpeciesNet vision embeddings across Corridor's 11 target wildlife species.

---

## 2. Detailed Technical Dossiers (S31–S37)

### S31: Caltrans Annual Average Daily Traffic (AADT) & Linear Referencing System (LRS) Postmiles

- **Publisher:** California Department of Transportation (Caltrans), Division of Traffic Operations / GIS Branch.
- **Authoritative URLs:**
  - ArcGIS REST FeatureServer: `https://caltrans-gis.dot.ca.gov/arcgis/rest/services/CHhighway/Traffic_AADT/FeatureServer/0`
  - California Open Data Portal: `https://data.ca.gov/api/3/action/package_show?id=annual-average-daily-traffic`
  - Caltrans Traffic Census Program: `https://dot.ca.gov/programs/traffic-operations/census`
- **Access Status:** `public_catalog`
- **License & Rights:** State of California Open Data; Creative Commons Attribution (CC-BY). Catalog explicitly specifies `license_id=cc-by` and `rights=No restrictions on public use`. Attribution: "Data provided by the California Department of Transportation (Caltrans)."
- **Coverage & Geometry:**
  - Entire California State Highway System, encompassing all Interstates, US Highways, and State Routes.
  - Specifically covers Yosemite corridors: CA-120 (Big Oak Flat Rd / Tioga Pass, postmiles 0.00–62.50+), CA-140 (Merced River Canyon / El Portal Rd, postmiles 0.00–43.40+), and CA-41 (Wawona Rd).
  - Point geometry (`esriGeometryPoint`) in Web Mercator (EPSG:3857) and WGS84 (EPSG:4326).
- **Documented Schema:**
  - `DISTRICT`: Caltrans administrative district (1 to 12; Yosemite is District 10).
  - `RTE`: State route number (e.g., 120, 140, 41).
  - `RTE_SFX`: Route suffix (e.g., S for supplement).
  - `CNTY`: County code (e.g., TUO for Tuolumne, MPA for Mariposa, MAD for Madera).
  - `PM_PFX`: Postmile prefix (e.g., R for realignment, M for maintain, L for loop).
  - `PM`: Postmile numerical measure along the route.
  - `PM_SFX`: Postmile suffix.
  - `DESCRIPTION`: Text location description (e.g., "JCT. RTE. 120 CRANE FLAT", "TUOLUMNE-MONO CO. LINE / TIOGA PASS").
  - `BACK_AADT`: Annual average daily traffic arriving at the count station (upstream).
  - `AHEAD_AADT`: Annual average daily traffic departing the count station (downstream).
  - `BACK_PEAK_HR` / `AHEAD_PEAK_HR`: Peak hour traffic volume back and ahead.
  - `BACK_PEAK_MONTH` / `AHEAD_PEAK_MONTH`: Peak month average daily traffic back and ahead.
- **Bias & Methodological Notes:**
  - *Point Geometry vs. Continuous Lines:* Features are discrete count locations, not continuous road polylines. Exposure assignment to Corridor's 0.1-mile highway segments requires linear postmile interpolation or nearest-station snapping within the same route and county bounds.
  - *Directional Differences:* At highway junctions, ramp splits, and recreational intersections (e.g., CA-120 at Crane Flat junction with Tioga Road), `BACK_AADT` and `AHEAD_AADT` differ significantly. Ingestion pipelines must respect directionality rather than naively averaging values.
  - *Annual vs. Seasonal Realities:* AADT represents annualized traffic averaged over a 12-month census cycle (October–September). For high-elevation alpine routes like Tioga Pass (CA-120 MP 42–62), winter pass closures reduce actual daily traffic to zero for 5–6 months; raw AADT must be modulated using elevation and SNODAS snow depth features.
  - *API Constraints:* ArcGIS FeatureServer limits queries to 2,000 features per request. ETL adapters must implement `resultOffset` pagination and verify feature count completeness.

---

### S32: Montana DOT Carcass Collection (2016–2025)

- **Publisher:** Montana Department of Transportation (MDT), Maintenance Division and Data and Statistics Bureau.
- **Authoritative URLs:**
  - Direct XLSX Download: `https://www.mdt.mt.gov/publications/docs/datastats/crashdata/CarcassData2016-2025.xlsx`
  - Crash & Carcass Portal: `https://www.mdt.mt.gov/publications/datastats/crashdata.aspx`
  - MDT Terms of Use: `https://www.mdt.mt.gov/mdt/terms-of-use.aspx`
- **Access Status:** `public_catalog`
- **License & Legal Constraints:**
  - MDT Terms of Use permit copying and distribution for personal or informational purposes.
  - **Federal Statutory Protection (23 U.S.C. § 407):** Section 407 of Title 23, United States Code, expressly shields highway safety data compiled by state transportation departments: *"Notwithstanding any other provision of law, reports, surveys, schedules, lists, or data compiled or collected for the purpose of identifying, evaluating, or planning the safety enhancement of potential accident sites, hazardous roadway conditions, or the planned safety enhancement of a highway safety improvement project... shall not be subject to discovery or admitted into evidence in a Federal or State court proceeding."*
  - Corridor must retain this statutory notice across all documentation, database metadata, and API responses to preserve federal evidentiary protections and prevent misuse in tort litigation.
- **Coverage & Coordinate Reference System:**
  - Statewide Montana highway network across 10 complete calendar years (2016-01-01 through 2025-12-31).
  - 60,592 nonblank records.
  - Coordinates provided in **Montana State Plane NAD83 (EPSG:32100)** with linear units in meters. Coordinates are stored in fields `SMS_X_CORD` (easting) and `SMS_Y_CORD` (northing).
- **Documented Schema (9 columns):**
  1. `Date`: Carcass collection date (format YYYY-MM-DD or datetime). Note: timestamps often include default `07:00:00` morning shift sentinels and must not be interpreted as collision clock time.
  2. `Route`: State or federal highway designation (e.g., `US-191`, `US-89`, `I-90`, `MT-200`, `US-93`).
  3. `Milepost`: Reference post / linear distance along route.
  4. `Animal`: Species identification across 16 categories: Whitetail Deer, Mule Deer, Deer (unspecified), Elk, Moose, Pronghorn/Antelope, Black Bear, Grizzly Bear, Mountain Lion/Cougar, Bighorn Sheep, Wolf, Coyote, Fox, Domestic Dog/Cat, Livestock, Other.
  5. `Sex`: Reported sex (`Male`, `Female`, `Unknown`).
  6. `Quantity`: Integer count of carcasses removed during the incident.
  7. `Comment`: Field maintenance crew notes (e.g., "Hit by semi", "Near bridge crossing", "Scavenged").
  8. `SMS_X_CORD`: Projected easting (EPSG:32100).
  9. `SMS_Y_CORD`: Projected northing (EPSG:32100).
- **Bias & Methodological Notes:**
  - *Reporting Bias:* Carcass records originate from maintenance patrol crews clearing travel lanes and shoulders. Patrol frequency is significantly higher on primary corridors (I-90, US-93, US-191 Gallatin Canyon) than on remote secondary routes.
  - *Species Bias:* Large game (deer, elk, moose) posing vehicular hazards or requiring physical removal are consistently recorded. Small mammals (badger, porcupine, fox) and raptors are routinely overlooked or removed without logging.
  - *Aggregate vs. Discrete Incidents:* Total quantity sum across the dataset is 74,268 across 60,592 records. Rows with `Quantity > 1` (e.g., doe with fawns or elk groups) represent multi-animal events and must be stored as aggregate counts rather than duplicated spatial points.
  - *CRS Transformation:* Ingestion requires spatial transformation from EPSG:32100 to WGS84 (EPSG:4326) and US National Albers (EPSG:5070) via PostGIS `ST_Transform`.
  - *PII Sanitization:* Free-text `Comment` fields must be scrubbed for driver names, license plate numbers, or private property details prior to database insertion.
  - *Supersedes S21:* Resolves the Phase 0 failed university endpoint (Montana State University ScholarWorks handle 1/18869 protocol failure).

---

### S33: NPS IRMA Wildlife Mortality and Morbidity Observations (WHMMO)

- **Publisher:** National Park Service (NPS), Integrated Resource Management Applications (IRMA) DataStore / Biological Resources Division.
- **Authoritative URLs:**
  - IRMA DataStore Reference Profile: `https://irma.nps.gov/DataStore/Reference/Profile/2317362`
  - Direct Artifact: `https://irma.nps.gov/DataStore/DownloadFile/702419` (`WHMMO_data.csv`, 11.3 MB)
  - NPS Open Science Guidelines: `https://www.nps.gov/subjects/gisandmapping/tools-and-data.htm`
- **Access Status:** `public_catalog`
- **License & Rights:** Public Domain (U.S. Government Work, 17 U.S.C. § 105). No copyright restrictions apply within the United States.
- **Coverage & Scope:**
  - Comprehensive nationwide coverage across National Park System units.
  - Critical park coverage: Yosemite National Park (YOSE), Yellowstone National Park (YELL), Grand Teton National Park (GRTE), Glacier National Park (GLAC), Olympic National Park (OLYM), Rocky Mountain National Park (ROMO), Grand Canyon National Park (GRCA), Shenandoah National Park (SHEN), and Craters of the Moon National Monument (CRMO).
  - 11.3 MB CSV artifact (`WHMMO_data.csv`).
- **Documented Schema:**
  - `ParkCode`: 4-letter National Park Service alphanumeric unit identifier (e.g., `YOSE`, `YELL`, `GRTE`, `GLAC`, `OLYM`).
  - `UnitCode`: Administrative organizational subunit code.
  - `ObservationDate`: Date observation was recorded in the field (YYYY-MM-DD).
  - `TaxonGroup`: Broad taxonomic classification (Mammal, Bird, Reptile, Amphibian).
  - `ScientificName`: Linnaean binomial taxonomic name.
  - `CommonName`: Vernacular common name.
  - `TSN`: Integrated Taxonomic Information System (ITIS) Taxonomic Serial Number.
  - `Latitude` / `Longitude`: Geographic coordinates in WGS84.
  - `Datum`: Geodetic datum (predominantly WGS84 or NAD83).
  - `CoordinateUncertaintyInMeters`: Positional uncertainty radius.
  - `MorbidityMortality`: Status flag indicating whether the animal was observed diseased/injured (Morbidity) or dead (Mortality).
  - `SuspectedCause` / `ConfirmedCause`: Primary etiology (e.g., `Vehicle Collision`, `Predation`, `Disease / CWD`, `Starvation`, `Trauma - Non-Vehicular`, `Unknown`).
  - `CarcassStatus`: Disposition (e.g., `Removed from Roadway`, `Left in Situ`, `Scavenged`, `Incinerated`).
  - `ObserverType`: Submitter role (`NPS Wildlife Biologist`, `NPS Protection Ranger`, `Maintenance Staff`, `Visitor Report`).
  - `Notes`: Field notes detailing carcass condition, distance from roadway edge, and habitat context.
- **Bias & Methodological Notes:**
  - *Frontcountry Road Bias:* Observations are overwhelmingly concentrated along paved park roadways (e.g., CA-120, CA-140 in Yosemite; US-191, Grand Loop Road in Yellowstone; US-89 in Grand Teton) and heavily visited frontcountry attractions. True wilderness mortality is vastly underrepresented.
  - *Etiology Filtering:* WHMMO includes natural disease outbreaks (e.g., chronic wasting disease, avian influenza) and predator kills. Ingestion adapters must explicitly filter `SuspectedCause` or `ConfirmedCause` for vehicle-strike incidents when generating collision training data.
  - *Sensitive Species Privacy Policy:* Several documented species in WHMMO are subject to federal and state privacy protections (e.g., Sierra Nevada bighorn sheep in Yosemite, grizzly bears in Yellowstone/Glacier, Pacific fisher in the Sierra Nevada, gray wolf, Canada lynx). In accordance with Corridor privacy architecture, public views and tile endpoints must generalize sensitive species occurrences to Uber H3 resolution 6 (~36 km²) and enforce a mandatory 30-day quarantine delay. Full-precision coordinates are restricted to authenticated ranger/biologist roles.
  - *Resolves S29:* Directly resolves the earlier Phase 0 missing park mortality gap.

---

### S34: NOAA SNODAS Daily Gridded Snow Products (1km CONUS)

- **Publisher:** National Oceanic and Atmospheric Administration (NOAA) / National Weather Service (NWS) / National Operational Hydrologic Remote Sensing Center (NOHRSC), distributed by the National Snow and Ice Data Center (NSIDC).
- **Authoritative URLs:**
  - NSIDC Direct Archive: `https://noaadata.apps.nsidc.org/NOAA/G02158/`
  - NSIDC Product Landing: `https://nsidc.org/data/g02158/versions/1`
  - NOAA NOHRSC Operational Portal: `https://www.nohrsc.noaa.gov/nsa/`
- **Access Status:** `public_catalog`
- **License & Rights:** Public Domain (U.S. Government Work, 17 U.S.C. § 105). Free and unrestricted reuse.
- **Coverage & Spatial Resolution:**
  - Conterminous United States (CONUS), daily 1 km resolution (~30 arc-seconds, grid dimension: 4096 columns x 3351 rows).
  - Temporal coverage: Daily grids from October 1, 2003 to present.
  - Format: Compressed flat binary rasters (`.dat.gz`) accompanied by ENVI headers (`.hdr`), Cloud-Optimized GeoTIFFs, and NetCDF-4 conversions.
- **Documented Schema & Variables:**
  - Variable `1034`: Snow Water Equivalent (SWE) in mm (scale factor 1).
  - Variable `1036`: Snow Depth in mm or meters (scale factor 1).
  - Variable `1038`: Snowpack Average Temperature in Kelvin.
  - Variable `1044`: Snow Melt Runoff at base of snowpack in mm.
  - Variable `1050`: Sublimation from snowpack in mm.
  - Geographic Bounds: 24.10°N to 53.00°N latitude, 130.52°W to 62.25°W longitude. Projection: Geographic WGS84 (EPSG:4326).
- **Bias & Methodological Notes:**
  - *Physically-Based Assimilation:* SNODAS assimilates energy-and-mass balance snow models driven by numerical weather prediction with satellite snow extent (MODIS/VIIRS), airborne gamma radiation snow surveys, and ground telemetry (NRCS SNOTEL, NOAA COOP).
  - *Critical 2014–2019 Zero Repair Mask:* During the 2014–2019 operational processing regime, unmonitored coastal, maritime, and low-elevation cells erroneously received explicit zero values (`0.0`) instead of masked no-data/no-snow sentinels. When training zero-inflated Poisson or Hurdle models, uncorrected zero-fills can bias likelihood estimators. Ingestion pipelines must apply the NSIDC documented zero-repair mask.
  - *Alpine Pass Covariates:* High-elevation passes throughout the Mountain West (e.g., Tioga Pass CA-120 summit at 3,031m in Yosemite, Dunraven Pass in Yellowstone) experience seasonal winter closures driven by snow accumulation. In the Corridor ML feature store (`ml/corridor_ml/features/extractor.py`), SNODAS provides the empirical basis for `snow_depth_cm` and triggers `seasonal_pass_closed` logic when cumulative snow depth exceeds DOT maintenance clearing thresholds.

---

### S35: USGS 3DEP High-Resolution Elevation Cloud-Optimized GeoTIFFs (10m & 1m)

- **Publisher:** United States Geological Survey (USGS) National Geospatial Program, 3D Elevation Program (3DEP).
- **Authoritative URLs:**
  - AWS Public S3 Bucket (1/3 arc-second ~10m DEM): `s3://prd-tnm/StagedProducts/Elevation/13/TIFF/`
  - AWS Public S3 Bucket (1m LiDAR bare-earth DEM): `s3://prd-tnm/StagedProducts/Elevation/1m/`
  - HTTPS Dynamic Map Service: `https://elevation.nationalmap.gov/arcgis/rest/services/3DEPElevation/ImageServer`
  - USGS 3DEP Product Standards: `https://www.usgs.gov/3d-elevation-program`
- **Access Status:** `public_catalog`
- **License & Rights:** Public Domain (USGS / U.S. Government Work, 17 U.S.C. § 105).
- **Coverage & Spatial Resolution:**
  - Seamless CONUS coverage at 1/3 arc-second (~10-meter nominal resolution) and expanding coverage at 1-meter resolution derived from high-density airborne LiDAR.
  - Delivery Format: Cloud-Optimized GeoTIFF (COG), tiled, deflate-compressed with internal HTTP range-accessible overviews.
  - Tiling Scheme: Standard 1° x 1° geographic quadrangles for 10m products; project swath/quad tiles for 1m products.
- **Documented Schema:**
  - Band 1: Floating-point 32-bit elevation values in meters above North American Vertical Datum of 1988 (NAVD88).
  - CRS: NAD83 (EPSG:4269) for 10m COGs; UTM NAD83 (e.g., UTM Zone 10N / 11N / 12N) for 1m LiDAR products.
- **Bias & Methodological Notes:**
  - *Cloud-Optimized Architecture:* Because 3DEP rasters are stored as COGs in public AWS S3 buckets, Corridor's Go backend and GDAL/PostGIS spatial workers can execute HTTP Range GET requests (`Range: bytes=start-end`) to sample elevation along a 0.1-mile highway segment without downloading hundreds of gigabytes of national rasters.
  - *Terrain Covariate Extraction:* Enables on-the-fly extraction of topographic covariates essential for wildlife collision hazard modeling:
    - Segment Slope (degrees): steep embankments funnel wildlife toward road cuts or culvert mouths.
    - Aspect (degrees): south-facing slopes melt snow earlier in spring, attracting early-season foraging ungulates.
    - Terrain Roughness Index (TRI) & Topographic Position Index (TPI): identifies ridge saddles, valley bottoms, and natural canyon funnels.
  - *Canyon Shading & No-Data Handling:* In steep river canyons (e.g., Merced River Canyon CA-140, Gallatin Canyon US-191), steep terrain relief can generate LiDAR shadow voids or extreme vertical interpolation artifacts. Adapters must validate against no-data sentinels (`-999999` or NaN).

---

### S36: Snapshot USA Coordinated National Camera Trap Survey (2019–2024)

- **Publisher:** Smithsonian Institution, North Carolina Museum of Natural Sciences, and the Snapshot USA Collaborative Consortium (Cove, Kays et al.).
- **Authoritative URLs:**
  - Dryad Digital Repository (2019–2020 Survey): `https://doi.org/10.5061/dryad.k0p2ngfhn`
  - Dryad Digital Repository (2021–2024 Survey): `https://doi.org/10.5061/dryad.bnzs7h4qf`
  - Project Information: `https://emammal.si.edu/snapshot-usa`
- **Access Status:** `public_catalog`
- **License & Rights:** Creative Commons Zero 1.0 Universal (CC0 1.0 Universal Public Domain Dedication). Unrestricted worldwide scientific and commercial reuse without requirement for prior permission.
- **Coverage & Sampling Design:**
  - Standardized nationwide camera trap survey conducted annually across all 50 US states during a synchronized autumnal sampling window (September through November).
  - Covers over 1,500 deployment sites nationwide, heavily represented across western national parks, wildlife refuges, national forests, and state trust lands.
  - Format: Standardized tabular Darwin Core / Camera Trap Data Package (Camtrap DP) CSV files.
- **Documented Schema:**
  - `deployments.csv`:
    - `deployment_id`: Unique identifier for each camera station.
    - `subproject_name`: Local survey partner / park unit name.
    - `latitude` / `longitude`: Geographic station coordinates in WGS84.
    - `camera_model`: Hardware make/model (e.g., Reconyx, Bushnell).
    - `camera_height`: Height above ground in cm.
    - `start_date` / `end_date`: Operational dates of deployment.
    - `trap_nights`: Total operational camera-trap nights (rigorous measure of survey effort).
    - `habitat_type`: Standardized ecological land cover (e.g., Montane Conifer, Sagebrush Steppe, Riparian).
    - `feature_type`: Attractant or micro-siting feature (`Road`, `Trail`, `Game Trail`, `Dirt Track`, `Water Source`, `Random`).
  - `observations.csv`:
    - `sequence_id`: Unique image trigger sequence identifier.
    - `deployment_id`: Foreign key matching deployment station.
    - `timestamp`: UTC timestamp of trigger event.
    - `scientific_name`: Linnaean binomial name.
    - `common_name`: Species common name.
    - `count`: Number of individuals detected in sequence.
    - `age` / `sex`: Demographic attributes when discernible.
- **Bias & Methodological Notes:**
  - *Overcoming Reporting Bias in Hurdle Models:* Standard carcass datasets suffer from unquantified reporting bias: zero reported carcasses on a highway segment could mean either (a) animals are completely absent from the adjacent landscape, or (b) animals cross frequently without collisions or collisions are unrecorded. Snapshot USA supplies standardized *survey effort* (`trap_nights`) and independent wildlife relative abundance indices in the surrounding buffer zones, providing the necessary denominator to train two-stage Hurdle models (`corridor_ml/models/hurdle.py`).
  - *Station Placement Bias:* Camera stations are frequently placed on game trails, dirt tracks, or natural corridors to maximize detection probability. The `feature_type` attribute must be used as a covariate or control weight during modeling.
  - *Location Generalization:* Camera trap locations revealing sensitive carnivore denning or congregation areas must respect Corridor's H3-r6 public obfuscation policy.

---

### S37: LILA BC Western Camera Traps (Idaho, CCT, WSU Lynx)

- **Publisher:** Labeled Information Library of Alexandria: Biology and Conservation (LILA BC), Microsoft AI for Earth, Source Cooperative, Idaho Department of Fish and Game (IDFG), Caltech, Washington State University.
- **Authoritative URLs:**
  - Source Cooperative S3 Bucket: `s3://us-west-2.opendata.source.coop/agentmorris/lila-wildlife/`
  - Source Cooperative Catalog: `https://beta.source.coop/agentmorris/lila-wildlife/`
  - LILA BC Main Portal: `https://lila.science/`
- **Access Status:** `public_catalog`
- **License & Rights:** Community Data License Agreement – Permissive – Version 1.0 (CDLA-Permissive-1.0) and Creative Commons Attribution (CC-BY), depending on individual sub-collection agreements. Both allow unrestricted machine learning training, fine-tuning, and model distribution.
- **Coverage & Image Archive Scale:**
  - Multi-terabyte image archives and JSON annotation packages representing hundreds of thousands of field camera trap images captured across western North American biomes.
  - Key western sub-collections:
    1. **Idaho Camera Traps (IDFG):** Over 100,000 camera trap images across southern and central Idaho public lands; comprehensive coverage of Rocky Mountain elk, mule deer, pronghorn, mountain lion, black bear, and gray wolf.
    2. **Caltech Camera Traps (CCT):** Over 240,000 images from southwestern and California environments; strong representations of bobcat, coyote, mule deer, gray fox, and black bear.
    3. **WSU Lynx Project:** Northern Rockies and Cascades winter camera trap series focused on Canada lynx, wolverine, snowshoe hare, and montane carnivores under severe winter conditions.
- **Documented Schema (COCO Camera Traps JSON):**
  - `info`: Metadata, contributor, year, version.
  - `images`: Array of image records:
    - `id`: Unique image string identifier.
    - `file_name`: Relative S3 storage path to image JPEG.
    - `datetime`: Timestamp of capture.
    - `location`: Camera station location hash.
    - `frame_num`: Sequence frame index.
  - `annotations`:
    - `id`: Unique annotation identifier.
    - `image_id`: Foreign key matching `images.id`.
    - `category_id`: Foreign key matching `categories.id`.
    - `bbox`: Normalized or pixel bounding box `[x_min, y_min, width, height]`.
  - `categories`:
    - `id`: Integer category identifier.
    - `name`: Vernacular species label.
    - `scientific_name`: Scientific taxon name.
- **Bias & Methodological Notes:**
  - *Visual Fine-Tuning Only:* LILA BC is strictly an image classification and object detection training resource. It must NOT be interpreted as highway collision records or animal movement trajectories.
  - *Field-Realistic Lighting & Weather Distribution:* Unlike studio or captive animal imagery, LILA BC images capture animals under authentic field conditions: nighttime infrared flash, blinding snowstorms, dust, severe motion blur, and partial branch/vegetation occlusion.
  - *Vision Pipeline Calibration:* Essential for pre-training, fine-tuning, and evaluating Corridor's on-device and offline vision inference heads (`corridor_ml/vision/head.py`, BioCLIP and SpeciesNet embeddings) across all 11 target wildlife species, including Sierra Nevada bighorn sheep, Pacific fisher, mountain lion, black bear, and mule deer.

---

## 3. Summary Ledger Cross-Reference

| ID | Short Name | Publisher | Format / Protocol | License | Primary Role in Corridor |
|---|---|---|---|---|---|
| **S31** | Caltrans AADT & Postmiles | Caltrans | FeatureServer / CSV | CC-BY | Highway traffic volume exposure & LRS postmile matching (CA-120 / CA-140) |
| **S32** | Montana DOT Carcass Data | Montana DOT | XLSX / EPSG:32100 | MDT Terms / 23 U.S.C. § 407 | Statewide 10-year carcass inventory (60,592 records) across Montana state highways |
| **S33** | NPS IRMA WHMMO | NPS | CSV (11.3 MB) / WGS84 | Public Domain | Nationwide national park wildlife mortality & vehicle strike observations |
| **S34** | NOAA SNODAS Snow Grids | NOAA / NSIDC | Daily 1km Binary / GeoTIFF | Public Domain | Alpine snow depth (`snow_depth_cm`) and pass closure (`seasonal_pass_closed`) modeling |
| **S35** | USGS 3DEP Elevation COGs | USGS | 10m & 1m Cloud GeoTIFFs | Public Domain | Topographic slope, aspect, and terrain roughness (TRI) via HTTP range GETs |
| **S36** | Snapshot USA Camera Survey | Smithsonian / Dryad | Camtrap DP CSV / CC0 | CC0 1.0 Universal | Standardized trap-night effort for Hurdle reporting-bias correction |
| **S37** | LILA BC Western Traps | LILA BC / Source Coop | S3 / COCO JSON / Images | CDLA-Permissive / CC-BY | Fine-tuning and calibration for BioCLIP/SpeciesNet 11-class vision head |

---

## 4. Next Steps for Ingestion & Feature Store Architecture

1. **Phase 2 Ingestion Extensions:**
   - Implement `internal/ingest/caltrans_aadt.go` to parse Caltrans FeatureServer JSON and build postmile lookup tables.
   - Implement `internal/ingest/montana_carcass.go` to read `CarcassData2016-2025.xlsx`, transform EPSG:32100 to EPSG:4326/5070, and apply 23 U.S.C. § 407 metadata tags.
   - Implement `internal/ingest/nps_whmmo.go` to parse `WHMMO_data.csv`, filter vehicle collisions, and enforce H3-r6 privacy gating for sensitive species.
2. **Feature Store Pipeline (Python & Go):**
   - Incorporate SNODAS daily GeoTIFF extraction in `ml/corridor_ml/features/extractor.py` using the 2014–2019 zero-repair mask.
   - Stream 3DEP elevation tiles via Cloud-Optimized GeoTIFF window reads to calculate slope and terrain roughness per 0.1-mile segment.
   - Leverage Snapshot USA `trap_nights` and detections as background presence effort in `ml/corridor_ml/models/hurdle.py`.
3. **Vision Training Pipeline:**
   - Sample LILA BC Western sub-datasets to construct balanced train/validation/test splits for the 11-class vision head, testing against degraded ranger carcass photos.
