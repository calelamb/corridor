# Utah and Nevada service check

**Checked:** 2026-11-03 UTC (local evening of 2026-11-02).  
**Method:** Anonymous `curl` against ArcGIS REST and ArcGIS Hub search. Each row below records HTTP status, `Content-Type`, and the response `Date` header or the UTC clock time of the request. Field names come from `?f=pjson`. Counts come from `returnCountOnly=true` and are the JSON `count` field. No token was supplied. No bulk extract was saved into Git.

S22 in `data/SOURCES.md` still stands: the Utah Roadkill Reporter app requires a request. This note does not change that.

## What responded

| Service | Result |
| --- | --- |
| UDOT Crash Locations MapServer | HTTP 200. Statewide crash points, with a `WILD_ANIMAL_RELATED` flag. Not a carcass inventory. |
| UDOT UPlan carcass layers | HTTP 200. Contract geography and one published roadkill extract. Not a current statewide carcass feed. |
| UDWR `Wildlife_Crossing_3_view` | HTTP 200. `count` 119. |
| Older `DWR_Wildlife_Crossings` | HTTP 200. `count` 47. Data last edit 2016-04-13. |
| `Wildlife_Crossing_public` | HTTP 200 body is error code 499, `Token Required`. |
| `Fence_Survey_PublicView` | HTTP 200 body is error code 400, `Invalid URL`. |
| UDOT Wildlife Mitigation Projects MapServer | HTTP 404. |
| UWMI `migration_corridors` | HTTP 200. Linked from the migration page. `count` 82. |
| NDOT wildlife crossings | HTTP 200 on both the department server and the GeoHub copy. Layer `Crossings` `count` 33. Six Pequop Summit rows on `IR80`. |
| Utah Roadkill Reporter site | HTTP 200 HTML app page. No public feature service in the page. |

A parser was added for the crash MapServer schema only: `internal/ingest/udot.go`. `go test -count=1 -run TestUDOT ./internal/ingest/` passed (`ok corridor/internal/ingest`). The test rows are synthetic.

## UDOT crash locations

Service root, checked 2026-11-03T00:37:15Z:

`https://maps.udot.utah.gov/central/rest/services/TrafficAndSafety/Crash_Locations/MapServer?f=pjson`

- HTTP 200, `content-type: text/plain; charset=UTF-8`, `Date: Tue, 03 Nov 2026 00:37:15 GMT`.
- Service description: crash locations with personal identification removed, date range 1/1/2017 to current, refreshed nightly.
- Item `06c397a4c16242e4bc2928053f856def` snippet says the service shows crashes back to 1/1/2018. The layer list has no 2017 layer. A query to layer id 11, which is not in that list, returned HTTP 200 with `{"error":{"code":404,"message":"Layer not found"}}`.
- `spatialReference.wkid` 26912. `maxRecordCount` 2000.
- Item `licenseInfo`: "This data is for informational purposes and must be field verified prior to being used on a project. UDOT makes no warranty with respect to the accuracy, completeness or usefulness of this content or consequential damages resulting from the use or misuse of the content or any of the information contained herein." Contact given: `udotgis@utah.gov`. That is a disclaimer, not an open-license grant. Redistribution: `license_unresolved`.

Layers returned by the MapServer:

| Layer id | Name |
| --- | --- |
| 14 | Crash Locations 2025 |
| 13 | Crash Locations 2024 |
| 12 | Crash Locations 2023 |
| 7 | Crash Locations 2022 |
| 6 | Crash Locations 2021 |
| 8 | Crash Locations 2020 |
| 9 | Crash Locations 2019 |
| 10 | Crash Locations 2018 |

