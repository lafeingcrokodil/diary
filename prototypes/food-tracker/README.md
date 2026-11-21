# Food tracker prototype

A single-page prototype of the food diary, built as a Claude artifact, to try
out the roadmap's features on a phone before they're built into the app.
Everything is in [`index.html`](index.html): markup, styles and script, with no
build step.

It covers:

- **Nutrients:** the user's own list (name, unit, optional kcal per gram),
  starting with fat, carbohydrates, fibre and protein, with archiving and
  deletion. Deleting a nutrient removes its values from foods.
- **Foods:** unit (g or ml), kcal per 100, named servings, nutrient values per
  100 (blank means unknown), an optional barcode and a description or
  ingredient list, with archiving and deletion.
- **Recipes:** source URL, ingredients (foods, in the food's unit or as
  servings, with notes), steps and notes, with nutrition computed from the
  current foods. Recipes open in a read-only view for cooking, with larger
  text and ingredients and steps that can be ticked off, and an editor behind
  an Edit button.
- **Diary:** per day, body weight and energy expended (both optional) and a
  flat list of entries, each a food or a fraction of a recipe, with an
  optional time. Totals are computed from the current foods and recipes, and
  shown as incomplete when a value is unknown.
- **Trends:** over the last 7, 30 or 90 days, net energy per day (eaten minus
  expended, coloured by whether energy eaten was within its target) and body
  weight, with the change over the period.
- **Targets (M7):** up to six per nutrient and for energy (min/max per day, as
  a percentage of energy expended, and per kg body weight). The effective range
  is the highest minimum to the lowest maximum; targets whose inputs are
  missing are skipped, and conflicts are shown as a warning.
- **Shopping lists (M8):** several lists, with typed items and a recipe's
  ingredients (scaled, and combined with matching unticked items).
- Foods and recipes that recipes or diary entries use can be archived but not
  deleted, and a used food's unit can't be changed nor its used servings
  removed. Renaming a used serving updates the recipes and diary entries that
  use it.

Beyond the roadmap:

- A built-in table of about 100 common foods (typical values per 100 g), which
  are copied into the user's foods when first logged.
- An estimate of energy expended from age, height, weight, steps and
  bouldering, used when none is entered for a day.
- Per-kg targets use the latest weight logged on or before the day, not only
  that day's weight.
- In the food editor, Claude can estimate blank values (energy, a portion
  weight and nutrients) from the name, brand or restaurant, description or
  ingredient list, and photos that aren't saved. Estimated values are marked
  until they're edited.
- A backup of all the user's data as a JSON file in a "Diary Backups" folder
  in their Google Drive.

## Running it

The page expects the artifact runtime's `window.claude.use()`: `db` and `user`
for storage per user, `sample` for Claude's estimates, and `mcp` with the
Google Drive connector for backups. Opened as a plain file, it falls back to
the browser's `localStorage` and hides the Claude and Google Drive features.
