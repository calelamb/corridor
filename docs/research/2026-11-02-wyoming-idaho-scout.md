# Wyoming and Idaho scout — 2026-11-03 UTC

Curl checks for four claims: a WYDOT roadside carcass service, WWRI problem spots, a Wyoming fence inventory, and IDFG roadkill records that include US-20 Island Park and US-30 Montpelier. Responses below are metadata, `returnCountOnly`, `returnDistinctValues` with `returnGeometry=false`, or `outStatistics`. No feature geometries were written into this note or into Git. Counts appear only when a query returned them, and each count cites that query. This pass did not edit `data/SOURCES.md`.

## What changed for S30

`data/SOURCES.md` S30 currently says Wyoming/Nevada carcasses were unverified and that no suitable licensed raw dataset was verified in that pass. This request checked Wyoming and Idaho only. Nevada was not requested and was not checked.

A public WYDOT Roadkill Reports layer answered anonymously with a schema, an empty `copyrightText`, and a feature count. The WYDOT website disclaimer fetched in the same pass does not grant reuse of that layer. S30’s “no licensed raw dataset” statement still matches this request. The “unverified” statement does not match the Wyoming roadside service: the endpoint, field names, and count below were verified on 2026-11-03 UTC. No raw WYDOT file was acquired, so there is still no retrieval SHA-256 for a carcass artifact.

## WYDOT roadside carcass

Layer metadata: `https://gisservices.wyoroad.info/arcgis/rest/services/Roadkill/Roadkill_Reports/FeatureServer/0?f=pjson`

| | |
| --- | --- |
| Response Date | 2026-11-03T00:38:02Z |
| HTTP | 200 |
| Content-Type | `text/plain;charset=UTF-8` |
| Geometry | `esriGeometryPoint`; extent spatial reference wkid 102100, latestWkid 3857 |
| MaxRecordCount | 2000 |
| Service item id | `c229c363360a45d8b3fa81d2ce3ea9a7` |
| `copyrightText` | empty |
| License | `license_unresolved` |

Service description text: “The dataset contains records of reported animal roadkill that have happened in Wyoming over the years 2022 to present, collected via the 511 app. The dataset contains information about the reports such as the time and location, as well as the species that was reported.”

Field names: `pk`, `received_ts`, `status`, `location`, `route`, `highway`, `milepost`, `cardinal_direction`, `county`, `district`, `longitude`, `latitude`, `species`, `other_species`, `prohibited_species`, `collected`, `road_blockage`, `in_excluded_zone`.

Count query: `https://gisservices.wyoroad.info/arcgis/rest/services/Roadkill/Roadkill_Reports/FeatureServer/0/query?where=1%3D1&returnCountOnly=true&f=json`

| | |
| --- | --- |
| Response Date | 2026-11-03T00:38:41Z |
| HTTP | 200 |
| Content-Type | `application/json;charset=UTF-8` |
| Body | `{"count": 8816}` |

`received_ts` statistics, `returnGeometry=false`, `where=1=1`, min and max on `received_ts`: response Date 2026-11-03T00:40:06Z, HTTP 200, Content-Type `application/json;charset=UTF-8`. The service returned `min_ts` 1643894752041 and `max_ts` 1793642577151, which are 2022-02-03T13:25:52Z and 2026-11-02T18:02:57Z.

Distinct `species`, `returnGeometry=false`, `where=1=1`: response Date 2026-11-03T00:38:41Z, HTTP 200, Content-Type `application/json;charset=UTF-8`. Values returned: `ANTELOPE`, `BIGHORN SHEEP`, `BLACK BEAR`, `COYOTE`, `DEER`, `ELK`, `GRIZZLY BEAR`, `MOOSE`, `MOUNTAIN GOAT`, `MOUNTAIN LION`, `WILD BISON`, `WILD TURKEY`, `WOLF`.

The same host’s service directory `https://gisservices.wyoroad.info/arcgis/rest/services?f=pjson` returned HTTP 200, Content-Type `text/plain;charset=UTF-8`, Date 2026-11-03T00:38:01Z. Folder names included `Roadkill` and `Snowfence`. The Roadkill folder lists `Roadkill/Roadkill_Reports` as FeatureServer and MapServer (directory response Date 2026-11-03T00:38:43Z, HTTP 200, Content-Type `text/plain;charset=UTF-8`).

