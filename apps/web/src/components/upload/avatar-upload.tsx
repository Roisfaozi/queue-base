"use client";

import { useState, useRef, useCallback } from "react";
import { Avatar, AvatarFallback, AvatarImage } from "~/components/ui/avatar";
import { Button } from "~/components/ui/button";
import { Progress } from "~/components/ui/progress";
import { Camera, Upload, X, Loader2, User } from "lucide-react";
import { toast } from "sonner";
import { isAcceptedFile, uploadSingleFile } from "~/lib/upload";
import { cn } from "~/lib/utils";

export interface AvatarUploadProps {
	/** URL of the current avatar */
	value?: string | null;
	/** User's full name to generate fallback initials */
	name?: string;
	/** Callback invoked with the new URL when upload finishes, or empty string on remove */
	onChange?: (url: string) => void;
	/** Callback invoked when a valid file is selected */
	onFileSelect?: (file: File) => void;
	/** Optional custom upload handler; defaults to direct upload (<15MB) or TUS (>=15MB) */
	onUpload?: (file: File) => Promise<string>;
	/** Display size */
	size?: "sm" | "md" | "lg" | "xl";
	/** Avatar shape */
	shape?: "circle" | "square";
	/** Accepted image MIME types or extensions */
	accept?: string;
	/** Maximum file size in bytes (default 50MB; <15MB uses direct upload, >=15MB uses TUS) */
	maxSize?: number;
	/** Whether the upload control is disabled */
	disabled?: boolean;
	/** Additional CSS class for the wrapper */
	className?: string;
	/** Whether to display a remove button when an avatar exists */
	showRemove?: boolean;
	/** Backend upload type metadata (default: "avatar") */
	uploadType?: string;
}

const sizeClasses = {
	sm: {
		container: "h-16 w-16",
		text: "text-xs",
		icon: "h-4 w-4",
		btn: "h-6 w-6",
	},
	md: {
		container: "h-24 w-24",
		text: "text-base",
		icon: "h-5 w-5",
		btn: "h-7 w-7",
	},
	lg: {
		container: "h-32 w-32",
		text: "text-xl",
		icon: "h-6 w-6",
		btn: "h-8 w-8",
	},
	xl: {
		container: "h-40 w-40",
		text: "text-2xl",
		icon: "h-8 w-8",
		btn: "h-9 w-9",
	},
};

function getInitials(name?: string): string {
	if (!name) return "";
	const parts = name.trim().split(/\s+/);
	if (parts.length === 1) return parts[0].slice(0, 2).toUpperCase();
	return (parts[0][0] + parts[parts.length - 1][0]).toUpperCase();
}

