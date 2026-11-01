<script lang="ts">
	import { resolve } from '$app/paths';
	import { type FoodForm, parseForm } from '#lib/food-form.ts';
	import type { FoodInput } from '#lib/foods.ts';
	import type { Nutrient } from '#lib/nutrients.ts';

	interface Props {
		// The values to start with.
		initial: FoodForm;
		// All of the user's nutrients. Only active ones get a field, but the
		// amounts of archived ones are kept.
		nutrients: Nutrient[];
		// Whether the food is archived, or undefined for a new food. It decides
		// which buttons there are.
		archived?: boolean;
		// Saves the food. If it throws, the error is shown and the form stays.
		onsave: (input: FoodInput) => Promise<void>;
		// Permanently deletes the food, for archived foods.
		ondelete?: () => Promise<void>;
	}

	const { initial, nutrients, archived, onsave, ondelete }: Props = $props();

	const active = $derived(nutrients.filter((nutrient) => !nutrient.archived));

	// The form starts with the initial values, and then belongs to the user.
	// svelte-ignore state_referenced_locally
	const form = $state<FoodForm>(structuredClone(initial));
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

	// Saves the form, with the food archived or not. Archiving and unarchiving
	// save any changes in the form too.
	const saveAs = async (archive: boolean): Promise<void> => {
		const parsed = parseForm(form, nutrients);
		if (!parsed.ok) {
			({ error } = parsed);
			return;
		}
		await run(async () => onsave({ ...parsed.input, archived: archive }));
	};

	const save = async (event: SubmitEvent): Promise<void> => {
		event.preventDefault();
		await saveAs(archived ?? false);
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

	const addServing = (): void => {
		form.servings.push({ name: '', amount: '' });
	};

	const removeServing = (index: number): void => {
		form.servings.splice(index, 1);
	};
</script>

<form onsubmit={save}>
	<label>
		Name
		<input autocomplete="off" required bind:value={form.name} />
	</label>
	<label>
		Brand (optional)
		<input autocomplete="off" bind:value={form.brand} />
	</label>
	<label>
		Unit
		<select bind:value={form.unit}>
			<option value="g">grams (g)</option>
			<option value="ml">millilitres (ml)</option>
		</select>
	</label>
	<label>
		Energy (kcal per 100 {form.unit})
		<input inputmode="decimal" required bind:value={form.kcalPer100} />
	</label>

	<fieldset>
		<legend>Servings</legend>
		{#each form.servings as serving, index (index)}
			<div class="serving">
				<label>
					Name
					<input autocomplete="off" placeholder="e.g. slice" bind:value={serving.name} />
				</label>
				<label>
					Amount ({form.unit})
					<input inputmode="decimal" bind:value={serving.amount} />
				</label>
				<button
					class="secondary"
					onclick={(): void => {
						removeServing(index);
					}}
					type="button">Remove</button
				>
			</div>
		{/each}
		<button class="secondary" onclick={addServing} type="button">Add a serving</button>
	</fieldset>

	<fieldset>
		<legend>Nutrients per 100 {form.unit}</legend>
		{#each active as nutrient (nutrient.id)}
			<label>
				{nutrient.name} ({nutrient.unit})
				<input inputmode="decimal" placeholder="unknown" bind:value={form.nutrients[nutrient.id]} />
			</label>
		{:else}
			<p class="muted">
				No nutrients yet. <a href={resolve('/nutrients/new')}>Add a nutrient</a> to track it.
			</p>
		{/each}
	</fieldset>

	{#if error !== undefined}
		<p role="alert">{error}</p>
	{/if}

	<div class="actions">
		<button disabled={busy} type="submit">Save</button>
		{#if archived === false}
			<button
				class="secondary"
				disabled={busy}
				onclick={async (): Promise<void> => saveAs(true)}
				type="button">Archive</button
			>
		{:else if archived === true}
			<button
				class="secondary"
				disabled={busy}
				onclick={async (): Promise<void> => saveAs(false)}
				type="button">Unarchive</button
			>
			{#if ondelete !== undefined}
				<button class="danger" disabled={busy} onclick={remove} type="button"
					>Delete permanently</button
				>
			{/if}
		{/if}
	</div>
</form>

<style>
	form {
		display: grid;
		gap: var(--space);
	}

	/* A serving's name and amount side by side, with Remove below them on
	   narrow screens and next to them on wider ones. */
	.serving {
		display: grid;
		grid-template-columns: 2fr 1fr;
		gap: 0.5rem;
		align-items: end;
	}

	.serving button {
		grid-column: 1 / -1;
	}

	@media (min-width: 30rem) {
		.serving {
			grid-template-columns: 2fr 1fr auto;
		}

		.serving button {
			grid-column: auto;
		}
	}

	.actions {
		display: flex;
		flex-wrap: wrap;
		gap: 0.75rem;
	}
</style>
