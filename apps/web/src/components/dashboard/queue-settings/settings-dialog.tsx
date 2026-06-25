"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useCallback, useEffect, useState } from "react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";
import { Icon } from "~/components/shared/icon";
import { Button } from "~/components/ui/button";
import {
	Dialog,
	DialogContent,
	DialogDescription,
	DialogFooter,
	DialogHeader,
	DialogTitle,
} from "~/components/ui/dialog";
import {
	Form,
	FormControl,
	FormDescription,
	FormField,
	FormItem,
	FormLabel,
	FormMessage,
} from "~/components/ui/form";
import { Input } from "~/components/ui/input";
import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from "~/components/ui/select";
import { Switch } from "~/components/ui/switch";
import {
	type Branch,
	branchesApi,
	type Counter,
	countersApi,
	type Service,
	servicesApi,
	type Setting,
	settingsApi,
} from "~/lib/api/qms";
import { useOrganizationStore } from "~/stores/use-organization-store";

const SCOPE_TYPES = ["tenant", "branch", "service", "counter"] as const;

const settingsSchema = z.object({
	scope_type: z.enum(SCOPE_TYPES),
	scope_id: z.string().min(1, "Scope target is required."),
	key: z.string().min(1, "Key is required.").max(100),
	value: z.string().min(1, "Value is required."),
	value_type: z.enum(["string", "number", "boolean", "json"]).optional(),
	is_active: z.boolean().optional(),
});

type SettingsFormValues = z.infer<typeof settingsSchema>;

interface SettingsDialogProps {
	open: boolean;
	onOpenChange: (open: boolean) => void;
	setting?: Setting | null;
	onSuccess: () => void;
}

