# Roadmap

A web app for storing recipes, tracking daily food intake (including
nutritional information) and creating shopping lists.

The app is built incrementally: each milestone is split into small PRs that
can each be reviewed comfortably on their own.

## Architecture

- **Back end:** Go, using the standard library's `net/http`. Each area of the
  app (foods, recipes, ...) has a storage interface with an in-memory
  implementation (for unit tests) and a Firestore implementation (tested
  against the Firestore emulator). Switching databases means adding another
  implementation.
- **Front end:** SvelteKit + TypeScript, built as a single-page app with
  `adapter-static`, mobile-first. The built assets are embedded into the Go
  binary with `go:embed`, so the whole app is a single container: no separate
  hosting, no CORS.
- **Auth:** Firebase Authentication with the Google provider. The front end
  obtains an ID token; the Go server verifies it and checks the email against
  an allowlist.
- **Infrastructure:** GCP, managed with Terraform (state in a GCS bucket):
  Cloud Run (scales to zero), Firestore (Native mode), Artifact Registry
  (with a cleanup policy), and Workload Identity Federation so that GitHub
  Actions can deploy without long-lived keys. The aim is to stay within the
  free tier.

## Milestones

### M0 – Walking skeleton

- [x] Go module, `/healthz` endpoint, test, Makefile, roadmap
- [x] GitHub Actions CI (`go vet`, `go test`, `golangci-lint`)
- [x] SvelteKit + TypeScript app scaffold, with linting in CI
- [x] Serve the front end from the Go binary, local dev setup
- [x] Dockerfile

### M1 – Deployed to GCP

- [x] Terraform state bucket and API enablement
- [ ] Artifact Registry and Cloud Run
- [ ] Continuous deployment from `main` via Workload Identity Federation
- [x] Budget alert

### M2 – Authentication

- [ ] Firebase project and Google provider (Terraform)
- [ ] Front-end sign-in flow
- [ ] Go middleware verifying ID tokens against an email allowlist

### M3 – Foods

- [ ] Domain types and storage interface (in-memory implementation)
- [ ] CRUD API
- [ ] Firestore implementation, tested against the emulator
- [ ] List and edit UI (calories and serving sizes only)

### M4 – Nutrients

- [ ] Seeded nutrients with targets
- [ ] Nutrient content on foods

### M5 – Recipes

- [ ] Recipes with ingredients
- [ ] Computed nutrition (decide how recipe yield is defined)

### M6 – Diary

- [ ] Daily log with body weight and energy expended
- [ ] Diary entries and daily totals

### M7 – Targets

- [ ] Per-nutrient ranges (per day, per 1000 kcal, per kg body weight)
- [ ] Progress against targets in the UI

### M8 – Shopping lists

- [ ] Shopping lists and items

### Later ideas

- Import foods from Open Food Facts / USDA FoodData Central (maybe barcodes)
- Installable PWA
- Labels and search
- Data export and backups

## Open questions

To be discussed when we reach the relevant milestone:

- Should diary entries be able to reference a single food as well as a recipe?
- How is a recipe's yield defined (servings, total cooked weight)?
- Should past diary entries change when a recipe is edited, or keep a
  snapshot?
- Where does food data come from (manual entry, imports)?
- Are shopping lists generated from recipes, manual, or both? One list or
  several?
- How are missing body weight / energy expended values handled?
- Which nutrients beyond fat, carbohydrates, fibre and protein should be
  tracked?
