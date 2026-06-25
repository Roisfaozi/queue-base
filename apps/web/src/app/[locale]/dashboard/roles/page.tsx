"use client";

import { Icon } from "~/components/shared/icon";
import { useDashboardShell } from "../_components/dashboard-shell-context";
import { RolesProvider } from "./_components/roles-context";
import { RolesGrid } from "./_components/roles-grid";
import { RolesHeader } from "./_components/roles-header";
import { RolesModals } from "./_components/roles-modals";

export default function RolesPage() {
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
		<RolesProvider>
			<div className="space-y-6">
				<RolesHeader />
				<RolesGrid />
				<RolesModals />
			</div>
		</RolesProvider>
	);
}
