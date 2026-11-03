# Canada / Banff endpoint scout

**Observed:** 2026-11-03 UTC, by `curl` from this workstation. Nothing was committed and no file was written under `data/`. Byte sizes and row counts below are counts of responses this session actually received. They are not checksums, and they are not a claim that a publisher’s full historical archive was mirrored.

Banff National Park is a Parks Canada unit in Alberta. It is not a unit of the U.S. National Park Service. Glacier National Park of Canada, which appears in one national summary below, is likewise a Parks Canada unit and is not the U.S. Glacier National Park. No response cited 23 U.S.C. The reuse text that did load is Canadian open-government licence text.

Statuses used here:

- `anonymous_csv` — an unauthenticated GET or byte-range returned CSV.
- `catalog_metadata` — a catalog or article page returned; the data file was not the body.
- `license_unresolved` — a service answered, and the body contained no licence grant.
- `download_refused` — the landing page loaded, and the file URL did not.
- `not_found` — HTTP 404.
- `descriptive_page` — program prose, with no dataset file in the body.

## Parks Canada program pages

These pages describe the Trans-Canada Highway (TCH) mitigation program. They do not host a download and they do not say that the long-term monitoring database is public or request-only.

| Requested URL | UTC | Status | Content-Type | What the body says |
| --- | --- | --- | --- | --- |
| `https://parks.canada.ca/pn-np/ab/banff/nature` | 2026-11-03T00:38:33Z | 200 `descriptive_page` | `text/html; charset=utf-8` | Date modified 2025-02-17. Says multi-year Banff datasets are on the Government of Canada Open Data portal and links “Wildlife crossing structures” to `http://open.canada.ca/data/en/dataset/d5fd45b9-7c70-4339-a68b-e60d7cbf42e8` and “Wildlife mortality” to `http://open.canada.ca/data/en/dataset/6aa18934-5fec-4f12-8dd7-8356555d0576`. The ArcGIS phrase on the page links to `http://open.canada.ca/data/en/organization/pc?res_type=api`, which redirected (see below). |
| `https://parks.canada.ca/pn-np/ab/banff/nature/conservation/transport/tch-rtc` | 2026-11-03T00:38:35Z | 200 `descriptive_page` | `text/html; charset=utf-8` | Date modified 2024-04-22. Says the TCH in the park was twinned in phases from 1981, about 83 km, with goals to improve motorist safety, reduce wildlife mortality and fragmentation, and improve movement of goods. |
| `https://www.pc.gc.ca/en/pn-np/ab/banff/nature/conservation/transport/tch-rtc` | 2026-11-03T00:38:35Z | 301 then 200 | first `text/html; charset=UTF-8`, then `text/html; charset=utf-8` | One redirect to the `parks.canada.ca` TCH URL above. Final body is that same page. |
| `https://parks.canada.ca/pn-np/ab/banff/nature/conservation/transport/tch-rtc/passages-crossings` | 2026-11-03T00:38:34Z | 200 `descriptive_page` | `text/html; charset=utf-8` | Date modified 2022-11-19. Says monitoring has run since 1996; states 44 structures (six overpasses and 38 underpasses) and 82 km of fencing; says collisions dropped by more than 80% after fencing and crossings. Names collaboration with the Western Transportation Institute, the Miistakis Institute, and the Woodcock and Wilburforce foundations. No file link, no login, and no sentence that the monitoring database is public or available on request. |

`https://open.canada.ca/data/en/organization/pc?res_type=api&q=banff`, requested 2026-11-03T00:39:43Z, returned 301, then 301, then 200 `text/html; charset=utf-8` at `https://search.open.canada.ca/opendata/?owner_org=pc` (305,290 bytes). The Banff and API filters were not retained on the final URL. A CKAN search the same morning, `package_search?q=banff%20arcgis&fq=organization:pc` at 2026-11-03T00:38:32Z, was HTTP 200 `application/json;charset=utf-8` with `"count": 0`. A second search, `q=Banff%20GIS%20geodatabase` and `fq=organization:pc` at 2026-11-03T00:41:58Z, was also HTTP 200 with `"count": 0`.

