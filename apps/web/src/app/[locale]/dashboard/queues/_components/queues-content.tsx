"use client";

import { useCallback, useEffect, useState } from "react";
import { toast } from "sonner";
import { useDashboardShell } from "~/app/[locale]/dashboard/_components/dashboard-shell-context";
import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from "~/components/ui/select";
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
	branchesApi,
	queuesApi,
	type Branch,
	type Queue,
	type QueueStatsResponse,
} from "~/lib/api/qms";

export function QueuesContent() {
	const { currentOrganization } = useDashboardShell();
	const [queues, setQueues] = useState<Queue[]>([]);
	const [branches, setBranches] = useState<Branch[]>([]);
	const [selectedBranchId, setSelectedBranchId] = useState<string>("");
	const [isLoading, setIsLoading] = useState(true);
	const [error, setError] = useState<any>(null);
	const [stats, setStats] = useState<QueueStatsResponse | null>(null);

	const [registerOpen, setRegisterOpen] = useState(false);
	const [forwardOpen, setForwardOpen] = useState(false);
	const [detailOpen, setDetailOpen] = useState(false);

	const [selectedQueue, setSelectedQueue] = useState<Queue | null>(null);

	const fetchBranches = useCallback(async () => {
		if (!currentOrganization) return;
		try {
			const resp = await branchesApi.getAll();
			const nextBranches = resp.data || [];
			setBranches(nextBranches);
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

	const fetchQueues = useCallback(async () => {
		if (!currentOrganization || !selectedBranchId) {
			setQueues([]);
			setIsLoading(false);
			return;
		}
		setIsLoading(true);
		setError(null);
		try {
			const resp = await queuesApi.getAll({ branch_id: selectedBranchId });
			setQueues(resp.data || []);
		} catch (err: any) {
			setError(err);
		} finally {
			setIsLoading(false);
		}
	}, [currentOrganization, selectedBranchId]);

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

	useEffect(() => {
		fetchBranches();
	}, [fetchBranches]);

	useEffect(() => {
		fetchQueues();
	}, [fetchQueues]);

	useEffect(() => {
		fetchStats();
	}, [fetchStats]);

	const handleRegister = () => {
		setRegisterOpen(true);
	};

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
			fetchQueues();
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
					<Button onClick={handleRegister} disabled={!selectedBranchId}>
						<Icon name="Plus" className="mr-2 h-4 w-4" />
						Register Queue
					</Button>
				</div>
			</div>

			{statCards.length > 0 && (
				<div className="grid gap-4 md:grid-cols-4">
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
				</div>
			)}

			<QueueTable
				queues={queues}
				isLoading={isLoading}
				error={error}
				canTransition
				canForward
				onTransition={handleTransition}
				onForward={handleForward}
				onView={handleView}
				onRegister={handleRegister}
			/>

			<QueueRegisterDialog
				open={registerOpen}
				onOpenChange={setRegisterOpen}
				branches={branches}
				defaultBranchId={selectedBranchId}
				onSuccess={() => {
					fetchQueues();
					fetchStats();
				}}
			/>

			<QueueForwardDialog
				open={forwardOpen}
				onOpenChange={setForwardOpen}
				queue={selectedQueue}
				onSuccess={() => {
					fetchQueues();
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
