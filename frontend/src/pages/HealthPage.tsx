import { useCallback, useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { ArrowRight, Search } from 'lucide-react';
import { listFindings } from '../api/findings';
import { getTableStats } from '../api/tables';
import { ApiError } from '../api/client';
import type { FindingsListResponse, MetricRange } from '../types/dashboard';
import type { TableStatsResponse } from '../types/tables';
import { Layout } from '../components/Layout';
import { StatusBadge } from '../components/StatusBadge';
import { HealthIssueCard } from '../components/HealthIssueCard';
import { formatRelativeTimeShort } from '../lib/format';
import { useDatabaseInstance } from '../lib/databaseInstance';

const NEEDS_ATTENTION_LIMIT = 5;

const SEVERITY_RANK: Record<string, number> = { critical: 0, warning: 1, info: 2 };

// Issues are an open-ended tracker, not a time-series view — use the widest
// available range so open issues aren't hidden by a short window (same
// rationale as ISSUES_RANGE in IssuesPage.tsx).
const HEALTH_FINDINGS_RANGE: MetricRange = '7d';

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
        listFindings({ status: 'open', range: HEALTH_FINDINGS_RANGE, limit: 50, databaseInstanceId }),
        getTableStats({ databaseInstanceId }),
      ]);
      setData(findingsResult);
      setTableStats(tableStatsResult);
      setError(null);
    } catch (err) {
      const message =
        err instanceof ApiError
          ? `The Postgresome API returned an error (${err.status}). Try refreshing.`
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
        <div className="flex flex-col gap-6">
          <section
            className="grid gap-4 border-b pb-5 md:grid-cols-[minmax(0,1.35fr)_minmax(260px,0.85fr)]"
            style={{ borderColor: 'var(--border-subtle)' }}
          >
            <div>
              <div className="mb-2 flex flex-wrap items-center gap-3">
                <h2
                  className="m-0 text-[var(--fs-h1)] font-semibold"
                  style={{ color: 'var(--text-primary)', letterSpacing: 'var(--ls-tight)', fontFamily: 'var(--font-mono)' }}
                >
                  {instance.database_name}
                </h2>
                <StatusBadge status={instance.status} />
              </div>
              <div className="max-w-[760px] text-[14px] leading-[1.6]" style={{ color: 'var(--text-secondary)' }}>
                {total === 0
                  ? 'No open issues are currently blocking this database. Keep monitoring for new findings.'
                  : `${total} open issue${total === 1 ? '' : 's'} need review. Start with the highest-severity findings, confirm the evidence, and then apply the recommended fix.`}
              </div>
              {total > 0 && (
                <div className="mt-2 text-[13px]" style={{ color: 'var(--text-muted)' }}>
                  {counts.critical} critical · {counts.warning} warning · {counts.info} informational
                </div>
              )}
              <div className="mt-2 text-[13px]" style={{ color: 'var(--text-muted)' }}>
                Last analyzed {tableStats?.collected_at ? formatRelativeTimeShort(tableStats.collected_at) : '—'}
                {instance.host && (
                  <>
                    {' · '}
                    <span style={{ fontFamily: 'var(--font-mono)' }}>{instance.host}</span>
                  </>
                )}
              </div>
            </div>

            <div
              className="rounded-[var(--radius-lg)] border px-4 py-3"
              style={{ borderColor: 'var(--border-subtle)', background: 'var(--surface-raised)' }}
            >
              <div className="text-[12px] font-medium" style={{ color: 'var(--text-muted)' }}>
                Next step
              </div>
              <div className="mt-1 text-[14px] leading-[1.55]" style={{ color: 'var(--text-secondary)' }}>
                {total === 0
                  ? 'No remediation work is queued right now.'
                  : counts.critical > 0
                    ? 'Review the critical issues first. They are the fastest path to reducing time to fix.'
                    : 'Review the warning issues below and confirm which ones are still affecting production.'}
              </div>
              {total > 0 && (
                <Link
                  to="/issues"
                  className="mt-3 inline-flex items-center gap-1.5 text-[13px] font-medium no-underline"
                  style={{ color: 'var(--text-link)' }}
                >
                  Open full issue queue
                  <ArrowRight size={14} />
                </Link>
              )}
            </div>
          </section>

          {needsAttention.length > 0 && (
            <section>
              <div className="mb-3">
                <h3 className="m-0 text-[18px] font-semibold" style={{ color: 'var(--text-primary)', letterSpacing: 'var(--ls-snug)' }}>
                  Priority issues
                </h3>
                <div className="mt-1 text-[13px]" style={{ color: 'var(--text-muted)' }}>
                  Each issue shows the problem, the evidence behind it, and the recommended fix.
                </div>
              </div>
              <div className="flex flex-col gap-3">
                {needsAttention.map((finding) => (
                  <HealthIssueCard key={finding.id} finding={finding} />
                ))}
              </div>
              {total > NEEDS_ATTENTION_LIMIT && (
                <Link
                  to="/issues"
                  className="mt-3 inline-flex items-center gap-1.5 text-[13px] font-medium no-underline"
                  style={{ color: 'var(--text-link)' }}
                >
                  View all {total} open issues
                  <ArrowRight size={14} />
                </Link>
              )}
            </section>
          )}
        </div>
      ) : null}
    </Layout>
  );
}
