.PHONY: run build image web test cover vet lint

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

# Builds the container image.
image:
	docker build --tag diary .

web:
	cd web && pnpm install --frozen-lockfile && pnpm build

$(FRONTEND):
	mkdir -p web/build
	echo '<!doctype html><title>Diary</title><p>The front end has not been built. Run make web.</p>' > $@

# The minimum percentage of statements in internal/ that the tests must cover.
# The other packages only wire things together, so they aren't counted.
MIN_COVERAGE := 80

test: $(FRONTEND)
	go test -race -coverprofile=coverage.out ./...

# Runs the tests and fails if coverage of internal/ is below MIN_COVERAGE. Each
# line of the profile after the first is "file:range statements count".
cover: test
	@awk -v min=$(MIN_COVERAGE) ' \
		NR > 1 && $$1 ~ /\/internal\// { total += $$2; if ($$3 > 0) covered += $$2 } \
		END { \
			pct = total ? 100 * covered / total : 0; \
			printf "Coverage of internal/: %.1f%% (minimum %d%%)\n", pct, min; \
			exit pct < min \
		}' coverage.out

vet: $(FRONTEND)
	go vet ./...

lint: $(FRONTEND)
	golangci-lint run ./...
