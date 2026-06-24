"use client";

import { useCallback, useEffect, useState } from "react";
import { toast } from "sonner";
import { useDashboardShell } from "~/app/[locale]/dashboard/_components/dashboard-shell-context";
import { CounterTable, CounterDialog } from "~/components/dashboard/counters";
import { Icon } from "~/components/shared/icon";
import { Button } from "~/components/ui/button";
import {
	countersApi,
	branchesApi,
	type Counter,
	type Branch,
} from "~/lib/api/qms";

export function CountersContent() {
	const { currentOrganization } = useDashboardShell();
	const [counters, setCounters] = useState<Counter[]>([]);
	const [branches, setBranches] = useState<Branch[]>([]);
	const [isLoading, setIsLoading] = useState(true);
	const [error, setError] = useState<any>(null);
	const [dialogOpen, setDialogOpen] = useState(false);
	const [selectedCounter, setSelectedCounter] = useState<Counter | null>(null);

	const fetchData = useCallback(async () => {
		if (!currentOrganization) return;
		setIsLoading(true);
		setError(null);
		try {
			const [countersResp, branchesResp] = await Promise.all([
				countersApi.getAll(),
				branchesApi.getAll(),
			]);
			setCounters(countersResp.data || []);
			setBranches(branchesResp.data || []);
		} catch (err: any) {
			setError(err);
		} finally {
			setIsLoading(false);
		}
	}, [currentOrganization]);

	useEffect(() => {
		fetchData();
	}, [fetchData]);

	const handleCreate = () => {
		setSelectedCounter(null);
		setDialogOpen(true);
	};

	const handleEdit = (counter: Counter) => {
		setSelectedCounter(counter);
		setDialogOpen(true);
	};

	const handleDelete = async (counter: Counter) => {
		try {
			await countersApi.delete(counter.id);
			toast.success(`Counter "${counter.code}" deleted`);
			fetchData();
		} catch (err: any) {
			toast.error(err.message || "Failed to delete counter");
		}
	};

	if (!currentOrganization) return null;

	return (
		<>
			<div className="flex items-center justify-between">
				<div>
					<h2 className="text-2xl font-bold tracking-tight">Counters</h2>
					<p className="text-muted-foreground">
						Manage counters across all branches.
					</p>
				</div>
				<Button onClick={handleCreate}>
					<Icon name="Plus" className="mr-2 h-4 w-4" />
					Add Counter
				</Button>
			</div>

			<CounterTable
				counters={counters}
				isLoading={isLoading}
				error={error}
				canUpdate
				canDelete
				onEdit={handleEdit}
				onDelete={handleDelete}
				onCreateCounter={handleCreate}
			/>

			<CounterDialog
				open={dialogOpen}
				onOpenChange={setDialogOpen}
				counter={selectedCounter}
				branches={branches}
				onSuccess={fetchData}
			/>
		</>
	);
}