export function AvatarUpload({
	value,
	name,
	onChange,
	onFileSelect,
	onUpload,
	size = "md",
	shape = "circle",
	accept = "image/jpeg,image/png,image/webp,image/gif",
	maxSize = 50 * 1024 * 1024, // 50MB (<15MB direct, >=15MB TUS)
	disabled = false,
	className,
	showRemove = true,
	uploadType = "avatar",
}: AvatarUploadProps) {
	const [previewUrl, setPreviewUrl] = useState<string | null>(null);
	const [isUploading, setIsUploading] = useState(false);
	const [progress, setProgress] = useState(0);
	const [isDragOver, setIsDragOver] = useState(false);
	const inputRef = useRef<HTMLInputElement>(null);

	const activeImage = previewUrl ?? value ?? "";
	const sizeConfig = sizeClasses[size];
	const initials = getInitials(name);

	const handleFile = useCallback(
		async (file: File) => {
			if (!isAcceptedFile(file, accept)) {
				toast.error(
					`Format berkas "${file.name}" tidak didukung. Harap pilih gambar (${accept}).`,
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

			// Show immediate client-side preview
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
						uploadType,
						onProgress: (p) => setProgress(p),
					});
				}

				toast.success("Foto profil berhasil diunggah.");
				onChange?.(resultUrl);
			} catch (error: any) {
				// Revert local preview on error
				setPreviewUrl(null);
				toast.error(error.message || "Gagal mengunggah foto profil.");
			} finally {
				setIsUploading(false);
				setProgress(0);
				if (inputRef.current) inputRef.current.value = "";
			}
		},
		[accept, maxSize, onFileSelect, onUpload, uploadType, onChange],
	);

	const handleRemove = useCallback(
		(e: React.MouseEvent) => {
			e.stopPropagation();
			setPreviewUrl(null);
			if (inputRef.current) inputRef.current.value = "";
			onChange?.("");
			toast.success("Foto profil berhasil dihapus.");
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
		<div className={cn("inline-flex flex-col items-center gap-3", className)}>
			<div
				onDragOver={onDragOver}
				onDragLeave={onDragLeave}
				onDrop={onDrop}
				onClick={() => {
					if (!disabled && !isUploading) inputRef.current?.click();
				}}
				className={cn(
					"group relative cursor-pointer select-none transition-all duration-200 focus-visible:outline-hidden focus-visible:ring-2 focus-visible:ring-primary focus-visible:ring-offset-2",
					shape === "circle" ? "rounded-full" : "rounded-2xl",
					disabled && "cursor-not-allowed opacity-60",
					isDragOver && "ring-primary ring-4 ring-offset-2 scale-105",
				)}
				tabIndex={disabled ? -1 : 0}
				role="button"
				aria-label="Upload avatar"
			>
				<Avatar
					className={cn(
						sizeConfig.container,
						shape === "circle" ? "rounded-full" : "rounded-2xl",
						"border-border bg-muted/40 shadow-xs border-2 transition-transform",
					)}
				>
					<AvatarImage
						src={activeImage}
						alt={name || "Avatar"}
						className="h-full w-full object-cover"
					/>
					<AvatarFallback
						className={cn(
							sizeConfig.text,
							shape === "circle" ? "rounded-full" : "rounded-2xl",
							"bg-primary/10 text-primary font-semibold uppercase",
						)}
					>
						{initials || <User className={sizeConfig.icon} />}
					</AvatarFallback>
				</Avatar>

				{/* Hover Overlay */}
				{!disabled && !isUploading && (
					<div
						className={cn(
							"absolute inset-0 flex flex-col items-center justify-center bg-black/50 text-white opacity-0 transition-opacity duration-200 group-hover:opacity-100",
							shape === "circle" ? "rounded-full" : "rounded-2xl",
						)}
					>
						<Camera className={cn(sizeConfig.icon, "mb-1")} />
						<span className="text-[10px] font-medium tracking-tight">Ubah</span>
					</div>
				)}

				{/* Uploading Spinner & Progress Overlay */}
				{isUploading && (
					<div
						className={cn(
							"absolute inset-0 flex flex-col items-center justify-center bg-black/65 text-white backdrop-blur-xs",
							shape === "circle" ? "rounded-full" : "rounded-2xl",
						)}
					>
						<Loader2
							className={cn(sizeConfig.icon, "animate-spin text-primary")}
						/>
						{progress > 0 && (
							<span className="mt-1 text-[11px] font-semibold">
								{progress}%
							</span>
						)}
					</div>
				)}

				{/* Remove Button Badge */}
				{showRemove && Boolean(activeImage) && !isUploading && !disabled && (
					<Button
						type="button"
						variant="destructive"
						size="icon"
						onClick={handleRemove}
						className={cn(
							"absolute -top-1 -right-1 z-10 shadow-md",
							sizeConfig.btn,
							"rounded-full p-0",
						)}
						title="Hapus foto profil"
						aria-label="Hapus foto profil"
					>
						<X className="h-3 w-3" />
					</Button>
				)}
			</div>

			{/* Progress Bar (Visible during upload) */}
			{isUploading && (
				<div className="w-full max-w-[120px] space-y-1 text-center">
					<Progress value={progress} className="h-1.5" />
				</div>
			)}

			{/* Action Links/Buttons below */}
			{!isUploading && !disabled && (
				<div className="flex items-center gap-2">
					<Button
						type="button"
						variant="outline"
						size="sm"
						onClick={() => inputRef.current?.click()}
						className="h-8 gap-1.5 text-xs"
					>
						<Upload className="h-3.5 w-3.5" />
						Unggah Foto
					</Button>

					{showRemove && Boolean(activeImage) && (
						<Button
							type="button"
							variant="ghost"
							size="sm"
							onClick={handleRemove}
							className="text-muted-foreground hover:text-destructive h-8 px-2 text-xs"
						>
							Hapus
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
