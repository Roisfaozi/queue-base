import type {
	BranchService,
	CallerActionRequest,
	CallerActionResponse,
	CallerLoginRequest,
	CallerLoginResponse,
	CallerMeResponse,
	Counter,
	EffectiveQueueConfigResponse,
	BranchResponse,
	BranchUpsertRequest,
	QMSClientResponse,
	QMSClientUpdateRequest,
	Queue,
	QueueJourney,
	QMSClientCreateRequest,
	QMSClientCreateResponse,
	QMSClientCredentialCreateRequest,
	QMSClientCredentialCreateResponse,
	QueueStatsResponse,
	ScannerCheckInResponse,
	Service,
	SignageCurrentCallResponse,
	SignageMeResponse,
	OperatorAssignmentRequest,
	OperatorAssignmentResponse,
	VisitJourney,
} from "@casbin/api-types";

import { api } from "./client";

export type Branch = BranchResponse;
export type BranchUpsertPayload = BranchUpsertRequest;

export type EffectiveQueueConfig = EffectiveQueueConfigResponse;
export type {
	CallerActionResponse,
	CallerMeResponse,
	Counter,
	BranchService,
	QMSClientResponse,
	QMSClientUpdateRequest,
	Queue,
	QueueJourney,
	QueueStatsResponse,
	ScannerCheckInResponse,
	Service,
	SignageCurrentCallResponse,
	SignageMeResponse,
	OperatorAssignmentRequest,
	OperatorAssignmentResponse,
	VisitJourney,
};

export const qmsClientsApi = {
	getAll: () => api.get<{ data: QMSClientResponse[] }>("/qms-clients"),
	getById: (id: string) =>
		api.get<{ data: QMSClientResponse }>(`/qms-clients/${id}`),
	create: (data: QMSClientCreateRequest) =>
		api.post<{ data: QMSClientCreateResponse }>("/qms-clients", data),
	update: (id: string, data: QMSClientUpdateRequest) =>
		api.patch<{ data: QMSClientResponse }>(`/qms-clients/${id}`, data),
	deactivate: (id: string) => api.delete(`/qms-clients/${id}`),
	createCredential: (data: QMSClientCredentialCreateRequest) =>
		api.post<{ data: QMSClientCredentialCreateResponse }>(
			"/qms-clients/credentials",
			data,
		),
};

export const callerApi = {
	login: (data: CallerLoginRequest, headers: QMSClientHeaders) =>
		api.post<{ data: CallerLoginResponse }>("/caller/login", data, {
			headers: qmsClientHeaders(headers),
		}),
	me: (headers: QMSClientHeaders) =>
		api.get<{ data: CallerMeResponse }>("/caller/me", {
			headers: qmsClientHeaders(headers),
		}),
	action: (
		journeyId: string,
		data: CallerActionRequest,
		headers: QMSClientHeaders,
	) =>
		api.post<{ data: CallerActionResponse }>(
			`/caller/queue-journeys/${journeyId}/action`,
			data,
			{ headers: qmsClientHeaders(headers) },
		),
};

type QMSClientHeaders = { clientId: string; apiKey: string };

function qmsClientHeaders(headers: QMSClientHeaders) {
	return {
		"X-Client-ID": headers.clientId,
		"X-API-Key": headers.apiKey,
	};
}

export const signageApi = {
	me: (headers: QMSClientHeaders) =>
		api.get<{ data: SignageMeResponse }>("/signage/me", {
			headers: qmsClientHeaders(headers),
		}),
	getCurrentCalls: (headers: QMSClientHeaders) =>
		api.get<{ data: SignageCurrentCallResponse[] }>("/signage/current-calls", {
			headers: qmsClientHeaders(headers),
		}),
	getQueues: (headers: QMSClientHeaders) =>
		api.get<{ data: Queue[] }>("/signage/queues", {
			headers: qmsClientHeaders(headers),
		}),
};

// -----------------------------------------------------------------------------
// BRANCHES API
// -----------------------------------------------------------------------------
export const branchesApi = {
	getAll: () => api.get<{ data: Branch[] }>("/branches"),
	getById: (id: string) => api.get<{ data: Branch }>(`/branches/${id}`),
	create: (data: BranchUpsertPayload) =>
		api.post<{ data: Branch }>("/branches", data),
	update: (id: string, data: BranchUpsertPayload) =>
		api.put<{ data: Branch }>(`/branches/${id}`, data),
	delete: (id: string) => api.delete(`/branches/${id}`),
};

