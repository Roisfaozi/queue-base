import { useUploadStore } from "@casbin/hooks";
import { UploadProgress } from "./upload-progress";
import { NexusButton } from "../nexus-button";
import { Progress } from "../progress";
import { cn } from "../lib/utils";
import {
	ChevronUp,
	ChevronDown,
	X,
	CheckCircle2,
	AlertCircle,
	Loader2,
	RotateCw,
	Ban,
} from "lucide-react";

function formatBytes(bytes: number) {
	if (bytes === 0) return "0 B";
	const k = 1024;
	const sizes = ["B", "KB", "MB", "GB"];
	const i = Math.floor(Math.log(bytes) / Math.log(k));
	return `${parseFloat((bytes / k ** i).toFixed(1))} ${sizes[i]}`;
}

function formatSpeed(bytesPerSec?: number) {
	if (!bytesPerSec || bytesPerSec <= 0) return "";
	return `${formatBytes(bytesPerSec)}/s`;
}

function formatEta(seconds?: number) {
	if (seconds === undefined || seconds < 0) return "";
	if (seconds === 0) return "finishing...";
	if (seconds < 60) return `${seconds}s left`;
	const m = Math.floor(seconds / 60);
	const s = seconds % 60;
	if (m < 60) return `${m}m ${s}s left`;
	const h = Math.floor(m / 60);
	return `${h}h ${m % 60}m left`;
}

export function UploadManager() {
	const items = useUploadStore((s) => s.uploadQueue);
	const open = useUploadStore((s) => s.managerOpen);
	const setOpen = useUploadStore((s) => s.setManagerOpen);
	const pauseUpload = useUploadStore((s) => s.pauseUpload);
	const resumeUpload = useUploadStore((s) => s.resumeUpload);
	const cancelUpload = useUploadStore((s) => s.cancelUpload);
	const retryUpload = useUploadStore((s) => s.retryUpload);
	const clearCompleted = useUploadStore((s) => s.clearCompleted);
	const retryAllFailed = useUploadStore((s) => s.retryAllFailed);
	const cancelAllUploading = useUploadStore((s) => s.cancelAllUploading);
	const cancelAllQueued = useUploadStore((s) => s.cancelAllQueued);

	const summary = {
		total: items.length,
		active: items.filter(
			(i) => i.status === "preparing" || i.status === "uploading",
		).length,
		queued: items.filter((i) => i.status === "queued").length,
		done: items.filter((i) => i.status === "success").length,
		failed: items.filter((i) => i.status === "error").length,
		canceled: items.filter((i) => i.status === "canceled").length,
		progress: items.length
			? Math.round(items.reduce((s, i) => s + i.progress, 0) / items.length)
			: 0,
		speedBps: items
			.filter((i) => i.status === "uploading")
			.reduce((s, i) => s + (i.speedBps ?? 0), 0),
		etaSeconds: (() => {
			const activeItems = items.filter((i) => i.status === "uploading");
			const totalSpeed = activeItems.reduce((s, i) => s + (i.speedBps ?? 0), 0);
			const remainingBytes = activeItems.reduce(
				(s, i) => s + Math.max(0, i.fileSize - (i.bytesUploaded ?? 0)),
				0,
			);
			return totalSpeed > 0
				? Math.round(remainingBytes / totalSpeed)
				: undefined;
		})(),
	};

	if (summary.total === 0) return null;

	const showSummaryLine = [
		`Uploading ${summary.active} of ${summary.total} file${summary.total > 1 ? "s" : ""}`,
		summary.done > 0 ? `${summary.done} completed` : null,
		summary.failed > 0 ? `${summary.failed} failed` : null,
		summary.canceled > 0 ? `${summary.canceled} canceled` : null,
		formatSpeed(summary.speedBps),
		formatEta(summary.etaSeconds),
	]
		.filter(Boolean)
		.join(" · ");

	return (
		<aside
			aria-label="Upload manager"
			aria-live="polite"
			className="fixed right-4 bottom-4 z-50 w-[calc(100vw-2rem)] max-w-md"
		>
			<div className="bg-card text-card-foreground border-border overflow-hidden rounded-[var(--radius-lg,12px)] border shadow-xl">
				{/* Header */}
				<div
					onClick={() => setOpen(!open)}
					className="hover:bg-muted/30 flex cursor-pointer items-center justify-between gap-3 border-b p-3 transition-colors"
				>
					<div className="flex min-w-0 items-center gap-2.5">
						{summary.active > 0 ? (
							<Loader2 className="text-primary h-4 w-4 shrink-0 animate-spin" />
						) : summary.failed > 0 ? (
							<AlertCircle className="text-danger h-4 w-4 shrink-0" />
						) : (
							<CheckCircle2 className="text-success h-4 w-4 shrink-0" />
						)}

						<div className="min-w-0">
							<p className="text-foreground truncate text-xs font-medium">
								{summary.active > 0 ? showSummaryLine : "No active uploads"}
							</p>
						</div>
					</div>

					<div className="flex shrink-0 items-center gap-1">
						<NexusButton
							variant="ghost"
							size="icon"
							onClick={(e) => {
								e.stopPropagation();
								setOpen(!open);
							}}
							className="h-7 w-7"
							aria-label={open ? "Collapse uploads" : "Expand uploads"}
						>
							{open ? (
								<ChevronDown className="h-4 w-4" />
							) : (
								<ChevronUp className="h-4 w-4" />
							)}
						</NexusButton>

						{summary.done > 0 && (
							<NexusButton
								variant="ghost"
								size="icon"
								onClick={(e) => {
									e.stopPropagation();
									clearCompleted();
								}}
								className="text-muted-foreground hover:text-danger h-7 w-7"
								aria-label="Clear completed uploads"
							>
								<X className="h-3.5 w-3.5" />
							</NexusButton>
						)}
					</div>
				</div>

				{/* Collapsed mini-progress */}
				{!open && summary.active > 0 && (
					<Progress value={summary.progress} className="h-1 rounded-none" />
				)}

				{/* Expanded Panel */}
				{open && (
					<div className="space-y-3 p-3">
						{/* Bulk control actions */}
						<div className="flex flex-wrap items-center justify-between gap-2 border-b pb-2 text-xs">
							<span className="text-muted-foreground">
								Overall: {summary.progress}%
							</span>

							<div className="flex items-center gap-1">
								{summary.failed > 0 && (
									<NexusButton
										variant="ghost"
										size="sm"
										onClick={retryAllFailed}
										className="h-7 px-2 text-xs"
									>
										<RotateCw className="mr-1 h-3 w-3" />
										Retry failed
									</NexusButton>
								)}

								{summary.active > 0 && (
									<NexusButton
										variant="ghost"
										size="sm"
										onClick={cancelAllUploading}
										className="text-muted-foreground hover:text-danger h-7 px-2 text-xs"
									>
										<Ban className="mr-1 h-3 w-3" />
										Cancel active
									</NexusButton>
								)}

								{summary.queued > 0 && (
									<NexusButton
										variant="ghost"
										size="sm"
										onClick={cancelAllQueued}
										className="text-muted-foreground hover:text-danger h-7 px-2 text-xs"
									>
										Cancel queued
									</NexusButton>
								)}
							</div>
						</div>

						{/* List */}
						<div className="max-h-[380px] space-y-2 overflow-y-auto pr-1">
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
					</div>
				)}
			</div>
		</aside>
	);
}
