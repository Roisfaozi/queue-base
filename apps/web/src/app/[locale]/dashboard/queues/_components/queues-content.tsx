"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import { toast } from "sonner";
import { useDashboardShell } from "~/app/[locale]/dashboard/_components/dashboard-shell-context";
import {
	QueueTable,
	QueueRegisterDialog,
	QueueDetailSheet,
	QueueForwardDialog,
} from "~/components/dashboard/queues";
import { Icon } from "~/components/shared/icon";
import { Button } from "~/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "~/components/ui/card";
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
	queuesApi,
	servicesApi,
	type Branch,
	type Counter,
	type Queue,
	type QueueJourney,
	type QueueStatsResponse,
	type Service,
} from "~/lib/api/qms";

type ViewMode = "queues" | "service" | "counter";

const STATUS_FILTERS = [
	"all",
	"waiting",
	"calling",
	"serving",
	"skipped",
	"canceled",
	"completed",
] as const;

export function QueuesContent() {
	const { currentOrganization } = useDashboardShell();
	const [queues, setQueues] = useState<Queue[]>([]);
	const [journeys, setJourneys] = useState<QueueJourney[]>([]);
	const [branches, setBranches] = useState<Branch[]>([]);
	const [services, setServices] = useState<Service[]>([]);
	const [counters, setCounters] = useState<Counter[]>([]);
	const [selectedBranchId, setSelectedBranchId] = useState<string>("");
	const [selectedServiceId, setSelectedServiceId] = useState<string>("all");
	const [selectedCounterId, setSelectedCounterId] = useState<string>("all");
	const [selectedStatus, setSelectedStatus] = useState<string>("all");
	const [viewMode, setViewMode] = useState<ViewMode>("queues");
	const [isLoading, setIsLoading] = useState(true);
	const [error, setError] = useState<any>(null);
	const [stats, setStats] = useState<QueueStatsResponse | null>(null);

	const [registerOpen, setRegisterOpen] = useState(false);
	const [forwardOpen, setForwardOpen] = useState(false);
	const [detailOpen, setDetailOpen] = useState(false);
	const [selectedQueue, setSelectedQueue] = useState<Queue | null>(null);

	const fetchReferences = useCallback(async () => {
		if (!currentOrganization) return;
		try {
			const [branchResp, serviceResp, counterResp] = await Promise.all([
				branchesApi.getAll(),
				servicesApi.getAll(),
				countersApi.getAll(),
			]);
			const nextBranches = branchResp.data || [];
			setBranches(nextBranches);
			setServices(
				(serviceResp.data || []).filter(
					(service) => service.status === "active",
				),
			);
			setCounters(
				(counterResp.data || []).filter(
					(counter) => counter.status === "active",
				),
			);
			setSelectedBranchId((current) => {
				if (current && nextBranches.some((branch) => branch.id === current)) {
					return current;
				}
				return nextBranches[0]?.id || "";
			});
		} catch (err: any) {
			setError(err);
		}
	}, [currentOrganization]);

	const fetchStats = useCallback(async () => {
		if (!currentOrganization || !selectedBranchId) {
			setStats(null);
			return;
		}
		try {
			const resp = await queuesApi.getQueueStats(selectedBranchId);
			setStats(resp.data);
		} catch {
			setStats(null);
		}
	}, [currentOrganization, selectedBranchId]);

	const fetchViewData = useCallback(async () => {
		if (!currentOrganization || !selectedBranchId) {
			setQueues([]);
			setJourneys([]);
			setIsLoading(false);
			return;
		}

		setIsLoading(true);
		setError(null);

		try {
			if (viewMode === "queues") {
				const resp = await queuesApi.getAll({
					branch_id: selectedBranchId,
					status: selectedStatus !== "all" ? selectedStatus : undefined,
					service_id:
						selectedServiceId !== "all" ? selectedServiceId : undefined,
				});
				setQueues(resp.data || []);
				setJourneys([]);
				return;
			}

			if (viewMode === "service") {
				if (selectedServiceId === "all") {
					setJourneys([]);
					setQueues([]);
					return;
				}
				const resp = await queuesApi.getJourneysByBranchAndService(
					selectedBranchId,
					selectedServiceId,
					{ status: selectedStatus !== "all" ? selectedStatus : undefined },
				);
				setJourneys(resp.data || []);
				setQueues([]);
				return;
			}

			if (selectedCounterId === "all") {
				setJourneys([]);
				setQueues([]);
				return;
			}

			const resp = await queuesApi.getJourneysByBranchAndCounter(
				selectedBranchId,
				selectedCounterId,
				{ status: selectedStatus !== "all" ? selectedStatus : undefined },
			);
			setJourneys(resp.data || []);
			setQueues([]);
		} catch (err: any) {
			setError(err);
		} finally {
			setIsLoading(false);
		}
	}, [
		currentOrganization,
		selectedBranchId,
		selectedCounterId,
		selectedServiceId,
		selectedStatus,
		viewMode,
	]);

	useEffect(() => {
		fetchReferences();
	}, [fetchReferences]);

	useEffect(() => {
		fetchStats();
	}, [fetchStats]);

	useEffect(() => {
		fetchViewData();
	}, [fetchViewData]);

	const handleView = (queue: Queue) => {
		setSelectedQueue(queue);
		setDetailOpen(true);
	};

	const handleForward = (queue: Queue) => {
		setSelectedQueue(queue);
		setForwardOpen(true);
	};

	const handleTransition = async (
		queue: Queue,
		action: "call" | "serve" | "complete" | "skip" | "cancel",
	) => {
		try {
			await queuesApi.transition(queue.id, { action });
			toast.success(`Ticket ${queue.ticket_no} marked as ${action}`);
			fetchViewData();
			fetchStats();
		} catch (err: any) {
			toast.error(err.message || `Failed to ${action} queue`);
		}
	};

	const statCards = stats
		? [
				{
					label: "Total Today",
					value: stats.total_queues_today,
					icon: "ListOrdered",
				},
				{
					label: "Active Journeys",
					value: stats.total_active_journeys,
					icon: "Activity",
				},
				{
					label: "Completed Visits",
					value: stats.total_completed_visits,
					icon: "CheckCircle2",
				},
				{
					label: "Waiting Services",
					value: Object.values(stats.waiting_by_service || {}).reduce(
						(total, count) => total + count,
						0,
					),
					icon: "Clock3",
				},
			]
		: [];

	const waitingByServiceEntries = useMemo(() => {
		if (!stats?.waiting_by_service) return [];
		return Object.entries(stats.waiting_by_service)
			.filter(([, count]) => count > 0)
			.map(([serviceId, count]) => {
				const service = services.find((item) => item.id === serviceId);
				return {
					serviceId,
					count,
					label: service ? `${service.code} — ${service.name}` : serviceId,
				};
			});
	}, [services, stats?.waiting_by_service]);

	const selectedService = services.find(
		(service) => service.id === selectedServiceId,
	);
	const selectedCounter = counters.find(
		(counter) => counter.id === selectedCounterId,
	);

	if (!currentOrganization) return null;

	return (
		<>
			<div className="flex items-center justify-between">
				<div>
					<h2 className="text-2xl font-bold tracking-tight">
						Queues Dashboard
					</h2>
					<p className="text-muted-foreground">
						Live monitoring and management of branch queues.
					</p>
				</div>
				<div className="flex items-center gap-3">
					<div className="min-w-[220px]">
						<Select
							value={selectedBranchId}
							onValueChange={setSelectedBranchId}
						>
							<SelectTrigger>
								<SelectValue placeholder="Select branch" />
							</SelectTrigger>
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
					</div>
					<Button
						onClick={() => setRegisterOpen(true)}
						disabled={!selectedBranchId}
					>
						<Icon name="Plus" className="mr-2 h-4 w-4" />
						Register Queue
					</Button>
				</div>
			</div>

			{statCards.length > 0 && (
				<div className="grid gap-4 xl:grid-cols-5 md:grid-cols-2">
					{statCards.map((card) => (
						<Card key={card.label}>
							<CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
								<CardTitle className="text-muted-foreground text-sm font-medium">
									{card.label}
								</CardTitle>
								<Icon
									name={card.icon as any}
									className="text-muted-foreground h-4 w-4"
								/>
							</CardHeader>
							<CardContent>
								<div className="text-3xl font-bold">{card.value}</div>
							</CardContent>
						</Card>
					))}
					<Card>
						<CardHeader className="pb-2">
							<CardTitle className="text-muted-foreground text-sm font-medium">
								Waiting By Service
							</CardTitle>
						</CardHeader>
						<CardContent>
							<div className="space-y-2 text-sm">
								{waitingByServiceEntries.length === 0 ? (
									<p className="text-muted-foreground">
										No waiting service queues.
									</p>
								) : (
									waitingByServiceEntries.map((entry) => (
										<div
											key={entry.serviceId}
											className="flex items-center justify-between gap-3"
										>
											<span className="truncate text-muted-foreground">
												{entry.label}
											</span>
											<span className="font-semibold">{entry.count}</span>
										</div>
									))
								)}
							</div>
						</CardContent>
					</Card>
				</div>
			)}

			<div className="grid gap-4 rounded-lg border p-4 md:grid-cols-4 xl:grid-cols-5">
				<div>
					<p className="mb-2 text-sm font-medium">View</p>
					<Select
						value={viewMode}
						onValueChange={(value) => setViewMode(value as ViewMode)}
					>
						<SelectTrigger>
							<SelectValue placeholder="Select view" />
						</SelectTrigger>
						<SelectContent>
							<SelectItem value="queues">Queues</SelectItem>
							<SelectItem value="service">Caller By Service</SelectItem>
							<SelectItem value="counter">Caller By Counter</SelectItem>
						</SelectContent>
					</Select>
				</div>
				<div>
					<p className="mb-2 text-sm font-medium">Status</p>
					<Select value={selectedStatus} onValueChange={setSelectedStatus}>
						<SelectTrigger>
							<SelectValue placeholder="All statuses" />
						</SelectTrigger>
						<SelectContent>
							{STATUS_FILTERS.map((status) => (
								<SelectItem key={status} value={status}>
									{status === "all" ? "All statuses" : status}
								</SelectItem>
							))}
						</SelectContent>
					</Select>
				</div>
				<div>
					<p className="mb-2 text-sm font-medium">Service</p>
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
				<div>
					<p className="mb-2 text-sm font-medium">Counter</p>
					<Select
						value={selectedCounterId}
						onValueChange={setSelectedCounterId}
					>
						<SelectTrigger>
							<SelectValue placeholder="All counters" />
						</SelectTrigger>
						<SelectContent>
							<SelectItem value="all">All counters</SelectItem>
							{counters
								.filter(
									(counter) =>
										!selectedBranchId || counter.branch_id === selectedBranchId,
								)
								.map((counter) => (
									<SelectItem key={counter.id} value={counter.id}>
										{counter.code} — {counter.name}
									</SelectItem>
								))}
						</SelectContent>
					</Select>
				</div>
				<div className="flex items-end">
					<div className="text-sm text-muted-foreground">
						{viewMode === "service" && selectedService
							? `Showing active journeys for ${selectedService.code}.`
							: viewMode === "counter" && selectedCounter
								? `Showing active journeys for ${selectedCounter.code}.`
								: "Showing queue master list for selected branch."}
					</div>
				</div>
			</div>

			{viewMode === "queues" ? (
				<QueueTable
					queues={queues}
					isLoading={isLoading}
					error={error}
					canTransition
					canForward
					onTransition={handleTransition}
					onForward={handleForward}
					onView={handleView}
					onRegister={() => setRegisterOpen(true)}
				/>
			) : (
				<Card>
					<CardHeader>
						<CardTitle>
							{viewMode === "service"
								? "Active Journeys By Service"
								: "Active Journeys By Counter"}
						</CardTitle>
					</CardHeader>
					<CardContent>
						{isLoading ? (
							<div className="flex items-center gap-2 text-muted-foreground">
								<Icon name="Loader" className="h-4 w-4 animate-spin" />
								Loading journeys...
							</div>
						) : error ? (
							<p className="text-sm text-destructive">
								{error.message || "Failed to load journeys"}
							</p>
						) : journeys.length === 0 ? (
							<p className="text-sm text-muted-foreground">
								No journeys found for current filters.
							</p>
						) : (
							<div className="space-y-3">
								{journeys.map((journey) => (
									<div
										key={journey.id}
										className="flex items-center justify-between rounded-md border p-3"
									>
										<div>
											<p className="font-medium">
												Queue {journey.queue_id.slice(0, 8)}…
											</p>
											<p className="text-sm text-muted-foreground">
												Seq #{journey.seq_no} • Service{" "}
												{journey.service_id.slice(0, 8)}…
												{journey.counter_id
													? ` • Counter ${journey.counter_id.slice(0, 8)}…`
													: ""}
											</p>
										</div>
										<div className="text-right">
											<p className="font-semibold capitalize">
												{journey.status}
											</p>
										</div>
									</div>
								))}
							</div>
						)}
					</CardContent>
				</Card>
			)}

			<QueueRegisterDialog
				open={registerOpen}
				onOpenChange={setRegisterOpen}
				branches={branches}
				defaultBranchId={selectedBranchId}
				onSuccess={() => {
					fetchViewData();
					fetchStats();
				}}
			/>

			<QueueForwardDialog
				open={forwardOpen}
				onOpenChange={setForwardOpen}
				queue={selectedQueue}
				onSuccess={() => {
					fetchViewData();
					fetchStats();
				}}
			/>

			<QueueDetailSheet
				open={detailOpen}
				onOpenChange={setDetailOpen}
				queue={selectedQueue}
			/>
		</>
	);
}
