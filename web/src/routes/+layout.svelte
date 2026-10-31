<script lang="ts">
	import { onMount } from 'svelte';
	import favicon from '#lib/assets/favicon.svg';
	import { initAuth, session, signIn, signOut } from '#lib/auth.svelte.ts';
	import type { LayoutProps } from './$types';

	const { children }: LayoutProps = $props();

	onMount(() => {
		initAuth().catch((error: unknown) => {
			session.error = String(error);
		});
	});
</script>

<svelte:head>
	<link href={favicon} rel="icon" />
</svelte:head>

{#if session.error !== undefined}
	<p role="alert">{session.error}</p>
{/if}

{#if session.user === undefined}
	<p>Loading…</p>
{:else if session.user === null}
	<main>
		<h1>Diary</h1>
		<button onclick={signIn} type="button">Sign in with Google</button>
	</main>
{:else}
	<header>
		<span>{session.user.email}</span>
		<button onclick={signOut} type="button">Sign out</button>
	</header>
	{@render children()}
{/if}
