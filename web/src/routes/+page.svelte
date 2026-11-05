<script lang="ts">
	import DayPage from '#lib/DayPage.svelte';
	import { localDate } from '#lib/dates.ts';
</script>

<DayPage date={localDate(new Date())} />
