"use client";

import { useEffect, useState } from "react";
import { toast } from "sonner";
import { Button } from "~/components/ui/button";
import {
	Dialog,
	DialogContent,
	DialogDescription,
	DialogFooter,
	DialogHeader,
	DialogTitle,
} from "~/components/ui/dialog";
import { Input } from "~/components/ui/input";
import { Label } from "~/components/ui/label";
import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from "~/components/ui/select";
import { Icon } from "~/components/shared/icon";
import { settingsApi } from "~/lib/api/qms";

const CONFIG_FIELDS = [
	{ key: "queue_reset_time", label: "Queue Reset Time", placeholder: "04:00" },
	{ key: "ticket_prefix", label: "Ticket Prefix", placeholder: "A" },
	{
		key: "default_estimated_duration",
		label: "Default Duration (min)",
		placeholder: "5",
	},
	{
		key: "numbering_strategy",
		label: "Numbering Strategy",
		placeholder: "daily_branch_sequence",
		type: "select",
		options: ["sequential", "daily_branch_sequence", "random"],
	},
	{
		key: "allow_recall",
		label: "Allow Recall",
		type: "select",
		options: ["true", "false"],
	},
	{
		key: "allow_skip",
		label: "Allow Skip",
		type: "select",
		options: ["true", "false"],
	},
	{
		key: "allow_cancel",
		label: "Allow Cancel",
		type: "select",
		options: ["true", "false"],
	},
	{
		key: "allow_forward",
		label: "Allow Forward",
		type: "select",
		options: ["true", "false"],
	},
	{
		key: "auto_call_next",
		label: "Auto Call Next",
		type: "select",
		options: ["true", "false"],
	},
	{
		key: "require_counter",
		label: "Require Counter",
		type: "select",
		options: ["true", "false"],
	},
	{
		key: "allow_forward_from",
		label: "Allow Forward From",
		type: "select",
		options: ["true", "false"],
	},
	{
		key: "allow_forward_to",
		label: "Allow Forward To",
		type: "select",
		options: ["true", "false"],
	},
];

export type OverrideScope = {
	type: "tenant" | "branch" | "branch_service" | "counter";
	branchId?: string;
	branchServiceId?: string;
	counterId?: string;
};

type QueueConfigDialogProps = {
	open: boolean;
	onOpenChange: (open: boolean) => void;
	scope: OverrideScope;
	onSuccess: () => void;
};

export function QueueConfigDialog({
	open,
	onOpenChange,
	scope,
	onSuccess,
}: QueueConfigDialogProps) {
	const [values, setValues] = useState<Record<string, string>>({});
	const [saving, setSaving] = useState(false);

	useEffect(() => {
		setValues({});
	}, [open]);

	const setField = (key: string, value: string) => {
		setValues((prev) => ({ ...prev, [key]: value }));
	};

	const handleSave = async () => {
		setSaving(true);
		try {
			const payload: Record<string, any> = { ...values };
			// Parse numeric field
			if (payload.default_estimated_duration) {
				payload.default_estimated_duration = String(
					parseInt(payload.default_estimated_duration, 10),
				);
			}
			// Parse boolean fields
			for (const field of CONFIG_FIELDS) {
				if (field.type === "select" && payload[field.key] !== undefined) {
					const val = payload[field.key];
					if (val === "true" || val === "false") {
						payload[field.key] = val === "true";
					}
				}
			}

			switch (scope.type) {
				case "tenant":
					await settingsApi.updateTenant(payload);
					break;
				case "branch":
					await settingsApi.updateBranch(scope.branchId!, payload);
					break;
				case "branch_service":
					await settingsApi.updateBranchService(
						scope.branchId!,
						scope.branchServiceId!,
						payload,
					);
					break;
				case "counter":
					await settingsApi.updateCounter(
						scope.branchId!,
						scope.counterId!,
						payload,
					);
					break;
			}
			toast.success("Queue config override saved");
			onSuccess();
			onOpenChange(false);
		} catch {
			toast.error("Failed to save override");
		} finally {
			setSaving(false);
		}
	};

	if (!open) return null;

	return (
		<Dialog open={open} onOpenChange={onOpenChange}>
			<DialogContent className="max-h-[90vh] overflow-y-auto sm:max-w-[500px]">
				<DialogHeader>
					<DialogTitle>Override Queue Config</DialogTitle>
					<DialogDescription>
						Set typed configuration for scope:{" "}
						<span className="font-medium">{scope.type}</span>
						{scope.branchId && (
							<span className="block text-xs text-muted-foreground">
								Branch: {scope.branchId}
							</span>
						)}
					</DialogDescription>
				</DialogHeader>
				<div className="grid gap-4 py-2">
					{CONFIG_FIELDS.map((field) => (
						<div key={field.key} className="grid gap-1.5">
							<Label htmlFor={field.key} className="text-sm">
								{field.label}
							</Label>
							{field.type === "select" ? (
								<Select
									value={values[field.key] ?? ""}
									onValueChange={(v) => setField(field.key, v)}
								>
									<SelectTrigger id={field.key}>
										<SelectValue placeholder="Default" />
									</SelectTrigger>
									<SelectContent>
										<SelectItem value="">Default</SelectItem>
										{field.options?.map((opt) => (
											<SelectItem key={opt} value={opt}>
												{opt}
											</SelectItem>
										))}
									</SelectContent>
								</Select>
							) : (
								<Input
									id={field.key}
									placeholder={field.placeholder}
									value={values[field.key] ?? ""}
									onChange={(e) => setField(field.key, e.target.value)}
								/>
							)}
						</div>
					))}
				</div>
				<DialogFooter>
					<Button
						variant="outline"
						onClick={() => onOpenChange(false)}
						disabled={saving}
					>
						Cancel
					</Button>
					<Button
						onClick={handleSave}
						disabled={saving || Object.keys(values).length === 0}
					>
						{saving && (
							<Icon name="Loader" className="mr-2 h-4 w-4 animate-spin" />
						)}
						Save Override
					</Button>
				</DialogFooter>
			</DialogContent>
		</Dialog>
	);
}
