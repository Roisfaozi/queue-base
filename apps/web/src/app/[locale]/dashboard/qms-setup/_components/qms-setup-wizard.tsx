"use client";

import Link from "next/link";
import { useEffect, useMemo, useState } from "react";
import { toast } from "sonner";
import type { QMSClientCreateResponse } from "@casbin/api-types";
import { useDashboardShell } from "~/app/[locale]/dashboard/_components/dashboard-shell-context";
import { Badge } from "~/components/ui/badge";
import { Button } from "~/components/ui/button";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "~/components/ui/card";
import {
	branchServicesApi,
	branchesApi,
	countersApi,
	qmsClientsApi,
	servicesApi,
	type Branch,
	type BranchService,
	type Counter,
	type Service,
} from "~/lib/api/qms";

const steps = [
	{
		key: "tenant",
		title: "Tenant Profile",
		href: "/dashboard/organization/settings",
	},
	{ key: "branch", title: "Branches", href: "/dashboard/branches" },
	{ key: "service", title: "Services", href: "/dashboard/services" },
	{
		key: "branch_service",
		title: "Branch Services",
		href: "/dashboard/queues",
	},
	{ key: "counter", title: "Counters", href: "/dashboard/counters" },
	{ key: "qms_client", title: "QMS Clients", href: "/dashboard/qms-clients" },
] as const;

export function QMSSetupWizard() {
	const { currentOrganization } = useDashboardShell();
	const [branches, setBranches] = useState<Branch[]>([]);
	const [services, setServices] = useState<Service[]>([]);
	const [branchServices, setBranchServices] = useState<BranchService[]>([]);
	const [counters, setCounters] = useState<Counter[]>([]);
	const [clients, setClients] = useState<QMSClientCreateResponse[]>([]);
	const [isLoading, setIsLoading] = useState(true);

	useEffect(() => {
		let mounted = true;
		async function load() {
			if (!currentOrganization) return;
			setIsLoading(true);
			try {
				const [branchResp, serviceResp, counterResp, clientResp] =
					await Promise.all([
						branchesApi.getAll(),
						servicesApi.getAll(),
						countersApi.getAll(),
						qmsClientsApi.getAll(),
					]);
				const nextBranches = branchResp.data || [];
				const branchServiceResp = await Promise.all(
					nextBranches.map((branch) =>
						branchServicesApi.getByBranch(branch.id),
					),
				);
				if (!mounted) return;
				setBranches(nextBranches);
				setServices(serviceResp.data || []);
				setCounters(counterResp.data || []);
				setClients(clientResp.data || []);
				setBranchServices(branchServiceResp.flatMap((resp) => resp.data || []));
			} catch (error: any) {
				toast.error(error.message || "Failed to load setup progress");
			} finally {
				if (mounted) setIsLoading(false);
			}
		}
		load();
		return () => {
			mounted = false;
		};
	}, [currentOrganization]);

	const activeBranches = useMemo(
		() => branches.filter((branch) => branch.status === "active"),
		[branches],
	);
	const activeServices = useMemo(
		() => services.filter((service) => service.status === "active"),
		[services],
	);
	const activeBranchServices = useMemo(
		() => branchServices.filter((branchService) => branchService.is_active),
		[branchServices],
	);
	const activeCounters = useMemo(
		() => counters.filter((counter) => counter.status === "active"),
		[counters],
	);
	const stepDone = {
		tenant: true,
		branch: activeBranches.length > 0,
		service: activeServices.length > 0,
		branch_service: activeBranchServices.length > 0,
		counter: activeCounters.length > 0,
		qms_client: clients.length > 0,
	};
	const completed = Object.values(stepDone).filter(Boolean).length;
	const percent = Math.round((completed / steps.length) * 100);

	if (!currentOrganization) return null;

	return (
		<div className="space-y-6">
			<div className="flex items-center justify-between gap-3">
				<div>
					<h2 className="text-2xl font-bold tracking-tight">
						QMS Setup Wizard
					</h2>
					<p className="text-muted-foreground">
						Minimal guided setup. Existing forms stay source of truth.
					</p>
				</div>
				<Badge variant="outline">{percent}%</Badge>
			</div>

			<Card>
				<CardHeader>
					<CardTitle>Progress</CardTitle>
					<CardDescription>
						{activeBranches.length} active branches, {activeServices.length}{" "}
						active services, {activeBranchServices.length} active
						branch-services, {activeCounters.length} active counters,{" "}
						{clients.length} QMS clients.
					</CardDescription>
				</CardHeader>
				<CardContent>
					{isLoading ? (
						<p className="text-muted-foreground">Loading setup status...</p>
					) : (
						<div className="space-y-3">
							{steps.map((step) => {
								const done = stepDone[step.key];
								return (
									<div
										key={step.key}
										className="flex items-center justify-between rounded-lg border p-3"
									>
										<div>
											<p className="font-medium">{step.title}</p>
											<p className="text-muted-foreground text-sm">
												Open existing form / page.
											</p>
										</div>
										<div className="flex items-center gap-2">
											<Badge variant={done ? "default" : "secondary"}>
												{done ? "done" : "todo"}
											</Badge>
											<Button asChild variant="outline" size="sm">
												<Link href={step.href}>Open</Link>
											</Button>
										</div>
									</div>
								);
							})}
						</div>
					)}
				</CardContent>
			</Card>
		</div>
	);
}
