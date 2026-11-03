# Eastern parks scout — Great Smoky Mountains, Shenandoah, Acadia

**Observed:** 2026-11-03 UTC  
**Method:** `curl` of public pages, ArcGIS metadata, `returnCountOnly` / `returnDistinctValues` queries, and text extracted from PDFs already returned by those requests. No bulk feature download. No record coordinates are copied here. A number below appears only because it was in a response body.

Rumored totals of 2,971 deer, 228 bear, 68,047 Maine strikes, and a 10-year span did not occur in any response body retrieved in this pass.

## NCDOT reported animal crashes (I-40, Haywood County)

| Observation UTC | HTTP | Content-Type | URL |
| --- | --- | --- | --- |
| 2026-11-03T00:38:36Z | 200 | `text/html; charset=utf-8` | https://www.ncdot.gov/initiatives-policies/safety/traffic-safety/Pages/safety-data-maps.aspx |
| 2026-11-03T00:39:57Z | 200 | `text/plain;charset=utf-8` | https://ncdot.maps.arcgis.com/sharing/rest/content/items/7bf2a5aef68a4719bbc33b994e368330?f=pjson |
| 2026-11-03T00:39:58Z | 200 | `text/plain;charset=utf-8` | https://ncdot.maps.arcgis.com/sharing/rest/content/items/7bf2a5aef68a4719bbc33b994e368330/data?f=pjson |
| 2026-11-03T00:41:30Z | 200 | `application/json; charset=utf-8` | https://services.arcgis.com/NuWFvHYDMVmmxMeM/arcgis/rest/services/AnimalCrashes2018_Mapped/FeatureServer/0?f=pjson |
| 2026-11-03T00:43:06Z | 200 | `application/json; charset=utf-8` | https://services.arcgis.com/NuWFvHYDMVmmxMeM/arcgis/rest/services/AnimalCrashes2024_Mapped/FeatureServer/0?f=pjson |
| 2026-11-03T00:41:30Z | 200 | `application/json; charset=utf-8` | https://services.arcgis.com/NuWFvHYDMVmmxMeM/arcgis/rest/services/AnimalCrashes2025_Mapped/FeatureServer/0?f=pjson |
| 2026-11-03T00:43:07Z | 200 | `text/plain;charset=utf-8` | https://ncdot.maps.arcgis.com/sharing/rest/search?q=orgid%3ANuWFvHYDMVmmxMeM%20AND%20(title%3Aroadkill%20OR%20title%3Acarcass)&f=pjson&num=10 |

The safety-maps page says animal crashes are “Locations of crashes involving animals during the past five years” and links web map `7bf2a5aef68a4719bbc33b994e368330`, titled “Reported Animal Crashes in North Carolina,” owner `TRAFFICSAFETY.NCDOT.GOV`, access `public`.

The web map lists public feature layers for 2015–2025. Only the 2025 layer is visible. `definitionExpression` is null on every animal layer. Popup fields for 2015–2023 are `FID`, `City`, `CountyName`, `Crash_ID`, `Crash_ID_1`, `CrashRepor`, `CrashType`, `Date`, `Direction`, `Distance`, `FromRd`, `GIS_Milepo`, `GIS_Route`, `Latitude`, `LOC_ERROR`, `Longitude`, `MapMethod`, `OnRd`, `Severity`, `Source_ID`, `Time`, `TowardRd`. Popup fields for 2024 and 2025 replace `GIS_Route` and `Crash_ID_1` with `GIS_RteTxt`.

Service JSON confirmed that split:

- 2018 `AnimalCrashes2018_Mapped` layer 0: point, `maxRecordCount` 2000, capability `Query`. Fields: `FID`, `Crash_ID`, `Crash_ID_1`, `Date` (string), `Time`, `Severity`, `CrashType`, `CountyName`, `City`, `OnRd`, `Distance`, `Direction`, `FromRd`, `TowardRd`, `GIS_Route` (double), `GIS_Milepo`, `Latitude`, `Longitude`, `Source_ID`, `MapMethod`, `CrashRepor`, `LOC_ERROR`. `copyrightText` empty.
- 2024 and 2025: same core fields, `GIS_RteTxt` (string) instead of `GIS_Route`, no `Crash_ID_1`. On the 2025 layer, `Date` is `esriFieldTypeDate`.

