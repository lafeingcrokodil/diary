<script lang="ts">
	import type { Food } from '#lib/foods.ts';
	import type { Nutrient } from '#lib/nutrients.ts';
	import { formatTotal, nutritionRows } from '#lib/recipe-nutrition.ts';
	import type { Recipe } from '#lib/recipes.ts';

	interface Props {
		recipe: Recipe;
		// All of the user's nutrients. Only active ones are shown.
		nutrients: Nutrient[];
		// All of the user's foods, archived or not, to name the ingredients
		// whose amounts aren't known.
		foods: Food[];
	}

	const { recipe, nutrients, foods }: Props = $props();

	const rows = $derived(nutritionRows(recipe, nutrients, foods));
</script>

<section>
	<h2>Nutrition</h2>
	<p class="muted">For the whole recipe, as last saved.</p>
	<dl>
		{#each rows as row (row.name)}
			<div class="row">
				<dt>{row.name}</dt>
				<dd>{formatTotal(row.total, row.unit)}</dd>
				{#if row.missing.length > 0}
					<dd class="missing muted">Not known for {row.missing.join(', ')}</dd>
				{/if}
			</div>
		{/each}
	</dl>
</section>

<style>
	section {
		padding: var(--space);
		margin-bottom: var(--space);
		background: var(--surface);
		border: 1px solid var(--border);
		border-radius: var(--radius);
	}

	h2 {
		margin: 0;
		font-size: 1.125rem;
	}

	p {
		margin: 0 0 0.5rem;
		font-size: 0.875rem;
	}

	dl {
		margin: 0;
	}

	/* A nutrient's name and total on one line, with any note below them. */
	.row {
		display: grid;
		grid-template-columns: 1fr auto;
		gap: 0 var(--space);
		padding: 0.375rem 0;
	}

	.row + .row {
		border-top: 1px solid var(--border);
	}

	dd {
		margin: 0;
		font-variant-numeric: tabular-nums;
		text-align: right;
	}

	.missing {
		grid-column: 1 / -1;
		font-size: 0.875rem;
		text-align: left;
	}
</style>
