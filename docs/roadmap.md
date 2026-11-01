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
  hosting, no CORS. Firebase Hosting serves it on a custom domain by forwarding
  all requests to Cloud Run.
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
- [x] Artifact Registry and Cloud Run
- [x] Continuous deployment from `main` via Workload Identity Federation
- [x] Budget alert

### M2 – Authentication

- [x] Firebase project, with Hosting serving the app on `diary.lafeingcrokodil.com`
- [x] Firebase Authentication with the Google provider (Terraform, plus a
      one-off step in the Firebase console)
- [x] Front-end sign-in flow
- [x] Go middleware verifying ID tokens against an email allowlist
- [x] Front end sending the ID token to the API, and showing whether the
      account is allowed

### M3 – Foods

- [x] Domain types and storage interface (in-memory implementation)
- [x] CRUD API
- [x] Firestore implementation, tested against the emulator
- [x] List and edit UI (calories and serving sizes only)

### M4 – Nutrients

Each user manages their own list of nutrients, like foods. A nutrient has a
name, a unit and an optional energy content (kcal per gram). The energy
content is only used to convert energy-based targets into amounts; a food's
energy always comes from its own kcal value, so overlapping nutrients (such as
sugar and carbohydrates) aren't counted twice.

- [x] Domain types and storage interface (in-memory and Firestore)
- [x] CRUD API, with archiving and permanent deletion
- [x] Nutrients page, with an archive view
- [x] Nutrient content on foods (API and food editor), with deleting a
      nutrient also removing its values from foods
- [x] Archiving and permanent deletion for foods

### M5 – Recipes

A recipe has a name, an optional source URL, ingredients, steps and optional
notes. Each ingredient is one of the user's own foods (no sub-recipes or
free-text ingredients for now), with an optional note, and an amount either in
the food's unit or as a number of one of its servings. A recipe's yield is the
whole recipe: the diary will log a fraction of it (e.g. 0.5 for half a pot of
soup).

A recipe's nutrition is computed from its foods' current values, so correcting
a food corrects every recipe that uses it. A nutrient whose amount is unknown
for some ingredient is shown as incomplete, not counted as zero. So that
recipes' amounts keep their meaning, a food that a recipe uses can't be
deleted, have its unit changed, or have a serving that the recipe uses renamed
or removed.

- [x] Domain types and storage interface (in-memory implementation)
- [x] Firestore implementation, tested against the emulator
- [ ] CRUD API, with archiving and permanent deletion
- [ ] Computed nutrition (domain function and API)
- [ ] Protecting foods that recipes use (deletion, unit changes, servings)
- [ ] Recipe list and editor UI
- [ ] Nutrition in the recipe UI

### M6 – Diary

- [ ] Daily log with body weight and energy expended
- [ ] Diary entries and daily totals

### M7 – Targets

- [ ] Per-nutrient ranges: up to six optional targets (min/max per day, as a
      percentage of energy expended, and per kg body weight). The effective
      range is the highest minimum to the lowest maximum; targets whose inputs
      are missing for a day are skipped, and conflicting ranges are shown as a
      warning.
- [ ] Progress against targets in the UI

### M8 – Shopping lists

- [ ] Shopping lists and items

### Later ideas

- Import foods from Open Food Facts / USDA FoodData Central (maybe barcodes)
- Installable PWA
- Labels and search
- Data export and backups
- Default nutrients for new users (fat, carbohydrates, fibre and protein, with
  their kcal per gram)

## Open questions

To be discussed when we reach the relevant milestone:

- Should diary entries be able to reference a single food as well as a recipe?
- Should past diary entries change when a recipe or food is edited, or keep a
  snapshot? (This also decides what permanently deleting a food does to
  diary entries that use it.)
- Where does food data come from (manual entry, imports)?
- Are shopping lists generated from recipes, manual, or both? One list or
  several?
