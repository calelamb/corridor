<script lang="ts">
	import { academicPapers } from './catalog';
	import EvidenceCard from './EvidenceCard.svelte';

	type CategoryFilter = 'all' | 'mitigation' | 'bias' | 'ecology' | 'economics';

	let activeCategory = $state<CategoryFilter>('all');
	let searchQuery = $state('');

	const categoryOptions: { id: CategoryFilter; label: string }[] = [
		{ id: 'all', label: 'All' },
		{ id: 'mitigation', label: 'Mitigation Efficacy' },
		{ id: 'bias', label: 'Reporting Bias' },
		{ id: 'ecology', label: 'Migration Ecology' },
		{ id: 'economics', label: 'Collision Economics' }
	];

	let filteredPapers = $derived(
		academicPapers.filter((paper) => {
			const matchesCategory = activeCategory === 'all' || paper.category === activeCategory;
			const q = searchQuery.toLowerCase().trim();
			const matchesSearch =
				q === '' ||
				paper.title.toLowerCase().includes(q) ||
				paper.authors.toLowerCase().includes(q) ||
				paper.journal.toLowerCase().includes(q) ||
				paper.key_findings.toLowerCase().includes(q) ||
				paper.applied_in_corridor.toLowerCase().includes(q);

			return matchesCategory && matchesSearch;
		})
	);
</script>

<div class="evidence-library" aria-label="Scientific Evidence Library">
	<div class="library-toolbar">
		<div class="search-box">
			<label for="evidence-search" class="sr-only">Search academic papers</label>
			<input
				id="evidence-search"
				type="search"
				placeholder="Search papers, authors, findings, or algorithms..."
				bind:value={searchQuery}
				class="search-input"
			/>
			{#if searchQuery}
				<button
					type="button"
					class="clear-btn"
					onclick={() => (searchQuery = '')}
					aria-label="Clear search"
				>
					✕
				</button>
			{/if}
		</div>

		<div class="category-filters" role="group" aria-label="Filter by topic">
			{#each categoryOptions as opt (opt.id)}
				<button
					type="button"
					class="filter-btn"
					class:active={activeCategory === opt.id}
					aria-pressed={activeCategory === opt.id}
					onclick={() => (activeCategory = opt.id)}
				>
					{opt.label}
				</button>
			{/each}
		</div>
	</div>

	<div class="results-meta" role="status" aria-live="polite">
		<span>Showing {filteredPapers.length} peer-reviewed research publications</span>
	</div>

	{#if filteredPapers.length === 0}
		<div class="no-results">
			<p>No academic publications match your current search and category filters.</p>
			<button
				type="button"
				class="reset-btn"
				onclick={() => {
					activeCategory = 'all';
					searchQuery = '';
				}}
			>
				Reset All Filters
			</button>
		</div>
	{:else}
		<div class="papers-grid">
			{#each filteredPapers as paper (paper.id)}
				<EvidenceCard {paper} />
			{/each}
		</div>
	{/if}
</div>

<style>
	.evidence-library {
		display: flex;
		flex-direction: column;
		gap: 20px;
	}

	.library-toolbar {
		display: flex;
		flex-direction: column;
		gap: 16px;
		background: var(--panel, #ffffff);
		padding: 20px;
		border: 1px solid var(--line, #e2e8f0);
		border-radius: 8px;
	}

	.search-box {
		position: relative;
		width: 100%;
	}

	.search-input {
		width: 100%;
		min-height: 48px;
		padding: 10px 16px;
		font-size: 1rem;
		border: 1px solid var(--line, #cbd5e1);
		border-radius: 6px;
		background: var(--surface, #f8fafc);
		color: var(--ink, #0f172a);
	}

	.search-input:focus {
		outline: 2px solid var(--focus, #2563eb);
		border-color: transparent;
	}

	.clear-btn {
		position: absolute;
		right: 12px;
		top: 50%;
		transform: translateY(-50%);
		min-width: 48px;
		min-height: 48px;
		background: transparent;
		border: none;
		font-size: 1.1rem;
		color: var(--muted, #64748b);
		cursor: pointer;
		display: flex;
		align-items: center;
		justify-content: center;
	}

	.category-filters {
		display: flex;
		flex-wrap: wrap;
		gap: 10px;
	}

	.filter-btn {
		min-height: 48px;
		min-width: 48px;
		padding: 10px 18px;
		font-size: 0.9rem;
		font-weight: 600;
		border: 1px solid var(--line, #cbd5e1);
		border-radius: 6px;
		background: var(--surface, #f8fafc);
		color: var(--ink, #1e293b);
		cursor: pointer;
		display: inline-flex;
		align-items: center;
		justify-content: center;
	}

	.filter-btn:hover {
		background: #e2e8f0;
	}

	.filter-btn.active {
		background: var(--accent, #0d9488);
		border-color: var(--accent, #0d9488);
		color: var(--accent-ink, #ffffff);
	}

	.results-meta {
		font-size: 0.9rem;
		font-weight: 600;
		color: var(--ink, #183c32);
	}

	.papers-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(360px, 1fr));
		gap: 20px;
	}

	@media (max-width: 640px) {
		.papers-grid {
			grid-template-columns: 1fr;
		}
	}

	.no-results {
		text-align: center;
		padding: 48px 24px;
		background: var(--panel, #ffffff);
		border: 1px solid var(--line, #e2e8f0);
		border-radius: 8px;
		color: var(--muted, #64748b);
	}

	.no-results p {
		margin-bottom: 16px;
		font-size: 1rem;
	}

	.reset-btn {
		min-height: 48px;
		padding: 10px 20px;
		background: var(--accent, #0d9488);
		color: var(--accent-ink, #ffffff);
		border: none;
		border-radius: 6px;
		font-weight: 600;
		cursor: pointer;
	}

	.sr-only {
		position: absolute;
		width: 1px;
		height: 1px;
		padding: 0;
		margin: -1px;
		overflow: hidden;
		clip: rect(0, 0, 0, 0);
		white-space: nowrap;
		border-width: 0;
	}
</style>
