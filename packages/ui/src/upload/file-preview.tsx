import { cn } from "../lib/utils";
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
				"group bg-card text-card-foreground border-border hover:border-primary/40 relative flex flex-col overflow-hidden rounded-[var(--radius-lg,12px)] border shadow-xs transition-all",
				className,
			)}
		>
			<div className="bg-muted/40 relative flex aspect-video w-full items-center justify-center overflow-hidden">
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
					<Icon className="text-muted-foreground h-10 w-10 transition-transform group-hover:scale-110" />
				)}
			</div>

			<div className="space-y-1 p-3">
				<p
					className="text-foreground truncate text-sm font-medium"
					title={file.name}
				>
					{file.name}
				</p>
				<div className="text-muted-foreground flex items-center justify-between text-xs">
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
						className="text-primary hover:text-primary-hover hover:underline inline-flex items-center gap-1 pt-1 text-xs font-medium"
					>
						View uploaded file
						<ExternalLink className="h-3 w-3" />
					</a>
				)}
			</div>
		</div>
	);
}
