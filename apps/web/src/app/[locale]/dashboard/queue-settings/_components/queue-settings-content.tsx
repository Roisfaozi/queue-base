"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import { toast } from "sonner";
import { useDashboardShell } from "~/app/[locale]/dashboard/_components/dashboard-shell-context";
import {
	ResolvePanel,
	SettingsDialog,
} from "~/components/dashboard/queue-settings";
import { Icon } from "~/components/shared/icon";
import { Badge } from "~/components/ui/badge";
import { Button } from "~/components/ui/button";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "~/components/ui/card";
import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from "~/components/ui/select";
import {
	branchesApi,
	countersApi,
	servicesApi,
	settingsApi,
	type Branch,
	type Counter,
	type EffectiveQueueConfig,
	type Service,
} from "~/lib/api/qms";

const EFFECTIVE_FIELDS: Array<{
	key: keyof EffectiveQueueConfig;
	label: string;
	description: string;
}> = [
	{
		key: "queue_reset_time",
		label: "Queue Reset Time",
		description: "Batas pergantian business day untuk ticket harian.",
	},
	{
		key: "ticket_prefix",
		label: "Ticket Prefix",
		description: "Prefix ticket efektif untuk konteks yang dipilih.",
	},
	{
		key: "numbering_strategy",
		label: "Numbering Strategy",
		description: "Strategi penomoran queue efektif.",
	},
	{
		key: "default_estimated_duration",
		label: "Default Estimated Duration",
		description: "Estimasi default durasi layanan bila tersedia.",
	},
];

