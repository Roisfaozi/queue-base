"use client";

import { Icon } from "~/components/shared/icon";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "~/components/ui/tabs";
import { useDashboardShell } from "../_components/dashboard-shell-context";
import { AccessRightsProvider } from "./_components/access-rights-context";
import { AccessRightsList } from "./_components/access-rights-list";
import { CreateArDialog } from "./_components/create-ar-dialog";
import { EndpointsList } from "./_components/endpoints-list";
import { RegisterEpDialog } from "./_components/register-ep-dialog";

export default function AccessRightsPage() {
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
		<AccessRightsProvider>
			<div className="space-y-6">
				<div className="flex items-center justify-between">
					<div>
						<h2 className="text-2xl font-bold tracking-tight">
							Access Rights & Endpoints
						</h2>
						<p className="text-muted-foreground">
							Define resource groups and register API endpoints.
						</p>
					</div>
				</div>

				<Tabs defaultValue="access-rights" className="w-full">
					<TabsList className="grid w-full max-w-[400px] grid-cols-2">
						<TabsTrigger value="access-rights">Access Rights</TabsTrigger>
						<TabsTrigger value="endpoints">All Endpoints</TabsTrigger>
					</TabsList>

					<TabsContent value="access-rights" className="mt-4 space-y-4">
						<div className="flex justify-end">
							<CreateArDialog />
						</div>
						<AccessRightsList />
					</TabsContent>

					<TabsContent value="endpoints" className="mt-4 space-y-4">
						<div className="flex justify-end">
							<RegisterEpDialog />
						</div>
						<EndpointsList />
					</TabsContent>
				</Tabs>
			</div>
		</AccessRightsProvider>
	);
}
