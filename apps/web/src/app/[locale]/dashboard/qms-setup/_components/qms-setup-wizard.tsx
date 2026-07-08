"use client";

import Link from "next/link";
import { useEffect, useMemo, useState } from "react";
import { toast } from "sonner";
import type { QMSClientCreateResponse } from "@casbin/api-types";
import { useDashboardShell } from "~/app/[locale]/dashboard/_components/dashboard-shell-context";
import { Icon } from "~/components/shared/icon";
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
		description: "Pastikan tenant aktif dan profile dasar siap dipakai cabang.",
		requirement: "Organization context terpilih.",
	},
	{
		key: "branch",
		title: "Branches",
		href: "/dashboard/branches",
		description: "Buat minimal satu branch aktif dengan profile lengkap.",
		requirement: "Butuh branch status active.",
	},
	{
		key: "service",
		title: "Services",
		href: "/dashboard/services",
		description:
			"Daftarkan service aktif beserta audio dan narrative bila perlu.",
		requirement: "Butuh minimal satu service active.",
	},
	{
		key: "branch_service",
		title: "Branch Services",
		href: "/dashboard/queues",
		description: "Hubungkan service ke branch tujuan operasional.",
		requirement: "Butuh branch active dan service active.",
	},
	{
		key: "counter",
		title: "Counters",
		href: "/dashboard/counters",
		description: "Aktifkan counter pada branch untuk jalur pelayanan.",
		requirement: "Butuh branch active.",
	},
	{
		key: "qms_client",
		title: "QMS Clients",
		href: "/dashboard/qms-clients",
		description: "Daftarkan caller, signage, scanner, atau kiosk device.",
		requirement:
			"Butuh branch active; opsional bind ke branch-service/counter.",
	},
	{
		key: "operator_assignment",
		title: "Operator Assignments",
		href: "/dashboard/operator-assignments",
		description: "Assign operator ke counter untuk mulai layani antrian.",
		requirement: "Butuh branch, user, dan counter active.",
	},
] as const;

