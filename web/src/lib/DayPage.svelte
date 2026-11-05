<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import DayEditor from '#lib/DayEditor.svelte';
	import { addDays, formatDate, isDate, localDate } from '#lib/dates.ts';
	import { toForm } from '#lib/day-form.ts';
	import { type Day, type DayInput, getDay, putDay } from '#lib/days.ts';

	interface Props {
		// The day to show, which must be a real calendar date.
		date: string;
	}

	const { date }: Props = $props();

	const today = localDate(new Date());
	const day: Promise<Day> = $derived(getDay(date));

	// How many days the buttons go back or forward.
	const STEP = 1;

	const previous = $derived(addDays(date, -STEP));
	const next = $derived(addDays(date, STEP));

	const goToDate = async (event: Event & { currentTarget: HTMLInputElement }): Promise<void> => {
		// The field is empty while the user is still typing a date.
		const chosen = event.currentTarget.value;
		if (isDate(chosen)) {
			await goto(resolve('/days/[date]', { date: chosen }));
		}
	};

	const save = async (input: DayInput): Promise<void> => {
		await putDay(date, input);
	};
</script>

<main>
	<h1>{formatDate(date)}</h1>

	<nav aria-label="Days">
		<a class="button secondary" href={resolve('/days/[date]', { date: previous })}>‹ Previous</a>
		<input aria-label="Date" onchange={goToDate} type="date" value={date} />
		<a class="button secondary" href={resolve('/days/[date]', { date: next })}>Next ›</a>
	</nav>
	{#if date !== today}
		<p><a href={resolve('/')}>Back to today</a></p>
	{/if}

	<section>
		<h2>Body</h2>
		{#key date}
			{#await day}
				<p class="muted">Loading…</p>
			{:then loaded}
				<DayEditor initial={toForm(loaded)} onsave={save} />
			{:catch error}
				<p role="alert">Loading the day failed: {String(error)}</p>
			{/await}
		{/key}
	</section>

	<section>
		<h2>Manage</h2>
		<ul>
			<li><a href={resolve('/foods')}>Foods</a></li>
			<li><a href={resolve('/nutrients')}>Nutrients</a></li>
			<li><a href={resolve('/recipes')}>Recipes</a></li>
		</ul>
	</section>
</main>

<style>
	/* The date fills the space between the buttons. */
	nav {
		display: flex;
		gap: 0.5rem;
	}

	nav input {
		flex: 1;
		min-width: 0;
	}

	section {
		margin-top: 1.5rem;
	}

	h2 {
		margin: 0 0 0.75rem;
		font-size: 1.125rem;
	}
</style>
