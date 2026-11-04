"use client";

import {
	AlertDialog,
	AlertDialogAction,
	AlertDialogCancel,
	AlertDialogContent,
	AlertDialogDescription,
	AlertDialogFooter,
	AlertDialogHeader,
	AlertDialogTitle,
} from "~/components/ui/alert-dialog";
import { Button } from "~/components/ui/button";
import { useUploadStore } from "~/lib/upload";

/**
 * Surfaces a prompt when a user uploads a file with the same name in the same target.
 * Resolution: replace, keep both, or skip — applied per pending entry.
 */
export function UploadDuplicateDialog() {
	const pending = useUploadStore((s) => s.duplicatePrompt.pending);
	const resolve = useUploadStore((s) => s.resolveDuplicate);
	const dismiss = useUploadStore((s) => s.dismissDuplicatePrompt);

	const current = pending[0];
	const open = Boolean(current);

	return (
		<AlertDialog open={open} onOpenChange={(o) => !o && dismiss()}>
			<AlertDialogContent>
				<AlertDialogHeader>
					<AlertDialogTitle>File already exists</AlertDialogTitle>
					<AlertDialogDescription>
						<span className="font-medium text-foreground">
							{current?.file.name}
						</span>{" "}
						already exists in this folder. Choose how to handle this upload.
						{pending.length > 1 && (
							<span className="mt-2 block text-xs text-muted-foreground">
								{pending.length - 1} more conflict
								{pending.length - 1 > 1 ? "s" : ""} after this.
							</span>
						)}
					</AlertDialogDescription>
				</AlertDialogHeader>
				<AlertDialogFooter className="gap-2 sm:gap-2">
					<AlertDialogCancel asChild>
						<Button
							variant="ghost"
							onClick={() => current && resolve(current.existingId, "skip")}
						>
							Skip
						</Button>
					</AlertDialogCancel>
					<Button
						variant="outline"
						onClick={() => current && resolve(current.existingId, "keep-both")}
					>
						Keep both
					</Button>
					<AlertDialogAction asChild>
						<Button
							onClick={() => current && resolve(current.existingId, "replace")}
						>
							Replace
						</Button>
					</AlertDialogAction>
				</AlertDialogFooter>
			</AlertDialogContent>
		</AlertDialog>
	);
}
