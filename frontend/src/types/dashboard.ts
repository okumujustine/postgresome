export type MetricRange = '15m' | '1h' | '6h' | '24h' | '7d';

export type InstanceStatus = 'healthy' | 'warning' | 'critical' | 'unknown';

export interface DashboardInstance {
  id: string;
  database_name: string;
  host: string;
  source_id?: string;
  source_kind?: string;
  provider?: string;
  status: InstanceStatus | string;
}

export interface DatabaseInstancesResponse {
  database_instances: DashboardInstance[];
}

export interface DashboardMetric {
  value: number;
  unit: string;
  trend_percent: number;
}

export interface DashboardSummary {
  active_connections: DashboardMetric;
  cache_hit_ratio: DashboardMetric;
  rollback_rate: DashboardMetric;
  slow_queries: DashboardMetric;
}

export interface MetricQueryPoint {
  time: string;
  value: number;
}

export interface MetricQueryResponse {
  metric_key: string;
  range: string;
  interval: string;
  points: MetricQueryPoint[];
}
