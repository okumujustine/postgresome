import { apiGet } from './client';
import type { TableStatsResponse } from '../types/tables';

export function getTableStats(databaseInstanceId: string): Promise<TableStatsResponse> {
  return apiGet<TableStatsResponse>('/api/tables', {
    database_instance_id: databaseInstanceId,
  });
}

