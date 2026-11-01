.PHONY: run build image web test cover emulator vet lint

# The Go binary embeds web/build, so the Go targets need that
# directory to exist. If the front end hasn't been built, a placeholder page is
# created so that the Go code still compiles.
FRONTEND := web/build/index.html

# The Firebase project whose users can sign in, unless set otherwise. The
# server also requires ALLOWED_EMAILS, which has no default.
export FIREBASE_PROJECT_ID ?= lafeingcrokodil-diary

# Builds the front end and runs the server on http://localhost:8080. Unless
# FIRESTORE_EMULATOR_HOST is set otherwise, the server uses the emulator that
# `make emulator` runs, so that local runs don't touch the real database.
run: export FIRESTORE_EMULATOR_HOST ?= localhost:8085
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

# The minimum percentage of statements that the tests must cover in each
# package in internal/. The other packages only wire things together, so they
# aren't checked.
MIN_COVERAGE := 80

# Test helper packages, which only other packages' tests run, so their coverage
# isn't checked. Keep in sync with the exclusions in .golangci.yml.
TEST_HELPERS := internal/food/foodtest internal/nutrient/nutrienttest internal/recipe/recipetest

test: $(FRONTEND)
	go test -race -coverprofile=coverage.out ./...

# The module's path, which starts every package path in the coverage profile.
MODULE := $(shell go list -m)

# Runs the tests and fails if any package in internal/ has less than
# MIN_COVERAGE coverage. Each line of the profile after the first is
# "file:range statements count", where file is a path that starts with the
# module's path, and a file's package is its directory.
cover: test
	@awk -v min=$(MIN_COVERAGE) -v module=$(MODULE) -v helpers="$(TEST_HELPERS)" ' \
		BEGIN { \
			n = split(helpers, list, " "); \
			for (i = 1; i <= n; i++) helper[list[i]] = 1 \
		} \
		NR > 1 { \
			pkg = $$1; \
			sub(/\/[^\/]*$$/, "", pkg); \
			pkg = substr(pkg, length(module) + 2); \
			if (pkg !~ /^internal\// || pkg in helper) next; \
			total[pkg] += $$2; \
			if ($$3 > 0) covered[pkg] += $$2 \
		} \
		END { \
			for (pkg in total) { \
				pct = total[pkg] ? 100 * covered[pkg] / total[pkg] : 100; \
				printf "%s: %.1f%% (minimum %d%%)\n", pkg, pct, min; \
				if (pct < min) failed = 1 \
			} \
			exit failed \
		}' coverage.out

# Runs the Firestore emulator, which the Firestore tests need. Run the tests
# with FIRESTORE_EMULATOR_HOST=localhost:8085 make test; without it, they fail
# rather than being skipped, so that they can't silently stop running.
emulator:
	gcloud emulators firestore start --host-port=localhost:8085

vet: $(FRONTEND)
	go vet ./...

lint: $(FRONTEND)
	golangci-lint run ./...
