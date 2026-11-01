import type { NutrientInput, Unit } from '#lib/nutrients.ts';

// The values of a nutrient's form fields, as the user typed them. Whether the
// nutrient is archived isn't part of the form.
export interface NutrientForm {
	name: string;
	unit: Unit;
	kcalPerGram: string;
}

// The fields of a nutrient that its form edits.
export type NutrientFields = Omit<NutrientInput, 'archived'>;

// The outcome of reading a form: either the nutrient's fields, or what's wrong.
export type ParsedForm = { ok: true; input: NutrientFields } | { ok: false; error: string };

// Returns the form for a new nutrient.
export const emptyForm = (): NutrientForm => ({ name: '', unit: 'g', kcalPerGram: '' });

// Returns a form with a nutrient's current values. An energy of 0 is shown as
// an empty field, as it's what an empty field means.
export const toForm = (input: NutrientInput): NutrientForm => ({
	name: input.name,
	unit: input.unit,
	kcalPerGram: input.kcalPerGram === 0 ? '' : String(input.kcalPerGram)
});

// Checks a form, with the same rules as the API, and returns the nutrient's
// fields. The energy is optional, and may use a decimal comma.
export const parseForm = (form: NutrientForm): ParsedForm => {
	const name = form.name.trim();
	if (name === '') {
		return { ok: false, error: 'The name is required.' };
	}
	const energy = form.kcalPerGram.trim();
	const kcalPerGram = energy === '' ? 0 : Number(energy.replace(',', '.'));
	if (!Number.isFinite(kcalPerGram) || kcalPerGram < 0) {
		return { ok: false, error: 'The energy must be a number, 0 or more, or left empty.' };
	}
	return { ok: true, input: { name, unit: form.unit, kcalPerGram } };
};
