import type { Food } from '#lib/foods.ts';
import { readNumber } from '#lib/numbers.ts';
import type { Ingredient, RecipeInput } from '#lib/recipes.ts';

// The values of an ingredient's form fields, as the user entered them.
export interface IngredientRow {
	// The food's label (see FoodChoices).
	food: string;
	amount: string;
	// The name of one of the food's servings, or empty for the food's unit.
	serving: string;
	note: string;
}

// The values of a recipe's form fields, as the user entered them.
export interface RecipeForm {
	name: string;
	sourceUrl: string;
	ingredients: IngredientRow[];
	steps: string[];
	notes: string;
}

// The fields of a recipe that its form edits.
export type RecipeFields = Omit<RecipeInput, 'archived'>;

// The outcome of reading a form: either a recipe to save, or what's wrong.
export type ParsedForm = { ok: true; input: RecipeFields } | { ok: false; error: string };

// The foods that ingredients can be, with a label for each, which is what the
// user picks and sees. Labels are unique, so that each one names one food.
export class FoodChoices {
	readonly #byLabel = new Map<string, Food>();
	readonly #labels = new Map<string, string>();

	// The given foods must be all of the user's foods, archived or not, sorted
	// by name.
	public constructor(foods: readonly Food[]) {
		for (const food of foods) {
			const base = food.brand === '' ? food.name : `${food.name} (${food.brand})`;
			let label = base;
			for (let count = 2; this.#byLabel.has(label); count += 1) {
				label = `${base} [${String(count)}]`;
			}
			this.#byLabel.set(label, food);
			this.#labels.set(food.id, label);
		}
	}

	// Returns the food with the given label, if any.
	public find(label: string): Food | undefined {
		return this.#byLabel.get(label.trim());
	}

	// Returns the label of the food with the given ID, or an empty string if
	// there's no such food.
	public labelOf(foodId: string): string {
		return this.#labels.get(foodId) ?? '';
	}

	// Returns the labels of the foods that aren't archived, which are the ones
	// to offer for new ingredients.
	public activeLabels(): string[] {
		return [...this.#byLabel].filter(([, food]) => !food.archived).map(([label]) => label);
	}
}

// Returns an empty ingredient row.
export const emptyIngredient = (): IngredientRow => ({
	food: '',
	amount: '',
	serving: '',
	note: ''
});

// Returns the form for a new recipe.
export const emptyForm = (): RecipeForm => ({
	name: '',
	sourceUrl: '',
	ingredients: [emptyIngredient()],
	steps: [''],
	notes: ''
});

// Returns a form with a recipe's current values.
export const toForm = (input: RecipeFields, choices: FoodChoices): RecipeForm => ({
	name: input.name,
	sourceUrl: input.sourceUrl,
	ingredients: input.ingredients.map((ingredient) => ({
		food: choices.labelOf(ingredient.foodId),
		amount: String(ingredient.amount),
		serving: ingredient.serving,
		note: ingredient.note
	})),
	steps: [...input.steps],
	notes: input.notes
});

// Reads the source URL, or returns an error message.
const readSourceUrl = (text: string): { sourceUrl: string } | string => {
	const sourceUrl = text.trim();
	if (sourceUrl === '') {
		return { sourceUrl };
	}
	const parsed = URL.parse(sourceUrl);
	if (parsed === null || (parsed.protocol !== 'http:' && parsed.protocol !== 'https:')) {
		return 'The source must be a web address that starts with http:// or https://.';
	}
	return { sourceUrl };
};

// Returns the food with the given label, or an error message for the
// ingredient called name.
const findFood = (label: string, name: string, choices: FoodChoices): Food | string => {
	if (label === '') {
		return `${name} needs a food.`;
	}
	return (
		choices.find(label) ?? `${name}: there's no food called "${label}". Choose one from the list.`
	);
};

// Whether a row was left empty. Its serving doesn't count, since there's always
// one selected.
const isEmpty = (row: IngredientRow): boolean =>
	row.food.trim() === '' && row.amount.trim() === '' && row.note.trim() === '';

// Reads the amount of food in a row, or returns an error message for the
// ingredient called name.
const readAmount = (
	row: IngredientRow,
	food: Food,
	name: string
): Pick<Ingredient, 'serving' | 'amount'> | string => {
	if (row.serving !== '' && !food.servings.some((serving) => serving.name === row.serving)) {
		return `${name}: ${food.name} has no serving called "${row.serving}".`;
	}
	const amount = readNumber(row.amount);
	if (amount === undefined || amount <= 0) {
		return `${name} needs an amount greater than 0.`;
	}
	return { serving: row.serving, amount };
};

// Reads one ingredient row: undefined if it was left empty, or an error
// message if it's invalid. position counts from 1.
const readIngredient = (
	row: IngredientRow,
	position: number,
	choices: FoodChoices
): Ingredient | string | undefined => {
	if (isEmpty(row)) {
		return undefined;
	}
	const name = `Ingredient ${String(position)}`;
	const food = findFood(row.food.trim(), name, choices);
	if (typeof food === 'string') {
		return food;
	}
	const amount = readAmount(row, food, name);
	if (typeof amount === 'string') {
		return amount;
	}
	return { foodId: food.id, ...amount, note: row.note.trim() };
};

// Reads the ingredient rows, ignoring empty ones, or returns an error message.
const readIngredients = (form: RecipeForm, choices: FoodChoices): Ingredient[] | string => {
	const ingredients: Ingredient[] = [];
	for (const [index, row] of form.ingredients.entries()) {
		const ingredient = readIngredient(row, index + 1, choices);
		if (typeof ingredient === 'string') {
			return ingredient;
		}
		if (ingredient !== undefined) {
			ingredients.push(ingredient);
		}
	}
	return ingredients;
};

// Checks a form, with the same rules as the API, and returns the recipe it
// describes. choices must have all of the user's foods, archived or not.
// Empty ingredient rows and steps are left out, and numbers may use a decimal
// comma.
export const parseForm = (form: RecipeForm, choices: FoodChoices): ParsedForm => {
	const name = form.name.trim();
	if (name === '') {
		return { ok: false, error: 'The name is required.' };
	}
	const source = readSourceUrl(form.sourceUrl);
	if (typeof source === 'string') {
		return { ok: false, error: source };
	}
	const ingredients = readIngredients(form, choices);
	if (typeof ingredients === 'string') {
		return { ok: false, error: ingredients };
	}
	return {
		ok: true,
		input: {
			name,
			...source,
			ingredients,
			steps: form.steps.map((step) => step.trim()).filter((step) => step !== ''),
			notes: form.notes.trim()
		}
	};
};

// Swaps two items, so that rows can be moved up and down, unless either index
// is out of range.
export const swap = (items: unknown[], first: number, second: number): void => {
	const inRange = (index: number): boolean =>
		Number.isInteger(index) && index >= 0 && index < items.length;
	if (!inRange(first) || !inRange(second)) {
		return;
	}
	[items[first], items[second]] = [items[second], items[first]];
};
