<script lang="ts">
	import { resolve } from '$app/paths';
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
	{:then list}
		{#if list.length === 0}
			<p class="muted">No foods yet.</p>
		{:else}
			<ul>
				{#each list as food (food.id)}
					<li>
						<a href={resolve('/foods/[id]', { id: food.id })}>
							<span class="name">{food.name}</span>
							<span class="muted">
								{#if food.brand !== ''}{food.brand} ·{/if}
								{food.kcalPer100} kcal per 100 {food.unit}
							</span>
						</a>
					</li>
				{/each}
			</ul>
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
