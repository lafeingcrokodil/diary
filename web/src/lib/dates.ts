// Calendar dates, written like 2026-11-05. They have no time zone: a date
// means that day wherever the user is.

const DATE_PATTERN = /^\d{4}-\d{2}-\d{2}$/v;

const pad = (value: number, length: number): string => String(value).padStart(length, '0');

const YEAR_DIGITS = 4;
const MONTH_DAY_DIGITS = 2;

// Formats the calendar date of a Date's UTC fields.
const fromUTC = (date: Date): string =>
	[
		pad(date.getUTCFullYear(), YEAR_DIGITS),
		pad(date.getUTCMonth() + 1, MONTH_DAY_DIGITS),
		pad(date.getUTCDate(), MONTH_DAY_DIGITS)
	].join('-');

// Returns the date as midnight UTC, or undefined if it isn't a real date.
// Working in UTC means that adding days never skips or repeats a day because
// of daylight saving time.
const toUTC = (date: string): Date | undefined => {
	if (!DATE_PATTERN.test(date)) {
		return undefined;
	}
	const parsed = new Date(`${date}T00:00:00Z`);
	// Date accepts some days that don't exist (e.g. 2026-02-30) by rolling
	// over, so the date must also be what formatting it gives back.
	return !Number.isNaN(parsed.getTime()) && fromUTC(parsed) === date ? parsed : undefined;
};

// Whether text is a real calendar date, written like 2026-11-05.
export const isDate = (text: string): boolean => toUTC(text) !== undefined;

// Returns the user's local date at the given time.
export const localDate = (now: Date): string =>
	[
		pad(now.getFullYear(), YEAR_DIGITS),
		pad(now.getMonth() + 1, MONTH_DAY_DIGITS),
		pad(now.getDate(), MONTH_DAY_DIGITS)
	].join('-');

// Returns the date a number of days after date (before it, if days is
// negative). date must be a real calendar date.
export const addDays = (date: string, days: number): string => {
	const parsed = toUTC(date);
	if (parsed === undefined) {
		throw new RangeError(`${date} isn't a date.`);
	}
	parsed.setUTCDate(parsed.getUTCDate() + days);
	return fromUTC(parsed);
};

// Formats a date for people to read, e.g. "Thursday, 5 November 2026" in
// British English. locales defaults to the browser's languages.
export const formatDate = (date: string, locales?: Intl.LocalesArgument): string => {
	const parsed = toUTC(date);
	if (parsed === undefined) {
		throw new RangeError(`${date} isn't a date.`);
	}
	return parsed.toLocaleDateString(locales, {
		weekday: 'long',
		day: 'numeric',
		month: 'long',
		year: 'numeric',
		timeZone: 'UTC'
	});
};