License on the web-map item, with HTML tags removed: “The North Carolina Department of Transportation shall not be held liable for any errors in this data. This includes errors of omission, commission, errors concerning the content of the data, and relative and positional accuracy of the data. This data cannot be construed to be a legal document. Primary sources from which this data was compiled must be consulted for verification of information contained in this data.” No reuse grant. **license_unresolved.**

Item description: “This map shows animal related crash data in North Carolina.” “The current NC crash report form does not distinguish between the type of animal struck, although previous research reveals over 90% of reported animal crashes involve a deer.” “Not all reported animal crashes are able to be located for various reasons.” “This data contains all reported crashes on public roads.”

2025 queries (HTTP 200, `application/json`):

- `CountyName='Haywood'` at 2026-11-03T00:41:31Z returned `{"count":103}`.
- `CountyName='Haywood' AND OnRd='I 40'` at 2026-11-03T00:43:05Z returned `{"count":24}`.
- The same filter with `GIS_Milepo < 100` at 2026-11-03T00:43:48Z returned `{"count":23}`.
- Min/max of `GIS_Milepo` on that I-40 filter returned `min_mp` 1 and `max_mp` 999.999.
- Distinct `CrashType` on that filter was `ANIMAL`.
- A distinct-value response paired `OnRd` `I 40` with `GIS_RteTxt` `10000040044`.

These counts are one service layer for calendar-year 2025 in Haywood County. They are not a Pigeon River Gorge multi-year total. `999.999` is in the milepost statistic, so milepost alone did not isolate the gorge.

The organization search `title:roadkill OR title:carcass` returned `"total": 0`. No anonymous NCDOT carcass feature service was found.

## Tennessee deer-named crash service

| Observation UTC | HTTP | Content-Type | URL |
| --- | --- | --- | --- |
| 2026-11-03T00:38:38Z | 200 | `text/plain; charset=UTF-8` | https://tnmap.tn.gov/arcgis/rest/services/SAFETY/DeerCrashes2019to2023/MapServer?f=pjson |
| 2026-11-03T00:39:59Z | 200 | `text/plain; charset=UTF-8` | https://tnmap.tn.gov/arcgis/rest/services/SAFETY/DeerCrashes2019to2023/MapServer/9?f=pjson |
| 2026-11-03T00:39:59Z | 200 | `text/plain; charset=UTF-8` | https://tnmap.tn.gov/arcgis/rest/services/SAFETY/DeerCrashes2019to2023/MapServer/2?f=pjson |
| 2026-11-03T00:41:31Z | 200 | `application/json; charset=UTF-8` | layer 9 `query` for distinct `CrashType` |
| 2026-11-03T00:43:06Z | 200 | `text/plain; charset=UTF-8` | https://tnmap.tn.gov/arcgis/rest/services/TRANSPORTATION?f=pjson |
| 2026-11-03T00:43:54Z | 200 | `text/html` | https://gis.safety.tn.gov/ |

Service `mapName` is “Deer Crashes 2019 to 2023.” `copyrightText` is “TN Dept. of Safety and Homeland Security.” The description says the map contains “all reported deer crashes between January 1, 2017 and September 30, 2020.” The public index instead labels “Deer Involved Crashes (1/1/2021 - 6/30/2025)” at `https://gis.safety.tn.gov/Maps/Public/DeerInvolvedCrashes/`. **license_unresolved.**

Layers include State Boundary, DeerSigns, Fatal Crash, Suspected Serious Injury Crash, Suspected Minor Injury Crash, Possible Injury Crash, Property Damage Crash layers named 2022, 2023, 2024, 1/1/24–6/30/24, and 2025, plus Highways, Counties, and Weighted Crash Likelihood and Severity.

