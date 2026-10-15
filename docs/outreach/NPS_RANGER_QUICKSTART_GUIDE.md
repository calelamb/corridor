# National Park Service (NPS) Ranger Field Quickstart Guide
## Tactical Reference Card for Vehicle Visors, Patrol Toughbooks, and Dispatch Consoles

**Document ID:** `CORRIDOR-NPS-QUICKSTART-CARD-V1`  
**Target Audience:** Protection Rangers, Wildlife Biologists, Dispatchers, and Road Maintenance Crews  
**Form Factor:** 1-Page Printable / Laminated Visor Card (8.5" × 11" or Vehicle Visor Insert)  
**Field Reporting URL:** `https://corridorwildlife.org/report/`  
**Ranger Portal:** `https://corridorwildlife.org/rangers/`  
**Mitigation & Grant Hub:** `https://corridorwildlife.org/?workspace=mitigate`  
**License:** Apache-2.0 (100% Free Public Good — Zero IT Tickets, Zero Procurement, Zero Per-Seat Fees)

---

```
┌────────────────────────────────────────────────────────────────────────────────────────┐
│                        CORRIDOR FIELD PATROL QUICK REFERENCE                           │
│  1. OPEN: corridorwildlife.org/report/ (Works completely offline in dead zones)        │
│  2. TAP SPECIES -> TAP PHOTO (Optional) -> TAP SUBMIT (<10 seconds total)             │
│  3. AUTO-SYNC: Reports queue locally in IndexedDB; auto-upload when back in range      │
└────────────────────────────────────────────────────────────────────────────────────────┘
```

---

## 1. Zero-Admin Desktop Setup

Install Corridor directly on Government Furnished Equipment (GFE) workstations, ruggedized patrol Toughbooks, iPads, and dispatch stations. **No IT help desk ticket, no administrative elevation (UAC), and no software licenses required.** Operates within standard NIST SP 800-53 / FISMA browser security boundaries.

### Windows 10 / 11 GFE (Patrol Toughbooks & Dispatch Desktops)
*Supported Browsers: Microsoft Edge, Google Chrome*
1. Open `https://corridorwildlife.org` in **Edge** or **Chrome**.
2. Look at the far right of the address URL bar and click the **Install App icon** (🖥️ or `⊕` next to the bookmark star).
   * *Alternative:* Click browser menu `(⋯)` > **Apps** > **Install Corridor**.
3. Confirm **Install**. Corridor launches instantly in a dedicated, distraction-free native window and pins to your Windows **Taskbar** and **Start Menu**.

### macOS & Field Tablets (Safari & Chrome)
*Supported: macOS Sonoma+, iPadOS, iOS, Android Rugged Tablets*
- **macOS Safari:** Open `https://corridorwildlife.org` > Menu bar **File** > **Add to Dock...** > Click **Add**. Launch anytime from your Dock or Spotlight (`Cmd + Space`).
- **Apple iPad / iPhone:** Open Safari > Tap **Share** icon (square with upward arrow) > Scroll and tap **Add to Home Screen** > Tap **Add**.
- **Android / Samsung Active Tough Tablets:** Open Chrome > Tap `(⋮)` menu > Tap **Install app** or **Add to Home screen**.

### Dispatch Console Multi-Monitor Setup
- Launch Corridor PWA on secondary CAD/GIS displays.
- Keep the **Live Incident Feed** (`/v1/stream/reports`) active. Dispatchers receive real-time visual banners and audible chimes the instant field rangers submit carcass observations or animal-on-highway hazards.

### Keyboard Shortcuts (Toughbook & Dispatch Console)
| Shortcut | Action | Operational Context |
| :--- | :--- | :--- |
| `Tab` / `Shift + Tab` | Navigate interactive elements | Fast navigation between species chips & form fields |
| `Space` / `Enter` | Activate selected item | Select active species chip or toggle camera capture |
| `Ctrl + Enter` (Win) / `Cmd + Enter` (Mac) | **Submit Carcass Report** | 1-key instant submit from any form field |
| `Esc` | Close Modal / Drawer | Dismiss safety acknowledgment, install banner, or drawers |
| `Ctrl + Shift + S` / `Cmd + Shift + S` | **Force Manual Sync** | Force push all pending queued reports to server |
| `Ctrl + P` / `Cmd + P` | Print Reference / PDF | Export current report or screen to laminated print |

---

## 2. Offline Field Patrol Logging (<10 Seconds)

Designed for rangers wearing winter duty gloves, parked on high-speed highway shoulders, or operating in deep canyon dead zones with zero cell service.

### The 3-Tap Glove-Friendly Roadside Workflow
```
   [ TAP 1: SPECIES ]        [ TAP 2: PHOTO ]         [ TAP 3: SUBMIT ]
 ┌──────────────────────┐  ┌──────────────────────┐  ┌──────────────────────┐
 │  🫎 Elk       🦌 Mule │  │  📷 Capture Carcass  │  │  ✓ Submit Carcass    │
 │  🫎 Moose    🐐 Prong│  │     Photo (Optional) │  │     Report           │
 │  🐻 Bear     🐾 Other│  │     (Auto-compressed)│  │     (<10s Complete)  │
 └──────────────────────┘  └──────────────────────┘  └──────────────────────┘
```

