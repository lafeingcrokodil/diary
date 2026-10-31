# diary

A personal web app for storing recipes, tracking what I eat (including
nutritional information) and creating shopping lists.

See [docs/roadmap.md](docs/roadmap.md) for the plan.

## Development

Requires:

- the Go version specified in `go.mod`, or later
- [golangci-lint](https://golangci-lint.run/docs/welcome/install/) v2.15.0
  (for `make lint`)

```bash
make test   # run the tests
make vet    # run go vet
make lint   # run golangci-lint (see .golangci.yml)
make run    # start the server on http://localhost:8080
```

Check that it's running:

```bash
curl http://localhost:8080/healthz
```
