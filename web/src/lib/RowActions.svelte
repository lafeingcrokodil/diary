<script lang="ts">
	interface Props {
		// What the row is, e.g. "ingredient 2", for screen readers.
		name: string;
		// Whether the row can move up or down.
		first: boolean;
		last: boolean;
		onup: () => void;
		ondown: () => void;
		onremove: () => void;
	}

	const { name, first, last, onup, ondown, onremove }: Props = $props();
</script>

<!-- Buttons to move a row of a list up or down, or remove it. -->
<div>
	<button
		class="secondary"
		aria-label="Move {name} up"
		disabled={first}
		onclick={onup}
		type="button">↑</button
	>
	<button
		class="secondary"
		aria-label="Move {name} down"
		disabled={last}
		onclick={ondown}
		type="button">↓</button
	>
	<button class="secondary" aria-label="Remove {name}" onclick={onremove} type="button"
		>Remove</button
	>
</div>

<style>
	div {
		display: flex;
		gap: 0.5rem;
	}
</style>
