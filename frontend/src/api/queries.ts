import { apiGet } from './client';
import type { QueryStatsResponse } from '../types/queries';

export interface GetQueryStatsParams {
  agentId?: string;
  databaseInstanceId?: string;
}

export function getQueryStats(params: GetQueryStatsParams = {}): Promise<QueryStatsResponse> {
  return apiGet<QueryStatsResponse>('/api/queries', {
    agent_id: params.agentId,
    database_instance_id: params.databaseInstanceId,
  });
}
