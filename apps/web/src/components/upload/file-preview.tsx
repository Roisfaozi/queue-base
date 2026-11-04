"use client";

import { cn } from "~/lib/utils";
import {
	FileText,
	FileImage,
	FileVideo,
	FileAudio,
	FileArchive,
	File,
	ExternalLink,
} from "lucide-react";

export interface FilePreviewProps {
	file: File;
	url?: string;
	className?: string;
}

function getFileIcon(type: string) {
	if (type.startsWith("image/")) return FileImage;
	if (type.startsWith("video/")) return FileVideo;
	if (type.startsWith("audio/")) return FileAudio;
	if (
		type.includes("zip") ||
		type.includes("rar") ||
		type.includes("tar") ||
		type.includes("compressed")
	)
		return FileArchive;
	if (
		type.includes("pdf") ||
		type.includes("document") ||
		type.includes("text")
	)
		return FileText;
	return File;
}

function formatBytes(bytes: number) {
	if (bytes === 0) return "0 B";
	const k = 1024;
	const sizes = ["B", "KB", "MB", "GB"];
	const i = Math.floor(Math.log(bytes) / Math.log(k));
	return `${parseFloat((bytes / k ** i).toFixed(1))} ${sizes[i]}`;
}

export function FilePreview({ file, url, className }: FilePreviewProps) {
	const Icon = getFileIcon(file.type);
	const isImage = file.type.startsWith("image/");
	const objectUrl = isImage ? URL.createObjectURL(file) : null;

	return (
		<div
			className={cn(
				"group relative flex flex-col overflow-hidden rounded-[var(--radius-lg)] border border-border bg-card text-card-foreground shadow-xs transition-all hover:border-primary/40",
				className,
			)}
		>
			<div className="relative flex aspect-video w-full items-center justify-center overflow-hidden bg-muted/40">
				{isImage && objectUrl ? (
					<img
						src={objectUrl}
						alt={file.name}
						className="h-full w-full object-cover transition-transform duration-200 group-hover:scale-105"
						onLoad={() => {
							if (objectUrl) URL.revokeObjectURL(objectUrl);
						}}
					/>
				) : (
					<Icon className="h-10 w-10 text-muted-foreground transition-transform group-hover:scale-110" />
				)}
			</div>

			<div className="space-y-1 p-3">
				<p
					className="truncate text-sm font-medium text-foreground"
					title={file.name}
				>
					{file.name}
				</p>
				<div className="flex items-center justify-between text-xs text-muted-foreground">
					<span>{formatBytes(file.size)}</span>
					<span className="uppercase">
						{file.name.split(".").pop() || "file"}
					</span>
				</div>

				{url && (
					<a
						href={url}
						target="_blank"
						rel="noopener noreferrer"
						className="inline-flex items-center gap-1 pt-1 text-xs font-medium text-primary hover:underline"
					>
						View uploaded file
						<ExternalLink className="h-3 w-3" />
					</a>
				)}
			</div>
		</div>
	);
}
