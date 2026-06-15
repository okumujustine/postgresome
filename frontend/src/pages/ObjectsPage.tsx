import { useEffect, useMemo, useState } from 'react';
import { ArrowUpRight } from 'lucide-react';
import { Link } from 'react-router-dom';
import { getTableStats } from '../api/tables';
import { listFindings } from '../api/findings';
import { ApiError } from '../api/client';
import { AppShell } from '../components/app-shell';
import { SeverityBadge } from '../components/status-badges';
import { DetailCard } from '../components/ui/card';
import { useDatabaseInstance } from '../lib/databaseInstance';
import { formatPercent, formatRelativeTime } from '../lib/format';
import type { IssueQueueItem } from '../types/issues';
import type { TableStat } from '../types/tables';

function deadRowRatio(table: TableStat) {
  const total = table.live_rows + table.dead_rows;
  if (total === 0) return 0;
  return (table.dead_rows / total) * 100;
}

function indexUsageRatio(table: TableStat) {
  const totalScans = table.index_scans + table.sequential_scans;
  if (totalScans === 0) return 0;
  return (table.index_scans / totalScans) * 100;
}

export function ObjectsPage() {
  const { selectedId, loading: instanceLoading } = useDatabaseInstance();
  const [tables, setTables] = useState<TableStat[]>([]);
  const [findings, setFindings] = useState<IssueQueueItem[]>([]);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (instanceLoading || !selectedId) return;

    Promise.all([
      getTableStats(selectedId),
      listFindings({ databaseInstanceId: selectedId, range: '7d', status: 'open', limit: 50 }),
    ])
      .then(([tablesResult, findingsResult]) => {
        setTables(tablesResult.tables);
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

  const findingsByTable = useMemo(() => {
    const grouped = new Map<string, IssueQueueItem[]>();
    for (const finding of findings) {
      if (finding.resource_type !== 'table') continue;
      const current = grouped.get(finding.resource_name) ?? [];
      current.push(finding);
      grouped.set(finding.resource_name, current);
    }
    return grouped;
  }, [findings]);

  const tableRows = useMemo(
    () =>
      [...tables].sort((a, b) => {
        const keyA = `${a.schema_name}.${a.table_name}`;
        const keyB = `${b.schema_name}.${b.table_name}`;
        const findingScoreA = findingsByTable.has(keyA) ? 0 : 1;
        const findingScoreB = findingsByTable.has(keyB) ? 0 : 1;
        if (findingScoreA !== findingScoreB) return findingScoreA - findingScoreB;
        return deadRowRatio(b) - deadRowRatio(a);
      }),
    [tables, findingsByTable],
  );

  return (
    <AppShell title="Database" subtitle="Tables and indexes health, with diagnosis first and object evidence second.">
      <div className="space-y-6">
        {error ? <div className="rounded-xl border border-[var(--danger)] bg-[var(--danger-soft)] px-4 py-3 text-[13px] text-[var(--danger)]">{error}</div> : null}

        <DetailCard title="Database objects health" description="Tables with the strongest evidence of bloat, vacuum drift, or index inefficiency are surfaced first.">
          <div className="space-y-4">
            {tableRows.slice(0, 8).map((table) => {
              const key = `${table.schema_name}.${table.table_name}`;
              const finding = findingsByTable.get(key)?.[0];
              return (
                <div key={key} className="flex items-start gap-4 rounded-xl border border-[var(--border)] bg-[var(--panel)] px-4 py-4">
                  <div className="min-w-0 flex-1">
                    <div className="flex flex-wrap items-center gap-2">
                      {finding ? <SeverityBadge severity={finding.severity} /> : null}
                      <div className="text-[14px] font-semibold text-[var(--foreground)]">{key}</div>
                    </div>
                    <div className="mt-2 text-[13px] leading-6 text-[var(--muted-foreground)]">
                      {finding
                        ? finding.problem_summary || finding.evidence_summary
                        : describeTableHealth(table)}
                    </div>
                  </div>
                  <div className="hidden min-w-[110px] md:block">
                    <div className="text-[12px] text-[var(--muted-foreground)]">Health</div>
                    <div className="mt-1 text-[14px] font-medium text-[var(--foreground)]">{tableHealthLabel(table, finding)}</div>
                  </div>
                  <div className="hidden min-w-[110px] md:block">
                    <div className="text-[12px] text-[var(--muted-foreground)]">Bloat risk</div>
                    <div className="mt-1 text-[14px] font-medium text-[var(--foreground)]">{formatPercent(deadRowRatio(table))}</div>
                  </div>
                  <div className="hidden min-w-[120px] lg:block">
                    <div className="text-[12px] text-[var(--muted-foreground)]">Vacuum status</div>
                    <div className="mt-1 text-[14px] font-medium text-[var(--foreground)]">
                      {table.last_autovacuum_at ? formatRelativeTime(table.last_autovacuum_at) : table.last_vacuum_at ? formatRelativeTime(table.last_vacuum_at) : 'No recent vacuum'}
                    </div>
                  </div>
                  {finding ? (
                    <Link to={`/findings/${finding.id}`} className="mt-1 inline-flex items-center gap-1 text-[13px] font-medium text-[var(--foreground)] no-underline">
                      Open
                      <ArrowUpRight size={13} />
                    </Link>
                  ) : null}
                </div>
              );
            })}
          </div>
        </DetailCard>
      </div>
    </AppShell>
  );
}

function describeTableHealth(table: TableStat) {
  const bloat = deadRowRatio(table);
  const indexUsage = indexUsageRatio(table);

  if (bloat > 20) {
    return 'Dead tuples are elevated for this table, which can point to table bloat risk and slower scans if maintenance is lagging.';
  }

  if (indexUsage < 50) {
    return 'This table is leaning on sequential scans more than indexes, which may indicate an inefficient access path or missing index.';
  }

  return 'No active diagnosis is attached, but the object is still being tracked for growth, vacuum cadence, and index efficiency.';
}

function tableHealthLabel(table: TableStat, finding?: IssueQueueItem) {
  if (finding?.severity === 'critical') return 'Needs action';
  if (finding?.severity === 'warning') return 'Watch';
  if (deadRowRatio(table) > 20 || indexUsageRatio(table) < 50) return 'Watch';
  return 'Healthy';
}
