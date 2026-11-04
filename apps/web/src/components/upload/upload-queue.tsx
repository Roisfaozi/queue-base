"use client";

import { useUploadStore } from "~/lib/upload";
import { UploadProgress } from "./upload-progress";
import { Button } from "~/components/ui/button";
import { Badge } from "~/components/ui/badge";
import { Card } from "~/components/ui/card";
import { cn } from "~/lib/utils";
import { Play, Trash2 } from "lucide-react";

export interface UploadQueueProps {
	className?: string;
}

export function UploadQueue({ className }: UploadQueueProps) {
	const items = useUploadStore((s) => s.uploadQueue);
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
		<Card className={cn("space-y-4 p-4", className)}>
			{/* Header */}
			<div className="flex flex-wrap items-center justify-between gap-2 border-b pb-3">
				<div className="flex items-center gap-2">
					<h3 className="text-sm font-semibold text-foreground">
						Upload Queue
					</h3>
					<span className="text-xs text-muted-foreground">
						({items.length} {items.length === 1 ? "file" : "files"})
					</span>
				</div>

				<div className="flex flex-wrap items-center gap-1.5">
					{uploading > 0 && <Badge variant="info">{uploading} uploading</Badge>}
					{queued > 0 && <Badge variant="secondary">{queued} queued</Badge>}
					{completed > 0 && <Badge variant="success">{completed} done</Badge>}
					{errors > 0 && <Badge variant="destructive">{errors} failed</Badge>}
					{canceled > 0 && (
						<Badge variant="secondary">{canceled} canceled</Badge>
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
					<Button
						variant="outline"
						size="sm"
						onClick={startAll}
						className="gap-1.5"
					>
						<Play className="h-3.5 w-3.5" />
						Start All
					</Button>
				) : (
					<div />
				)}

				{completed > 0 && (
					<Button
						variant="ghost"
						size="sm"
						onClick={clearCompleted}
						className="gap-1.5 text-muted-foreground hover:text-foreground"
					>
						<Trash2 className="h-3.5 w-3.5" />
						Clear Completed
					</Button>
				)}
			</div>
		</Card>
	);
}
