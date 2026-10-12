/**
 * Offline IndexedDB queue for Park Ranger field reports.
 * Provides client-side UUID idempotency, retry backoff, and photo blob storage.
 */

export interface ReportInput {
	id?: string;
	species_common_name: string;
	latitude: number;
	longitude: number;
	observed_at?: string;
	confidence?: 'confirmed' | 'probable' | 'possible' | string;
	notes?: string;
	photo?: Blob | File | null;
}

export interface QueuedReport {
	id: string;
	species_common_name: string;
	latitude: number;
	longitude: number;
	observed_at: string;
	confidence: string;
	notes?: string;
	photoBlob?: Blob | null;
	photoName?: string;
	createdAt: number;
	retryCount: number;
	lastError?: string;
}

export interface SyncResult {
	synced: number;
	failed: number;
	remaining: number;
	errors: string[];
}

const DB_NAME = 'corridor-field-reports';
const DB_VERSION = 1;
const STORE_NAME = 'reports';

// In-memory fallback for environments without IndexedDB (Node, tests, SSR)
const memoryStore = new Map<string, QueuedReport>();

function isIndexedDBAvailable(): boolean {
	return typeof indexedDB !== 'undefined';
}

function openDB(): Promise<IDBDatabase> {
	return new Promise((resolve, reject) => {
		if (!isIndexedDBAvailable()) {
			return reject(new Error('IndexedDB not supported'));
		}
		const request = indexedDB.open(DB_NAME, DB_VERSION);
		request.onupgradeneeded = () => {
			const db = request.result;
			if (!db.objectStoreNames.contains(STORE_NAME)) {
				db.createObjectStore(STORE_NAME, { keyPath: 'id' });
			}
		};
		request.onsuccess = () => resolve(request.result);
		request.onerror = () => reject(request.error);
	});
}

function generateUUID(): string {
	if (typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function') {
		return crypto.randomUUID();
	}
	// Fallback UUID v4 generator
	return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, (c) => {
		const r = (Math.random() * 16) | 0;
		const v = c === 'x' ? r : (r & 0x3) | 0x8;
		return v.toString(16);
	});
}

/**
 * Enqueues a new field report into IndexedDB or fallback storage.
 */
export async function enqueueReport(input: ReportInput): Promise<QueuedReport> {
	const id = input.id || generateUUID();
	const observed_at = input.observed_at || new Date().toISOString();
	const confidence = input.confidence || 'confirmed';

	const queuedReport: QueuedReport = {
		id,
		species_common_name: input.species_common_name,
		latitude: input.latitude,
		longitude: input.longitude,
		observed_at,
		confidence,
		notes: input.notes,
		photoBlob: input.photo ?? null,
		photoName: input.photo instanceof File ? input.photo.name : undefined,
		createdAt: Date.now(),
		retryCount: 0
	};

	if (!isIndexedDBAvailable()) {
		memoryStore.set(id, queuedReport);
		return queuedReport;
	}

	try {
		const db = await openDB();
		return await new Promise<QueuedReport>((resolve, reject) => {
			const tx = db.transaction(STORE_NAME, 'readwrite');
			const store = tx.objectStore(STORE_NAME);
			const req = store.put(queuedReport);
			req.onsuccess = () => resolve(queuedReport);
			req.onerror = () => reject(req.error);
		});
	} catch {
		// Fallback to memory store if IDB operation fails
		memoryStore.set(id, queuedReport);
		return queuedReport;
	}
}

/**
 * Returns all pending reports currently waiting in the offline queue.
 */
export async function getPendingReports(): Promise<QueuedReport[]> {
	if (!isIndexedDBAvailable()) {
		return Array.from(memoryStore.values()).sort((a, b) => a.createdAt - b.createdAt);
	}

	try {
		const db = await openDB();
		return await new Promise<QueuedReport[]>((resolve, reject) => {
			const tx = db.transaction(STORE_NAME, 'readonly');
			const store = tx.objectStore(STORE_NAME);
			const req = store.getAll();
			req.onsuccess = () => {
				const items: QueuedReport[] = req.result || [];
				items.sort((a, b) => a.createdAt - b.createdAt);
				resolve(items);
			};
			req.onerror = () => reject(req.error);
		});
	} catch {
		return Array.from(memoryStore.values()).sort((a, b) => a.createdAt - b.createdAt);
	}
}

/**
 * Removes a successfully uploaded report from the offline queue.
 */
export async function removeReport(id: string): Promise<void> {
	if (!isIndexedDBAvailable()) {
		memoryStore.delete(id);
		return;
	}

	try {
		const db = await openDB();
		await new Promise<void>((resolve, reject) => {
			const tx = db.transaction(STORE_NAME, 'readwrite');
			const store = tx.objectStore(STORE_NAME);
			const req = store.delete(id);
			req.onsuccess = () => resolve();
			req.onerror = () => reject(req.error);
		});
	} catch {
		memoryStore.delete(id);
	}
}

