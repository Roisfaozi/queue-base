"use client";

import {
	ExternalLink,
	File as FileIcon,
	FileAudio,
	Loader2,
	RotateCw,
	Upload,
	X,
} from "lucide-react";
import { useCallback, useRef, useState } from "react";
import { toast } from "sonner";
import { Button } from "~/components/ui/button";
import { isAcceptedFile } from "~/lib/upload";
import { cn } from "~/lib/utils";

export interface FileUploadProps {
	accept?: string;
	multiple?: boolean;
	maxSize?: number; // bytes, default 10MB
	value?: string; // existing file URL for single-file controlled mode
	onValueChange?: (url: string) => void;
	onUpload?: (file: File) => Promise<string | void>;
	onChange?: (files: File[]) => void;
	disabled?: boolean;
	compact?: boolean;
	formatHint?: string;
	className?: string;
}

export function FileUpload({
	accept,
	multiple = false,
	maxSize = 10 * 1024 * 1024,
	value,
	onValueChange,
	onUpload,
	onChange,
	disabled = false,
	compact = false,
	formatHint,
	className,
}: FileUploadProps) {
	const [files, setFiles] = useState<File[]>([]);
	const [dragOver, setDragOver] = useState(false);
	const [isUploading, setIsUploading] = useState(false);
	const inputRef = useRef<HTMLInputElement>(null);

	const isAudio = Boolean(
		accept?.includes("audio") ||
			(value && /\.(mp3|wav|ogg|m4a|aac)($|\?)/i.test(value)),
	);

	const handleFiles = useCallback(
		async (incoming: FileList | null) => {
			if (!incoming || disabled || isUploading) return;
			const all = Array.from(incoming);

			// Validate file format (especially critical for drag & drop)
			const rejectedFormat = all.filter((f) => !isAcceptedFile(f, accept));
			if (rejectedFormat.length > 0) {
				toast.error(
					`Format berkas "${rejectedFormat[0].name}" tidak didukung (hanya menerima: ${accept}).`,
				);
			}

			// Validate file size
			const oversized = all.filter((f) => f.size > maxSize);
			if (oversized.length > 0) {
				toast.error(
					`File "${oversized[0].name}" melebihi batas ukuran maksimal (${(maxSize / 1024 / 1024).toFixed(0)}MB).`,
				);
			}

			const valid = all
				.filter((f) => isAcceptedFile(f, accept))
				.filter((f) => f.size <= maxSize);

			if (valid.length === 0) {
				if (inputRef.current) inputRef.current.value = "";
				return;
			}

			const next = multiple ? [...files, ...valid] : valid.slice(0, 1);
			setFiles(next);
			onChange?.(next);

			// If onUpload handler is provided, perform upload for the selected file(s)
			if (onUpload && valid.length > 0) {
				setIsUploading(true);
				try {
					const targetFile = valid[0];
					const uploadedUrl = await onUpload(targetFile);
					if (uploadedUrl && typeof uploadedUrl === "string") {
						onValueChange?.(uploadedUrl);
					}
					toast.success(`Berkas "${targetFile.name}" berhasil diunggah.`);
				} catch (error: any) {
					toast.error(error.message || "Gagal mengunggah berkas.");
				} finally {
					setIsUploading(false);
					if (inputRef.current) inputRef.current.value = "";
				}
			} else {
				if (inputRef.current) inputRef.current.value = "";
			}
		},
		[
			accept,
			disabled,
			files,
			isUploading,
			maxSize,
			multiple,
			onChange,
			onUpload,
			onValueChange,
		],
	);

	const removeFile = (index: number) => {
		const next = files.filter((_, i) => i !== index);
		setFiles(next);
		onChange?.(next);
	};

	const handleClearValue = () => {
		setFiles([]);
		onChange?.([]);
		onValueChange?.("");
		if (inputRef.current) inputRef.current.value = "";
	};

	// 1. Controlled URL preview mode (e.g. existing audio or uploaded URL)
	if (value && !isUploading) {
		return (
			<div
				className={cn(
					compact
						? "space-y-1.5 rounded-[var(--radius-md)] border border-border bg-muted/40 p-2.5"
						: "space-y-2 rounded-[var(--radius-lg)] border border-border bg-muted/40 p-3",
					className,
				)}
			>
				<input
					ref={inputRef}
					type="file"
					accept={accept}
					multiple={multiple}
					disabled={disabled || isUploading}
					onChange={(e) => handleFiles(e.target.files)}
					className="hidden"
				/>

				<div className="flex items-center justify-between gap-2">
					<div className="flex items-center gap-2 min-w-0">
						{isAudio ? (
							<FileAudio
								className={cn(
									"shrink-0 text-primary",
									compact ? "h-3.5 w-3.5" : "h-4 w-4",
								)}
							/>
						) : (
							<FileIcon
								className={cn(
									"shrink-0 text-primary",
									compact ? "h-3.5 w-3.5" : "h-4 w-4",
								)}
							/>
						)}
						<span
							className={cn(
								"truncate font-mono text-muted-foreground",
								compact ? "text-[11px]" : "text-xs",
							)}
						>
							{value}
						</span>
					</div>
					<div className="flex items-center gap-1 shrink-0">
						<Button
							type="button"
							variant="ghost"
							size="icon"
							className={cn(
								"text-muted-foreground hover:text-foreground",
								compact ? "h-6 w-6" : "h-7 w-7",
							)}
							asChild
						>
							<a
								href={value}
								target="_blank"
								rel="noopener noreferrer"
								title="Buka berkas di tab baru"
							>
								<ExternalLink className={compact ? "h-3 w-3" : "h-3.5 w-3.5"} />
							</a>
						</Button>
						<Button
							type="button"
							variant="ghost"
							size="icon"
							className={cn(
								"text-destructive hover:bg-destructive/10",
								compact ? "h-6 w-6" : "h-7 w-7",
							)}
							onClick={handleClearValue}
							disabled={disabled || isUploading}
							title="Hapus berkas"
						>
							<X className={compact ? "h-3 w-3" : "h-3.5 w-3.5"} />
						</Button>
					</div>
				</div>

				{isAudio && (
					<audio
						controls
						src={value}
						className={cn("w-full rounded", compact ? "h-7" : "h-8")}
						preload="metadata"
					>
						Browser Anda tidak mendukung elemen audio.
					</audio>
				)}

				<div className="flex justify-end pt-0.5">
					<Button
						type="button"
						variant="outline"
						size="sm"
						className={compact ? "h-6 px-2 text-[11px]" : "h-7 text-xs"}
						onClick={() => inputRef.current?.click()}
						disabled={disabled || isUploading}
					>
						<RotateCw className="mr-1 h-3 w-3" />
						Ganti Berkas
					</Button>
				</div>
			</div>
		);
	}

	return (
		<div className={cn(compact ? "space-y-1.5" : "space-y-3", className)}>
			<div
				onDragOver={(e) => {
					e.preventDefault();
					if (!disabled && !isUploading) setDragOver(true);
				}}
				onDragLeave={() => setDragOver(false)}
				onDrop={(e) => {
					e.preventDefault();
					setDragOver(false);
					handleFiles(e.dataTransfer.files);
				}}
				onClick={() => {
					if (!disabled && !isUploading) inputRef.current?.click();
				}}
				className={cn(
					"flex flex-col items-center justify-center border-2 border-dashed transition-colors duration-200",
					compact
						? "rounded-[var(--radius-md)] p-2.5 gap-1 min-h-[70px]"
						: "rounded-[var(--radius-lg)] p-6 gap-2",
					disabled || isUploading
						? "cursor-not-allowed opacity-60 bg-muted/20 border-border"
						: "cursor-pointer",
					dragOver
						? "border-primary bg-primary/5"
						: "border-border hover:border-primary/50 hover:bg-muted/30",
				)}
			>
				{isUploading ? (
					<>
						<Loader2
							className={cn(
								"text-primary animate-spin",
								compact ? "h-5 w-5" : "h-8 w-8",
							)}
						/>
						<p
							className={cn(
								"text-muted-foreground font-medium",
								compact ? "text-xs" : "text-sm",
							)}
						>
							Mengunggah berkas ke storage...
						</p>
					</>
				) : (
					<>
						<Upload
							className={cn(
								"text-muted-foreground",
								compact ? "h-5 w-5" : "h-8 w-8",
							)}
						/>
						<p
							className={cn(
								"text-muted-foreground",
								compact ? "text-xs" : "text-sm",
							)}
						>
							Drag & drop or{" "}
							<span className="text-primary font-medium">browse</span>
						</p>
						<p
							className={cn(
								"text-muted-foreground text-center",
								compact ? "text-[11px] leading-tight" : "text-xs",
							)}
						>
							{formatHint ? (
								formatHint
							) : (
								<>
									{accept ? `Hanya format: ${accept} · ` : ""}Max{" "}
									{(maxSize / 1024 / 1024).toFixed(0)}MB per file
								</>
							)}
						</p>
					</>
				)}
				<input
					ref={inputRef}
					type="file"
					accept={accept}
					multiple={multiple}
					disabled={disabled || isUploading}
					onChange={(e) => handleFiles(e.target.files)}
					className="hidden"
				/>
			</div>

			{files.length > 0 && !value && (
				<ul className="space-y-2">
					{files.map((file, i) => (
						<li
							key={`${file.name}-${i}`}
							className="border-border bg-card text-card-foreground flex items-center gap-3 rounded-[var(--radius-md)] border p-3 shadow-xs"
						>
							{file.type.startsWith("audio/") ? (
								<FileAudio className="text-muted-foreground h-4 w-4 shrink-0" />
							) : (
								<FileIcon className="text-muted-foreground h-4 w-4 shrink-0" />
							)}
							<span className="text-foreground text-sm flex-1 truncate font-medium">
								{file.name}
							</span>
							<span className="text-muted-foreground text-xs">
								{(file.size / 1024).toFixed(1)} KB
							</span>
							<Button
								type="button"
								variant="ghost"
								size="icon"
								onClick={(e) => {
									e.stopPropagation();
									removeFile(i);
								}}
								className="h-7 w-7 text-muted-foreground hover:text-destructive"
								aria-label={`Remove ${file.name}`}
							>
								<X className="h-4 w-4" />
							</Button>
						</li>
					))}
				</ul>
			)}
		</div>
	);
}
