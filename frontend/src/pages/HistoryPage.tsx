import { useEffect, useState } from 'react';
import { ArrowRight } from 'lucide-react';
import { Link } from 'react-router-dom';
import { getDashboardOverview } from '../api/dashboard';
import { listFindings } from '../api/findings';
import { ApiError } from '../api/client';
import { AppShell } from '../components/app-shell';
import { DetailCard } from '../components/ui/card';
import { useDatabaseInstance } from '../lib/databaseInstance';
import { formatRelativeTimeShort } from '../lib/format';
import type { DashboardOverviewResponse, IssueQueueItem } from '../types/issues';

export function HistoryPage() {
  const { selectedId, loading: instanceLoading } = useDatabaseInstance();
  const [overview, setOverview] = useState<DashboardOverviewResponse | null>(null);
  const [findings, setFindings] = useState<IssueQueueItem[]>([]);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (instanceLoading || !selectedId) return;

    Promise.all([
      getDashboardOverview(selectedId, '24h'),
      listFindings({ databaseInstanceId: selectedId, range: '7d', status: 'open', limit: 20 }),
    ])
      .then(([overviewResult, findingsResult]) => {
        setOverview(overviewResult);
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

  const timelineItems = [...(overview?.findings.recent ?? []), ...findings.slice(0, 5)]
    .filter((item, index, list) => list.findIndex((candidate) => candidate.id === item.id) === index)
    .slice(0, 8);

  return (
    <AppShell title="History" subtitle="Historical memory helps explain what changed, when it changed, and whether the situation is getting better.">
      <div className="space-y-6">
        {error ? <div className="rounded-xl border border-[var(--danger)] bg-[var(--danger-soft)] px-4 py-3 text-[13px] text-[var(--danger)]">{error}</div> : null}

        <DetailCard title="Investigation timeline" description="Use this page to reconstruct the sequence of change during an incident.">
          <div className="space-y-4">
            {timelineItems.map((finding) => (
              <div key={finding.id} className="flex gap-4 rounded-xl border border-[var(--border)] bg-[var(--panel)] px-4 py-4">
                <div className="min-w-[72px] text-[12px] font-medium text-[var(--muted-foreground)]">{formatRelativeTimeShort(finding.last_seen_at)}</div>
                <div className="min-w-0 flex-1">
                  <div className="text-[14px] font-semibold text-[var(--foreground)]">{finding.change_summary || finding.problem_summary || finding.title}</div>
                  <div className="mt-1 text-[13px] leading-6 text-[var(--muted-foreground)]">{finding.evidence_summary}</div>
                </div>
                <Link to={`/findings/${finding.id}`} className="inline-flex items-center gap-1 text-[13px] font-medium text-[var(--foreground)] no-underline">
                  Open
                  <ArrowRight size={14} />
                </Link>
              </div>
            ))}
          </div>
        </DetailCard>

        <DetailCard title="Advanced evidence" description="Raw metrics are available, but only as supporting proof after you understand the diagnosis timeline.">
          <div className="flex items-center justify-between gap-4 rounded-xl border border-[var(--border)] bg-[var(--muted)] px-4 py-4">
            <div>
              <div className="text-[14px] font-semibold text-[var(--foreground)]">Open raw metrics</div>
              <div className="mt-1 text-[13px] leading-6 text-[var(--muted-foreground)]">
                Use raw time-series only when you need to verify that a signal is truly recovering.
              </div>
            </div>
            <Link
              to="/metrics"
              className="inline-flex h-10 items-center gap-2 rounded-[10px] border border-[var(--border)] bg-[var(--panel)] px-4 text-[14px] font-medium text-[var(--foreground)] no-underline transition-colors hover:bg-[var(--muted)]"
            >
              Open metrics
              <ArrowRight size={14} />
            </Link>
          </div>
        </DetailCard>
      </div>
    </AppShell>
  );
}
