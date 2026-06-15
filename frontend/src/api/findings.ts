import { apiGet } from './client';
import type { MetricRange } from '../types/dashboard';
import type { IssueDetailResponse, IssueQueueResponse } from '../types/issues';

export interface ListFindingsParams {
  databaseInstanceId: string;
  range?: MetricRange;
  severity?: string;
  category?: string;
  status?: string;
  limit?: number;
  offset?: number;
}

export function listFindings(params: ListFindingsParams): Promise<IssueQueueResponse> {
  return apiGet<IssueQueueResponse>('/api/findings', {
    database_instance_id: params.databaseInstanceId,
    range: params.range,
    severity: params.severity,
    category: params.category,
    status: params.status,
    limit: params.limit !== undefined ? String(params.limit) : undefined,
    offset: params.offset !== undefined ? String(params.offset) : undefined,
  });
}

export function getFinding(id: string, databaseInstanceId: string): Promise<IssueDetailResponse> {
  return apiGet<IssueDetailResponse>(`/api/findings/${id}`, {
    database_instance_id: databaseInstanceId,
  });
}

