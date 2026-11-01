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
			<div>
				<label>
					Name
					<input autocomplete="off" placeholder="e.g. slice" bind:value={serving.name} />
				</label>
				<label>
					Amount ({form.unit})
					<input inputmode="decimal" bind:value={serving.amount} />
				</label>
				<button
					onclick={(): void => {
						removeServing(index);
					}}
					type="button">Remove</button
				>
			</div>
		{/each}
		<button onclick={addServing} type="button">Add a serving</button>
	</fieldset>

	{#if error !== undefined}
		<p role="alert">{error}</p>
	{/if}

	<button disabled={busy} type="submit">Save</button>
	{#if ondelete !== undefined}
		<button disabled={busy} onclick={remove} type="button">Delete</button>
	{/if}
</form>
