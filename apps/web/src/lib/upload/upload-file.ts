import * as tus from "tus-js-client";
import { organizationsApi } from "~/lib/api/organizations";
import { branchesApi } from "~/lib/api/qms";
import { usersApi } from "~/lib/api/users";
import { mapUploadError } from "./upload-errors";

/**
 * Files smaller than this threshold (15 MB) use direct multipart HTTP upload
 * when a feature-specific direct endpoint is available (avatar, organization logo, branch logo).
 * Files at or above this threshold automatically use resumable TUS upload.
 */
export const TUS_THRESHOLD_BYTES = 15 * 1024 * 1024;

export type UploadFeatureType =
	| "avatar"
	| "organization_logo"
	| "branch_logo"
	| "generic"
	| (string & {});

export interface SingleUploadOptions {
	uploadType?: UploadFeatureType;
	endpoint?: string;
	metadata?: Record<string, string>;
	onProgress?: (progress: number) => void;
	/** Custom byte threshold before switching from direct upload to TUS (default: 15MB) */
	tusThresholdBytes?: number;
	/** Optional explicit direct-upload handler used when file.size < tusThresholdBytes */
	directUpload?: (file: File) => Promise<string>;
}

function resolveDirectUploadHandler(
	uploadType: string,
	customDirect?: (file: File) => Promise<string>,
): ((file: File) => Promise<string>) | null {
	if (customDirect) {
		return customDirect;
	}

	switch (uploadType) {
		case "avatar":
			return async (file: File) => {
				const res = await usersApi.uploadAvatar(file);
				return res.data?.avatar_url || "";
			};
		case "organization_logo":
			return async (file: File) => {
				const res = await organizationsApi.uploadLogo(file);
				return res.data?.url || "";
			};
		case "branch_logo":
		case "logo":
			return async (file: File) => {
				const res = await branchesApi.uploadLogo(file);
				return res.data?.url || "";
			};
		default:
			return null;
	}
}

/**
 * Uploads a single file:
 * - If `file.size < 15MB` and a feature-specific direct endpoint exists (`avatar`, `organization_logo`, `branch_logo`),
 *   performs a standard multipart HTTP upload.
 * - If `file.size >= 15MB` (or no direct endpoint applies), uses the TUS resumable upload protocol.
 */
export async function uploadSingleFile(
	file: File,
	options: SingleUploadOptions = {},
): Promise<string> {
	const uploadType = options.uploadType || "generic";
	const threshold = options.tusThresholdBytes ?? TUS_THRESHOLD_BYTES;
	const directHandler = resolveDirectUploadHandler(
		uploadType,
		options.directUpload,
	);

	// Regular multipart upload when file is strictly under 15MB
	if (file.size < threshold && directHandler) {
		options.onProgress?.(25);
		const url = await directHandler(file);
		options.onProgress?.(100);
		return url;
	}

	// Resumable TUS upload for files >= 15MB (or generic uploads)
	return new Promise((resolve, reject) => {
		const endpoint = options.endpoint || "/api/v1/upload/files/";
		// Backend TUS registry accepts "avatar" or "generic"
		const tusRegisteredType = uploadType === "avatar" ? "avatar" : "generic";

		const metadata: Record<string, string> = {
			type: tusRegisteredType,
			scope: uploadType,
			filename: file.name,
			filetype: file.type || "application/octet-stream",
			filesize: String(file.size),
			...(options.metadata || {}),
		};

		const upload = new tus.Upload(file, {
			endpoint,
			chunkSize: 5 * 1024 * 1024,
			retryDelays: [0, 1000, 3000, 5000],
			metadata,
			onProgress: (bytesUploaded, bytesTotal) => {
				const pct = Math.round((bytesUploaded / bytesTotal) * 100);
				options.onProgress?.(pct);
			},
			onSuccess: () => {
				resolve(upload.url || "");
			},
			onError: (error) => {
				const mapped = mapUploadError(error);
				reject(new Error(mapped.message || error.message));
			},
		});

		upload.start();
	});
}
