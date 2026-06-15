import { apiGet } from './client';
import type { QueryStatsResponse } from '../types/queries';

export function getQueryStats(databaseInstanceId: string): Promise<QueryStatsResponse> {
  return apiGet<QueryStatsResponse>('/api/queries', {
    database_instance_id: databaseInstanceId,
  });
}

