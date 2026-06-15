import type { DashboardInstance } from './dashboard';

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
  database_instance: DashboardInstance;
  collected_at: string | null;
  queries: QueryStat[];
}