1. **Tap 1: Select Species Chip**
   - Tap high-contrast, oversized target chip: **Elk** (🫎), **Mule Deer** (🦌), **White-tailed Deer** (🦌), **Moose** (🫎), **Pronghorn** (🐐), **Black Bear** (🐻), **Raccoon** (🦝), **Coyote** (🐺), or **Other** (🐾).
   - On-device vision model automatically assists with species identification when a photo is taken.
2. **Tap 2: Capture Photo (Optional)**
   - Tap large **[📷 Capture Carcass Photo]** button. Camera shutter opens instantly.
   - Snaps specimen, injury type, or vehicle impact point. Image is downsampled locally on-device to conserve memory and battery.
3. **Tap 3: Submit Report**
   - Tap vibrant green **[✓ Submit Carcass Report]** button. Form resets immediately, ready for next observation. Total time: **6 to 8 seconds**.

### Autonomous Highway Intelligence
- **Automatic Milepost Snapping:** Hardware GPS automatically matches coordinate to nearest 0.1-mile highway post marker and highway corridor name (e.g., *US-191 MP 28.4*).
- **Sub-Meter Geolocation:** Continuous background GPS listener locks high-accuracy coordinates (±3m) without requiring manual latitude/longitude entry.

### Dead-Zone Queue Instructions
Field patrols in Yellowstone's Gallatin Canyon, Glacier's Marias Pass, Grand Canyon's North Rim, or Olympic's Hoh River often have zero cellular connectivity. Corridor is engineered specifically for these dead zones:

```
┌────────────────────────────────────────────────────────────────────────────────────────┐
│                              DEAD-ZONE QUEUE PIPELINE                                  │
│                                                                                        │
│  [Patrol in Dead Zone]  ──>  [IndexedDB Client Storage]  ──>  [Auto-Sync on Connect]  │
│  Zero cell signal            Encrypted offline queue          Station Wi-Fi or LTE     │
│  Amber "Offline Mode"        UUID idempotency generated       Flushes queue in ms      │
└────────────────────────────────────────────────────────────────────────────────────────┘
```

1. **Automatic Offline Transition:** When leaving cellular coverage, status pill switches to amber **`Offline Mode`**.
2. **Persistent IndexedDB Storage:** All carcass records, timestamps, GPS coordinates, and photo blobs are saved into encrypted local browser storage (`corridor-field-reports`). Records survive device reboots, browser closures, and battery drains.
3. **Pending Queue Counter:** The header displays active queue status: `Pending: X`.
4. **Autonomous Reconnect Sync:** As soon as patrol vehicle enters LTE/5G coverage or connects to station Wi-Fi, Corridor automatically flushes the queue in the background.
5. **Manual Force Sync:** If immediate upload is required, tap the **`[Sync Now]`** button in the header.

---

## 3. Dual-Resolution Poaching Privacy

To protect vulnerable wildlife from poaching, illegal horn/antler collection, and commercial harassment, Corridor enforces a strict dual-resolution data firewall compliant with the Endangered Species Act and NPS Resource Management directives.

```
                     ┌─────────────────────────────────────────┐
                     │     Ranger Logs Carcass Observation     │
                     └────────────────────┬────────────────────┘
                                          │
                  ┌───────────────────────┴───────────────────────┐
                  ▼                                               ▼
┌───────────────────────────────────┐   ┌───────────────────────────────────┐
│     AUTHENTICATED PARK RANGER     │   │      PUBLIC / RESEARCH PORTAL     │
│       & WILDLIFE BIOLOGISTS       │   │        UNAUTHENTICATED ACCESS     │
├───────────────────────────────────┤   ├───────────────────────────────────┤
│ • Sub-meter GPS Coordinates       │   │ • Coarse H3-Res 6 Hexagons (~36km²)│
│ • Real-time Instant Feed (<1 sec) │   │ • Mandatory 30-Day Delay Window   │
│ • Unredacted Photos & Antler Data │   │ • Redacted Photos (No Trophy Data)│
│ • Full Investigative Patrol Notes │   │ • Generalized Genus-Level Counts  │
│ • Direct Law Enforcement Export   │   │ • Zero Point-Source Coordinates   │
└───────────────────────────────────┘   └───────────────────────────────────┘
```

### Sensitive Species Covered Under Automatic Spatial Obfuscation
- **Carnivores:** Grizzly bear (*Ursus arctos horribilis*), Gray wolf (*Canis lupus*), Canada lynx (*Lynx canadensis*), Wolverine (*Gulo gulo luscus*), Cougar.
- **Trophy & At-Risk Ungulates:** Bighorn sheep (*Ovis canadensis*), Mountain goat, Woodland caribou.
- **Raptors & Scavengers:** Bald eagle, Golden eagle, California condor.

