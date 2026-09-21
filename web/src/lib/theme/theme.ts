export type Theme = 'light' | 'dark';
export function readTheme(storage: Pick<Storage, 'getItem'> | null, systemDark: boolean): Theme {
	try {
		const stored = storage?.getItem('corridor-theme');
		if (stored === 'light' || stored === 'dark') return stored;
	} catch {
		/* Storage denial preserves system preference. */
	}
	return systemDark ? 'dark' : 'light';
}
export function saveTheme(storage: Pick<Storage, 'setItem'> | null, theme: Theme): void {
	try {
		storage?.setItem('corridor-theme', theme);
	} catch {
		/* Current theme remains usable without persistence. */
	}
}
export function browserStorage(): Storage | null {
	try {
		return window.localStorage;
	} catch {
		return null;
	}
}