## Open Government Licence – Canada

`https://open.canada.ca/en/open-government-licence-canada` at 2026-11-03T00:38:33Z returned HTTP 200, `text/html; charset=utf-8`. The page says it is version 2.0. Grant sentence in the body: “The Information Provider grants you a worldwide, royalty-free, perpetual, non-exclusive licence to use the Information, including for commercial purposes, subject to the terms below.” Default attribution sentence: “Contains information licensed under the Open Government Licence – Canada.” Governing-law sentence: “This licence is governed by the laws of the province of Ontario and the applicable laws of Canada.” Exemptions named on the page include personal information, third-party rights the provider cannot license, official symbols, and patents, trade-marks, and official marks.

## Wildlife Crossing Structures – Banff

Catalog: `https://open.canada.ca/data/en/api/3/action/package_show?id=d5fd45b9-7c70-4339-a68b-e60d7cbf42e8` at 2026-11-03T00:37:49Z, HTTP 200, `application/json;charset=utf-8`, 7,321 bytes, `"success": true`.

The JSON identifies publisher organization Parks Canada | Parcs Canada, `license_id` `ca-ogl-lgo`, `license_title` “Open Government Licence - Canada”, `license_url` the licence page above, `restrictions` `unrestricted`, `jurisdiction` `federal`, `frequency` `P1Y`, `time_period_coverage_start` `2014-10-14 00:00:00`, `time_period_coverage_end` `2017-07-11 00:00:00`, `metadata_modified` `2024-12-09T17:13:10.850885`, maintainer email `derek.petersen@pc.gc.ca`. `spatial` is an empty string. Notes say Parks Canada monitors 44 crossing structures along the TCH with motion-triggered cameras, and that the work also concerns Highway 93 South in Kootenay National Park, to measure mitigation success for road mortality and habitat connectivity. One resource: CSV, catalog `size` 13073, state `active`.

JSON-LD for the same record, `https://open.canada.ca/data/en/dataset/d5fd45b9-7c70-4339-a68b-e60d7cbf42e8.jsonld` at 2026-11-03T00:37:49Z, was HTTP 200, `application/ld+json`, 7,638 bytes.

File: `https://open.canada.ca/data/dataset/d5fd45b9-7c70-4339-a68b-e60d7cbf42e8/resource/bd8c589f-4023-4f37-8e07-d948a8fcb6d0/download/banff_np_wildlife_crossing_structures_2014-2017_data.csv` at 2026-11-03T00:39:38Z. HTTP 302 (`text/html; charset=utf-8`) then HTTP 200, `text/csv`, `Content-Length: 13073`. The redirect target host was `opencanada.blob.core.windows.net` under `opengovprod/resources/bd8c589f-4023-4f37-8e07-d948a8fcb6d0/`. The query string was a short-lived read signature and is omitted here. No account page appeared. Status: `anonymous_csv`.

The bytes are Windows-1252, not UTF-8 (a UTF-8 decode fails on byte `0xE9`). There are 183 physical lines: an English header, a French label row, and 181 data rows. English header:

`Region`, `Location`, `Year images were taken`, `Date of first image`, `Date of last image`, `number of hikers and runners`, `number of horse riders`, `number of bicyclists`, `number of cougars`, `number of elk`, `number of grizzly bears`, `number of lynx`, `number of wolves`, `number of deer`.

`Region` values on data rows: Banff Town, Bow Valley, Lake Louise, Massive, Panorama. `Location` values are 60 structure names, almost all prefixed `TCH` (examples of the name pattern: `TCH East Gate UP`, `TCH Redearth OP-East`, `TCH Wolverine OP_West`). There is no latitude, longitude, or other geometry column. `Year images were taken` values are 2014, 2015, 2016, and 2017. Image dates in the sampled rows are day/month/year. Count cells are sometimes blank and sometimes the digit `0`, so a blank cell was not written as zero in the file. Each data row is a count for one named structure over an image-date span, not one animal passage.

## Wildlife Mortality – Banff

