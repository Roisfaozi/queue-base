"use client";

import { useState, useRef, useCallback } from "react";
import { Button } from "~/components/ui/button";
import { Progress } from "~/components/ui/progress";
import {
	Building2,
	Upload,
	Trash2,
	RotateCw,
	Loader2,
	Image as ImageIcon,
} from "lucide-react";
import { toast } from "sonner";
import { isAcceptedFile, uploadSingleFile } from "~/lib/upload";
import { cn } from "~/lib/utils";

export interface LogoUploadProps {
	/** Current logo URL or Asset ID */
	value?: string | null;
	/** Organization or branch name (used for alt text and fallback) */
	name?: string;
	/** Callback invoked with the new URL / Asset ID when upload succeeds, or empty string on remove */
	onChange?: (url: string) => void;
	/** Callback invoked when a valid file is selected */
	onFileSelect?: (file: File) => void;
	/** Optional custom upload handler; defaults to direct upload (<15MB) or TUS (>=15MB) */
	onUpload?: (file: File) => Promise<string>;
	/** Target entity scope for direct logo upload (<15MB) */
	scope?: "organization" | "branch";
	/** Aspect ratio / layout mode */
	aspectRatio?: "square" | "wide";
	/** Accepted image MIME types or extensions */
	accept?: string;
	/** Maximum file size in bytes (default 50MB; <15MB uses direct upload, >=15MB uses TUS) */
	maxSize?: number;
	/** Title displayed when empty */
	title?: string;
	/** Description displayed when empty */
	description?: string;
	/** Whether the upload control is disabled */
	disabled?: boolean;
	/** Additional CSS class for the container */
	className?: string;
	/** Backend upload type metadata (overrides scope when provided) */
	uploadType?: string;
}

