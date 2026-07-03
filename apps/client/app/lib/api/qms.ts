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

type SignageQueueResponse = {
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
};

export const qmsClientsApi = {
	create: (data: QMSClientCreateRequest) =>
		apiClient.post<QMSClientCreateResponse>("/qms-clients", data),
	createCredential: (data: QMSClientCredentialCreateRequest) =>
		apiClient.post<QMSClientCredentialCreateResponse>(
			"/qms-clients/credentials",
			data,
		),
};

type QMSClientHeaders = { clientId: string; apiKey: string };

function qmsClientHeaders(headers: QMSClientHeaders) {
	return {
		"X-Client-ID": headers.clientId,
		"X-API-Key": headers.apiKey,
	};
}

export const callerApi = {
	login: (data: CallerLoginRequest, headers: QMSClientHeaders) =>
		apiClient.post<CallerLoginResponse>("/caller/login", data, undefined, {
			headers: qmsClientHeaders(headers),
		}),
	me: (headers: QMSClientHeaders) =>
		apiClient.get<CallerMeResponse>("/caller/me", undefined, {
			headers: qmsClientHeaders(headers),
		}),
	action: (
		journeyId: string,
		data: CallerActionRequest,
		headers: QMSClientHeaders,
	) =>
		apiClient.post<CallerActionResponse>(
			`/caller/queue-journeys/${journeyId}/action`,
			data,
			undefined,
			{ headers: qmsClientHeaders(headers) },
		),
};

export const signageApi = {
	me: (headers: QMSClientHeaders) =>
		apiClient.get<SignageMeResponse>("/signage/me", undefined, {
			headers: qmsClientHeaders(headers),
		}),
	currentCalls: (headers: QMSClientHeaders) =>
		apiClient.get<SignageCurrentCallResponse[]>(
			"/signage/current-calls",
			undefined,
			{
				headers: qmsClientHeaders(headers),
			},
		),
	queues: (headers: QMSClientHeaders) =>
		apiClient.get<SignageQueueResponse[]>("/signage/queues", undefined, {
			headers: qmsClientHeaders(headers),
		}),
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
