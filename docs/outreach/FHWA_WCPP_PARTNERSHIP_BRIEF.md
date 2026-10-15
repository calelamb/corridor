# Federal Highway Administration (FHWA) Wildlife Crossings Pilot Program (WCPP) Partnership Brief
## How Corridor Accelerates National Park Service and State DOT Grant Applications Under 23 U.S.C. § 171

**Document ID:** `CORRIDOR-FHWA-WCPP-BRIEF-V1`  
**Target Audience:** National Park Service (NPS) Superintendents, Denver Service Center (DSC) Transportation Planners, State Department of Transportation (DOT) Wildlife Program Managers, and FHWA Division Administrators  
**Statutory Authority:** 23 U.S.C. § 171 (Wildlife Crossings Pilot Program; Infrastructure Investment and Jobs Act § 11123)  
**Platform License:** Apache-2.0 (100% Free and Open Source Public Good)  
**Web & Grant Hub:** `https://corridorwildlife.org/?workspace=mitigate`

---

## 1. Executive Summary & Program Context

Section 11123 of the Infrastructure Investment and Jobs Act (IIJA), codified at **23 U.S.C. § 171**, established the landmark **$350 million Wildlife Crossings Pilot Program (WCPP)** under the Federal Highway Administration (FHWA). The program provides competitive federal grants to reduce wildlife-vehicle collisions (WVCs) and restore terrestrial and aquatic habitat connectivity across transportation corridors.

Under 23 U.S.C. § 171(b), **Federal Land Management Agencies (FLMAs)**—specifically including the **National Park Service (NPS)**—are eligible entities authorized to apply directly or partner jointly with State DOTs, Tribal Nations, and regional transportation authorities.

However, assembling a competitive WCPP grant application requires rigorous quantitative evidence: multi-year collision frequency analyses, species vulnerability classifications, validated engineering cost schedules, and empirical Benefit-Cost Ratios (BCRs) computed in accordance with USDOT guidance. For many park units and regional DOT districts, assembling these analyses through private environmental consulting firms costs tens of thousands of dollars and requires 6 to 12 months of lead time.