### Privacy Safeguards
1. **Public Hexagonal Obfuscation:** Public maps and open APIs aggregate observations into Uber H3 resolution-6 hexagonal bins (~36 km² / ~14 sq mi). Point coordinates are never exposed to public network traffic.
2. **30-Day Quarantine Delay:** Observations of sensitive species are withheld from public dashboards for 30 calendar days to prevent illegal tracking or opportunistic scavenging of trophy parts.
3. **Law Enforcement & Biology Access:** Verified DOI/NPS accounts retain instantaneous access to unredacted GPS coordinates, high-resolution photo evidence, and full patrol notes.

---

## 4. FHWA Grant Brief Export

Produce publication-grade, two-page grant application briefs for the **$350M FHWA Wildlife Crossings Pilot Program (WCPP, 23 U.S.C. § 171)** in under 30 seconds.

```
┌────────────────────────────────────────────────────────────────────────────────────────┐
│                          1-CLICK FHWA WCPP GRANT EXPORTER                              │
│                                                                                        │
│  [Select Park Corridor] ──> [Set Capital Budget] ──> [Click Export 2-Page Brief (PDF)] │
│   e.g. US-191 Gallatin       e.g. $10,000,000         Full Benefit-Cost Analysis (BCR) │
└────────────────────────────────────────────────────────────────────────────────────────┘
```

### 30-Second Grant Brief Generation Steps
1. Navigate to **Mitigation Hub**: `https://corridorwildlife.org/?workspace=mitigate`.
2. **Select Highway Corridor:** Choose your park corridor (e.g., *Yellowstone US-191 MP 10–35*, *Grand Teton US-89 MP 380–420*, *Glacier US-2 Marias Pass*, *Rocky Mountain US-34*, *Yosemite CA-120*, *Grand Canyon AZ-64*, *Olympic US-101*, or *Shenandoah Skyline Drive*).
3. **Input Available Budget:** Enter project capital target (e.g., `$8,000,000`).
4. **Click `[Export FHWA WCPP Brief (PDF)]`:** The system runs an algorithmic knapsack optimizer in pure Go and downloads a publication-grade, 2-page executive summary ready for immediate submittal to FHWA and State DOT partners.

### What the 2-Page Executive Brief Delivers
- **Page 1: Project Alignment & Collision Hotspots:**
  - Highway corridor profile, milepost extents, and multi-agency crash history.
  - Network Kernel Density Estimation (NKDE) and Getis-Ord Gi* hotspot maps.
  - Threatened/endangered focal species collision mortality records.
- **Page 2: Knapsack-Optimized Engineering Package & Economics:**
  - Recommended structural treatments: wildlife overpass, arch underpasses, 8-foot exclusion fencing, jump-outs/escape ramps, and double Texas wildlife guards.
  - Empirical crash reduction rate (**86% reduction** based on Rytwinski et al. meta-analysis).
  - USDOT Comprehensive Crash Valuations (**$14,000** deer, **$45,000** elk, **$83,000** moose).
  - **Benefit-Cost Ratio (BCR)**, Net Present Value (NPV), and Payback Period over 20-year design life.
  - Pre-filled statutory crosswalk matching 23 U.S.C. § 171 evaluation criteria.

---

## Quick Troubleshooting & Field Verification Tips

| Issue | Quick Fix |
| :--- | :--- |
| **Yellow "Offline Mode" Indicator** | Normal behavior in cellular dead zones. Continue logging normally. All data is queued in local IndexedDB and will auto-upload when you regain signal. |
| **GPS Accuracy Warning (>30m)** | Ensure Toughbook/tablet has clear line-of-sight through vehicle windshield. Tap **[📍 Update GPS]** to refresh satellite fix before submitting. |
| **Pending Queue Not Clearing at Station** | Verify vehicle is connected to station Wi-Fi. Open Corridor app and tap the blue **`[Sync Now]`** button in the top right header. |
| **Accidentally Logged Wrong Species** | Re-open observation in Dispatch feed, or add correction note in the Field Notes box (e.g., *"Correction: Mule deer female, not elk"*). |
| **Camera Won't Open on Tablet** | Verify browser permissions: Tap lock icon 🔒 next to browser URL > Ensure **Camera** and **Location** permissions are toggled to **Allow**. |
| **Never Clear Browser Cookies/Cache on Patrol** | Do not clear browser cache while `Pending: X` shows un-synced reports. Once `Pending: 0`, all data is safely mirrored in the central database. |

---

### Ranger Station Dispatch Emergency Contacts
- **Corridor Web Admin & Technical Support:** `support@corridorwildlife.org`
- **NPS Resource Management Liaison:** `rms@corridorwildlife.org`
- **System Status & Uptime Dashboard:** `https://corridorwildlife.org/health`

*Keep this reference card laminated and clipped to vehicle sun visors or mounted beside patrol Toughbooks.*
