<script lang="ts">
	import { resolve } from '$app/paths';
	import type { Food } from '#lib/foods.ts';

	interface Props {
		foods: Food[];
	}

	const { foods }: Props = $props();
</script>

<ul>
	{#each foods as food (food.id)}
		<li>
			<a href={resolve('/foods/[id]', { id: food.id })}>
				<span class="name">{food.name}</span>
				<span class="muted">
					{#if food.brand !== ''}{food.brand} ·{/if}
					{food.kcalPer100} kcal per 100 {food.unit}
				</span>
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
