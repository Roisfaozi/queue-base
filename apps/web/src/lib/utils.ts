import { generateRandomString, type RandomReader } from "@oslojs/crypto/random";
export { cn, nFormatter, hasFileNameSpaces, formatDate } from "@casbin/utils";

export class FreePlanLimitError extends Error {
	constructor(message?: string) {
		super(message || "Free plan limit reached. Please upgrade your plan.");
		this.name = "FreePlanLimitError";
	}
}

export function isRedirectError(error: unknown): boolean {
	return (
		error !== null &&
		typeof error === "object" &&
		"digest" in error &&
		typeof error.digest === "string" &&
		error.digest.includes("NEXT_REDIRECT")
	);
}

const alphanumeric =
	"abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789";

export function generateId(length = 10): string {
	const random: RandomReader = {
		read(bytes) {
			crypto.getRandomValues(bytes);
		},
	};
	return generateRandomString(random, alphanumeric, length);
}

/**
 * Validates and sanitizes returnTo parameters to prevent Open Redirect attacks.
 * Only allows relative paths on the same host (starting with a single "/").
 * Strictly blocks external schemes, protocol-relative paths ("//"), backslashes ("/\"),
 * control characters, and malformed inputs.
 */
export function getSafeReturnTo(
	returnTo: string | null | undefined,
	defaultPath = "/dashboard",
): string {
	if (!returnTo) return defaultPath;

	try {
		const decoded = decodeURIComponent(returnTo).trim();

		// Reject any string with control characters or tabs
		if (/[\r\n\t]/.test(decoded)) {
			return defaultPath;
		}

		// Must start with exactly one forward slash, not "//", "/\", or "\\"
		if (
			!decoded.startsWith("/") ||
			decoded.startsWith("//") ||
			decoded.startsWith("/\\") ||
			decoded.startsWith("\\")
		) {
			return defaultPath;
		}

		// Parse against a dummy base origin to verify host and protocol do not diverge
		const dummyBase = "http://localhost";
		const parsed = new URL(decoded, dummyBase);

		if (parsed.origin !== dummyBase || parsed.protocol !== "http:") {
			return defaultPath;
		}

		return parsed.pathname + parsed.search + parsed.hash;
	} catch {
		return defaultPath;
	}
}