Layer 14 field names (`?f=pjson`, HTTP 200, `text/plain; charset=UTF-8`, 2026-11-03T00:38:00Z): `OBJECTID`, `SHAPE`, `CRASH_ID`, `CRASH_DATETIME`, `COUNTY_NAME`, `REGION_NAME`, `MAIN_ROAD_NAME`, `ROUTE`, `START_ACCUM`, `ROUTE_DIRECTION`, `ROADWAY_TYPE_CD`, `EXIT_NUMBER`, `RAMP_ID`, `LOCATION_DESC`, `LAT_UTM_Y`, `LONG_UTM_X`, `CRASH_SEVERITY_DESC`, `LIGHT_CONDITION_DESC`, `WEATHER_CONDITION_DESC`, `MANNER_COLLISION_DESC`, `PAVEMENT_DESC`, `ROADWAY_SURF_CONDITION_DESC`, `ROADWAY_JUNCT_FEATURE_DESC`, `ROAD_JURISDICTION_DESC`, `WORK_ZONE_RELATED_YNU`, `HORIZONTAL_ALIGNMENT_DESC`, `VERTICAL_ALIGNMENT_DESC`, `ROADWAY_CONTRIB_CIRCUM_DESC`, `NUMBER_VEHICLES_INVOLVED`, `FIRST_HARMFUL_EVENT_DESC`, `MOTOR_CARRIER_INVOLVED_YN`, `NUMBER_FATALITIES`, `NUMBER_FOUR_INJURIES`, `NUMBER_THREE_INJURIES`, `NUMBER_TWO_INJURIES`, `NUMBER_ONE_INJURIES`, `PEDESTRIAN_INVOLVED`, `BICYCLIST_INVOLVED`, `MOTORCYCLE_INVOLVED`, `DUI`, `AGGRESSIVE_DRIVING`, `DISREGARD_TRAFFIC_CONTROL_DEV`, `DISTRACTED_DRIVING`, `DROWSY_DRIVING`, `SPEED_RELATED`, `IMPROPER_RESTRAINT`, `IMPROPER_RESTRAINT_FATAL_TOT`, `IMPROPER_RESTRAINT_LEVEL4_TOT`, `UNRESTRAINED`, `UNRESTRAINED_FATAL_TOT`, `UNRESTRAINED_LEVEL4_TOT`, `WILD_ANIMAL_RELATED`, `DOMESTIC_ANIMAL_RELATED`, `ROADWAY_DEPARTURE`, `OVERTURN_ROLLOVER`, `COMMERCIAL_MOTOR_VEH_INVOLVED`, `TEENAGE_DRIVER_INVOLVED`, `OLDER_DRIVER_INVOLVED`, `SINGLE_VEHICLE`, `TRAIN_INVOLVED`, `RAILROAD_CROSSING`, `TRANSIT_VEHICLE_INVOLVED`, `COLLISION_WITH_FIXED_OBJECT`, `HIT_AND_RUN`, `DEBRIS`, `DIVIDED_HIGHWAY`, `FUNCTIONAL_CLASS`, `ERR_CD`, `CURRENT_AS_OF_DATE`.

`WILD_ANIMAL_RELATED` and `DOMESTIC_ANIMAL_RELATED` are strings of length 1. There is no species field. Layer 10 (2018) field names matched layer 14. Other years were not field-listed; the `WILD_ANIMAL_RELATED='Y'` query succeeded on each listed layer, so that field exists there.

Distinct values on layer 14, `returnDistinctValues=true` for `WILD_ANIMAL_RELATED`: `Y` and `N`.

`count` from `returnCountOnly=true` and `where=WILD_ANIMAL_RELATED='Y'` (HTTP 200, `text/plain; charset=UTF-8`, 2026-11-03T00:39:33Z through 00:39:37Z):

| Layer | Year | `count` |
| --- | --- | --- |
| 14 | 2025 | 1372 |
| 13 | 2024 | 2482 |
| 12 | 2023 | 2496 |
| 7 | 2022 | 2551 |
| 6 | 2021 | 2389 |
| 8 | 2020 | 2357 |
| 9 | 2019 | 2822 |
| 10 | 2018 | 2901 |

