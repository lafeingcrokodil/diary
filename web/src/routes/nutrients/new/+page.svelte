<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import NutrientEditor from '#lib/NutrientEditor.svelte';
	import { type NutrientFields, emptyForm } from '#lib/nutrient-form.ts';
	import { createNutrient } from '#lib/nutrients.ts';

	const save = async (input: NutrientFields): Promise<void> => {
		await createNutrient({ ...input, archived: false });
		await goto(resolve('/nutrients'));
	};
</script>

<main>
	<h1>New nutrient</h1>
	<NutrientEditor initial={emptyForm()} onsave={save} />
	<p><a href={resolve('/nutrients')}>Back to nutrients</a></p>
</main>
