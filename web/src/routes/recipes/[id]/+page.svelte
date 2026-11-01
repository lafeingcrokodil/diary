<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { page } from '$app/state';
	import RecipeEditor from '#lib/RecipeEditor.svelte';
	import RecipeNutrition from '#lib/RecipeNutrition.svelte';
	import { type Food, listFoods } from '#lib/foods.ts';
	import { type Nutrient, listNutrients } from '#lib/nutrients.ts';
	import { FoodChoices, toForm } from '#lib/recipe-form.ts';
	import {
		type Recipe,
		type RecipeInput,
		deleteRecipe,
		getRecipe,
		updateRecipe
	} from '#lib/recipes.ts';

	const { id } = page.params;
	const loading: Promise<[Recipe, Food[], Nutrient[]]> = Promise.all([
		getRecipe(id ?? ''),
		listFoods(),
		listNutrients()
	]);

	const save = async (input: RecipeInput): Promise<void> => {
		await updateRecipe(id ?? '', input);
		await goto(resolve('/recipes'));
	};

	const remove = async (): Promise<void> => {
		await deleteRecipe(id ?? '');
		await goto(resolve('/recipes'));
	};
</script>

<main>
	{#await loading}
		<p class="muted">Loading…</p>
	{:then [recipe, foods, nutrients]}
		<h1>{recipe.name}</h1>
		{#if recipe.archived}
			<p class="muted">This recipe is archived. Unarchive it to use it again.</p>
		{/if}
		{#if recipe.sourceUrl !== ''}
			<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -- The source is another website, not one of the app's routes. -->
			<p><a href={recipe.sourceUrl} rel="noopener noreferrer" target="_blank">Source</a></p>
		{/if}
		<RecipeNutrition {foods} {nutrients} {recipe} />
		<RecipeEditor
			archived={recipe.archived}
			{foods}
			initial={toForm(recipe, new FoodChoices(foods))}
			ondelete={remove}
			onsave={save}
		/>
	{:catch error}
		<p role="alert">Loading the recipe failed: {String(error)}</p>
	{/await}
	<p><a href={resolve('/recipes')}>Back to recipes</a></p>
</main>
