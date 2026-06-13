import { useCallback, useEffect, useState } from 'react';
import { AlertTriangle, Clock, Database, OctagonAlert, Search, Server } from 'lucide-react';
import { listFindings } from '../api/findings';
import { getTableStats } from '../api/tables';
import { ApiError } from '../api/client';
import type { FindingsListResponse } from '../types/dashboard';
import type { TableStatsResponse } from '../types/tables';
import { Layout } from '../components/Layout';
import { Card } from '../components/Card';
import { MetricCard } from '../components/MetricCard';
import { StatusBadge } from '../components/StatusBadge';
import { HealthIssueCard } from '../components/HealthIssueCard';
import { formatRelativeTimeShort } from '../lib/format';
import { useDatabaseInstance } from '../lib/databaseInstance';

const NEEDS_ATTENTION_LIMIT = 5;

const SEVERITY_RANK: Record<string, number> = { critical: 0, warning: 1, info: 2 };

export function HealthPage() {
  const [data, setData] = useState<FindingsListResponse | null>(null);
  const [tableStats, setTableStats] = useState<TableStatsResponse | null>(null);
  const [refreshing, setRefreshing] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const { selectedId, loading: instanceLoading } = useDatabaseInstance();

  const load = useCallback(async (databaseInstanceId: string, isRefresh: boolean) => {
    if (isRefresh) {
      setRefreshing(true);
    }

    try {
      const [findingsResult, tableStatsResult] = await Promise.all([
        listFindings({ status: 'open', limit: 50, databaseInstanceId }),
        getTableStats({ databaseInstanceId }),
      ]);
      setData(findingsResult);
      setTableStats(tableStatsResult);
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

  const instance = data?.database_instance ?? { id: '', database_name: '', host: '', status: 'unknown' };
  const counts = data?.severity_counts ?? { critical: 0, warning: 0, info: 0 };
  const total = data?.total ?? 0;

  const needsAttention = [...(data?.findings ?? [])]
    .sort((a, b) => {
      const rankDiff = (SEVERITY_RANK[a.severity] ?? 3) - (SEVERITY_RANK[b.severity] ?? 3);
      if (rankDiff !== 0) return rankDiff;
      return new Date(b.last_seen_at).getTime() - new Date(a.last_seen_at).getTime();
    })
    .slice(0, NEEDS_ATTENTION_LIMIT);

  let statusSummary = 'All checks passing';
  if (counts.critical > 0) {
    statusSummary = `${counts.critical} critical issue${counts.critical === 1 ? '' : 's'} need attention`;
  } else if (counts.warning > 0) {
    statusSummary = `${counts.warning} warning${counts.warning === 1 ? '' : 's'} found`;
  }

  return (
    <Layout title="Health" onRefresh={() => load(selectedId, true)} refreshing={refreshing}>
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
          Checking database health…
        </div>
      ) : data ? (
        <div className="flex flex-col gap-5">
          <Card title="Database status">
            <div className="flex flex-col gap-3">
              <div className="flex flex-wrap items-center gap-3">
                <h2
                  className="m-0 text-[var(--fs-h1)] font-semibold"
                  style={{ color: 'var(--text-primary)', letterSpacing: 'var(--ls-tight)', fontFamily: 'var(--font-mono)' }}
                >
                  {instance.database_name}
                </h2>
                <StatusBadge status={instance.status} />
              </div>
              <div className="flex flex-wrap items-center gap-4">
                <span className="inline-flex items-center gap-[6px] text-[12.5px]" style={{ color: 'var(--text-muted)' }}>
                  <Server size={13} />
                  <span style={{ fontFamily: 'var(--font-mono)' }}>{instance.host}</span>
                </span>
                <span className="inline-flex items-center gap-[6px] text-[12.5px]" style={{ color: 'var(--text-muted)' }}>
                  <Database size={13} />
                  <span style={{ fontFamily: 'var(--font-mono)' }}>{instance.id}</span>
                </span>
              </div>
              <div className="text-sm" style={{ color: 'var(--text-secondary)' }}>
                {statusSummary}
              </div>
            </div>
          </Card>

          <div className="grid gap-[14px]" style={{ gridTemplateColumns: 'repeat(auto-fit, minmax(180px, 1fr))' }}>
            <MetricCard label="Open issues" value={total.toString()} icon={<AlertTriangle size={14} />} hideFooter />
            <MetricCard label="Critical" value={counts.critical.toString()} icon={<OctagonAlert size={14} />} hideFooter />
            <MetricCard label="Warnings" value={counts.warning.toString()} icon={<AlertTriangle size={14} />} hideFooter />
            <MetricCard
              label="Last checked"
              value={tableStats?.collected_at ? formatRelativeTimeShort(tableStats.collected_at) : '—'}
              icon={<Clock size={14} />}
              hideFooter
            />
          </div>

          <Card title="Needs attention" subtitle={total === 0 ? undefined : `${total} open issue${total === 1 ? '' : 's'}`}>
            {needsAttention.length === 0 ? (
              <div className="text-sm" style={{ color: 'var(--text-muted)' }}>
                No issues need attention — all checks passing.
              </div>
            ) : (
              <div className="flex flex-col gap-3">
                {needsAttention.map((finding) => (
                  <HealthIssueCard key={finding.id} finding={finding} />
                ))}
              </div>
            )}
          </Card>
        </div>
      ) : null}
    </Layout>
  );
}
