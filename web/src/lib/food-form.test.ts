import { describe, expect, it } from 'vitest';
import { type FoodForm, emptyForm, parseForm, toForm } from '#lib/food-form.ts';
import type { FoodInput } from '#lib/foods.ts';
import type { Nutrient } from '#lib/nutrients.ts';

// The user's nutrients: protein is active, and fibre is archived.
const nutrients: Nutrient[] = [
	{ id: 'protein', name: 'Protein', unit: 'g', kcalPerGram: 4, archived: false },
	{ id: 'fibre', name: 'Fibre', unit: 'g', kcalPerGram: 2, archived: true }
];

const bread: FoodInput = {
	name: 'Rye bread',
	brand: 'Bakery',
	unit: 'g',
	kcalPer100: 210,
	servings: [{ name: 'slice', amount: 40 }],
	nutrients: { protein: 8.5, fibre: 6 }
};

const breadForm = (): FoodForm => toForm(bread);

describe('parseForm', () => {
	it('reads a food', () => {
		expect(parseForm(breadForm(), nutrients)).toStrictEqual({ ok: true, input: bread });
	});

	it('round-trips a food through toForm', () => {
		const milk: FoodInput = {
			name: 'Milk',
			brand: '',
			unit: 'ml',
			kcalPer100: 64.5,
			servings: [],
			nutrients: {}
		};
		expect(parseForm(toForm(milk), nutrients)).toStrictEqual({ ok: true, input: milk });
	});

	it('trims text and accepts a decimal comma', () => {
		const form = { ...breadForm(), name: '  Rye bread ', brand: ' Bakery ', kcalPer100: ' 210,5 ' };
		expect(parseForm(form, nutrients)).toStrictEqual({
			ok: true,
			input: { ...bread, kcalPer100: 210.5 }
		});
	});

	it('ignores empty serving rows', () => {
		const form = breadForm();
		form.servings.push({ name: ' ', amount: '' });
		expect(parseForm(form, nutrients)).toStrictEqual({ ok: true, input: bread });
	});

	it("keeps amounts of archived nutrients, which the form doesn't show", () => {
		const form = breadForm();
		form.nutrients.protein = '9';
		expect(parseForm(form, nutrients)).toStrictEqual({
			ok: true,
			input: { ...bread, nutrients: { protein: 9, fibre: 6 } }
		});
	});

	it('treats empty nutrient fields as unknown, and accepts a decimal comma', () => {
		const form = breadForm();
		form.nutrients.protein = ' ';
		form.nutrients.fibre = '6,5';
		expect(parseForm(form, nutrients)).toStrictEqual({
			ok: true,
			input: { ...bread, nutrients: { fibre: 6.5 } }
		});
	});

	it('drops amounts of nutrients that no longer exist', () => {
		const form = breadForm();
		form.nutrients.deleted = '1';
		expect(parseForm(form, nutrients)).toStrictEqual({ ok: true, input: bread });
	});

	it.each<[string, (form: FoodForm) => void, string]>([
		[
			'no name',
			(form): void => {
				form.name = ' ';
			},
			'name is required'
		],
		[
			'no energy',
			(form): void => {
				form.kcalPer100 = '';
			},
			'energy'
		],
		[
			"energy that isn't a number",
			(form): void => {
				form.kcalPer100 = 'lots';
			},
			'energy'
		],
		[
			'negative energy',
			(form): void => {
				form.kcalPer100 = '-1';
			},
			'energy'
		],
		[
			'a serving without a name',
			(form): void => {
				form.servings.push({ name: '', amount: '30' });
			},
			'needs a name'
		],
		[
			'a serving without an amount',
			(form): void => {
				form.servings.push({ name: 'loaf', amount: '' });
			},
			'"loaf"'
		],
		[
			'a serving with amount 0',
			(form): void => {
				form.servings.push({ name: 'loaf', amount: '0' });
			},
			'"loaf"'
		],
		[
			"a nutrient amount that isn't a number",
			(form): void => {
				form.nutrients.protein = 'lots';
			},
			'The amount of Protein'
		],
		[
			'a negative nutrient amount',
			(form): void => {
				form.nutrients.protein = '-1';
			},
			'The amount of Protein'
		],
		[
			'two servings with the same name',
			(form): void => {
				form.servings.push({ name: 'slice', amount: '30' });
			},
			'more than one'
		]
	])('rejects %s', (_name, change, error) => {
		const form = breadForm();
		change(form);

		const result = parseForm(form, nutrients);
		expect(result.ok).toBe(false);
		expect(result.ok ? '' : result.error).toContain(error);
	});
});

describe('emptyForm', () => {
	it("isn't valid until the required fields are filled in", () => {
		expect(parseForm(emptyForm(), nutrients).ok).toBe(false);
	});
});
