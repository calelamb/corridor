# Global roadkill, citizen science and image data sweep

Checked 2026-09-21 UTC. This is research/acquisition evidence, not production
coverage. `downloaded` below means privately cached originals; publication still
requires record QA, provenance and the shared sensitive-location policy.

## Catalog

| ID | Dataset and exact access | Rights evidence / access | Coverage and usable information | Corridor use and remaining gate |
| --- | --- | --- | --- | --- |
| G01 | **Global Roadkill Data, Figshare v5**: [metadata](https://api.figshare.com/v2/articles/25714233), [version DOI](https://doi.org/10.6084/m9.figshare.25714233.v5), [CSV](https://ndownloader.figshare.com/files/53393273) | CC BY 4.0 explicitly in version metadata. **Downloaded**, 90,502,616 bytes; publisher MD5 verified and SHA256 recorded. | Inspected: 177,428 unique occurrence rows, 54 country codes, 1971–2024; `numberOfRoadkill` sums to 208,570. US: 5,211 rows, 8,854 animals, 1983–2023. Fields include IDs, taxonomy, uncertainty, partial dates, road ID/type, survey type/effort, source references. | Best new immediately accessible US collision pilot candidate. Preserve aggregate multiplicities; never turn one aggregate into invented individual events. Confirm precise publication policy, dedup and source lineage. |
| G02 | **Global Roadkill opportunistic records**, GBIF key `65908f95-5ab6-48d9-bf6d-6da274ed730e`: [metadata](https://api.gbif.org/v1/dataset/65908f95-5ab6-48d9-bf6d-6da274ed730e), [DwC-A](https://ipt.gbif.pt/ipt/archive.do?r=globalroadkilldataset) | CC BY 4.0, metadata version 1.8. Metadata/count query downloaded; archive not acquired. | Live US indexed count 3,344, years 1983–2023. Opportunistic records; no survey denominator. | Adapter/updated distribution of the same source family as G01, **not additive independent data**. Compare occurrence IDs and versions before selection. |
| G03 | **Global Roadkill systematic records**, GBIF key `d3b6cb30-0a64-4f82-91ea-0bb14637ee17`: [metadata](https://api.gbif.org/v1/dataset/d3b6cb30-0a64-4f82-91ea-0bb14637ee17), [DwC-A](https://ipt.gbif.pt/ipt/archive.do?r=global_roadkill_surveys) | CC BY 4.0, metadata version 1.10. Metadata/count query downloaded; archive not acquired. | Live US indexed count 1,867, years 2004–2023. Sampling/event structure requires archive inspection. | Potential effort-aware model input. Keep sampling events, absences if explicitly supplied, survey duration and repeat visits. Do not infer unobserved road-days as surveyed zeros. |
| G04 | **Roadkill Reports / Anecdata**, [GBIF metadata](https://api.gbif.org/v1/dataset/053d3e84-a440-42ae-a8ab-acda381469bb), [DwC-A](https://ipt.gbif.us/archive.do?r=roadkill-109) | CC BY 4.0, version 1.2; metadata/count query acquired. | Live US count 386, 2015–2023; citizen reports with observation date/species/location. Public catalog extent includes points beyond stated US scope. | Secondary observations. Verify country/coordinate anomalies, specimen meaning and duplicates with other aggregators. No effort or movement inference. |
| G05 | **iNaturalist licensed observations**: [developer/data access guidance](https://www.inaturalist.org/pages/developers), [API practices](https://www.inaturalist.org/pages/api+recommended+practices), [licenses](https://help.inaturalist.org/en/support/solutions/articles/151000175695) | Per-record and per-media licenses; default CC BY-NC is not unrestricted commercial reuse. Prefer CC0/CC BY records and separately cleared media. Exports/GBIF preferred for bulk; account/export may be needed. No new occurrence export acquired. | Global species/time/public-uncertainty observations, with annotation/project/field evidence when present. Weekly GBIF archive is large. | Habitat/presence context and explicitly verified roadkill. Dead does not imply traffic death; nearby road does not imply roadkill. Respect obscured/private coordinates. |
| G06 | **Taiwan Roadkill Observation Network**: [metadata](https://api.gbif.org/v1/dataset/db09684b-0fd1-431e-b5fa-4c1532fbdb14), [published DwC-A](https://ipt.taibif.tw/archive.do?r=tw_roadkill_data) | Published GBIF dataset CC BY 4.0, version 1.11. Metadata acquired. [Native query policy](https://roadkill.tw/data/queryform/occurrence) distinguishes released GBIF/TaiBIF data from permission-controlled native data. | Metadata time scope 2011-08 through 2017-12; Taiwan. Do not infer newer coverage from recent metadata modification. | International transfer/adapter testing; not western-US coverage. Use the specifically released distribution and verify its fields/precision. |
| G07 | **BOKU Project Roadkill**: [GBIF metadata](https://api.gbif.org/v1/dataset/d0d5ef85-71b2-4da6-b6f6-c1c3d60987d3), [DwC-A](https://www.zobodat.at/gbif/roadkill/DwCA.zip), [publisher FAQ](https://roadkill.at/en/explore/faqsen) | CC BY 4.0; metadata acquired. GBIF describes quality-level 1; more reviewed quality-level 2 is separately released on Zenodo. | Primarily Austria; publisher describes 2014–2023 records and a shift to Austria-only reporting in 2021. | International QA/transfer experiments. Pin quality tier and version; avoid duplicate imports of the two quality levels. Not an animal tracking dataset. |
| G08 | **Luxembourg ANF roadkill**: [GBIF metadata](https://api.gbif.org/v1/dataset/32651421-c517-48a3-a80c-cb2d866ea1f6), [DwC-A](https://ipt.mnhn.lu/ipt/archive.do?r=gbif_roadkill_anf) | CC0 1.0, version 1.52; metadata acquired. | Luxembourg agency/citizen records. Catalog notes records were formerly included in another museum dataset. Exact row/time schema not inspected. | Reusable international comparison candidate; dedup against parent museum records. No US coverage claim. |
| G09 | **NatureScot Deer Vehicle Collisions**: [catalog](https://www.data.gov.uk/dataset/838b88d8-7509-435c-9649-90f1881b5ad7/deer-vehicle-collisions), [GeoJSON ZIP](https://gis-downloads.nature.scot/DVC_SCOTLAND_GEOJSON_4326.zip) | **Downloaded**. OGL v3 in catalog access constraints **and embedded XML**; credit Deer Initiative/SNH. Top-level catalog license field is blank, but this does not override the explicit embedded license. | Inspected 16,383 Point features, 2008–2018, Scotland. Fields: incident date, species, road number, year/month, grid accuracy and source-class indicators. GeoJSON EPSG:4326; other distributions use British National Grid. | Strong international map/adapter baseline. Preserve accuracy and report origin; public generalized output only after QA. Not western-US training evidence. |
| G10 | **British Columbia WARS 1978–2025**: [official access page](https://www2.gov.bc.ca/gov/content/transportation/transportation-infrastructure/engineering-standards-guidelines/environmental-management/wildlife-management/wildlife-accident-reporting-system), [2026 data-use license](https://www2.gov.bc.ca/assets/gov/driving-and-transportation/transportation-infrastructure/engineering-standards-and-guidelines/environment/wars/historical-data/mott_wars_data_use_licence_agreement_2026.pdf) | **Restricted**: reviewed PDF forbids redistribution and derived data products/services for distribution, including free/noncommercial use. Page/PDF evidence acquired; **no WARS datasets downloaded**. | Province highway maintenance carcass records, Excel intervals from 1978–2025. Direct spreadsheet links are on the page. | Cannot include in the distributed/public Corridor product under these terms. Separate permission is needed; free download is insufficient. |
| G11 | **North American Camera Trap Images (NACTI)**: [publisher/license/downloads](https://lila.science/datasets/nacti) | CDLA permissive variant explicitly linked by publisher; verify linked license/version with chosen files. Metadata links documented but not downloaded. | Publisher describes 3.7M images from five US locations, 28 categories; COCO Camera Traps JSON. Images roughly 1.37TB, metadata JSON 44MB/CSV31MB. | Species classifier training/evaluation; sample per species/location rather than download everything. Not collision labels, GPS paths or blanket license for other LILA datasets. |
| G12 | **Caltech Camera Traps**: [publisher/license/downloads](https://lila.science/datasets/caltech-camera-traps) | CDLA permissive variant explicitly linked by publisher. No images acquired. | Publisher describes 243,100 images, 140 southwestern-US camera locations, 21 animal categories; location-based train/validation splits. Image archive105GB; annotations9MB/35MB; split4KB. | Useful regional species-recognition benchmark. Respect spatial splits and don't equate image count with independent encounters. |
| G13 | **Wildlife Insights public projects**: [download rules](https://www.wildlifeinsights.org/get-started/data-download/public), [FAQ](https://wildlifeinsights.org/faq) | Per-project CC0/CC BY/CC BY-NC; account and agreement required for download. No account created or agreement accepted; no data acquired. | Global camera detections/deployments; public sensitive locations obfuscated. Image rights separate from detection data. | Potential seasonal occupancy/detection context. Select named projects and rights before claiming coverage. Cameras observe detections, not continuous animal trajectories. |

## Inspected US pilot: Interstate 84 / 84–86 records

G01 contains 2,200 US rows whose road IDs are `Interstate 84` or `Interstate 84/86`,
2004–2020, representing 2,312 reported animals. The mix is 1,416 systematic and
784 opportunistic rows, heavily dominated by barn owls (1,574 rows). Mule deer
account for 113 rows. Preserve source strata rather than marketing this as a
representative deer dataset. Coordinate uncertainty is recorded as30m for these
rows, but that is a source attribute, not verified survey accuracy.

Associated references include [Boves 2011 Boise State thesis](https://scholarworks.boisestate.edu/td/458),
[its wildlife-management paper](https://doi.org/10.1002/jwmg.378), and
[an Ibis study](https://doi.org/10.1111/ibi.12593). Validate the precise highway
extent from cleared spatial records and source methods before advertising a
continuous covered highway. Other US rows include Florida and local-road studies.

This is an acquisition lead that can make the map useful. It is not yet a
validated collision-prediction corpus, and it provides no animal trajectories.

## Version and unit traps discovered

- Figshare v5 and the live GBIF pair each total5,211 US rows in this check, but
  survey labels split differently: Figshare3,281 opportunistic/1,930 systematic
  versus GBIF3,344/1,867. Matching totals do not prove equivalent releases.
- The CSV has177,428 rows and208,570 summed animals. Current GBIF descriptions
  mention274,132 compilation records. Pin a distribution/version and derive
  counts from its actual file; never blend catalog marketing totals.
- G01 has9,418 rows globally and926 US rows whose multiplicity is not1.
  Aggregates belong in the aggregate data model with their stated spatial/time
  uncertainty, not expanded fabricated points.
- Only two US rows lack a day under the narrow blank/NA check, but a complete
  date-validity audit still must reject sentinels and impossible dates.
- Empty reference fields remain in2,385 US rows. Retain dataset-level attribution
  and inspect upstream provenance before record-level redistribution.

## Acquisition and inspection evidence

Originals remain ignored under `data/raw/sweep-2026-09-21/`. Metadata, receipts
and aggregate-only profiles are in `data/evidence/sweep-2026-09-21/global/`.
The final sweep manifest preserves exact URLs, acquisition times, bytes and
SHA256 values. Dataset hashes and evidence-page hashes are explicitly distinct.
No ingestion, model training, public coordinate output or production coverage
has been added by this research task.

Priority: G01 US pilot first; compare G02/G03 lineage before choosing one source
representation; G04 supplements observations; G09 is a strong separately labeled
international testbed. G10 is excluded from distributed products pending permission.
