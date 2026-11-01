<script lang="ts">
	import { resolve } from '$app/paths';
	import NutrientList from '#lib/NutrientList.svelte';
	import { type Nutrient, listNutrients } from '#lib/nutrients.ts';

	const nutrients: Promise<Nutrient[]> = listNutrients();
</script>

<main>
	<div class="title">
		<h1>Nutrients</h1>
		<a class="button" href={resolve('/nutrients/new')}>Add a nutrient</a>
	</div>

	{#await nutrients}
		<p class="muted">Loading…</p>
	{:then all}
		{@const active = all.filter((nutrient) => !nutrient.archived)}
		{@const archived = all.filter((nutrient) => nutrient.archived)}
		{#if active.length === 0}
			<p class="muted">No nutrients yet.</p>
		{:else}
			<NutrientList nutrients={active} />
		{/if}
		{#if archived.length > 0}
			<details>
				<summary>Archived ({archived.length})</summary>
				<NutrientList nutrients={archived} />
			</details>
		{/if}
	{:catch error}
		<p role="alert">Loading the nutrients failed: {String(error)}</p>
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
