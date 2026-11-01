import type { FoodInput, Serving, Unit } from '#lib/foods.ts';

// The values of a food's form fields, as the user typed them.
export interface FoodForm {
	name: string;
	brand: string;
	unit: Unit;
	kcalPer100: string;
	servings: { name: string; amount: string }[];
}

// The outcome of reading a form: either a food to save, or what's wrong.
export type ParsedForm = { ok: true; input: FoodInput } | { ok: false; error: string };

// Returns the form for a new food.
export const emptyForm = (): FoodForm => ({
	name: '',
	brand: '',
	unit: 'g',
	kcalPer100: '',
	servings: []
});

// Returns a form with a food's current values.
export const toForm = (input: FoodInput): FoodForm => ({
	name: input.name,
	brand: input.brand,
	unit: input.unit,
	kcalPer100: String(input.kcalPer100),
	servings: input.servings.map((serving) => ({
		name: serving.name,
		amount: String(serving.amount)
	}))
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

// Checks a form, with the same rules as the API, and returns the food it
// describes. Numbers may use a decimal comma.
export const parseForm = (form: FoodForm): ParsedForm => {
	const name = form.name.trim();
	if (name === '') {
		return { ok: false, error: 'The name is required.' };
	}
	const kcalPer100 = readNumber(form.kcalPer100);
	if (kcalPer100 === undefined || kcalPer100 < 0) {
		return { ok: false, error: 'The energy must be a number, 0 or more.' };
	}
	const servings = readServings(form);
	if (typeof servings === 'string') {
		return { ok: false, error: servings };
	}
	return {
		ok: true,
		input: { name, brand: form.brand.trim(), unit: form.unit, kcalPer100, servings }
	};
};
