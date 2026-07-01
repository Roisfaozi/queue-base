import { api } from "./client";

export interface Branch {
	id: string;
	tenant_id: string;
	code: string;
	name: string;
	status: "active" | "inactive";
	created_at: number;
	updated_at: number;
}

export interface Service {
	id: string;
	tenant_id: string;
	code: string;
	name: string;
	status: "active" | "inactive";
	is_pharmacy: boolean;
	is_pharmacy_reception: boolean;
	created_at: number;
	updated_at: number;
}

export interface Counter {
	id: string;
	tenant_id: string;
	branch_id: string;
	code: string;
	name: string;
	status: "active" | "inactive";
	created_at: number;
	updated_at: number;
}

export interface Setting {
	id: string;
	tenant_id: string;
	scope_type: "tenant" | "branch" | "service" | "counter";
	scope_id: string;
	key: string;
	value: string;
	value_type: "string" | "number" | "boolean" | "json";
	is_active: boolean;
	created_at: number;
	updated_at: number;
}

export interface Queue {
	id: string;
	tenant_id: string;
	branch_id: string;
	queue_date: string;
	ticket_no: string;
	queue_no: number;
	patient_id?: string;
	patient_name?: string;
	status: string;
	current_journey_id?: string;
	created_at: number;
	updated_at: number;
}

export interface QueueJourney {
	id: string;
	queue_id: string;
	service_id: string;
	counter_id?: string;
	seq_no: number;
	status: string;
	created_at: number;
	updated_at: number;
}

export interface VisitJourney {
	id: string;
	queue_id: string;
	tenant_id: string;
	event_type: string;
	payload?: string;
	created_at: number;
}

export interface QueueStatsResponse {
	total_queues_today: number;
	total_active_journeys: number;
	total_completed_visits: number;
	waiting_by_service: Record<string, number>;
}

export interface ScannerCheckInResponse {
	action: "register" | "forward";
	queue: Queue;
}

// -----------------------------------------------------------------------------
// BRANCHES API
// -----------------------------------------------------------------------------
export const branchesApi = {
	getAll: () => api.get<{ data: Branch[] }>("/branches"),
	getById: (id: string) => api.get<{ data: Branch }>(`/branches/${id}`),
	create: (data: { code: string; name: string }) =>
		api.post<{ data: Branch }>("/branches", data),
	update: (
		id: string,
		data: { code?: string; name?: string; status?: "active" | "inactive" },
	) => api.put<{ data: Branch }>(`/branches/${id}`, data),
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
		is_pharmacy: boolean;
		is_pharmacy_reception: boolean;
	}) => api.post<{ data: Service }>("/services", data),
	update: (
		id: string,
		data: {
			code?: string;
			name?: string;
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
	create: (data: { branch_id: string; code: string; name: string }) =>
		api.post<{ data: Counter }>("/counters", data),
	update: (
		id: string,
		data: { code?: string; name?: string; status?: "active" | "inactive" },
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
		headers: { clientId: string; apiKey: string },
	) =>
		api.post<{ data: ScannerCheckInResponse }>("/scanner/check-in", data, {
			headers: {
				"X-Client-ID": headers.clientId,
				"X-API-Key": headers.apiKey,
			},
		}),
};

// -----------------------------------------------------------------------------
// SETTINGS API
// -----------------------------------------------------------------------------
export const settingsApi = {
	resolve: (params: {
		key: string;
		branch_id?: string;
		service_id?: string;
		counter_id?: string;
	}) => {
		const searchParams = new URLSearchParams();
		searchParams.append("key", params.key);
		if (params.branch_id) searchParams.append("branch_id", params.branch_id);
		if (params.service_id) searchParams.append("service_id", params.service_id);
		if (params.counter_id) searchParams.append("counter_id", params.counter_id);

		return api.get<{ data: Setting }>(
			`/settings/resolve?${searchParams.toString()}`,
		);
	},
	getById: (id: string) => api.get<{ data: Setting }>(`/settings/${id}`),
	create: (data: {
		scope_type: "tenant" | "branch" | "service" | "counter";
		scope_id: string;
		key: string;
		value: string;
		value_type?: "string" | "number" | "boolean" | "json";
	}) => api.post<{ data: Setting }>("/settings", data),
	update: (id: string, data: { value?: string; is_active?: boolean }) =>
		api.put<{ data: Setting }>(`/settings/${id}`, data),
	delete: (id: string) => api.delete(`/settings/${id}`),
};
