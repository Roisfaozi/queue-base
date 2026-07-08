"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import { toast } from "sonner";
import { useDashboardShell } from "~/app/[locale]/dashboard/_components/dashboard-shell-context";
import { useWebSocket } from "~/components/shared/providers/websocket-provider";
import { Icon } from "~/components/shared/icon";
import { Button } from "~/components/ui/button";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "~/components/ui/card";
import { Badge } from "~/components/ui/badge";
import { Input } from "~/components/ui/input";
import { Label } from "~/components/ui/label";
import {
	signageApi,
	type Queue,
	type SignageCurrentCallResponse,
	type SignageMeResponse,
} from "~/lib/api/qms";

export function SignageContent() {
	const { currentOrganization } = useDashboardShell();
	const { subscribe, unsubscribe } = useWebSocket();
	const [clientId, setClientId] = useState("");
	const [apiKey, setApiKey] = useState("");
	const [isLoadingMe, setIsLoadingMe] = useState(false);
	const [isLoadingCalls, setIsLoadingCalls] = useState(false);
	const [isLoadingQueues, setIsLoadingQueues] = useState(false);
	const [me, setMe] = useState<SignageMeResponse | null>(null);
	const [currentCalls, setCurrentCalls] = useState<
		SignageCurrentCallResponse[]
	>([]);
	const [queues, setQueues] = useState<Queue[]>([]);

	const headers = useMemo(
		() => ({ clientId: clientId.trim(), apiKey: apiKey.trim() }),
		[apiKey, clientId],
	);

	const hasHeaders = !!headers.clientId && !!headers.apiKey;

	const handleMe = useCallback(async () => {
		setIsLoadingMe(true);
		try {
			const response = await signageApi.me(headers);
			setMe(response.data);
			toast.success("Signage info loaded");
		} catch (error: any) {
			toast.error(error.message || "Failed to load signage info");
			setMe(null);
		} finally {
			setIsLoadingMe(false);
		}
	}, [headers]);

	const handleCurrentCalls = useCallback(async () => {
		setIsLoadingCalls(true);
		try {
			const response = await signageApi.getCurrentCalls(headers);
			setCurrentCalls(response.data || []);
		} catch (error: any) {
			toast.error(error.message || "Failed to load current calls");
			setCurrentCalls([]);
		} finally {
			setIsLoadingCalls(false);
		}
	}, [headers]);

	const handleQueues = useCallback(async () => {
		setIsLoadingQueues(true);
		try {
			const response = await signageApi.getQueues(headers);
			setQueues(response.data || []);
		} catch (error: any) {
			toast.error(error.message || "Failed to load queues");
			setQueues([]);
		} finally {
			setIsLoadingQueues(false);
		}
	}, [headers]);

	useEffect(() => {
		if (!hasHeaders) return;
		void handleMe();
		void handleCurrentCalls();
		void handleQueues();
	}, [handleCurrentCalls, handleMe, handleQueues, hasHeaders]);

	useEffect(() => {
		if (!me?.tenant_id || !me?.branch_id) return;
		const channel = `queue:${me.tenant_id}:${me.branch_id}`;
		const onMessage = (message: any) => {
			if (message?.type !== "queue_update") return;
			void handleCurrentCalls();
			void handleQueues();
		};
		subscribe(channel, onMessage);
		return () => unsubscribe(channel, onMessage);
	}, [
		me?.tenant_id,
		me?.branch_id,
		handleCurrentCalls,
		handleQueues,
		subscribe,
		unsubscribe,
	]);

	if (!currentOrganization) return null;

	return (
		<div className="space-y-6">
			<div>
				<h2 className="text-2xl font-bold tracking-tight">Signage Display</h2>
				<p className="text-muted-foreground">
					Live display untuk signage client. Status tenant, current calls, dan
					waiting queues tampil otomatis setelah credential valid.
				</p>
			</div>

			<Card>
				<CardHeader>
					<CardTitle>Display Binding</CardTitle>
					<CardDescription>
						Isi credential binding sekali, lalu display refresh otomatis.
					</CardDescription>
				</CardHeader>
				<CardContent className="grid gap-4 md:grid-cols-2">
					<div className="space-y-2">
						<Label htmlFor="signage-client-id">X-Client-ID</Label>
						<Input
							id="signage-client-id"
							value={clientId}
							onChange={(e) => setClientId(e.target.value)}
							placeholder="signage-device-01"
						/>
					</div>
					<div className="space-y-2">
						<Label htmlFor="signage-api-key">X-API-Key</Label>
						<Input
							id="signage-api-key"
							type="password"
							value={apiKey}
							onChange={(e) => setApiKey(e.target.value)}
							placeholder="••••••••"
						/>
					</div>
				</CardContent>
			</Card>

			<div className="flex flex-wrap gap-2">
				<Button onClick={handleMe} disabled={!hasHeaders || isLoadingMe}>
					{isLoadingMe && (
						<Icon name="Loader" className="mr-2 h-4 w-4 animate-spin" />
					)}
					Refresh Display
				</Button>
				<Button
					variant="secondary"
					onClick={handleCurrentCalls}
					disabled={!hasHeaders || isLoadingCalls}
				>
					{isLoadingCalls && (
						<Icon name="Loader" className="mr-2 h-4 w-4 animate-spin" />
					)}
					Refresh Calls
				</Button>
				<Button
					variant="outline"
					onClick={handleQueues}
					disabled={!hasHeaders || isLoadingQueues}
				>
					{isLoadingQueues && (
						<Icon name="Loader" className="mr-2 h-4 w-4 animate-spin" />
					)}
					Refresh Queues
				</Button>
			</div>

			{hasHeaders && !me && !isLoadingMe && (
				<Card>
					<CardContent className="py-6 text-sm text-muted-foreground">
						Display ready. Refresh akan jalan otomatis setelah binding valid.
					</CardContent>
				</Card>
			)}

			{me && (
				<Card>
					<CardHeader>
						<CardTitle>Client Info</CardTitle>
					</CardHeader>
					<CardContent>
						<div className="grid grid-cols-2 gap-2 text-sm">
							<div>Client ID:</div>
							<div className="font-mono">{me.client_id}</div>
							<div>Tenant:</div>
							<div className="font-mono">{me.tenant_id}</div>
							<div>Branch:</div>
							<div className="font-mono">{me.branch_name || me.branch_id}</div>
							{me.service_name && (
								<>
									<div>Service:</div>
									<div className="font-mono">{me.service_name}</div>
								</>
							)}
							{me.counter_display_name && (
								<>
									<div>Counter:</div>
									<div className="font-mono">{me.counter_display_name}</div>
								</>
							)}
							{me.running_text && (
								<>
									<div>Running Text:</div>
									<div className="font-mono">{me.running_text}</div>
								</>
							)}
							{me.audio_id && (
								<>
									<div>Audio:</div>
									<div className="font-mono">{me.audio_id}</div>
								</>
							)}
							{me.narrative_instruction_id && (
								<>
									<div>Narrative:</div>
									<div className="font-mono">{me.narrative_instruction_id}</div>
								</>
							)}
							<div>Name:</div>
							<div className="font-mono">{me.name}</div>
							<div>Client Type:</div>
							<div className="font-mono">{me.client_type}</div>
						</div>
					</CardContent>
				</Card>
			)}

			{currentCalls.length > 0 && (
				<Card>
					<CardHeader>
						<CardTitle>Current Calls ({currentCalls.length})</CardTitle>
					</CardHeader>
					<CardContent>
						<div className="divide-y text-sm">
							{currentCalls.map((call) => (
								<div
									key={call.queue_id}
									className="flex items-center justify-between py-2"
								>
									<div>
										<span className="font-mono font-semibold text-lg">
											{call.ticket_no}
										</span>
										<span className="text-muted-foreground ml-2">
											─ {call.counter_display_name || call.counter_id}
										</span>
									</div>
									<Badge variant="outline">
										{call.service_type || call.service_id}
									</Badge>
								</div>
							))}
						</div>
					</CardContent>
				</Card>
			)}

			{queues.length > 0 && (
				<Card>
					<CardHeader>
						<CardTitle>Waiting Queues ({queues.length})</CardTitle>
					</CardHeader>
					<CardContent>
						<div className="divide-y text-sm">
							{queues.map((q) => (
								<div
									key={q.id}
									className="flex items-center justify-between py-2"
								>
									<div>
										<span className="font-mono font-semibold">
											{q.ticket_no}
										</span>
										<span className="text-muted-foreground ml-2">
											{q.patient_name || "—"} — {q.status}
										</span>
									</div>
									<Badge variant="outline">#{q.queue_no}</Badge>
								</div>
							))}
						</div>
					</CardContent>
				</Card>
			)}

			{currentCalls.length === 0 && me && !isLoadingCalls && (
				<Card>
					<CardContent className="py-6 text-sm text-muted-foreground">
						No current calls. Display akan ikut update saat queue berubah.
					</CardContent>
				</Card>
			)}

			{queues.length === 0 && me && !isLoadingQueues && (
				<Card>
					<CardContent className="py-6 text-sm text-muted-foreground">
						No waiting queues for current binding.
					</CardContent>
				</Card>
			)}
		</div>
	);
}
