<script lang="ts">
	import { type DayForm, parseForm } from '#lib/day-form.ts';
	import type { DayInput } from '#lib/days.ts';

	interface Props {
		// The values to start with.
		initial: DayForm;
		// Saves the day. If it throws, the error is shown and the form stays.
		onsave: (input: DayInput) => Promise<void>;
	}

	const { initial, onsave }: Props = $props();

	// The form starts with the initial values, and then belongs to the user.
	// svelte-ignore state_referenced_locally
	const form = $state<DayForm>(structuredClone(initial));
	let error = $state<string | undefined>(undefined);
	let busy = $state(false);
	// Whether the form is unchanged since it was last saved.
	let saved = $state(false);

	// Saves the form's values, showing what went wrong (if anything) and
	// disabling the button while it runs.
	const saveValues = async (input: DayInput): Promise<void> => {
		busy = true;
		try {
			await onsave(input);
			saved = true;
		} catch (error_) {
			error = String(error_);
		} finally {
			busy = false;
		}
	};

	const save = async (event: SubmitEvent): Promise<void> => {
		event.preventDefault();
		error = undefined;
		saved = false;
		const parsed = parseForm(form);
		if (parsed.ok) {
			await saveValues(parsed.input);
		} else {
			({ error } = parsed);
		}
	};
</script>

<form
	oninput={(): void => {
		saved = false;
	}}
	onsubmit={save}
>
	<div class="fields">
		<label>
			Weight (kg)
			<input inputmode="decimal" bind:value={form.weightKg} />
		</label>
		<label>
			Energy expended (kcal)
			<input inputmode="decimal" bind:value={form.kcalExpended} />
		</label>
	</div>

	{#if error !== undefined}
		<p role="alert">{error}</p>
	{/if}

	<div class="actions">
		<button disabled={busy} type="submit">Save</button>
		{#if saved}
			<span class="muted" role="status">Saved.</span>
		{/if}
	</div>
</form>

<style>
	form {
		display: grid;
		gap: var(--space);
	}

	/* The two fields are short, so they share a row, even on a phone. */
	.fields {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 0.75rem;
		align-items: end;
	}

	.actions {
		display: flex;
		gap: 0.75rem;
		align-items: center;
	}
</style>
