"use client";

import { Icon } from "~/components/shared/icon";
import { useDashboardShell } from "../_components/dashboard-shell-context";
import { QueueSettingsContent } from "./_components/queue-settings-content";

export default function QueueSettingsPage() {
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
			<QueueSettingsContent />
		</div>
	);
}
