import { apiGet } from './client';
import type { DashboardOverview, MetricRange } from '../types/dashboard';

export interface DashboardOverviewParams {
  agentId?: string;
  databaseInstanceId?: string;
  range?: MetricRange;
}

export function getDashboardOverview(params: DashboardOverviewParams = {}): Promise<DashboardOverview> {
  return apiGet<DashboardOverview>('/api/dashboard/overview', {
    agent_id: params.agentId,
    database_instance_id: params.databaseInstanceId,
    range: params.range,
  });
}
