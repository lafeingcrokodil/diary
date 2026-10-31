import { type FirebaseOptions, initializeApp } from 'firebase/app';
import {
	type Auth,
	GoogleAuthProvider,
	type User,
	getAuth,
	getRedirectResult,
	onAuthStateChanged,
	signOut as firebaseSignOut,
	signInWithRedirect
} from 'firebase/auth';

interface Session {
	// Undefined until Firebase has determined whether someone is signed in.
	user: User | null | undefined;
	error: string | undefined;
}

export const session = $state<Session>({ user: undefined, error: undefined });

let auth: Auth | undefined;

// Checks the fields that Firebase Authentication needs.
const isFirebaseOptions = (value: unknown): value is FirebaseOptions =>
	typeof value === 'object' &&
	value !== null &&
	'apiKey' in value &&
	typeof value.apiKey === 'string' &&
	'projectId' in value &&
	typeof value.projectId === 'string';

// Fetches the project's Firebase configuration, which Firebase Hosting serves
// at a reserved URL.
const loadConfig = async (): Promise<FirebaseOptions> => {
	const response = await fetch('/__/firebase/init.json');
	if (!response.ok) {
		throw new Error(`Loading the Firebase configuration failed: ${String(response.status)}`);
	}
	const config: unknown = await response.json();
	if (!isFirebaseOptions(config)) {
		throw new Error('The Firebase configuration is invalid.');
	}
	return config;
};

// Initializes Firebase and starts tracking who is signed in.
export const initAuth = async (): Promise<void> => {
	const config = await loadConfig();

	// Run the sign-in flow on the app's own domain rather than on the
	// configured firebaseapp.com domain, so that it doesn't depend on
	// third-party cookies, which some browsers block.
	auth = getAuth(initializeApp({ ...config, authDomain: location.host }));

	onAuthStateChanged(auth, (user) => {
		session.user = user;
	});

	try {
		// Completes a sign-in redirect, if there was one, and reports its errors.
		await getRedirectResult(auth);
	} catch (error) {
		session.error = `Signing in failed: ${String(error)}`;
	}
};

const requireAuth = (): Auth => {
	if (auth === undefined) {
		throw new Error('Firebase Authentication has not been initialized.');
	}
	return auth;
};

export const signIn = async (): Promise<void> => {
	await signInWithRedirect(requireAuth(), new GoogleAuthProvider());
};

export const signOut = async (): Promise<void> => {
	await firebaseSignOut(requireAuth());
};
