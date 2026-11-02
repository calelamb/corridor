# NPS Ranger SEO, Desktop App Discovery & Outreach Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Transform Corridor into the definitive discoverable desktop and field application for National Park Service (NPS) rangers and federal transportation planners through high-impact SEO, W3C desktop PWA manifest & one-click install flow, a dedicated `/rangers` hub, and targeted outreach packages.

**Architecture:** SvelteKit frontend enhancements embedded in the single Go binary with full W3C Web App Manifest metadata and desktop installation hooks (`beforeinstallprompt`). Structured data (JSON-LD `SoftwareApplication`, `GovernmentApplication`, `FAQPage`) and semantic SEO meta tags on all routes, complemented by a dedicated `/rangers` landing route and ready-to-deploy federal outreach playbooks in `docs/outreach/`.

**Tech Stack:** SvelteKit 5 (runes), TypeScript, W3C Web App Manifest, JSON-LD Schema.org, Go embedded web server (`internal/web`), Vitest, Playwright E2E, Axe accessibility engine.

**Spec:** [NPS Ranger SEO & Desktop App Design Spec](../specs/2026-10-14-nps-ranger-seo-and-desktop-app-design.md)

## Global Constraints

- Svelte 5 runes mode (`$state`, `$props`, `$derived`) for all frontend components.
- Zero CGo dependencies; maintain zero Docker on macOS constraint.
- Glove-friendly interactive controls with minimum 48px x 48px touch targets.
- 100% WCAG 2.2 AA compliance verified via automated Axe checks with 0 violations.
- Clean conventional commits without AI attribution.
- Preserve sensitive species privacy guarantees (H3-r6 + 30-day delay for public views).

## Review Focus

- Web App Manifest desktop installability: manifest must declare standalone display, valid icons, and desktop shortcuts without breaking service worker caching.
- Search indexing coverage: robots.txt and sitemap.xml must accurately index all canonical public routes including `/rangers/`.
- FAQPage and SoftwareApplication JSON-LD schema validity: must pass Google Rich Results test validation with 0 syntax or semantic errors.
- Desktop install prompt dismissal: `beforeinstallprompt` event handling must gracefully handle unsupported browsers (Safari/Firefox) without throwing runtime errors or console warnings.
- Server-side routing redirects: Go HTTP router in `internal/web/handler.go` must cleanly serve and redirect clean URLs (`/rangers` to `/rangers/`) and static assets.

---

### Task 1: W3C Desktop Web App Manifest, App Icons & Install Infrastructure

**Files:**
- Create: `web/static/manifest.webmanifest`
- Create: `web/static/icons/icon-192.svg`
- Create: `web/static/icons/icon-512.svg`
- Modify: `web/src/app.html`
- Modify: `web/src/service-worker.ts`
- Test: `web/tests/manifest.spec.ts`

**Interfaces:**
- Consumes: W3C Web App Manifest specification, Corridor brand tokens (`#102b23`, `#163d31`).
- Produces: Installable desktop PWA manifest linked via `<link rel="manifest" href="/manifest.webmanifest" />`.

- [ ] **Step 1: Write the failing Playwright test for manifest and desktop metadata**

```typescript
// web/tests/manifest.spec.ts
import { test, expect } from '@playwright/test';

test.describe('Desktop Web App Manifest & Icons', () => {
  test('serves valid web manifest with standalone display and desktop shortcuts', async ({ request }) => {
    const response = await request.get('/manifest.webmanifest');
    expect(response.status()).toBe(200);
    expect(response.headers()['content-type']).toContain('application/manifest+json');
    const manifest = await response.json();
    expect(manifest.name).toContain('Corridor');
    expect(manifest.display).toBe('standalone');
    expect(manifest.categories).toContain('government');
    expect(manifest.shortcuts.length).toBeGreaterThanOrEqual(3);
  });
});
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd web && npx playwright test tests/manifest.spec.ts`  
Expected: FAIL with 404 on `/manifest.webmanifest`

- [ ] **Step 3: Implement Web App Manifest and SVG icons**

