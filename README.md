# diary

A personal web app for storing recipes, tracking what I eat (including
nutritional information) and creating shopping lists.

See [docs/roadmap.md](docs/roadmap.md) for the plan.

## Development

Requires the Go version specified in `go.mod`, or later.

```bash
make test   # run the tests
make vet    # run go vet
make run    # start the server on http://localhost:8080
```

Check that it's running:

```bash
curl http://localhost:8080/healthz
```
