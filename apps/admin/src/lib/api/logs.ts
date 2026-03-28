import { apiClient } from './client';

// ─── Types ────────────────────────────────────────────────────────────────────

export interface LogEntry {
  id: string;
  actor_id: string;
  actor_role: string;
  actor_name?: string; // Only visible to admin
  action: string;
  description?: string; // Human-readable description
  entity_type: string;
  entity_id: string;
  metadata?: Record<string, unknown>;
  ip_address?: string;
  created_at: string;
}

export interface LogsListResponse {
  data: LogEntry[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}

export interface LogFilters {
  action?: string;
  entity_type?: string;
  start_date?: string;
  end_date?: string;
  page?: number;
  page_size?: number;
}

// ─── API Functions ────────────────────────────────────────────────────────────

export async function getLogs(
  filters: LogFilters = {}
): Promise<LogsListResponse> {
  const params = new URLSearchParams();
  if (filters.action) params.set('action', filters.action);
  if (filters.entity_type) params.set('entity_type', filters.entity_type);
  if (filters.start_date) params.set('start_date', filters.start_date);
  if (filters.end_date) params.set('end_date', filters.end_date);
  if (filters.page) params.set('page', String(filters.page));
  if (filters.page_size) params.set('page_size', String(filters.page_size));

  const response = await apiClient.get<{ data: LogsListResponse }>(
    `/logs?${params.toString()}`
  );
  return response.data.data;
}

export async function getLogDetail(id: string): Promise<LogEntry> {
  const response = await apiClient.get<{ data: LogEntry }>(`/logs/${id}`);
  return response.data.data;
}