Create `web/static/manifest.webmanifest` declaring `name`, `short_name`, `display: standalone`, `theme_color`, `background_color`, shortcuts to `/report/`, `/?workspace=mitigate`, `/methodology/`, `/rangers/`, and icons (`web/static/icons/icon-192.svg` and `icon-512.svg`). Link manifest in `web/src/app.html`. Update `service-worker.ts` cache list to include manifest and icons.

- [ ] **Step 4: Run test to verify it passes**

Run: `cd web && npx playwright test tests/manifest.spec.ts`  
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add web/static/manifest.webmanifest web/static/icons/ web/src/app.html web/src/service-worker.ts web/tests/manifest.spec.ts
git commit -m "feat(pwa): add desktop web app manifest, icons, and shortcuts"
```

---

### Task 2: Technical SEO, Structured Data (JSON-LD), Canonical Tags & Sitemap Expansion

**Files:**
- Modify: `web/src/app.html`
- Modify: `web/static/sitemap.xml`
- Modify: `web/static/robots.txt`
- Modify: `web/src/routes/+page.svelte`
- Modify: `web/src/routes/report/+page.svelte`
- Modify: `web/src/routes/methodology/+page.svelte`
- Modify: `web/src/routes/data/+page.svelte`
- Test: `web/tests/seo.spec.ts`

**Interfaces:**
- Consumes: Schema.org `SoftwareApplication`, `GovernmentService`, `FAQPage`, OpenGraph and Twitter Card specs.
- Produces: Search-engine-optimized HTML heads with rich JSON-LD markup and sitemap declarations.

- [ ] **Step 1: Write the failing test for SEO metadata and structured JSON-LD**

```typescript
// web/tests/seo.spec.ts
import { test, expect } from '@playwright/test';

test.describe('Technical SEO & Structured Data', () => {
  test('homepage contains canonical link, open graph, and rich JSON-LD schema', async ({ page }) => {
    await page.goto('/');
    const canonical = await page.locator('link[rel="canonical"]').getAttribute('href');
    expect(canonical).toBe('https://corridorwildlife.org/');

    const jsonLd = await page.locator('script[type="application/ld+json"]').textContent();
    expect(jsonLd).toBeTruthy();
    const data = JSON.parse(jsonLd!);
    expect(data['@type']).toBe('SoftwareApplication');
    expect(data.operatingSystem).toContain('Windows');
    expect(data.offers.price).toBe('0');
  });

  test('sitemap.xml lists all core routes with valid priorities', async ({ request }) => {
    const res = await request.get('/sitemap.xml');
    expect(res.status()).toBe(200);
    const text = await res.text();
    expect(text).toContain('https://corridorwildlife.org/rangers/');
  });
});
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd web && npx playwright test tests/seo.spec.ts`  
Expected: FAIL due to missing canonical link or missing `/rangers/` in sitemap

- [ ] **Step 3: Implement structured data, canonical tags, and updated sitemap**

Update `web/src/app.html` with expanded JSON-LD schema (`SoftwareApplication` with desktop OS support, offers: price $0). Add canonical `<link rel="canonical" href="..." />` and OpenGraph tags to all routes. Add `/rangers/` to `web/static/sitemap.xml` with priority 0.95.

- [ ] **Step 4: Run test to verify it passes**

Run: `cd web && npx playwright test tests/seo.spec.ts`  
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add web/src/app.html web/static/sitemap.xml web/static/robots.txt web/src/routes/ web/tests/seo.spec.ts
git commit -m "feat(seo): enrich JSON-LD schema, canonical tags, and sitemap entries"
```

---

### Task 3: Dedicated NPS Ranger Discovery Hub (`/rangers`)

**Files:**
- Create: `web/src/routes/rangers/+page.svelte`
- Modify: `web/src/routes/+layout.svelte`
- Modify: `internal/web/handler.go`
- Test: `web/tests/rangers.spec.ts`

