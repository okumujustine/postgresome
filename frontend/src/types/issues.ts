import type { DashboardInstance } from './dashboard';

export interface IssueQueueItem {
  id: string;
  severity: string;
  category: string;
  title: string;
  message: string;
  recommendation: string;
  status: string;
  problem_summary: string;
  evidence_summary: string;
  impact_summary: string;
  suggested_action: string;
  confidence_label: string;
  confidence_score: number;
  baseline_value: number;
  baseline_label: string;
  change_summary: string;
  verification_hint: string;
  verification_status: string;
  verification_summary: string;
  regression_count: number;
  last_regressed_at?: string;
  improving_since?: string;
  verified_fixed_at?: string;
  rule_key: string;
  resource_type: string;
  resource_name: string;
  current_value: number;
  threshold_value: number;
  occurrence_count: number;
  first_seen_at: string;
  last_seen_at: string;
  detected_at: string;
}

export interface IssueQueueResponse {
  database_instance: DashboardInstance;
  severity_counts: { critical: number; warning: number; info: number };
  total: number;
  limit: number;
  offset: number;
  findings: IssueQueueItem[];
}

export interface IssueHistoricalContext {
  current_value: number;
  previous_value: number;
  baseline_value: number;
  change_percent: number;
  trend_window: string;
  baseline_label: string;
}

export interface IssueEvidencePoint {
  time: string;
  series: string;
  value: number;
}

export interface IssueEvidenceItem {
  id: string;
  observed_at: string;
  evidence_type: string;
  role: string;
  label: string;
  summary: string;
  metric_key: string;
  reference_id: string;
  current_value: number;
  baseline_value: number;
  change_percent: number;
  confidence_score: number;
  metadata?: Record<string, unknown>;
}

export interface IssueDetailResponse {
  database_instance: DashboardInstance;
  finding: IssueQueueItem;
  evidence?: IssueEvidenceItem[];
  historical_context?: IssueHistoricalContext;
  evidence_points?: IssueEvidencePoint[];
}

export interface DashboardOverviewResponse {
  database_instance: DashboardInstance;
  summary: {
    active_connections: { value: number; unit: string; trend_percent: number };
    cache_hit_ratio: { value: number; unit: string; trend_percent: number };
    rollback_rate: { value: number; unit: string; trend_percent: number };
    slow_queries: { value: number; unit: string; trend_percent: number };
  };
  findings: {
    critical: number;
    warning: number;
    info: number;
    recent: IssueQueueItem[];
  };
}
