"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useCallback, useEffect, useMemo, useState } from "react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";
import { useDashboardShell } from "~/app/[locale]/dashboard/_components/dashboard-shell-context";
import { Icon } from "~/components/shared/icon";
import { Button } from "~/components/ui/button";
import { Badge } from "~/components/ui/badge";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "~/components/ui/card";
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
import { Switch } from "~/components/ui/switch";
import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from "~/components/ui/select";
import { branchesApi, type Branch } from "~/lib/api/qms";
import { branchActivationSchema } from "~/lib/validations/branch";

const branchSchema = z
	.object({
		code: z.string().min(1, "Code is required.").max(50),
		name: z.string().min(1, "Name is required.").max(255),
		address: z.string().optional().default(""),
		city: z.string().optional().default(""),
		province: z.string().optional().default(""),
		postal_code: z.string().optional().default(""),
		phone: z.string().optional().default(""),
		email: z.string().email().optional().or(z.literal("")),
		logo_asset_id: z.string().optional().default(""),
		running_text: z.string().optional().default(""),
		timezone: z.string().optional().default(""),
		status: z.enum(["active", "inactive", "draft"]),
	})
	.superRefine((data, ctx) => {
		if (data.status === "active") {
			const result = branchActivationSchema.safeParse({
				address: data.address,
				city: data.city,
				province: data.province,
				phone: data.phone,
				timezone: data.timezone,
			});
			if (!result.success) {
				for (const issue of result.error.issues) {
					ctx.addIssue({
						code: z.ZodIssueCode.custom,
						path: issue.path,
						message: issue.message,
					});
				}
			}
		}
	});

type BranchFormValues = z.infer<typeof branchSchema>;

