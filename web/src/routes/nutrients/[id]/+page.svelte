<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { page } from '$app/state';
	import NutrientEditor from '#lib/NutrientEditor.svelte';
	import { toForm } from '#lib/nutrient-form.ts';
	import {
		type Nutrient,
		type NutrientInput,
		deleteNutrient,
		getNutrient,
		updateNutrient
	} from '#lib/nutrients.ts';

	const { id } = page.params;
	const nutrient: Promise<Nutrient> = getNutrient(id ?? '');

	const save = async (input: NutrientInput): Promise<void> => {
		await updateNutrient(id ?? '', input);
		await goto(resolve('/nutrients'));
	};

	const remove = async (): Promise<void> => {
		await deleteNutrient(id ?? '');
		await goto(resolve('/nutrients'));
	};
</script>

<main>
	{#await nutrient}
		<p class="muted">Loading…</p>
	{:then loaded}
		<h1>{loaded.name}</h1>
		{#if loaded.archived}
			<p class="muted">This nutrient is archived. Unarchive it to use it again.</p>
		{/if}
		<NutrientEditor
			archived={loaded.archived}
			initial={toForm(loaded)}
			ondelete={remove}
			onsave={save}
		/>
	{:catch error}
		<p role="alert">Loading the nutrient failed: {String(error)}</p>
	{/await}
	<p><a href={resolve('/nutrients')}>Back to nutrients</a></p>
</main>
