"use client";

import { useCallback, useEffect, useState } from "react";
import { toast } from "sonner";
import { useDashboardShell } from "~/app/[locale]/dashboard/_components/dashboard-shell-context";
import { ServiceDialog, ServiceTable } from "~/components/dashboard/services";
import { Icon } from "~/components/shared/icon";
import { Button } from "~/components/ui/button";
import { type Service, servicesApi } from "~/lib/api/qms";

export function ServicesContent() {
	const { currentOrganization } = useDashboardShell();
	const [services, setServices] = useState<Service[]>([]);
	const [isLoading, setIsLoading] = useState(true);
	const [error, setError] = useState<any>(null);
	const [dialogOpen, setDialogOpen] = useState(false);
	const [selectedService, setSelectedService] = useState<Service | null>(null);

	const fetchServices = useCallback(async () => {
		if (!currentOrganization) return;
		setIsLoading(true);
		setError(null);
		try {
			const resp = await servicesApi.getAll();
			setServices(resp.data || []);
		} catch (err: any) {
			setError(err);
		} finally {
			setIsLoading(false);
		}
	}, [currentOrganization]);

	useEffect(() => {
		fetchServices();
	}, [fetchServices]);

	const handleCreate = () => {
		setSelectedService(null);
		setDialogOpen(true);
	};

	const handleEdit = (service: Service) => {
		setSelectedService(service);
		setDialogOpen(true);
	};

	const handleDelete = async (service: Service) => {
		try {
			await servicesApi.delete(service.id);
			toast.success(`Service "${service.code}" deleted`);
			fetchServices();
		} catch (err: any) {
			toast.error(err.message || "Failed to delete service");
		}
	};

	if (!currentOrganization) return null;

	return (
		<>
			<div className="flex items-center justify-between">
				<div>
					<h2 className="text-2xl font-bold tracking-tight">Services</h2>
					<p className="text-muted-foreground">
						Manage queue service flows for this organization.
					</p>
				</div>
				<Button onClick={handleCreate}>
					<Icon name="Plus" className="mr-2 h-4 w-4" />
					Add Service
				</Button>
			</div>

			<ServiceTable
				services={services}
				isLoading={isLoading}
				error={error}
				canUpdate
				canDelete
				onEdit={handleEdit}
				onDelete={handleDelete}
				onCreateService={handleCreate}
			/>

			<ServiceDialog
				open={dialogOpen}
				onOpenChange={setDialogOpen}
				service={selectedService}
				onSuccess={fetchServices}
			/>
		</>
	);
}
