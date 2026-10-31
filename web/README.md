# web

The diary app's front end: a [SvelteKit](https://svelte.dev/docs/kit) app
written in TypeScript and built as a single-page app with
`@sveltejs/adapter-static`.

## Development

Requires the Node.js version in `.nvmrc` (or later) and pnpm (the version is
pinned in `package.json`; with [Corepack](https://pnpm.io/installation#using-corepack)
enabled, it's picked up automatically).

```bash
pnpm install   # install dependencies
pnpm dev       # start a dev server on http://localhost:5173 (see below)
pnpm check     # type-check
pnpm lint      # check formatting (Prettier) and lint (ESLint)
pnpm format    # fix formatting
pnpm build     # build the app into build/
```

In development, the dev server forwards requests under `/api` to the Go server
on http://localhost:8080 (see the [root README](../README.md)). In production,
the Go server serves the built app itself.