`https://www.dot.state.wy.us/home/administration/disclaimer.html` returned HTTP 200, Content-Type `text/html;charset=UTF-8`, Date 2026-11-03T00:42:53Z. The page text says the State of Wyoming does not endorse referenced or linked sites and is not responsible for those sites’ content. That sentence is about links. It does not state a reuse license for Roadkill Reports.

`internal/ingest/wydot.go` parses this layer’s attribute names. `WYDOTRoadkillArtifact` records `license_unresolved`, the metadata response time above, and an empty SHA-256 because no feature download was stored.

## WWRI problem spots

ArcGIS item search `https://www.arcgis.com/sharing/rest/search?f=pjson&num=10&q=wyoming%20wildlife%20roadways%20problem%20spots` returned HTTP 200, Content-Type `text/plain;charset=utf-8`, one result. The batch that issued this call started at 2026-11-03T00:38:36Z; the response Date header was absent.

Item `f65dfd53c25c4b6488b74fc45a7f8c5f`, title “Roadway Initiative Problem Spots”, type Web Map, owner `KNordyke@wgfd`, access public. Snippet: “Map showing problem spots for big game collisions on Wyoming highways as identified in 2017 Summit”. `license` and `licenseInfo` are null. Item `modified` 1774380955000 is 2026-03-24T19:35:55Z. License: `license_unresolved`.

The web map data names the operational layer “Problem Spots” at `https://services6.arcgis.com/cWzdqIyxbijuhPLw/arcgis/rest/services/WYDOT_all_data/FeatureServer/0`.

Service root `https://services6.arcgis.com/cWzdqIyxbijuhPLw/arcgis/rest/services/WYDOT_all_data/FeatureServer?f=pjson`: response Date 2026-11-03T00:40:01Z, HTTP 200, Content-Type `application/json; charset=utf-8`. `copyrightText` empty. Description: “Joint project between WGFD and WYDOT to identify collision problem spots and possible solutions to abate collision frequency.” Layers: 0 `Statewide Scope Problems Spots` (polyline), 1 `Overlaps` (polyline), 2 `WYDOT Highways` (polyline), 3 `WYDOT Districts` (polygon).

Layer 0 metadata, same timestamp and content type: name `Statewide Scope Problems Spots`, geometry `esriGeometryPolyline`, `copyrightText` empty, maxRecordCount 1000. Extent spatial reference WKT is named `wylam` (NAD83 Lambert conformal conic, meters). License: `license_unresolved`.

Field names: `OBJECTID`, `Species`, `Data`, `Road_Jurisdiction`, `Road_No`, `MP_start`, `MP_end`, `Season`, `Herd`, `Description`, `Issues`, `Solutions`, `Contacts`, `Comments`, `Scope_Reason`, `Priority`, `Priority_Reason`, `Scope`, `ML_Route`, `Shape__Length`.

Count query: `https://services6.arcgis.com/cWzdqIyxbijuhPLw/arcgis/rest/services/WYDOT_all_data/FeatureServer/0/query?where=1%3D1&returnCountOnly=true&f=json`

| | |
| --- | --- |
| Response Date | 2026-11-03T00:40:02Z |
| HTTP | 200 |
| Content-Type | `application/json; charset=utf-8` |
| Body | `{"count": 206}` |

Distinct `Species`, `returnGeometry=false`: response Date 2026-11-03T00:41:04Z, HTTP 200, Content-Type `application/json; charset=utf-8`. Values: `Antelope`, `Bighorn Sheep`, `Elk`, `Moose`, `Mule Deer`.

`https://www.arcgis.com/apps/MapSeries/index.html?appid=ef666ba292b74c56a339efc10fca5332` returned HTTP 200, Content-Type `text/html`, 647 bytes, at request time 2026-11-03T00:47:22Z. The Date header was absent. The static document title is “Item Replacement” and the body is a client-side `<arcgis-portal-replacement-item-redirect>` element. That HTML response does not include problem-spot fields. The feature layer cited above is the layer named in the web map JSON.

## Wyoming fence inventory