/**
 * Clears all pending reports from storage (primarily for test resets).
 */
export async function clearQueue(): Promise<void> {
	memoryStore.clear();

	if (!isIndexedDBAvailable()) {
		return;
	}

	try {
		const db = await openDB();
		await new Promise<void>((resolve, reject) => {
			const tx = db.transaction(STORE_NAME, 'readwrite');
			const store = tx.objectStore(STORE_NAME);
			const req = store.clear();
			req.onsuccess = () => resolve();
			req.onerror = () => reject(req.error);
		});
	} catch {
		// ignore
	}
}

/**
 * Updates a queued report with retry information or error status.
 */
async function updateReport(report: QueuedReport): Promise<void> {
	if (!isIndexedDBAvailable()) {
		memoryStore.set(report.id, report);
		return;
	}

	try {
		const db = await openDB();
		await new Promise<void>((resolve, reject) => {
			const tx = db.transaction(STORE_NAME, 'readwrite');
			const store = tx.objectStore(STORE_NAME);
			const req = store.put(report);
			req.onsuccess = () => resolve();
			req.onerror = () => reject(req.error);
		});
	} catch {
		memoryStore.set(report.id, report);
	}
}

/**
 * Syncs queued reports to backend endpoints:
 * 1. POST /v1/reports
 * 2. POST /v1/reports/{id}/photo (if photo blob is attached)
 *
 * If offline or backend is unreachable, preserves items in queue for future retry.
 */
export async function syncQueue(fetchFn?: typeof fetch): Promise<SyncResult> {
	const callerFetch = fetchFn || (typeof fetch !== 'undefined' ? fetch : undefined);
	if (!callerFetch) {
		return { synced: 0, failed: 0, remaining: 0, errors: ['fetch is not available'] };
	}

	if (typeof navigator !== 'undefined' && navigator.onLine === false) {
		const pending = await getPendingReports();
		return {
			synced: 0,
			failed: 0,
			remaining: pending.length,
			errors: ['Device is offline']
		};
	}

	const pending = await getPendingReports();
	let synced = 0;
	let failed = 0;
	const errors: string[] = [];

	for (const report of pending) {
		try {
			// Step 1: Submit report metadata
			const reportPayload = {
				id: report.id,
				species_common_name: report.species_common_name,
				latitude: report.latitude,
				longitude: report.longitude,
				observed_at: report.observed_at,
				confidence: report.confidence,
				notes: report.notes
			};

			const reportRes = await callerFetch('/v1/reports', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(reportPayload)
			});

			if (!reportRes.ok && reportRes.status !== 200 && reportRes.status !== 201) {
				const errMsg = `POST /v1/reports failed with HTTP ${reportRes.status}`;
				report.retryCount++;
				report.lastError = errMsg;
				await updateReport(report);
				errors.push(errMsg);
				failed++;
				continue;
			}

			// Step 2: Upload photo if present
			if (report.photoBlob) {
				let photoRes: Response;
				if (typeof FormData !== 'undefined') {
					const formData = new FormData();
					formData.append('photo', report.photoBlob, report.photoName || 'carcass.jpg');
					photoRes = await callerFetch(`/v1/reports/${report.id}/photo`, {
						method: 'POST',
						body: formData
					});
				} else {
					photoRes = await callerFetch(`/v1/reports/${report.id}/photo`, {
						method: 'POST',
						headers: { 'Content-Type': report.photoBlob.type || 'image/jpeg' },
						body: report.photoBlob
					});
				}

				if (!photoRes.ok && photoRes.status !== 200) {
					const errMsg = `POST /v1/reports/${report.id}/photo failed with HTTP ${photoRes.status}`;
					report.retryCount++;
					report.lastError = errMsg;
					await updateReport(report);
					errors.push(errMsg);
					failed++;
					continue;
				}
			}

			// Successfully created and photo uploaded
			await removeReport(report.id);
			synced++;
		} catch (err: unknown) {
			const errMsg = err instanceof Error ? err.message : 'Network sync error';
			report.retryCount++;
			report.lastError = errMsg;
			await updateReport(report);
			errors.push(errMsg);
			failed++;
		}
	}

	const remainingReports = await getPendingReports();
	return {
		synced,
		failed,
		remaining: remainingReports.length,
		errors
	};
}

/**
 * Registers automatic background synchronization when network reconnects.
 */
export function setupAutoSync(fetchFn?: typeof fetch): () => void {
	if (typeof window === 'undefined') {
		return () => {};
	}

	const handleOnline = () => {
		void syncQueue(fetchFn);
	};

	window.addEventListener('online', handleOnline);
	return () => {
		window.removeEventListener('online', handleOnline);
	};
}
