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
