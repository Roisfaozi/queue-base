"use client";

import { useEffect, useState, useCallback } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { Button } from "~/components/ui/button";
import {
	Dialog,
	DialogContent,
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
} from "~/components/ui/form";
import { Input } from "~/components/ui/input";
import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from "~/components/ui/select";
import { toast } from "sonner";
import { Icon } from "~/components/shared/icon";
import {
	queuesApi,
	servicesApi,
	type Service,
	type Branch,
} from "~/lib/api/qms";

const registerSchema = z.object({
	branch_id: z.string().min(1, "Branch is required."),
	service_id: z.string().min(1, "Service is required."),
	patient_name: z.string().min(2, "Name must be at least 2 characters."),
});

type RegisterFormValues = z.infer<typeof registerSchema>;

interface QueueRegisterDialogProps {
	open: boolean;
	onOpenChange: (open: boolean) => void;
	branches: Branch[];
	defaultBranchId?: string;
	onSuccess: () => void;
}

export function QueueRegisterDialog({
	open,
	onOpenChange,
	branches,
	defaultBranchId,
	onSuccess,
}: QueueRegisterDialogProps) {
	const [isLoading, setIsLoading] = useState(false);
	const [services, setServices] = useState<Service[]>([]);

	const form = useForm<RegisterFormValues>({
		resolver: zodResolver(registerSchema),
		defaultValues: {
			branch_id: "",
			service_id: "",
			patient_name: "",
		},
	});

	const fetchServices = useCallback(async () => {
		try {
			const resp = await servicesApi.getAll();
			setServices(resp.data?.filter((s) => s.status === "active") || []);
		} catch {
			// Handle silently
		}
	}, []);

	useEffect(() => {
		if (open) {
			form.reset({
				branch_id: defaultBranchId || "",
				service_id: "",
				patient_name: "",
			});
			fetchServices();
		}
	}, [open, form, fetchServices, defaultBranchId]);

	async function onSubmit(data: RegisterFormValues) {
		setIsLoading(true);
		try {
			await queuesApi.register({
				branch_id: data.branch_id,
				service_id: data.service_id,
				patient_name: data.patient_name,
			});
			toast.success("Queue registered successfully");
			onSuccess();
			onOpenChange(false);
		} catch (error: any) {
			toast.error(error.message || "Failed to register queue");
		} finally {
			setIsLoading(false);
		}
	}

	return (
		<Dialog open={open} onOpenChange={onOpenChange}>
			<DialogContent className="sm:max-w-[425px]">
				<DialogHeader>
					<DialogTitle>Register Patient</DialogTitle>
				</DialogHeader>
				<Form {...form}>
					<form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
						<FormField
							control={form.control}
							name="branch_id"
							render={({ field }) => (
								<FormItem>
									<FormLabel>Branch</FormLabel>
									<Select
										onValueChange={field.onChange}
										defaultValue={field.value}
									>
										<FormControl>
											<SelectTrigger>
												<SelectValue placeholder="Select a branch" />
											</SelectTrigger>
										</FormControl>
										<SelectContent>
											{branches.length === 0 ? (
												<SelectItem value="__no_branches__" disabled>
													No branches available
												</SelectItem>
											) : (
												branches.map((branch) => (
													<SelectItem key={branch.id} value={branch.id}>
														{branch.code} — {branch.name}
													</SelectItem>
												))
											)}
										</SelectContent>
									</Select>
									<FormMessage />
								</FormItem>
							)}
						/>
						<FormField
							control={form.control}
							name="service_id"
							render={({ field }) => (
								<FormItem>
									<FormLabel>Service</FormLabel>
									<Select
										onValueChange={field.onChange}
										defaultValue={field.value}
									>
										<FormControl>
											<SelectTrigger>
												<SelectValue placeholder="Select destination service" />
											</SelectTrigger>
										</FormControl>
										<SelectContent>
											{services.map((svc) => (
												<SelectItem key={svc.id} value={svc.id}>
													{svc.code} — {svc.name}
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
							name="patient_name"
							render={({ field }) => (
								<FormItem>
									<FormLabel>Patient Name</FormLabel>
									<FormControl>
										<Input placeholder="Enter patient name" {...field} />
									</FormControl>
									<FormMessage />
								</FormItem>
							)}
						/>
						<div className="pt-4 flex justify-end">
							<Button type="submit" disabled={isLoading}>
								{isLoading && (
									<Icon name="Loader" className="mr-2 h-4 w-4 animate-spin" />
								)}
								Issue Ticket
							</Button>
						</div>
					</form>
				</Form>
			</DialogContent>
		</Dialog>
	);
}