Layer 14 with `where=1=1` returned `{"count":33847}` at 2026-11-03T00:39:06Z. Other years were not counted without the wildlife filter. No crash features were downloaded. The parser decodes ArcGIS date fields as epoch milliseconds; that encoding was not confirmed on a live crash row.

The `maps.udot.utah.gov/central/rest/services` folder list (HTTP 200, 2026-11-03T00:38:53Z) has no service whose name contains wildlife, carcass, animal, crossing, fence, migration, roadkill, or deer. `Crash_Locations` is the wildlife-relevant service in that catalog because of the flag, not because of its name.

## Carcass and roadkill

`https://wildlifecollisions.utah.gov/` returned HTTP 200, `text/html; charset=utf-8`, at 2026-11-03T00:40:35Z. The page is the phone-app landing page. It describes reports of location, species, gender, age class, and photos, and it links to the Apple and Google apps. It does not publish a bulk URL or a feature service. DWR Hub dataset search `carcass OR roadkill` returned `numberMatched: 0` at `2026-11-03T00:40:32.417Z`. S22 stays `requires_request`.

UPlan org `pA2nEVnB6tquxgOW` does have public layers. They are not a replacement for S22.

### US 40 Roadkill Reports

`https://services.arcgis.com/pA2nEVnB6tquxgOW/arcgis/rest/services/US_40_Roadkill_Reports/FeatureServer/0`

- HTTP 200, `application/json; charset=utf-8`, 2026-11-03T00:44:08Z.
- Point layer, `latestWkid` 3857, `maxRecordCount` 2000, `editingInfo.dataLastEditDate` 2021-10-08T18:00:53Z.
- `count` for `where=1=1`: `{"count":5943}`.
- Fields: `FID`, `REPORT_DAT`, `SPECIES`, `AGE_CLASS`, `GENDER`, `HIGHWAY_RO`, `UDOT_REGIO`, `UDWR_REGIO`, `XYPHOID`, `RESPONDER_`, `COMMENTS`, `TAG_COLLAR`, `GPS_ACCURA`, `ADDRESS`, `RESPONDER1`, `RESPONDE_1`, `RESPONDE_2`, `WMU`, `ROUTE`, `ROUTE_DIR`, `MILEPOST`.
- Responder, address, and comment fields were not sampled.
- Distinct `ROUTE` values from `returnDistinctValues=true`: `0087`, `0311`, `0208`, `0045`, `0040`, `0191`, `0149`. The service is not limited to route `0040`.
- Item `c3d70ac43ec94143a3eb8ee76d7e6400` title is "US 40 Roadkill Reports". Snippet: "Roadkill reports from the Utah Carcass Data and Mitigation Efforts Data Access for use on the U.S. 40 Corridor Study story map." Owner `khorton@wcecengineers.com_uplan`. `licenseInfo` is the same UDOT disclaimer quoted above. Redistribution: `license_unresolved`.

### Contract geography, not carcass observations

These responded HTTP 200 with `application/json` and are polylines or polygons:

| URL | Geometry | `count` | `dataLastEditDate` | Fields checked |
| --- | --- | --- | --- | --- |
| `.../R2CarcassRemoval2018/FeatureServer/0` | polyline, wkid 26912 | 21 | 2016-02-18T23:28:27Z | `OBJECTID`, `contract__number`, `contractor`, `phone`, `cost`, `expires`, `pick_up_days`, `direction`, `route`, `area`, `beg_mile_post`, `end_mile_post`, `miles`, `station`, `LOC_ERROR` |
| `.../CarcassRemovalContracts2019/FeatureServer/0` | polyline, latestWkid 3857 | 38 | 2018-09-18T17:04:02Z | `FID`, `RouteLabel`, `Direction`, `StartingMP`, `EndingMP`, `TotalMiles`, `CostPerMil`, `Station`, `Area`, `Contractor`, `ContractNu`, `Phone`, `ContractTe`, `MainContac`, `AltContact`, `AltPhone`, `Link`, `Email`, `Email2`, `Shape__Length` |
| `.../Carcass_Removal_Contracts/FeatureServer/0` | polyline `contracts_linework`, latestWkid 3857 | 82 | 2017-07-18T22:51:16Z | `OBJECTID`, `REGION`, `ROUTE`, `UplanRoute`, `BMP`, `BegDisc`, `EMP`, `EndDisc`, `Length`, `Shape__Length` |

