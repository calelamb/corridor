<script lang="ts">
	import type { AcademicPaper } from './catalog';

	interface Props {
		paper: AcademicPaper;
	}

	let { paper }: Props = $props();

	function categoryLabel(cat: AcademicPaper['category']): string {
		switch (cat) {
			case 'mitigation':
				return 'Mitigation Efficacy';
			case 'bias':
				return 'Reporting Bias';
			case 'ecology':
				return 'Migration Ecology';
			case 'economics':
				return 'Collision Economics';
		}
	}
</script>

<article class="evidence-card" aria-labelledby="paper-{paper.id}">
	<div class="card-meta">
		<span class="category-pill category-{paper.category}">
			{categoryLabel(paper.category)}
		</span>
		<span class="year-badge">{paper.year}</span>
	</div>

	<h2 id="paper-{paper.id}" class="paper-title">{paper.title}</h2>
	<p class="paper-authors">{paper.authors}</p>
	<p class="paper-journal">{paper.journal}</p>

	<div class="findings-block">
		<h3>Key Scientific Finding</h3>
		<p>{paper.key_findings}</p>
	</div>

	<div class="applied-block">
		<div class="applied-header">
			<span class="corridor-mark" aria-hidden="true">⋈</span>
			<h3>Applied in Corridor</h3>
		</div>
		<p>{paper.applied_in_corridor}</p>
	</div>

	<div class="card-footer">
		<a
			href="https://doi.org/{paper.doi.replace(/^https?:\/\/doi\.org\//, '')}"
			target="_blank"
			rel="noreferrer"
			class="doi-link"
			aria-label="View DOI: {paper.doi}"
		>
			<span>DOI: {paper.doi}</span>
			<span aria-hidden="true" class="arrow">↗</span>
		</a>
	</div>
</article>

<style>
	.evidence-card {
		background: var(--panel, #ffffff);
		border: 1px solid var(--line, #e2e8f0);
		border-radius: 8px;
		padding: 24px;
		display: flex;
		flex-direction: column;
		gap: 14px;
		box-shadow: 0 2px 4px rgba(0, 0, 0, 0.04);
		transition:
			border-color 0.15s,
			box-shadow 0.15s;
	}

	.evidence-card:hover,
	.evidence-card:focus-within {
		border-color: var(--focus, #2563eb);
		box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
	}

	.card-meta {
		display: flex;
		justify-content: space-between;
		align-items: center;
	}

	.category-pill {
		font-size: 0.75rem;
		font-weight: 700;
		text-transform: uppercase;
		letter-spacing: 0.04em;
		padding: 4px 10px;
		border-radius: 9999px;
	}

	.category-mitigation {
		background: #dbeafe;
		color: #1e40af;
	}

	.category-bias {
		background: #fef3c7;
		color: #92400e;
	}

	.category-ecology {
		background: #d1fae5;
		color: #065f46;
	}

	.category-economics {
		background: #ede9fe;
		color: #5b21b6;
	}

	.year-badge {
		font-size: 0.85rem;
		font-weight: 600;
		color: var(--muted, #526258);
	}

	:global([data-theme='dark']) .year-badge {
		color: #bdcbbf;
	}

	.paper-title {
		font-size: 1.15rem;
		font-weight: 700;
		line-height: 1.35;
		margin: 0;
		color: var(--ink, #0f172a);
	}

	.paper-authors {
		font-size: 0.88rem;
		color: var(--ink, #183c32);
		margin: 0;
		line-height: 1.4;
	}

	:global([data-theme='dark']) .paper-authors {
		color: #f6f3ea;
	}

	.paper-journal {
		font-size: 0.82rem;
		font-style: italic;
		color: var(--muted, #526258);
		margin: 0;
	}

	:global([data-theme='dark']) .paper-journal {
		color: #bdcbbf;
	}

	.findings-block {
		background: var(--surface, #f8fafc);
		border-left: 3px solid var(--line, #64748b);
		padding: 12px 16px;
		border-radius: 0 6px 6px 0;
	}

	.findings-block h3 {
		font-size: 0.78rem;
		font-weight: 700;
		text-transform: uppercase;
		letter-spacing: 0.04em;
		color: var(--muted, #475569);
		margin: 0 0 4px;
	}

	:global([data-theme='dark']) .findings-block h3 {
		color: #bdcbbf;
	}

	.findings-block p {
		font-size: 0.9rem;
		line-height: 1.5;
		margin: 0;
		color: var(--ink, #1e293b);
	}

	.applied-block {
		background: #f0fdf4;
		border: 1px solid #bbf7d0;
		border-radius: 6px;
		padding: 12px 16px;
	}

	:global([data-theme='dark']) .applied-block {
		background: #14352a;
		border-color: #245745;
	}

	.applied-header {
		display: flex;
		align-items: center;
		gap: 6px;
		margin-bottom: 4px;
	}

	.corridor-mark {
		font-size: 0.95rem;
		color: #166534;
		font-weight: bold;
	}

	:global([data-theme='dark']) .corridor-mark {
		color: #86efac;
	}

	.applied-header h3 {
		font-size: 0.82rem;
		font-weight: 700;
		text-transform: uppercase;
		letter-spacing: 0.04em;
		color: #166534;
		margin: 0;
	}

	:global([data-theme='dark']) .applied-header h3 {
		color: #86efac;
	}

	.applied-block p {
		font-size: 0.88rem;
		line-height: 1.5;
		margin: 0;
		color: #14532d;
	}

	:global([data-theme='dark']) .applied-block p {
		color: #dcfce7;
	}

	.card-footer {
		margin-top: auto;
		padding-top: 8px;
	}

	.doi-link {
		display: inline-flex;
		align-items: center;
		gap: 6px;
		min-height: 48px;
		padding: 8px 14px;
		font-size: 0.82rem;
		font-weight: 600;
		color: var(--accent, #0d9488);
		text-decoration: none;
		border: 1px solid var(--line, #cbd5e1);
		border-radius: 6px;
		background: var(--surface, #f8fafc);
	}

	.doi-link:hover,
	.doi-link:focus-visible {
		background: #e2e8f0;
		text-decoration: underline;
		outline: 2px solid var(--focus, #2563eb);
	}

	.arrow {
		font-size: 0.95rem;
	}
</style>
