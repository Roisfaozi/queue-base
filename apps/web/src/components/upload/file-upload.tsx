"use client";

import { useCallback, useState, useRef } from "react";
import { cn } from "~/lib/utils";
import { Upload, X, File as FileIcon } from "lucide-react";
import { Button } from "~/components/ui/button";
import { toast } from "sonner";
import { isAcceptedFile } from "~/lib/upload";

export interface FileUploadProps {
	accept?: string;
	multiple?: boolean;
	maxSize?: number; // bytes, default 10MB
	onChange?: (files: File[]) => void;
	className?: string;
}

export function FileUpload({
	accept,
	multiple = false,
	maxSize = 10 * 1024 * 1024,
	onChange,
	className,
}: FileUploadProps) {
	const [files, setFiles] = useState<File[]>([]);
	const [dragOver, setDragOver] = useState(false);
	const inputRef = useRef<HTMLInputElement>(null);

	const handleFiles = useCallback(
		(incoming: FileList | null) => {
			if (!incoming) return;
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

			const next = multiple ? [...files, ...valid] : valid.slice(0, 1);
			setFiles(next);
			onChange?.(next);
		},
		[accept, files, maxSize, multiple, onChange],
	);

	const removeFile = (index: number) => {
		const next = files.filter((_, i) => i !== index);
		setFiles(next);
		onChange?.(next);
	};

	return (
		<div className={cn("space-y-3", className)}>
			<div
				onDragOver={(e) => {
					e.preventDefault();
					setDragOver(true);
				}}
				onDragLeave={() => setDragOver(false)}
				onDrop={(e) => {
					e.preventDefault();
					setDragOver(false);
					handleFiles(e.dataTransfer.files);
				}}
				onClick={() => inputRef.current?.click()}
				className={cn(
					"flex cursor-pointer flex-col items-center justify-center gap-2 rounded-[var(--radius-lg)] border-2 border-dashed p-8 transition-colors duration-200",
					dragOver
						? "border-primary bg-primary/5"
						: "border-border hover:border-primary/50 hover:bg-muted/30",
				)}
			>
				<Upload className="text-muted-foreground h-8 w-8" />
				<p className="text-muted-foreground text-sm">
					Drag & drop or{" "}
					<span className="text-primary font-medium">browse</span>
				</p>
				<p className="text-muted-foreground text-xs">
					{accept ? `Hanya format: ${accept} · ` : ""}Max{" "}
					{(maxSize / 1024 / 1024).toFixed(0)}MB per file
				</p>
				<input
					ref={inputRef}
					type="file"
					accept={accept}
					multiple={multiple}
					onChange={(e) => handleFiles(e.target.files)}
					className="hidden"
				/>
			</div>

			{files.length > 0 && (
				<ul className="space-y-2">
					{files.map((file, i) => (
						<li
							key={`${file.name}-${i}`}
							className="border-border bg-card text-card-foreground flex items-center gap-3 rounded-[var(--radius-md)] border p-3 shadow-xs"
						>
							<FileIcon className="text-muted-foreground h-4 w-4 shrink-0" />
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
