<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { page } from '$app/state';
	import FoodEditor from '#lib/FoodEditor.svelte';
	import { toForm } from '#lib/food-form.ts';
	import { type Food, type FoodInput, deleteFood, getFood, updateFood } from '#lib/foods.ts';

	const { id } = page.params;
	const food: Promise<Food> = getFood(id ?? '');

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
	{#await food}
		<p class="muted">Loading…</p>
	{:then loaded}
		<h1>{loaded.name}</h1>
		<FoodEditor initial={toForm(loaded)} ondelete={remove} onsave={save} />
	{:catch error}
		<p role="alert">Loading the food failed: {String(error)}</p>
	{/await}
	<p><a href={resolve('/foods')}>Back to foods</a></p>
</main>
