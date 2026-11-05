import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { type Day, getDay, putDay } from '#lib/days.ts';

// The tests don't sign in: every request gets this token.
vi.mock('#lib/auth.svelte.ts', (): { idToken: () => Promise<string> } => ({
	idToken: async (): Promise<string> => Promise.resolve('test-token')
}));

const monday: Day = { date: '2026-11-02', weightKg: 70.5, kcalExpended: 2300 };

const fetchMock = vi.fn<typeof fetch>();

// Makes the next request get a response with the given body.
const respond = (body: unknown): void => {
	fetchMock.mockResolvedValueOnce(new Response(JSON.stringify(body)));
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

describe('getDay', () => {
	it('returns the day', async () => {
		respond(monday);

		await expect(getDay('2026-11-02')).resolves.toStrictEqual(monday);

		const { path, init } = sentRequest();
		expect(path).toBe('/api/days/2026-11-02');
		expect(new Headers(init.headers).get('Authorization')).toBe('Bearer test-token');
	});

	it('returns a day that nothing is logged for', async () => {
		const empty: Day = { date: '2026-11-02', weightKg: null, kcalExpended: null };
		respond(empty);

		await expect(getDay('2026-11-02')).resolves.toStrictEqual(empty);
	});

	it.each([
		['a weight that is a string', { weightKg: '70.5' }],
		['no energy expended', { kcalExpended: undefined }],
		['no date', { date: undefined }]
	])('rejects a day with %s', async (_name, change) => {
		respond({ ...monday, ...change });

		await expect(getDay('2026-11-02')).rejects.toThrow('invalid day');
	});
});

describe('putDay', () => {
	it('sends the day and returns it as stored', async () => {
		respond(monday);

		const input = { weightKg: 70.5, kcalExpended: 2300 };
		await expect(putDay('2026-11-02', input)).resolves.toStrictEqual(monday);

		const { path, init, body } = sentRequest();
		expect(path).toBe('/api/days/2026-11-02');
		expect(init.method).toBe('PUT');
		expect(new Headers(init.headers).get('Content-Type')).toBe('application/json');
		expect(body).toStrictEqual(input);
	});
});