Item text for the 2018 layer: "This map shows the areas in which Region 2 has contracted out carcass removal areas on state routes." The same UDOT disclaimer applies to the 2018 and 2019 items. Phone, email, and contractor values were not read. `CarcassRemovalContractsData_gdb` also returned HTTP 200 and lists layers `R3_Carcass_Removal_Contracts` (polyline) and `Area_Check` (polygon). Those two layers were not counted and their fields were not read.

### I-70 study points

Both are on the UPlan org, HTTP 200, `application/json`, points, `latestWkid` 3857. Schema fields are only `Latitude`, `Longitude`, and `ObjectId`. There is no species, route, or date field in the layer schema. Species is only in the catalog title.

| Service | Catalog title | `count` | `dataLastEditDate` |
| --- | --- | --- | --- |
| `elk_vehicle_collisions_points/FeatureServer/0` | I-70 Elk Collisions RP 72 to 96 Feature Layer | 111 | 2025-11-10T20:58:37Z |
| `mule_deer_collisions_points/FeatureServer/0` | I-70 Mule Deer Collisions RP 72 to 96 Feature Layer | 72 | 2025-11-10T21:19:16Z |

No features were downloaded, so the milepost range in the title was not checked against rows.

## Utah crossing structures

The live anonymous inventory with `count` 119 is:

`https://services.arcgis.com/ZzrwjTRez6FJiOq4/arcgis/rest/services/Wildlife_Crossing_3_view/FeatureServer/0/query?where=1%3D1&returnCountOnly=true&f=pjson`

Response body: `{"count":119}`. HTTP 200, `application/json; charset=utf-8`, request UTC 2026-11-03T00:41:57Z. Item `a54a3cc299b943339e8745b0d6a58d17` is public, title "Utah Wildlife Crossings 2.0", owner `dwr.aehrhart`, `licenseInfo` empty. Redistribution: `license_unresolved`.

Layer 0 name `Wildlife_Crossing`, point, `wkid` 4326. `editingInfo.dataLastEditDate` 2022-09-12T18:31:15Z. `schemaLastEditDate` 2024-05-29T16:11:05Z. Distinct `CrossingType` values: `overpass` and `underpass`.

Fields: `objectid`, `globalid`, `CrossingName`, `CrossingType`, `OverpassType`, `Overpassother`, `UnderpassType`, `underpassother`, `MitigationMeasures`, `MitigationOther`, `Designpurpose`, `Designother`, `ConnectivityValue`, `connectivityother`, `targetspecies`, `targetother`, `StructureNum`, `length`, `width`, `height`, `substrateone`, `suboneother`, `secondlength`, `secondwidth`, `secondheight`, `substratetwo`, `subtwoother`, `thirdlength`, `thirdwidth`, `thirdheight`, `substratethree`, `subthreeother`, `structurelife`, `trafficvolume`, `camera`, `structurecomplete`, `structurecost`, `Funding`, `RoadName`, `Milemarker`, `protected`, `Video`, `research`, `notes`, `CreationDate`, `Creator`, `EditDate`, `Editor`, `StructureLat`, `StructureLong`, `MeasurementInst`, `MeasurementInstU`, `MeasurementInstO`, `MeasurementInstA`, `DWR_Region`, `pic1`, `pic2`, `pic3`, `pic4`, `pic5`, `pic6`, `pic7`, `pic8`, `pic9`, `pic10`. Picture fields were not downloaded.

