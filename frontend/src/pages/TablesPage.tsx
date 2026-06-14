import { useCallback, useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { Search } from 'lucide-react';
import { getTableStats } from '../api/tables';
import { listFindings } from '../api/findings';
import { ApiError } from '../api/client';
import type { TableStat, TableStatsResponse } from '../types/tables';
import type { DashboardFinding, MetricRange } from '../types/dashboard';
import { Layout } from '../components/Layout';
import { Card } from '../components/Card';
import { StatusBadge } from '../components/StatusBadge';
import { formatRelativeTime } from '../lib/format';
import { useDatabaseInstance } from '../lib/databaseInstance';

const SEVERITY_RANK: Record<string, number> = { critical: 0, warning: 1, info: 2 };
const STATUS_RANK: Record<string, number> = { critical: 0, warning: 1, healthy: 2 };

// Issues are an open-ended tracker — use the widest available range so
// currently-open table findings aren't hidden by a short window (same
// rationale as ISSUES_RANGE in IssuesPage.tsx).
const FINDINGS_RANGE: MetricRange = '7d';
const FINDINGS_LIMIT = 200;

function deadRowRatio(table: TableStat): number {
  const total = table.live_rows + table.dead_rows;
  if (total === 0) return 0;
  return (table.dead_rows / total) * 100;
}

// indexUsageRatio returns the percentage of scans served by an index rather
// than a sequential scan, or null if the table has too few scans to judge.
function indexUsageRatio(table: TableStat): number | null {
  const totalScans = table.index_scans + table.sequential_scans;
  if (totalScans === 0) return null;
  return (table.index_scans / totalScans) * 100;
}

function tableStatus(findings: DashboardFinding[]): 'critical' | 'warning' | 'healthy' {
  if (findings.some((f) => f.severity === 'critical')) return 'critical';
  if (findings.some((f) => f.severity === 'warning')) return 'warning';
  return 'healthy';
}

function topFinding(findings: DashboardFinding[]): DashboardFinding | null {
  if (findings.length === 0) return null;
  return [...findings].sort((a, b) => {
    const rankDiff = (SEVERITY_RANK[a.severity] ?? 3) - (SEVERITY_RANK[b.severity] ?? 3);
    if (rankDiff !== 0) return rankDiff;
    return new Date(b.last_seen_at).getTime() - new Date(a.last_seen_at).getTime();
  })[0];
}

export function TablesPage() {
  const [data, setData] = useState<TableStatsResponse | null>(null);
  const [findings, setFindings] = useState<DashboardFinding[]>([]);
  const [refreshing, setRefreshing] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const { selectedId, loading: instanceLoading } = useDatabaseInstance();

  const load = useCallback(async (databaseInstanceId: string, isRefresh: boolean) => {
    if (isRefresh) {
      setRefreshing(true);
    }

    try {
      const [tableStatsResult, findingsResult] = await Promise.all([
        getTableStats({ databaseInstanceId }),
        listFindings({ status: 'open', range: FINDINGS_RANGE, limit: FINDINGS_LIMIT, databaseInstanceId }),
      ]);
      setData(tableStatsResult);
      setFindings(findingsResult.findings);
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

  const tables = data?.tables ?? [];

  const findingsByTable = new Map<string, DashboardFinding[]>();
  for (const finding of findings) {
    if (finding.resource_type !== 'table') continue;
    const list = findingsByTable.get(finding.resource_name) ?? [];
    list.push(finding);
    findingsByTable.set(finding.resource_name, list);
  }

  const sortedTables = tables
    .map((table) => {
      const key = `${table.schema_name}.${table.table_name}`;
      const tableFindings = findingsByTable.get(key) ?? [];
      return { table, key, status: tableStatus(tableFindings), finding: topFinding(tableFindings) };
    })
    .sort((a, b) => {
      const rankDiff = STATUS_RANK[a.status] - STATUS_RANK[b.status];
      if (rankDiff !== 0) return rankDiff;
      return a.key.localeCompare(b.key);
    });

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
              No table statistics yet — they&apos;ll appear once the Postgresome agent completes its next collection.
            </div>
          </Card>
        ) : (
          <div className="flex flex-col gap-4">
            <div className="border-b pb-4" style={{ borderColor: 'var(--border-subtle)' }}>
              <h2 className="m-0 text-[18px] font-semibold" style={{ color: 'var(--text-primary)', letterSpacing: 'var(--ls-snug)' }}>
                Table health
              </h2>
              <div className="mt-1 max-w-[780px] text-[13px] leading-[1.55]" style={{ color: 'var(--text-muted)' }}>
                Tables with open findings float to the top. Use this view to see which relations are unhealthy, what
                triggered the finding, and what fix should come next.
                {data.collected_at ? ` Snapshot from ${formatRelativeTime(data.collected_at)}.` : ''}
              </div>
            </div>

            <Card title="Table review">
              <div className="flex flex-col">
                {sortedTables.map(({ table, key, status, finding }, index) => {
                  const ratio = deadRowRatio(table);
                  const idxUsage = indexUsageRatio(table);

                  return (
                    <div key={key} className="py-3" style={{ borderTop: index === 0 ? 'none' : '1px solid var(--border-subtle)' }}>
                      <div className="flex flex-wrap items-center justify-between gap-3">
                        <div className="flex items-center gap-3">
                          <span className="text-sm" style={{ fontFamily: 'var(--font-mono)', color: 'var(--text-primary)' }}>
                            {key}
                          </span>
                          <StatusBadge status={status} size="sm" />
                        </div>
                        {finding && (
                          <Link
                            to={`/issues/${finding.id}`}
                            className="inline-flex h-[var(--control-h-sm)] shrink-0 items-center rounded-[var(--radius-md)] border px-3 text-[13px] font-medium no-underline"
                            style={{ background: 'var(--surface-raised)', color: 'var(--text-secondary)', borderColor: 'var(--border-default)' }}
                          >
                            Investigate
                          </Link>
                        )}
                      </div>

                      {finding && (
                        <div className="mt-2 flex flex-col gap-1 text-[13px]" style={{ color: 'var(--text-secondary)' }}>
                          <div>
                            <span className="font-medium" style={{ color: 'var(--text-primary)' }}>
                              Observed problem:{' '}
                            </span>
                            {finding.title}
                          </div>
                          {finding.recommendation && (
                            <div>
                              <span className="font-medium" style={{ color: 'var(--text-primary)' }}>
                                Recommended fix:{' '}
                              </span>
                              {finding.recommendation}
                            </div>
                          )}
                        </div>
                      )}

                      <div className="mt-2 text-xs tabular" style={{ color: 'var(--text-faint)' }}>
                        {table.live_rows.toLocaleString()} rows · {ratio.toFixed(1)}% dead · {idxUsage === null ? '—' : `${idxUsage.toFixed(1)}%`} index usage · {table.sequential_scans.toLocaleString()} seq scans
                      </div>
                    </div>
                  );
                })}
              </div>
            </Card>
          </div>
        )
      ) : null}
    </Layout>
  );
}