**Corridor** (https://corridorwildlife.org) completely automates this technical analysis. Built as an open-source public good under the Apache-2.0 license, Corridor combines normalized multi-agency carcass records, 0.1-mile highway segmentation, peer-reviewed mitigation science, and algorithmic knapsack optimization to generate publication-grade, two-page FHWA WCPP Project Briefs (`/v1/mitigation/{id}/brief.pdf`) in seconds.

---

## 2. Statutory Alignment: How Corridor Satisfies 23 U.S.C. § 171 Criteria

FHWA evaluates competitive WCPP proposals against specific statutory criteria defined in 23 U.S.C. § 171(d) and § 171(e). The table below outlines how Corridor directly provides the required empirical evidence for each statutory dimension:

| Statutory Criterion (23 U.S.C. § 171) | Federal Grant Requirement | Corridor Analytical Engine & Evidence Provided |
| :--- | :--- | :--- |
| **Criterion 1: Safety & Crash Reduction** (§ 171(e)(1)(A)) | Leverage historical crash and carcass data to demonstrate substantial reduction in human fatalities, injuries, and animal mortality. | - Normalizes heterogeneous state DOT crash databases and park carcass logs into uniform **0.1-mile linear highway segments**.<br>- Applies peer-reviewed meta-analytic mitigation efficacy rates (**86% collision reduction** for combined wildlife overpasses/underpasses and exclusionary fencing; *Rytwinski et al. 2016*).<br>- Quantifies expected annual crash reductions per segment across target species. |
| **Criterion 2: Ecological Connectivity & Habitat Linkages** (§ 171(e)(1)(B)) | Improve terrestrial or aquatic movement corridors for migratory ungulates, threatened and endangered species, and apex carnivores. | - Integrates USGS Western Association of Fish and Wildlife Agencies (WAFWA) ungulate migration corridors and The Nature Conservancy (TNC) Resilient and Connected Network.<br>- Evaluates habitat fragmentation metrics across park boundary corridors (e.g., Yellowstone US-191, Grand Teton US-89, Glacier US-2, Rocky Mountain US-34, Yosemite CA-120, Grand Canyon AZ-64, Olympic US-101, Shenandoah Skyline Drive / US-211).<br>- Classifies focal species impacts across listed carnivores (grizzly bear, Canada lynx, wolf) and critical ungulate herds. |
| **Criterion 3: Cost-Effectiveness & Benefit-Cost Ratio (BCR)** (§ 171(e)(1)(C)) | Demonstrate high economic return on federal investment through validated Benefit-Cost Analysis (BCA). | - Implements an automated **Benefit-Cost Ratio (BCR) Knapsack Engine** in pure Go.<br>- Employs official USDOT comprehensive crash valuations: **$14,000** for deer, **$45,000** for elk, and **$83,000** for moose (excluding ecosystem service values).<br>- Calculates Net Present Value (NPV), internal rate of return, and payback period across a standard 20-year structural design life with transparent sensitivity bounds. |
| **Criterion 4: Interagency & Tribal Partnerships** (§ 171(e)(1)(D)) | Foster multi-jurisdictional collaboration between federal land managers, state DOTs, wildlife agencies, and Tribal governments. | - Unifies fragmented data silos: harmonizes internal NPS maintenance and ranger carcass logs with State DOT highway patrol collision reports into one single source of truth.<br>- Features open API endpoints (`GET /v1/mitigation`) and standardized GeoJSON/CSV exports accessible to all partnering agencies.<br>- Provides a neutral, transparent scientific framework that builds consensus between park resource managers and state highway engineers. |
| **Criterion 5: Innovative Technology & Field Operations** (§ 171(e)(1)(E)) | Utilize innovative monitoring tools, modern spatial analysis, and advanced construction practices. | - Deploys a zero-install W3C desktop and mobile PWA with sub-10-second offline carcass logging for field rangers.<br>- Incorporates on-device lightweight INT8 quantized ONNX vision models for instantaneous species verification.<br>- Uses TreeSHAP explainability algorithms and Network Kernel Density Estimation (NKDE) along road networks to isolate micro-topographical collision drivers (culverts, vegetation cover, curvature). |

---

## 3. The Corridor Mitigation Prioritization & Knapsack Engine

The core computational driver behind Corridor's grant support is its **Mitigation Prioritization Engine** (`internal/prioritize`), implemented natively in Go for microsecond execution and complete scientific reproducibility.

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                      Input Parameters & Data Layers                         │
│  - 0.1-Mile Highway Segments with Observed & Predicted Collision Risk       │
│  - Species Breakdown (Mule Deer, Elk, Moose, Pronghorn, Carnivores)         │
│  - Capital Budget Ceiling (e.g., $10,000,000 WCPP Grant Request)            │
│  - Candidate Mitigation Treatments (Overpass, Underpass, Exclusion Fencing) │
└──────────────────────────────────────┬──────────────────────────────────────┘
                                       │
                                       ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│              Benefit-Cost Ratio (BCR) Evaluation Model                      │
│  Annual Benefit ($) = ∑ (Annual Collisions_s × USDOT Valuation_s × 0.86)   │
│  Life-Cycle Cost ($) = Capital Construction + 20-Yr Inspection/Maintenance │
│  NPV & BCR Calculation: BCR = Present Value (Benefits) / Total Cost         │
└──────────────────────────────────────┬──────────────────────────────────────┘
                                       │
                                       ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                     Greedy Knapsack Optimization Engine                     │
│  Selects optimal set of contiguous highway segments that maximizes net      │
│  collision reduction and dollar savings without exceeding grant budget.      │
└──────────────────────────────────────┬──────────────────────────────────────┘
                                       │
                                       ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│           Publication-Grade 2-Page FHWA WCPP PDF Project Brief               │
│  Automated PDF generation (`GET /v1/mitigation/{id}/brief.pdf`) ready       │
│  for direct inclusion into Grants.gov / FHWA application packages.           │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 3.1 Economic Modeling Parameters
- **Collision Reduction Factor:** 86% based on global peer-reviewed meta-analyses for comprehensive wildlife fencing paired with dedicated overpasses/underpasses (*Rytwinski et al. 2016, Transportation Research Part D*).
- **USDOT Comprehensive Crash Costs:** Derived from published federal transportation economic guidelines and state DOT synthesis studies (WSDOT 2022, CDOT 2023):
  - White-tailed & Mule Deer: **$14,000** per collision (vehicle repair, medical care, towing, lost productivity).
  - Rocky Mountain & Roosevelt Elk: **$45,000** per collision.
  - Shiras & Western Moose: **$83,000** per collision (reflecting catastrophic cabin-intrusion and severe injury rates).
- **Structure Lifecycle:** Modeled across a standard 20-year structural analysis period with 3% annual real social discount rate and 1.5% annual maintenance contingency.

---

## 4. The 2-Page FHWA WCPP PDF Project Brief

Corridor's embedded reporting service (`GET /v1/mitigation/{id}/brief.pdf`) produces a standardized, two-page, publication-grade PDF engineered specifically for federal grant reviewers:

### Page 1: Executive Summary, Cartography & Safety Profile
- **Project Title & Administrative Header:** Park unit identifier, state highway designation, beginning and ending mileposts (e.g., *Yellowstone National Park — US-191 Gallatin Canyon Wildlife Mitigation Project, MP 18.0 to 24.5*).
- **High-Resolution Geospatial Inset Map:** Cartographic depiction of the 0.1-mile highway corridor, showing historical collision locations, predicted risk heatmaps, surrounding park/forest boundaries, and proposed structure placements.
- **5-Year Historical Collision Summary:** Breakdown of animal strikes by species, seasonal peak distributions (winter range vs summer dispersal), and severity levels.
- **Micro-Hotspot Identification:** Statistical clustering results from Network Kernel Density Estimation (NKDE) and Getis-Ord Gi* hotspot tests, confirming statistical significance ($p < 0.01$).

### Page 2: Economic Valuation, Engineering Scope & Grant Justification
- **Benefit-Cost Analysis (BCA) Table:** Itemized breakdown of capital construction costs, 20-year operations/maintenance costs, avoided annual collision savings, Net Present Value (NPV), and final Benefit-Cost Ratio (BCR).
- **Recommended Engineering Treatment Scope:** Specification of crossing dimensions (e.g., 50-meter vegetated overpass for grizzly bears and ungulates, dual 4m × 7m underpasses, 8-foot high-tensile wildlife exclusionary fencing, and double jump-out escape ramps).
- **Ecological & Migration Linkage Justification:** Narrative and spatial citations connecting the segment to documented ungulate migration pathways and endangered carnivore connectivity corridors.
- **Statutory Eligibility Attestation:** Formally attests to compliance with 23 U.S.C. § 171 eligibility requirements, NEPA readiness, and non-federal cost-share alignment.

---

## 5. Practical Implementation Guide for Park Planners & State DOTs

National Park Service units and their State DOT partners can utilize Corridor to prepare their WCPP applications in four straightforward steps:

### Step 1: Corridor Delineation & Data Validation
1. Access the Corridor Mitigation Workspace at `https://corridorwildlife.org/?workspace=mitigate`.
2. Select your target highway segment (e.g., Grand Teton US-89 from Moose Junction to Moran Junction, Yellowstone US-191 Gallatin Canyon, Glacier US-2 Marias Pass, or Grand Canyon AZ-64 Desert View Drive).
3. Review the combined dataset of state DOT incident reports and park ranger carcass logs. Ensure all recent park records have been uploaded or synchronized.

### Step 2: Configure Budget & Treatment Parameters
1. Enter your anticipated grant funding ceiling (e.g., $10,000,000 federal request with $2,500,000 non-federal local match).
2. Specify structural preferences based on terrain and target species (e.g., large mammal overpass for elk/grizzly vs bridge extension underpass for riparian species).
3. Execute the knapsack optimization to identify the highest-yielding contiguous highway sub-segments.

### Step 3: Export the Publication-Grade Grant Brief
1. Click **"Export FHWA WCPP Brief (PDF)"** or issue an HTTP request to `GET /v1/mitigation/{id}/brief.pdf`.
2. Review the resulting two-page brief with your Division of Resource Management, Chief Ranger, and State DOT district engineers.
3. Incorporate the generated charts, Benefit-Cost Ratio calculations, and map graphics directly into your formal grant narrative in Grants.gov (Form SF-424 and Project Narrative Attachment).

### Step 4: Joint Interagency Co-Sponsorship
1. Transmit the Corridor brief to your State DOT Federal Aid Coordinator as the foundation for a joint application.
2. Under 23 U.S.C. § 171, joint applications between Federal Land Management Agencies and State DOTs receive priority consideration for projects traversing federal lands.

---

## 6. Program Inquiries & Technical Support

The Corridor development team provides complimentary technical assistance to National Park Service units, Tribal Nations, and State DOTs assembling WCPP grant proposals.

- **Platform URL:** `https://corridorwildlife.org`
- **NPS Ranger Hub:** `https://corridorwildlife.org/rangers/`
- **Methodology & Model Documentation:** `https://corridorwildlife.org/methodology/`
- **Grant Assistance Email:** `grants@corridorwildlife.org`
- **Source Code Repository:** `https://github.com/corridor/corridor` (Apache-2.0 License)
