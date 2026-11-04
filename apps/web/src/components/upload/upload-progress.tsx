"use client";

import type { UploadItem } from "~/lib/upload";
import { Progress } from "~/components/ui/progress";
import { Button } from "~/components/ui/button";
import { cn } from "~/lib/utils";
import {
	CheckCircle2,
	AlertCircle,
	Pause,
	Play,
	RotateCw,
	X,
	Loader2,
	Ban,
	Clock,
} from "lucide-react";

export interface UploadProgressProps {
	item: UploadItem;
	onPause?: () => void;
	onResume?: () => void;
	onCancel?: () => void;
	onRetry?: () => void;
	className?: string;
}

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

const statusConfig = {
	queued: { icon: Clock, label: "Queued", color: "text-muted-foreground" },
	preparing: { icon: Loader2, label: "Preparing", color: "text-primary" },
	uploading: { icon: Loader2, label: "Uploading", color: "text-primary" },
	paused: { icon: Pause, label: "Paused", color: "text-amber-500" },
	success: {
		icon: CheckCircle2,
		label: "Completed",
		color: "text-emerald-500",
	},
	error: { icon: AlertCircle, label: "Failed", color: "text-destructive" },
	canceled: { icon: Ban, label: "Canceled", color: "text-muted-foreground" },
};

export function UploadProgress({
	item,
	onPause,
	onResume,
	onCancel,
	onRetry,
	className,
}: UploadProgressProps) {
	const config = statusConfig[item.status];
	const StatusIcon = config.icon;

	return (
		<div
			className={cn(
				"group relative flex items-center gap-3 rounded-[var(--radius-md)] border border-border bg-card p-3 text-card-foreground shadow-xs transition-colors hover:border-primary/40",
				className,
			)}
		>
			<StatusIcon
				className={cn(
					"h-5 w-5 shrink-0",
					config.color,
					(item.status === "preparing" || item.status === "uploading") &&
						"animate-spin",
				)}
			/>

			<div className="min-w-0 flex-1 space-y-1">
				<div className="flex items-center justify-between gap-2">
					<p
						className="truncate text-sm font-medium text-foreground"
						title={item.relativePath || item.fileName}
					>
						{item.relativePath || item.fileName}
					</p>
					<span className="shrink-0 text-xs text-muted-foreground">
						{formatBytes(item.fileSize)}
					</span>
				</div>

				{item.targetFolderName && (
					<p className="truncate text-[11px] text-muted-foreground/80">
						to: {item.targetFolderName}
					</p>
				)}

				{(item.status === "preparing" ||
					item.status === "uploading" ||
					item.status === "paused") && (
					<div className="space-y-1">
						<Progress value={item.progress} className="h-1.5" />
						<div className="flex items-center justify-between text-[11px] text-muted-foreground">
							<span>{item.progress}%</span>
							{item.status === "uploading" &&
								(item.speedBps || item.etaSeconds !== undefined) && (
									<span>
										{[formatSpeed(item.speedBps), formatEta(item.etaSeconds)]
											.filter(Boolean)
											.join(" · ")}
									</span>
								)}
						</div>
					</div>
				)}

				{item.status === "error" && (item.errorMessage || item.error) && (
					<p className="truncate text-xs text-destructive">
						{item.errorMessage || item.error}
						{(item.retryCount ?? 0) > 0 && ` (attempt ${item.retryCount! + 1})`}
					</p>
				)}
			</div>

			<div className="flex shrink-0 items-center gap-1 opacity-0 transition-opacity group-hover:opacity-100">
				{item.status === "uploading" && onPause && (
					<Button
						variant="ghost"
						size="icon"
						onClick={onPause}
						className="h-7 w-7"
						aria-label={`Pause upload of ${item.fileName}`}
					>
						<Pause className="h-3.5 w-3.5" />
					</Button>
				)}

				{item.status === "paused" && onResume && (
					<Button
						variant="ghost"
						size="icon"
						onClick={onResume}
						className="h-7 w-7"
						aria-label={`Resume upload of ${item.fileName}`}
					>
						<Play className="h-3.5 w-3.5" />
					</Button>
				)}

				{item.status === "error" && onRetry && (
					<Button
						variant="ghost"
						size="icon"
						onClick={onRetry}
						className="h-7 w-7"
						aria-label={`Retry upload of ${item.fileName}`}
					>
						<RotateCw className="h-3.5 w-3.5" />
					</Button>
				)}

				{item.status !== "success" &&
					item.status !== "canceled" &&
					onCancel && (
						<Button
							variant="ghost"
							size="icon"
							onClick={onCancel}
							className="h-7 w-7 text-muted-foreground hover:text-destructive"
							aria-label={`Cancel upload of ${item.fileName}`}
						>
							<X className="h-3.5 w-3.5" />
						</Button>
					)}
			</div>
		</div>
	);
}
