"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import { toast } from "sonner";
import { useDashboardShell } from "~/app/[locale]/dashboard/_components/dashboard-shell-context";
import { Icon } from "~/components/shared/icon";
import { useWebSocket } from "~/components/shared/providers/websocket-provider";
import { Badge } from "~/components/ui/badge";
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
	callerApi,
	queuesApi,
	type CallerActionResponse,
	type CallerMeResponse,
	type QueueJourney,
	type Service,
	servicesApi,
} from "~/lib/api/qms";

const ACTIONS = ["call", "serve", "complete", "skip", "cancel"] as const;

type BranchOption = { id: string; code: string; name: string; status: string };

export function CallerContent() {
	const { currentOrganization } = useDashboardShell();
	const { subscribe, unsubscribe } = useWebSocket();
	const [clientId, setClientId] = useState("");
	const [apiKey, setApiKey] = useState("");
	const [username, setUsername] = useState("");
	const [password, setPassword] = useState("");
	const [branchId, setBranchId] = useState("");
	const [serviceId, setServiceId] = useState("");
	const [journeyId, setJourneyId] = useState("");
	const [action, setAction] = useState<(typeof ACTIONS)[number]>("call");
	const [isLoggingIn, setIsLoggingIn] = useState(false);
	const [isLoadingJourneys, setIsLoadingJourneys] = useState(false);
	const [isActing, setIsActing] = useState(false);
	const [me, setMe] = useState<CallerMeResponse | null>(null);
	const [branches, setBranches] = useState<BranchOption[]>([]);
	const [services, setServices] = useState<Service[]>([]);
	const [journeys, setJourneys] = useState<QueueJourney[]>([]);
	const [lastAction, setLastAction] = useState<CallerActionResponse | null>(
		null,
	);

	const headers = useMemo(
		() => ({ clientId: clientId.trim(), apiKey: apiKey.trim() }),
		[apiKey, clientId],
	);

	const fetchRefs = useCallback(async () => {
		try {
			const [branchResp, serviceResp] = await Promise.all([
				branchesApi.getAll(),
				servicesApi.getAll(),
			]);
			setBranches((branchResp.data || []).filter((b) => b.status === "active"));
			setServices(
				(serviceResp.data || []).filter((s) => s.status === "active"),
			);
		} catch {
			// ponytail: refs only help picker labels; submit path still shows real error.
		}
	}, []);

	useEffect(() => {
		if (currentOrganization) {
			fetchRefs();
		}
	}, [currentOrganization, fetchRefs]);

	const fetchJourneys = useCallback(async () => {
		if (!branchId || !serviceId || !headers.clientId || !headers.apiKey) {
			setJourneys([]);
			return;
		}
		setIsLoadingJourneys(true);
		try {
			const response = await queuesApi.getJourneysByBranchAndService(
				branchId,
				serviceId,
				{
					status: "waiting",
				},
			);
			setJourneys(response.data || []);
		} catch (error: any) {
			toast.error(error.message || "Failed to load queue journeys");
			setJourneys([]);
		} finally {
			setIsLoadingJourneys(false);
		}
	}, [branchId, headers.apiKey, headers.clientId, serviceId]);

	const handleLogin = async () => {
		setIsLoggingIn(true);
		setLastAction(null);
		try {
			await callerApi.login({ username: username.trim(), password }, headers);
			const meResponse = await callerApi.me(headers);
			setMe(meResponse.data);
			setBranchId(meResponse.data.context.branch_id);
			setServiceId(meResponse.data.context.branch_service_id || "");
			toast.success("Caller login success");
		} catch (error: any) {
			toast.error(error.message || "Caller login failed");
			setMe(null);
		} finally {
			setIsLoggingIn(false);
		}
	};

	const handleRefreshMe = async () => {
		try {
			const meResponse = await callerApi.me(headers);
			setMe(meResponse.data);
			toast.success("Caller context refreshed");
		} catch (error: any) {
			toast.error(error.message || "Failed to load caller context");
		}
	};

	useEffect(() => {
		if (!me?.context?.tenant_id || !me?.context?.branch_id) return;
		const channel = `queue:${me.context.tenant_id}:${me.context.branch_id}`;
		const onMessage = (message: any) => {
			if (message?.type !== "queue_update") return;
			void fetchJourneys();
			void handleRefreshMe();
		};
		subscribe(channel, onMessage);
		return () => unsubscribe(channel, onMessage);
	}, [
		fetchJourneys,
		handleRefreshMe,
		me?.context?.branch_id,
		me?.context?.tenant_id,
		subscribe,
		unsubscribe,
	]);

	const handleAction = async () => {
		if (!journeyId.trim()) return;
		setIsActing(true);
		setLastAction(null);
		try {
			const response = await callerApi.action(
				journeyId.trim(),
				{ action },
				headers,
			);
			setLastAction(response.data);
			toast.success(`Caller action ${action} success`);
			await fetchJourneys();
		} catch (error: any) {
			toast.error(error.message || "Caller action failed");
		} finally {
			setIsActing(false);
		}
	};

	const canLogin =
		!!headers.clientId &&
		!!headers.apiKey &&
		username.trim().length >= 3 &&
		password.length >= 8;
	const canAct = !!headers.clientId && !!headers.apiKey && !!journeyId.trim();

	if (!currentOrganization) return null;

	return (
		<div className="space-y-6">
			<div>
				<h2 className="text-2xl font-bold tracking-tight">Caller</h2>
				<p className="text-muted-foreground">
					Thin admin surface untuk login caller, inspect bound context, dan
					trigger action endpoint.
				</p>
			</div>

			<Card>
				<CardHeader>
					<CardTitle>Client Credentials</CardTitle>
					<CardDescription>
						Caller endpoints butuh session operator plus client credential
						headers.
					</CardDescription>
				</CardHeader>
				<CardContent className="grid gap-4 md:grid-cols-2">
					<div className="space-y-2">
						<Label htmlFor="caller-client-id">X-Client-ID</Label>
						<Input
							id="caller-client-id"
							value={clientId}
							onChange={(e) => setClientId(e.target.value)}
							placeholder="caller-device-01"
						/>
					</div>
					<div className="space-y-2">
						<Label htmlFor="caller-api-key">X-API-Key</Label>
						<Input
							id="caller-api-key"
							type="password"
							value={apiKey}
							onChange={(e) => setApiKey(e.target.value)}
							placeholder="••••••••"
						/>
					</div>
				</CardContent>
			</Card>

			<div className="grid gap-6 xl:grid-cols-2">
				<Card>
					<CardHeader>
						<CardTitle>Operator Login</CardTitle>
						<CardDescription>
							Login operator lewat `POST /caller/login`, lalu cek `GET
							/caller/me`.
						</CardDescription>
					</CardHeader>
					<CardContent className="space-y-4">
						<div className="space-y-2">
							<Label htmlFor="caller-username">Username</Label>
							<Input
								id="caller-username"
								value={username}
								onChange={(e) => setUsername(e.target.value)}
								placeholder="operator"
							/>
						</div>
						<div className="space-y-2">
							<Label htmlFor="caller-password">Password</Label>
							<Input
								id="caller-password"
								type="password"
								value={password}
								onChange={(e) => setPassword(e.target.value)}
								placeholder="••••••••"
							/>
						</div>
						<div className="flex gap-2">
							<Button onClick={handleLogin} disabled={!canLogin || isLoggingIn}>
								{isLoggingIn && (
									<Icon name="Loader" className="mr-2 h-4 w-4 animate-spin" />
								)}
								Login Caller
							</Button>
							<Button
								variant="outline"
								onClick={handleRefreshMe}
								disabled={!canLogin}
							>
								Refresh Me
							</Button>
						</div>

						{me && (
							<div className="rounded-md border bg-muted/30 p-3 text-sm space-y-2">
								<p className="font-medium">Bound Context</p>
								<div className="flex flex-wrap gap-2">
									<Badge variant="secondary">
										Tenant: {me.context.tenant_name || me.context.tenant_id}
									</Badge>
									<Badge variant="secondary">
										Branch: {me.context.branch_name || me.context.branch_id}
									</Badge>
									{me.context.service_name && (
										<Badge variant="secondary">
											Service: {me.context.service_name}
										</Badge>
									)}
									{me.context.display_name && (
										<Badge variant="secondary">
											Counter: {me.context.display_name}
										</Badge>
									)}
								</div>
								<p className="text-muted-foreground text-xs">
									Permissions: {me.permissions.join(", ") || "-"}
								</p>
							</div>
						)}
					</CardContent>
				</Card>

				<Card>
					<CardHeader>
						<CardTitle>Journey Action</CardTitle>
						<CardDescription>
							Action tipis untuk current branch/service scope. Picker hanya
							bantu isi ID journey.
						</CardDescription>
					</CardHeader>
					<CardContent className="space-y-4">
						<div className="grid gap-4 md:grid-cols-2">
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
						</div>

						<div className="flex gap-2">
							<Button
								variant="outline"
								onClick={fetchJourneys}
								disabled={!branchId || !serviceId || isLoadingJourneys}
							>
								{isLoadingJourneys && (
									<Icon name="Loader" className="mr-2 h-4 w-4 animate-spin" />
								)}
								Load Waiting Journeys
							</Button>
						</div>

						<div className="space-y-2">
							<Label>Suggested Journey</Label>
							<Select value={journeyId} onValueChange={setJourneyId}>
								<SelectTrigger>
									<SelectValue placeholder="Pick waiting journey" />
								</SelectTrigger>
								<SelectContent>
									{journeys.map((journey) => (
										<SelectItem key={journey.id} value={journey.id}>
											{journey.id} — seq {journey.seq_no} — {journey.status}
										</SelectItem>
									))}
								</SelectContent>
							</Select>
						</div>

						<div className="space-y-2">
							<Label htmlFor="caller-journey-id">Journey ID</Label>
							<Input
								id="caller-journey-id"
								value={journeyId}
								onChange={(e) => setJourneyId(e.target.value)}
								placeholder="queue-journey-uuid"
							/>
						</div>

						<div className="space-y-2">
							<Label>Action</Label>
							<Select
								value={action}
								onValueChange={(value) =>
									setAction(value as (typeof ACTIONS)[number])
								}
							>
								<SelectTrigger>
									<SelectValue placeholder="Select action" />
								</SelectTrigger>
								<SelectContent>
									{ACTIONS.map((item) => (
										<SelectItem key={item} value={item}>
											{item}
										</SelectItem>
									))}
								</SelectContent>
							</Select>
						</div>

						<Button onClick={handleAction} disabled={!canAct || isActing}>
							{isActing && (
								<Icon name="Loader" className="mr-2 h-4 w-4 animate-spin" />
							)}
							Run Action
						</Button>

						{lastAction && (
							<div className="rounded-md border bg-muted/30 p-3 text-sm space-y-1">
								<p className="font-medium">Last action result</p>
								<p>Status: {lastAction.status}</p>
								<p>Journey: {lastAction.journey_id}</p>
								{lastAction.track_no && <p>Track: {lastAction.track_no}</p>}
								{lastAction.queue_no ? (
									<p>Queue No: {lastAction.queue_no}</p>
								) : null}
							</div>
						)}
					</CardContent>
				</Card>
			</div>
		</div>
	);
}