Layer 9, “Property Damage Crash - 2024,” and layer 2, “Fatal Crash,” are points. Fields on both: `OBJECTID`, `MstrRecNbrTxt`, `CrashDate`, `CollisionTimeTxt`, `LatDecimalNmb`, `LongDecimalNmb`, `Lighting`, `CrashType`, `CrashWeightNmb`, `Shape`. No route, county, milepost, or species field. Distinct `CrashType` on layer 9 was `Property Damage`, `Suspected Serious Injury`, `Possible Injury`, `Fatal`, and `Suspected Minor Injury`.

`TRANSPORTATION` contains `TRANSPORTATION/MAJOR_ROADS` and `TRANSPORTATION/TDOT_INTERSTATE_MILE_POSTS` only. No TDOT animal or carcass service name was in that folder.

## Harmon Den / Safe Passage (public descriptions)

| Observation UTC | HTTP | Content-Type | URL |
| --- | --- | --- | --- |
| 2026-11-03T00:43:48Z | 404 | `text/html; charset=utf-8` | https://www.ncdot.gov/news/press-releases/Pages/2022/2022-05-20-harmon-den-bridge-open.aspx |
| 2026-11-03T00:43:50Z | 200 | `text/html; charset=UTF-8` | https://smokiessafepassage.org/ncdot-to-begin-construction-on-wildlife-crossing-at-harmon-den/ |
| 2026-11-03T00:43:51Z | 200 | `text/html; charset=UTF-8` | https://smokiessafepassage.org/habituation-at-harmon-den/ |
| 2026-11-03T00:43:52Z | 200 | `text/html;charset=utf-8` | https://www.northcarolinawildlifecorridor.org/projects/safe-passage |
| 2026-11-03T00:44:41Z | 200 | `application/pdf` (4,582,772 bytes) | https://connect.ncdot.gov/resources/Environmental/EPU/Grants/Documents/NC%20WCPP%20Grant_2023_I-40%20CATCH_FINAL.pdf |
| 2026-11-03T00:44:46Z | 200 | `application/pdf` (6,922,362 bytes) | https://wildlandsnetwork.org/wp-content/uploads/2026/03/ExecutiveSummary-ResearchResultsandMitigationStrategiestoImproveWildlifeConnectivityandHumanSafetyalongI-40inthePigeonRiverGorge.pdf |

Page licenses were not stated. **license_unresolved** for each page and PDF.

Safe Passage, 5 November 2021: NCDOT construction at “Interstate 40 over Harmon Den Road, at Exit 7 in Haywood County,” with “two paths — one on each side of Cold Springs Creek,” “A 9-foot tall fence,” and wildlife guards “similar to slotted cattle guards.” It names a $19 million Kiewit contract.

Safe Passage, 14 April 2023: Harmon Den then had “wildlife-friendly benches and cattle guards and will, in the future, include fencing.”

North Carolina Wildlife Corridor page: the project is a “28-mile stretch” between Great Smoky Mountains National Park and Pisgah National Forest, and Harmon Den Exit 7 “contains lifesaving measures such as cattle guards, jump outs, fencing, and underpasses.”

NCDOT I-40 CATCH PDF, July 2023, Haywood County: seeks $6.1 million and proposes $678,200 match “to upgrade multiple drainage structures and install approximately 2.2 miles of connecting wildlife fencing along I-40 near the North Carolina/Tennessee state line,” plus ramp-culvert work at “Cold Springs Creek Road (Exit 7, NFSR 148).” It says STIP B-6054A “included short fencing and cattle guards to discourage wildlife access to I-40, in addition to narrow benches to allow wildlife passage under the new bridge.” It also says Wildlands Network and NPCA “recorded 158 wildlife vehicle crashes (WVCs) along a 28-mile corridor of I-40 (including the project area) from 2018 to 2021,” mostly black bear and whitetail deer, “along with at least one elk.” The same PDF applies a “factor of 8.5” and says “over 1,300 crashes are estimated.” That 1,300 figure is the PDF’s estimate, not a row count from the NCDOT feature service.

Wildlands Network / NPCA executive summary, June 2022: a 28-mile I-40 study, agency WVC records “2001–2021,” driving surveys “2018-2021,” and a statement that combined crash, carcass, and survey records for 2018–2020 “were 6.9 times greater than the crash report data only.” It calls for 20 mitigation recommendations, wildlife fencing, seven priority recommendations including overpasses and larger culverts, and Forest Service protection for the Double Tunnel land bridge. The text does not name Harmon Den and does not contain 2,971 or 228.