Owner search `https://www.arcgis.com/sharing/rest/search?f=pjson&num=15&q=owner%3AKNordyke%40wgfd%20(fence%20OR%20fencing%20OR%20%22problem%20spots%22)` returned HTTP 200, Content-Type `text/plain;charset=utf-8`, Date 2026-11-03T00:40:07Z, total 5. One result is the Feature Service “Fences”, URL `https://services6.arcgis.com/cWzdqIyxbijuhPLw/arcgis/rest/services/Fences/FeatureServer`, `licenseInfo` null.

Service root `https://services6.arcgis.com/cWzdqIyxbijuhPLw/arcgis/rest/services/Fences/FeatureServer?f=pjson`: response Date 2026-11-03T00:41:03Z, HTTP 200, Content-Type `application/json; charset=utf-8`. `copyrightText` empty. Description: “This data reflects existing fences on the landscape in western and central Wyoming.” One layer, id 0, name `Fences`, geometry polyline. License: `license_unresolved`.

Layer 0 metadata, same timestamp and content type: maxRecordCount 1000, extent wkid 26912 (latestWkid 26912). `copyrightText` empty.

Field names: `OBJECTID_1`, `OBJECTID`, `TYPE`, `NUM_STR`, `FEN_HEIGHT`, `POST_DIST`, `STR1_TYPE`, `STR1_SPACE`, `STR2_TYPE`, `STR2_SPACE`, `STR3_TYPE`, `STR3_SPACE`, `STR4_TYPE`, `STR4_SPACE`, `STR5_TYPE`, `STR5_SPACE`, `STR6_TYPE`, `STR6_SPACE`, `FEN_CON`, `COMMENT`, `DATE_`, `SOURCE`, `OBSERVER`, `Max_PDOP`, `Max_HDOP`, `Corr_Type`, `Rcvr_Type`, `GPS_Date`, `GPS_Time`, `Update_Sta`, `Feat_Name`, `Datafile`, `Unfilt_Pos`, `Filt_Pos`, `Data_Dicti`, `GPS_Week`, `GPS_Second`, `GPS_Length`, `GPS_3DLeng`, `Avg_Vert_P`, `Avg_Horz_P`, `Worst_Vert`, `Worst_Horz`, `Line_ID`, `Length_mi`, `Shape_Leng`, `FNODE_`, `TNODE_`, `LPOLY_`, `RPOLY_`, `DDFENCE_`, `DDFENCE_ID`, `FENCECODE`, `AREA`, `OWNER`, `TOWNSHIP__`, `RANGE__WES`, `SECTION`, `SUBSECTION`, `AUTHORIZED`, `FOUND`, `MAP`, `YEAR_BUILT`, `LENGTH`, `DATA_SOURC`, `RECOMMENDA`, `LENGTH_FEE`, `NET`, `SegmentID`, `PostType`, `PostDist`, `NumStrings`, `StringType`, `StringDist`, `StringTy00`, `StringDi00`, `StringTy01`, `StringDi01`, `StringTy02`, `FenConditi`, `Operationa`, `Owner1`, `Owner1Dir`, `Owner2`, `Owner2Dir`, `Comments`, `Zone`, `Segid`, `Groupid`, `Segidrev`, `Joinid`, `Lenmiles`, `Lenfeet`, `Status`, `Year`, `Notes`, `Hold`, `Name`, `Shape__Length`.

The field list has fence structure, ownership, and GPS-quality names. It has no species field.

Count query: `https://services6.arcgis.com/cWzdqIyxbijuhPLw/arcgis/rest/services/Fences/FeatureServer/0/query?where=1%3D1&returnCountOnly=true&f=json`

| | |
| --- | --- |
| Response Date | 2026-11-03T00:41:04Z |
| HTTP | 200 |
| Content-Type | `application/json; charset=utf-8` |
| Body | `{"count": 14789}` |

### WYDOT snow fences, separate inventory

`https://gisservices.wyoroad.info/arcgis/rest/services/Snowfence/Snowfence_External_View/MapServer/1?f=pjson`

| | |
| --- | --- |
| Response Date | 2026-11-03T00:40:06Z |
| HTTP | 200 |
| Content-Type | `text/plain;charset=UTF-8` |
| Geometry | `esriGeometryPolyline` |
| `copyrightText` | `WYDOT & Trihydro Corporation.` |
| License | `license_unresolved` (copyright line present; no reuse grant in the layer JSON) |

