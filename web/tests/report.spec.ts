import { test, expect } from '@playwright/test';
import AxeBuilder from '@axe-core/playwright';

test.describe('Park Ranger Field Reporting PWA', () => {
	test('passes WCAG 2.2 AA accessibility audit with and without safety modal', async ({ page }) => {
		await page.goto('/report/');
		const modalAudit = await new AxeBuilder({ page }).analyze();
		expect(modalAudit.violations).toEqual([]);

		await page.getByRole('button', { name: /Passenger|Parked Safely|Acknowledge/i }).click();
		const formAudit = await new AxeBuilder({ page }).analyze();
		expect(formAudit.violations).toEqual([]);
	});
	test('driver safety modal appears on initial launch and can be dismissed', async ({ page }) => {
		await page.goto('/report/');

		// Modal should be visible on initial visit
		const modal = page.getByRole('dialog', { name: /Driver Safety|Do Not Use While Driving/i });
		await expect(modal).toBeVisible();
		await expect(page.getByText(/Do Not Use While Driving|park vehicles/i)).toBeVisible();

		// Dismiss modal
		const ackButton = page.getByRole('button', { name: /Passenger|Parked Safely|Acknowledge/i });
		await expect(ackButton).toBeVisible();
		await ackButton.click();

		// Modal should now be closed
		await expect(modal).not.toBeVisible();

		// On reload in same session, modal should remain dismissed
		await page.reload();
		await expect(
			page.getByRole('dialog', { name: /Driver Safety|Do Not Use While Driving/i })
		).not.toBeVisible();
	});

	test('all primary touch targets measure >= 48px x 48px for glove-friendly operation', async ({
		page
	}) => {
		await page.goto('/report/');

		// Dismiss safety modal first to inspect form
		const ackButton = page.getByRole('button', { name: /Passenger|Parked Safely|Acknowledge/i });
		const ackBox = await ackButton.boundingBox();
		expect(ackBox).not.toBeNull();
		if (ackBox) {
			expect(ackBox.width).toBeGreaterThanOrEqual(48);
			expect(ackBox.height).toBeGreaterThanOrEqual(48);
		}
		await ackButton.click();

		// Check species chips touch targets
		const speciesButtons = page.locator('button[data-species-chip]');
		const count = await speciesButtons.count();
		expect(count).toBeGreaterThanOrEqual(8);

		for (let i = 0; i < count; i++) {
			const box = await speciesButtons.nth(i).boundingBox();
			expect(box).not.toBeNull();
			if (box) {
				expect(box.width).toBeGreaterThanOrEqual(48);
				expect(box.height).toBeGreaterThanOrEqual(48);
			}
		}

		// Check photo capture button
		const photoBtn = page.locator('.btn-photo');
		const photoBox = await photoBtn.boundingBox();
		expect(photoBox).not.toBeNull();
		if (photoBox) {
			expect(photoBox.width).toBeGreaterThanOrEqual(48);
			expect(photoBox.height).toBeGreaterThanOrEqual(48);
		}

		// Check submit button
		const submitBtn = page.getByRole('button', { name: /Submit Carcass Report/i });
		const submitBox = await submitBtn.boundingBox();
		expect(submitBox).not.toBeNull();
		if (submitBox) {
			expect(submitBox.width).toBeGreaterThanOrEqual(48);
			expect(submitBox.height).toBeGreaterThanOrEqual(48);
		}

		// Check GPS update button
		const gpsBtn = page.getByRole('button', { name: /GPS|Location/i });
		const gpsBox = await gpsBtn.boundingBox();
		expect(gpsBox).not.toBeNull();
		if (gpsBox) {
			expect(gpsBox.width).toBeGreaterThanOrEqual(48);
			expect(gpsBox.height).toBeGreaterThanOrEqual(48);
		}
	});

	test('3-tap rapid carcass reporting workflow selects species and submits', async ({ page }) => {
		let reportPayload: Record<string, unknown> | null = null;
		await page.route('**/v1/reports', async (route) => {
			if (route.request().method() === 'POST') {
				reportPayload = route.request().postDataJSON();
				await route.fulfill({
					status: 201,
					contentType: 'application/json',
					body: JSON.stringify({
						status: 'success',
						data: { id: reportPayload?.id || 'mock-report-id' }
					})
				});
			} else {
				await route.continue();
			}
		});

		await page.goto('/report/');

		// Dismiss safety modal
		await page.getByRole('button', { name: /Passenger|Parked Safely|Acknowledge/i }).click();

		// Tap 1: Select Mule Deer species chip
		const muleDeerChip = page.getByRole('button', { name: /Mule Deer/i });
		await muleDeerChip.click();
		await expect(muleDeerChip).toHaveAttribute('aria-pressed', 'true');

		// Provide mock GPS coordinates if not already present
		await page.fill('#latitude', '43.8000');
		await page.fill('#longitude', '-110.2000');

		// Tap 3: Click Submit Carcass Report
		const submitBtn = page.getByRole('button', { name: /Submit Carcass Report/i });
		await submitBtn.click();

		// Verify success message / feedback
		await expect(page.getByText(/Report submitted|Report saved/i)).toBeVisible();
		expect(reportPayload).not.toBeNull();
		expect(reportPayload?.['species_common_name']).toBe('mule_deer');
	});

	test('offline network enqueues report and reconnecting triggers sync', async ({
		page,
		context
	}) => {
		let syncCount = 0;
		await page.route('**/v1/reports', async (route) => {
			syncCount++;
			await route.fulfill({
				status: 201,
				contentType: 'application/json',
				body: JSON.stringify({ status: 'success', data: { id: 'test-sync-id' } })
			});
		});

		await page.goto('/report/');

		// Dismiss safety modal
		await page.getByRole('button', { name: /Passenger|Parked Safely|Acknowledge/i }).click();

		// Go offline
		await context.setOffline(true);

		// Select Elk species chip
		const elkChip = page.getByRole('button', { name: /Elk/i });
		await elkChip.click();
		await page.fill('#latitude', '44.1000');
		await page.fill('#longitude', '-110.5000');

		// Submit offline
		await page.getByRole('button', { name: /Submit Carcass Report/i }).click();

		// Should show queued indicator / pending count >= 1
		await expect(page.locator('.queue-pill')).toHaveText(/Pending: 1/);
		expect(syncCount).toBe(0);

		// Reconnect to online
		await context.setOffline(false);

		// Trigger online event to trigger automatic background sync
		await page.evaluate(() => window.dispatchEvent(new Event('online')));

		// Automatic background sync runs, and pending count drops to 0
		await expect(page.locator('.queue-pill')).toHaveText(/Pending: 0/);
		expect(syncCount).toBeGreaterThanOrEqual(1);
	});
});
