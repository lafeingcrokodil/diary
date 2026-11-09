import type { DayInput } from '#lib/days.ts';
import { readOptionalNumber } from '#lib/numbers.ts';

// The values of a day's form fields, as the user typed them.
export interface DayForm {
	weightKg: string;
	kcalExpended: string;
}

// The outcome of reading a form: either the day's values, or what's wrong.
export type ParsedForm = { ok: true; input: DayInput } | { ok: false; error: string };

// Returns a form with a day's current values. A value that isn't logged is an
// empty field.
export const toForm = (input: DayInput): DayForm => ({
	weightKg: input.weightKg === null ? '' : String(input.weightKg),
	kcalExpended: input.kcalExpended === null ? '' : String(input.kcalExpended)
});

// Checks a form, with the same rules as the API, and returns the day's values.
// Both are optional: an empty field isn't logged. They may use a decimal comma.
export const parseForm = (form: DayForm): ParsedForm => {
	const weightKg = readOptionalNumber(form.weightKg);
	if (weightKg === undefined || (weightKg !== null && weightKg <= 0)) {
		return { ok: false, error: 'The weight must be a number above 0, or left empty.' };
	}
	const kcalExpended = readOptionalNumber(form.kcalExpended);
	if (kcalExpended === undefined || (kcalExpended !== null && kcalExpended < 0)) {
		return { ok: false, error: 'The energy expended must be a number, 0 or more, or left empty.' };
	}
	return { ok: true, input: { weightKg, kcalExpended } };
};
