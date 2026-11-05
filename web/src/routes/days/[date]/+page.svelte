<script lang="ts">
	import { resolve } from '$app/paths';
	import { page } from '$app/state';
	import DayPage from '#lib/DayPage.svelte';
	import { isDate } from '#lib/dates.ts';

	const date = $derived(page.params.date ?? '');
</script>

{#if isDate(date)}
	<DayPage {date} />
{:else}
	<main>
		<p role="alert">"{date}" isn't a date.</p>
		<p><a href={resolve('/')}>Go to today</a></p>
	</main>
{/if}
