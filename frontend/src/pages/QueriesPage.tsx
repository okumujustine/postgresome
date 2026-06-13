import { useCallback, useEffect, useState } from 'react';
import { Search } from 'lucide-react';
import { getQueryStats } from '../api/queries';
import { ApiError } from '../api/client';
import type { QueryStatsResponse } from '../types/queries';
import { Layout } from '../components/Layout';
import { Card } from '../components/Card';
import { formatBytes, formatDuration, formatRelativeTime } from '../lib/format';
import { useDatabaseInstance } from '../lib/databaseInstance';

export function QueriesPage() {
  const [data, setData] = useState<QueryStatsResponse | null>(null);
  const [refreshing, setRefreshing] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const { selectedId, loading: instanceLoading } = useDatabaseInstance();

  const load = useCallback(async (databaseInstanceId: string, isRefresh: boolean) => {
    if (isRefresh) {
      setRefreshing(true);
    }

    try {
      const result = await getQueryStats({ databaseInstanceId });
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

  const queries = data?.queries ?? [];

  return (
    <Layout
      title="Queries"
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
          Loading queries…
        </div>
      ) : data ? (
        queries.length === 0 ? (
          <Card title="Queries">
            <div className="text-sm" style={{ color: 'var(--text-muted)' }}>
              No query statistics yet. Enable pg_stat_statements on the monitored database.
            </div>
          </Card>
        ) : (
          <Card
            title="Queries"
            subtitle={
              data.collected_at
                ? `Sorted by impact (calls × execution time) · Snapshot from ${formatRelativeTime(data.collected_at)}`
                : 'Sorted by impact (calls × execution time)'
            }
          >
            <div className="overflow-x-auto">
              <table className="w-full text-left text-[13px]" style={{ borderCollapse: 'collapse' }}>
                <thead>
                  <tr style={{ borderBottom: '1px solid var(--border-subtle)' }}>
                    <th className="py-2 pr-4 font-medium" style={{ color: 'var(--text-muted)' }}>Query</th>
                    <th className="py-2 pr-4 font-medium" style={{ color: 'var(--text-muted)' }}>Calls</th>
                    <th className="py-2 pr-4 font-medium" style={{ color: 'var(--text-muted)' }}>Average time</th>
                    <th className="py-2 pr-4 font-medium" style={{ color: 'var(--text-muted)' }}>Total impact</th>
                    <th className="py-2 pr-4 font-medium" style={{ color: 'var(--text-muted)' }}>Rows</th>
                    <th className="py-2 pr-4 font-medium" style={{ color: 'var(--text-muted)' }}>Disk usage</th>
                  </tr>
                </thead>
                <tbody>
                  {queries.map((q) => (
                    <tr key={q.query_id} style={{ borderBottom: '1px solid var(--border-subtle)' }}>
                      <td className="max-w-[480px] py-2 pr-4">
                        <code
                          title={q.query}
                          className="block overflow-hidden whitespace-nowrap"
                          style={{ textOverflow: 'ellipsis', color: 'var(--text-secondary)' }}
                        >
                          {q.query}
                        </code>
                      </td>
                      <td className="py-2 pr-4 tabular" style={{ color: 'var(--text-primary)' }}>
                        {q.calls.toLocaleString()}
                      </td>
                      <td className="py-2 pr-4 tabular" style={{ color: 'var(--text-primary)' }}>
                        {formatDuration(q.mean_exec_time_ms)}
                      </td>
                      <td className="py-2 pr-4 font-semibold tabular" style={{ color: 'var(--text-primary)' }}>
                        {formatDuration(q.total_exec_time_ms)}
                      </td>
                      <td className="py-2 pr-4 tabular" style={{ color: 'var(--text-primary)' }}>
                        {q.rows_returned.toLocaleString()}
                      </td>
                      <td className="py-2 pr-4 tabular" style={{ color: 'var(--text-primary)' }}>
                        {formatBytes(q.shared_blocks_read)}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </Card>
        )
      ) : null}
    </Layout>
  );
}
