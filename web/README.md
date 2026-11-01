# web

The diary app's front end: a [SvelteKit](https://svelte.dev/docs/kit) app
written in TypeScript and built as a single-page app with
`@sveltejs/adapter-static`.

## Development

Requires the Node.js version in `.nvmrc` (or later) and the pnpm version in
the `packageManager` field of `package.json`, e.g. installed with
`npm install --global pnpm@<version>`.

```bash
pnpm install   # install dependencies
pnpm dev       # start a dev server on http://localhost:5173 (see below)
pnpm check     # type-check
pnpm lint      # check formatting (Prettier) and lint (ESLint)
pnpm test      # run the tests (Vitest)
pnpm format    # fix formatting
pnpm build     # build the app into build/
```

In development, the dev server forwards requests under `/api` to the Go server
on http://localhost:8080 (see the [root README](../README.md)), and requests
under `/__/` (the Firebase configuration and sign-in pages) to the deployed
Firebase Hosting site. In production, the Go server serves the built app itself,
behind Firebase Hosting.

## Signing in

The app signs in with Google through Firebase Authentication
(`src/lib/auth.svelte.ts`), using a redirect to Google and back. It loads its
Firebase configuration from `/__/firebase/init.json`, which Firebase Hosting
serves, and runs the sign-in flow on the domain that the app is served from.

To sign in on the dev server, the OAuth client used for Google sign-in needs to
accept `localhost` too (see [infra/README.md](../infra/README.md#sign-in)):

- authorized JavaScript origin: `http://localhost:5173`
- authorized redirect URI: `http://localhost:5173/__/auth/handler`
