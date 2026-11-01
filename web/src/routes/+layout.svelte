<script lang="ts">
	import { onMount } from 'svelte';
	import { resolve } from '$app/paths';
	import favicon from '#lib/assets/favicon.svg';
	import { ApiError, getMe } from '#lib/api.ts';
	import { initAuth, session, signIn, signOut } from '#lib/auth.svelte.ts';
	import type { LayoutProps } from './$types';

	const { children }: LayoutProps = $props();

	// Whether the API lets the signed-in user use the app.
	let access = $state<'checking' | 'allowed' | 'denied'>('checking');

	onMount(() => {
		initAuth().catch((error: unknown) => {
			session.error = String(error);
		});
	});

	// Checks with the API whenever someone signs in.
	$effect(() => {
		if (!session.user) {
			return;
		}
		access = 'checking';
		getMe().then(
			() => {
				access = 'allowed';
			},
			(error: unknown) => {
				if (error instanceof ApiError && error.isForbidden) {
					access = 'denied';
				} else {
					session.error = `Checking access failed: ${String(error)}`;
				}
			}
		);
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
		<a href={resolve('/')}>Diary</a>
		<span>{session.user.email}</span>
		<button onclick={signOut} type="button">Sign out</button>
	</header>
	{#if access === 'allowed'}
		{@render children()}
	{:else if access === 'denied'}
		<main>
			<p>This account isn't allowed to use the app.</p>
		</main>
	{:else}
		<p>Loading…</p>
	{/if}
{/if}