Catalog: `https://open.canada.ca/data/en/api/3/action/package_show?id=6aa18934-5fec-4f12-8dd7-8356555d0576` at 2026-11-03T00:38:31Z, HTTP 200, `application/json;charset=utf-8`, 7,637 bytes. Same licence fields as the crossing record (`ca-ogl-lgo`, unrestricted). `time_period_coverage_start` `2005-01-11 00:00:00`, `time_period_coverage_end` `2017-12-31 00:00:00`, `metadata_modified` `2024-12-09T17:13:13.388831`, `geographic_region` value `48` (code left as returned). Notes say rail- and vehicle-caused mortality and management destruction for Banff, 2005–2017, and that “each species mortality is recorded by date, location, count, age, gender, and cause.” The same notes point recent records at `https://open.canada.ca/data/en/dataset/cc5ea139-c628-46dc-ac55-a5b3351b7fdf`.

That older id, requested as package_show at 2026-11-03T00:38:31Z, was HTTP 404 `application/json;charset=utf-8` with `"message": "Not found"`. The HTML URL at 2026-11-03T00:39:43Z was HTTP 404 `text/html; charset=utf-8`. Status of that id: `not_found`.

Mortality file: `https://open.canada.ca/data/dataset/6aa18934-5fec-4f12-8dd7-8356555d0576/resource/f4d96701-e59f-4d3d-8af0-b3b772fc68e3/download/banff_np_wildlife_mortality_2005-2017_data.csv` at 2026-11-03T00:39:39Z. HTTP 302 then 200, `text/csv`, `Content-Length: 50797`, same blob host pattern, unauthenticated. Status: `anonymous_csv`. Windows-1252. English header, then a French label row, then 893 data rows. Header columns: `Unique ID`, `Date`, `National Park`, `Wildlife Species`, `Total`, `Age / Gender`, `Mortality Cause`. `National Park` on the sampled data rows is `Banff`. There is no location, kilometre, latitude, or longitude column in the retrieved header, despite the catalog sentence that says location is recorded. Dates are day/month/year. `Mortality Cause` tallies in the 893 data rows: Highway 421, Railway 388, Management Destruction 82, HIGHWAY 2. Species labels include elk, deer, coyote, black bear, moose, wolf, grizzly bear, lynx, and cougar. The file mixes highway collisions, railway collisions, and management destruction, and it cannot be placed on the TCH from these columns.

Dictionary file: `https://open.canada.ca/data/dataset/6aa18934-5fec-4f12-8dd7-8356555d0576/resource/190a46d9-f5ec-4d0b-a482-f63a5c78c67f/download/banff_np_wildlife_mortality_2005-2017_data_dictionary.csv` at 2026-11-03T00:39:39Z. HTTP 302 then 200, `text/csv`, `Content-Length: 1973`. It lists controlled values for species, age/gender, and the three cause labels Highway, Management Destruction, and Railway. It does not define a location field.

## Human–wildlife coexistence (national), the catalog’s successor link

The nature page links coexistence incidents to `http://open.canada.ca/data/en/dataset/743a0b4a-9e33-4b12-981a-9f9fd3dd1680`. Package search `q=human-wildlife%20coexistence%20incidents` at 2026-11-03T00:38:32Z, HTTP 200 `application/json;charset=utf-8`, returned that record as “Human-wildlife coexistence incidents managed by Parks Canada”, `license_id` `ca-ogl-lgo`, `license_title` “Open Government Licence - Canada”. Notes say the record documents incidents and responses from 2010 to 2024 and define an incident as a potential conflict assigned to staff. This is a national Parks Canada record, not a Banff-only extract. Several resources are multi-megabyte; this scout read headers or small summaries only and did not store those files in the repo.

