<script lang="ts">
	import { resolve } from '$app/paths';
	import FoodList from '#lib/FoodList.svelte';
	import { type Food, listFoods } from '#lib/foods.ts';

	const foods: Promise<Food[]> = listFoods();
</script>

<main>
	<div class="title">
		<h1>Foods</h1>
		<a class="button" href={resolve('/foods/new')}>Add a food</a>
	</div>

	{#await foods}
		<p class="muted">Loading…</p>
	{:then all}
		{@const active = all.filter((food) => !food.archived)}
		{@const archived = all.filter((food) => food.archived)}
		{#if active.length === 0}
			<p class="muted">No foods yet.</p>
		{:else}
			<FoodList foods={active} />
		{/if}
		{#if archived.length > 0}
			<details>
				<summary>Archived ({archived.length})</summary>
				<FoodList foods={archived} />
			</details>
		{/if}
	{:catch error}
		<p role="alert">Loading the foods failed: {String(error)}</p>
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
