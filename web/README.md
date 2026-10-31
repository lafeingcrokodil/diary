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
pnpm dev       # start a dev server on http://localhost:5173
pnpm check     # type-check
pnpm lint      # check formatting (Prettier) and lint (ESLint)
pnpm format    # fix formatting
pnpm build     # build the app into build/
```
