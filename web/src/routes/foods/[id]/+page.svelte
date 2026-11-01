<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { page } from '$app/state';
	import FoodEditor from '#lib/FoodEditor.svelte';
	import { toForm } from '#lib/food-form.ts';
	import { type Food, type FoodInput, deleteFood, getFood, updateFood } from '#lib/foods.ts';
	import { type Nutrient, listNutrients } from '#lib/nutrients.ts';

	const { id } = page.params;
	const loading: Promise<[Food, Nutrient[]]> = Promise.all([getFood(id ?? ''), listNutrients()]);

	const save = async (input: FoodInput): Promise<void> => {
		await updateFood(id ?? '', input);
		await goto(resolve('/foods'));
	};

	const remove = async (): Promise<void> => {
		await deleteFood(id ?? '');
		await goto(resolve('/foods'));
	};
</script>

<main>
	{#await loading}
		<p class="muted">Loading…</p>
	{:then [food, nutrients]}
		<h1>{food.name}</h1>
		{#if food.archived}
			<p class="muted">This food is archived. Unarchive it to use it again.</p>
		{/if}
		<FoodEditor
			archived={food.archived}
			initial={toForm(food)}
			{nutrients}
			ondelete={remove}
			onsave={save}
		/>
	{:catch error}
		<p role="alert">Loading the food failed: {String(error)}</p>
	{/await}
	<p><a href={resolve('/foods')}>Back to foods</a></p>
</main>