Header descriptions, full file, at 2026-11-03T00:39:40Z: `https://open.canada.ca/data/dataset/743a0b4a-9e33-4b12-981a-9f9fd3dd1680/resource/7f1ce4da-f174-4394-a53d-398b30c0f69d/download/pca-national-human-wildlife-coexistence-header-descriptions.csv`. HTTP 302 then 200, `text/csv`, `Content-Length: 5544`. The latitude row says: “Decimal degrees latitude of the approximate incident location within 1500m (root-mean-squared error approx. 1000m).” The longitude row uses the same 1500 m / about 1000 m wording. Column names in this description file are `Latitude Public` and `Longitude Public`.

Incidents file, range only, at 2026-11-03T00:39:41Z: `https://open.canada.ca/data/dataset/743a0b4a-9e33-4b12-981a-9f9fd3dd1680/resource/eaadb900-e1e5-4844-9617-4b8d76f0ea80/download/pca-human-wildlife-coexistence-incidents-detailed-records-2010-2024.csv`. HTTP 302 then 206, `text/csv`, `Content-Range: bytes 0-2500/15551403`. Header actually in the file: `Incident Number`, `Incident Date`, `Field Unit`, `Protected Heritage Area`, `Public Latitude`, `Public Longitude`, `In Park`, `Incident Type`, `Sum of Total Staff Involved`, `Sum of Total Staff Hours`. Those coordinate names do not match the description file’s `Latitude Public` / `Longitude Public`. Species is not in this header. Row count was not taken. Status: `anonymous_csv` for the header only.

Animals file, range only, at 2026-11-03T00:41:54Z: `https://open.canada.ca/data/dataset/743a0b4a-9e33-4b12-981a-9f9fd3dd1680/resource/d6892a77-5e2f-407e-80ee-3b76466a4d80/download/pca-human-wildlife-coexistence-animals-involved-detailed-records-2010-2024.csv`. HTTP 302 then 206, `text/csv`, `Content-Range: bytes 0-1800/20027792`. Header: `Incident Number`, `Incident Date`, `Field Unit`, `Protected Heritage Area`, `Incident Type`, `Species Common Name`, `Sum of Number of Animals`, `Animal Health Status`, `Cause of Animal Health Status`, `Animal Behaviour`, `Reason for Animal Behaviour`, `Animal Attractant`, `Deterrents Used`, `Animal Response to Deterrents`. No coordinate columns. `Incident Number` is the join key named in the header-description file. This scout did not join the files or count Banff rows in either large CSV.

Collision summary, full small file, at 2026-11-03T00:39:42Z: `https://open.canada.ca/data/dataset/743a0b4a-9e33-4b12-981a-9f9fd3dd1680/resource/72b501f4-bf39-4254-9a19-8cce0b4fdd80/download/pca-number-of-animals-killed-by-collisions-with-vehicles-and-trains-by-site-2010-2024.csv`. HTTP 302 then 200, `text/csv`, `Content-Length: 47894`. Header is `Protected Heritage Area`, `Species Common Name`, `Activity Type`, then `Total Animals 2010` through `Total Animals 2024`. 633 data rows. `Activity Type` has two values, `Driving` and `Railway`. `Protected Heritage Area` has 30 names, including `Banff National Park of Canada` on 88 of those rows (72 distinct species-name values on the Banff rows). Also present: `Glacier National Park of Canada`, `Yoho National Park of Canada`, `Kootenay National Park of Canada`, and `Jasper National Park of Canada`. These rows are park-by-species-by-year sums. They have no coordinates and are not individual carcasses.

## Alberta Wildlife Watch carcass CSV

This is Alberta Transportation and Economic Corridors, not Parks Canada, and the catalog does not say the rows are inside Banff.

`https://open.alberta.ca/api/3/action/package_show?id=alberta-wildlife-watch-animal-carcass-records` at 2026-11-03T00:41:54Z, HTTP 200, `application/json;charset=utf-8`, 11,131 bytes. Title “Alberta Wildlife Watch animal carcass records”. Notes: “The Alberta Wildlife Watch program collects data related to animal-vehicle collisions in Alberta. This data are records of animal carcasses resulting from animal-vehicle collisions reported within the province of Alberta.” `license_id` `OGLA`, `license_title` “Open Government Licence - Alberta”, `license_url` `https://open.alberta.ca/licence`, `sensitivity` `unrestricted`, spatial coverage text `Alberta`, `metadata_modified` `2026-03-05T22:41:50.664118`. Ten yearly CSVs, 2016–2025, are listed. The 2025 resource URL is `https://open.alberta.ca/dataset/3fe2eb9a-1398-4198-be6a-cd0a940385e7/resource/a1ddc795-3cc5-42dc-b710-fb8d476fd413/download/open-data-aww-2025.csv` with catalog size 390506.

