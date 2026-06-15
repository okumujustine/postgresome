import { useEffect, useMemo, useState } from 'react';
import { ArrowUpRight, ChevronRight } from 'lucide-react';
import { Link } from 'react-router-dom';
import { getQueryStats } from '../api/queries';
import { listFindings } from '../api/findings';
import { ApiError } from '../api/client';
import { AppShell } from '../components/app-shell';
import { SeverityBadge } from '../components/status-badges';
import { DetailCard } from '../components/ui/card';
import { useDatabaseInstance } from '../lib/databaseInstance';
import { formatBytes, formatDuration } from '../lib/format';
import type { IssueQueueItem } from '../types/issues';
import type { QueryStat } from '../types/queries';

export function QueriesPage() {
  const { selectedId, loading: instanceLoading } = useDatabaseInstance();
  const [queries, setQueries] = useState<QueryStat[]>([]);
  const [findings, setFindings] = useState<IssueQueueItem[]>([]);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (instanceLoading || !selectedId) return;

    Promise.all([
      getQueryStats(selectedId),
      listFindings({ databaseInstanceId: selectedId, range: '7d', status: 'open', category: 'queries', limit: 50 }),
    ])
      .then(([queryResult, findingsResult]) => {
        setQueries(queryResult.queries);
        setFindings(findingsResult.findings);
        setError(null);
      })
      .catch((err) => {
        const message =
          err instanceof ApiError
            ? `The Postgresome API returned an error (${err.status}).`
            : 'Unable to reach the Postgresome API. Is it running?';
        setError(message);
      });
  }, [selectedId, instanceLoading]);

  const findingsByQuery = useMemo(
    () => new Map(findings.filter((finding) => finding.resource_type === 'query').map((finding) => [finding.resource_name, finding] as const)),
    [findings],
  );

  const rows = useMemo(
    () =>
      [...queries].sort((a, b) => {
        const scoreA = findingsByQuery.has(a.query_id) ? 0 : 1;
        const scoreB = findingsByQuery.has(b.query_id) ? 0 : 1;
        if (scoreA !== scoreB) return scoreA - scoreB;
        return b.total_exec_time_ms - a.total_exec_time_ms;
      }),
    [queries, findingsByQuery],
  );

  return (
    <AppShell title="Queries" subtitle="Slow queries, regressions, and execution changes.">
      <div className="space-y-6">
        {error ? <div className="rounded-xl border border-[var(--danger)] bg-[var(--danger-soft)] px-4 py-3 text-[13px] text-[var(--danger)]">{error}</div> : null}

        <DetailCard title="Developer debugging experience" description="This page is for regressions, expensive queries, and new slow behavior. It is not a database table dump.">
          <div className="space-y-3">
            {rows.slice(0, 1).map((query) => {
              const finding = findingsByQuery.get(query.query_id);
              return (
                <div key={query.query_id} className="rounded-xl border border-[var(--border)] bg-[var(--muted)] px-5 py-5">
                  <div className="flex flex-wrap items-center gap-2">
                    {finding ? <SeverityBadge severity={finding.severity} /> : null}
                    <span className="text-[12px] font-medium text-[var(--muted-foreground)]">Primary query regression</span>
                  </div>
                  <code className="mt-3 block overflow-hidden text-ellipsis whitespace-nowrap text-[14px] text-[var(--foreground)]">{query.query}</code>
                  <p className="mt-3 text-[14px] leading-6 text-[var(--body)]">{finding?.evidence_summary || deriveQueryReason(query)}</p>
                  <div className="mt-4 flex flex-wrap gap-3 text-[13px] text-[var(--muted-foreground)]">
                    <span>Impact: {formatDuration(query.total_exec_time_ms)}</span>
                    <span>Change: {formatDuration(query.mean_exec_time_ms)} avg</span>
                    <span>Reason: {finding ? 'Diagnosis matched' : 'Behavioral outlier'}</span>
                  </div>
                </div>
              );
            })}
          </div>
        </DetailCard>

        <DetailCard title="Query evidence" description="Problem, change, and reason come first. Low-value detail stays out of the way until you need it.">
          <div className="divide-y divide-[var(--border)]">
            {rows.slice(0, 8).map((query) => {
              const finding = findingsByQuery.get(query.query_id);
              return (
                <div key={query.query_id} className="flex items-start gap-4 py-4">
                  <div className="min-w-0 flex-1">
                    <div className="flex flex-wrap items-center gap-2">
                      {finding ? <SeverityBadge severity={finding.severity} /> : null}
                      <span className="text-[12px] text-[var(--muted-foreground)]">
                        {query.user_name} on {query.database_name}
                      </span>
                    </div>
                    <code className="mt-2 block overflow-hidden text-ellipsis whitespace-nowrap text-[14px] text-[var(--foreground)]">{query.query}</code>
                    <div className="mt-2 text-[13px] leading-6 text-[var(--muted-foreground)]">
                      {finding?.evidence_summary || deriveQueryReason(query)}
                    </div>
                  </div>
                  <div className="hidden min-w-[130px] md:block">
                    <div className="text-[12px] text-[var(--muted-foreground)]">Impact</div>
                    <div className="mt-1 text-[14px] font-medium text-[var(--foreground)]">{formatDuration(query.total_exec_time_ms)}</div>
                  </div>
                  <div className="hidden min-w-[120px] md:block">
                    <div className="text-[12px] text-[var(--muted-foreground)]">Change</div>
                    <div className="mt-1 text-[14px] font-medium text-[var(--foreground)]">{formatDuration(query.mean_exec_time_ms)} avg</div>
                  </div>
                  <div className="hidden min-w-[160px] lg:block">
                    <div className="text-[12px] text-[var(--muted-foreground)]">Reason</div>
                    <div className="mt-1 text-[14px] font-medium text-[var(--foreground)]">
                      {finding ? 'Diagnosis matched' : query.shared_blocks_read > query.shared_blocks_hit ? 'Disk-heavy execution' : 'Execution outlier'}
                    </div>
                  </div>
                  {finding ? (
                    <Link to={`/findings/${finding.id}`} className="mt-1 inline-flex items-center gap-1 text-[13px] font-medium text-[var(--foreground)] no-underline">
                      Open
                      <ChevronRight size={14} />
                    </Link>
                  ) : (
                    <div className="mt-1 text-[12px] text-[var(--muted-helper)]">
                      <ArrowUpRight size={14} />
                    </div>
                  )}
                </div>
              );
            })}
          </div>
        </DetailCard>
      </div>
    </AppShell>
  );
}

function deriveQueryReason(query: QueryStat) {
  if (query.shared_blocks_read > query.shared_blocks_hit) {
    return `This query is reading ${formatBytes(query.shared_blocks_read)} from disk, which suggests cache misses or a wider scan path than usual.`;
  }

  if (query.mean_exec_time_ms > 250) {
    return `Average execution time is ${formatDuration(query.mean_exec_time_ms)}, which makes this a likely regression candidate during an incident.`;
  }

  return `This query accumulated ${query.calls.toLocaleString()} calls and ${formatDuration(query.total_exec_time_ms)} total execution time, enough to deserve review.`;
}