// -----------------------------------------------------------------------------
// SERVICES API
// -----------------------------------------------------------------------------
export const servicesApi = {
	getAll: () => api.get<{ data: Service[] }>("/services"),
	getById: (id: string) => api.get<{ data: Service }>(`/services/${id}`),
	create: (data: {
		code: string;
		name: string;
		type?: string;
		default_estimated_duration?: number;
		audio_id?: string;
		audio_en?: string;
		narrative_instruction_id?: string;
		narrative_instruction_en?: string;
		is_pharmacy: boolean;
		is_pharmacy_reception: boolean;
	}) => api.post<{ data: Service }>("/services", data),
	update: (
		id: string,
		data: {
			code?: string;
			name?: string;
			type?: string;
			default_estimated_duration?: number;
			audio_id?: string;
			audio_en?: string;
			narrative_instruction_id?: string;
			narrative_instruction_en?: string;
			status?: "active" | "inactive";
			is_pharmacy?: boolean;
			is_pharmacy_reception?: boolean;
		},
	) => api.put<{ data: Service }>(`/services/${id}`, data),
	delete: (id: string) => api.delete(`/services/${id}`),
};

// -----------------------------------------------------------------------------
// COUNTERS API
// -----------------------------------------------------------------------------
export const countersApi = {
	getAll: () => api.get<{ data: Counter[] }>("/counters"),
	getById: (id: string) => api.get<{ data: Counter }>(`/counters/${id}`),
	create: (data: {
		branch_id: string;
		branch_service_id?: string;
		code: string;
		name: string;
		display_name?: string;
	}) => api.post<{ data: Counter }>("/counters", data),
	update: (
		id: string,
		data: {
			branch_service_id?: string;
			code?: string;
			name?: string;
			display_name?: string;
			status?: "active" | "inactive";
		},
	) => api.put<{ data: Counter }>(`/counters/${id}`, data),
	delete: (id: string) => api.delete(`/counters/${id}`),
};

// -----------------------------------------------------------------------------
// QUEUES API
// -----------------------------------------------------------------------------
export const queuesApi = {
	getAll: (params?: {
		branch_id?: string;
		status?: string;
		queue_date?: string;
		service_id?: string;
	}) => {
		const searchParams = new URLSearchParams();
		if (params?.branch_id) searchParams.append("branch_id", params.branch_id);
		if (params?.status) searchParams.append("status", params.status);
		if (params?.queue_date)
			searchParams.append("queue_date", params.queue_date);
		if (params?.service_id)
			searchParams.append("service_id", params.service_id);

		const query = searchParams.toString();
		return api.get<{ data: Queue[] }>(`/queues${query ? `?${query}` : ""}`);
	},
	getById: (id: string) => api.get<{ data: Queue }>(`/queues/${id}`),
	register: (data: {
		branch_id: string;
		service_id: string;
		patient_name: string;
		patient_id?: string;
	}) => api.post<{ data: Queue }>("/queues", data),
	transition: (
		id: string,
		data: { action: "call" | "serve" | "complete" | "skip" | "cancel" },
	) => api.post<{ data: Queue }>(`/queues/${id}/transition`, data),
	forward: (
		id: string,
		data: { destination_service_id: string; destination_counter_id?: string },
	) => api.post<{ data: Queue }>(`/queues/${id}/forward`, data),
	getVisitJourneys: (id: string) =>
		api.get<{ data: VisitJourney[] }>(`/queues/${id}/visit-journeys`),
	getQueueStats: (branchId: string) =>
		api.get<{ data: QueueStatsResponse }>(`/branches/${branchId}/queue-stats`),
	getJourneysByBranchAndService: (
		branchId: string,
		serviceId: string,
		params?: { status?: string; queue_date?: string },
	) => {
		const searchParams = new URLSearchParams();
		if (params?.status) searchParams.append("status", params.status);
		if (params?.queue_date)
			searchParams.append("queue_date", params.queue_date);
		return api.get<{ data: QueueJourney[] }>(
			`/branches/${branchId}/services/${serviceId}/queue-journeys${searchParams.toString() ? `?${searchParams.toString()}` : ""}`,
		);
	},
	getJourneysByBranchAndCounter: (
		branchId: string,
		counterId: string,
		params?: { status?: string; queue_date?: string },
	) => {
		const searchParams = new URLSearchParams();
		if (params?.status) searchParams.append("status", params.status);
		if (params?.queue_date)
			searchParams.append("queue_date", params.queue_date);
		return api.get<{ data: QueueJourney[] }>(
			`/branches/${branchId}/counters/${counterId}/queue-journeys${searchParams.toString() ? `?${searchParams.toString()}` : ""}`,
		);
	},
	updateTenant: (data: Record<string, unknown>) =>
		api.patch<{ data: void }>("/queue-config", data),
	updateBranch: (branchId: string, data: Record<string, unknown>) =>
		api.patch<{ data: void }>(`/branches/${branchId}/queue-config`, data),
	updateBranchService: (
		branchId: string,
		branchServiceId: string,
		data: Record<string, unknown>,
	) =>
		api.patch<{ data: void }>(
			`/branches/${branchId}/services/${branchServiceId}/queue-config`,
			data,
		),
	updateCounter: (
		branchId: string,
		counterId: string,
		data: Record<string, unknown>,
	) =>
		api.patch<{ data: void }>(
			`/branches/${branchId}/counters/${counterId}/queue-config`,
			data,
		),
	resetBranch: (branchId: string, field: string) =>
		api.delete(`/branches/${branchId}/queue-config/${field}`),
	resetBranchService: (
		branchId: string,
		branchServiceId: string,
		field: string,
	) =>
		api.delete(
			`/branches/${branchId}/services/${branchServiceId}/queue-config/${field}`,
		),
	resetCounter: (branchId: string, counterId: string, field: string) =>
		api.delete(
			`/branches/${branchId}/counters/${counterId}/queue-config/${field}`,
		),
};

