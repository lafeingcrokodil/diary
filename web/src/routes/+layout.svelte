<script lang="ts">
	import '../app.css';
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
	<div class="status">
		<p role="alert">{session.error}</p>
	</div>
{/if}

{#if session.user === undefined}
	<p class="status muted">Loading…</p>
{:else if session.user === null}
	<main>
		<h1>Diary</h1>
		<button onclick={signIn} type="button">Sign in with Google</button>
	</main>
{:else}
	<header>
		<a href={resolve('/')}>Diary</a>
		<span>{session.user.email}</span>
		<button class="secondary" onclick={signOut} type="button">Sign out</button>
	</header>
	{#if access === 'allowed'}
		{@render children()}
	{:else if access === 'denied'}
		<main>
			<p>This account isn't allowed to use the app.</p>
		</main>
	{:else}
		<p class="status muted">Loading…</p>
	{/if}
{/if}

<style>
	/* The bar at the top: the app's name, who's signed in, and signing out. */
	header {
		display: flex;
		gap: 0.75rem;
		align-items: center;
		padding: 0.5rem var(--space);
		background: var(--surface);
		border-bottom: 1px solid var(--border);
	}

	header a {
		font-weight: 600;
		color: var(--text);
		text-decoration: none;
	}

	/* Long email addresses shrink rather than pushing the button off screen. */
	header span {
		flex: 1;
		min-width: 0;
		overflow: hidden;
		font-size: 0.875rem;
		color: var(--muted);
		text-align: right;
		text-overflow: ellipsis;
		white-space: nowrap;
	}
</style>
