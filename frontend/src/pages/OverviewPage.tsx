import { useCallback, useEffect, useState } from 'react';
import { Database, Search, Server } from 'lucide-react';
import { listFindings } from '../api/findings';
import { ApiError } from '../api/client';
import type { FindingsListResponse, MetricRange } from '../types/dashboard';
import { Layout } from '../components/Layout';
import { Card } from '../components/Card';
import { FindingCard } from '../components/FindingCard';
import { StatusBadge } from '../components/StatusBadge';
import { SeverityPill } from '../components/SeverityPill';
import { formatRelativeTime } from '../lib/format';

const OVERVIEW_FINDINGS_LIMIT = 10;
const SECTION_LIMIT = 5;

export function OverviewPage() {
  const [range, setRange] = useState<MetricRange>('1h');
  const [data, setData] = useState<FindingsListResponse | null>(null);
  const [refreshing, setRefreshing] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const load = useCallback(async (currentRange: MetricRange, isRefresh: boolean) => {
    if (isRefresh) {
      setRefreshing(true);
    }

    try {
      const result = await listFindings({ range: currentRange, limit: OVERVIEW_FINDINGS_LIMIT });
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
    // load() only updates state after its internal await, not synchronously.
    // eslint-disable-next-line react-hooks/set-state-in-effect
    load(range, false);
  }, [range, load]);

  const loading = data === null && error === null;

  const instance = data?.database_instance ?? { id: '', database_name: '', host: '', status: 'unknown' };
  const counts = data?.severity_counts ?? { critical: 0, warning: 0, info: 0 };
  const findings = data?.findings ?? [];
  const hasInstance = Boolean(data?.database_instance?.id);

  const criticalFindings = findings.filter((f) => f.severity === 'critical' || f.severity === 'warning').slice(0, SECTION_LIMIT);
  const recentFindings = findings.slice(0, SECTION_LIMIT);

  let statusSummary = 'All checks passing';
  if (counts.critical > 0) {
    statusSummary = `${counts.critical} critical issue${counts.critical === 1 ? '' : 's'} need attention`;
  } else if (counts.warning > 0) {
    statusSummary = `${counts.warning} warning${counts.warning === 1 ? '' : 's'} found`;
  }

  return (
    <Layout
      title="Overview"
      databaseName={hasInstance ? instance.database_name : undefined}
      status={instance.status}
      range={range}
      onRangeChange={setRange}
      onRefresh={() => load(range, true)}
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
          Loading overview…
        </div>
      ) : data ? (
        <div className="flex flex-col gap-5">
          <Card title="Database status">
            {hasInstance ? (
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
            ) : (
              <div className="text-sm" style={{ color: 'var(--text-muted)' }}>
                No database instance registered yet.
              </div>
            )}
          </Card>

          <Card
            title="Critical findings"
            subtitle={`${counts.critical} critical · ${counts.warning} warning`}
          >
            {criticalFindings.length === 0 ? (
              <div className="text-sm" style={{ color: 'var(--text-muted)' }}>
                No critical issues — nice work.
              </div>
            ) : (
              <div className="flex flex-col gap-3">
                {criticalFindings.map((finding) => (
                  <FindingCard key={finding.id} finding={finding} />
                ))}
              </div>
            )}
          </Card>

          <Card title="Recent changes">
            {recentFindings.length === 0 ? (
              <div className="text-sm" style={{ color: 'var(--text-muted)' }}>
                No recent activity.
              </div>
            ) : (
              <div className="flex flex-col gap-1">
                {recentFindings.map((finding) => (
                  <div key={finding.id} className="flex items-center gap-3 py-[6px]">
                    <SeverityPill severity={finding.severity} />
                    <span className="min-w-0 flex-1 truncate text-sm" style={{ color: 'var(--text-primary)' }}>
                      {finding.title}
                    </span>
                    <span className="shrink-0 text-xs" style={{ color: 'var(--text-faint)' }}>
                      {formatRelativeTime(finding.detected_at)}
                    </span>
                  </div>
                ))}
              </div>
            )}
          </Card>
        </div>
      ) : null}
    </Layout>
  );
}
