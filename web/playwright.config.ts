import { defineConfig, devices } from '@playwright/test';
export default defineConfig({
	testDir: 'tests',
	use: { baseURL: process.env.PLAYWRIGHT_BASE_URL ?? 'http://127.0.0.1:4173' },
	projects: [
		{ name: 'chromium', use: { ...devices['Desktop Chrome'] } },
		{ name: 'firefox', use: { ...devices['Desktop Firefox'] } },
		{ name: 'webkit', use: { ...devices['Desktop Safari'] } }
	],
	webServer: process.env.PLAYWRIGHT_BASE_URL
		? undefined
		: {
				command: 'npm run build && npm run preview -- --host 127.0.0.1',
				port: 4173,
				reuseExistingServer: !process.env.CI
			}
});
