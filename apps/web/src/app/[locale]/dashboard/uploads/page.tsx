"use client";

import { FileUploader, UploadQueue, FilePreview } from "~/components/upload";
import { Card } from "~/components/ui/card";
import { useUploadStore } from "~/lib/upload";

export default function UploadsDashboardPage() {
	const items = useUploadStore((s) => s.uploadQueue);
	const completed = items.filter((i) => i.status === "success");

	return (
		<div className="space-y-6">
			<div>
				<h1 className="text-foreground text-2xl font-bold tracking-tight">
					File & Folder Upload
				</h1>
				<p className="text-muted-foreground mt-1 text-sm">
					Resumable file and directory uploads with TUS protocol — drag & drop,
					pause, resume, and retry.
				</p>
			</div>

			<div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
				<FileUploader />
				<FileUploader
					directory
					maxFiles={200}
					title="Upload folder"
					description="Select a folder · Files keep their relative paths"
					className="min-h-[188px]"
				/>
			</div>

			<UploadQueue />

			{completed.length > 0 && (
				<div className="space-y-3">
					<h3 className="text-foreground text-sm font-semibold">
						Completed Uploads
					</h3>
					<div className="grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4">
						{completed.map((item) => (
							<FilePreview key={item.id} file={item.file} url={item.url} />
						))}
					</div>
				</div>
			)}

			{items.length === 0 && (
				<Card className="p-8 text-center text-sm text-muted-foreground">
					No uploads yet. Drop files or folder above, or paste anywhere
					(Ctrl+V).
				</Card>
			)}
		</div>
	);
}
