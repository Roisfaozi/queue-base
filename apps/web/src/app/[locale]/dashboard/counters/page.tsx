"use client";

import { Icon } from "~/components/shared/icon";
import { useDashboardShell } from "../_components/dashboard-shell-context";
import { Button } from "~/components/ui/button";

export default function CountersPage() {
	const { currentOrganization, isLoading } = useDashboardShell();

	if (isLoading) {
		return (
			<div className="flex h-[400px] items-center justify-center rounded-lg border-2 border-dashed">
				<p className="text-muted-foreground">Loading organization context...</p>
			</div>
		);
	}

	if (!currentOrganization) {
		return (
			<div className="flex h-[400px] items-center justify-center rounded-lg border-2 border-dashed">
				<div className="text-center">
					<Icon
						name="Building2"
						className="text-muted-foreground/50 mx-auto h-8 w-8"
					/>
					<p className="text-muted-foreground mt-2">
						Please select an organization first.
					</p>
				</div>
			</div>
		);
	}

	return (
		<div className="space-y-6">
			<div className="flex items-center justify-between">
				<div>
					<h2 className="text-2xl font-bold tracking-tight">Counters</h2>
					<p className="text-muted-foreground">
						Manage counters across branches for this organization.
					</p>
				</div>
				<Button>
					<Icon name="Plus" className="mr-2 h-4 w-4" />
					Add Counter
				</Button>
			</div>

			<div className="rounded-md border">
				<div className="p-4 text-center text-sm text-muted-foreground">
					Counters table will go here.
				</div>
			</div>
		</div>
	);
}
