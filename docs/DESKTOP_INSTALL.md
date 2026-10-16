# Corridor Desktop & GFE Workstation Installation Guide
## Zero-Admin Deployment Guide for Park IT Specialists, Resource Managers, and Field Rangers

**Document ID:** `CORRIDOR-DESKTOP-INSTALL-V1`  
**Target Audience:** National Park Service (NPS) Rangers, Wildlife Biologists, GIS Technicians, Dispatchers, and Park IT Specialists  
**Operating Systems:** Windows 10/11 Government Furnished Equipment (GFE), macOS 11+, Linux (Ubuntu/RHEL/Debian), iPadOS/Android  
**License:** Apache-2.0 (100% Free Public Good — Zero IT Tickets, Zero Procurement, Zero Per-Seat Licensing)  
**Production URL:** `https://corridorwildlife.org`  

---

## 1. Executive Overview

Corridor is engineered as an offline-first Progressive Web Application (PWA) with native OS desktop integration. For federal and state agency personnel—including the National Park Service (NPS), U.S. Forest Service (USFS), Bureau of Land Management (BLM), and State Departments of Transportation—Corridor installs and runs with **strictly zero administrator privileges, zero User Account Control (UAC) prompts, and zero installation of unverified binary executables**.

By executing directly inside enterprise-approved, sandboxed browser environments (Microsoft Edge on Windows GFE, Google Chrome, or Apple Safari), Corridor delivers a native desktop window experience while maintaining full compliance with federal cybersecurity directives, including **NIST SP 800-53 Rev. 5**, **FISMA**, and **Section 508 / WCAG 2.2 AA**.

```
┌────────────────────────────────────────────────────────────────────────────────────────┐
│                        ZERO-ADMIN GFE DEPLOYMENT ARCHITECTURE                          │
│                                                                                        │
│  Standard GFE User Profile        Enterprise Sandboxed Process       Secure Local Data │
│  ┌───────────────────────┐        ┌───────────────────────────┐      ┌───────────────┐ │
│  │ ~/Desktop/Corridor    │───────>│ msedge.exe / Chrome       │─────>│ IndexedDB     │ │
│  │ Start Menu / Programs │        │ Mode: --app=https://...   │      │ (Dead-zone    │ │
│  │ User-space shortcut   │        │ Chromium OS Sandbox       │      │  report queue)│ │
│  └───────────────────────┘        └───────────────────────────┘      └───────────────┘ │
│                                                                                        │
│  [✓] Zero Admin / Sudo       [✓] No MSI / EXE Payloads      [✓] NIST SP 800-53 Compliant│
└────────────────────────────────────────────────────────────────────────────────────────┘
```

---

## 2. 1-Minute Automated Desktop Installation

For rapid deployment across dispatch consoles, ranger station workstations, and field Toughbooks, Corridor provides pre-packaged zero-admin deployment scripts.

### 2.1. Windows 10 / 11 GFE (Microsoft Edge / Google Chrome)

On any standard GFE workstation running Windows 10 or Windows 11:

1. Open standard PowerShell (no "Run as Administrator" needed):
   ```powershell
   # From local repository clone or flash drive:
   powershell -ExecutionPolicy Bypass -File .\scripts\Install-Corridor.ps1
   ```

2. **What this automated script does:**
   - Detects the enterprise-approved browser (probes native `msedge.exe`, followed by `chrome.exe`).
   - Caches the high-resolution Corridor emblem icon in `$env:LOCALAPPDATA\Corridor\corridor.ico`.
   - Creates a native desktop shortcut at `$HOME\Desktop\Corridor.lnk`.
   - Creates a native Start Menu application launcher at `$env:APPDATA\Microsoft\Windows\Start Menu\Programs\Corridor.lnk`.
   - Sets the target to `--app=https://corridorwildlife.org`, giving Corridor its own distraction-free window, taskbar icon, and native window controls.

3. **Clean Uninstallation:**
   ```powershell
   powershell -ExecutionPolicy Bypass -File .\scripts\Install-Corridor.ps1 -Uninstall
   ```

---

### 2.2. macOS Workstations & Field MacBooks (Safari / Chrome / Edge)

On macOS Big Sur (11.0) through macOS Sonoma / Sequoia:

1. Open Terminal and execute:
   ```bash
   # From local repository clone:
   ./scripts/install-desktop.sh
   ```

2. **What this automated script does:**
   - Creates a native application bundle in your user folder: `~/Applications/Corridor.app`.
   - Automatically builds an Apple ICNS multi-resolution icon using native macOS image utilities (`qlmanage`, `sips`, and `iconutil`).
   - Configures the app launcher to open Corridor in standalone frameless app window mode via Google Chrome, Microsoft Edge, Brave, or Safari.
   - Creates a desktop shortcut at `~/Desktop/Corridor.app`.
   - Indexes Corridor in Spotlight (`Cmd + Space` -> Type "Corridor") and Launchpad.

3. **Clean Uninstallation:**
   ```bash
   ./scripts/install-desktop.sh --uninstall
   ```

