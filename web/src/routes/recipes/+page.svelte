<script lang="ts">
	import { resolve } from '$app/paths';
	import RecipeList from '#lib/RecipeList.svelte';
	import { type Recipe, listRecipes } from '#lib/recipes.ts';

	const recipes: Promise<Recipe[]> = listRecipes();
</script>

<main>
	<div class="title">
		<h1>Recipes</h1>
		<a class="button" href={resolve('/recipes/new')}>Add a recipe</a>
	</div>

	{#await recipes}
		<p class="muted">Loading…</p>
	{:then all}
		{@const active = all.filter((recipe) => !recipe.archived)}
		{@const archived = all.filter((recipe) => recipe.archived)}
		{#if active.length === 0}
			<p class="muted">No recipes yet.</p>
		{:else}
			<RecipeList recipes={active} />
		{/if}
		{#if archived.length > 0}
			<details>
				<summary>Archived ({archived.length})</summary>
				<RecipeList recipes={archived} />
			</details>
		{/if}
	{:catch error}
		<p role="alert">Loading the recipes failed: {String(error)}</p>
	{/await}
</main>

<style>
	.title {
		display: flex;
		gap: var(--space);
		align-items: center;
		justify-content: space-between;
	}

	details {
		margin-top: var(--space);
	}

	summary {
		padding: 0.5rem 0;
		color: var(--muted);
		cursor: pointer;
	}
</style>
