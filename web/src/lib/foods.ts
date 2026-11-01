import { apiFetch } from '#lib/api.ts';

// What a food's amounts are measured in.
export type Unit = 'g' | 'ml';

// A named amount of a food, e.g. "slice", in the food's unit.
export interface Serving {
	name: string;
	amount: number;
}

// A food as the user enters it, with its energy per 100 of its unit.
export interface FoodInput {
	name: string;
	brand: string;
	unit: Unit;
	kcalPer100: number;
	servings: Serving[];
	// The amounts of the user's nutrients per 100 of the food's unit, by
	// nutrient ID. A nutrient without an amount has an unknown amount.
	nutrients: Record<string, number>;
	// Archived foods are kept, but are hidden until they're unarchived.
	archived: boolean;
}

// A stored food.
export interface Food extends FoodInput {
	id: string;
}

const isServing = (value: unknown): value is Serving =>
	typeof value === 'object' &&
	value !== null &&
	'name' in value &&
	typeof value.name === 'string' &&
	'amount' in value &&
	typeof value.amount === 'number';

const isAmounts = (value: unknown): value is Record<string, number> =>
	typeof value === 'object' &&
	value !== null &&
	!Array.isArray(value) &&
	Object.values(value).every((amount) => typeof amount === 'number');

const isFood = (value: unknown): value is Food =>
	typeof value === 'object' &&
	value !== null &&
	'id' in value &&
	typeof value.id === 'string' &&
	'name' in value &&
	typeof value.name === 'string' &&
	'brand' in value &&
	typeof value.brand === 'string' &&
	'unit' in value &&
	(value.unit === 'g' || value.unit === 'ml') &&
	'kcalPer100' in value &&
	typeof value.kcalPer100 === 'number' &&
	'servings' in value &&
	Array.isArray(value.servings) &&
	value.servings.every(isServing) &&
	'nutrients' in value &&
	isAmounts(value.nutrients) &&
	'archived' in value &&
	typeof value.archived === 'boolean';

const readFood = async (response: Response): Promise<Food> => {
	const value: unknown = await response.json();
	if (!isFood(value)) {
		throw new Error('The API responded with an invalid food.');
	}
	return value;
};

const foodPath = (id: string): string => `/api/foods/${encodeURIComponent(id)}`;

const jsonRequest = (method: string, body: FoodInput): RequestInit => ({
	method,
	headers: { 'Content-Type': 'application/json' },
	body: JSON.stringify(body)
});

// Returns all of the signed-in user's foods, including archived ones, sorted
// by name.
export const listFoods = async (): Promise<Food[]> => {
	const response = await apiFetch('/api/foods');
	const value: unknown = await response.json();
	if (!Array.isArray(value) || !value.every(isFood)) {
		throw new Error('The API responded with an invalid list of foods.');
	}
	return value;
};

// Returns one of the signed-in user's foods.
export const getFood = async (id: string): Promise<Food> => readFood(await apiFetch(foodPath(id)));

// Stores a new food, and returns it with its ID.
export const createFood = async (input: FoodInput): Promise<Food> =>
	readFood(await apiFetch('/api/foods', jsonRequest('POST', input)));

// Replaces one of the signed-in user's foods, and returns it as stored.
// Archiving and unarchiving a food are updates of its archived field.
export const updateFood = async (id: string, input: FoodInput): Promise<Food> =>
	readFood(await apiFetch(foodPath(id), jsonRequest('PUT', input)));

// Permanently deletes one of the signed-in user's foods. Only archived foods
// can be deleted.
export const deleteFood = async (id: string): Promise<void> => {
	await apiFetch(foodPath(id), { method: 'DELETE' });
};
