import { apiGet } from './client';
import type { FindingsListResponse, MetricRange } from '../types/dashboard';

export interface ListFindingsParams {
  agentId?: string;
  databaseInstanceId?: string;
  range?: MetricRange;
  severity?: string;
  category?: string;
  status?: string;
  limit?: number;
  offset?: number;
}

export function listFindings(params: ListFindingsParams = {}): Promise<FindingsListResponse> {
  return apiGet<FindingsListResponse>('/api/findings', {
    agent_id: params.agentId,
    database_instance_id: params.databaseInstanceId,
    range: params.range,
    severity: params.severity,
    category: params.category,
    status: params.status,
    limit: params.limit !== undefined ? String(params.limit) : undefined,
    offset: params.offset !== undefined ? String(params.offset) : undefined,
  });
}
