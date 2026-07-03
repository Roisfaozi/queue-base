import { apiClient } from "./client";
import type {
	CallerActionRequest,
	CallerActionResponse,
	CallerLoginRequest,
	CallerLoginResponse,
	CallerMeResponse,
	EffectiveQueueConfigResponse,
	QMSClientCreateRequest,
	QMSClientCreateResponse,
	QMSClientCredentialCreateRequest,
	QMSClientCredentialCreateResponse,
	SignageCurrentCallResponse,
	SignageMeResponse,
} from "@casbin/api-types";

export const qmsClientsApi = {
	create: (data: QMSClientCreateRequest) =>
		apiClient.post<QMSClientCreateResponse>("/qms-clients", data),
	createCredential: (data: QMSClientCredentialCreateRequest) =>
		apiClient.post<QMSClientCredentialCreateResponse>(
			"/qms-clients/credentials",
			data,
		),
};

export const callerApi = {
	login: (data: CallerLoginRequest) =>
		apiClient.post<CallerLoginResponse>("/caller/login", data),
	me: () => apiClient.get<CallerMeResponse>("/caller/me"),
	action: (journeyId: string, data: CallerActionRequest) =>
		apiClient.post<CallerActionResponse>(
			`/caller/queue-journeys/${journeyId}/action`,
			data,
		),
};

export const signageApi = {
	me: () => apiClient.get<SignageMeResponse>("/signage/me"),
	currentCalls: () =>
		apiClient.get<SignageCurrentCallResponse[]>("/signage/current-calls"),
	queues: () => apiClient.get<unknown[]>("/signage/queues"),
};

export const qmsSettingsApi = {
	effective: (params?: {
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
		return apiClient.get<EffectiveQueueConfigResponse>(
			`/settings/effective${query ? `?${query}` : ""}`,
		);
	},
};