Description begins: “This data set depicts snow fences in Wyoming. It contains the locations of all snow fences inventoried since project inception in May 2005.” Extent spatial reference WKT is named `WYDOT Albers`, unit `Foot_US`.

Field names: `objectid`, `globalid`, `fenceid`, `fenceno`, `lrs_name`, `highway_name`, `milepost`, `district`, `offsetdist`, `offsetdir`, `fencetype`, `anchortype`, `silllength`, `bottomgap`, `panelwidth`, `height`, `length_ft`, `vertical_post_shape_size`, `vertical_post_spacing`, `totalpanel`, `date_constructed`, `most_recent_maint_date`, `comments`, `signed`, `invdate`, `created_user`, `created_date`, `last_edited_user`, `last_edited_date`, `shape`, `st_length(shape)`.

Count query: `https://gisservices.wyoroad.info/arcgis/rest/services/Snowfence/Snowfence_External_View/MapServer/1/query?where=1%3D1&returnCountOnly=true&f=json` — response Date 2026-11-03T00:41:11Z, HTTP 200, Content-Type `application/json;charset=UTF-8`, body `{"count": 4756}`.

### Public service named Wildlife_Fence that is not this inventory

`https://services.arcgis.com/Cif2vyIafUMtRuVN/ArcGIS/rest/services/Wildlife_Fence/FeatureServer?f=pjson` returned HTTP 200, Content-Type `application/json; charset=utf-8`, during the curl batch that started 2026-11-03T00:38:36Z. The Date header was absent. Service description: “Features extracted from Lidar - Wildlife Fence”. The layer list contains id 12, name `Wildlife_Fence`. A request for layer 0 returned HTTP 200, Content-Type `application/json; charset=utf-8`, with an ArcGIS error that layer 0 was not found (same batch).

Layer 12 metadata: `https://services.arcgis.com/Cif2vyIafUMtRuVN/ArcGIS/rest/services/Wildlife_Fence/FeatureServer/12?f=pjson` — response Date 2026-11-03T00:40:02Z, HTTP 200, Content-Type `application/json; charset=utf-8`. Geometry polyline. Extent spatial reference wkid 2036, latestWkid 2953. `copyrightText` empty. Fields include `Chainage_start_Location`, `Chainage_end_Location`, `OMR_Code`, `Route`, `X_Coordinate`, and `Y_Coordinate`. This is a different service from the WGFD Fences layer above.

## IDFG roadkill

Layer metadata: `https://gisportal-idfg.idaho.gov/hosting/rest/services/Roadkill/Roadkill_Observations/FeatureServer/1?f=pjson`

| | |
| --- | --- |
| Response Date | 2026-11-03T00:38:02Z |
| HTTP | 200 |
| Content-Type | `text/plain; charset=UTF-8` |
| Geometry | `esriGeometryPoint`; extent wkid 102100, latestWkid 3857 |
| MaxRecordCount | 100000 |
| Service item id | `d51848101f994df9b6b794738830d62c` |
| `copyrightText` | empty |
| `typeIdField` | `species`, with 12 renderer classes listed below |

Field names: `OBJECTID`, `species`, `source`, `observed`, `salvaged`, `sex`, `lifeStage`, `lifeState`, `disposition`, `decomposition`, `latitude`, `longitude`, `highway`, `milepost`, `note`, `reported`, `county`, `region`, `gmu`, `path`, `GlobalID`.

Renderer class labels on that response: Mule Deer (`Odocoileus hemionus`), White-tailed Deer (`Odocoileus virginianus`), Elk (`Cervus canadensis`), Moose (`Alces americanus`), Deer (`Odocoileus`), Domestic Cat (`Felis catus`), Common Raccoon (`Procyon lotor`), Coyote (`Canis latrans`), Skunks and Stink Badgers (Mephitidae), North American Porcupine (`Erethizon dorsatum`), Pronghorn (`Antilocapra americana`), American Badger (`Taxidea taxus`).