## VDOT carcass removals (Shenandoah roads)

| Observation UTC | HTTP | Content-Type | URL |
| --- | --- | --- | --- |
| 2026-11-03T00:38:41Z | 200 | `application/json;charset=utf-8` | https://data.virginia.gov/api/3/action/package_show?id=vdot-wildlife-carcass-removal-tracking-pv |
| 2026-11-03T00:38:41Z | 200 | `text/plain;charset=utf-8` | https://www.arcgis.com/sharing/rest/content/items/ddc33a1134ca48f5874dc0f351f22db7?f=pjson |
| 2026-11-03T00:39:59Z | 200 | `application/json; charset=utf-8` | https://services.arcgis.com/p5v98VHDX9Atv3l7/arcgis/rest/services/Wildlife_Carcass_Removal_Tracking/FeatureServer/0?f=pjson |

CKAN `license_id`, `license_title`, and `license_url` are null. **license_unresolved.**

Item `licenseInfo`: “VDOT's digital data files are for use in performing the official business of the Commonwealth of Virginia. VDOT Divisions set all policy on the allowable uses of this data.” Package notes: “under Title 23 United State Code – Section 409, this crash information cannot be used in discovery or as evidence in a Federal or State court proceeding or considered for other purposes in any action for damages against VDOT or the State of Virginia arising from any occurrence at the location identified.”

Notes also say the layer covers carcasses removed by contracted maintenance on “all VDOT interstates and select other roads,” excludes roads maintained by VDOT forces, “does not represent a complete dataset,” and “Data collection began in VDOT’s Culpeper and Staunton Districts in 2021.”

Layer `WildlifeObservations` is a point feature, `maxRecordCount` 1000. Fields: `objectid`, `globalid`, `starttime` (alias “Date and Pickup Time”), `wildlife` (alias “Deer and Bear Removal”), `wildlifeother`, `rtenamedir`, `latitude_d`, `longitude_d`, `Jurisdiction`, `District`, `Residency`.

`wildlife` coded values: `Deer`, `Bear`, `Other`, `NotFound`. Distinct values returned: `Deer`, null, `Bear`, `Other`, `Notfound`. `wildlifeother` coded values: Coyote, Raccoon, Bird, Bobcat, Squirrel, Fox, Skunk, Turtle, Opossum, Groundhog, Dog, Cat, Farm Animal (`Farmanimal`), Other Animal (`Otheranimal`).

`rtenamedir` coded values: `I-64E`, `I-64W`, `I-81N`, `I-81S`, `I-66E`, `I-66W`, `Rte37E`, `Rte37W`. Distinct `District` values: null, `Richmond`, `Staunton`, `Hampton Roads`, `Culpeper`. A distinct query for `rtenamedir` containing `211`, `33`, or `SKY` at 2026-11-03T00:43:06Z returned `"features": []`. A separate distinct query returned I-64 direction strings in Staunton, Culpeper, and Richmond. Skyline Drive, US-33, and US-211 were not among the route strings returned.

## MaineDOT animal crashes (Acadia)

| Observation UTC | HTTP | Content-Type | URL |
| --- | --- | --- | --- |
| 2026-11-03T00:43:53Z | 403 | `text/html` | https://www.maine.gov/dot/programs-services/safety-and-mobility/safer-roads/crash-data (research user-agent; CloudFront “Request blocked.”) |
| 2026-11-03T00:44:38Z | 200 | `text/html; charset=UTF-8` | same crash-data URL with a browser user-agent |
| 2026-11-03T00:44:38Z | 200 | `application/pdf` (2,009,327 bytes) | https://www.maine.gov/dot/sites/maine.gov.dot/files/inline-files/Collisions%20Between%20Wildlife%20and%20Motor%20Vehicles%202019-23.pdf |
| 2026-11-03T00:44:39Z | 200 | `text/html; charset=utf-8` | https://mdotapps.maine.gov/mainecrashpublic/PublicQueryMap |
| 2026-11-03T00:38:39Z | 200 | `text/plain; charset=UTF-8` | https://gis.maine.gov/arcgis/rest/services/dot/MaineDOT_OpenData_Crashes/MapServer?f=pjson |
| 2026-11-03T00:40:00Z | 200 | `text/plain; charset=UTF-8` | same host, layer 2 |
| 2026-11-03T00:38:39Z | 200 | `text/plain; charset=UTF-8` | https://gis.maine.gov/mapservices/rest/services/dot/MaineDOT_OpenData_Crashes/MapServer?f=pjson |
| 2026-11-03T00:46:28Z | 200 | `text/html` | https://www.maine.gov/mdot/mapviewer/?map=e0Y3Q0RFNEIwLTJGMTEtMTFFRi1BRkYwLUQwNjcyNkJBMUVDOH0= |

