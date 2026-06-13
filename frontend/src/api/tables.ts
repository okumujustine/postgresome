import { apiGet } from './client';
import type { TableStatsResponse } from '../types/tables';

export interface GetTableStatsParams {
  agentId?: string;
  databaseInstanceId?: string;
}

export function getTableStats(params: GetTableStatsParams = {}): Promise<TableStatsResponse> {
  return apiGet<TableStatsResponse>('/api/tables', {
    agent_id: params.agentId,
    database_instance_id: params.databaseInstanceId,
  });
}
