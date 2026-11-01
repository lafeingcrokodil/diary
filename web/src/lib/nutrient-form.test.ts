import { describe, expect, it } from 'vitest';
import { type NutrientForm, emptyForm, parseForm, toForm } from '#lib/nutrient-form.ts';
import type { NutrientInput } from '#lib/nutrients.ts';

const protein: NutrientInput = { name: 'Protein', unit: 'g', kcalPerGram: 4, archived: false };

describe('parseForm', () => {
	it('reads a nutrient', () => {
		expect(parseForm(toForm(protein))).toStrictEqual({
			ok: true,
			input: { name: 'Protein', unit: 'g', kcalPerGram: 4 }
		});
	});

	it('treats an empty energy as 0, and shows 0 as empty', () => {
		const salt: NutrientInput = { name: 'Salt', unit: 'mg', kcalPerGram: 0, archived: true };
		expect(toForm(salt).kcalPerGram).toBe('');
		expect(parseForm(toForm(salt))).toStrictEqual({
			ok: true,
			input: { name: 'Salt', unit: 'mg', kcalPerGram: 0 }
		});
	});

	it('trims text and accepts a decimal comma', () => {
		const form: NutrientForm = { name: ' Alcohol  ', unit: 'g', kcalPerGram: ' 6,9 ' };
		expect(parseForm(form)).toStrictEqual({
			ok: true,
			input: { name: 'Alcohol', unit: 'g', kcalPerGram: 6.9 }
		});
	});

	it.each<[string, Partial<NutrientForm>, string]>([
		['no name', { name: ' ' }, 'The name is required.'],
		["an energy that isn't a number", { kcalPerGram: 'four' }, 'The energy must be'],
		['a negative energy', { kcalPerGram: '-1' }, 'The energy must be']
	])('rejects %s', (_name, change, error) => {
		const result = parseForm({ ...toForm(protein), ...change });
		expect(result.ok).toBe(false);
		expect(result.ok ? '' : result.error).toContain(error);
	});
});

describe('emptyForm', () => {
	it('needs a name', () => {
		expect(parseForm(emptyForm())).toStrictEqual({ ok: false, error: 'The name is required.' });
	});
});
