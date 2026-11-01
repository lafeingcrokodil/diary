import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { ApiError } from '#lib/api.ts';
import {
	type Food,
	type FoodInput,
	createFood,
	deleteFood,
	getFood,
	listFoods,
	updateFood
} from '#lib/foods.ts';

// The tests don't sign in: every request gets this token.
vi.mock('#lib/auth.svelte.ts', (): { idToken: () => Promise<string> } => ({
	idToken: async (): Promise<string> => Promise.resolve('test-token')
}));

const bread: FoodInput = {
	name: 'Rye bread',
	brand: 'Bakery',
	unit: 'g',
	kcalPer100: 210,
	servings: [{ name: 'slice', amount: 40 }],
	nutrients: { protein: 8.5 },
	archived: false
};

const storedBread: Food = { id: 'food-1', ...bread };

const STATUS_OK = 200;
const STATUS_CREATED = 201;
const STATUS_NO_CONTENT = 204;
const STATUS_NOT_FOUND = 404;

const fetchMock = vi.fn<typeof fetch>();

// Makes the next request get a response with the given body and status.
const respond = (body: unknown, status = STATUS_OK): void => {
	fetchMock.mockResolvedValueOnce(
		new Response(body === undefined ? null : JSON.stringify(body), { status })
	);
};

// Returns the path, options and decoded JSON body (if any) of the only request
// that was sent.
const sentRequest = (): { path: string; init: RequestInit; body: unknown } => {
	expect(fetchMock).toHaveBeenCalledOnce();
	const [path, init = {}] = fetchMock.mock.lastCall ?? [];
	if (typeof path !== 'string') {
		throw new TypeError('The request was sent without a string path.');
	}
	const body: unknown = typeof init.body === 'string' ? JSON.parse(init.body) : undefined;
	return { path, init, body };
};

beforeEach(() => {
	vi.stubGlobal('fetch', fetchMock);
});

afterEach(() => {
	fetchMock.mockReset();
	vi.unstubAllGlobals();
});

describe('listFoods', () => {
	it('returns the foods', async () => {
		respond([storedBread]);

		await expect(listFoods()).resolves.toStrictEqual([storedBread]);

		const { path, init } = sentRequest();
		expect(path).toBe('/api/foods');
		expect(new Headers(init.headers).get('Authorization')).toBe('Bearer test-token');
	});

	it.each([
		['an unknown unit', { unit: 'oz' }],
		['no nutrients', { nutrients: undefined }],
		["a nutrient amount that isn't a number", { nutrients: { protein: 'lots' } }],
		['a list of nutrients', { nutrients: [bread.nutrients] }],
		['no archived flag', { archived: undefined }]
	])('rejects a food with %s', async (_name, change) => {
		respond([{ ...storedBread, ...change }]);

		await expect(listFoods()).rejects.toThrow('invalid list of foods');
	});
});

describe('getFood', () => {
	it('returns the food', async () => {
		respond(storedBread);

		await expect(getFood('food-1')).resolves.toStrictEqual(storedBread);
		expect(sentRequest().path).toBe('/api/foods/food-1');
	});

	it('escapes the ID', async () => {
		respond(storedBread);

		await getFood('a/b');
		expect(sentRequest().path).toBe('/api/foods/a%2Fb');
	});

	it('reports a missing food, with the reason the API gave', async () => {
		fetchMock.mockResolvedValueOnce(
			new Response('404 page not found\n', { status: STATUS_NOT_FOUND })
		);

		const error: unknown = await getFood('missing').catch((error_: unknown) => error_);
		expect(error).toBeInstanceOf(ApiError);
		expect(String(error)).toContain('404 page not found');
	});

	it('rejects an invalid response', async () => {
		respond({ ...storedBread, servings: [{ name: 'slice' }] });

		await expect(getFood('food-1')).rejects.toThrow('invalid food');
	});
});

describe('createFood', () => {
	it('sends the food and returns it as stored', async () => {
		respond(storedBread, STATUS_CREATED);

		await expect(createFood(bread)).resolves.toStrictEqual(storedBread);

		const { path, init, body } = sentRequest();
		expect(path).toBe('/api/foods');
		expect(init.method).toBe('POST');
		expect(new Headers(init.headers).get('Content-Type')).toBe('application/json');
		expect(body).toStrictEqual(bread);
	});
});

describe('updateFood', () => {
	it('sends the food and returns it as stored', async () => {
		const changed = { ...bread, name: 'Wholemeal rye bread' };
		respond({ id: 'food-1', ...changed });

		await expect(updateFood('food-1', changed)).resolves.toStrictEqual({
			id: 'food-1',
			...changed
		});

		const { path, init, body } = sentRequest();
		expect(path).toBe('/api/foods/food-1');
		expect(init.method).toBe('PUT');
		expect(body).toStrictEqual(changed);
	});
});

describe('deleteFood', () => {
	it('deletes the food', async () => {
		respond(undefined, STATUS_NO_CONTENT);

		await deleteFood('food-1');

		const { path, init } = sentRequest();
		expect(path).toBe('/api/foods/food-1');
		expect(init.method).toBe('DELETE');
	});
});