---

### 2.3. Linux Workstations (Debian, Ubuntu, Fedora, RHEL)

For GIS workstations running Linux:

1. Open terminal and run:
   ```bash
   ./scripts/install-desktop.sh
   ```

2. **What this automated script does:**
   - Installs user-space launcher command to `~/.local/bin/corridor`.
   - Installs high-resolution vector icon to `~/.local/share/icons/hicolor/scalable/apps/corridor.svg`.
   - Registers standard XDG Freedesktop launcher at `~/.local/share/applications/corridor.desktop` with categories `Utility;Science;Geography;`.
   - Configures native Quick Actions for **New Field Carcass Report**, **Mitigation Planner**, **Hotspot Lab**, and **NPS Ranger Hub**.

3. **Clean Uninstallation:**
   ```bash
   ./scripts/install-desktop.sh --uninstall
   ```

---

## 3. 1-Minute Manual Browser Installation (No Terminal Required)

If command-line tools are restricted or preferred by non-technical rangers, Corridor can be installed in under 60 seconds directly through the browser graphical user interface.

### 3.1. Windows GFE via Microsoft Edge (Recommended for Park Toughbooks)
1. Launch **Microsoft Edge** and navigate to `https://corridorwildlife.org`.
2. Look at the right-hand side of the address URL bar. Click the **App Available** icon (a small computer monitor with a plus sign `⊕` or down arrow).
3. In the pop-up modal, click **Install**.
4. Check the options to **Pin to taskbar** and **Pin to Start**, then click **Allow**.
5. Corridor opens immediately as an isolated desktop application.

### 3.2. Google Chrome (Windows, macOS, Linux)
1. Open **Google Chrome** and visit `https://corridorwildlife.org`.
2. In the right side of the omnibox address bar, click the **Install Corridor** icon (`⊕`).
   * *Alternative:* Click the Chrome menu `(⋮)` in the top-right corner > **Save and share** (or **Cast, save, and share**) > **Install Corridor...**
3. Click **Install**. Corridor will appear in your OS taskbar, Start Menu, or Applications folder.

### 3.3. macOS Safari (macOS Sonoma 14+ "Add to Dock")
1. Open **Safari** and go to `https://corridorwildlife.org`.
2. In the top Apple menu bar, select **File** > **Add to Dock...**.
3. Confirm the application title ("Corridor") and icon, then click **Add**.
4. Launch Corridor at any time directly from the macOS Dock or Spotlight search.

### 3.4. Ruggedized Field Tablets (iPadOS & Android Toughpads)
- **Apple iPad / iPad Mini (Patrol Vehicles):**
  1. Open Safari at `https://corridorwildlife.org/report/`.
  2. Tap the **Share** button (box with upward arrow) in the toolbar.
  3. Scroll down and tap **Add to Home Screen**.
  4. Tap **Add**. The Corridor wildlife crossing icon appears on the iPad home screen.
- **Samsung Galaxy Active / Android Rugged Tablets:**
  1. Open Chrome at `https://corridorwildlife.org/report/`.
  2. Tap the three dots menu `(⋮)` > tap **Install app** or **Add to Home screen**.
  3. Confirm installation. The app launches full-screen without browser URL bars.

---

## 4. Offline Usage & Field Persistence Architecture

National parks frequently feature severe cellular dead zones along mountain passes, canyon corridors, and remote boundaries (e.g., US-191 in Yellowstone, US-2 in Glacier, CA-120 Tioga Pass in Yosemite). Corridor is designed specifically to operate continuously in zero-connectivity environments.

### 4.1. Local Cache Persistence
- **ServiceWorker Caching:** Upon initial visit, Corridor's W3C ServiceWorker caches the complete application bundle, styling, UI components, raster map tile basemaps, and machine learning client classifiers in the browser's protected `CacheStorage`.
- **IndexedDB Dead-Zone Queue:**
  - Carcass observations, GPS coordinates, species classifications, and collision evidence photos are immediately serialized into an encrypted browser IndexedDB queue (`field_reports_offline`).
  - Reports receive a cryptographically generated UUID idempotency key before storage.
  - The offline queue indicator in the navigation bar displays pending reports count (e.g., `⚡ 3 pending reports queued`).
- **Automatic Re-sync Upon Connectivity:**
  - When the patrol vehicle returns to ranger station Wi-Fi or cellular service, the ServiceWorker automatically flushes the queue via background HTTP POST transactions.
  - Rangers can also force an immediate manual synchronization by pressing `Ctrl + Shift + S` (Windows) or `Cmd + Shift + S` (macOS).

### 4.2. Recommended 10-Second Patrol Workflow
```
[Arrive at Carcass] 
       │
       ▼
[Open Desktop / Home Screen Shortcut] ──> (Loads offline in <1 second)
       │
       ▼
[Tap Species Chip] ──────────────────────> (e.g. Elk, Mule Deer, Pronghorn)
       │
       ▼
[Optional: Tap Camera Capture] ──────────> (EXIF coordinates auto-extracted)
       │
       ▼
[Press Ctrl+Enter / Tap Submit] ─────────> (<10 seconds total time on shoulder)
       │
       ▼
[Safely Resume Patrol] ──────────────────> (Auto-syncs upon entering network range)
```

