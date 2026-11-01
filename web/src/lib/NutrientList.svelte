<script lang="ts">
	import { resolve } from '$app/paths';
	import type { Nutrient } from '#lib/nutrients.ts';

	interface Props {
		nutrients: Nutrient[];
	}

	const { nutrients }: Props = $props();

	// Describes a nutrient's unit and energy, e.g. "g · 4 kcal/g".
	const details = (nutrient: Nutrient): string =>
		nutrient.kcalPerGram === 0
			? nutrient.unit
			: `${nutrient.unit} · ${String(nutrient.kcalPerGram)} kcal/g`;
</script>

<ul>
	{#each nutrients as nutrient (nutrient.id)}
		<li>
			<a href={resolve('/nutrients/[id]', { id: nutrient.id })}>
				<span class="name">{nutrient.name}</span>
				<span class="muted">{details(nutrient)}</span>
			</a>
		</li>
	{/each}
</ul>

<style>
	ul {
		padding: 0;
		margin: 0;
		list-style: none;
		background: var(--surface);
		border: 1px solid var(--border);
		border-radius: var(--radius);
	}

	li + li {
		border-top: 1px solid var(--border);
	}

	/* The whole row is the link, so it's easy to tap. */
	li a {
		display: grid;
		padding: 0.75rem var(--space);
		color: inherit;
		text-decoration: none;
	}

	.name {
		font-weight: 600;
	}

	.muted {
		font-size: 0.875rem;
	}
</style>
