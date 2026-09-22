# Sites-guided ranger interface

The user requested the installed Sites plugin for Corridor's UI. The Sites
building workflow was applied to the existing Svelte/MapLibre working surface,
preserving the native Go backend, package lockfile, licenses and privacy policy.

## Design and implementation

The field-atlas direction uses forest-green navigation, a larger primary map,
readable controls and a distinct purple treatment for movement analysis. The
existing collision filters, timeline, linked list, source context, study model,
assessment guidance, dark mode and error states remain connected to the APIs.

On phones the map appears before the optional filter form. Filters retain their
state while collapsed. The movement shortcut moves keyboard focus to the analysis
panel, and a results link returns to the map. Desktop rails scroll independently.
Existing native HTML controls and Svelte state were reused without new packages.

## Hosting boundary

The Sites plugin supports static exports and Cloudflare Worker-compatible server
output. Corridor currently serves its map tiles, licensed basemap and live model
through the local Go service backed by PostGIS. There is no hosted API origin.
Publishing only the static frontend would leave those capabilities disconnected.
No standalone Sites deployment or migration of the Go/PostGIS backend is claimed.
The complete updated UI is served by the existing local Docker application.

## Verification

Two new interaction tests were written first and failed on the previous UI:
mobile filter disclosure/preservation and direct movement-workspace focus.
The existing responsive, accessibility, prediction and live-map tests cover the
retained journeys. Final run results are recorded in PROGRESS.md.

Review checked focus handling, native disclosure/anchor behavior, responsive
scrolling, source attribution and absence of changes to model/data semantics.
The larger font sizes initially exposed a hidden mobile heading and a zoomed
header overflow; both were corrected and verified with the accessibility suite.
No new endpoints, credentials, raw records or unsafe HTML were introduced.

The first populated mobile Lighthouse report measured performance 91,
accessibility 100 and LCP 3.09 seconds. The repeated-run harness then stalled
without producing a second report and was stopped. No new three-run median or
map-ready timing is claimed. The previously documented map-ready and JavaScript
bundle budget gaps remain open; this UI increment does not close performance
acceptance.
