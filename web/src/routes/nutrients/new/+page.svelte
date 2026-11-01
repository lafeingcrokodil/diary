<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import NutrientEditor from '#lib/NutrientEditor.svelte';
	import { emptyForm } from '#lib/nutrient-form.ts';
	import { type NutrientInput, createNutrient } from '#lib/nutrients.ts';

	const save = async (input: NutrientInput): Promise<void> => {
		await createNutrient(input);
		await goto(resolve('/nutrients'));
	};
</script>

<main>
	<h1>New nutrient</h1>
	<NutrientEditor initial={emptyForm()} onsave={save} />
	<p><a href={resolve('/nutrients')}>Back to nutrients</a></p>
</main>
