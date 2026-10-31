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

```bash
make run    # build the front end and start the server on http://localhost:8080
make build  # build the front end and a server binary (bin/diary)
make image  # build the container image (diary), which needs only Docker
make web    # build only the front end
make test   # run the tests
make vet    # run go vet
make lint   # run golangci-lint (see .golangci.yml)
```

`make test`, `make vet` and `make lint` don't need the front end: if it hasn't
been built, they create a placeholder page so that the Go code compiles.

### Working on the front end

For instant reloading while you edit the front end, run the Go server and Vite's
dev server side by side:

```bash
make run              # Go server on http://localhost:8080
```

```bash
cd web && pnpm dev    # front end on http://localhost:5173
```

Then open http://localhost:5173. Vite forwards requests under `/api` to the Go
server.
