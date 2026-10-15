# National Park Service (NPS) Ranger & Resource Management Adoption Playbook
## Deploying Corridor for Highway Wildlife-Vehicle Collision Mitigation, Carcass Reporting, and FHWA WCPP Grant Prioritization

**Document ID:** `CORRIDOR-NPS-PLAYBOOK-V1`  
**Target Agency:** National Park Service (U.S. Department of the Interior)  
**Applicable Units:** National Parks, National Preserves, National Recreation Areas, and Parkway Units  
**Platform License:** Apache-2.0 (100% Open Source and Free Public Good)  
**Platform Repository:** `https://github.com/corridor/corridor`  
**Web & Desktop PWA Hub:** `https://corridorwildlife.org/rangers/`

---

## 1. Executive Summary & Purpose

Each year, thousands of native ungulates and apex carnivores are killed on federally owned or state-administered highways traversing National Park Service (NPS) units. Beyond devastating wildlife mortality, wildlife-vehicle collisions (WVCs) endanger park visitors and staff, damage government-furnished vehicles, consume ranger patrol hours in humane animal dispatch and carcass disposal, and fragment critical migration corridors connecting park interiors to adjacent federal, state, and tribal lands.

**Corridor** is a purpose-built, high-performance, open-source wildlife highway intelligence platform. It provides National Park Service units with:
1. **Zero-Friction Desktop Installation:** A W3C Progressive Web Application (PWA) that installs on Government Furnished Equipment (GFE) workstations (Windows, macOS, Linux, ChromeOS) with one click—requiring **zero administrative rights** and zero software licensing expenditure.
2. **Sub-10-Second Offline Carcass Reporting:** Field-tested for ranger patrol vehicles operating in mountain dead zones, featuring one-handed glove-friendly controls, automatic highway milepost snapping, local IndexedDB queuing, and on-device species identification assistance.
3. **Dual-Resolution Poaching Protection:** Automated spatial and temporal privacy gating. Unauthenticated public viewers see only generalized H3 resolution-6 hexagon aggregates (~36 km²) delayed by 30 days, while verified park biologists and law enforcement rangers retain instantaneous access to unredacted, sub-meter GPS coordinates.
4. **Automated FHWA WCPP 2-Page Grant Briefs:** A built-in knapsack optimization engine that evaluates highway segments using peer-reviewed mitigation efficacy (86% collision reduction for crossing structures and exclusion fencing; Rytwinski et al. 2016) and U.S. Department of Transportation (USDOT) comprehensive crash valuations ($14,000 per deer, $45,000 per elk, $83,000 per moose) to produce publication-grade two-page grant briefs ready for submission under 23 U.S.C. § 171.

This playbook establishes a standard operating procedure for park superintendents, wildlife biologists, chief rangers, communications supervisors, and regional transportation liaisons to adopt Corridor across park operations.

---

## 2. Priority National Park Service Units & Strategic Highway Corridors

Corridor is pre-configured with high-resolution 0.1-mile segmentation, topological network models, and habitat layers for major park corridors across North America. The following seven flagship units represent primary targets for initial operational deployment:

| National Park Service Unit | Key Highway Corridors | Focal Species & Ecological Concerns | Primary Interagency Partners |
| :--- | :--- | :--- | :--- |
| **Yellowstone National Park (YELL)** | **US-191** (Gallatin Canyon, West Yellowstone to Big Sky/Bozeman), **US-89** (Gardiner North Entrance), **US-212** (Northeast Entrance / Beartooth Highway) | Grizzly bear (*Ursus arctos horribilis*), Gray wolf (*Canis lupus*), Bison, Elk, Moose, Pronghorn | Montana Dept. of Transportation (MDT), Custer Gallatin National Forest, Wyoming Dept. of Transportation (WYDOT) |
| **Grand Teton National Park (GRTE)** | **US-89/US-26/US-191** (Jackson Hole Valley spine: Moose, Moran, and Buffalo Valley junctions) | Moose (*Alces alces shirasi*), Migratory Elk (Jackson Herd), Pronghorn ("Path of the Pronghorn"), Grizzly bear | WYDOT, Bridger-Teton National Forest, U.S. Fish and Wildlife Service (National Elk Refuge) |
| **Glacier National Park (GLAC)** | **US-2** (Southern boundary: West Glacier to East Glacier across Marias Pass, Essex, Middle Fork Flathead River) | Grizzly bear, Canada lynx (*Lynx canadensis*), Bighorn sheep, Moose, Elk, Wolverine | MDT, Flathead National Forest, Blackfeet Nation, Burlington Northern Santa Fe (BNSF) Railway |
| **Rocky Mountain National Park (ROMO)** | **US-34** (Trail Ridge Road, Fall River Corridor, Kawuneeche Valley), **US-36** (Lyons to Estes Park approach), **CO-7** | Rocky Mountain elk (heavy winter range concentration), Bighorn sheep, Mule deer, Shiras moose | Colorado Dept. of Transportation (CDOT), Arapaho and Roosevelt National Forests, Town of Estes Park |
| **Yosemite National Park (YOSE)** | **CA-120** (Tioga Road high Sierra crossing, Big Oak Flat Road), **CA-140** (Merced River Canyon), **CA-41** (Wawona Road) | Sierra Nevada bighorn sheep (Federally Endangered), Pacific fisher, California black bear, Mule deer | California Dept. of Transportation (Caltrans District 10 & 6), Stanislaus & Sierra National Forests |
| **Olympic National Park (OLYM)** | **US-101** (Olympic Loop encircling the peninsula: Lake Crescent, Soleduck, Hoh, and Quinault river corridors) | Roosevelt elk (*Cervus canadensis roosevelti* - world's largest herd), Black bear, Cougar, Columbian black-tailed deer | Washington State Dept. of Transportation (WSDOT Olympic Region), Olympic National Forest, Tribal Nations |
| **Shenandoah National Park (SHEN)** | **Skyline Drive** (MP 0 to 105 Blue Ridge spine), **US-211** (Thornton Gap crossing), **US-33** (Swift Run Gap crossing), **I-64** (adjacent south) | White-tailed deer (acute collision density), Black bear, Bobcat | Virginia Dept. of Transportation (VDOT Culpeper and Staunton Districts), Augusta and Rockingham Counties |

---

## 3. Park Organizational Structure & Stakeholder Alignment

A successful park-wide rollout coordinates across four functional divisions and regional liaisons:

```
                  ┌──────────────────────────────────────────────┐
                  │          Park Superintendent & Staff         │
                  │   Interagency Agreements & FHWA Leadership   │
                  └──────────────────────┬───────────────────────┘
                                         │
         ┌───────────────────────────────┼───────────────────────────────┐
         │                               │                               │
┌────────▼──────────────┐     ┌──────────▼────────────┐     ┌────────────▼──────────┐
│ Division of Resource  │     │ Division of Visitor & │     │ Communication Center  │
│ Management & Science  │     │ Resource Protection   │     │ (Interagency Dispatch)│
│ (RMS)                 │     │ (VRP - Law Enf./Field)│     │                       │
│ - Wildlife Biologists │     │ - Chief Park Ranger   │     │ - Dispatch Supervisors│
│ - GIS Coordinators    │     │ - District Patrols    │     │ - Real-Time SSE Feed  │
│ - Data Exports/Models │     │ - Roadkill Loggers    │     │ - Incident Chimes     │
└────────┬──────────────┘     └──────────┬────────────┘     └────────────┬──────────┘
         │                               │                               │
         └───────────────────────────────┼───────────────────────────────┘
                                         │
                  ┌──────────────────────▼───────────────────────┐
                  │ Facility Management & Maintenance / DOTs     │
                  │ - Carcass Removal Crews & Regional DOT Lead │
                  │ - FHWA WCPP Project Application Coordination │
                  └──────────────────────────────────────────────┘
```

### 3.1 Division of Resource Management and Science (RMS)
- **Key Roles:** Wildlife Biologists, Ungulate & Carnivore Ecologists, GIS Coordinators, Natural Resource Specialists.
- **Corridor Workflow:**
  - Review and export unredacted, high-precision carcass records via authenticated API or CSV/GeoJSON downloads.
  - Evaluate collision cluster density against GPS-collar migration corridors and seasonal range polygons.
  - Run network Kernel Density Estimation (NKDE) and Getis-Ord Gi* hotspot analytics to validate wildlife crossing locations.
  - Author and approve ecological impact summaries embedded in FHWA grant applications.

### 3.2 Division of Visitor and Resource Protection (VRP)
- **Key Roles:** Chief Park Ranger, Patrol District Supervisors, Field Protection Rangers, Wildlife Management Technicians (bear/bison jam squads).
- **Corridor Workflow:**
  - Install Corridor PWA on patrol vehicle Toughbooks and mobile phones.
  - Use 10-second roadside reporting to record strikes, injured wildlife dispatches, and animal highway crossings without impeding traffic control.
  - Benefit from offline capability in canyon and backcountry road segments with zero cellular connectivity.

### 3.3 Dispatch Centers & Interagency Communications
- **Key Roles:** Lead Telecommunications Specialists, Interagency Dispatchers (e.g., Yellowstone Communications Center, Teton Interagency Dispatch Center, Olympic Peninsula Interagency Communications Center).
- **Corridor Workflow:**
  - Keep Corridor open in a dedicated desktop browser window or standalone PWA tab.
  - Monitor the real-time Server-Sent Events (SSE) live incident stream (`/v1/stream/reports`), receiving audible and visual chimes whenever field rangers log a carcass or injured animal hazard.
  - Coordinate immediate road maintenance or wildlife staff responses without radio clutter.

### 3.4 Division of Facility Management & Road Maintenance
- **Key Roles:** Roads & Trails Supervisors, Maintenance Equipment Operators, Regional DOT Highway Maintenance Crews.
- **Corridor Workflow:**
  - Log carcass pickups during daily highway sweep operations.
  - Identify recurring fence breaches, gate failures, or drainage culverts functioning as informal wildlife passages.

### 3.5 Regional Transportation Liaisons & Federal Highway Partners
- **Key Roles:** NPS Denver Service Center (DSC) Transportation Project Managers, NPS Regional Transportation Program Managers (Regions 1 through 12), State DOT Environmental and Wildlife Liaisons.
- **Corridor Workflow:**
  - Generate quantitative 2-page project briefs for upcoming Federal Highway Administration (FHWA) Wildlife Crossings Pilot Program (WCPP) grant cycles.
  - Leverage Corridor's benefit-cost ratio (BCR) knapsack optimization to prioritize multi-million dollar capital improvement requests with state DOTs.

---

## 4. Core Platform Capabilities & Value Propositions

### 4.1 100% Free Public Good (Apache-2.0 License)
Federal agencies frequently face long, expensive procurement cycles for proprietary software. Corridor is released under the permissive **Apache-2.0 open-source license**. 
- **Zero Cost:** No per-seat licenses, no software maintenance agreements, no recurring SaaS subscriptions.
- **No Vendor Lock-In:** Parks can self-host the single Go binary on internal infrastructure or utilize the hosted public instance at `https://corridorwildlife.org`.
- **Open Data & Reproducible Science:** All analytical formulas (NKDE, Getis-Ord Gi*, TreeSHAP, knapsack optimization) are fully documented in `docs/` and inspectable in pure Go source code.

### 4.2 1-Click Desktop PWA Installation on Government Furnished Equipment (GFE)
NPS field and dispatch computers operate under strict Department of the Interior (DOI) IT security policies that prohibit unapproved `.exe` or `.msi` installations.
- Corridor complies fully with W3C Progressive Web Application standards (`/manifest.webmanifest`).
- Any user opening `https://corridorwildlife.org/rangers/` in Microsoft Edge, Google Chrome, Apple Safari, or Mozilla Firefox can click **"Install Desktop App"** in the navigation header.
- The platform installs instantly as an isolated, sandboxed desktop application with its own window title bar, desktop shortcut, and dock icon.
- **Zero administrative privileges** are required to install, run, or update the application.

### 4.3 10-Second Offline Field Carcass Logging
Patrol rangers responding to roadside strikes in deep river canyons (e.g., Gallatin Canyon on US-191, Lake Crescent on US-101) cannot rely on LTE/5G connectivity.
- **IndexedDB Persistence:** Reports submitted while disconnected are safely stored in local browser storage and automatically synchronize with the server once connectivity is restored.
- **Automatic Spatial Snapping:** Device GPS coordinates instantly snap to the nearest highway segment, state route identifier, and linear milepost marker.
- **On-Device Vision Classifier:** Built-in lightweight INT8 quantized ONNX vision model provides instantaneous species identification suggestion chips (Mule deer, Elk, Moose, Pronghorn, Bighorn sheep, Black bear, Grizzly bear, Mountain lion) from carcass photographs.
- **Glove-Friendly Touch Targets:** Minimum 48px × 48px touch bounding boxes designed for cold-weather tactical patrol gloves.

### 4.4 Dual-Resolution Sensitive Species Poaching Protection
Public roadkill tracking tools risk revealing the coordinates of vulnerable wildlife, attracting trophy hunters, poachers, or unmanaged tourist crowds. Corridor implements an immutable dual-resolution privacy model:
- **Public & Unauthenticated Viewers:**
  - Location coordinates for all sensitive species (all apex predators, listed ungulates, raptors, and rare carnivores) are generalized to **H3 resolution-6 spatial hexagons** (~36 km² cell area).
  - All public reports are subject to a **mandatory 30-day temporal quarantine delay** before appearing on public maps or API endpoints.
- **Authenticated Park Biologists & Rangers:**
  - Park personnel authenticate via OpenID Connect (OIDC) agency single sign-on or cryptographically secured agency API keys.
  - Rangers have immediate, real-time access to sub-meter GPS coordinates, precise highway mileposts, carcass disposition notes, and incident photographs.

### 4.5 Automated FHWA WCPP 2-Page PDF Grant Briefs
The Bipartisan Infrastructure Law established the **$350 million Wildlife Crossings Pilot Program (WCPP)** under 23 U.S.C. § 171. Applying for these competitive grants typically requires costly consultant studies.
- Corridor's **Mitigation Prioritization Engine** performs greedy knapsack optimization across target highway segments.
- Calculates empirical Benefit-Cost Ratios (BCRs) using peer-reviewed treatment effectiveness data (86% crash reduction for combined wildlife overpasses/underpasses and exclusion fencing, Rytwinski et al. 2016) and USDOT crash valuations ($14,000 for deer, $45,000 for elk, $83,000 for moose).
- With a single click (`GET /v1/mitigation/{id}/brief.pdf`), Corridor compiles and renders a publication-grade, two-page PDF grant brief complete with project map, collision history, species impact analysis, economic return calculation, and engineering specifications ready for grant submission.

---

## 5. Step-by-Step Park Onboarding Roadmap

```
Week 1 (Days 1–7)     Week 2–3 (Days 8–21)    Week 4–5 (Days 22–35)   Week 6–8 (Days 36–60)
┌─────────────────┐   ┌───────────────────┐   ┌───────────────────┐   ┌───────────────────┐
│  Phase 1: Pilot │   │ Phase 2: Role     │   │ Phase 3: Field    │   │ Phase 4: DOT &    │
│  Desktop Install│──▶│ Access & Legacy   │──▶│ Rollout & Live    │──▶│ FHWA Grant Brief  │
│  & Verification │   │ Data Sync         │   │ Dispatch Feed     │   │ Production        │
└─────────────────┘   └───────────────────┘   └───────────────────┘   └───────────────────┘
```

### Phase 1: Rapid Pilot Setup (Week 1 / Days 1–7)
1. **Desktop App Installation:**
   - Wildlife Lead and Dispatch Supervisor navigate to `https://corridorwildlife.org/rangers/`.
   - Click "Install Desktop App" on two primary workstation computers and one field patrol Toughbook or tablet.
   - Verify that the app launches in standalone window mode without administrative prompts.
2. **Offline Simulation Test:**
   - Set device to Airplane Mode.
   - Navigate to `/report/` and submit a test carcass observation with a photograph.
   - Verify the report queues in IndexedDB and displays a pending badge.
   - Re-enable connectivity and verify the queued report synchronizes.

### Phase 2: Role-Based Access & Historical Sync (Weeks 2–3 / Days 8–21)
1. **Access Authorization:**
   - Designate park wildlife staff requiring unredacted GPS access.
   - Connect Corridor to agency authentication or request agency API keys from the platform administrator.
2. **Legacy Record Harmonization:**
   - Export park carcass logs from internal spreadsheets, wildlife incident databases, or dispatch logs.
   - Format records to Corridor's open schema (`timestamp`, `latitude`, `longitude`, `species`, `sex`, `age_class`, `disposition`).
   - Run bulk ingestion to backfill historical collision distributions for park corridors.

### Phase 3: Field Crew Rollout & Dispatch Integration (Weeks 4–5 / Days 22–35)
1. **Field Staff Briefing:**
   - Conduct a 15-minute roll-call demonstration for VRP protection rangers and road maintenance operators.
   - Distribute the 1-page Corridor Field Card detailing the 10-second reporting flow.
2. **Dispatch Center Live Incident Integration:**
   - Launch Corridor on a secondary dispatch console monitor.
   - Verify that real-time Server-Sent Events (`/v1/stream/reports`) play audible chimes when new field submissions occur along park highways.

### Phase 4: Interagency DOT Alignment & FHWA WCPP Grant Generation (Weeks 6–8 / Days 36–60)
1. **Mitigation Optimization Run:**
   - Open Corridor's Mitigation Prioritization Workspace (`/?workspace=mitigate`).
   - Select the park's primary highway corridor (e.g., US-191 Gallatin Canyon or US-89 Jackson Hole).
   - Enter estimated budget ceilings and target crossing structure types.
2. **Interagency Coordination Session:**
   - Convene a joint meeting between Park RMS, VRP, and State DOT regional engineers (e.g., WYDOT District 3 or MDT Bozeman District).
   - Review Corridor's benefit-cost calculations and identified top-tier mitigation segments.
3. **Export Grant Documentation:**
   - Export 2-page FHWA WCPP PDF Project Briefs for the top 3 ranked segments.
   - Attach briefs directly to state DOT and federal grant application packages for upcoming WCPP funding opportunities.

---

## 6. IT Security, Data Privacy, and GFE Compliance

### 6.1 GFE Workstation Safety
- **No Native Executable Binaries:** The desktop PWA operates entirely within the sandboxed browser execution environment (Chromium or WebKit). It cannot modify system registry keys, execute unauthorized background services, or access files outside its browser sandbox.
- **FIPS 140-3 Encryption in Transit:** All traffic is transmitted exclusively over TLS 1.3 with forward secrecy.
- **Zero Third-Party Tracking:** Corridor contains zero commercial ad trackers, marketing pixels, or analytics trackers.

### 6.2 Data Sovereignty & Interagency Export
- All data collected by park personnel remains in public domain ownership under federal open government standards.
- Parks may export complete unredacted datasets at any time in open formats (CSV, GeoJSON, Parquet) for use in ArcGIS Pro, R, QGIS, or internal NPS NRSS databases.

---

## 7. Adoption Checklist

- [ ] **Step 1:** Visit `https://corridorwildlife.org/rangers/` and install the desktop PWA on station PC.
- [ ] **Step 2:** Test 10-second offline reporting on a field mobile device or patrol laptop.
- [ ] **Step 3:** Designate primary point of contact in Division of Resource Management and Science.
- [ ] **Step 4:** Establish agency credentials for unredacted sensitive species access.
- [ ] **Step 5:** Harmonize legacy carcass data from the past 3–5 seasons into Corridor.
- [ ] **Step 6:** Activate Live Dispatch SSE feed in the park communications center.
- [ ] **Step 7:** Run Corridor Mitigation Prioritization on park boundary highways.
- [ ] **Step 8:** Generate 2-page FHWA WCPP PDF project briefs and schedule interagency review with State DOT.

*For technical assistance, agency OIDC integration, or bulk data ingestion support, contact the Corridor Federal Outreach Team at `outreach@corridorwildlife.org`.*