export function LogoUpload({
	value,
	name,
	onChange,
	onFileSelect,
	onUpload,
	scope = "organization",
	aspectRatio = "wide",
	accept = "image/png,image/jpeg,image/webp,image/gif",
	maxSize = 50 * 1024 * 1024, // 50MB (<15MB direct, >=15MB TUS)
	title = "Unggah Logo",
	description = "Format PNG, JPG, WEBP, atau GIF (<15MB upload biasa, ≥15MB via TUS)",
	disabled = false,
	className,
	uploadType,
}: LogoUploadProps) {
	const resolvedUploadType =
		uploadType ?? (scope === "branch" ? "branch_logo" : "organization_logo");
	const [previewUrl, setPreviewUrl] = useState<string | null>(null);
	const [isUploading, setIsUploading] = useState(false);
	const [progress, setProgress] = useState(0);
	const [isDragOver, setIsDragOver] = useState(false);
	const inputRef = useRef<HTMLInputElement>(null);

	const activeImage = previewUrl ?? value ?? "";

	const handleFile = useCallback(
		async (file: File) => {
			if (!isAcceptedFile(file, accept)) {
				toast.error(
					`Format berkas "${file.name}" tidak didukung. Harap pilih logo gambar (${accept}).`,
				);
				return;
			}

			if (file.size > maxSize) {
				toast.error(
					`Ukuran berkas (${(file.size / 1024 / 1024).toFixed(1)}MB) melebihi batas maksimal ${(maxSize / 1024 / 1024).toFixed(0)}MB.`,
				);
				return;
			}

			onFileSelect?.(file);

			// Immediate local preview
			const localUrl = URL.createObjectURL(file);
			setPreviewUrl(localUrl);
			setIsUploading(true);
			setProgress(0);

			try {
				let resultUrl: string;
				if (onUpload) {
					resultUrl = await onUpload(file);
				} else {
					resultUrl = await uploadSingleFile(file, {
						uploadType: resolvedUploadType,
						onProgress: (p) => setProgress(p),
					});
				}

				toast.success("Logo berhasil diunggah.");
				onChange?.(resultUrl);
			} catch (error: any) {
				// Revert local preview on error
				setPreviewUrl(null);
				toast.error(error.message || "Gagal mengunggah logo.");
			} finally {
				setIsUploading(false);
				setProgress(0);
				if (inputRef.current) inputRef.current.value = "";
			}
		},
		[accept, maxSize, onFileSelect, onUpload, resolvedUploadType, onChange],
	);

	const handleRemove = useCallback(
		(e: React.MouseEvent) => {
			e.stopPropagation();
			setPreviewUrl(null);
			if (inputRef.current) inputRef.current.value = "";
			onChange?.("");
			toast.success("Logo berhasil dihapus.");
		},
		[onChange],
	);

	const onDragOver = (e: React.DragEvent) => {
		e.preventDefault();
		if (!disabled && !isUploading) setIsDragOver(true);
	};

	const onDragLeave = () => {
		setIsDragOver(false);
	};

	const onDrop = (e: React.DragEvent) => {
		e.preventDefault();
		setIsDragOver(false);
		if (disabled || isUploading) return;
		const file = e.dataTransfer.files?.[0];
		if (file) handleFile(file);
	};

	return (
		<div className={cn("space-y-3", className)}>
			{/* Preview & Dropzone Container */}
			<div
				onDragOver={onDragOver}
				onDragLeave={onDragLeave}
				onDrop={onDrop}
				onClick={() => {
					if (!disabled && !isUploading) inputRef.current?.click();
				}}
				className={cn(
					"group relative flex cursor-pointer flex-col items-center justify-center overflow-hidden rounded-[var(--radius-lg)] border-2 transition-all duration-200 select-none",
					aspectRatio === "wide"
						? "aspect-2/1 min-h-[160px] w-full"
						: "aspect-square w-48",
					activeImage
						? "border-border bg-card hover:border-primary/50"
						: "border-border border-dashed hover:border-primary/50 hover:bg-muted/30",
					isDragOver && "border-primary bg-primary/5 scale-[1.01]",
					disabled && "cursor-not-allowed opacity-60",
				)}
				tabIndex={disabled ? -1 : 0}
				role="button"
				aria-label={activeImage ? "Change logo" : "Upload logo"}
			>
				{/* Empty State */}
				{!activeImage && (
					<div className="flex flex-col items-center justify-center p-6 text-center">
						<div
							className={cn(
								"mb-3 flex h-12 w-12 items-center justify-center rounded-full transition-colors",
								isDragOver
									? "bg-primary/10 text-primary"
									: "bg-muted text-muted-foreground",
							)}
						>
							<Building2 className="h-6 w-6" />
						</div>
						<p className="text-foreground text-sm font-semibold">{title}</p>
						<p className="text-muted-foreground mt-1 text-xs">{description}</p>
					</div>
				)}

				{/* Active Logo Image View */}
				{activeImage && (
					<div className="relative flex h-full w-full items-center justify-center p-4">
						<img
							src={activeImage}
							alt={name || "Logo"}
							className="max-h-full max-w-full object-contain transition-transform duration-200 group-hover:scale-105"
						/>

						{/* Hover Action Overlay */}
						{!disabled && !isUploading && (
							<div className="absolute inset-0 flex items-center justify-center gap-2 bg-black/50 opacity-0 backdrop-blur-2xs transition-opacity duration-200 group-hover:opacity-100">
								<Button
									type="button"
									variant="secondary"
									size="sm"
									className="gap-1.5 text-xs shadow-md"
									onClick={(e) => {
										e.stopPropagation();
										inputRef.current?.click();
									}}
								>
									<RotateCw className="h-3.5 w-3.5" />
									Ganti
								</Button>
								<Button
									type="button"
									variant="destructive"
									size="sm"
									className="gap-1.5 text-xs shadow-md"
									onClick={handleRemove}
								>
									<Trash2 className="h-3.5 w-3.5" />
									Hapus
								</Button>
							</div>
						)}
					</div>
				)}

				{/* Uploading Overlay */}
				{isUploading && (
					<div className="absolute inset-0 flex flex-col items-center justify-center bg-black/60 p-4 text-white backdrop-blur-xs">
						<Loader2 className="text-primary h-7 w-7 animate-spin" />
						<p className="mt-2 text-xs font-medium">Mengunggah logo...</p>
						<div className="mt-3 w-40">
							<Progress value={progress} className="h-1.5" />
						</div>
						{progress > 0 && (
							<span className="mt-1 text-[11px] text-muted-foreground">
								{progress}%
							</span>
						)}
					</div>
				)}
			</div>

			{/* Auxiliary helper controls under the box */}
			{!isUploading && !disabled && (
				<div className="flex items-center justify-between text-xs">
					<span className="text-muted-foreground">
						{activeImage
							? "Klik atau seret logo baru untuk mengganti"
							: "Mendukung drag & drop berkas"}
					</span>
					{activeImage && (
						<Button
							type="button"
							variant="ghost"
							size="sm"
							onClick={handleRemove}
							className="text-muted-foreground hover:text-destructive h-7 px-2 text-xs"
						>
							Hapus Logo
						</Button>
					)}
				</div>
			)}

			{/* Hidden file input */}
			<input
				ref={inputRef}
				type="file"
				accept={accept}
				onChange={(e) => {
					const file = e.target.files?.[0];
					if (file) handleFile(file);
				}}
				className="hidden"
				disabled={disabled || isUploading}
			/>
		</div>
	);
}
