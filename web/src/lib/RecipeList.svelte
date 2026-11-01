<script lang="ts">
	import { resolve } from '$app/paths';
	import { formatTotal } from '#lib/recipe-nutrition.ts';
	import type { Recipe } from '#lib/recipes.ts';

	interface Props {
		recipes: Recipe[];
	}

	const { recipes }: Props = $props();

	// Describes a recipe's size and energy, e.g. "3 ingredients · 1,058 kcal".
	const describe = (recipe: Recipe): string => {
		const count = recipe.ingredients.length;
		if (count === 0) {
			return 'No ingredients yet';
		}
		const ingredients = count === 1 ? '1 ingredient' : `${String(count)} ingredients`;
		return `${ingredients} · ${formatTotal(recipe.nutrition.kcal, 'kcal')}`;
	};
</script>

<ul>
	{#each recipes as recipe (recipe.id)}
		<li>
			<a href={resolve('/recipes/[id]', { id: recipe.id })}>
				<span class="name">{recipe.name}</span>
				<span class="muted">{describe(recipe)}</span>
			</a>
		</li>
	{/each}
</ul>

<style>
	ul {
		padding: 0;
		margin: 0;
		list-style: none;
		background: var(--surface);
		border: 1px solid var(--border);
		border-radius: var(--radius);
	}

	li + li {
		border-top: 1px solid var(--border);
	}

	/* The whole row is the link, so it's easy to tap. */
	li a {
		display: grid;
		padding: 0.75rem var(--space);
		color: inherit;
		text-decoration: none;
	}

	.name {
		font-weight: 600;
	}

	.muted {
		font-size: 0.875rem;
	}
</style>
