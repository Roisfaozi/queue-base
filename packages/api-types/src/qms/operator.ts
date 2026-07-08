export interface OperatorAssignmentRequest {
  branch_id: string;
  user_id: string;
  counter_id: string;
}

export interface OperatorAssignmentResponse {
  id: string;
  tenant_id?: string;
  branch_id: string;
  user_id: string;
  counter_id: string;
  assigned_at: number;
  unassigned_at?: number;
}
