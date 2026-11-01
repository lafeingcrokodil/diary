<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import RecipeEditor from '#lib/RecipeEditor.svelte';
	import { type Food, listFoods } from '#lib/foods.ts';
	import { emptyForm } from '#lib/recipe-form.ts';
	import { type RecipeInput, createRecipe } from '#lib/recipes.ts';

	const foods: Promise<Food[]> = listFoods();

	const save = async (input: RecipeInput): Promise<void> => {
		await createRecipe(input);
		await goto(resolve('/recipes'));
	};
</script>

<main>
	<h1>New recipe</h1>
	{#await foods}
		<p class="muted">Loading…</p>
	{:then loaded}
		<RecipeEditor foods={loaded} initial={emptyForm()} onsave={save} />
	{:catch error}
		<p role="alert">Loading the foods failed: {String(error)}</p>
	{/await}
	<p><a href={resolve('/recipes')}>Back to recipes</a></p>
</main>
