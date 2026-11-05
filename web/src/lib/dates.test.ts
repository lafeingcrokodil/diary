import { describe, expect, it } from 'vitest';
import { addDays, formatDate, isDate, localDate } from '#lib/dates.ts';

describe('isDate', () => {
	it.each(['2026-11-05', '2028-02-29', '0001-01-01'])('accepts %s', (text) => {
		expect(isDate(text)).toBe(true);
	});

	it.each(['', 'today', '2026-02-29', '2026-13-01', '2026-11-31', '2026-1-5', '2026-11-05T00:00'])(
		'rejects %j',
		(text) => {
			expect(isDate(text)).toBe(false);
		}
	);
});

describe('localDate', () => {
	it('uses the local date, not the UTC one', () => {
		// Without a time zone, these are local times, so the test passes in any
		// time zone.
		const lateEvening = new Date('2026-11-05T23:59');
		const earlyMorning = new Date('2026-11-06T00:01');
		expect(localDate(lateEvening)).toBe('2026-11-05');
		expect(localDate(earlyMorning)).toBe('2026-11-06');
	});
});

const BACK = -1;

describe('addDays', () => {
	it.each<[string, number, string]>([
		['2026-11-05', 1, '2026-11-06'],
		['2026-11-05', BACK, '2026-11-04'],
		['2026-11-05', 0, '2026-11-05'],
		['2026-12-31', 1, '2027-01-01'],
		['2028-03-01', BACK, '2028-02-29'],
		// The day that clocks go back in much of Europe.
		['2026-10-25', 1, '2026-10-26']
	])('%s plus %i days is %s', (date, days, want) => {
		expect(addDays(date, days)).toBe(want);
	});

	it('rejects dates that are not real', () => {
		expect(() => addDays('2026-02-30', 1)).toThrow(RangeError);
	});
});

describe('formatDate', () => {
	it('formats the date with its weekday, whatever the time zone', () => {
		expect(formatDate('2026-11-05', 'en-GB')).toBe('Thursday, 5 November 2026');
		expect(formatDate('2026-11-05', 'en-US')).toBe('Thursday, November 5, 2026');
	});

	it('rejects dates that are not real', () => {
		expect(() => formatDate('nope')).toThrow(RangeError);
	});
});
