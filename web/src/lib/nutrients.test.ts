import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { ApiError } from '#lib/api.ts';
import {
	type Nutrient,
	type NutrientInput,
	createNutrient,
	deleteNutrient,
	getNutrient,
	listNutrients,
	updateNutrient
} from '#lib/nutrients.ts';

// The tests don't sign in: every request gets this token.
vi.mock('#lib/auth.svelte.ts', (): { idToken: () => Promise<string> } => ({
	idToken: async (): Promise<string> => Promise.resolve('test-token')
}));

const protein: NutrientInput = { name: 'Protein', unit: 'g', kcalPerGram: 4, archived: false };

const storedProtein: Nutrient = { id: 'nutrient-1', ...protein };

const STATUS_OK = 200;
const STATUS_CREATED = 201;
const STATUS_NO_CONTENT = 204;
const STATUS_CONFLICT = 409;

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

describe('listNutrients', () => {
	it('returns the nutrients', async () => {
		const salt: Nutrient = {
			id: 'nutrient-2',
			name: 'Salt',
			unit: 'mg',
			kcalPerGram: 0,
			archived: true
		};
		respond([storedProtein, salt]);

		await expect(listNutrients()).resolves.toStrictEqual([storedProtein, salt]);

		const { path, init } = sentRequest();
		expect(path).toBe('/api/nutrients');
		expect(new Headers(init.headers).get('Authorization')).toBe('Bearer test-token');
	});

	it.each([
		['an unknown unit', { unit: 'oz' }],
		['no archived flag', { archived: undefined }]
	])('rejects a nutrient with %s', async (_name, change) => {
		respond([{ ...storedProtein, ...change }]);

		await expect(listNutrients()).rejects.toThrow('invalid list of nutrients');
	});
});

describe('getNutrient', () => {
	it('returns the nutrient', async () => {
		respond(storedProtein);

		await expect(getNutrient('nutrient-1')).resolves.toStrictEqual(storedProtein);
		expect(sentRequest().path).toBe('/api/nutrients/nutrient-1');
	});

	it('escapes the ID', async () => {
		respond(storedProtein);

		await getNutrient('a/b');
		expect(sentRequest().path).toBe('/api/nutrients/a%2Fb');
	});

	it('rejects an invalid response', async () => {
		respond({ ...storedProtein, kcalPerGram: '4' });

		await expect(getNutrient('nutrient-1')).rejects.toThrow('invalid nutrient');
	});
});

describe('createNutrient', () => {
	it('sends the nutrient and returns it as stored', async () => {
		respond(storedProtein, STATUS_CREATED);

		await expect(createNutrient(protein)).resolves.toStrictEqual(storedProtein);

		const { path, init, body } = sentRequest();
		expect(path).toBe('/api/nutrients');
		expect(init.method).toBe('POST');
		expect(new Headers(init.headers).get('Content-Type')).toBe('application/json');
		expect(body).toStrictEqual(protein);
	});
});

describe('updateNutrient', () => {
	it('sends the nutrient and returns it as stored', async () => {
		const archived = { ...protein, archived: true };
		respond({ id: 'nutrient-1', ...archived });

		await expect(updateNutrient('nutrient-1', archived)).resolves.toStrictEqual({
			id: 'nutrient-1',
			...archived
		});

		const { path, init, body } = sentRequest();
		expect(path).toBe('/api/nutrients/nutrient-1');
		expect(init.method).toBe('PUT');
		expect(body).toStrictEqual(archived);
	});
});

describe('deleteNutrient', () => {
	it('deletes the nutrient', async () => {
		respond(undefined, STATUS_NO_CONTENT);

		await deleteNutrient('nutrient-1');

		const { path, init } = sentRequest();
		expect(path).toBe('/api/nutrients/nutrient-1');
		expect(init.method).toBe('DELETE');
	});

	it('reports that a nutrient must be archived first', async () => {
		fetchMock.mockResolvedValueOnce(
			new Response('Only archived nutrients can be deleted.\n', { status: STATUS_CONFLICT })
		);

		const error: unknown = await deleteNutrient('nutrient-1').catch((error_: unknown) => error_);
		expect(error).toBeInstanceOf(ApiError);
		expect(String(error)).toContain('Only archived nutrients can be deleted.');
	});
});