A January 2024 Utah DWR news release quotes staff saying at least 119 structures had been completed, and that passage also discusses fences and fish structures. That sentence is a press quote. It is not the `count` field above. This service's observed `CrossingType` values are `overpass` and `underpass`, and its data-last-edit timestamp is 2022-09-12.

Older service, still anonymous:

`https://services.arcgis.com/ZzrwjTRez6FJiOq4/arcgis/rest/services/DWR_Wildlife_Crossings/FeatureServer/0`

- HTTP 200, `application/json`, 2026-11-03T00:37:16Z. Description on the service root: "Locations of UDWR wildlife crossings throughout the state of Utah". Copyright text and item `accessInformation`: `Cramer-USU, sch`. Item `licenseInfo` is null. Redistribution: `license_unresolved`.
- `count`: `{"count":47}`.
- `editingInfo.dataLastEditDate`: 2016-04-13T17:58:39Z. Item spatial reference text: `NAD_1983_UTM_Zone_12N`.
- Fields: `OBJECTID`, `Crossing_Name`, `Road`, `MP`, `UTM_X`, `UTM_Y`, `Type`, `Height`, `Width`, `Length`, `Target_Species`, `Date_Installed`, `Photo_URL`.
- Distinct `Type`: `Arch Bridge`, `Arch Culvert`, `Bridge`, `Bridge overpass`, `Concrete box culvert`, `Concrete box?`, `Concrete culvert`, `Concrete culvert 3 sections`, `Corrugated steel culvert`, `Corrugated steel squash pipe`, `Single Span Bridge`.

These did not open:

- `Wildlife_Crossing_public/FeatureServer/0?f=pjson` at 2026-11-03T00:37:15Z: HTTP 200, `application/json`, body `code` 499, message `Token Required`, `messageCode` `GWM_0003`.
- `Fence_Survey_PublicView/FeatureServer/0?f=pjson` at the same minute: HTTP 200, `application/json`, body `code` 400, message `Invalid URL`.
- The crossing page `https://wildlifemigration.utah.gov/land-animals/crossing/` (HTTP 200 HTML, 2026-11-03T00:40:34Z) still embeds experience `5d80fdda0d7d47a5b94071377d43a9aa`. That experience was not re-downloaded. The September 2026 sweep already tied it to the token-required service.
- `https://maps.udot.utah.gov/arcgis/rest/services/uplan/Wildlife_Mitigation_Projects_PUBLIC/MapServer?f=pjson` at 2026-11-03T00:41:15Z: HTTP 404, `text/html; charset=iso-8859-1`. Item `6e9bb221d34443a792e8dd584e19162d` still lists that URL. Its description says the data are static, have no update cycle, and are current as of 2012. Item modified 2026-10-14T17:21:19Z. Same UDOT disclaimer. No features were available to count.

## Utah Wildlife Migration Initiative corridors

The corridors page `https://wildlifemigration.utah.gov/land-animals/corridors/` returned HTTP 200 HTML at 2026-11-03T00:40:33Z and links Hub dataset `61ca21abff034bb6b209031f694a6fbf`. That item's service is:

`https://services.arcgis.com/ZzrwjTRez6FJiOq4/arcgis/rest/services/migration_corridors/FeatureServer/0`

- HTTP 200, `application/json`, 2026-11-03T00:41:57Z.
- Layer name `corridors_2182022`, polygon. Service description: `migration corridors`.
- `count`: `{"count":82}`.
- `dataLastEditDate`: 2023-04-11T20:01:24Z. Hub item modified 2024-05-29T15:52:06Z.
- Distinct `Species`: `Mule Deer` only.
- Fields: `FID`, `migratType`, `area_name`, `HerdSize`, `CollarGPS`, `acres`, `km2`, `Species`, `Shape__Area`, `Shape__Length`, `DateRange`.
- Spatial reference is a WKT named `GCS_unknown` with datum `D_WGS_1984` and degree units. It is not an EPSG code in the payload.
- Hub `licenseInfo`: "The information contained in these data is dynamic and may change over time. It is the responsibility of the data user to use the data appropriately and consistently within the limitations of geospatial data in general and these data in particular. UDWR provides no warranty nor accepts any responsibility or liability for any inaccurate or incomplete data associated with this file." `accessInformation`: "Utah Division of Wildlife Resources". That is a warranty disclaimer. Redistribution: `license_unresolved`.

