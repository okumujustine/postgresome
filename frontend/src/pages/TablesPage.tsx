import { useCallback, useEffect, useState } from 'react';
import { Search } from 'lucide-react';
import { getTableStats } from '../api/tables';
import { ApiError } from '../api/client';
import type { TableStat, TableStatsResponse } from '../types/tables';
import { Layout } from '../components/Layout';
import { Card } from '../components/Card';
import { formatRelativeTime } from '../lib/format';
import { useDatabaseInstance } from '../lib/databaseInstance';

const DEAD_ROW_WARNING_RATIO = 20;

function deadRowRatio(table: TableStat): number {
  const total = table.live_rows + table.dead_rows;
  if (total === 0) return 0;
  return (table.dead_rows / total) * 100;
}

export function TablesPage() {
  const [data, setData] = useState<TableStatsResponse | null>(null);
  const [refreshing, setRefreshing] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const { selectedId, loading: instanceLoading } = useDatabaseInstance();

  const load = useCallback(async (databaseInstanceId: string, isRefresh: boolean) => {
    if (isRefresh) {
      setRefreshing(true);
    }

    try {
      const result = await getTableStats({ databaseInstanceId });
      setData(result);
      setError(null);
    } catch (err) {
      const message =
        err instanceof ApiError
          ? `The Postgresome API returned an error (${err.status}).`
          : 'Unable to reach the Postgresome API. Is it running?';
      setError(message);
    } finally {
      setRefreshing(false);
    }
  }, []);

  useEffect(() => {
    if (instanceLoading || !selectedId) return;
    // load() only updates state after its internal await, not synchronously.
    // eslint-disable-next-line react-hooks/set-state-in-effect
    load(selectedId, false);
  }, [selectedId, instanceLoading, load]);

  const loading = data === null && error === null;

  const tables = data?.tables ?? [];

  return (
    <Layout
      title="Tables"
      onRefresh={() => load(selectedId, true)}
      refreshing={refreshing}
    >
      {error && (
        <div
          className="mb-6 rounded-[var(--radius-lg)] border px-4 py-3 text-sm"
          style={{ borderColor: 'rgba(207,34,46,0.25)', background: 'var(--danger-tint)', color: 'var(--danger)' }}
        >
          {error}
        </div>
      )}

      {loading && !data ? (
        <div className="flex items-center gap-2 text-sm" style={{ color: 'var(--text-muted)' }}>
          <Search size={14} className="animate-pulse" />
          Loading tables…
        </div>
      ) : data ? (
        tables.length === 0 ? (
          <Card title="Tables">
            <div className="text-sm" style={{ color: 'var(--text-muted)' }}>
              No table statistics yet.
            </div>
          </Card>
        ) : (
          <Card
            title="Tables"
            subtitle={data.collected_at ? `Snapshot from ${formatRelativeTime(data.collected_at)}` : undefined}
          >
            <div className="overflow-x-auto">
              <table className="w-full text-left text-[13px]" style={{ borderCollapse: 'collapse' }}>
                <thead>
                  <tr style={{ borderBottom: '1px solid var(--border-subtle)' }}>
                    <th className="py-2 pr-4 font-medium" style={{ color: 'var(--text-muted)' }}>Table</th>
                    <th className="py-2 pr-4 font-medium" style={{ color: 'var(--text-muted)' }}>Live rows</th>
                    <th className="py-2 pr-4 font-medium" style={{ color: 'var(--text-muted)' }}>Dead rows</th>
                    <th className="py-2 pr-4 font-medium" style={{ color: 'var(--text-muted)' }}>Seq scans</th>
                    <th className="py-2 pr-4 font-medium" style={{ color: 'var(--text-muted)' }}>Vacuum status</th>
                  </tr>
                </thead>
                <tbody>
                  {tables.map((table) => {
                    const ratio = deadRowRatio(table);
                    const ratioHigh = ratio > DEAD_ROW_WARNING_RATIO;
                    const lastVacuum = table.last_autovacuum_at ?? table.last_vacuum_at;

                    return (
                      <tr key={`${table.schema_name}.${table.table_name}`} style={{ borderBottom: '1px solid var(--border-subtle)' }}>
                        <td className="py-2 pr-4" style={{ color: 'var(--text-primary)', fontFamily: 'var(--font-mono)' }}>
                          {table.schema_name}.{table.table_name}
                        </td>
                        <td className="py-2 pr-4 tabular" style={{ color: 'var(--text-primary)' }}>
                          {table.live_rows.toLocaleString()}
                        </td>
                        <td className="py-2 pr-4 tabular" style={{ color: ratioHigh ? 'var(--warning)' : 'var(--text-primary)' }}>
                          {table.dead_rows.toLocaleString()} ({ratio.toFixed(1)}%)
                        </td>
                        <td className="py-2 pr-4 tabular" style={{ color: 'var(--text-primary)' }}>
                          {table.sequential_scans.toLocaleString()}
                        </td>
                        <td
                          className="py-2 pr-4"
                          style={{ color: !lastVacuum && ratioHigh ? 'var(--warning)' : 'var(--text-secondary)' }}
                        >
                          {lastVacuum ? formatRelativeTime(lastVacuum) : 'Never'}
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>
          </Card>
        )
      ) : null}
    </Layout>
  );
}
