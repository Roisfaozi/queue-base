import { z } from "zod";

// ── Primitives ──
export const emailSchema = z.string().trim().email().max(255);
export const timestampSchema = z.number().optional();

// ── Entities ──

export interface User {
	id: string;
	name: string;
	email: string;
	username?: string; // Optional for legacy support
	avatar_url?: string;
	avatarUrl?: string; // Legacy support
	picture?: string; // Legacy support
	role?: string;
	status?: string;
	created_at?: number;
	updated_at?: number;
}

export interface Role {
	id: string;
	name: string;
	description?: string;
	created_at?: number;
	updated_at?: number;
}

export interface Organization {
	id: string;
	name: string;
	slug: string;
	owner_id: string;
	status: string;
	settings?: Record<string, unknown>;
	created_at?: number;
	updated_at?: number;
}

export interface OrgMember {
	id: string;
	user_id: string;
	organization_id: string;
	role_id: string;
	status: string;
	joined_at?: number;
	user?: User;
}

export interface Project {
	id: string;
	name: string;
	slug?: string; // Optional for legacy support
	domain?: string; // Legacy support
	user_id?: string; // Legacy support
	description?: string;
	status: string;
	organization_id: string;
	created_at?: number;
	updated_at?: number;
}

export interface Resource {
	id: string;
	name: string;
	slug: string;
	description?: string;
	status?: string;
	created_at?: number;
	updated_at?: number;
}

export interface Endpoint {
	id: string;
	name: string;
	method: "GET" | "POST" | "PUT" | "PATCH" | "DELETE";
	path: string;
	resource_id: string;
	resource_name?: string;
	description?: string;
	auth_required?: boolean;
	status?: string;
	created_at?: number;
	updated_at?: number;
}

export interface AccessRight {
	id: string;
	name: string;
	resource: string;
	action: string;
	conditions?: Record<string, unknown>;
	created_at?: number;
	updated_at?: number;
}

export interface Permission {
	id: string;
	role_id: string;
	access_right_id: string;
	granted: boolean;
	role_name?: string;
	access_right_name?: string;
	created_at?: number;
	updated_at?: number;
}

// ── Requests & Responses ──

export interface LoginRequest {
	username: string;
	password: string;
}

export interface RegisterRequest {
	name: string;
	email: string;
	username: string;
	password: string;
}

export interface LoginResponse {
	access_token: string;
	refresh_token: string;
	expires_at: string;
	expires_in: number;
	token_type: string;
	user: User;
}

export interface TokenResponse {
	access_token: string;
	refresh_token: string;
	expires_at: string;
	expires_in: number;
	token_type: string;
}

export interface QMSClientCreateRequest {
	branch_id: string;
	client_type: "caller" | "signage" | "scanner" | "kiosk";
	name: string;
}

export interface QMSClientCreateResponse {
	id: string;
	tenant_id?: string;
	branch_id: string;
	client_type: "caller" | "signage" | "scanner" | "kiosk";
	name: string;
	branch_service_id?: string;
	counter_id?: string;
	is_active: boolean;
	created_at: number;
}

export interface QMSClientResponse extends QMSClientCreateResponse {}

export interface QMSClientUpdateRequest {
	name?: string;
	branch_service_id?: string;
	counter_id?: string;
	is_active?: boolean;
}

export interface QMSClientCredentialCreateRequest {
	client_id: string;
	api_key: string;
}

export interface QMSClientCredentialCreateResponse {
	id: string;
	client_id: string;
	expires_at?: number;
	created_at: number;
}

export interface BranchResponse {
	id: string;
	tenant_id: string;
	code: string;
	name: string;
	address?: string;
	city?: string;
	province?: string;
	postal_code?: string;
	phone?: string;
	email?: string;
	logo_asset_id?: string;
	running_text?: string;
	timezone?: string;
	status: "draft" | "active" | "inactive";
	created_at: number;
	updated_at: number;
}

export interface Service {
	id: string;
	tenant_id: string;
	code: string;
	name: string;
	type?: string;
	default_estimated_duration?: number;
	audio_id?: string;
	audio_en?: string;
	narrative_instruction_id?: string;
	narrative_instruction_en?: string;
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
	branch_service_id?: string;
	code: string;
	name: string;
	display_name?: string;
	status: "active" | "inactive";
	created_at: number;
	updated_at: number;
}