export function QueueSettingsContent() {
	const { currentOrganization } = useDashboardShell();
	const [branches, setBranches] = useState<Branch[]>([]);
	const [services, setServices] = useState<Service[]>([]);
	const [counters, setCounters] = useState<Counter[]>([]);
	const [effectiveConfig, setEffectiveConfig] =
		useState<EffectiveQueueConfig | null>(null);
	const [isLoading, setIsLoading] = useState(true);
	const [isRefreshing, setIsRefreshing] = useState(false);
	const [error, setError] = useState<string | null>(null);
	const [dialogOpen, setDialogOpen] = useState(false);
	const [selectedBranchId, setSelectedBranchId] = useState<string>("all");
	const [selectedServiceId, setSelectedServiceId] = useState<string>("all");
	const [selectedCounterId, setSelectedCounterId] = useState<string>("all");

	const selectedBranch =
		selectedBranchId === "all" ? undefined : selectedBranchId;
	const selectedService =
		selectedServiceId === "all" ? undefined : selectedServiceId;
	const selectedCounter =
		selectedCounterId === "all" ? undefined : selectedCounterId;

	const availableCounters = useMemo(() => {
		if (!selectedBranch) return counters;
		return counters.filter((counter) => counter.branch_id === selectedBranch);
	}, [counters, selectedBranch]);

	const fetchEffectiveConfig = useCallback(async () => {
		if (!currentOrganization) return;
		setIsRefreshing(true);
		setError(null);
		try {
			const response = await settingsApi.getEffective({
				branch_id: selectedBranch,
				service_id: selectedService,
				counter_id: selectedCounter,
			});
			setEffectiveConfig(response.data);
		} catch (err: any) {
			setError(err?.message || "Failed to load effective queue config");
			setEffectiveConfig(null);
		} finally {
			setIsRefreshing(false);
		}
	}, [currentOrganization, selectedBranch, selectedCounter, selectedService]);

	const fetchInitialData = useCallback(async () => {
		if (!currentOrganization) return;
		setIsLoading(true);
		setError(null);
		try {
			const [
				branchResponse,
				serviceResponse,
				counterResponse,
				effectiveResponse,
			] = await Promise.all([
				branchesApi.getAll(),
				servicesApi.getAll(),
				countersApi.getAll(),
				settingsApi.getEffective(),
			]);
			setBranches(branchResponse.data || []);
			setServices(serviceResponse.data || []);
			setCounters(counterResponse.data || []);
			setEffectiveConfig(effectiveResponse.data);
		} catch (err: any) {
			setError(err?.message || "Failed to load queue settings context");
			setEffectiveConfig(null);
		} finally {
			setIsLoading(false);
			setIsRefreshing(false);
		}
	}, [currentOrganization]);

	useEffect(() => {
		fetchInitialData();
	}, [fetchInitialData]);

	useEffect(() => {
		if (!currentOrganization || isLoading) return;
		fetchEffectiveConfig();
	}, [currentOrganization, fetchEffectiveConfig, isLoading]);

	useEffect(() => {
		if (
			selectedCounter !== undefined &&
			!availableCounters.some((counter) => counter.id === selectedCounter)
		) {
			setSelectedCounterId("all");
		}
	}, [availableCounters, selectedCounter]);

	const handleCreate = () => {
		setDialogOpen(true);
	};

	const handleDialogSuccess = async () => {
		toast.success("Queue override saved");
		await fetchInitialData();
	};

	if (!currentOrganization) return null;

	return (
		<>
			<div className="flex items-center justify-between gap-4">
				<div>
					<h2 className="text-2xl font-bold tracking-tight">Queue Settings</h2>
					<p className="text-muted-foreground">
						Typed configuration runtime for tenant, branch, service, and counter
						scope.
					</p>
				</div>
				<div className="flex items-center gap-2">
					<Button
						variant="outline"
						onClick={() => void fetchEffectiveConfig()}
						disabled={isLoading || isRefreshing}
					>
						{isRefreshing ? (
							<Icon name="Loader" className="mr-2 h-4 w-4 animate-spin" />
						) : (
							<Icon name="RefreshCw" className="mr-2 h-4 w-4" />
						)}
						Refresh Effective
					</Button>
					<Button onClick={handleCreate}>
						<Icon name="Plus" className="mr-2 h-4 w-4" />
						Add Override
					</Button>
				</div>
			</div>

			<div className="grid gap-6 xl:grid-cols-[minmax(0,1fr)_400px]">
				<div className="space-y-6">
					<Card>
						<CardHeader>
							<CardTitle className="text-lg">Runtime Context</CardTitle>
							<CardDescription>
								Pilih scope untuk lihat effective config dari endpoint `GET
								/settings/effective`.
							</CardDescription>
						</CardHeader>
						<CardContent className="grid gap-4 md:grid-cols-3">
							<div className="space-y-2">
								<p className="text-sm font-medium">Branch</p>
								<Select
									value={selectedBranchId}
									onValueChange={(value) => {
										setSelectedBranchId(value);
										setSelectedCounterId("all");
									}}
								>
									<SelectTrigger>
										<SelectValue placeholder="All branches" />
									</SelectTrigger>
									<SelectContent>
										<SelectItem value="all">All branches</SelectItem>
										{branches.map((branch) => (
											<SelectItem key={branch.id} value={branch.id}>
												{branch.code} — {branch.name}
											</SelectItem>
										))}
									</SelectContent>
								</Select>
							</div>

							<div className="space-y-2">
								<p className="text-sm font-medium">Service</p>
								<Select
									value={selectedServiceId}
									onValueChange={setSelectedServiceId}
								>
									<SelectTrigger>
										<SelectValue placeholder="All services" />
									</SelectTrigger>
									<SelectContent>
										<SelectItem value="all">All services</SelectItem>
										{services.map((service) => (
											<SelectItem key={service.id} value={service.id}>
												{service.code} — {service.name}
											</SelectItem>
										))}
									</SelectContent>
								</Select>
							</div>

							<div className="space-y-2">
								<p className="text-sm font-medium">Counter</p>
								<Select
									value={selectedCounterId}
									onValueChange={setSelectedCounterId}
								>
									<SelectTrigger>
										<SelectValue placeholder="All counters" />
									</SelectTrigger>
									<SelectContent>
										<SelectItem value="all">All counters</SelectItem>
										{availableCounters.map((counter) => (
											<SelectItem key={counter.id} value={counter.id}>
												{counter.code} — {counter.name}
											</SelectItem>
										))}
									</SelectContent>
								</Select>
							</div>
						</CardContent>
					</Card>

					<Card>
						<CardHeader>
							<CardTitle className="text-lg">Effective Queue Config</CardTitle>
							<CardDescription>
								Current runtime values from typed inheritance chain.
							</CardDescription>
						</CardHeader>
						<CardContent>
							{isLoading ? (
								<div className="flex min-h-40 items-center justify-center text-sm text-muted-foreground">
									<Icon name="Loader" className="mr-2 h-4 w-4 animate-spin" />
									Loading queue config...
								</div>
							) : error ? (
								<div className="rounded-md border border-destructive/20 bg-destructive/10 p-4 text-sm text-destructive">
									{error}
								</div>
							) : effectiveConfig ? (
								<div className="grid gap-4 md:grid-cols-2">
									{EFFECTIVE_FIELDS.map((field) => {
										const value = effectiveConfig[field.key];
										return (
											<div
												key={field.key}
												className="rounded-lg border bg-muted/20 p-4"
											>
												<div className="flex items-start justify-between gap-3">
													<div>
														<p className="text-sm font-medium">{field.label}</p>
														<p className="text-muted-foreground mt-1 text-xs">
															{field.description}
														</p>
													</div>
													<Badge variant="outline">effective</Badge>
												</div>
												<p className="mt-4 break-all font-mono text-sm font-semibold">
													{value || "—"}
												</p>
											</div>
										);
									})}
								</div>
							) : (
								<div className="rounded-md border border-dashed p-8 text-center text-sm text-muted-foreground">
									No effective config returned.
								</div>
							)}
						</CardContent>
					</Card>
				</div>

				<div className="space-y-4">
					<Card>
						<CardHeader>
							<CardTitle className="text-lg">Design Notes</CardTitle>
							<CardDescription>
								Page now follows typed-config runtime, not generic settings
								list.
							</CardDescription>
						</CardHeader>
						<CardContent className="space-y-3 text-sm text-muted-foreground">
							<p>
								Core QMS config resolves from typed tenant, branch, service, and
								counter tables.
							</p>
							<p>
								Generic `settings` remains only for compatibility and manual
								overrides outside core typed flow.
							</p>
						</CardContent>
					</Card>
					<ResolvePanel />
				</div>
			</div>

			<SettingsDialog
				open={dialogOpen}
				onOpenChange={setDialogOpen}
				setting={null}
				onSuccess={() => void handleDialogSuccess()}
			/>
		</>
	);
}
