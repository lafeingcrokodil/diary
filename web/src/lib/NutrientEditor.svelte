<script lang="ts">
	import { type NutrientFields, type NutrientForm, parseForm } from '#lib/nutrient-form.ts';
	import { units } from '#lib/nutrients.ts';

	interface Props {
		// The values to start with.
		initial: NutrientForm;
		// Saves the nutrient's fields. If it throws, the error is shown and the
		// form stays.
		onsave: (input: NutrientFields) => Promise<void>;
		// Archives the nutrient, for nutrients that exist and aren't archived.
		onarchive?: () => Promise<void>;
		// Unarchives the nutrient, for archived nutrients.
		onunarchive?: () => Promise<void>;
		// Permanently deletes the nutrient, for archived nutrients.
		ondelete?: () => Promise<void>;
	}

	const { initial, onsave, onarchive, onunarchive, ondelete }: Props = $props();

	// The form starts with the initial values, and then belongs to the user.
	// svelte-ignore state_referenced_locally
	const form = $state<NutrientForm>(structuredClone(initial));
	let error = $state<string | undefined>(undefined);
	let busy = $state(false);

	// Runs an action, showing its error (if any) and disabling the buttons while
	// it runs.
	const run = async (action: () => Promise<void>): Promise<void> => {
		busy = true;
		error = undefined;
		try {
			await action();
		} catch (error_) {
			error = String(error_);
		} finally {
			busy = false;
		}
	};

	const save = async (event: SubmitEvent): Promise<void> => {
		event.preventDefault();
		const parsed = parseForm(form);
		if (!parsed.ok) {
			({ error } = parsed);
			return;
		}
		await run(async () => onsave(parsed.input));
	};

	const remove = async (): Promise<void> => {
		if (
			ondelete !== undefined &&
			// eslint-disable-next-line no-alert -- A confirmation dialog is all this needs.
			confirm(`Permanently delete "${form.name}"? This can't be undone.`)
		) {
			await run(ondelete);
		}
	};
</script>

<form onsubmit={save}>
	<label>
		Name
		<input autocomplete="off" required bind:value={form.name} />
	</label>
	<label>
		Unit
		<select bind:value={form.unit}>
			{#each units as unit (unit)}
				<option value={unit}>{unit}</option>
			{/each}
		</select>
	</label>
	<label>
		Energy (kcal per gram, optional)
		<input inputmode="decimal" placeholder="e.g. 4 for protein" bind:value={form.kcalPerGram} />
	</label>

	{#if error !== undefined}
		<p role="alert">{error}</p>
	{/if}

	<div class="actions">
		<button disabled={busy} type="submit">Save</button>
		{#if onarchive !== undefined}
			<button
				class="secondary"
				disabled={busy}
				onclick={async (): Promise<void> => run(onarchive)}
				type="button">Archive</button
			>
		{/if}
		{#if onunarchive !== undefined}
			<button
				class="secondary"
				disabled={busy}
				onclick={async (): Promise<void> => run(onunarchive)}
				type="button">Unarchive</button
			>
		{/if}
		{#if ondelete !== undefined}
			<button class="danger" disabled={busy} onclick={remove} type="button"
				>Delete permanently</button
			>
		{/if}
	</div>
</form>

<style>
	form {
		display: grid;
		gap: var(--space);
	}

	.actions {
		display: flex;
		flex-wrap: wrap;
		gap: 0.75rem;
	}
</style>