export interface BranchService {
	id: string;
	tenant_id: string;
	branch_id: string;
	service_id: string;
	custom_name?: string;
	is_active: boolean;
	sort_order: number;
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

export type BranchUpsertRequest = Partial<
	Pick<
		BranchResponse,
		| "code"
		| "name"
		| "address"
		| "city"
		| "province"
		| "postal_code"
		| "phone"
		| "email"
		| "logo_asset_id"
		| "running_text"
		| "timezone"
		| "status"
	>
>;

export const branchActivationSchema = z.object({
	address: z.string().trim().min(1),
	city: z.string().trim().min(1),
	province: z.string().trim().min(1),
	phone: z.string().trim().min(1),
	timezone: z.string().trim().min(1),
});

export interface CallerContext {
	tenant_id: string;
	tenant_name?: string;
	branch_id: string;
	branch_name?: string;
	branch_service_id?: string;
	service_name?: string;
	counter_id?: string;
	counter_name?: string;
	display_name?: string;
}

export interface CallerLoginRequest {
	username: string;
	password: string;
}

export interface CallerActionRequest {
	action: "call" | "serve" | "complete" | "skip" | "cancel";
}

export interface CallerActionResponse {
	success: boolean;
	track_no?: string;
	queue_no?: number;
	status: string;
	journey_id: string;
}

export interface CallerLoginResponse {
	access_token: string;
	context: CallerContext;
	permissions: string[];
}

export interface CallerMeResponse extends CallerLoginResponse {}

export interface SignageMeResponse {
	client_id: string;
	tenant_id: string;
	branch_id: string;
	branch_service_id?: string;
	counter_id?: string;
	client_type: string;
	name: string;
	running_text?: string;
	logo_asset_id?: string;
	branch_name?: string;
	service_name?: string;
	counter_display_name?: string;
	audio_id?: string;
	audio_en?: string;
	narrative_instruction_id?: string;
	narrative_instruction_en?: string;
}

export interface SignageCurrentCallResponse {
	queue_id: string;
	ticket_no: string;
	counter_id: string;
	counter_display_name?: string;
	service_id: string;
	service_type?: string;
	audio_id?: string;
	audio_en?: string;
	narrative_instruction_id?: string;
	narrative_instruction_en?: string;
}

export interface EffectiveQueueConfigResponse {
	tenant_id: string;
	branch_id?: string;
	service_id?: string;
	counter_id?: string;
	tenant?: {
		tenant_id: string;
	};
	branch?: {
		branch_id?: string;
		effective_logo_asset_id?: string;
	};
	queue_reset_time: string;
	queue_reset_time_source?: string;
	queue_reset_time_inherited?: boolean;
	ticket_prefix: string;
	ticket_prefix_source?: string;
	ticket_prefix_inherited?: boolean;
	numbering_strategy: string;
	numbering_strategy_source?: string;
	numbering_strategy_inherited?: boolean;
	default_estimated_duration?: string;
	default_estimated_duration_source?: string;
	default_estimated_duration_inherited?: boolean;
	allow_forward?: boolean;
	allow_skip?: boolean;
	allow_recall?: boolean;
	allow_cancel?: boolean;
	auto_call_next?: boolean;
	max_service_duration?: number;
	min_service_duration?: number;
	require_counter?: boolean;
	allow_forward_from?: boolean;
	allow_forward_to?: boolean;
	audio_id?: string;
	audio_en?: string;
	narrative_instruction_id?: string;
	narrative_instruction_en?: string;
	effective_until?: string;
}

export interface PaginatedResponse<T> {
	data: T[];
	meta: { total: number };
}

// ── Dashboard & Stats ──

export interface DashboardStats {
	total_users: number;
	total_roles: number;
	total_org_members: number;
	total_audit_logs: number;
}

export interface ActivityMetric {
	date: string;
	logins: number;
	audits: number;
}

export interface SystemInsight {
	avg_latency_ms: number;
	error_rate: number;
	uptime_percent: number;
	most_active_role: string;
}

// ── Zod Schemas ──

export const userSchema = z.object({
	id: z.string(),
	name: z.string().trim().min(1).max(100),
	email: emailSchema,
	username: z.string().trim().min(1).max(50),
	avatar_url: z.string().url().optional(),
	role: z.string().optional(),
	status: z.string().optional(),
	created_at: timestampSchema,
	updated_at: timestampSchema,
});

export const roleSchema = z.object({
	id: z.string(),
	name: z.string().trim().min(1).max(100),
	description: z.string().max(500).optional(),
	created_at: timestampSchema,
	updated_at: timestampSchema,
});

export const organizationSchema = z.object({
	id: z.string(),
	name: z.string().trim().min(1).max(200),
	slug: z.string(),
	owner_id: z.string(),
	status: z.string(),
	settings: z.record(z.unknown()).optional(),
	created_at: timestampSchema,
	updated_at: timestampSchema,
});

export const orgMemberSchema = z.object({
	id: z.string(),
	user_id: z.string(),
	organization_id: z.string(),
	role_id: z.string(),
	status: z.string(),
	joined_at: timestampSchema,
	user: z.lazy(() => userSchema).optional(),
});

export const projectSchema = z.object({
	id: z.string(),
	name: z.string().trim().min(1).max(200),
	slug: z.string(),
	description: z.string().max(1000).optional(),
	status: z.string(),
	organization_id: z.string(),
	created_at: timestampSchema,
	updated_at: timestampSchema,
});

export const resourceSchema = z.object({
	id: z.string(),
	name: z.string().trim().min(1).max(100),
	slug: z.string().trim().min(1).max(100),
	description: z.string().max(500).optional(),
	status: z.string().optional(),
	created_at: timestampSchema,
	updated_at: timestampSchema,
});

export const endpointSchema = z.object({
	id: z.string(),
	name: z.string().trim().min(1).max(200),
	method: z.enum(["GET", "POST", "PUT", "PATCH", "DELETE"]),
	path: z.string().trim().min(1).max(500),
	resource_id: z.string(),
	resource_name: z.string().optional(),
	description: z.string().max(500).optional(),
	auth_required: z.boolean().optional(),
	status: z.string().optional(),
	created_at: timestampSchema,
	updated_at: timestampSchema,
});

export const accessRightSchema = z.object({
	id: z.string(),
	name: z.string(),
	resource: z.string(),
	action: z.string(),
	conditions: z.record(z.unknown()).optional(),
	created_at: timestampSchema,
	updated_at: timestampSchema,
});

export const permissionSchema = z.object({
	id: z.string(),
	role_id: z.string(),
	access_right_id: z.string(),
	granted: z.boolean(),
	role_name: z.string().optional(),
	access_right_name: z.string().optional(),
	created_at: timestampSchema,
	updated_at: timestampSchema,
});

export const loginRequestSchema = z.object({
	username: z.string().trim().min(1, "Username is required"),
	password: z.string().min(1, "Password is required"),
});

export const registerRequestSchema = z.object({
	name: z.string().trim().min(1, "Name is required").max(100),
	email: emailSchema,
	username: z.string().trim().min(3, "Min 3 characters").max(50),
	password: z.string().min(8, "Min 8 characters"),
});

export const loginResponseSchema = z.object({
	access_token: z.string(),
	refresh_token: z.string(),
	expires_at: z.string(),
	expires_in: z.number(),
	token_type: z.string(),
	user: userSchema,
});

export const tokenResponseSchema = z.object({
	access_token: z.string(),
	refresh_token: z.string(),
	expires_at: z.string(),
	expires_in: z.number(),
	token_type: z.string(),
});

export function paginatedSchema(itemSchema: z.ZodTypeAny) {
	return z.object({
		data: z.array(itemSchema),
		meta: z.object({ total: z.number() }),
	});
}

export const dashboardStatsSchema = z.object({
	total_users: z.number(),
	total_roles: z.number(),
	total_org_members: z.number(),
	total_audit_logs: z.number(),
});

export const activityMetricSchema = z.object({
	date: z.string(),
	logins: z.number(),
	audits: z.number(),
});

export const systemInsightSchema = z.object({
	avg_latency_ms: z.number(),
	error_rate: z.number(),
	uptime_percent: z.number(),
	most_active_role: z.string(),
});

export * from "./qms/operator";