Related public services on the same org, found through Hub search `migration` (`numberMatched` 3) and ArcGIS org search:

| Service | Layer | Geometry | `count` | Notes |
| --- | --- | --- | --- | --- |
| `Stopovers_2021/FeatureServer/0` | `stopovers_2182022` | polygon | 30 | Same disclaimer. `dataLastEditDate` 2023-03-27T18:50:37Z. Fields match the corridor layer except there is no `DateRange`. WKT is an Albers projection (parallels 29.5 and 45.5, origin latitude 23, central meridian -96, meters) with datum name `D_unknown` and spheroid GRS80. |
| `Unmapped_Mule_Deer_Migration_Corridors/FeatureServer/0` | same name | polygon, wkid 4326 | 2 | `dataLastEditDate` 2022-12-08T16:33:57Z. Fields: `OBJECTID_1`, `SPECIES`, `km2`, `acres`, `date_updated`, `Shape__Area`, `Shape__Length`. Hub `licenseInfo` empty. `license_unresolved`. |
| `Uinta_Migration_Corridors/FeatureServer` | four layers | polygon | 1, 1, 3, 3 | Layer 0 `SouthSlope_Stopovers` count 1; layer 1 `NorthSlope_Stopovers` count 1; layer 2 `SouthSlope` count 3; layer 3 `NorthSlope` count 3. Counts at 2026-11-03T00:45:07Z. Field names were read for layer 0 only: `FID`, `area_name`, `migratType`, `HerdSize`, `CollarGPS`, `acres`, `km2`, `Species`, `Shape__Area`, `Shape__Length`. Layer 0 `dataLastEditDate` 2021-10-07T21:43:39Z. Albers WKT, datum name `D_North_American_1983`. |

No corridor polygons were downloaded.

UDOT also hosts `Migration_Corridors/FeatureServer/0` (org `pA2nEVnB6tquxgOW`). HTTP 200. Polygon, latestWkid 3857. `count` 9. `dataLastEditDate` 2026-03-30T19:11:59Z. Item snippet: "U.S. 191 South of Duschene StoryMap wildlife migration corridors data." Same UDOT disclaimer. Fields: `FID`, `FID_1`, `migratType`, `area_name`, `HerdSize`, `CollarGPS`, `acres`, `km2`, `Species`, `SHAPE_Leng`, `SHAPE_Area`, `Shape__Area`, `Shape__Length`. This is a nine-feature UDOT copy, not the 82-feature UWMI service.

## NDOT crossings and I-80 Pequop

Two hosts returned the same layer names and the same layer-2 count.

Department server, HTTP 200, `text/plain; charset=UTF-8`, `Date: Tue, 03 Nov 2026 00:37:16 GMT`:

`https://gis.dot.nv.gov/agsphs/rest/services/GeoHub/Wildlife_Crossing_Half_Mile_Buffer_Zone/FeatureServer?f=pjson`

GeoHub copy, HTTP 200, `application/json`, request 2026-11-03T00:41:59Z:

`https://services1.arcgis.com/9Y4hSlLf13E9S0Eo/arcgis/rest/services/Wildlife_Crossing_Half_Mile_Buffer_Zone/FeatureServer?f=pjson`

| Layer id | Name | Geometry | On-prem `count` | Hosted `count` |
| --- | --- | --- | --- | --- |
| 2 | Crossings | point | 33 | 33 |
| 3 | Crossings Buffer Half Mile | polygon | 33 | not counted |

