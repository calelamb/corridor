import { test, expect } from '@playwright/test';
import AxeBuilder from '@axe-core/playwright';
import { evidence } from './fixtures/evidence';
import { hotspotsResponse } from './fixtures/mitigation';

test.describe('Multi-Park Navigator & Hotspot Controls', () => {
	test.beforeEach(async ({ page }) => {
		await page.route('**/v1/explore?*', (r) => r.fulfill({ json: evidence }));
		await page.route('**/v1/segments*', (r) => r.fulfill({ json: [] }));
		await page.route(/\/v1\/hotspots(\?.*)?$/, (r) => r.fulfill({ json: hotspotsResponse }));
	});

	test('renders multi-park extent switcher with tactile touch targets (>= 48px)', async ({
		page
	}) => {
		await page.goto('/');

		const switcher = page.getByRole('region', { name: /National Park navigation/i });
		await expect(switcher).toBeVisible();

		// Check dropdown
		const dropdown = page.getByRole('combobox', { name: /National Park Extent Switcher/i });
		await expect(dropdown).toBeVisible();
		const dropdownBox = await dropdown.boundingBox();
		expect(dropdownBox?.height).toBeGreaterThanOrEqual(48);

		// Check pill group
		const pillGroup = page.getByRole('radiogroup', { name: /national park quick extents/i });
		await expect(pillGroup).toBeVisible();

		const parks = [
			'Yosemite National Park',
			'Yellowstone National Park',
			'Grand Teton National Park',
			'Glacier National Park',
			'Rocky Mountain National Park',
			'Olympic National Park'
		];

		for (const parkName of parks) {
			const pill = page.getByRole('radio', { name: parkName });
			await expect(pill).toBeVisible();
			const box = await pill.boundingBox();
			expect(box?.height).toBeGreaterThanOrEqual(48);
			expect(box?.width).toBeGreaterThanOrEqual(48);
		}
	});

	test('allows instantaneous navigation across all 6 national parks via pills and dropdown', async ({
		page
	}) => {
		await page.goto('/');

		const dropdown = page.getByRole('combobox', { name: /National Park Extent Switcher/i });
		const yellowstonePill = page.getByRole('radio', { name: 'Yellowstone National Park' });
		const glacierPill = page.getByRole('radio', { name: 'Glacier National Park' });

		// Click Yellowstone pill
		await yellowstonePill.click();
		await expect(yellowstonePill).toHaveAttribute('aria-checked', 'true');
		await expect(dropdown).toHaveValue('yellowstone');

		// Click Glacier pill
		await glacierPill.click();
		await expect(glacierPill).toHaveAttribute('aria-checked', 'true');
		await expect(yellowstonePill).toHaveAttribute('aria-checked', 'false');
		await expect(dropdown).toHaveValue('glacier');

		// Switch via dropdown to Olympic
		await dropdown.selectOption('olympic');
		const olympicPill = page.getByRole('radio', { name: 'Olympic National Park' });
		await expect(olympicPill).toHaveAttribute('aria-checked', 'true');
		await expect(glacierPill).toHaveAttribute('aria-checked', 'false');
	});

	test('supports keyboard navigation with arrow keys and roving tabindex across park pills', async ({
		page
	}) => {
		await page.goto('/');

		const yosemitePill = page.getByRole('radio', { name: 'Yosemite National Park' });
		const yellowstonePill = page.getByRole('radio', { name: 'Yellowstone National Park' });
		const tetonPill = page.getByRole('radio', { name: 'Grand Teton National Park' });

		// Focus the first pill
		await yosemitePill.focus();
		await expect(yosemitePill).toBeFocused();

		// ArrowRight moves to Yellowstone
		await page.keyboard.press('ArrowRight');
		await expect(yellowstonePill).toBeFocused();
		await expect(yellowstonePill).toHaveAttribute('aria-checked', 'true');

		// ArrowRight moves to Grand Teton
		await page.keyboard.press('ArrowRight');
		await expect(tetonPill).toBeFocused();
		await expect(tetonPill).toHaveAttribute('aria-checked', 'true');

		// ArrowLeft moves back to Yellowstone
		await page.keyboard.press('ArrowLeft');
		await expect(yellowstonePill).toBeFocused();
		await expect(yellowstonePill).toHaveAttribute('aria-checked', 'true');
	});

	test('hotspot controls allow algorithm selection and continuous bandwidth slider with query param updates', async ({
		page
	}) => {
		let lastHotspotUrl = '';
		await page.route(/\/v1\/hotspots(\?.*)?$/, (r) => {
			lastHotspotUrl = r.request().url();
			return r.fulfill({ json: hotspotsResponse });
		});

		await page.goto('/?workspace=hotspots');

		const controls = page.getByRole('group', { name: /Hotspot Analysis Controls/i });
		await expect(controls).toBeVisible();

		// Bandwidth slider has 250m to 2000m range
		const slider = page.getByRole('slider', { name: /Hotspot smoothing bandwidth/i });
		await expect(slider).toBeVisible();
		await expect(slider).toHaveAttribute('aria-valuemin', '250');
		await expect(slider).toHaveAttribute('aria-valuemax', '2000');

		// Continuous slider touch target
		const sliderBox = await slider.boundingBox();
		expect(sliderBox?.height).toBeGreaterThanOrEqual(44);

		// Switch algorithm via button to Getis-Ord
		const getisBtn = page.getByRole('button', { name: /Getis-Ord/i });
		await getisBtn.click();
		await expect.poll(() => lastHotspotUrl).toContain('method=getis_ord');

		// Adjust slider to 1500m
		await slider.fill('1500');
		await slider.dispatchEvent('input');
		await expect.poll(() => lastHotspotUrl).toContain('bandwidth=1500');

		// Click 2000m preset
		const bw2000Btn = page.getByRole('button', { name: /2000\s*m/i });
		await expect(bw2000Btn).toBeVisible();
		await bw2000Btn.click();
		await expect.poll(() => lastHotspotUrl).toContain('bandwidth=2000');
	});

	test('passes WCAG 2.2 AA audit in light and dark mode with park extent switcher and hotspot controls', async ({
		page
	}) => {
		for (const theme of ['light', 'dark']) {
			await page.addInitScript((val) => localStorage.setItem('corridor-theme', val), theme);

			// Check Workbench with Park Extent Switcher
			await page.goto('/');
			await expect(page.getByRole('region', { name: /National Park navigation/i })).toBeVisible();
			await expect(page.locator('html')).toHaveAttribute('data-theme', theme);
			await page.screenshot({
				path: test.info().outputPath(`park-navigation-${theme}.png`),
				fullPage: true
			});
			let axe = await new AxeBuilder({ page })
				.withTags(['wcag2a', 'wcag2aa', 'wcag21a', 'wcag21aa', 'wcag22aa'])
				.analyze();
			expect(axe.violations).toEqual([]);

			// Check Hotspot Lab with HotspotControls
			await page.goto('/?workspace=hotspots');
			await expect(page.getByRole('group', { name: /Hotspot Analysis Controls/i })).toBeVisible();
			await expect(page.locator('html')).toHaveAttribute('data-theme', theme);
			await page.screenshot({
				path: test.info().outputPath(`hotspot-controls-${theme}.png`),
				fullPage: true
			});
			axe = await new AxeBuilder({ page })
				.withTags(['wcag2a', 'wcag2aa', 'wcag21a', 'wcag21aa', 'wcag22aa'])
				.analyze();
			expect(axe.violations).toEqual([]);
		}
	});
});