FeatureServer root `https://gisportal-idfg.idaho.gov/hosting/rest/services/Roadkill/Roadkill_Observations/FeatureServer?f=pjson`: response Date 2026-11-03T00:38:41Z, HTTP 200, Content-Type `text/plain; charset=UTF-8`. `copyrightText` empty. The description says the page allows downloads of raw roadkill reported by the public, Idaho Fish and Game, or the Idaho Transportation Department; that the same roadkill can be reported more than once; that users assume responsibility for usability; and that Idaho Fish and Game does not assume liability and makes no warranty by distributing the data. The same description says the dataset displays roadkill observations reported in Idaho from 1977 to the current date, and it recommends downloading the full CSV because of open-data filter issues.

Portal item `https://gisportal-idfg.idaho.gov/portal/sharing/rest/content/items/d51848101f994df9b6b794738830d62c?f=pjson`: response Date 2026-11-03T00:38:02Z, HTTP 200, Content-Type `text/plain; charset=utf-8`. `access` is `public`. The `license` field is null. `licenseInfo` is a liability disclaimer: the data are a best representation only, users release Idaho Fish and Game and related State of Idaho entities from liability, and distribution is not a warranty. License for reuse: `license_unresolved`.

`https://idfg.idaho.gov/terms` returned HTTP 503, Content-Type `text/html`, Date 2026-11-03T00:38:44Z, and again HTTP 503, Content-Type `text/html`, Date 2026-11-03T00:40:06Z. `https://idfg.idaho.gov/species/roadkill` returned HTTP 503, Content-Type `text/html`, Date 2026-11-03T00:40:07Z. The general terms page was not readable on this request.

Statewide count: `https://gisportal-idfg.idaho.gov/hosting/rest/services/Roadkill/Roadkill_Observations/FeatureServer/1/query?where=1%3D1&returnCountOnly=true&f=json`

| | |
| --- | --- |
| Response Date | 2026-11-03T00:38:41Z |
| HTTP | 200 |
| Content-Type | `application/json; charset=UTF-8` |
| Body | `{"count": 70651}` |

`observed` min/max statistics, `where=1=1`, `returnGeometry=false`: response Date 2026-11-03T00:40:05Z, HTTP 200, Content-Type `application/json; charset=UTF-8`. Returned epochs `-202694400000` and `1798502400000`, which are 1963-07-31T00:00:00Z and 2026-12-29T00:00:00Z. The maximum is later than this request. The service description’s “1977 to the current date” sentence and this minimum are both from this request; they disagree.

### US-20, US-30, Island Park, Montpelier

These are attribute filters. They do not locate a town by coordinates.

Highway strings from `returnDistinctValues` on `highway`, `returnGeometry=false`, `where=UPPER(highway) LIKE '%20%' OR UPPER(highway) LIKE '%30%'`: response Date 2026-11-03T00:40:04Z, HTTP 200, Content-Type `application/json; charset=UTF-8`. Values: `Higgway 20`, `Highway 20`, `Highway 30`, `Hwy 20`, `Hwy 20/26`, `Hwy 20/26/`, `Hwy 2026/9`, `Hwy 30`, `ID-200`, `SH-200`, `ST Hwy 20`, `US 20`, `Us hwy 20`, `US-20`, `US-20 Corr`, `US-30`, `Us20`. `ID-200` and `SH-200` matched because the filter asked for the digits 20.

