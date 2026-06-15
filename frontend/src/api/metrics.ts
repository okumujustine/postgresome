import { apiGet } from './client';
import type { MetricQueryResponse, MetricRange } from '../types/dashboard';

export function queryMetric(metricKey: string, databaseInstanceId: string, range: MetricRange = '24h') {
  return apiGet<MetricQueryResponse>('/api/metrics/query', {
    metric_key: metricKey,
    database_instance_id: databaseInstanceId,
    range,
  });
}

