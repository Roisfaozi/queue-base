"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import { toast } from "sonner";
import { useDashboardShell } from "~/app/[locale]/dashboard/_components/dashboard-shell-context";
import { Icon } from "~/components/shared/icon";
import { Button } from "~/components/ui/button";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "~/components/ui/card";
import { Input } from "~/components/ui/input";
import { Label } from "~/components/ui/label";
import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from "~/components/ui/select";
import {
	branchesApi,
	countersApi,
	scannerApi,
	servicesApi,
	type Queue,
} from "~/lib/api/qms";

type BranchOption = { id: string; code: string; name: string; status: string };
type ServiceOption = { id: string; code: string; name: string; status: string };
type CounterOption = { id: string; code: string; name: string; status: string };

export function ScannerContent() {
	const { currentOrganization } = useDashboardShell();
	const [clientId, setClientId] = useState("");
	const [apiKey, setApiKey] = useState("");
	const [action, setAction] = useState<"register" | "forward">("register");
	const [branchId, setBranchId] = useState("");
	const [serviceId, setServiceId] = useState("");
	const [patientName, setPatientName] = useState("");
	const [queueId, setQueueId] = useState("");
	const [destinationServiceId, setDestinationServiceId] = useState("");
	const [destinationCounterId, setDestinationCounterId] = useState("");
	const [isLoading, setIsLoading] = useState(false);
	const [lastResult, setLastResult] = useState<Queue | null>(null);
	const [branches, setBranches] = useState<BranchOption[]>([]);
	const [services, setServices] = useState<ServiceOption[]>([]);
	const [counters, setCounters] = useState<CounterOption[]>([]);

	const fetchRefs = useCallback(async () => {
		try {
			const [bResp, sResp, cResp] = await Promise.all([
				branchesApi.getAll(),
				servicesApi.getAll(),
				countersApi.getAll(),
			]);
			setBranches((bResp.data || []).filter((b) => b.status === "active"));
			setServices((sResp.data || []).filter((s) => s.status === "active"));
			setCounters((cResp.data || []).filter((c) => c.status === "active"));
		} catch {
			// ignore reference fetch failure; submit path still reports real error
		}
	}, []);

	useEffect(() => {
		if (currentOrganization) {
			fetchRefs();
		}
	}, [currentOrganization, fetchRefs]);

	const canRegister = useMemo(
		() => !!branchId && !!serviceId && patientName.trim().length >= 2,
		[branchId, serviceId, patientName],
	);

	const canForward = useMemo(
		() => !!branchId && !!queueId.trim() && !!destinationServiceId,
		[branchId, queueId, destinationServiceId],
	);

	const canSubmit = useMemo(
		() =>
			!!clientId.trim() &&
			!!apiKey.trim() &&
			(action === "register" ? canRegister : canForward),
		[action, apiKey, canForward, canRegister, clientId],
	);

	const handleSubmit = async () => {
		setIsLoading(true);
		setLastResult(null);
		try {
			const response =
				action === "register"
					? await scannerApi.checkIn(
							{
								action: "register",
								branch_id: branchId,
								service_id: serviceId,
								patient_name: patientName,
							},
							{ clientId, apiKey },
						)
					: await scannerApi.checkIn(
							{
								action: "forward",
								branch_id: branchId,
								queue_id: queueId,
								destination_service_id: destinationServiceId,
								destination_counter_id: destinationCounterId || undefined,
							},
							{ clientId, apiKey },
						);

			setLastResult(response.data.queue);
			toast.success(
				action === "register"
					? "Queue registered via scanner"
					: "Queue forwarded via scanner",
			);
		} catch (error: any) {
			toast.error(error.message || "Scanner request failed");
		} finally {
			setIsLoading(false);
		}
	};

	if (!currentOrganization) return null;

	return (
		<div className="space-y-6">
			<div>
				<h2 className="text-2xl font-bold tracking-tight">Scanner Check-in</h2>
				<p className="text-muted-foreground">
					Register atau forward antrean via credential device scanner.
				</p>
			</div>

			<Card>
				<CardHeader>
					<CardTitle>Scanner Credentials</CardTitle>
					<CardDescription>
						Gunakan header scanner. Secret tidak masuk request body.
					</CardDescription>
				</CardHeader>
				<CardContent className="grid gap-4 md:grid-cols-2">
					<div className="space-y-2">
						<Label htmlFor="client-id">X-Client-ID</Label>
						<Input
							id="client-id"
							value={clientId}
							onChange={(e) => setClientId(e.target.value)}
							placeholder="scanner-device-01"
						/>
					</div>
					<div className="space-y-2">
						<Label htmlFor="api-key">X-API-Key</Label>
						<Input
							id="api-key"
							type="password"
							value={apiKey}
							onChange={(e) => setApiKey(e.target.value)}
							placeholder="••••••••"
						/>
					</div>
				</CardContent>
			</Card>

			<Card>
				<CardHeader>
					<CardTitle>Check-in Request</CardTitle>
					<CardDescription>
						Surface admin untuk verifikasi flow scanner register dan forward.
					</CardDescription>
				</CardHeader>
				<CardContent className="grid gap-4 md:grid-cols-2">
					<div className="space-y-2">
						<Label>Action</Label>
						<Select
							value={action}
							onValueChange={(value) =>
								setAction(value as "register" | "forward")
							}
						>
							<SelectTrigger>
								<SelectValue placeholder="Select action" />
							</SelectTrigger>
							<SelectContent>
								<SelectItem value="register">Register</SelectItem>
								<SelectItem value="forward">Forward</SelectItem>
							</SelectContent>
						</Select>
					</div>

					<div className="space-y-2">
						<Label>Branch</Label>
						<Select value={branchId} onValueChange={setBranchId}>
							<SelectTrigger>
								<SelectValue placeholder="Select branch" />
							</SelectTrigger>
							<SelectContent>
								{branches.map((branch) => (
									<SelectItem key={branch.id} value={branch.id}>
										{branch.code} — {branch.name}
									</SelectItem>
								))}
							</SelectContent>
						</Select>
					</div>

					{action === "register" ? (
						<>
							<div className="space-y-2">
								<Label>Service</Label>
								<Select value={serviceId} onValueChange={setServiceId}>
									<SelectTrigger>
										<SelectValue placeholder="Select service" />
									</SelectTrigger>
									<SelectContent>
										{services.map((service) => (
											<SelectItem key={service.id} value={service.id}>
												{service.code} — {service.name}
											</SelectItem>
										))}
									</SelectContent>
								</Select>
							</div>
							<div className="space-y-2">
								<Label htmlFor="patient-name">Patient Name</Label>
								<Input
									id="patient-name"
									value={patientName}
									onChange={(e) => setPatientName(e.target.value)}
									placeholder="John Doe"
								/>
							</div>
						</>
					) : (
						<>
							<div className="space-y-2">
								<Label htmlFor="queue-id">Queue ID</Label>
								<Input
									id="queue-id"
									value={queueId}
									onChange={(e) => setQueueId(e.target.value)}
									placeholder="550e8400-e29b-41d4-a716-446655440500"
								/>
							</div>
							<div className="space-y-2">
								<Label>Destination Service</Label>
								<Select
									value={destinationServiceId}
									onValueChange={setDestinationServiceId}
								>
									<SelectTrigger>
										<SelectValue placeholder="Select destination service" />
									</SelectTrigger>
									<SelectContent>
										{services.map((service) => (
											<SelectItem key={service.id} value={service.id}>
												{service.code} — {service.name}
											</SelectItem>
										))}
									</SelectContent>
								</Select>
							</div>
							<div className="space-y-2 md:col-span-2">
								<Label>Destination Counter</Label>
								<Select
									value={destinationCounterId}
									onValueChange={setDestinationCounterId}
								>
									<SelectTrigger>
										<SelectValue placeholder="Optional destination counter" />
									</SelectTrigger>
									<SelectContent>
										{counters.map((counter) => (
											<SelectItem key={counter.id} value={counter.id}>
												{counter.code} — {counter.name}
											</SelectItem>
										))}
									</SelectContent>
								</Select>
							</div>
						</>
					)}
				</CardContent>
			</Card>

			<div className="flex justify-end">
				<Button onClick={handleSubmit} disabled={!canSubmit || isLoading}>
					{isLoading ? (
						<Icon name="Loader" className="mr-2 h-4 w-4 animate-spin" />
					) : (
						<Icon name="Scan" className="mr-2 h-4 w-4" />
					)}
					{action === "register" ? "Register" : "Forward"} via Scanner
				</Button>
			</div>

			{lastResult && (
				<Card>
					<CardHeader>
						<CardTitle>Last Queue Result</CardTitle>
						<CardDescription>
							Response queue terakhir dari scanner endpoint.
						</CardDescription>
					</CardHeader>
					<CardContent>
						<pre className="overflow-x-auto rounded-md bg-muted p-4 text-sm">
							{JSON.stringify(lastResult, null, 2)}
						</pre>
					</CardContent>
				</Card>
			)}
		</div>
	);
}
