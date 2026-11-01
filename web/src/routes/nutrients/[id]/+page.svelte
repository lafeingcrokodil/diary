<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { page } from '$app/state';
	import NutrientEditor from '#lib/NutrientEditor.svelte';
	import { type NutrientFields, toForm } from '#lib/nutrient-form.ts';
	import { type Nutrient, deleteNutrient, getNutrient, updateNutrient } from '#lib/nutrients.ts';

	const { id } = page.params;
	const nutrient: Promise<Nutrient> = getNutrient(id ?? '');

	const back = async (): Promise<void> => goto(resolve('/nutrients'));

	// Saving keeps the nutrient archived or not, as it was.
	const save = async (loaded: Nutrient, input: NutrientFields): Promise<void> => {
		await updateNutrient(id ?? '', { ...input, archived: loaded.archived });
		await back();
	};

	// Archives or unarchives the nutrient as it's stored, without any unsaved
	// changes from the form.
	const setArchived = async (loaded: Nutrient, archived: boolean): Promise<void> => {
		await updateNutrient(id ?? '', { ...loaded, archived });
		await back();
	};

	const remove = async (): Promise<void> => {
		await deleteNutrient(id ?? '');
		await back();
	};
</script>

<main>
	{#await nutrient}
		<p class="muted">Loading…</p>
	{:then loaded}
		<h1>{loaded.name}</h1>
		{#if loaded.archived}
			<p class="muted">This nutrient is archived. Unarchive it to use it again.</p>
			<NutrientEditor
				initial={toForm(loaded)}
				ondelete={remove}
				onsave={async (input: NutrientFields): Promise<void> => save(loaded, input)}
				onunarchive={async (): Promise<void> => setArchived(loaded, false)}
			/>
		{:else}
			<NutrientEditor
				initial={toForm(loaded)}
				onarchive={async (): Promise<void> => setArchived(loaded, true)}
				onsave={async (input: NutrientFields): Promise<void> => save(loaded, input)}
			/>
		{/if}
	{:catch error}
		<p role="alert">Loading the nutrient failed: {String(error)}</p>
	{/await}
	<p><a href={resolve('/nutrients')}>Back to nutrients</a></p>
</main>
