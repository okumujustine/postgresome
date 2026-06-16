export interface DatabaseInstance {
  id: string;
  database_name: string;
  host: string;
  source_id?: string;
  source_kind?: string;
  provider?: string;
  status: string;
}

export interface Impact {
  code: string;
  label: string;
  summary: string;
}

export interface Action {
  code: string;
  label: string;
  summary: string;
}

export interface Finding {
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
  primary_impact: Impact;
  secondary_impacts?: Impact[];
  primary_action: Action;
  secondary_actions?: Action[];
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

export interface FindingListResponse {
  database_instance: DatabaseInstance;
  severity_counts: {
    critical: number;
    warning: number;
    info: number;
  };
  total: number;
  limit: number;
  offset: number;
  findings: Finding[];
}

export interface FindingEvidenceItem {
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

export interface FindingHistoricalContext {
  current_value: number;
  previous_value: number;
  baseline_value: number;
  change_percent: number;
  trend_window: string;
  baseline_label: string;
}

export interface FindingEvidencePoint {
  time: string;
  series: string;
  value: number;
}

export interface RelatedQuery {
  collected_at: string;
  query_id: string;
  database_name: string;
  user_name: string;
  query: string;
  calls: number;
  total_exec_time_ms: number;
  mean_exec_time_ms: number;
  min_exec_time_ms: number;
  max_exec_time_ms: number;
  rows_returned: number;
  shared_blocks_read: number;
  shared_blocks_hit: number;
}

export interface RelatedTable {
  collected_at: string;
  schema_name: string;
  table_name: string;
  live_rows: number;
  dead_rows: number;
  sequential_scans: number;
  sequential_rows_read: number;
  index_scans: number;
  index_rows_fetched: number;
  rows_inserted: number;
  rows_updated: number;
  rows_deleted: number;
  last_vacuum_at?: string;
  last_autovacuum_at?: string;
  last_analyze_at?: string;
  last_autoanalyze_at?: string;
}

export interface FindingDetailResponse {
  database_instance: DatabaseInstance;
  finding: Finding;
  evidence?: FindingEvidenceItem[];
  historical_context?: FindingHistoricalContext;
  evidence_points?: FindingEvidencePoint[];
  related_query?: RelatedQuery;
  related_table?: RelatedTable;
  alert_payload?: Record<string, unknown>;
}

export interface Source {
  id: string;
  kind: string;
  provider: string;
  name: string;
  configured: boolean;
  setup_state: string;
  last_check_status: string;
  last_check_started_at?: string;
  last_check_completed_at?: string;
  last_check_error?: string;
}

export interface SourceDatabase {
  id: string;
  name: string;
  host: string;
}

export interface SourceRecord {
  source: Source;
  database: SourceDatabase;
  instance: DatabaseInstance;
}

export interface ListSourcesResponse {
  sources: SourceRecord[];
}

export interface CreateSourceInput {
  source: {
    kind: string;
    provider: string;
    name: string;
  };
  database: {
    name: string;
    host: string;
  };
  connection: {
    uri: string;
  };
}

export interface RunCheckupResponse {
  status: string;
  database_instance_id: string;
  detected_at: string;
  warnings: string[];
  findings: Finding[];
}

export interface QueryStat {
  query_id: string;
  database_name: string;
  user_name: string;
  query: string;
  calls: number;
  total_exec_time_ms: number;
  mean_exec_time_ms: number;
  min_exec_time_ms: number;
  max_exec_time_ms: number;
  rows_returned: number;
  shared_blocks_read: number;
  shared_blocks_hit: number;
}

export interface QueryStatsResponse {
  database_instance: DatabaseInstance;
  collected_at?: string;
  queries: QueryStat[];
}