type StepKey = (typeof steps)[number]["key"];
export function QMSSetupWizard() {
	const { currentOrganization } = useDashboardShell();
	const [branches, setBranches] = useState<Branch[]>([]);
	const [services, setServices] = useState<Service[]>([]);
	const [branchServices, setBranchServices] = useState<BranchService[]>([]);
	const [counters, setCounters] = useState<Counter[]>([]);
	const [clients, setClients] = useState<QMSClientCreateResponse[]>([]);
	const [isLoading, setIsLoading] = useState(true);
	const [activeStep, setActiveStep] = useState<StepKey>("tenant");

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
	const draftBranches = useMemo(
		() => branches.filter((branch) => branch.status === "draft"),
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
		operator_assignment: true,
	};
	const firstTodo =
		steps.find((step) => !stepDone[step.key])?.key ??
		steps[steps.length - 1].key;

	useEffect(() => {
		setActiveStep(firstTodo);
	}, [firstTodo]);

	const completed = Object.values(stepDone).filter(Boolean).length;
	const percent = Math.round((completed / steps.length) * 100);
	const activeIndex = steps.findIndex((step) => step.key === activeStep);
	const currentStep = steps[activeIndex] ?? steps[0];
	const currentDone = stepDone[currentStep.key];
	const blockers = useMemo(() => {
		switch (currentStep.key) {
			case "tenant":
				return [] as string[];
			case "branch":
				const br: string[] = [];
				if (activeBranches.length === 0 && draftBranches.length === 0) {
					br.push(
						"Belum ada branch. Buat branch baru dengan isi Address, City, Province, Phone, dan Timezone.",
					);
				}
				if (draftBranches.length > 0) {
					br.push(
						"Branch draft butuh dilengkapi: Address, City, Province, Phone, Timezone, lalu update status jadi active.",
					);
				}
				return br;
			case "service":
				return activeServices.length > 0 ? [] : ["Belum ada service active."];
			case "branch_service":
				return [
					...(activeBranches.length > 0 ? [] : ["Branch active belum ada."]),
					...(activeServices.length > 0 ? [] : ["Service active belum ada."]),
					...(activeBranchServices.length > 0
						? []
						: ["Belum ada branch-service active."]),
				];
			case "counter":
				return [
					...(activeBranches.length > 0 ? [] : ["Branch active belum ada."]),
					...(activeCounters.length > 0 ? [] : ["Belum ada counter active."]),
				];
			case "operator_assignment":
				return [];
			case "qms_client":
				return [
					...(activeBranches.length > 0 ? [] : ["Branch active belum ada."]),
					...(clients.length > 0 ? [] : ["Belum ada QMS client."]),
				];
		}
	}, [
		currentStep.key,
		activeBranches.length,
		activeServices.length,
		activeBranchServices.length,
		activeCounters.length,
		clients.length,
	]);

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
						{activeBranches.length} active branches
						{draftBranches.length > 0 ? `, ${draftBranches.length} draft` : ""},{" "}
						{activeServices.length} active services,{" "}
						{activeBranchServices.length} active branch-services,{" "}
						{activeCounters.length} active counters, {clients.length} QMS
						clients.
					</CardDescription>
				</CardHeader>
				<CardContent>
					{isLoading ? (
						<p className="text-muted-foreground">Loading setup status...</p>
					) : (
						<div className="space-y-3">
							{steps.map((step, index) => {
								const done = stepDone[step.key];
								const current = step.key === currentStep.key;
								return (
									<div
										key={step.key}
										className={`flex items-center justify-between rounded-lg border p-3 ${
											current ? "border-primary bg-primary/5" : ""
										}`}
									>
										<div>
											<p className="font-medium">
												{index + 1}. {step.title}
											</p>
											<p className="text-muted-foreground text-sm">
												{step.description}
											</p>
										</div>
										<div className="flex items-center gap-2">
											<Button
												variant="ghost"
												size="sm"
												onClick={() => setActiveStep(step.key)}
											>
												{current ? "Viewing" : "View"}
											</Button>
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

							<Card className="border-dashed">
								<CardHeader>
									<CardTitle className="flex items-center gap-2 text-base">
										<Icon
											name={currentDone ? "CircleCheck" : "CircleArrowRight"}
											className="h-4 w-4"
										/>
										Current Step: {currentStep.title}
									</CardTitle>
									<CardDescription>{currentStep.requirement}</CardDescription>
								</CardHeader>
								<CardContent className="space-y-4">
									<p className="text-sm text-muted-foreground">
										{currentStep.description}
									</p>
									{blockers.length > 0 ? (
										<div className="rounded-md border border-amber-500/30 bg-amber-500/5 p-3 text-sm">
											<p className="font-medium">Blockers</p>
											<ul className="text-muted-foreground mt-2 space-y-1">
												{blockers.map((blocker) => (
													<li key={blocker}>- {blocker}</li>
												))}
											</ul>
										</div>
									) : (
										<div className="rounded-md border border-emerald-500/30 bg-emerald-500/5 p-3 text-sm font-medium">
											Step ready. Lanjut ke tahap berikutnya.
										</div>
									)}
									<div className="flex flex-wrap gap-2">
										<Button
											variant="outline"
											size="sm"
											onClick={() =>
												setActiveStep(steps[Math.max(activeIndex - 1, 0)].key)
											}
											disabled={activeIndex === 0}
										>
											Previous
										</Button>
										<Button asChild size="sm">
											<Link href={currentStep.href}>Open Step Page</Link>
										</Button>
										<Button
											variant="secondary"
											size="sm"
											onClick={() => window.location.reload()}
										>
											Refresh Status
										</Button>
										<Button
											variant="outline"
											size="sm"
											onClick={() =>
												setActiveStep(
													steps[Math.min(activeIndex + 1, steps.length - 1)]
														.key,
												)
											}
											disabled={activeIndex === steps.length - 1}
										>
											Next
										</Button>
									</div>
								</CardContent>
							</Card>
						</div>
					)}
				</CardContent>
			</Card>
		</div>
	);
}
