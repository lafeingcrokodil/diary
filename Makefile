.PHONY: run build web test vet lint

# The Go binary embeds web/build, so the Go targets need that
# directory to exist. If the front end hasn't been built, a placeholder page is
# created so that the Go code still compiles.
FRONTEND := web/build/index.html

# Builds the front end and runs the server on http://localhost:8080.
run: web
	go run ./cmd/diary

# Builds the front end and a server binary that includes it.
build: web
	go build -o bin/diary ./cmd/diary

web:
	cd web && pnpm install --frozen-lockfile && pnpm build

$(FRONTEND):
	mkdir -p web/build
	echo '<!doctype html><title>Diary</title><p>The front end has not been built. Run make web.</p>' > $@

test: $(FRONTEND)
	go test -race ./...

vet: $(FRONTEND)
	go vet ./...

lint: $(FRONTEND)
	golangci-lint run ./...
