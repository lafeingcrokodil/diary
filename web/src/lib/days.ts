import { apiFetch } from '#lib/api.ts';

// What the user logs for a day. A value that isn't logged is null.
export interface DayInput {
	weightKg: number | null;
	kcalExpended: number | null;
}

// A day as the API returns it, with its date (e.g. 2026-11-05).
export interface Day extends DayInput {
	date: string;
}

const isNumberOrNull = (value: unknown): value is number | null =>
	value === null || typeof value === 'number';

const isDay = (value: unknown): value is Day =>
	typeof value === 'object' &&
	value !== null &&
	'date' in value &&
	typeof value.date === 'string' &&
	'weightKg' in value &&
	isNumberOrNull(value.weightKg) &&
	'kcalExpended' in value &&
	isNumberOrNull(value.kcalExpended);

const readDay = async (response: Response): Promise<Day> => {
	const value: unknown = await response.json();
	if (!isDay(value)) {
		throw new Error('The API responded with an invalid day.');
	}
	return value;
};

const dayPath = (date: string): string => `/api/days/${encodeURIComponent(date)}`;

// Returns the signed-in user's day with the given date. A day that nothing is
// logged for has null values.
export const getDay = async (date: string): Promise<Day> => readDay(await apiFetch(dayPath(date)));

// Replaces what the signed-in user logged for the day with the given date, and
// returns the day as stored.
export const putDay = async (date: string, input: DayInput): Promise<Day> =>
	readDay(
		await apiFetch(dayPath(date), {
			method: 'PUT',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(input)
		})
	);
