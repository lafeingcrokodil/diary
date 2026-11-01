import { apiFetch } from '#lib/api.ts';

// What a nutrient's amounts are measured in.
export type Unit = 'g' | 'mg' | 'µg';

export const units: readonly Unit[] = ['g', 'mg', 'µg'];

// A nutrient as the user enters it. kcalPerGram is 0 for nutrients that
// provide no energy, or whose energy isn't known.
export interface NutrientInput {
	name: string;
	unit: Unit;
	kcalPerGram: number;
	archived: boolean;
}

// A stored nutrient.
export interface Nutrient extends NutrientInput {
	id: string;
}

const isUnit = (value: unknown): value is Unit => units.some((unit) => unit === value);

const isNutrient = (value: unknown): value is Nutrient =>
	typeof value === 'object' &&
	value !== null &&
	'id' in value &&
	typeof value.id === 'string' &&
	'name' in value &&
	typeof value.name === 'string' &&
	'unit' in value &&
	isUnit(value.unit) &&
	'kcalPerGram' in value &&
	typeof value.kcalPerGram === 'number' &&
	'archived' in value &&
	typeof value.archived === 'boolean';

const readNutrient = async (response: Response): Promise<Nutrient> => {
	const value: unknown = await response.json();
	if (!isNutrient(value)) {
		throw new Error('The API responded with an invalid nutrient.');
	}
	return value;
};

const nutrientPath = (id: string): string => `/api/nutrients/${encodeURIComponent(id)}`;

const jsonRequest = (method: string, body: NutrientInput): RequestInit => ({
	method,
	headers: { 'Content-Type': 'application/json' },
	body: JSON.stringify(body)
});

// Returns all of the signed-in user's nutrients, including archived ones,
// sorted by name.
export const listNutrients = async (): Promise<Nutrient[]> => {
	const response = await apiFetch('/api/nutrients');
	const value: unknown = await response.json();
	if (!Array.isArray(value) || !value.every(isNutrient)) {
		throw new Error('The API responded with an invalid list of nutrients.');
	}
	return value;
};

// Returns one of the signed-in user's nutrients.
export const getNutrient = async (id: string): Promise<Nutrient> =>
	readNutrient(await apiFetch(nutrientPath(id)));

// Stores a new nutrient, and returns it with its ID.
export const createNutrient = async (input: NutrientInput): Promise<Nutrient> =>
	readNutrient(await apiFetch('/api/nutrients', jsonRequest('POST', input)));

// Replaces one of the signed-in user's nutrients, and returns it as stored.
// Archiving and unarchiving a nutrient are updates of its archived field.
export const updateNutrient = async (id: string, input: NutrientInput): Promise<Nutrient> =>
	readNutrient(await apiFetch(nutrientPath(id), jsonRequest('PUT', input)));

// Permanently deletes one of the signed-in user's nutrients. Only archived
// nutrients can be deleted.
export const deleteNutrient = async (id: string): Promise<void> => {
	await apiFetch(nutrientPath(id), { method: 'DELETE' });
};