That 2025 URL at 2026-11-03T00:41:55Z, range `bytes=0-1200`, was HTTP 206, `text/csv`, `Content-Range: bytes 0-1200/390506`. No login. Header, after a UTF-8 BOM and a leading space on the first name: `Month Observed`, `Year Observed`, `Species`, `Latitude`, `Longitude`. A species value in the range was quoted and contained a comma. The body contains decimal-degree coordinates at full printed precision. This note does not repeat those coordinates. The header has no park, highway, road, or day field. Whether any coordinate falls inside Banff National Park was not tested. Status: `anonymous_csv` for the header and the licence. Row count was not taken.

`https://open.alberta.ca/licence` at 2026-11-03T00:41:55Z, HTTP 200, `text/html; charset=utf-8`. The page says it is version 2.2 of the Open Government Licence – Alberta. Grant sentence matches the Canadian federal pattern, including commercial use. Attribution sentence: “Contains information licensed under the Open Government Licence – Alberta.”

## ArcGIS

`https://www.arcgis.com/sharing/rest/search?q=Banff%20National%20Park%20wildlife&f=pjson&num=8` at 2026-11-03T00:41:55Z, HTTP 200, `text/plain;charset=utf-8`. `total` was 67; eight items were returned. One item is type Feature Service, `access` `public`, `owner` `res55155_USG`, URL `https://services2.arcgis.com/PYn6bWCjT6bhw1z3/arcgis/rest/services/Safe_Passages_in_Banff_National_Park_Res2_WFL1/FeatureServer`. Another item is a Web Mapping Application, owner `Esri_cy_CA`, URL `http://www.highwaywilding.org/viewer/`.

FeatureServer root at 2026-11-03T00:43:01Z, `?f=pjson`, HTTP 200, `application/json; charset=utf-8`, 5,725 bytes. `serviceDescription`: “Showing Wildlife Crossing Structures in Banff National Park”. `copyrightText` is empty. `capabilities` is `Query`. `spatialReference.latestWkid` is 3857. `maxRecordCount` is 2000. Layers returned:

- 4 `Banff_Roads_TransCanada`, polyline. Fields are road-network attributes (`NGD_UID`, `NAME`, `TYPE`, `CLASS`, `RANK`, address-range fields, census-subdivision fields). `copyrightText` empty.
- 5 `Banff_Roads`, polyline (layer id only; fields not requested).
- 6 `Buffered_NPAECs_Banff`, polygon. `copyrightText`: “Urban Parks and Ecological Corridors, Protected Areas Establishment Branch, Protected Areas Establishment and Conservation Directorate, Parks Canada”. Description text is about Parks Canada’s National Program for Ecological Corridors. Fields: `ID`, `Name_En`, `Nom_Fr`, `Desc_En`, `Desc_Fr`, plus shape measures.
- 12 `lrnf000r25a_e`, polyline, same National Geographic Database-style road fields, empty copyright.
- table 15 `YearlyElkMortality.csv`: `Year`, `Elk_Moralities` (spelled that way in the service).
- table 16 `highway_mortalities_2005_2017.csv`: `Year`, `Highway_Mortalities`.

Layer `0` at 2026-11-03T00:43:02Z was HTTP 200 with an error object: code 400, “The requested layer (layerId: 0) was not found.” Layer documents for 4, 6, 12, 15, and 16 were HTTP 200 at 2026-11-03T00:43:53Z–00:43:54Z. No layer in the service document is a crossing-structure point layer, and the mortality tables are two-column annual totals, not the Open Government event CSV. Status: `license_unresolved`. The empty service copyright is not a reuse grant. Feature counts were not queried.