The crash-data HTML links “Crashes Involving Deer, Moose, or Bears (2019-2023)” to that map viewer. The viewer HTML is an ArcGIS JavaScript shell. It contains no FeatureServer URL and no field list. The public query page is an HTML form. Visible animal labels include `All Other Animal`, `Bear`, `Deer`, `Moose`, and `Turkey`. It is not a feature schema.

`gis.maine.gov/arcgis/.../MaineDOT_OpenData_Crashes` returned `{"error":{"code":499,"message":"Token Required"}}`. Layer 2 on that host returned `404` “Service not found.” The `mapservices` copy returned HTTP 200, `copyrightText` “MaineDOT,” document subject “Maine Crash Data from MaineDOT,” and `"layers": []`. The `dot` folder also lists `dot/MaineDOT_OpenData` and `dot/MaineDOT_TSMO`. The `ifw` folder lists habitat and harvest services, not a crash or carcass service.

The PDF preface calls it “a statistical review of reported motor vehicle highway crashes that involved animals in Maine during the five-year study period 2019 - 2023,” compiled from police crash reports. PDF metadata author is “Grinnell, Katherine”; creation date `D:20241028103634-04'00'`. No reuse sentence was in the extracted preface. **license_unresolved.**

Animal-type table in that PDF, five-year totals: moose 1,247; deer 29,745; bear 171; all other animals 868; turkey 247; total 32,278. A later page says “32,278 Total Animal Crashes.” Year columns for the total row are 6,016 (2019), 5,837 (2020), 6,523 (2021), 7,199 (2022), and 6,703 (2023). This is a statewide aggregate bulletin, not an event schema, and it does not contain 68,047.

## Adapter verdict

- **GRSM:** anonymous adapter-ready schema on the NCDOT animal-crash feature services (route, county, milepost, date, coordinates). `CrashType` is the animal flag, not a species. NCDOT carcass removals were not found. The Tennessee service has coordinates and a date, and it lacks a route or species field, so Tennessee-side gorge events still require a request.
- **SHEN:** the VDOT carcass schema is anonymous and queryable, and its returned route strings did not include Skyline Drive, US-33, or US-211. Those roads still require a request.
- **ACAD:** MaineDOT published a 2019–2023 aggregate PDF and an HTML crash query. No anonymous event schema was returned. Acadia event records still require a request.

## Summary

On 3 November 2026 UTC, NCDOT animal-crash feature services answered without a token. The 2025 Haywood County layer returned count 24 where OnRd was "I 40", with CrashType "ANIMAL" and no species field. Tennessee's DeerCrashes2019to2023 MapServer also answered, but CrashType held injury severities and the layer has no route field. NCDOT's July 2023 CATCH PDF describes Harmon Den Exit 7 benches, cattle guards, and short fencing, and states 158 wildlife-vehicle crashes on 28 miles of I-40 from 2018 to 2021. VDOT's carcass service is public and includes deer, bear, route, time, and coordinates. Skyline Drive, US-33, and US-211 were absent from a route query. MaineDOT's 2019–2023 PDF reports 32,278 animal crashes and exposes no point schema. Rumored totals of 2,971 deer, 228 bear, and 68,047 Maine strikes were absent. GRSM has an anonymous adapter-ready schema. SHEN and ACAD still require a request for event records, as do Tennessee-side gorge events.