**Interfaces:**
- Consumes: Design spec Section 4 (NPS Ranger capabilities, FHWA grant briefs, dual-resolution privacy, 10s carcass logging).
- Produces: `/rangers` route with rich content, `FAQPage` JSON-LD schema, direct desktop install triggers, and agency quickstart guidance.

- [ ] **Step 1: Write the failing test for the `/rangers` route**

```typescript
// web/tests/rangers.spec.ts
import { test, expect } from '@playwright/test';

test.describe('NPS Ranger Discovery Hub (/rangers)', () => {
  test('renders ranger hub with core capabilities and FAQ structured data', async ({ page }) => {
    await page.goto('/rangers/');
    await expect(page.locator('h1')).toContainText(/National Park Service|Ranger/i);
    await expect(page.getByText(/100% Free|Public Good/i).first()).toBeVisible();
    await expect(page.getByText(/FHWA/i).first()).toBeVisible();

    const faqScript = page.locator('script[data-schema="faq"]');
    await expect(faqScript).toHaveCount(1);
    const content = await faqScript.textContent();
    const faq = JSON.parse(content!);
    expect(faq['@type']).toBe('FAQPage');
  });
});
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd web && npx playwright test tests/rangers.spec.ts`  
Expected: FAIL with 404 on `/rangers/`

- [ ] **Step 3: Implement `/rangers/+page.svelte` and navigation link**

Create `web/src/routes/rangers/+page.svelte` with:
- Accessible hero section with value propositions tailored for NPS rangers.
- Key features breakdown: 10s offline reporting, unredacted vs generalized privacy, FHWA WCPP PDF exporter, real-time dispatch SSE.
- Park unit spotlights: Yellowstone US-191, Grand Teton US-89/26, Glacier US-2, Rocky Mountain US-34.
- Embedded `FAQPage` JSON-LD structured data for rich snippets.
- Add "NPS Rangers" link in `web/src/routes/+layout.svelte` navigation.
- Ensure `internal/web/handler.go` handles clean URL 301 redirect for `/rangers`.

- [ ] **Step 4: Run test to verify it passes**

Run: `cd web && npx playwright test tests/rangers.spec.ts`  
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add web/src/routes/rangers/ web/src/routes/+layout.svelte internal/web/handler.go web/tests/rangers.spec.ts
git commit -m "feat(rangers): add dedicated NPS ranger discovery hub with FAQ structured data"
```

---

### Task 4: In-App Desktop Install Prompt & Navigation CTA

**Files:**
- Create: `web/src/lib/components/DesktopInstallBanner.svelte`
- Modify: `web/src/routes/+layout.svelte`
- Modify: `web/src/routes/rangers/+page.svelte`
- Test: `web/tests/install-prompt.spec.ts`

**Interfaces:**
- Consumes: W3C `beforeinstallprompt` event, local storage persistence.
- Produces: Accessible install button in header and banner on ranger hub triggering native desktop app installation.

- [ ] **Step 1: Write the failing unit/component test for desktop install banner**

```typescript
// web/tests/install-prompt.spec.ts
import { test, expect } from '@playwright/test';

test.describe('Desktop App Install Flow', () => {
  test('header contains desktop app install CTA or ranger action link', async ({ page }) => {
    await page.goto('/');
    const nav = page.locator('nav[aria-label="Main navigation"]');
    await expect(nav.getByRole('link', { name: /Rangers/i })).toBeVisible();
  });
});
```

- [ ] **Step 2: Run test to verify it fails or passes as baseline**

Run: `cd web && npx playwright test tests/install-prompt.spec.ts`

- [ ] **Step 3: Implement `DesktopInstallBanner.svelte` and integrate into Layout**

Build `DesktopInstallBanner.svelte` using Svelte 5 runes:
- Listens for `beforeinstallprompt`.
- Provides "Install Desktop App" button that calls `.prompt()`.
- Provides fallback instructions for browsers without `beforeinstallprompt` (Chrome/Edge desktop install icon in address bar, Safari "Add to Dock").
- Embeds seamless "Install App" button in the site navigation bar.

- [ ] **Step 4: Run test to verify it passes**

Run: `cd web && npx playwright test tests/install-prompt.spec.ts`  
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add web/src/lib/components/DesktopInstallBanner.svelte web/src/routes/+layout.svelte web/tests/install-prompt.spec.ts
git commit -m "feat(ui): add desktop PWA install banner and navigation button"
```

