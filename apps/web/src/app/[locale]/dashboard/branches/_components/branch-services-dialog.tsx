"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import { toast } from "sonner";
import { Icon } from "~/components/shared/icon";
import { Badge } from "~/components/ui/badge";
import { Button } from "~/components/ui/button";
import {
	Dialog,
	DialogContent,
	DialogDescription,
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
import { Switch } from "~/components/ui/switch";
import {
	Table,
	TableBody,
	TableCell,
	TableHead,
	TableHeader,
	TableRow,
} from "~/components/ui/table";
import {
	branchServicesApi,
	servicesApi,
	type Branch,
	type BranchService,
	type Service,
} from "~/lib/api/qms";

interface BranchServicesDialogProps {
	open: boolean;
	onOpenChange: (open: boolean) => void;
	branch: Branch | null;
	onSuccess?: () => void;
}

export function BranchServicesDialog({
	open,
	onOpenChange,
	branch,
	onSuccess,
}: BranchServicesDialogProps) {
	const [branchServices, setBranchServices] = useState<BranchService[]>([]);
	const [allServices, setAllServices] = useState<Service[]>([]);
	const [isLoading, setIsLoading] = useState(false);
	const [isSubmitting, setIsSubmitting] = useState(false);

	// New branch service form state
	const [selectedServiceId, setSelectedServiceId] = useState("");
	const [customName, setCustomName] = useState("");
	const [sortOrder, setSortOrder] = useState<number>(1);

	const fetchData = useCallback(async () => {
		if (!branch?.id) return;
		setIsLoading(true);
		try {
			const [bsResp, sResp] = await Promise.all([
				branchServicesApi.getByBranch(branch.id),
				servicesApi.getAll(),
			]);
			const bs = bsResp.data || [];
			setBranchServices(bs);
			setAllServices((sResp.data || []).filter((s) => s.status === "active"));
			setSortOrder(bs.length + 1);
		} catch (error: any) {
			toast.error(error.message || "Failed to load branch services");
		} finally {
			setIsLoading(false);
		}
	}, [branch?.id]);

	useEffect(() => {
		if (open && branch?.id) {
			fetchData();
			setSelectedServiceId("");
			setCustomName("");
		}
	}, [open, branch?.id, fetchData]);

	const linkedServiceIds = useMemo(
		() => new Set(branchServices.map((bs) => bs.service_id)),
		[branchServices],
	);

	const availableServices = useMemo(
		() => allServices.filter((s) => !linkedServiceIds.has(s.id)),
		[allServices, linkedServiceIds],
	);

	const handleLinkService = async () => {
		if (!branch?.id || !selectedServiceId) {
			toast.error("Please select a service to link");
			return;
		}

		setIsSubmitting(true);
		try {
			await branchServicesApi.create(branch.id, {
				service_id: selectedServiceId,
				custom_name: customName.trim() || undefined,
				is_active: true,
				sort_order: sortOrder || branchServices.length + 1,
			});
			toast.success("Service linked to branch");
			setSelectedServiceId("");
			setCustomName("");
			await fetchData();
			onSuccess?.();
		} catch (error: any) {
			toast.error(error.message || "Failed to link service");
		} finally {
			setIsSubmitting(false);
		}
	};

	const handleToggleActive = async (bs: BranchService) => {
		if (!branch?.id) return;
		try {
			const newActive = !bs.is_active;
			await branchServicesApi.update(branch.id, bs.id, {
				is_active: newActive,
			});
			setBranchServices((prev) =>
				prev.map((item) =>
					item.id === bs.id ? { ...item, is_active: newActive } : item,
				),
			);
			toast.success(
				`Service ${newActive ? "activated" : "deactivated"} for this branch`,
			);
			onSuccess?.();
		} catch (error: any) {
			toast.error(error.message || "Failed to toggle service status");
		}
	};

	const handleDelete = async (branchServiceId: string) => {
		if (!branch?.id) return;
		try {
			await branchServicesApi.delete(branch.id, branchServiceId);
			toast.success("Service unlinked from branch");
			await fetchData();
			onSuccess?.();
		} catch (error: any) {
			toast.error(error.message || "Failed to unlink service");
		}
	};

	if (!branch) return null;

	return (
		<Dialog open={open} onOpenChange={onOpenChange}>
			<DialogContent className="max-h-[90vh] overflow-y-auto sm:max-w-3xl">
				<DialogHeader>
					<DialogTitle className="flex items-center gap-2">
						<Icon name="Globe" className="h-5 w-5 text-primary" />
						Branch Services — {branch.code} ({branch.name})
					</DialogTitle>
					<DialogDescription>
						Manage services operationalized specifically at this branch.
					</DialogDescription>
				</DialogHeader>

				<div className="space-y-6 pt-2">
					{/* Add / Link Service Section */}
					<div className="rounded-lg border bg-muted/30 p-4">
						<h4 className="mb-3 text-sm font-semibold">Link New Service</h4>
						<div className="grid gap-3 sm:grid-cols-4">
							<div className="sm:col-span-2">
								<Label htmlFor="bs-service" className="text-xs">
									Master Service
								</Label>
								<Select
									value={selectedServiceId}
									onValueChange={setSelectedServiceId}
								>
									<SelectTrigger id="bs-service" className="mt-1">
										<SelectValue placeholder="Select active service..." />
									</SelectTrigger>
									<SelectContent>
										{availableServices.length === 0 ? (
											<SelectItem value="__none__" disabled>
												{allServices.length === 0
													? "No master services available"
													: "All active services already linked"}
											</SelectItem>
										) : (
											availableServices.map((service) => (
												<SelectItem key={service.id} value={service.id}>
													{service.code} — {service.name}
												</SelectItem>
											))
										)}
									</SelectContent>
								</Select>
							</div>
							<div>
								<Label htmlFor="bs-custom-name" className="text-xs">
									Custom Name (Opt)
								</Label>
								<Input
									id="bs-custom-name"
									placeholder="e.g. Loket 1"
									value={customName}
									onChange={(e) => setCustomName(e.target.value)}
									className="mt-1"
								/>
							</div>
							<div>
								<Label htmlFor="bs-sort-order" className="text-xs">
									Sort Order
								</Label>
								<Input
									id="bs-sort-order"
									type="number"
									min={1}
									value={sortOrder}
									onChange={(e) => setSortOrder(Number(e.target.value))}
									className="mt-1"
								/>
							</div>
						</div>
						<div className="mt-3 flex justify-end">
							<Button
								size="sm"
								onClick={handleLinkService}
								disabled={isSubmitting || !selectedServiceId}
							>
								{isSubmitting ? (
									<Icon name="Loader" className="mr-1.5 h-4 w-4 animate-spin" />
								) : (
									<Icon name="Plus" className="mr-1.5 h-4 w-4" />
								)}
								Link to Branch
							</Button>
						</div>
					</div>

					{/* List of currently linked services */}
					<div className="space-y-2">
						<div className="flex items-center justify-between">
							<h4 className="text-sm font-semibold">
								Active Branch Services ({branchServices.length})
							</h4>
							<Button
								variant="ghost"
								size="sm"
								onClick={fetchData}
								disabled={isLoading}
							>
								{isLoading ? (
									<Icon name="Loader" className="h-3.5 w-3.5 animate-spin" />
								) : (
									<Icon name="RefreshCw" className="h-3.5 w-3.5" />
								)}
							</Button>
						</div>

						{isLoading ? (
							<div className="py-8 text-center text-sm text-muted-foreground">
								Loading branch services...
							</div>
						) : branchServices.length === 0 ? (
							<div className="rounded-md border border-dashed py-8 text-center text-sm text-muted-foreground">
								No services linked to this branch yet. Link one above to start
								operations.
							</div>
						) : (
							<div className="rounded-md border">
								<Table>
									<TableHeader>
										<TableRow>
											<TableHead className="w-12">#</TableHead>
											<TableHead>Service</TableHead>
											<TableHead>Custom Name</TableHead>
											<TableHead className="w-24 text-center">Status</TableHead>
											<TableHead className="w-24 text-right">Actions</TableHead>
										</TableRow>
									</TableHeader>
									<TableBody>
										{branchServices.map((bs) => {
											const master = allServices.find(
												(s) => s.id === bs.service_id,
											);
											return (
												<TableRow key={bs.id}>
													<TableCell className="font-mono text-xs text-muted-foreground">
														{bs.sort_order}
													</TableCell>
													<TableCell>
														<div className="flex items-center gap-2">
															<span className="font-mono text-xs font-semibold">
																{master?.code || "—"}
															</span>
															<span className="text-sm">
																{master?.name || bs.service_id}
															</span>
														</div>
													</TableCell>
													<TableCell className="text-sm text-muted-foreground">
														{bs.custom_name || "—"}
													</TableCell>
													<TableCell className="text-center">
														<div className="flex items-center justify-center gap-2">
															<Switch
																checked={bs.is_active}
																onCheckedChange={() => handleToggleActive(bs)}
															/>
															<Badge
																variant={bs.is_active ? "default" : "secondary"}
																className={
																	bs.is_active
																		? "bg-emerald-500 hover:bg-emerald-600 text-xs"
																		: "text-xs"
																}
															>
																{bs.is_active ? "active" : "inactive"}
															</Badge>
														</div>
													</TableCell>
													<TableCell className="text-right">
														<Button
															variant="ghost"
															size="sm"
															className="text-destructive hover:bg-destructive/10"
															onClick={() => handleDelete(bs.id)}
														>
															<Icon name="Trash2" className="h-4 w-4" />
														</Button>
													</TableCell>
												</TableRow>
											);
										})}
									</TableBody>
								</Table>
							</div>
						)}
					</div>
				</div>
			</DialogContent>
		</Dialog>
	);
}
