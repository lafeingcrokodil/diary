import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import {
	type Recipe,
	type RecipeInput,
	createRecipe,
	deleteRecipe,
	getRecipe,
	listRecipes,
	updateRecipe
} from '#lib/recipes.ts';

// The tests don't sign in: every request gets this token.
vi.mock('#lib/auth.svelte.ts', (): { idToken: () => Promise<string> } => ({
	idToken: async (): Promise<string> => Promise.resolve('test-token')
}));

const soup: RecipeInput = {
	name: 'Lentil soup',
	sourceUrl: 'https://example.com/lentil-soup',
	ingredients: [
		{ foodId: 'lentils', serving: '', amount: 200, note: 'rinsed' },
		{ foodId: 'onion', serving: 'medium onion', amount: 1, note: '' }
	],
	steps: ['Fry the onion.', 'Add the lentils and simmer.'],
	notes: 'Freezes well.',
	archived: false
};

const storedSoup: Recipe = {
	id: 'recipe-1',
	...soup,
	nutrition: {
		kcal: { amount: 760, missing: [] },
		nutrients: { protein: { amount: 51.5, missing: [] }, fibre: { amount: 20, missing: [1] } }
	}
};

const STATUS_OK = 200;
const STATUS_CREATED = 201;
const STATUS_NO_CONTENT = 204;

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

describe('listRecipes', () => {
	it('returns the recipes', async () => {
		respond([storedSoup]);

		await expect(listRecipes()).resolves.toStrictEqual([storedSoup]);

		const { path, init } = sentRequest();
		expect(path).toBe('/api/recipes');
		expect(new Headers(init.headers).get('Authorization')).toBe('Bearer test-token');
	});

	it.each([
		['no source', { sourceUrl: undefined }],
		[
			'an ingredient without a note',
			{ ingredients: [{ foodId: 'onion', serving: '', amount: 1 }] }
		],
		["an amount that isn't a number", { ingredients: [{ ...soup.ingredients[0], amount: '1' }] }],
		["a step that isn't a string", { steps: [1] }],
		['no archived flag', { archived: undefined }],
		['no nutrition', { nutrition: undefined }],
		['a list of nutrients', { nutrition: { ...storedSoup.nutrition, nutrients: [] } }],
		['an invalid total', { nutrition: { ...storedSoup.nutrition, kcal: { amount: 1 } } }],
		[
			'an invalid missing ingredient',
			{ nutrition: { ...storedSoup.nutrition, kcal: { amount: 1, missing: ['1'] } } }
		]
	])('rejects a recipe with %s', async (_name, change) => {
		respond([{ ...storedSoup, ...change }]);

		await expect(listRecipes()).rejects.toThrow('invalid list of recipes');
	});
});

describe('getRecipe', () => {
	it('returns the recipe', async () => {
		respond(storedSoup);

		await expect(getRecipe('a/b')).resolves.toStrictEqual(storedSoup);
		expect(sentRequest().path).toBe('/api/recipes/a%2Fb');
	});

	it('rejects an invalid response', async () => {
		respond({ ...storedSoup, steps: 'Boil.' });

		await expect(getRecipe('recipe-1')).rejects.toThrow('invalid recipe');
	});
});

describe('createRecipe', () => {
	it('sends the recipe and returns it as stored', async () => {
		respond(storedSoup, STATUS_CREATED);

		await expect(createRecipe(soup)).resolves.toStrictEqual(storedSoup);

		const { path, init, body } = sentRequest();
		expect(path).toBe('/api/recipes');
		expect(init.method).toBe('POST');
		expect(new Headers(init.headers).get('Content-Type')).toBe('application/json');
		expect(body).toStrictEqual(soup);
	});
});

describe('updateRecipe', () => {
	it("sends only the input's fields and returns the recipe as stored", async () => {
		respond(storedSoup);

		// A stored recipe works as input, but its ID and nutrition aren't sent.
		await expect(updateRecipe('recipe-1', storedSoup)).resolves.toStrictEqual(storedSoup);

		const { path, init, body } = sentRequest();
		expect(path).toBe('/api/recipes/recipe-1');
		expect(init.method).toBe('PUT');
		expect(body).toStrictEqual(soup);
	});
});

describe('deleteRecipe', () => {
	it('deletes the recipe', async () => {
		respond(undefined, STATUS_NO_CONTENT);

		await deleteRecipe('recipe-1');

		const { path, init } = sentRequest();
		expect(path).toBe('/api/recipes/recipe-1');
		expect(init.method).toBe('DELETE');
	});
});
