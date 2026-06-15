import { apiGet } from './client';
import type { DashboardOverviewResponse } from '../types/issues';
import type { MetricRange } from '../types/dashboard';

export function getDashboardOverview(databaseInstanceId: string, range: MetricRange = '24h') {
  return apiGet<DashboardOverviewResponse>('/api/dashboard/overview', {
    database_instance_id: databaseInstanceId,
    range,
  });
}

