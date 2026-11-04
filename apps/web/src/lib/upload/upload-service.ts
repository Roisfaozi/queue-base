import { api } from "~/lib/api/client";

export interface UploadRequest {
	file: File;
	targetFolderId?: string;
	relativePath?: string;
	signal?: AbortSignal;
	onProgress?: (progress: number) => void;
}

export interface UploadResult {
	url: string;
}

const UPLOAD_ENDPOINT = "/upload/files/";

export const uploadService = {
	async uploadFile({
		file,
		targetFolderId,
		relativePath,
		signal,
		onProgress,
	}: UploadRequest): Promise<UploadResult> {
		const form = new FormData();
		form.append("file", file);
		if (targetFolderId) form.append("folderId", targetFolderId);
		if (relativePath) form.append("relativePath", relativePath);

		return api.request<UploadResult>(UPLOAD_ENDPOINT, {
			method: "POST",
			body: form,
			signal,
		});
	},
};
