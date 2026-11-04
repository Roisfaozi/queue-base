"use client";

import {
	ExternalLink,
	Link as LinkIcon,
	Loader2,
	Music,
	Trash2,
	UploadCloud,
} from "lucide-react";
import { type ChangeEvent, useRef, useState } from "react";
import { toast } from "sonner";
import { Button } from "~/components/ui/button";
import {
	FormControl,
	FormDescription,
	FormItem,
	FormLabel,
	FormMessage,
} from "~/components/ui/form";
import { Input } from "~/components/ui/input";
import { servicesApi } from "~/lib/api/qms";

interface AudioUploadFieldProps {
	label: string;
	description?: string;
	value?: string;
	onChange: (val: string) => void;
	disabled?: boolean;
}

const MAX_FILE_SIZE = 5 * 1024 * 1024; // 5MB

export function AudioUploadField({
	label,
	description,
	value,
	onChange,
	disabled = false,
}: AudioUploadFieldProps) {
	const [isUploading, setIsUploading] = useState(false);
	const [showManualInput, setShowManualInput] = useState(false);
	const fileInputRef = useRef<HTMLInputElement>(null);

	const handleFileSelect = async (e: ChangeEvent<HTMLInputElement>) => {
		const file = e.target.files?.[0];
		if (!file) return;

		// Validate file size
		if (file.size > MAX_FILE_SIZE) {
			toast.error("File size exceeds 5MB limit");
			if (fileInputRef.current) fileInputRef.current.value = "";
			return;
		}

		// Validate MIME type
		const validTypes = [
			"audio/mpeg",
			"audio/mp3",
			"audio/wav",
			"audio/x-wav",
			"audio/ogg",
			"audio/x-m4a",
			"audio/m4a",
			"audio/aac",
			"audio/mp4",
		];
		const isAudioExt = /\.(mp3|wav|ogg|m4a|aac)$/i.test(file.name);
		if (!validTypes.includes(file.type) && !isAudioExt) {
			toast.error("Please upload an audio file (.mp3, .wav, .ogg, .m4a)");
			if (fileInputRef.current) fileInputRef.current.value = "";
			return;
		}

		setIsUploading(true);
		try {
			const res = await servicesApi.uploadAudio(file);
			if (res.data?.url) {
				onChange(res.data.url);
				toast.success("Audio file uploaded successfully");
			} else {
				throw new Error("No URL returned from server");
			}
		} catch (error: any) {
			toast.error(error.message || "Failed to upload audio file");
		} finally {
			setIsUploading(false);
			if (fileInputRef.current) fileInputRef.current.value = "";
		}
	};

	const handleRemove = () => {
		onChange("");
		if (fileInputRef.current) fileInputRef.current.value = "";
	};

	return (
		<FormItem className="space-y-2">
			<div className="flex items-center justify-between">
				<FormLabel>{label}</FormLabel>
				<Button
					type="button"
					variant="ghost"
					size="sm"
					className="h-6 px-2 text-xs text-muted-foreground hover:text-foreground"
					onClick={() => setShowManualInput(!showManualInput)}
				>
					<LinkIcon className="mr-1 h-3 w-3" />
					{showManualInput ? "Hide URL Input" : "Edit URL"}
				</Button>
			</div>

			<FormControl>
				<div className="space-y-2">
					<input
						ref={fileInputRef}
						type="file"
						accept="audio/mpeg,audio/wav,audio/ogg,audio/x-m4a,audio/m4a,audio/aac"
						className="hidden"
						onChange={handleFileSelect}
						disabled={disabled || isUploading}
					/>

					{value ? (
						<div className="rounded-lg border bg-muted/40 p-3 space-y-2">
							<div className="flex items-center justify-between gap-2">
								<div className="flex items-center gap-2 min-w-0">
									<Music className="h-4 w-4 shrink-0 text-primary" />
									<span className="truncate text-xs font-mono text-muted-foreground">
										{value}
									</span>
								</div>
								<div className="flex items-center gap-1 shrink-0">
									<Button
										type="button"
										variant="ghost"
										size="icon"
										className="h-7 w-7 text-muted-foreground hover:text-foreground"
										asChild
									>
										<a
											href={value}
											target="_blank"
											rel="noopener noreferrer"
											title="Open audio in new tab"
										>
											<ExternalLink className="h-3.5 w-3.5" />
										</a>
									</Button>
									<Button
										type="button"
										variant="ghost"
										size="icon"
										className="h-7 w-7 text-destructive hover:bg-destructive/10"
										onClick={handleRemove}
										disabled={disabled || isUploading}
										title="Remove audio"
									>
										<Trash2 className="h-3.5 w-3.5" />
									</Button>
								</div>
							</div>

							<audio
								controls
								src={value}
								className="h-8 w-full rounded"
								preload="metadata"
							>
								Your browser does not support the audio element.
							</audio>

							<div className="flex justify-end pt-1">
								<Button
									type="button"
									variant="outline"
									size="sm"
									className="h-7 text-xs"
									onClick={() => fileInputRef.current?.click()}
									disabled={disabled || isUploading}
								>
									{isUploading ? (
										<>
											<Loader2 className="mr-1.5 h-3 w-3 animate-spin" />
											Uploading...
										</>
									) : (
										<>
											<UploadCloud className="mr-1.5 h-3 w-3" />
											Replace Audio
										</>
									)}
								</Button>
							</div>
						</div>
					) : (
						<div
							className="flex flex-col items-center justify-center rounded-lg border border-dashed p-4 text-center hover:bg-muted/30 transition-colors cursor-pointer"
							onClick={() =>
								!isUploading && !disabled && fileInputRef.current?.click()
							}
						>
							{isUploading ? (
								<div className="flex flex-col items-center gap-1.5 py-1">
									<Loader2 className="h-6 w-6 animate-spin text-primary" />
									<p className="text-xs font-medium text-foreground">
										Uploading audio to storage...
									</p>
								</div>
							) : (
								<div className="flex flex-col items-center gap-1 py-1">
									<div className="rounded-full bg-primary/10 p-2 text-primary">
										<UploadCloud className="h-5 w-5" />
									</div>
									<p className="text-xs font-medium text-foreground">
										Click to upload audio file
									</p>
									<p className="text-[11px] text-muted-foreground">
										MP3, WAV, OGG, M4A up to 5MB
									</p>
								</div>
							)}
						</div>
					)}

					{showManualInput && (
						<div className="pt-1">
							<Input
								value={value || ""}
								onChange={(e) => onChange(e.target.value)}
								placeholder="https://... or /uploads/..."
								disabled={disabled || isUploading}
								className="font-mono text-xs"
							/>
						</div>
					)}
				</div>
			</FormControl>

			{description && <FormDescription>{description}</FormDescription>}
			<FormMessage />
		</FormItem>
	);
}
