<script lang="ts">
	import { resolve } from '$app/paths';
	import { type Food, listFoods } from '#lib/foods.ts';

	const foods: Promise<Food[]> = listFoods();
</script>

<main>
	<h1>Foods</h1>
	<a href={resolve('/foods/new')}>Add a food</a>

	{#await foods}
		<p>Loading…</p>
	{:then list}
		{#if list.length === 0}
			<p>No foods yet.</p>
		{:else}
			<ul>
				{#each list as food (food.id)}
					<li>
						<a href={resolve('/foods/[id]', { id: food.id })}>{food.name}</a>
						{#if food.brand !== ''}({food.brand}){/if}
						– {food.kcalPer100} kcal per 100 {food.unit}
					</li>
				{/each}
			</ul>
		{/if}
	{:catch error}
		<p role="alert">Loading the foods failed: {String(error)}</p>
	{/await}
</main>