export function BranchesContent() {
	const { currentOrganization } = useDashboardShell();
	const [branches, setBranches] = useState<Branch[]>([]);
	const [isLoading, setIsLoading] = useState(true);
	const [error, setError] = useState<any>(null);
	const [dialogOpen, setDialogOpen] = useState(false);
	const [selectedBranch, setSelectedBranch] = useState<Branch | null>(null);
	const isEdit = !!selectedBranch;

	const form = useForm<BranchFormValues>({
		resolver: zodResolver(branchSchema),
		defaultValues: {
			code: "",
			name: "",
			address: "",
			city: "",
			province: "",
			postal_code: "",
			phone: "",
			email: "",
			logo_asset_id: "",
			running_text: "",
			timezone: "Asia/Jakarta",
			status: "draft",
		},
	});

	const fetchBranches = useCallback(async () => {
		if (!currentOrganization) return;
		setIsLoading(true);
		setError(null);
		try {
			const response = await branchesApi.getAll();
			setBranches(response.data || []);
		} catch (err: any) {
			setError(err);
		} finally {
			setIsLoading(false);
		}
	}, [currentOrganization]);

	useEffect(() => {
		fetchBranches();
	}, [fetchBranches]);

	useEffect(() => {
		if (dialogOpen) {
			form.reset({
				code: selectedBranch?.code || "",
				name: selectedBranch?.name || "",
				address: selectedBranch?.address || "",
				city: selectedBranch?.city || "",
				province: selectedBranch?.province || "",
				postal_code: selectedBranch?.postal_code || "",
				phone: selectedBranch?.phone || "",
				email: selectedBranch?.email || "",
				logo_asset_id: selectedBranch?.logo_asset_id || "",
				running_text: selectedBranch?.running_text || "",
				timezone: selectedBranch?.timezone || "Asia/Jakarta",
				status: selectedBranch?.status || "draft",
			});
		}
	}, [dialogOpen, selectedBranch, form]);

	const activeCount = useMemo(
		() => branches.filter((branch) => branch.status === "active").length,
		[branches],
	);

	async function onSubmit(data: BranchFormValues) {
		try {
			const payload = {
				code: data.code,
				name: data.name,
				address: data.address || undefined,
				city: data.city || undefined,
				province: data.province || undefined,
				postal_code: data.postal_code || undefined,
				phone: data.phone || undefined,
				email: data.email || undefined,
				logo_asset_id: data.logo_asset_id || undefined,
				running_text: data.running_text || undefined,
				timezone: data.timezone || undefined,
				status: data.status,
			};
			if (isEdit && selectedBranch) {
				await branchesApi.update(selectedBranch.id, payload);
				toast.success("Branch updated");
			} else {
				await branchesApi.create(payload);
				toast.success("Branch created");
			}
			setDialogOpen(false);
			setSelectedBranch(null);
			await fetchBranches();
		} catch (error: any) {
			toast.error(error.message || "Failed to save branch");
		}
	}

	async function handleDelete(branch: Branch) {
		try {
			await branchesApi.delete(branch.id);
			toast.success("Branch deleted");
			await fetchBranches();
		} catch (error: any) {
			toast.error(error.message || "Failed to delete branch");
		}
	}

	if (!currentOrganization) return null;

	return (
		<>
			<div className="flex items-center justify-between">
				<div>
					<h2 className="text-2xl font-bold tracking-tight">Branches</h2>
					<p className="text-muted-foreground">
						Manage branch profile and activation rules.
					</p>
				</div>
				<Button
					onClick={() => {
						setSelectedBranch(null);
						setDialogOpen(true);
					}}
				>
					<Icon name="Plus" className="mr-2 h-4 w-4" />
					Add Branch
				</Button>
			</div>

			<Card>
				<CardHeader>
					<CardTitle>Branch Summary</CardTitle>
					<CardDescription>
						{branches.length} branches total, {activeCount} active.
					</CardDescription>
				</CardHeader>
				<CardContent>
					{isLoading ? (
						<p className="text-muted-foreground">Loading branches...</p>
					) : error ? (
						<div className="space-y-2">
							<p className="text-destructive text-sm">
								Failed to load branches.
							</p>
							<Button variant="outline" size="sm" onClick={fetchBranches}>
								Retry
							</Button>
						</div>
					) : branches.length === 0 ? (
						<p className="text-muted-foreground">No branches found.</p>
					) : (
						<div className="space-y-3">
							{branches.map((branch) => (
								<div
									key={branch.id}
									className="flex items-center justify-between rounded-lg border p-3"
								>
									<div>
										<div className="flex items-center gap-2">
											<p className="font-medium">
												{branch.code} — {branch.name}
											</p>
											<Badge
												variant={
													branch.status === "active" ? "default" : "secondary"
												}
												className={
													branch.status === "active"
														? "bg-emerald-500 hover:bg-emerald-600"
														: ""
												}
											>
												{branch.status}
											</Badge>
										</div>
										<p className="text-muted-foreground text-sm">
											{branch.city || "—"}, {branch.province || "—"} ·{" "}
											{branch.timezone || "—"}
										</p>
									</div>
									<div className="flex items-center gap-2">
										<Button
											variant="outline"
											size="sm"
											onClick={() => {
												setSelectedBranch(branch);
												setDialogOpen(true);
											}}
										>
											Edit
										</Button>
										<Button
											variant="destructive"
											size="sm"
											onClick={() => handleDelete(branch)}
										>
											Delete
										</Button>
									</div>
								</div>
							))}
						</div>
					)}
				</CardContent>
			</Card>

			<Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
				<DialogContent className="max-h-[90vh] overflow-y-auto sm:max-w-2xl">
					<DialogHeader>
						<DialogTitle>{isEdit ? "Edit Branch" : "Add Branch"}</DialogTitle>
						<DialogDescription>
							Branch can be activated only if address, city, province, phone,
							and timezone are filled.
						</DialogDescription>
					</DialogHeader>
					<Form {...form}>
						<form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
							<div className="grid gap-4 md:grid-cols-2">
								<FormField
									control={form.control}
									name="code"
									render={({ field }) => (
										<FormItem>
											<FormLabel>Code</FormLabel>
											<FormControl>
												<Input {...field} />
											</FormControl>
											<FormMessage />
										</FormItem>
									)}
								/>
								<FormField
									control={form.control}
									name="name"
									render={({ field }) => (
										<FormItem>
											<FormLabel>Name</FormLabel>
											<FormControl>
												<Input {...field} />
											</FormControl>
											<FormMessage />
										</FormItem>
									)}
								/>
								<FormField
									control={form.control}
									name="address"
									render={({ field }) => (
										<FormItem>
											<FormLabel>Address</FormLabel>
											<FormControl>
												<Input {...field} />
											</FormControl>
											<FormMessage />
										</FormItem>
									)}
								/>
								<FormField
									control={form.control}
									name="city"
									render={({ field }) => (
										<FormItem>
											<FormLabel>City</FormLabel>
											<FormControl>
												<Input {...field} />
											</FormControl>
											<FormMessage />
										</FormItem>
									)}
								/>
								<FormField
									control={form.control}
									name="province"
									render={({ field }) => (
										<FormItem>
											<FormLabel>Province</FormLabel>
											<FormControl>
												<Input {...field} />
											</FormControl>
											<FormMessage />
										</FormItem>
									)}
								/>
								<FormField
									control={form.control}
									name="postal_code"
									render={({ field }) => (
										<FormItem>
											<FormLabel>Postal Code</FormLabel>
											<FormControl>
												<Input {...field} />
											</FormControl>
											<FormMessage />
										</FormItem>
									)}
								/>
								<FormField
									control={form.control}
									name="phone"
									render={({ field }) => (
										<FormItem>
											<FormLabel>Phone</FormLabel>
											<FormControl>
												<Input {...field} />
											</FormControl>
											<FormMessage />
										</FormItem>
									)}
								/>
								<FormField
									control={form.control}
									name="email"
									render={({ field }) => (
										<FormItem>
											<FormLabel>Email</FormLabel>
											<FormControl>
												<Input {...field} />
											</FormControl>
											<FormMessage />
										</FormItem>
									)}
								/>
								<FormField
									control={form.control}
									name="timezone"
									render={({ field }) => (
										<FormItem>
											<FormLabel>Timezone</FormLabel>
											<FormControl>
												<Input {...field} />
											</FormControl>
											<FormMessage />
										</FormItem>
									)}
								/>
								<FormField
									control={form.control}
									name="logo_asset_id"
									render={({ field }) => (
										<FormItem>
											<FormLabel>Logo Asset ID</FormLabel>
											<FormControl>
												<Input {...field} />
											</FormControl>
											<FormMessage />
										</FormItem>
									)}
								/>
								<FormField
									control={form.control}
									name="running_text"
									render={({ field }) => (
										<FormItem>
											<FormLabel>Running Text</FormLabel>
											<FormControl>
												<Input {...field} />
											</FormControl>
											<FormMessage />
										</FormItem>
									)}
								/>
								<FormField
									control={form.control}
									name="status"
									render={({ field }) => (
										<FormItem>
											<FormLabel>Status</FormLabel>
											<Select
												value={field.value}
												onValueChange={field.onChange}
											>
												<FormControl>
													<SelectTrigger>
														<SelectValue placeholder="Select status" />
													</SelectTrigger>
												</FormControl>
												<SelectContent>
													<SelectItem value="draft">Draft</SelectItem>
													<SelectItem value="active">Active</SelectItem>
													<SelectItem value="inactive">Inactive</SelectItem>
												</SelectContent>
											</Select>
											<FormMessage />
										</FormItem>
									)}
								/>
							</div>
							<DialogFooter>
								<Button
									type="button"
									variant="outline"
									onClick={() => setDialogOpen(false)}
								>
									Cancel
								</Button>
								<Button type="submit">
									{isEdit ? "Save Changes" : "Create Branch"}
								</Button>
							</DialogFooter>
						</form>
					</Form>
				</DialogContent>
			</Dialog>
		</>
	);
}