`http://www.highwaywilding.org/viewer/` at 2026-11-03T00:43:02Z, HTTP 200, `text/html`, 7,408 bytes. The body is an ArcGIS Viewer for Flex shell. It tells the client to load Flash and does not contain a FeatureServer or MapServer URL.

## Highway Wilding and the bear-genetics paper

`https://highwaywilding.org/` and `https://www.highwaywilding.org/` at 2026-11-03T00:41:56Z and 2026-11-03T00:41:57Z both returned HTTP 200, `text/html; charset=UTF-8`, 14,878 bytes. The page says Highway Wilding is a partnership among Parks Canada, the Western Transportation Institute, the Miistakis Institute, the Wilburforce Foundation, and the Woodcock Foundation, focused on TCH mitigation in Banff, and that the collaborative project was completed in 2013 and the site is an archive. The homepage HTML has no dataset, CSV, or KMZ link.

Sawaya, Kalinowski, and Clevenger, genetic connectivity of two bear species at Banff crossings:

- `https://pmc.ncbi.nlm.nih.gov/articles/PMC4027379/` at 2026-11-03T00:41:56Z, HTTP 200, `text/html; charset=utf-8`, 230,741 bytes. Data-availability sentence: “Data available from the Dryad Digital Repository: http://dx.doi.org/10.5061/dryad.5q3b3.”
- `https://datadryad.org/dataset/doi:10.5061/dryad.5q3b3` at 2026-11-03T00:43:02Z, HTTP 200, `text/html; charset=utf-8`. The page lists `Banff Bear DNA Data.xlsx`, 44.71 KB, and says the workbook has description, black bear, and grizzly bear worksheets of microsatellite data.
- `https://datadryad.org/api/v2/datasets/doi:10.5061/dryad.5q3b3` at 2026-11-03T00:41:56Z was HTTP 404, `text/html; charset=UTF-8`, empty body. The same path with an encoded DOI, `doi:10.5061%2Fdryad.5q3b3`, at 2026-11-03T00:43:02Z was HTTP 200, `application/json; charset=utf-8`, 3,938 bytes. `license` is `https://spdx.org/licenses/CC0-1.0.html`. `visibility` is `public`. `storageSize` is 44714. Usage notes repeat the 105 black bear and 134 grizzly bear microsatellite description. Locations named in the JSON: Canada, Alberta, Banff National Park.
- `https://datadryad.org/api/v2/datasets/doi:10.5061%2Fdryad.5q3b3/download` at 2026-11-03T00:44:51Z, GET, HTTP 401, `application/json; charset=utf-8`, body `{"error":"Unauthorized, must have current bearer token"}`.
- `https://datadryad.org/downloads/file_stream/66077` (the file href on the landing page) at 2026-11-03T00:45:04Z, HTTP 403, `text/html`, 118 bytes, title “403 Forbidden”.

The xlsx bytes were not read. Status: `download_refused` for those two file URLs. The PMC sentence says the genotypes are in Dryad; this client did not retrieve them. The paper is not a crossing inventory and not a carcass table. No page fetched in this scout said the 1996–present camera or track-pad database is available on request. The crossings program page describes that program and does not offer the file.

## Recommended next adapter

Recommend one Go adapter, for the anonymous Open Government CSV `banff_np_wildlife_crossing_structures_2014-2017_data.csv` only. Decode Windows-1252, skip the French label row, and store one count record per named structure and image span, under the Open Government Licence – Canada, version 2.0, with the attribution sentence from that licence. The schema has structure names and species counts. It has no coordinates, so this adapter does not draw crossings on the TCH.

No second adapter yet. The 2005–2017 Banff mortality CSV is anonymous and licensed, and its retrieved columns have no location. The national coexistence collision file is a park-level sum. The Alberta carcass CSV has coordinates and a provincial licence, and its header does not identify Banff or a highway. The ArcGIS FeatureServer answers anonymously, its service copyright is empty, and it exposes roads, a corridor polygon, and two annual totals rather than crossing points or carcass events.
