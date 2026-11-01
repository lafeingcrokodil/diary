import { apiFetch } from '#lib/api.ts';

// An amount of one of the user's foods in a recipe.
export interface Ingredient {
	foodId: string;
	// The name of one of the food's servings, or empty if the amount is in the
	// food's unit.
	serving: string;
	// The number of servings if serving is set, and otherwise the amount in the
	// food's unit.
	amount: number;
	// Optional, e.g. "finely chopped".
	note: string;
}

// A recipe as the user enters it. Its yield is the whole recipe.
export interface RecipeInput {
	name: string;
	// Optional: an http or https URL that the recipe comes from.
	sourceUrl: string;
	ingredients: Ingredient[];
	steps: string[];
	// Optional, free-form text.
	notes: string;
	// Archived recipes are kept, but are hidden until they're unarchived.
	archived: boolean;
}

// A total over a recipe's ingredients.
export interface Total {
	// The sum over the ingredients whose amount is known.
	amount: number;
	// The positions (from 0) of the ingredients whose amount isn't known, so
	// that amount is only a lower bound if there are any.
	missing: number[];
}

// The nutritional information of a whole recipe, worked out from its foods'
// current values.
export interface Nutrition {
	kcal: Total;
	// Totals by nutrient ID, in each nutrient's unit, for the nutrients whose
	// amount is known for at least one ingredient.
	nutrients: Record<string, Total>;
}

// A stored recipe, with its nutrition.
export interface Recipe extends RecipeInput {
	id: string;
	nutrition: Nutrition;
}

const isObject = (value: unknown): value is object =>
	typeof value === 'object' && value !== null && !Array.isArray(value);

const isIngredient = (value: unknown): value is Ingredient =>
	isObject(value) &&
	'foodId' in value &&
	typeof value.foodId === 'string' &&
	'serving' in value &&
	typeof value.serving === 'string' &&
	'amount' in value &&
	typeof value.amount === 'number' &&
	'note' in value &&
	typeof value.note === 'string';

const isTotal = (value: unknown): value is Total =>
	isObject(value) &&
	'amount' in value &&
	typeof value.amount === 'number' &&
	'missing' in value &&
	Array.isArray(value.missing) &&
	value.missing.every((index) => typeof index === 'number');

const isNutrition = (value: unknown): value is Nutrition =>
	isObject(value) &&
	'kcal' in value &&
	isTotal(value.kcal) &&
	'nutrients' in value &&
	isObject(value.nutrients) &&
	Object.values(value.nutrients).every(isTotal);

const isRecipe = (value: unknown): value is Recipe =>
	isObject(value) &&
	'id' in value &&
	typeof value.id === 'string' &&
	'name' in value &&
	typeof value.name === 'string' &&
	'sourceUrl' in value &&
	typeof value.sourceUrl === 'string' &&
	'ingredients' in value &&
	Array.isArray(value.ingredients) &&
	value.ingredients.every(isIngredient) &&
	'steps' in value &&
	Array.isArray(value.steps) &&
	value.steps.every((step) => typeof step === 'string') &&
	'notes' in value &&
	typeof value.notes === 'string' &&
	'archived' in value &&
	typeof value.archived === 'boolean' &&
	'nutrition' in value &&
	isNutrition(value.nutrition);

const readRecipe = async (response: Response): Promise<Recipe> => {
	const value: unknown = await response.json();
	if (!isRecipe(value)) {
		throw new Error('The API responded with an invalid recipe.');
	}
	return value;
};

const recipePath = (id: string): string => `/api/recipes/${encodeURIComponent(id)}`;

// Only the input's own fields are sent, since the API rejects unknown ones
// (such as a stored recipe's nutrition).
const jsonRequest = (method: string, input: RecipeInput): RequestInit => {
	const { name, sourceUrl, ingredients, steps, notes, archived } = input;
	return {
		method,
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({ name, sourceUrl, ingredients, steps, notes, archived })
	};
};

// Returns all of the signed-in user's recipes, including archived ones, sorted
// by name.
export const listRecipes = async (): Promise<Recipe[]> => {
	const response = await apiFetch('/api/recipes');
	const value: unknown = await response.json();
	if (!Array.isArray(value) || !value.every(isRecipe)) {
		throw new Error('The API responded with an invalid list of recipes.');
	}
	return value;
};

// Returns one of the signed-in user's recipes.
export const getRecipe = async (id: string): Promise<Recipe> =>
	readRecipe(await apiFetch(recipePath(id)));

// Stores a new recipe, and returns it with its ID.
export const createRecipe = async (input: RecipeInput): Promise<Recipe> =>
	readRecipe(await apiFetch('/api/recipes', jsonRequest('POST', input)));

// Replaces one of the signed-in user's recipes, and returns it as stored.
// Archiving and unarchiving a recipe are updates of its archived field.
export const updateRecipe = async (id: string, input: RecipeInput): Promise<Recipe> =>
	readRecipe(await apiFetch(recipePath(id), jsonRequest('PUT', input)));

// Permanently deletes one of the signed-in user's recipes. Only archived
// recipes can be deleted.
export const deleteRecipe = async (id: string): Promise<void> => {
	await apiFetch(recipePath(id), { method: 'DELETE' });
};
