export type MetricRange = '15m' | '1h' | '6h' | '24h' | '7d';

export type InstanceStatus = 'healthy' | 'warning' | 'critical' | 'unknown';

export interface DashboardInstance {
  id: string;
  database_name: string;
  host: string;
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

export interface DashboardFinding {
  id: string;
  severity: string;
  category: string;
  title: string;
  message: string;
  recommendation: string;
  detected_at: string;
}

export interface DashboardFindings {
  critical: number;
  warning: number;
  info: number;
  recent: DashboardFinding[];
}

export interface DashboardOverview {
  database_instance: DashboardInstance;
  summary: DashboardSummary;
  findings: DashboardFindings;
}

export interface FindingsListResponse {
  database_instance: DashboardInstance;
  severity_counts: { critical: number; warning: number; info: number };
  total: number;
  limit: number;
  offset: number;
  findings: DashboardFinding[];
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
