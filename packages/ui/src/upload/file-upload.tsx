import { useState, useRef } from "react";
import { cn } from "../lib/utils";
import { Upload, X, File } from "lucide-react";
import { NexusButton } from "../nexus-button";

export interface FileUploadProps {
	accept?: string;
	multiple?: boolean;
	maxSize?: number; // in bytes, default 10MB
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

	const handleFiles = (incoming: FileList | null) => {
		if (!incoming) return;
		const valid = Array.from(incoming).filter((f) => f.size <= maxSize);
		const next = multiple ? [...files, ...valid] : valid.slice(0, 1);
		setFiles(next);
		onChange?.(next);
	};

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
					"border-border hover:border-primary/50 hover:bg-muted/30 flex cursor-pointer flex-col items-center justify-center gap-2 rounded-[var(--radius-lg,12px)] border-2 border-dashed p-6 transition-colors",
					dragOver && "border-primary bg-primary/5",
				)}
			>
				<Upload className="text-muted-foreground h-8 w-8" />
				<div className="text-center">
					<p className="text-foreground text-sm font-medium">
						Drag & drop or{" "}
						<span className="text-primary underline">browse</span>
					</p>
					<p className="text-muted-foreground text-xs">
						Max {Math.round(maxSize / (1024 * 1024))}MB
					</p>
				</div>
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
				<ul className="space-y-1.5">
					{files.map((file, i) => (
						<li
							key={`${file.name}-${i}`}
							className="bg-muted/40 border-border flex items-center justify-between rounded-md border px-3 py-1.5 text-xs"
						>
							<span className="flex items-center gap-2 truncate">
								<File className="text-muted-foreground h-3.5 w-3.5 shrink-0" />
								<span className="truncate">{file.name}</span>
								<span className="text-muted-foreground shrink-0">
									({(file.size / 1024).toFixed(1)} KB)
								</span>
							</span>
							<NexusButton
								type="button"
								variant="ghost"
								size="icon"
								onClick={() => removeFile(i)}
								className="text-muted-foreground hover:text-danger h-6 w-6"
								aria-label={`Remove ${file.name}`}
							>
								<X className="h-3 w-3" />
							</NexusButton>
						</li>
					))}
				</ul>
			)}
		</div>
	);
}
