import { apiGet } from './client';
import type { MetricQueryResponse, MetricRange } from '../types/dashboard';

export interface QueryMetricsParams {
  metricKey: string;
  range?: MetricRange;
  interval?: string;
  databaseInstanceId?: string;
  agentId?: string;
}

export function queryMetrics(params: QueryMetricsParams): Promise<MetricQueryResponse> {
  return apiGet<MetricQueryResponse>('/api/metrics/query', {
    metric_key: params.metricKey,
    range: params.range,
    interval: params.interval,
    database_instance_id: params.databaseInstanceId,
    agent_id: params.agentId,
  });
}