| Query where-clause | Response Date | HTTP | Content-Type | Count |
| --- | --- | --- | --- | --- |
| `UPPER(highway) LIKE '%US-20%' OR UPPER(highway) LIKE '%US 20%'` | 2026-11-03T00:40:03Z | 200 | `application/json; charset=UTF-8` | 2415 |
| `UPPER(highway) IN ('US-20','US 20','US HWY 20','HWY 20','HIGHWAY 20','HIGGWAY 20','ST HWY 20','US20','US-20 CORR')` | 2026-11-03T00:42:10Z | 200 | `application/json; charset=UTF-8` | 2448 |
| `UPPER(highway) LIKE '%US-30%' OR UPPER(highway) LIKE '%US 30%'` | 2026-11-03T00:40:03Z | 200 | `application/json; charset=UTF-8` | 2828 |
| `UPPER(highway) IN ('US-30','US 30','HWY 30','HIGHWAY 30')` | 2026-11-03T00:42:11Z | 200 | `application/json; charset=UTF-8` | 2833 |
| `UPPER(note) LIKE '%ISLAND PARK%'` | 2026-11-03T00:41:05Z | 200 | `application/json; charset=UTF-8` | 25 |
| `UPPER(note) LIKE '%ISLAND PARK%' AND UPPER(highway)='US-20'` | 2026-11-03T00:42:53Z | 200 | `application/json; charset=UTF-8` | 4 |
| `UPPER(county) LIKE '%FREMONT%' AND (UPPER(highway) LIKE '%US-20%' OR UPPER(highway) LIKE '%US 20%' OR UPPER(highway) LIKE '%HWY 20%' OR UPPER(highway) LIKE '%HIGHWAY 20%')` | 2026-11-03T00:41:05Z | 200 | `application/json; charset=UTF-8` | 0 |
| `UPPER(note) LIKE '%MONTPELIER%'` | 2026-11-03T00:41:05Z | 200 | `application/json; charset=UTF-8` | 15 |
| `UPPER(note) LIKE '%MONTPELIER%' AND UPPER(highway)='US-30'` | 2026-11-03T00:42:54Z | 200 | `application/json; charset=UTF-8` | 1 |
| `UPPER(county) LIKE '%BEAR LAKE%' AND (UPPER(highway) LIKE '%US-30%' OR UPPER(highway) LIKE '%US 30%' OR UPPER(highway) LIKE '%HWY 30%' OR UPPER(highway) LIKE '%HIGHWAY 30%')` | 2026-11-03T00:41:06Z | 200 | `application/json; charset=UTF-8` | 0 |

Each count used `returnCountOnly=true&f=json` on `https://gisportal-idfg.idaho.gov/hosting/rest/services/Roadkill/Roadkill_Observations/FeatureServer/1/query`.

Distinct `highway`, `county`, and `region` for `UPPER(note) LIKE '%ISLAND PARK%'`, `returnGeometry=false`: response Date 2026-11-03T00:42:08Z, HTTP 200, Content-Type `application/json; charset=UTF-8`. Combinations returned: highway `US-20` with null county and null region, and highway empty with null county and null region. The same query with `MONTPELIER` returned highway `US-30` and highway empty, again with null county and null region.

Distinct `species` where the uppercased value matches wolf, grizz, lion, lynx, wolverine, or black bear, `returnGeometry=false`: response Date 2026-11-03T00:40:05Z, HTTP 200, Content-Type `application/json; charset=UTF-8`. Values: `American Black Bear (Ursus americanus)`, `Black Bear (Ursus americanus americanus)`, `Bobcat (Lynx rufus)`, `Gray Wolf (Canis lupus)`, `Grizzly Bear (Ursus arctos)`, `Mountain Lion, Cougar, or Puma (Puma concolor)`, `Wolverine (Gulo gulo)`. The filter also matched `LYNX`; the only returned label containing that string is the bobcat label. These names are outside the 12 renderer classes.

## Privacy

The WYDOT Roadkill Reports schema includes `latitude` and `longitude`. The distinct species query returned `GRIZZLY BEAR`, `WOLF`, `MOUNTAIN LION`, `BLACK BEAR`, and `BIGHORN SHEEP`. Public release of those precise coordinates would need H3 resolution 6 and a 30-day delay. `received_ts` is the timestamp on that layer.

The IDFG schema includes `latitude` and `longitude`. The distinct species query returned grizzly bear, gray wolf, mountain lion, black bear, bobcat, and wolverine. Public release of those precise coordinates would need H3 resolution 6 and a 30-day delay. `observed` is the timestamp on that layer.

The WWRI problem-spot distinct species list returned `Bighorn Sheep` and did not return grizzly bear, wolf, or mountain lion. The features are polylines. ADR 0008 treats bighorn sheep as sensitive, so public release of those precise segments would need H3 resolution 6. The layer fields have `Season` and no observation timestamp, so a 30-day delay cannot be computed from the published fields.

The WGFD Fences schema and the WYDOT snow-fence schema do not include a species field.

## Parser check

`go test ./internal/ingest/ -count=1 -run WYDOT` passed. The test JSON uses invented coordinates and the field names listed for Roadkill Reports. No live feature body was saved in the repository.
