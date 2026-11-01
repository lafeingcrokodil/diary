import type { Food } from '#lib/foods.ts';
import type { Nutrient } from '#lib/nutrients.ts';
import type { Recipe, Total } from '#lib/recipes.ts';

// One line of a recipe's nutrition: energy, or one of the user's nutrients.
export interface NutritionRow {
	name: string;
	unit: string;
	// The total, or undefined if it isn't known for any ingredient.
	total: Total | undefined;
	// The names of the foods of the ingredients whose amount isn't known, so
	// that the total is only a lower bound if there are any.
	missing: string[];
}

// The decimal places to show: amounts under this get one, and larger ones get
// none.
const SMALL_AMOUNT = 10;

// Formats an amount for display, with fewer decimal places the larger it is.
// locale is the browser's if it's undefined.
export const formatAmount = (amount: number, locale?: string): string =>
	new Intl.NumberFormat(locale, {
		maximumFractionDigits: Math.abs(amount) < SMALL_AMOUNT ? 1 : 0
	}).format(amount);

// Formats a total with its unit, marking it as a lower bound ("≥") if some
// ingredients' amounts aren't known, or returns "unknown" if there's no total.
export const formatTotal = (total: Total | undefined, unit: string, locale?: string): string => {
	if (total === undefined) {
		return 'unknown';
	}
	const amount = `${formatAmount(total.amount, locale)} ${unit}`;
	return total.missing.length === 0 ? amount : `≥ ${amount}`;
};

// Returns the names of the foods of a recipe's ingredients at the given
// positions, without repeats. foods are the user's foods, by ID.
const foodNames = (recipe: Recipe, positions: number[], foods: Map<string, Food>): string[] => {
	const names = positions.map((position) => {
		const foodId = recipe.ingredients[position]?.foodId ?? '';
		return foods.get(foodId)?.name ?? 'an unknown food';
	});
	return [...new Set(names)];
};

// Returns the lines of a recipe's nutrition for the whole recipe: energy, then
// each of the user's nutrients that isn't archived, in the given order.
// foods must include all of the user's foods, archived or not.
export const nutritionRows = (
	recipe: Recipe,
	nutrients: readonly Nutrient[],
	foods: readonly Food[]
): NutritionRow[] => {
	const foodsById = new Map(foods.map((food) => [food.id, food]));
	const row = (name: string, unit: string, total: Total | undefined): NutritionRow => ({
		name,
		unit,
		total,
		missing: total === undefined ? [] : foodNames(recipe, total.missing, foodsById)
	});
	return [
		row('Energy', 'kcal', recipe.nutrition.kcal),
		...nutrients
			.filter((nutrient) => !nutrient.archived)
			.map((nutrient) => row(nutrient.name, nutrient.unit, recipe.nutrition.nutrients[nutrient.id]))
	];
};
