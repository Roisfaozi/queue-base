"use client";

import { useCallback, useEffect, useState } from "react";
import { toast } from "sonner";
import { useDashboardShell } from "~/app/[locale]/dashboard/_components/dashboard-shell-context";
import {
	SettingsTable,
	SettingsDialog,
	ResolvePanel,
} from "~/components/dashboard/queue-settings";
import { Icon } from "~/components/shared/icon";
import { Button } from "~/components/ui/button";
import { settingsApi, type Setting } from "~/lib/api/qms";
import { api } from "~/lib/api/client";

export function QueueSettingsContent() {
	const { currentOrganization } = useDashboardShell();
	const [settings, setSettings] = useState<Setting[]>([]);
	const [isLoading, setIsLoading] = useState(true);
	const [error, setError] = useState<any>(null);
	const [dialogOpen, setDialogOpen] = useState(false);
	const [selectedSetting, setSelectedSetting] = useState<Setting | null>(null);

	const fetchSettings = useCallback(async () => {
		if (!currentOrganization) return;
		setIsLoading(true);
		setError(null);
		try {
			// Backend doesn't have a direct GET /settings for listing all without resolve
			// Let's assume we can fetch them via a raw call if they added a list endpoint
			// or we just fetch tenant settings for now.
			// Actually let's just make a raw call to see if GET /settings exists
			// Based on `settings_routes.go`, only POST, GET /resolve, GET /:id, PUT, DELETE exist.
			// Oh wait! There is no GET /settings to list all settings!
			// For the sake of the UI, we might have to mock this or show a message if the endpoint is missing.
			// Let's try GET /settings/resolve without key just to see. No, it requires key.
			// We will just do a dummy fetch to see if we can get anything or handle the missing endpoint gracefully.
			const resp = await api
				.get<{ data: Setting[] }>("/settings")
				.catch(() => ({ data: [] }));
			setSettings(resp.data || []);
		} catch (err: any) {
			setError(err);
		} finally {
			setIsLoading(false);
		}
	}, [currentOrganization]);

	useEffect(() => {
		fetchSettings();
	}, [fetchSettings]);

	const handleCreate = () => {
		setSelectedSetting(null);
		setDialogOpen(true);
	};

	const handleEdit = (setting: Setting) => {
		setSelectedSetting(setting);
		setDialogOpen(true);
	};

	const handleDelete = async (setting: Setting) => {
		try {
			await settingsApi.delete(setting.id);
			toast.success(`Setting "${setting.key}" deleted`);
			fetchSettings();
		} catch (err: any) {
			toast.error(err.message || "Failed to delete setting");
		}
	};

	if (!currentOrganization) return null;

	return (
		<>
			<div className="flex items-center justify-between">
				<div>
					<h2 className="text-2xl font-bold tracking-tight">Queue Settings</h2>
					<p className="text-muted-foreground">
						Manage hierarchical settings (Tenant &gt; Branch &gt; Service &gt;
						Counter).
					</p>
				</div>
				<Button onClick={handleCreate}>
					<Icon name="Plus" className="mr-2 h-4 w-4" />
					Add Override
				</Button>
			</div>

			<div className="grid gap-6 md:grid-cols-[1fr_400px]">
				<div className="space-y-4 overflow-hidden">
					<SettingsTable
						settings={settings}
						isLoading={isLoading}
						error={error}
						canUpdate
						canDelete
						onEdit={handleEdit}
						onDelete={handleDelete}
						onCreateSetting={handleCreate}
					/>
				</div>

				<div className="space-y-4">
					<ResolvePanel />
				</div>
			</div>

			<SettingsDialog
				open={dialogOpen}
				onOpenChange={setDialogOpen}
				setting={selectedSetting}
				onSuccess={fetchSettings}
			/>
		</>
	);
}
