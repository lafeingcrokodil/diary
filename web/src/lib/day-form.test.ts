import { describe, expect, it } from 'vitest';
import { type DayForm, parseForm, toForm } from '#lib/day-form.ts';

describe('toForm', () => {
	it('shows values, and leaves values that are not logged empty', () => {
		expect(toForm({ weightKg: 70.5, kcalExpended: null })).toStrictEqual({
			weightKg: '70.5',
			kcalExpended: ''
		});
		expect(toForm({ weightKg: null, kcalExpended: 0 })).toStrictEqual({
			weightKg: '',
			kcalExpended: '0'
		});
	});
});

describe('parseForm', () => {
	it('reads both values', () => {
		expect(parseForm({ weightKg: '70.5', kcalExpended: '2300' })).toStrictEqual({
			ok: true,
			input: { weightKg: 70.5, kcalExpended: 2300 }
		});
	});

	it('treats empty fields as not logged', () => {
		expect(parseForm({ weightKg: ' ', kcalExpended: '' })).toStrictEqual({
			ok: true,
			input: { weightKg: null, kcalExpended: null }
		});
	});

	it('accepts a decimal comma, and no energy expended', () => {
		expect(parseForm({ weightKg: ' 70,5 ', kcalExpended: '0' })).toStrictEqual({
			ok: true,
			input: { weightKg: 70.5, kcalExpended: 0 }
		});
	});

	it.each<[string, DayForm, string]>([
		['a weight that is not a number', { weightKg: 'heavy', kcalExpended: '' }, 'The weight'],
		['a weight of 0', { weightKg: '0', kcalExpended: '' }, 'The weight'],
		['a negative weight', { weightKg: '-70', kcalExpended: '' }, 'The weight'],
		['an energy that is not a number', { weightKg: '', kcalExpended: 'lots' }, 'The energy'],
		['a negative energy', { weightKg: '', kcalExpended: '-1' }, 'The energy']
	])('rejects %s', (_name, form, error) => {
		const result = parseForm(form);
		expect(result.ok).toBe(false);
		expect(result.ok ? '' : result.error).toContain(error);
	});
});
