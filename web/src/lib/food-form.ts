import type { FoodInput, Serving, Unit } from '#lib/foods.ts';
import type { Nutrient } from '#lib/nutrients.ts';

// The values of a food's form fields, as the user typed them.
export interface FoodForm {
	name: string;
	brand: string;
	unit: Unit;
	kcalPer100: string;
	servings: { name: string; amount: string }[];
	// The amounts of nutrients, by nutrient ID. An empty field means that the
	// amount is unknown.
	nutrients: Record<string, string>;
}

// The outcome of reading a form: either a food to save, or what's wrong.
// The fields of a food that its form edits.
export type FoodFields = Omit<FoodInput, 'archived'>;

export type ParsedForm = { ok: true; input: FoodFields } | { ok: false; error: string };

// Returns the form for a new food.
export const emptyForm = (): FoodForm => ({
	name: '',
	brand: '',
	unit: 'g',
	kcalPer100: '',
	servings: [],
	nutrients: {}
});

// Returns a form with a food's current values.
export const toForm = (input: FoodFields): FoodForm => ({
	name: input.name,
	brand: input.brand,
	unit: input.unit,
	kcalPer100: String(input.kcalPer100),
	servings: input.servings.map((serving) => ({
		name: serving.name,
		amount: String(serving.amount)
	})),
	nutrients: Object.fromEntries(
		Object.entries(input.nutrients).map(([id, amount]) => [id, String(amount)])
	)
});

// Reads a number field, or returns undefined if it isn't a finite number.
const readNumber = (text: string): number | undefined => {
	const trimmed = text.trim();
	if (trimmed === '') {
		return undefined;
	}
	const value = Number(trimmed.replace(',', '.'));
	return Number.isFinite(value) ? value : undefined;
};

// Reads one serving row: undefined if it was left empty, or an error message
// if it's invalid.
const readServing = (row: FoodForm['servings'][number]): Serving | string | undefined => {
	const name = row.name.trim();
	if (name === '' && row.amount.trim() === '') {
		return undefined;
	}
	if (name === '') {
		return 'Every serving needs a name.';
	}
	const amount = readNumber(row.amount);
	if (amount === undefined || amount <= 0) {
		return `The serving "${name}" needs an amount greater than 0.`;
	}
	return { name, amount };
};

// Reads the serving rows, ignoring empty ones, or returns an error message.
const readServings = (form: FoodForm): Serving[] | string => {
	const servings: Serving[] = [];
	for (const row of form.servings) {
		const serving = readServing(row);
		if (typeof serving === 'string') {
			return serving;
		}
		if (serving !== undefined) {
			if (servings.some((other) => other.name === serving.name)) {
				return `There's more than one serving called "${serving.name}".`;
			}
			servings.push(serving);
		}
	}
	return servings;
};

// Reads the nutrient amounts, ignoring empty fields, or returns an error
// message. Amounts of nutrients that aren't in nutrients (e.g. because they
// were deleted) are dropped. Archived nutrients' amounts are kept, although
// the form doesn't show them, so that saving the food doesn't lose them.
const readNutrients = (
	form: FoodForm,
	nutrients: readonly Nutrient[]
): Record<string, number> | string => {
	const amounts: Record<string, number> = {};
	for (const nutrient of nutrients) {
		const text = form.nutrients[nutrient.id] ?? '';
		if (text.trim() !== '') {
			const amount = readNumber(text);
			if (amount === undefined || amount < 0) {
				return `The amount of ${nutrient.name} must be a number, 0 or more, or left empty.`;
			}
			amounts[nutrient.id] = amount;
		}
	}
	return amounts;
};

// Reads the servings and nutrient amounts, or returns an error message.
const readAmounts = (
	form: FoodForm,
	nutrients: readonly Nutrient[]
): Pick<FoodInput, 'servings' | 'nutrients'> | string => {
	const servings = readServings(form);
	if (typeof servings === 'string') {
		return servings;
	}
	const amounts = readNutrients(form, nutrients);
	if (typeof amounts === 'string') {
		return amounts;
	}
	return { servings, nutrients: amounts };
};

// Checks a form, with the same rules as the API, and returns the food it
// describes. nutrients are all of the user's nutrients, archived or not.
// Numbers may use a decimal comma.
export const parseForm = (form: FoodForm, nutrients: readonly Nutrient[]): ParsedForm => {
	const name = form.name.trim();
	if (name === '') {
		return { ok: false, error: 'The name is required.' };
	}
	const kcalPer100 = readNumber(form.kcalPer100);
	if (kcalPer100 === undefined || kcalPer100 < 0) {
		return { ok: false, error: 'The energy must be a number, 0 or more.' };
	}
	const amounts = readAmounts(form, nutrients);
	if (typeof amounts === 'string') {
		return { ok: false, error: amounts };
	}
	return {
		ok: true,
		input: { name, brand: form.brand.trim(), unit: form.unit, kcalPer100, ...amounts }
	};
};
