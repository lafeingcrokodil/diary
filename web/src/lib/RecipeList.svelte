<script lang="ts">
	import { resolve } from '$app/paths';
	import type { Recipe } from '#lib/recipes.ts';

	interface Props {
		recipes: Recipe[];
	}

	const { recipes }: Props = $props();

	const describe = (recipe: Recipe): string =>
		recipe.ingredients.length === 1
			? '1 ingredient'
			: `${String(recipe.ingredients.length)} ingredients`;
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
