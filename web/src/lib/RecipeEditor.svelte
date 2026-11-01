<script lang="ts">
	import { resolve } from '$app/paths';
	import type { Food } from '#lib/foods.ts';
	import {
		FoodChoices,
		type IngredientRow,
		type RecipeForm,
		emptyIngredient,
		parseForm,
		swap
	} from '#lib/recipe-form.ts';
	import type { RecipeInput } from '#lib/recipes.ts';
	import RowActions from '#lib/RowActions.svelte';

	interface Props {
		// The values to start with.
		initial: RecipeForm;
		// All of the user's foods. Only active ones are offered, but ingredients
		// can keep using archived ones.
		foods: Food[];
		// Whether the recipe is archived, or undefined for a new recipe. It
		// decides which buttons there are.
		archived?: boolean;
		// Saves the recipe. If it throws, the error is shown and the form stays.
		onsave: (input: RecipeInput) => Promise<void>;
		// Permanently deletes the recipe, for archived recipes.
		ondelete?: () => Promise<void>;
	}

	const { initial, foods, archived, onsave, ondelete }: Props = $props();

	const choices = $derived(new FoodChoices(foods));

	// Each editor's list of food labels needs its own ID, in case there's more
	// than one editor on a page.
	const listId = $props.id();

	// The form starts with the initial values, and then belongs to the user.
	// svelte-ignore state_referenced_locally
	const form = $state<RecipeForm>(structuredClone(initial));
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

	// Saves the form, with the recipe archived or not. Archiving and
	// unarchiving save any changes in the form too.
	const saveAs = async (archive: boolean): Promise<void> => {
		const parsed = parseForm(form, choices);
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

	// Switches a row back to its food's unit when its new food doesn't have
	// the serving that was selected.
	const changeFood = (row: IngredientRow): void => {
		const food = choices.find(row.food);
		if (!(food?.servings.some((serving) => serving.name === row.serving) ?? false)) {
			row.serving = '';
		}
	};
</script>

<form onsubmit={save}>
	<label>
		Name
		<input autocomplete="off" required bind:value={form.name} />
	</label>
	<label>
		Source (optional)
		<input autocomplete="off" placeholder="https://…" type="url" bind:value={form.sourceUrl} />
	</label>

	<fieldset>
		<legend>Ingredients</legend>
		<datalist id={listId}>
			{#each choices.activeLabels() as label (label)}
				<option value={label}></option>
			{/each}
		</datalist>
		{#if foods.length === 0}
			<p class="muted">
				No foods yet. <a href={resolve('/foods/new')}>Add a food</a> to use it in recipes.
			</p>
		{/if}
		{#each form.ingredients as row, index (index)}
			{const food = $derived(choices.find(row.food))}
			<div class="row">
				<p class="row-title">Ingredient {index + 1}</p>
				<label class="wide">
					Food
					<input
						autocomplete="off"
						list={listId}
						onchange={(): void => {
							changeFood(row);
						}}
						placeholder="Start typing a food"
						bind:value={row.food}
					/>
				</label>
				<label>
					Amount
					<input inputmode="decimal" bind:value={row.amount} />
				</label>
				<label>
					Unit
					<select disabled={food === undefined} bind:value={row.serving}>
						<option value="">{food?.unit ?? '–'}</option>
						{#each food?.servings ?? [] as serving (serving.name)}
							<option value={serving.name}>{serving.name} ({serving.amount} {food?.unit})</option>
						{/each}
					</select>
				</label>
				<label class="wide">
					Note (optional)
					<input autocomplete="off" placeholder="e.g. finely chopped" bind:value={row.note} />
				</label>
				<div class="wide">
					<RowActions
						name="ingredient {index + 1}"
						first={index === 0}
						last={index === form.ingredients.length - 1}
						ondown={(): void => {
							swap(form.ingredients, index, index + 1);
						}}
						onremove={(): void => {
							form.ingredients.splice(index, 1);
						}}
						onup={(): void => {
							swap(form.ingredients, index, index - 1);
						}}
					/>
				</div>
			</div>
		{/each}
		<button
			class="secondary"
			onclick={(): void => {
				form.ingredients.push(emptyIngredient());
			}}
			type="button">Add an ingredient</button
		>
	</fieldset>

	<fieldset>
		<legend>Steps</legend>
		{#each form.steps, index (index)}
			<div class="row">
				<label class="wide">
					Step {index + 1}
					<textarea rows="3" bind:value={form.steps[index]}></textarea>
				</label>
				<div class="wide">
					<RowActions
						name="step {index + 1}"
						first={index === 0}
						last={index === form.steps.length - 1}
						ondown={(): void => {
							swap(form.steps, index, index + 1);
						}}
						onremove={(): void => {
							form.steps.splice(index, 1);
						}}
						onup={(): void => {
							swap(form.steps, index, index - 1);
						}}
					/>
				</div>
			</div>
		{/each}
		<button
			class="secondary"
			onclick={(): void => {
				form.steps.push('');
			}}
			type="button">Add a step</button
		>
	</fieldset>

	<label>
		Notes (optional)
		<textarea rows="3" bind:value={form.notes}></textarea>
	</label>

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

	/* An ingredient or step: its fields, then buttons to move or remove it.
	   An ingredient's amount and unit sit side by side. */
	.row {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 0.5rem;
		padding-bottom: 0.75rem;
		border-bottom: 1px solid var(--border);
	}

	.wide,
	.row-title {
		grid-column: 1 / -1;
	}

	.row-title {
		margin: 0;
		font-weight: 600;
	}

	.actions {
		display: flex;
		flex-wrap: wrap;
		gap: 0.75rem;
	}
</style>
