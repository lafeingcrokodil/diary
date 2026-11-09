import { describe, expect, it } from 'vitest';
import { readNumber, readOptionalNumber } from '#lib/numbers.ts';

const WEIGHT = 70.5;

describe('readNumber', () => {
	it.each([
		{ text: '42', want: 42 },
		{ text: ' 70.5 ', want: 70.5 },
		{ text: '70,5', want: 70.5 },
		{ text: '-1', want: -1 },
		{ text: '0', want: 0 }
	])('reads $text as $want', ({ text, want }) => {
		expect(readNumber(text)).toBe(want);
	});

	it.each(['', ' ', 'four', '1,000.5', 'Infinity'])('rejects %j', (text) => {
		expect(readNumber(text)).toBeUndefined();
	});
});

describe('readOptionalNumber', () => {
	it('reads an empty field as null', () => {
		expect(readOptionalNumber(' ')).toBeNull();
	});

	it('reads other fields like readNumber', () => {
		expect(readOptionalNumber('70,5')).toBe(WEIGHT);
		expect(readOptionalNumber('four')).toBeUndefined();
	});
});