---

## 5. Federal Cybersecurity & FISMA Compliance Notes

Park IT Specialists and Information System Security Officers (ISSOs) are frequently asked whether Corridor requires an Authority to Operate (ATO), local administrative rights, or specialized firewall exceptions.

### 5.1. Zero Installation Footprint
Corridor creates **zero kernel drivers, zero Windows services, zero background daemons, zero scheduled tasks, and zero modifications to system-wide registries (`HKEY_LOCAL_MACHINE`)**. All desktop integration occurs purely within the logged-in ranger's personal user profile (`%LOCALAPPDATA%`, `%APPDATA%`, `~/Applications`).

### 5.2. Sandboxed Process Boundary
Corridor runs entirely inside the native Chromium / WebKit security sandbox already certified and deployed across all GFE machines:
- **Process Isolation:** The browser process executes with restricted OS tokens and integrity levels (`Low Integrity` on Windows).
- **System Call Filtering:** Chromium seccomp-bpf filters and Windows AppContainer rules prevent arbitrary system access.
- **No Binary Downloads:** No third-party `.exe`, `.msi`, `.dll`, or `.dylib` binary files are downloaded or executed.

### 5.3. Data Protection at Rest & In Transit
- **Data in Transit:** All communications with Corridor APIs enforce TLS 1.3 with strict HTTP Strict Transport Security (HSTS).
- **Data at Rest:** Offline IndexedDB databases and cached assets reside in the user's OS-encrypted browser profile directory (protected by Windows BitLocker / DPAPI and macOS FileVault).
- **Poaching Protection Privacy:** As required by federal wildlife protection standards, public incident feeds and exports automatically generalize sensitive species locations to H3 resolution 6 (~36 km²) and enforce a mandatory 30-day reporting delay. Only authenticated rangers with active credentials access unredacted high-precision coordinates.

### 5.4. NIST SP 800-53 Rev. 5 Security Controls Mapping

| Control ID | Control Name | Corridor Technical Implementation |
| :--- | :--- | :--- |
| **AC-6** | Least Privilege | Strictly zero administrator privileges or UAC elevation required. Installs in user-space. |
| **CM-7** | Least Functionality | No extraneous services or daemons; executes solely when launched by user in browser sandbox. |
| **SC-18** | Mobile Code | Governed by enterprise browser security configurations; sandbox prevents unauthorized file system access. |
| **SC-28** | Protection at Rest | Local report queue protected by user profile isolation and OS disk encryption (BitLocker/FileVault). |
| **IA-2 / IA-8** | Identification & Authentication | Seamlessly compatible with federal PIV / CAC card authentication and Entra ID / Okta SAML/OIDC SSO. |
| **CA-3** | System Interconnections | Standard outbound HTTPS (TCP 443) only; zero proprietary ports or incoming listening sockets. |

---

## 6. Dispatch Desk & Multi-Monitor Deployment

For emergency dispatch consoles (CAD) and park communications centers:

1. Launch Corridor in standalone mode.
2. Drag the Corridor window onto secondary or tertiary monitoring displays.
3. Open the **Live Incident Dispatch Stream** (`/v1/stream/reports`).
4. Dispatchers receive real-time Server-Sent Events (SSE) showing live carcass hazard submissions with audible hazard chimes and visual alert banners.
5. Road maintenance and patrol units can be immediately dispatched to remove carcass hazards before secondary vehicular collisions occur.

---

## 7. Keyboard Shortcuts Reference Card

For rapid operation on patrol vehicle Toughbooks and dispatch consoles:

| Shortcut (Windows) | Shortcut (macOS) | Action / Function |
| :--- | :--- | :--- |
| `Tab` / `Shift + Tab` | `Tab` / `Shift + Tab` | Cycle forward / backward through species chips and fields |
| `Space` / `Enter` | `Space` / `Enter` | Select highlighted species or activate camera capture |
| `Ctrl + Enter` | `Cmd + Enter` | **Instant Carcass Report Submit** (from any field) |
| `Ctrl + Shift + S` | `Cmd + Shift + S` | **Force Manual Queue Synchronization** to server |
| `Esc` | `Esc` | Close open drawers, modals, or safety dialogs |
| `Ctrl + P` | `Cmd + P` | Print FHWA-ready mitigation brief or laminated report |

---

## 8. Park IT Specialist Support & Assistance

For technical deployment questions, GPO packaging assistance, or integration with state DOT wildlife collision databases:
- **Engineering Repository:** `https://github.com/corridor/corridor`
- **NPS Ranger Hub:** `https://corridorwildlife.org/rangers/`
- **Methodology & Evidence Library:** `https://corridorwildlife.org/methodology`
- **FHWA WCPP Partnership Brief:** `docs/outreach/FHWA_WCPP_PARTNERSHIP_BRIEF.md`
