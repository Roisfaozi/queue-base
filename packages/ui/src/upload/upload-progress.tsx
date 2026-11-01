import type { UploadItem } from "@casbin/hooks";
import { Progress } from "../progress";
import { NexusButton } from "../nexus-button";
import { cn } from "../lib/utils";
import {
	File,
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
	paused: { icon: Pause, label: "Paused", color: "text-warning" },
	success: { icon: CheckCircle2, label: "Completed", color: "text-success" },
	error: { icon: AlertCircle, label: "Failed", color: "text-danger" },
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
				"group bg-card text-card-foreground border-border hover:border-primary/40 relative flex items-center gap-3 rounded-[var(--radius-md,8px)] border p-3 shadow-xs transition-colors",
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
						className="text-foreground truncate text-sm font-medium"
						title={item.relativePath || item.fileName}
					>
						{item.relativePath || item.fileName}
					</p>
					<span className="text-muted-foreground shrink-0 text-xs">
						{formatBytes(item.fileSize)}
					</span>
				</div>

				{item.targetFolderName && (
					<p className="text-muted-foreground/80 truncate text-[11px]">
						to: {item.targetFolderName}
					</p>
				)}

				{(item.status === "preparing" ||
					item.status === "uploading" ||
					item.status === "paused") && (
					<div className="space-y-1">
						<Progress value={item.progress} className="h-1.5" />
						<div className="text-muted-foreground flex items-center justify-between text-[11px]">
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
					<p className="text-danger truncate text-xs">
						{item.errorMessage || item.error}
						{(item.retryCount ?? 0) > 0 && ` (attempt ${item.retryCount! + 1})`}
					</p>
				)}
			</div>

			<div className="flex shrink-0 items-center gap-1 opacity-0 transition-opacity group-hover:opacity-100">
				{item.status === "uploading" && onPause && (
					<NexusButton
						variant="ghost"
						size="icon"
						onClick={onPause}
						className="h-7 w-7"
						aria-label={`Pause upload of ${item.fileName}`}
					>
						<Pause className="h-3.5 w-3.5" />
					</NexusButton>
				)}

				{item.status === "paused" && onResume && (
					<NexusButton
						variant="ghost"
						size="icon"
						onClick={onResume}
						className="h-7 w-7"
						aria-label={`Resume upload of ${item.fileName}`}
					>
						<Play className="h-3.5 w-3.5" />
					</NexusButton>
				)}

				{item.status === "error" && onRetry && (
					<NexusButton
						variant="ghost"
						size="icon"
						onClick={onRetry}
						className="h-7 w-7"
						aria-label={`Retry upload of ${item.fileName}`}
					>
						<RotateCw className="h-3.5 w-3.5" />
					</NexusButton>
				)}

				{item.status !== "success" &&
					item.status !== "canceled" &&
					onCancel && (
						<NexusButton
							variant="ghost"
							size="icon"
							onClick={onCancel}
							className="text-muted-foreground hover:text-danger h-7 w-7"
							aria-label={`Cancel upload of ${item.fileName}`}
						>
							<X className="h-3.5 w-3.5" />
						</NexusButton>
					)}
			</div>
		</div>
	);
}
