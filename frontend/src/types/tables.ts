import type { DashboardInstance } from './dashboard';

export interface TableStat {
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
  last_vacuum_at: string | null;
  last_autovacuum_at: string | null;
  last_analyze_at: string | null;
  last_autoanalyze_at: string | null;
}

export interface TableStatsResponse {
  database_instance: DashboardInstance;
  collected_at: string | null;
  tables: TableStat[];
}
