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
	countersApi,
	branchServicesApi,
	servicesApi,
	type Counter,
	type Branch,
	type BranchService,
	type Service,
} from "~/lib/api/qms";

const counterSchema = z.object({
	branch_id: z.string().min(1, "Branch is required."),
	branch_service_id: z.string().optional(),
	code: z.string().min(2, "Code must be at least 2 characters.").max(50),
	name: z.string().min(3, "Name must be at least 3 characters.").max(255),
	display_name: z.string().optional(),
	status: z.enum(["active", "inactive"]).optional(),
});

type CounterFormValues = z.infer<typeof counterSchema>;

interface CounterDialogProps {
	open: boolean;
	onOpenChange: (open: boolean) => void;
	counter?: Counter | null;
	branches: Branch[];
	onSuccess: () => void;
}

export function CounterDialog({
	open,
	onOpenChange,
	counter,
	branches,
	onSuccess,
}: CounterDialogProps) {
	const [isLoading, setIsLoading] = useState(false);
	const [branchServices, setBranchServices] = useState<BranchService[]>([]);
	const [servicesMap, setServicesMap] = useState<Record<string, Service>>({});
	const isEdit = !!counter;

	const form = useForm<CounterFormValues>({
		resolver: zodResolver(counterSchema),
		defaultValues: {
			branch_id: "",
			branch_service_id: "",
			code: "",
			name: "",
			display_name: "",
			status: "active",
		},
	});

	const selectedBranchId = form.watch("branch_id");

	useEffect(() => {
		async function loadServices() {
			try {
				const resp = await servicesApi.getAll();
				const map: Record<string, Service> = {};
				for (const s of resp.data || []) {
					map[s.id] = s;
				}
				setServicesMap(map);
			} catch {
				// ignore
			}
		}
		loadServices();
	}, []);

	useEffect(() => {
		async function loadBranchServices(branchId: string) {
			if (!branchId) {
				setBranchServices([]);
				return;
			}
			try {
				const resp = await branchServicesApi.getByBranch(branchId);
				setBranchServices(resp.data || []);
			} catch {
				setBranchServices([]);
			}
		}
		loadBranchServices(selectedBranchId);
	}, [selectedBranchId]);

	useEffect(() => {
		if (open) {
			form.reset({
				branch_id: counter?.branch_id || "",
				branch_service_id: counter?.branch_service_id || "",
				code: counter?.code || "",
				name: counter?.name || "",
				display_name: counter?.display_name || "",
				status: counter?.status || "active",
			});
		}
	}, [counter, open, form]);

	async function onSubmit(data: CounterFormValues) {
		setIsLoading(true);
		try {
			if (isEdit && counter) {
				await countersApi.update(counter.id, {
					code: data.code,
					name: data.name,
					branch_service_id: data.branch_service_id,
					display_name: data.display_name,
					status: data.status,
				});
				toast.success("Counter updated successfully");
			} else {
				await countersApi.create({
					branch_id: data.branch_id,
					branch_service_id: data.branch_service_id,
					code: data.code,
					name: data.name,
					display_name: data.display_name,
				});
				toast.success("Counter created successfully");
			}
			onSuccess();
			onOpenChange(false);
		} catch (error: any) {
			toast.error(error.message || "Failed to save counter");
		} finally {
			setIsLoading(false);
		}
	}

	return (
		<Dialog open={open} onOpenChange={onOpenChange}>
			<DialogContent className="sm:max-w-[425px]">
				<DialogHeader>
					<DialogTitle>{isEdit ? "Edit Counter" : "Add Counter"}</DialogTitle>
					<DialogDescription>
						{isEdit
							? "Update the configuration for this counter."
							: "Create a new counter assigned to a branch."}
					</DialogDescription>
				</DialogHeader>
				<Form {...form}>
					<form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
						{!isEdit && (
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
						)}

						<FormField
							control={form.control}
							name="branch_service_id"
							render={({ field }) => (
								<FormItem>
									<FormLabel>Assigned Service (Optional)</FormLabel>
									<Select
										onValueChange={field.onChange}
										value={field.value || ""}
									>
										<FormControl>
											<SelectTrigger>
												<SelectValue placeholder="Select a service" />
											</SelectTrigger>
										</FormControl>
										<SelectContent>
											{branchServices.length === 0 ? (
												<SelectItem value="__no_services__" disabled>
													No services assigned to this branch
												</SelectItem>
											) : (
												branchServices.map((bs) => {
													const sName =
														bs.custom_name ||
														servicesMap[bs.service_id]?.name ||
														"Unknown Service";
													const sCode = servicesMap[bs.service_id]?.code || "-";
													return (
														<SelectItem key={bs.id} value={bs.id}>
															{sCode} — {sName}
														</SelectItem>
													);
												})
											)}
										</SelectContent>
									</Select>
									<FormDescription>
										Link this counter to a specific service flow.
									</FormDescription>
									<FormMessage />
								</FormItem>
							)}
						/>

						<div className="grid grid-cols-2 gap-4">
							<FormField
								control={form.control}
								name="code"
								render={({ field }) => (
									<FormItem>
										<FormLabel>Counter Code</FormLabel>
										<FormControl>
											<Input placeholder="e.g. C1" {...field} />
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
									<FormLabel>Counter Name</FormLabel>
									<FormControl>
										<Input placeholder="e.g. Counter 1" {...field} />
									</FormControl>
									<FormMessage />
								</FormItem>
							)}
						/>

						<FormField
							control={form.control}
							name="display_name"
							render={({ field }) => (
								<FormItem>
									<FormLabel>Display Name</FormLabel>
									<FormControl>
										<Input placeholder="e.g. Layanan 1" {...field} />
									</FormControl>
									<FormMessage />
								</FormItem>
							)}
						/>

						<DialogFooter className="pt-4">
							<Button type="submit" disabled={isLoading}>
								{isLoading && (
									<Icon name="Loader" className="mr-2 h-4 w-4 animate-spin" />
								)}
								{isEdit ? "Save Changes" : "Create Counter"}
							</Button>
						</DialogFooter>
					</form>
				</Form>
			</DialogContent>
		</Dialog>
	);
}
