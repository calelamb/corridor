# Corridor NPS Ranger SEO, Desktop App Discovery & Federal Outreach Design

**Status:** Approved for execution  
**Date:** 2026-10-14  
**Target Audience:** National Park Service (NPS) Wildlife Rangers, Biological Resource Managers, Transportation Liaisons, Park Superintendents, and FHWA Wildlife Crossings Pilot Program (WCPP) Coordinators.  
**License:** Apache-2.0 (100% Free and Open Source for the Public Good)

---

## 1. Executive Summary & Intent

Corridor exists to eliminate wildlife-vehicle collisions across North America by predicting road risk, prioritizing high-yield mitigation structures (wildlife overpasses, underpasses, and exclusion fencing), and equipping rangers with rapid, offline field reporting tools.

To get Corridor directly onto the desktop computers and mobile devices of NPS rangers at zero cost:
1. **Search Engine Discovery (SEO):** Rank #1 for authoritative search terms queried by park biologists, resource managers, and federal transportation coordinators (e.g., "NPS roadkill reporting software", "wildlife crossing prioritization tool", "offline carcass reporting app", "FHWA wildlife crossings grant calculator").
2. **Frictionless Desktop Application Experience:** Provide a zero-install-friction Progressive Web Application (PWA) that installs with one click directly onto government workstations (Windows, macOS, Linux, ChromeOS) as a standalone desktop app—no IT administrator privileges required, functioning completely offline with local SQLite/IndexedDB caching.
3. **Dedicated NPS Ranger Hub (`/rangers`):** A purpose-built, high-conversion landing page detailing park-specific benefits (Yellowstone US-191, Grand Teton US-89/26, Glacier US-2), dual-resolution sensitive species protection (H3-r6 + 30-day delay for public vs unredacted for rangers), 2-page FHWA WCPP PDF brief generation, and 10-second field carcass logging.
4. **Targeted Federal Outreach Playbook:** A comprehensive, ethical outbound outreach toolkit (`docs/outreach/`) with contact directories, verified federal program pathways (NRSS Biological Resources Division, Park Resource Management Divisions, FHWA WCPP coordinators), and ready-to-send personalized email templates for park superintendents, chief rangers, and wildlife biologists.

---

## 2. Technical SEO & Discoverability Architecture

### 2.1 Search Engine Metadata & OpenGraph Standards
- **Global `web/src/app.html` & Layout Enhancements:**
  - Standardized `<title>` hierarchy: `[Page Title] · Corridor Wildlife Intelligence`
  - Canonical URL links on every route (`https://corridorwildlife.org/...`) to prevent duplicate indexing.
  - Social & Search Cards: Full OpenGraph (`og:title`, `og:description`, `og:image`, `og:url`, `og:type`) and Twitter Cards (`twitter:card: summary_large_image`).
  - Search engine crawlers enabled via `web/static/robots.txt` with sitemap declaration.
  - XML Sitemap (`web/static/sitemap.xml`) updated to include `/rangers/` with high priority (`0.95`).

### 2.2 Rich Structured Data (JSON-LD Schemas)
- **`SoftwareApplication` Schema:**
  - Name: "Corridor Wildlife Intelligence Platform"
  - Application Category: `GovernmentApplication`, `ScienceApplication`
  - Operating Systems: "Windows, macOS, Linux, Android, iOS, ChromeOS"
  - Offers: Price $0.00 USD (Free Public Good)
  - Feature List: "0.1-mile highway collision risk, TreeSHAP explanation drivers, 10-second offline carcass logging, dual-resolution sensitive species privacy, FHWA WCPP grant brief PDF generator"
- **`FAQPage` Schema (Targeting Search Rich Snippets):**
  - "Is Corridor free for National Park Service units and rangers?" -> Yes, 100% free under Apache-2.0.
  - "Can Corridor be installed on government desktop computers without administrative rights?" -> Yes, as a W3C Progressive Web Application directly from Edge/Chrome/Safari.
  - "How does Corridor protect sensitive and endangered wildlife from poachers?" -> Strict dual-resolution privacy gating with H3 resolution-6 spatial generalization and 30-day temporal delay for unauthenticated public viewers.
  - "How does Corridor assist with Federal Highway Administration (FHWA) grant applications?" -> Built-in benefit-cost ratio (BCR) knapsack engine that exports publication-grade 2-page FHWA Wildlife Crossings Pilot Program (WCPP) PDF project briefs.

