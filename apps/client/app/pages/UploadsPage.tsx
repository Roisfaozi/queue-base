import { PageHeader } from "@/components/layout/page-header";
import { FileUploader, UploadQueue, FilePreview, NexusCard } from "@casbin/ui";
import { useUploadStore } from "@casbin/hooks";

export default function UploadsPage() {
	const items = useUploadStore((s) => s.uploadQueue);
	const completed = items.filter((i) => i.status === "success");

	return (
		<div className="space-y-6">
			<PageHeader
				title="File Upload"
				description="Resumable file uploads with TUS protocol — drag & drop, pause, resume, retry"
			/>

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
				<div>
					<h3 className="text-foreground mb-3 text-sm font-semibold">
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
				<NexusCard className="text-muted-foreground p-8 text-center text-sm">
					No uploads yet. Drop files above or click to browse.
				</NexusCard>
			)}
		</div>
	);
}
