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
import { countersApi, type Counter, type Branch } from "~/lib/api/qms";

const counterSchema = z.object({
	branch_id: z.string().min(1, "Branch is required."),
	code: z.string().min(2, "Code must be at least 2 characters.").max(50),
	name: z.string().min(3, "Name must be at least 3 characters.").max(255),
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
	const isEdit = !!counter;

	const form = useForm<CounterFormValues>({
		resolver: zodResolver(counterSchema),
		defaultValues: {
			branch_id: "",
			code: "",
			name: "",
			status: "active",
		},
	});

	useEffect(() => {
		if (open) {
			form.reset({
				branch_id: counter?.branch_id || "",
				code: counter?.code || "",
				name: counter?.name || "",
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
					status: data.status,
				});
				toast.success("Counter updated successfully");
			} else {
				await countersApi.create({
					branch_id: data.branch_id,
					code: data.code,
					name: data.name,
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