---

## 3. Desktop Application Architecture & PWA Manifest

Federal IT workstations often restrict `.exe` or `.msi` installers without lengthy administrative security reviews. A desktop-optimized PWA overcomes this barrier completely while offering native desktop integration.

### 3.1 Web App Manifest (`web/static/manifest.webmanifest`)
- Name: "Corridor — Wildlife Highway Intelligence & Ranger Dispatch"
- Short Name: "Corridor"
- Display: `standalone`
- Orientation: `any`
- Background Color: `#102b23` (deep forest dark theme token)
- Theme Color: `#163d31` (forest brand token)
- Categories: `["government", "productivity", "utilities", "navigation"]`
- Desktop Shortcuts:
  1. "Ranger Field Report" -> `/report/`
  2. "Mitigation Prioritization" -> `/?workspace=mitigate`
  3. "Hotspot Analysis Lab" -> `/?workspace=hotspot`
  4. "Scientific Methodology" -> `/methodology/`
- Icons: Vector SVG and raster PNG icon declarations with maskable support.

### 3.2 In-App Desktop Install Promotion
- Component `DesktopInstallBanner.svelte`:
  - Listens to `beforeinstallprompt` browser event.
  - Renders a clean, accessible banner and header button ("Install Desktop App") with a badge "Free for NPS Rangers".
  - One-click installation triggers native browser desktop application prompt.
  - Seamlessly persists dismissed state in local storage.

---

## 4. Dedicated NPS Ranger Hub (`/rangers`)

A high-converting, authoritative page engineered specifically for federal park staff:
- **Hero:** "Built for the Keepers of the Wild. Free for National Park Service Rangers."
- **Key Capabilities for Park Operations:**
  1. *Sub-10-Second Carcass Logging:* Works in complete dead zones (IndexedDB queue), auto-detects highway milepost and segment, on-device species AI suggestion chips.
  2. *Dual-Resolution Poaching Protection:* Full unredacted GPS for verified park staff; public sees only generalized H3-r6 polygons delayed by 30 days.
  3. *FHWA Grant PDF Generator:* One-click generation of two-page funding briefs for wildlife overpasses and fencing along park boundary highways (e.g., US-191, US-89).
  4. *Live Dispatch & Incident Stream:* Real-time SSE chimes and dispatch feeds for ranger stations.
- **Direct Action Callouts:**
  - "Launch Desktop App"
  - "Download Ranger Quickstart Guide"
  - "Request Agency OIDC / SSO Integration"

---

## 5. Federal Outreach Playbook & Templates (`docs/outreach/`)

Structured materials for direct communication with federal and state agencies:
1. `NPS_RANGER_ADOPTION_PLAYBOOK.md`: Targeting roadmap across primary parks (Yellowstone, Grand Teton, Glacier, Rocky Mountain, Yosemite, Olympic, Shenandoah), National Park Service leadership directories, and regional DOT partnerships.
2. `NPS_EMAIL_TEMPLATES.md`: Tested, non-spammy, highly professional email templates for:
   - Park Wildlife Biologists & Resource Chiefs
   - Chief Park Rangers & Dispatch Coordinators
   - Park Superintendents & Transportation Planners
3. `FHWA_WCPP_PARTNERSHIP_BRIEF.md`: Briefing memo explaining how parks can use Corridor's BCR ranking and PDF exports to submit competitive applications for the $350M FHWA Wildlife Crossings Pilot Program.

---

## 6. Verification and Quality Criteria

- **Automated Tests:**
  - Unit tests in Vitest for install prompt, SEO metadata helpers, and PWA queue.
  - End-to-end Playwright tests verifying `/rangers` page rendering, metadata tags, structured data validation, and responsive mobile/desktop layouts.
  - WCAG 2.2 AA accessibility audit with zero Axe violations.
- **Lighthouse Performance & SEO:**
  - Lighthouse SEO score: 100/100.
  - Lighthouse Accessibility score: 100/100.
