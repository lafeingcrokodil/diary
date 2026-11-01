import { describe, expect, it } from 'vitest';
import type { Food } from '#lib/foods.ts';
import type { Nutrient } from '#lib/nutrients.ts';
import { formatAmount, formatTotal, nutritionRows } from '#lib/recipe-nutrition.ts';
import type { Recipe } from '#lib/recipes.ts';

const food = (id: string, name: string): Food => ({
	id,
	name,
	brand: '',
	unit: 'g',
	kcalPer100: 100,
	servings: [],
	nutrients: {},
	archived: false
});

const nutrients: Nutrient[] = [
	{ id: 'protein', name: 'Protein', unit: 'g', kcalPerGram: 4, archived: false },
	{ id: 'iron', name: 'Iron', unit: 'mg', kcalPerGram: 0, archived: false },
	{ id: 'fibre', name: 'Fibre', unit: 'g', kcalPerGram: 2, archived: true },
	{ id: 'salt', name: 'Salt', unit: 'g', kcalPerGram: 0, archived: false }
];

const ingredient = (foodId: string): Recipe['ingredients'][number] => ({
	foodId,
	serving: '',
	amount: 100,
	note: ''
});

// Soup's iron is only known for its lentils, and not for its onions (twice) or
// for a food that's gone.
const ingredients = [
	ingredient('onion'),
	ingredient('lentils'),
	ingredient('onion'),
	ingredient('gone')
];
const withoutIron = [...ingredients.keys()].filter(
	(position) => ingredients[position]?.foodId !== 'lentils'
);

const soup: Recipe = {
	id: 'soup',
	name: 'Soup',
	sourceUrl: '',
	ingredients,
	steps: [],
	notes: '',
	archived: false,
	nutrition: {
		kcal: { amount: 1057.5, missing: [] },
		nutrients: {
			protein: { amount: 31.35, missing: [] },
			iron: { amount: 6.64, missing: withoutIron },
			fibre: { amount: 22, missing: [] }
		}
	}
};

const foods = [food('onion', 'Onion'), food('lentils', 'Lentils')];

describe('formatAmount', () => {
	it.each([
		{ amount: 0, want: '0' },
		{ amount: 6.64, want: '6.6' },
		{ amount: 9.96, want: '10' },
		{ amount: 31.35, want: '31' },
		{ amount: 1057.5, want: '1,058' }
	])('formats $amount as $want', ({ amount, want }) => {
		expect(formatAmount(amount, 'en')).toBe(want);
	});

	it('uses the locale', () => {
		expect(formatAmount(soup.nutrition.nutrients.iron?.amount ?? 0, 'de')).toBe('6,6');
	});
});

describe('formatTotal', () => {
	const { protein, iron } = soup.nutrition.nutrients;

	it('formats a complete total', () => {
		expect(formatTotal(protein, 'g', 'en')).toBe('31 g');
	});

	it('marks an incomplete total as a lower bound', () => {
		expect(formatTotal(iron, 'mg', 'en')).toBe('≥ 6.6 mg');
	});

	it('reports an unknown total', () => {
		expect(formatTotal(undefined, 'g', 'en')).toBe('unknown');
	});
});

describe('nutritionRows', () => {
	it('lists energy and the active nutrients, naming foods whose amounts are missing', () => {
		expect(nutritionRows(soup, nutrients, foods)).toStrictEqual([
			{ name: 'Energy', unit: 'kcal', total: soup.nutrition.kcal, missing: [] },
			{ name: 'Protein', unit: 'g', total: soup.nutrition.nutrients.protein, missing: [] },
			{
				name: 'Iron',
				unit: 'mg',
				total: soup.nutrition.nutrients.iron,
				missing: ['Onion', 'an unknown food']
			},
			{ name: 'Salt', unit: 'g', total: undefined, missing: [] }
		]);
	});
});
