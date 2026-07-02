"use client";

import { useEffect, useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
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
	FormField,
	FormItem,
	FormLabel,
	FormMessage,
	FormDescription,
} from "~/components/ui/form";
import { Input } from "~/components/ui/input";
import { Switch } from "~/components/ui/switch";
import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from "~/components/ui/select";
import { toast } from "sonner";
import { Icon } from "~/components/shared/icon";
import { servicesApi, type Service } from "~/lib/api/qms";

const serviceSchema = z.object({
	code: z.string().min(2, "Code must be at least 2 characters.").max(50),
	name: z.string().min(3, "Name must be at least 3 characters.").max(255),
	type: z.string().optional(),
	default_estimated_duration: z.coerce.number().int().nonnegative().optional(),
	is_pharmacy: z.boolean().default(false),
	is_pharmacy_reception: z.boolean().default(false),
	status: z.enum(["active", "inactive"]).optional(),
});

type ServiceFormValues = z.infer<typeof serviceSchema>;

interface ServiceDialogProps {
	open: boolean;
	onOpenChange: (open: boolean) => void;
	service?: Service | null;
	onSuccess: () => void;
}

export function ServiceDialog({
	open,
	onOpenChange,
	service,
	onSuccess,
}: ServiceDialogProps) {
	const [isLoading, setIsLoading] = useState(false);
	const isEdit = !!service;

	const form = useForm<ServiceFormValues>({
		resolver: zodResolver(serviceSchema),
		defaultValues: {
			code: "",
			name: "",
			type: "",
			default_estimated_duration: 0,
			is_pharmacy: false,
			is_pharmacy_reception: false,
			status: "active",
		},
	});

	useEffect(() => {
		if (open) {
			form.reset({
				code: service?.code || "",
				name: service?.name || "",
				type: service?.type || "",
				default_estimated_duration: service?.default_estimated_duration || 0,
				is_pharmacy: service?.is_pharmacy || false,
				is_pharmacy_reception: service?.is_pharmacy_reception || false,
				status: service?.status || "active",
			});
		}
	}, [service, open, form]);

	async function onSubmit(data: ServiceFormValues) {
		setIsLoading(true);
		try {
			if (isEdit && service) {
				await servicesApi.update(service.id, {
					code: data.code,
					name: data.name,
					type: data.type,
					default_estimated_duration: data.default_estimated_duration,
					status: data.status,
					is_pharmacy: data.is_pharmacy,
					is_pharmacy_reception: data.is_pharmacy_reception,
				});
				toast.success("Service updated successfully");
			} else {
				await servicesApi.create({
					code: data.code,
					name: data.name,
					type: data.type,
					default_estimated_duration: data.default_estimated_duration,
					is_pharmacy: data.is_pharmacy,
					is_pharmacy_reception: data.is_pharmacy_reception,
				});
				toast.success("Service created successfully");
			}
			onSuccess();
			onOpenChange(false);
		} catch (error: any) {
			toast.error(error.message || "Failed to save service");
		} finally {
			setIsLoading(false);
		}
	}

	return (
		<Dialog open={open} onOpenChange={onOpenChange}>
			<DialogContent className="sm:max-w-[425px]">
				<DialogHeader>
					<DialogTitle>{isEdit ? "Edit Service" : "Add Service"}</DialogTitle>
					<DialogDescription>
						{isEdit
							? "Update the configuration for this queue service."
							: "Create a new service flow for your queue."}
					</DialogDescription>
				</DialogHeader>
				<Form {...form}>
					<form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
						<div className="grid grid-cols-2 gap-4">
							<FormField
								control={form.control}
								name="code"
								render={({ field }) => (
									<FormItem>
										<FormLabel>Service Code</FormLabel>
										<FormControl>
											<Input placeholder="e.g. A" {...field} />
										</FormControl>
										<FormMessage />
									</FormItem>
								)}
							/>
							{isEdit && (
								<FormField
									control={form.control}
									name="status"
									render={({ field }) => (
										<FormItem>
											<FormLabel>Status</FormLabel>
											<Select
												onValueChange={field.onChange}
												defaultValue={field.value}
											>
												<FormControl>
													<SelectTrigger>
														<SelectValue placeholder="Select status" />
													</SelectTrigger>
												</FormControl>
												<SelectContent>
													<SelectItem value="active">Active</SelectItem>
													<SelectItem value="inactive">Inactive</SelectItem>
												</SelectContent>
											</Select>
											<FormMessage />
										</FormItem>
									)}
								/>
							)}
						</div>

						<FormField
							control={form.control}
							name="name"
							render={({ field }) => (
								<FormItem>
									<FormLabel>Service Name</FormLabel>
									<FormControl>
										<Input placeholder="e.g. General Checkup" {...field} />
									</FormControl>
									<FormMessage />
								</FormItem>
							)}
						/>

						<FormField
							control={form.control}
							name="type"
							render={({ field }) => (
								<FormItem>
									<FormLabel>Service Type</FormLabel>
									<FormControl>
										<Input placeholder="e.g. primary" {...field} />
									</FormControl>
									<FormMessage />
								</FormItem>
							)}
						/>

						<FormField
							control={form.control}
							name="default_estimated_duration"
							render={({ field }) => (
								<FormItem>
									<FormLabel>Default Estimated Duration</FormLabel>
									<FormControl>
										<Input type="number" min={0} placeholder="15" {...field} />
									</FormControl>
									<FormMessage />
								</FormItem>
							)}
						/>

						<div className="space-y-4 rounded-lg border p-4">
							<h4 className="text-sm font-medium leading-none">
								Domain Specific Logic
							</h4>
							<FormField
								control={form.control}
								name="is_pharmacy"
								render={({ field }) => (
									<FormItem className="flex flex-row items-center justify-between rounded-lg border p-3 shadow-sm">
										<div className="space-y-0.5">
											<FormLabel>Pharmacy Flow</FormLabel>
											<FormDescription>
												Treats this service as a pharmacy counter.
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
							<FormField
								control={form.control}
								name="is_pharmacy_reception"
								render={({ field }) => (
									<FormItem className="flex flex-row items-center justify-between rounded-lg border p-3 shadow-sm">
										<div className="space-y-0.5">
											<FormLabel>Prescription Reception</FormLabel>
											<FormDescription>
												Uses prescription dropping logic before queueing.
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
						</div>

						<DialogFooter className="pt-4">
							<Button type="submit" disabled={isLoading}>
								{isLoading && (
									<Icon name="Loader" className="mr-2 h-4 w-4 animate-spin" />
								)}
								{isEdit ? "Save Changes" : "Create Service"}
							</Button>
						</DialogFooter>
					</form>
				</Form>
			</DialogContent>
		</Dialog>
	);
}
