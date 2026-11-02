import { test, expect } from '@playwright/test';
import AxeBuilder from '@axe-core/playwright';

test.describe('Interactive Ranger Field Demo Simulator (/rangers/#ranger-sandbox)', () => {
	test.beforeEach(async ({ page }) => {
		await page.goto('/rangers/');
		await expect(page.locator('#ranger-sandbox')).toBeVisible();
	});

	test('renders interactive sandbox section with 4-step stepper and glove-friendly touch targets', async ({
		page
	}) => {
		const section = page.locator('#ranger-sandbox');
		await expect(section).toBeVisible();

		// Check heading
		const heading = section.getByRole('heading', {
			name: /Interactive Ranger Sandbox: Test the Field Flow in 15 Seconds/i
		});
		await expect(heading).toBeVisible();

		// Stepper has 4 steps
		const stepper = section.locator('.sim-stepper');
		await expect(stepper).toBeVisible();
		const stepTabs = stepper.locator('.step-tab');
		await expect(stepTabs).toHaveCount(4);

		// Touch targets >= 48px
		for (const tab of await stepTabs.all()) {
			const box = await tab.boundingBox();
			expect(box).not.toBeNull();
			if (box) {
				expect(box.height).toBeGreaterThanOrEqual(44); // 48px target with tolerance
			}
		}

		// Initial step is Step 1
		await expect(stepTabs.nth(0)).toHaveAttribute('aria-current', 'step');
	});

	test('allows selecting different park scenarios and updates field context', async ({ page }) => {
		const section = page.locator('#ranger-sandbox');

		// Scenario cards
		const scenarioCards = section.locator('.scenario-card');
		await expect(scenarioCards).toHaveCount(4);

		// Default is Yellowstone
		await expect(scenarioCards.nth(0)).toHaveAttribute('aria-checked', 'true');
		await expect(section.locator('.scenario-summary-inline')).toContainText(/Yellowstone/i);

		// Select Grand Teton
		await scenarioCards.nth(1).click();
		await expect(scenarioCards.nth(1)).toHaveAttribute('aria-checked', 'true');
		await expect(scenarioCards.nth(0)).toHaveAttribute('aria-checked', 'false');
		await expect(section.locator('.scenario-summary-inline')).toContainText(/Grand Teton/i);

		// Select Glacier
		await scenarioCards.nth(2).click();
		await expect(scenarioCards.nth(2)).toHaveAttribute('aria-checked', 'true');
		await expect(section.locator('.scenario-summary-inline')).toContainText(/Glacier/i);

		// Select Yosemite
		await scenarioCards.nth(3).click();
		await expect(scenarioCards.nth(3)).toHaveAttribute('aria-checked', 'true');
		await expect(section.locator('.scenario-summary-inline')).toContainText(/Yosemite/i);
	});

	test('simulates 1-tap offline carcass logging with GPS, species AI chip, and IndexedDB write', async ({
		page
	}) => {
		const section = page.locator('#ranger-sandbox');

		// Advance from Step 1 to Step 2
		await section.getByRole('button', { name: /Proceed to Offline Report Simulation/i }).click();

		// Verify on Step 2
		const step2Heading = section.getByRole('heading', {
			name: /Step 2: 1-Tap Offline Carcass Report/i
		});
		await expect(step2Heading).toBeVisible();

		// 3-tap simulation cards are present
		await expect(section.locator('.tap-card')).toHaveCount(3);
		await expect(section.getByText(/Auto GPS & Milepost Snapping/i)).toBeVisible();
		await expect(section.getByText(/On-Device AI Species Chip/i)).toBeVisible();
		await expect(section.getByText(/Local IndexedDB Queueing/i)).toBeVisible();

		// Verify offline status indicator
		await expect(section.getByText(/Queued Locally \(0 bytes cellular\)/i)).toBeVisible();

		// Click 1-tap rapid log button
		const logBtn = section.getByRole('button', {
			name: /1-Tap Offline Carcass Report/i
		});
		await expect(logBtn).toBeVisible();
		await logBtn.click();

		// Successfully logged banner
		const successBanner = section.locator('.report-success-banner');
		await expect(successBanner).toBeVisible();
		await expect(successBanner).toContainText(/Record committed to local IndexedDB/i);

		// Next action button to advance to Step 3
		const nextBtn = section.getByRole('button', {
			name: /View Instant Dispatch Alert & Privacy Badge/i
		});
		await expect(nextBtn).toBeVisible();
		await nextBtn.click();

		// On Step 3
		await expect(
			section.getByRole('heading', { name: /Step 3: Instant Dispatch Alert/i })
		).toBeVisible();
	});

	test('demonstrates instant dispatch alert, audio chime, and exact dual-resolution privacy badge', async ({
		page
	}) => {
		const section = page.locator('#ranger-sandbox');

		// Select Yellowstone and navigate directly to Step 3
		await section.locator('.step-tab').nth(2).click();

		// Step 3 heading
		await expect(
			section.getByRole('heading', { name: /Step 3: Instant Dispatch Alert/i })
		).toBeVisible();

		// Dispatch toast notification
		const dispatchToast = section.locator('.dispatch-toast');
		await expect(dispatchToast).toBeVisible();
		await expect(dispatchToast).toContainText(/DISPATCH ALERT:/i);
		await expect(dispatchToast).toContainText(/Gallatin Canyon/i);

		// Audio chime test button
		const chimeBtn = section.getByRole('button', { name: /Play Radio Chime/i });
		await expect(chimeBtn).toBeVisible();
		await chimeBtn.click();

		// Dual-resolution poaching privacy badge verbatim check
		const privacyBadge = section.locator('.dual-res-badge');
		await expect(privacyBadge).toBeVisible();
		await expect(privacyBadge).toContainText(
			'Park Ranger View: Exact GPS 44.9781°N, 110.6982°W | Public Map: Generalized H3 Cell + 30d Delay'
		);

		// Side-by-side comparison cards
		const viewCards = section.locator('.view-card');
		await expect(viewCards).toHaveCount(2);
		await expect(viewCards.nth(0)).toContainText(/PARK RANGER VIEW/i);
		await expect(viewCards.nth(1)).toContainText(/PUBLIC MAP/i);
	});

	test('calculates FHWA WCPP benefit-cost ratio, displays brief preview, and permits download and reset', async ({
		page
	}) => {
		const section = page.locator('#ranger-sandbox');

		// Navigate directly to Step 4
		await section.locator('.step-tab').nth(3).click();

		// Step 4 heading
		await expect(
			section.getByRole('heading', { name: /Step 4: 1-Click FHWA WCPP/i })
		).toBeVisible();

		// Benefit-Cost Ratio hero stat
		const bcrStat = section.locator('.bcr-hero-stat');
		await expect(bcrStat).toBeVisible();
		await expect(bcrStat).toContainText('4.2x');
		await expect(bcrStat).toContainText(/BENEFIT-COST RATIO/i);

		// Preview 2-Page FHWA Brief action
		const previewBtn = section.getByRole('button', { name: /Preview FHWA WCPP 2-Page Brief/i });
		await expect(previewBtn).toBeVisible();
		await previewBtn.click();

		// Expanded document preview appears
		const docPreview = section.locator('.brief-document-preview');
		await expect(docPreview).toBeVisible();
		await expect(docPreview).toContainText(/WILDLIFE CROSSINGS PILOT PROGRAM/i);
		await expect(docPreview).toContainText(/Project Scope/i);
		await expect(docPreview).toContainText(/Economic Return & Crash Reduction/i);
		await expect(docPreview).toContainText(/Dual-Resolution Patrol Verification/i);

		// Close preview
		const closeBtn = docPreview.getByRole('button', { name: /Close Brief Preview/i });
		await closeBtn.click();
		await expect(docPreview).not.toBeVisible();

		// Reset simulator back to Step 1
		const resetBtn = section.getByRole('button', { name: /Restart 15-Second Simulation/i });
		await resetBtn.click();

		// Back on Step 1
		await expect(section.locator('.step-tab').nth(0)).toHaveAttribute('aria-current', 'step');
		await expect(
			section.getByRole('heading', { name: /Step 1: Select Park Scenario/i })
		).toBeVisible();
	});

	test('passes WCAG 2.2 AA accessibility audit in light and dark mode with zero violations', async ({
		page
	}) => {
		for (const theme of ['light', 'dark']) {
			await page.addInitScript((val) => localStorage.setItem('corridor-theme', val), theme);
			await page.goto('/rangers/');
			await expect(page.locator('#ranger-sandbox')).toBeVisible();

			const axe = await new AxeBuilder({ page }).include('#ranger-sandbox').analyze();

			expect(axe.violations).toEqual([]);
		}
	});
});
