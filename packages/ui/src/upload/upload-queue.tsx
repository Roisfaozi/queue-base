import { useUploadStore } from "@casbin/hooks";
import { UploadProgress } from "./upload-progress";
import { NexusButton } from "../nexus-button";
import { NexusBadge } from "../nexus-badge";
import { NexusCard } from "../nexus-card";
import { cn } from "../lib/utils";
import { Play, Check, Trash2 } from "lucide-react";

export interface UploadQueueProps {
	className?: string;
}

export function UploadQueue({ className }: UploadQueueProps) {
	const items = useUploadStore((s) => s.uploadQueue);
	const startUpload = useUploadStore((s) => s.startUpload);
	const startAll = useUploadStore((s) => s.startAll);
	const pauseUpload = useUploadStore((s) => s.pauseUpload);
	const resumeUpload = useUploadStore((s) => s.resumeUpload);
	const cancelUpload = useUploadStore((s) => s.cancelUpload);
	const retryUpload = useUploadStore((s) => s.retryUpload);
	const clearCompleted = useUploadStore((s) => s.clearCompleted);

	if (items.length === 0) return null;

	const uploading = items.filter(
		(i) => i.status === "preparing" || i.status === "uploading",
	).length;
	const queued = items.filter((i) => i.status === "queued").length;
	const completed = items.filter((i) => i.status === "success").length;
	const errors = items.filter((i) => i.status === "error").length;
	const canceled = items.filter((i) => i.status === "canceled").length;

	return (
		<NexusCard className={cn("space-y-4", className)}>
			{/* Header */}
			<div className="flex flex-wrap items-center justify-between gap-2 border-b pb-3">
				<div className="flex items-center gap-2">
					<h3 className="text-foreground text-sm font-semibold">
						Upload Queue
					</h3>
					<span className="text-muted-foreground text-xs">
						({items.length} {items.length === 1 ? "file" : "files"})
					</span>
				</div>

				<div className="flex flex-wrap items-center gap-1.5">
					{uploading > 0 && (
						<NexusBadge variant="info">{uploading} uploading</NexusBadge>
					)}
					{queued > 0 && (
						<NexusBadge variant="neutral">{queued} queued</NexusBadge>
					)}
					{completed > 0 && (
						<NexusBadge variant="success">{completed} done</NexusBadge>
					)}
					{errors > 0 && (
						<NexusBadge variant="danger">{errors} failed</NexusBadge>
					)}
					{canceled > 0 && (
						<NexusBadge variant="neutral">{canceled} canceled</NexusBadge>
					)}
				</div>
			</div>

			{/* List */}
			<div className="max-h-[360px] space-y-2 overflow-y-auto pr-1">
				{items.map((item) => (
					<UploadProgress
						key={item.id}
						item={item}
						onPause={() => pauseUpload(item.id)}
						onResume={() => resumeUpload(item.id)}
						onCancel={() => cancelUpload(item.id)}
						onRetry={() => retryUpload(item.id)}
					/>
				))}
			</div>

			{/* Actions */}
			<div className="flex items-center justify-between border-t pt-3">
				{queued > 0 ? (
					<NexusButton
						variant="outline"
						size="sm"
						onClick={startAll}
						className="gap-1.5"
					>
						<Play className="h-3.5 w-3.5" />
						Start All
					</NexusButton>
				) : (
					<div />
				)}

				{completed > 0 && (
					<NexusButton
						variant="ghost"
						size="sm"
						onClick={clearCompleted}
						className="text-muted-foreground gap-1.5"
					>
						<Trash2 className="h-3.5 w-3.5" />
						Clear Completed
					</NexusButton>
				)}
			</div>
		</NexusCard>
	);
}