// -----------------------------------------------------------------------------
// SCANNER API
// -----------------------------------------------------------------------------
export const scannerApi = {
	checkIn: (
		data: {
			action: "register" | "forward";
			branch_id: string;
			service_id?: string;
			patient_id?: string;
			patient_name?: string;
			queue_id?: string;
			destination_service_id?: string;
			destination_counter_id?: string;
		},
		headers: QMSClientHeaders,
	) =>
		api.post<{ data: ScannerCheckInResponse }>("/scanner/check-in", data, {
			headers: qmsClientHeaders(headers),
		}),
};

// -----------------------------------------------------------------------------
// BRANCH SERVICES API
// -----------------------------------------------------------------------------
export const branchServicesApi = {
	getByBranch: (branchId: string) =>
		api.get<{ data: BranchService[] }>(`/branches/${branchId}/services`),
};

// -----------------------------------------------------------------------------
// SETTINGS API
// -----------------------------------------------------------------------------
export const settingsApi = {
	getEffective: (params?: {
		branch_id?: string;
		service_id?: string;
		counter_id?: string;
	}) => {
		const searchParams = new URLSearchParams();
		if (params?.branch_id) searchParams.append("branch_id", params.branch_id);
		if (params?.service_id)
			searchParams.append("service_id", params.service_id);
		if (params?.counter_id)
			searchParams.append("counter_id", params.counter_id);

		const query = searchParams.toString();
		return api.get<{ data: EffectiveQueueConfigResponse }>(
			`/queue-config/effective${query ? `?${query}` : ""}`,
		);
	},
	updateTenant: (data: Record<string, unknown>) =>
		api.patch<{ data: void }>("/queue-config", data),
	updateBranch: (branchId: string, data: Record<string, unknown>) =>
		api.patch<{ data: void }>(`/branches/${branchId}/queue-config`, data),
	updateBranchService: (
		branchId: string,
		branchServiceId: string,
		data: Record<string, unknown>,
	) =>
		api.patch<{ data: void }>(
			`/branches/${branchId}/services/${branchServiceId}/queue-config`,
			data,
		),
	updateCounter: (
		branchId: string,
		counterId: string,
		data: Record<string, unknown>,
	) =>
		api.patch<{ data: void }>(
			`/branches/${branchId}/counters/${counterId}/queue-config`,
			data,
		),
	resetBranch: (branchId: string, field: string) =>
		api.delete(`/branches/${branchId}/queue-config/${field}`),
	resetBranchService: (
		branchId: string,
		branchServiceId: string,
		field: string,
	) =>
		api.delete(
			`/branches/${branchId}/services/${branchServiceId}/queue-config/${field}`,
		),
	resetCounter: (branchId: string, counterId: string, field: string) =>
		api.delete(
			`/branches/${branchId}/counters/${counterId}/queue-config/${field}`,
		),
};

export const operatorAssignmentsApi = {
  create: (data: OperatorAssignmentRequest) =>
    api.post<{ data: OperatorAssignmentResponse }>("/operator-counter-assignments", data),
  getAll: () =>
    api.get<{ data: OperatorAssignmentResponse[] }>("/operator-counter-assignments"),
  delete: (id: string) =>
    api.delete<{ data: void }>(`/operator-counter-assignments/${id}`),
};
