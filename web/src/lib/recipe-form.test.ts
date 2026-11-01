import { describe, expect, it } from 'vitest';
import type { Food } from '#lib/foods.ts';
import {
	FoodChoices,
	type RecipeFields,
	type RecipeForm,
	emptyForm,
	parseForm,
	swap,
	toForm
} from '#lib/recipe-form.ts';

const food = (id: string, name: string, changes: Partial<Food> = {}): Food => ({
	id,
	name,
	brand: '',
	unit: 'g',
	kcalPer100: 100,
	servings: [],
	nutrients: {},
	archived: false,
	...changes
});

// The user's foods, sorted by name as the API lists them: two lentils with the
// same name and brand, and an archived food.
const foods: Food[] = [
	food('lentils-1', 'Lentils'),
	food('lentils-2', 'Lentils'),
	food('onion', 'Onion', { brand: 'Farm', servings: [{ name: 'medium onion', amount: 150 }] }),
	food('stock', 'Stock', { archived: true })
];

const choices = new FoodChoices(foods);

const soup: RecipeFields = {
	name: 'Lentil soup',
	sourceUrl: 'https://example.com/lentil-soup',
	ingredients: [
		{ foodId: 'lentils-2', serving: '', amount: 200, note: 'rinsed' },
		{ foodId: 'onion', serving: 'medium onion', amount: 1.5, note: '' },
		{ foodId: 'stock', serving: '', amount: 1000, note: '' }
	],
	steps: ['Fry the onion.', 'Add the rest and simmer.'],
	notes: 'Freezes well.'
};

const soupForm = (): RecipeForm => toForm(soup, choices);

describe('FoodChoices', () => {
	it('labels foods by name and brand, uniquely', () => {
		expect(choices.labelOf('lentils-1')).toBe('Lentils');
		expect(choices.labelOf('lentils-2')).toBe('Lentils [2]');
		expect(choices.labelOf('onion')).toBe('Onion (Farm)');
		expect(choices.labelOf('missing')).toBe('');
	});

	it('finds foods by label', () => {
		expect(choices.find(' Lentils [2] ')?.id).toBe('lentils-2');
		expect(choices.find('Onion')).toBeUndefined();
	});

	it("only offers foods that aren't archived", () => {
		expect(choices.activeLabels()).toStrictEqual(['Lentils', 'Lentils [2]', 'Onion (Farm)']);
	});
});

describe('parseForm', () => {
	it('round-trips a recipe through toForm', () => {
		expect(parseForm(soupForm(), choices)).toStrictEqual({ ok: true, input: soup });
	});

	it('reads an empty form as a recipe without ingredients or steps', () => {
		const form = { ...emptyForm(), name: 'Idea' };
		expect(parseForm(form, choices)).toStrictEqual({
			ok: true,
			input: { name: 'Idea', sourceUrl: '', ingredients: [], steps: [], notes: '' }
		});
	});

	it('trims text, accepts a decimal comma, and leaves out empty rows and steps', () => {
		const form = soupForm();
		form.name = ' Lentil soup ';
		form.sourceUrl = ' https://example.com/lentil-soup ';
		form.notes = ' Freezes well. ';
		form.ingredients[0] = { ...form.ingredients[0], food: ' Lentils [2] ', note: ' rinsed ' };
		form.ingredients[1] = { ...form.ingredients[1], amount: '1,5' };
		form.ingredients.splice(1, 0, { food: ' ', amount: '', serving: '', note: '' });
		form.steps = [' ', ' Fry the onion. ', 'Add the rest and simmer.', ''];
		expect(parseForm(form, choices)).toStrictEqual({ ok: true, input: soup });
	});

	it.each<{ name: string; change: (form: RecipeForm) => void; error: string }>([
		{
			name: 'no name',
			change: (form): void => {
				form.name = ' ';
			},
			error: 'The name is required.'
		},
		{
			name: "a source that isn't a URL",
			change: (form): void => {
				form.sourceUrl = 'example.com';
			},
			error: 'The source must be a web address'
		},
		{
			name: 'a source with another scheme',
			change: (form): void => {
				form.sourceUrl = 'ftp://example.com/soup';
			},
			error: 'The source must be a web address'
		},
		{
			name: 'an ingredient without a food',
			change: (form): void => {
				form.ingredients[1].food = '';
			},
			error: 'Ingredient 2 needs a food.'
		},
		{
			name: 'an unknown food',
			change: (form): void => {
				form.ingredients[1].food = 'Onion';
			},
			error: `Ingredient 2: there's no food called "Onion".`
		},
		{
			name: 'an unknown serving',
			change: (form): void => {
				form.ingredients[0].serving = 'cup';
			},
			error: 'Ingredient 1: Lentils has no serving called "cup".'
		},
		{
			name: 'an ingredient without an amount',
			change: (form): void => {
				form.ingredients[1].amount = '';
			},
			error: 'Ingredient 2 needs an amount greater than 0.'
		},
		{
			name: 'an ingredient with an amount of 0',
			change: (form): void => {
				form.ingredients[1].amount = '0';
			},
			error: 'Ingredient 2 needs an amount greater than 0.'
		},
		{
			name: "an amount that isn't a number",
			change: (form): void => {
				form.ingredients[1].amount = 'lots';
			},
			error: 'Ingredient 2 needs an amount greater than 0.'
		}
	])('rejects $name', ({ change, error }) => {
		const form = soupForm();
		change(form);
		const parsed = parseForm(form, choices);
		expect(parsed.ok).toBe(false);
		expect(parsed.ok ? '' : parsed.error).toContain(error);
	});
});

describe('swap', () => {
	it('swaps two items', () => {
		const items = ['a', 'b', 'c'];
		swap(items, 1, 0);
		expect(items).toStrictEqual(['b', 'a', 'c']);
	});

	it('leaves the items alone if an index is out of range', () => {
		const items = ['a', 'b', 'c'];
		swap(items, 0, items.length);
		swap(items, 0, Number.NaN);
		swap(items, -items.length, 0);
		expect(items).toStrictEqual(['a', 'b', 'c']);
	});
});
