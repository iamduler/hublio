export type ExecutionStatus =
  | "created"
  | "queued"
  | "running"
  | "succeeded"
  | "failed"
  | "cancelled"
  | "expired"
  | "dead_letter";

export interface ExecutionStep {
  id: string;
  step_no: number;
  step_type: string;
  status: string;
  retry_attempt?: number;
  duration_ms?: number;
  error_message?: string;
  error_code?: string;
  started_at?: string;
  completed_at?: string;
}

export interface TimelineEntry {
  id?: string;
  event: string;
  message?: string;
  metadata?: Record<string, unknown>;
  created_at?: string;
}

export interface Execution {
  id: string;
  intent_id: string;
  status: ExecutionStatus;
  result?: string;
  retry_attempt?: number;
  current_step_no?: number;
  failure_reason?: string;
  steps?: ExecutionStep[];
  timeline?: TimelineEntry[];
  started_at?: string;
  completed_at?: string;
  created_at?: string;
}

export interface ExecutionTimeline {
  execution_id: string;
  status: string;
  timeline: TimelineEntry[];
}
