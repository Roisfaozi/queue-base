/**
 * Validates whether a file matches the standard HTML `accept` specification.
 * Supports:
 * - Extension matching: ".pdf", ".docx", ".png"
 * - Wildcard MIME matching: "image/*", "video/*", "audio/*"
 * - Exact MIME matching: "application/pdf", "image/png"
 */
export function isAcceptedFile(file: File, accept?: string): boolean {
	if (!accept || accept.trim() === "" || accept === "*") return true;

	const acceptedTypes = accept
		.split(",")
		.map((t) => t.trim().toLowerCase())
		.filter(Boolean);

	const fileName = file.name.toLowerCase();
	const fileType = (file.type || "").toLowerCase();

	return acceptedTypes.some((pattern) => {
		if (pattern.startsWith(".")) {
			return fileName.endsWith(pattern);
		}
		if (pattern.endsWith("/*")) {
			const prefix = pattern.slice(0, -1);
			return fileType.startsWith(prefix);
		}
		return fileType === pattern;
	});
}
