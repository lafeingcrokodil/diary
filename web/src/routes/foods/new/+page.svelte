<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import FoodEditor from '#lib/FoodEditor.svelte';
	import { emptyForm } from '#lib/food-form.ts';
	import { type FoodInput, createFood } from '#lib/foods.ts';
	import { type Nutrient, listNutrients } from '#lib/nutrients.ts';

	const nutrients: Promise<Nutrient[]> = listNutrients();

	const save = async (input: FoodInput): Promise<void> => {
		await createFood(input);
		await goto(resolve('/foods'));
	};
</script>

<main>
	<h1>New food</h1>
	{#await nutrients}
		<p class="muted">Loading…</p>
	{:then loaded}
		<FoodEditor initial={emptyForm()} nutrients={loaded} onsave={save} />
	{:catch error}
		<p role="alert">Loading the nutrients failed: {String(error)}</p>
	{/await}
	<p><a href={resolve('/foods')}>Back to foods</a></p>
</main>
