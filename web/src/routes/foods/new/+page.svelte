<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import FoodEditor from '#lib/FoodEditor.svelte';
	import { emptyForm } from '#lib/food-form.ts';
	import { type FoodInput, createFood } from '#lib/foods.ts';

	const save = async (input: FoodInput): Promise<void> => {
		await createFood(input);
		await goto(resolve('/foods'));
	};
</script>

<main>
	<h1>New food</h1>
	<FoodEditor initial={emptyForm()} onsave={save} />
	<p><a href={resolve('/foods')}>Back to foods</a></p>
</main>