export function SettingsDialog({
	open,
	onOpenChange,
	setting,
	onSuccess,
}: SettingsDialogProps) {
	const { currentOrganization } = useOrganizationStore();
	const [isLoading, setIsLoading] = useState(false);
	const [branches, setBranches] = useState<Branch[]>([]);
	const [services, setServices] = useState<Service[]>([]);
	const [counters, setCounters] = useState<Counter[]>([]);
	const isEdit = !!setting;

	const form = useForm<SettingsFormValues>({
		resolver: zodResolver(settingsSchema),
		defaultValues: {
			scope_type: "tenant",
			scope_id: "",
			key: "",
			value: "",
			value_type: "string",
			is_active: true,
		},
	});

	const scopeType = form.watch("scope_type");

	// Fetch options for scope selectors when scope_type changes
	const fetchScopeOptions = useCallback(async () => {
		try {
			if (scopeType === "branch") {
				const resp = await branchesApi.getAll();
				setBranches(resp.data || []);
			} else if (scopeType === "service") {
				const resp = await servicesApi.getAll();
				setServices(resp.data || []);
			} else if (scopeType === "counter") {
				const resp = await countersApi.getAll();
				setCounters(resp.data || []);
			}
		} catch {
			// silenty ignore fetch errors
		}
	}, [scopeType]);

	useEffect(() => {
		if (open) {
			fetchScopeOptions();
			// Reset scope_id when scope_type changes
			form.setValue("scope_id", "");
		}
	}, [scopeType, open, fetchScopeOptions, form]);

	useEffect(() => {
		if (open) {
			form.reset({
				scope_type: setting?.scope_type || "tenant",
				scope_id: setting?.scope_id || "",
				key: setting?.key || "",
				value: setting?.value || "",
				value_type: setting?.value_type || "string",
				is_active: setting?.is_active ?? true,
			});
		}
	}, [setting, open, form]);

	async function onSubmit(data: SettingsFormValues) {
		setIsLoading(true);
		try {
			if (isEdit && setting) {
				await settingsApi.update(setting.id, {
					value: data.value,
					is_active: data.is_active,
				});
				toast.success("Setting updated successfully");
			} else {
				await settingsApi.create({
					scope_type: data.scope_type,
					scope_id:
						data.scope_type === "tenant"
							? currentOrganization?.id || ""
							: data.scope_id || "",
					key: data.key,
					value: data.value,
					value_type: data.value_type,
				});
				toast.success("Setting created successfully");
			}
			onSuccess();
			onOpenChange(false);
		} catch (error: any) {
			toast.error(error.message || "Failed to save setting");
		} finally {
			setIsLoading(false);
		}
	}

	function renderScopeIdSelector() {
		if (scopeType === "tenant") return null;

		const items =
			scopeType === "branch"
				? branches
				: scopeType === "service"
					? services
					: scopeType === "counter"
						? counters
						: [];

		const label =
			scopeType === "branch"
				? "Branch"
				: scopeType === "service"
					? "Service"
					: "Counter";

		return (
			<FormField
				control={form.control}
				name="scope_id"
				render={({ field }) => (
					<FormItem>
						<FormLabel>{label}</FormLabel>
						<Select onValueChange={field.onChange} defaultValue={field.value}>
							<FormControl>
								<SelectTrigger>
									<SelectValue
										placeholder={`Select a ${label.toLowerCase()}`}
									/>
								</SelectTrigger>
							</FormControl>
							<SelectContent>
								{items.length === 0 ? (
									<SelectItem value="" disabled>
										No {label.toLowerCase()}s available
									</SelectItem>
								) : (
									items.map((item) => (
										<SelectItem key={item.id} value={item.id}>
											{item.code} — {item.name}
										</SelectItem>
									))
								)}
							</SelectContent>
						</Select>
						<FormMessage />
					</FormItem>
				)}
			/>
		);
	}

	return (
		<Dialog open={open} onOpenChange={onOpenChange}>
			<DialogContent className="sm:max-w-[500px]">
				<DialogHeader>
					<DialogTitle>
						{isEdit ? "Edit Setting" : "Add Setting Override"}
					</DialogTitle>
					<DialogDescription>
						{isEdit
							? "Update the value or status of this setting."
							: "Define a new configuration key at the chosen scope level."}
					</DialogDescription>
				</DialogHeader>
				<Form {...form}>
					<form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
						<div className="grid grid-cols-2 gap-4">
							<FormField
								control={form.control}
								name="scope_type"
								render={({ field }) => (
									<FormItem>
										<FormLabel>Scope Level</FormLabel>
										<Select
											onValueChange={field.onChange}
											defaultValue={field.value}
											disabled={isEdit}
										>
											<FormControl>
												<SelectTrigger>
													<SelectValue placeholder="Select scope" />
												</SelectTrigger>
											</FormControl>
											<SelectContent>
												{SCOPE_TYPES.map((t) => (
													<SelectItem key={t} value={t}>
														{t.charAt(0).toUpperCase() + t.slice(1)}
													</SelectItem>
												))}
											</SelectContent>
										</Select>
										<FormMessage />
									</FormItem>
								)}
							/>
							<FormField
								control={form.control}
								name="key"
								render={({ field }) => (
									<FormItem>
										<FormLabel>Setting Key</FormLabel>
										<FormControl>
											<Input
												placeholder="e.g. max_queue"
												{...field}
												disabled={isEdit}
											/>
										</FormControl>
										<FormMessage />
									</FormItem>
								)}
							/>
						</div>

						{renderScopeIdSelector()}

						<FormField
							control={form.control}
							name="value"
							render={({ field }) => (
								<FormItem>
									<FormLabel>Value</FormLabel>
									<FormControl>
										<Input placeholder="Setting value" {...field} />
									</FormControl>
									<FormMessage />
								</FormItem>
							)}
						/>

						<FormField
							control={form.control}
							name="value_type"
							render={({ field }) => (
								<FormItem>
									<FormLabel>Value Type</FormLabel>
									<Select
										onValueChange={field.onChange}
										defaultValue={field.value}
									>
										<FormControl>
											<SelectTrigger>
												<SelectValue placeholder="Select type" />
											</SelectTrigger>
										</FormControl>
										<SelectContent>
											<SelectItem value="string">String</SelectItem>
											<SelectItem value="number">Number</SelectItem>
											<SelectItem value="boolean">Boolean</SelectItem>
											<SelectItem value="json">JSON</SelectItem>
										</SelectContent>
									</Select>
									<FormMessage />
								</FormItem>
							)}
						/>

						{isEdit && (
							<FormField
								control={form.control}
								name="is_active"
								render={({ field }) => (
									<FormItem className="flex flex-row items-center justify-between rounded-lg border p-3 shadow-sm">
										<div className="space-y-0.5">
											<FormLabel>Active</FormLabel>
											<FormDescription>
												Inactive settings are ignored during resolution.
											</FormDescription>
										</div>
										<FormControl>
											<Switch
												checked={field.value}
												onCheckedChange={field.onChange}
											/>
										</FormControl>
									</FormItem>
								)}
							/>
						)}

						<DialogFooter className="pt-4">
							<Button type="submit" disabled={isLoading}>
								{isLoading && (
									<Icon name="Loader" className="mr-2 h-4 w-4 animate-spin" />
								)}
								{isEdit ? "Save Changes" : "Create Setting"}
							</Button>
						</DialogFooter>
					</form>
				</Form>
			</DialogContent>
		</Dialog>
	);
}
