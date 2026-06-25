"use client";

import { DashboardProvider } from "./_components/dashboard-context";
import { DashboardStats } from "./_components/dashboard-stats";
import { SystemInsights } from "./_components/system-insights";
import { RecentActivity } from "./_components/recent-activity";
import { QuickActions } from "./_components/quick-actions";
import { useDashboardShell } from "./_components/dashboard-shell-context";
import { Icon } from "~/components/shared/icon";

export default function DashboardPage() {
	const { currentOrganization, isLoading } = useDashboardShell();

	if (!isLoading && !currentOrganization) {
		return (
			<div className="flex min-h-[320px] items-center justify-center rounded-[var(--radius-lg)] border border-dashed">
				<div className="max-w-md text-center">
					<Icon
						name="Building2"
						className="text-muted-foreground/60 mx-auto mb-4 h-10 w-10"
					/>
					<h2 className="text-xl font-semibold tracking-tight">
						No organization context
					</h2>
					<p className="text-muted-foreground mt-2 text-sm">
						Select organization from sidebar first. Dashboard observability data
						stays tenant-scoped.
					</p>
				</div>
			</div>
		);
	}

	return (
		<DashboardProvider>
			<div className="space-y-[var(--spacing-gap)]">
				<DashboardStats />
				<SystemInsights />

				<div className="grid gap-[var(--spacing-gap)] md:grid-cols-7">
					<RecentActivity />
					<QuickActions />
				</div>
			</div>
		</DashboardProvider>
	);
}
