<script lang="ts">
	import { type FoodForm, parseForm } from '#lib/food-form.ts';
	import type { FoodInput } from '#lib/foods.ts';

	interface Props {
		// The values to start with.
		initial: FoodForm;
		// Saves the food. If it throws, the error is shown and the form stays.
		onsave: (input: FoodInput) => Promise<void>;
		// Deletes the food, for foods that already exist.
		ondelete?: () => Promise<void>;
	}

	const { initial, onsave, ondelete }: Props = $props();

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
		// eslint-disable-next-line no-alert -- A confirmation dialog is all this needs.
		if (ondelete !== undefined && confirm(`Delete "${form.name}"?`)) {
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

	{#if error !== undefined}
		<p role="alert">{error}</p>
	{/if}

	<div class="actions">
		<button disabled={busy} type="submit">Save</button>
		{#if ondelete !== undefined}
			<button class="danger" disabled={busy} onclick={remove} type="button">Delete</button>
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
		gap: 0.75rem;
		justify-content: space-between;
	}
</style>
