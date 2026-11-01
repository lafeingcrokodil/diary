// Reads a number that the user typed, which may use a decimal comma, or
// returns undefined if it isn't a finite number.
export const readNumber = (text: string): number | undefined => {
	const trimmed = text.trim();
	if (trimmed === '') {
		return undefined;
	}
	const value = Number(trimmed.replace(',', '.'));
	return Number.isFinite(value) ? value : undefined;
};