On-prem spatial reference `wkid` 26911. `maxRecordCount` 2000. Layer 2 fields: `GlobalID`, `created_user`, `created_date`, `last_edited_user`, `last_edited_date`, `Structure`, `StructureType`, `Status`, `FocalSpecies`, `Year`, `SiteName`, `System`, `RoadName`, `OBJECTID`. Layer 3 adds `BUFF_DIST`, `ORIG_FID`, `Shape__Area`, `Shape__Length`. Hosted layer fields were not re-read. Hosted service description: "Inventory of NDOT Wildlife Crossings with 1/2 mile buffers showing the extent of areas where hunting is not permissible."

Pequop query, on-prem layer 2, `returnGeometry=false`, 2026-11-03T00:43:28Z. `where=UPPER(SiteName) LIKE '%PEQUOP%'`. `count` for that filter was 6. Six `OBJECTID`s came back:

| OBJECTID | SiteName | RoadName | Structure | StructureType | Year | Status |
| --- | --- | --- | --- | --- | --- | --- |
| 10 | Pequop Summit | IR80 | Overpass | Steel Arches | 2017 | Under Construction |
| 11 | Pequop Summit | IR80 | Underpass | Med Concrete Box | Unknown | Complete |
| 12 | Pequop Summit | IR80 | Underpass | Multi Use Bridge | Unknown | Complete |
| 13 | Pequop Summit | IR80 | Underpass | Med Concrete Box | Unknown | Complete |
| 14 | Pequop Summit | IR80 | Overpass | Steel Arches | 2017 | Under Construction |
| 15 | Pequop Summit | IR80 | Underpass | Multi Use Bridge | Unknown | Complete |

`System` was `IR - Interstate` and `FocalSpecies` was `Lrg Ungulates` on all six in the earlier attribute query. OBJECTIDs 10 and 14, 11 and 13, and 12 and 15 share those attribute values. Geometry was not requested, so this check does not show whether the pairs are different locations. The attributes still say the overpasses are `Under Construction` with `Year` `2017`. That is the service value. It is not rewritten from the NDOT news page about 2018 completion.

ArcGIS item `a5309315ec15468587148d24ff4d0bd0` (the id on the department service) returned HTTP 200 with `Item does not exist or is inaccessible` (`CONT_0001`). Hosted item `4f28b5bf83c342aa8b752327b8fe0e56` is public, owner `OpenData_NDOT`, and `licenseInfo` is empty. The GeoHub learn page (`https://geohub-ndot.hub.arcgis.com/pages/learn`, HTTP 200, `text/html; charset=utf-8`, request 2026-11-03T00:42:00Z) contains the sentence "Anyone can use open data from the NDOT GeoHub at no cost." The fetched HTML did not include a license name. Redistribution: `license_unresolved`.

## Parser

`internal/ingest/udot.go` reads an ArcGIS feature set for the crash MapServer. Mapped attributes are `OBJECTID`, `CRASH_ID`, `CRASH_DATETIME`, `COUNTY_NAME`, `REGION_NAME`, `MAIN_ROAD_NAME`, `ROUTE`, `START_ACCUM`, `ROUTE_DIRECTION`, `LAT_UTM_Y`, `LONG_UTM_X`, `CRASH_SEVERITY_DESC`, `FIRST_HARMFUL_EVENT_DESC`, `WILD_ANIMAL_RELATED`, `DOMESTIC_ANIMAL_RELATED`, and `CURRENT_AS_OF_DATE`, plus point geometry `x` and `y` when present. Other live fields are left unmapped. `WildAnimal()` is true only when `WILD_ANIMAL_RELATED` is `Y`.

`internal/ingest/udot_test.go` uses invented rows and labels them synthetic. No crash, carcass, crossing, or corridor feature was copied into the test.

```text
go test -count=1 -run TestUDOT ./internal/ingest/
ok corridor/internal/ingest
```
