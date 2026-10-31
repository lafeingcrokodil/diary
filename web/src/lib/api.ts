import { idToken } from '#lib/auth.svelte.ts';

const STATUS_FORBIDDEN = 403;

// An error response from the API.
export class ApiError extends Error {
	public readonly status: number;

	public constructor(status: number) {
		super(`The API responded with status ${String(status)}.`);
		this.name = 'ApiError';
		this.status = status;
	}

	// Whether the user is signed in, but isn't allowed to do what they asked.
	public get isForbidden(): boolean {
		return this.status === STATUS_FORBIDDEN;
	}
}

// Sends a request to the API as the signed-in user, and throws an ApiError if
// the response isn't successful.
export const apiFetch = async (path: string, init: RequestInit = {}): Promise<Response> => {
	const headers = new Headers(init.headers);
	headers.set('Authorization', `Bearer ${await idToken()}`);

	const response = await fetch(path, { ...init, headers });
	if (!response.ok) {
		throw new ApiError(response.status);
	}
	return response;
};

export interface Me {
	email: string;
}

const isMe = (value: unknown): value is Me =>
	typeof value === 'object' &&
	value !== null &&
	'email' in value &&
	typeof value.email === 'string';

// Returns the signed-in user, as the API sees them. If the user isn't allowed
// to use the app, it throws an ApiError whose isForbidden is true.
export const getMe = async (): Promise<Me> => {
	const response = await apiFetch('/api/me');
	const me: unknown = await response.json();
	if (!isMe(me)) {
		throw new Error('The API responded with an invalid user.');
	}
	return me;
};