---

### Task 5: Federal Agency Outreach Package & Email Playbooks

**Files:**
- Create: `docs/outreach/NPS_RANGER_ADOPTION_PLAYBOOK.md`
- Create: `docs/outreach/NPS_EMAIL_TEMPLATES.md`
- Create: `docs/outreach/FHWA_WCPP_PARTNERSHIP_BRIEF.md`
- Test: `tests/outreach_docs_test.go` (verifies outreach documentation integrity, links, and non-empty sections)

**Interfaces:**
- Consumes: Section 5 of Design Spec, verified park highways (Yellowstone US-191, Grand Teton US-89, Glacier US-2).
- Produces: Ready-to-use outbound email templates, agency directory structure, and FHWA grant memo.

- [ ] **Step 1: Write Go test validating outreach documentation integrity**

```go
// tests/outreach_docs_test.go
package tests

import (
	"os"
	"strings"
	"testing"
)

func TestOutreachDocumentation(t *testing.T) {
	files := []string{
		"../docs/outreach/NPS_RANGER_ADOPTION_PLAYBOOK.md",
		"../docs/outreach/NPS_EMAIL_TEMPLATES.md",
		"../docs/outreach/FHWA_WCPP_PARTNERSHIP_BRIEF.md",
	}
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			t.Fatalf("failed to read %s: %v", f, err)
		}
		if len(data) < 500 {
			t.Errorf("file %s is too short: %d bytes", f, len(data))
		}
		content := string(data)
		if !strings.Contains(content, "National Park Service") && !strings.Contains(content, "NPS") {
			t.Errorf("file %s missing NPS reference", f)
		}
	}
}
```

- [ ] **Step 2: Run Go test to verify it fails**

Run: `go test -v ./tests/...`  
Expected: FAIL due to missing files in `docs/outreach/`

- [ ] **Step 3: Author outreach playbook, email templates, and FHWA grant brief**

Author:
1. `docs/outreach/NPS_RANGER_ADOPTION_PLAYBOOK.md` covering park targeting, regional contacts, and adoption workflows.
2. `docs/outreach/NPS_EMAIL_TEMPLATES.md` containing personalized, professional emails for Wildlife Biologists, Chief Rangers, and Superintendents.
3. `docs/outreach/FHWA_WCPP_PARTNERSHIP_BRIEF.md` demonstrating how Corridor satisfies 23 U.S.C. 171 criteria for grant scoring.

- [ ] **Step 4: Run Go test to verify it passes**

Run: `go test -v ./tests/...`  
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add docs/outreach/ tests/outreach_docs_test.go
git commit -m "docs(outreach): add NPS ranger adoption playbook, email templates, and FHWA brief"
```

---

### Task 6: Platform Verification, Accessibility Audit & E2E Validation

**Files:**
- Test: `web/tests/accessibility.spec.ts`
- Test: Full Go and Playwright test suites.

- [ ] **Step 1: Run Svelte check and linters**

Run: `cd web && npm run check && npm run lint`  
Expected: 0 errors, 0 warnings

- [ ] **Step 2: Run frontend unit tests**

Run: `cd web && npm run test:unit`  
Expected: All unit tests pass

- [ ] **Step 3: Run Playwright E2E suite with Axe accessibility audit**

Run: `cd web && npx playwright test`  
Expected: All tests pass including accessibility checks on `/rangers` and `/`

- [ ] **Step 4: Run Go backend test suite**

Run: `go test -v ./...`  
Expected: All packages pass

- [ ] **Step 5: Update `docs/PROGRESS.md` and commit**

```bash
git add docs/PROGRESS.md
git commit -m "chore: record completion of NPS ranger SEO, desktop app, and outreach increment"
```
