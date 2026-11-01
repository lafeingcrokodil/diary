# diary

A personal web app for storing recipes, tracking what I eat (including
nutritional information) and creating shopping lists.

See [docs/roadmap.md](docs/roadmap.md) for the plan.

## Development

The Go back end lives in the repository root; the SvelteKit front end lives in
[`web/`](web/README.md), and the Terraform configuration for GCP lives in
[`infra/`](infra/README.md). The Go binary embeds the built front end and serves
it, so the deployed app is a single binary.

Requires:

- the Go version specified in `go.mod`, or later
- [golangci-lint](https://golangci-lint.run/docs/welcome/install/) v2.15.0
  (for `make lint`)
- the front end's requirements (see [`web/README.md`](web/README.md)), for the
  targets that build it
- the [gcloud CLI](https://cloud.google.com/sdk/docs/install) with the
  `cloud-firestore-emulator` component, and Java 25 or later, for the Firestore
  tests (see below). Java 21 still works, but the emulator stops supporting it
  after gcloud 576.0.0.

```bash
make run      # build the front end and start the server on http://localhost:8080 (see below)
make build    # build the front end and a server binary (bin/diary)
make image    # build the container image (diary), which needs only Docker
make web      # build only the front end
make test     # run the tests
make cover    # run the tests and check the coverage of each package in internal/ (see Makefile)
make emulator # run the Firestore emulator on localhost:8085
make vet      # run go vet
make lint     # run golangci-lint (see .golangci.yml)
```

The server only lets allowed users use its API, so it needs their
comma-separated email addresses in `ALLOWED_EMAILS`, e.g.
`ALLOWED_EMAILS=you@example.com make run`. It checks sign-ins against the
Firebase project in `FIREBASE_PROJECT_ID`, which `make run` sets to the app's
project unless it's already set.

The server keeps its data in Firestore. `make run` points it to the Firestore
emulator on `localhost:8085` (via `FIRESTORE_EMULATOR_HOST`), so local runs
never touch the real database: start the emulator with `make emulator` first.
The emulator keeps its data in memory, so it starts empty each time.

The Firestore tests run against the Firestore emulator. Start it with
`make emulator`, then run the tests with `FIRESTORE_EMULATOR_HOST` pointing to
it, e.g. `FIRESTORE_EMULATOR_HOST=localhost:8085 make cover`. Without
`FIRESTORE_EMULATOR_HOST`, the Firestore tests are skipped, so `make cover`
fails on their package's coverage.

`make test`, `make vet` and `make lint` don't need the front end: if it hasn't
been built, they create a placeholder page so that the Go code compiles.

### Working on the front end

For instant reloading while you edit the front end, run the Go server and Vite's
dev server side by side:

```bash
make emulator                             # Firestore emulator on localhost:8085
```

```bash
ALLOWED_EMAILS=you@example.com make run   # Go server on http://localhost:8080
```

```bash
cd web && pnpm dev    # front end on http://localhost:5173
```

Then open http://localhost:5173. Vite forwards requests under `/api` to the Go
server, and the Firebase configuration and sign-in pages (under `/__/`) to the
deployed site.
