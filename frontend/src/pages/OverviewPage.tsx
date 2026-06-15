import { useEffect, useMemo, useState } from 'react';
import { ArrowRight, ChevronRight } from 'lucide-react';
import { Link } from 'react-router-dom';
import { getDashboardOverview } from '../api/dashboard';
import { listFindings } from '../api/findings';
import { ApiError } from '../api/client';
import { useDatabaseInstance } from '../lib/databaseInstance';
import { formatRelativeTimeShort } from '../lib/format';
import type { DashboardOverviewResponse, IssueQueueItem } from '../types/issues';
import { AppShell } from '../components/app-shell';

export function OverviewPage() {
  const { selectedId, loading: instanceLoading } = useDatabaseInstance();
  const [overview, setOverview] = useState<DashboardOverviewResponse | null>(null);
  const [allFindings, setAllFindings] = useState<IssueQueueItem[]>([]);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (instanceLoading || !selectedId) return;

    Promise.all([
      getDashboardOverview(selectedId, '24h'),
      listFindings({ databaseInstanceId: selectedId, status: 'open', range: '7d', limit: 12 }),
    ])
      .then(([overviewResult, findingsResult]) => {
        setOverview(overviewResult);
        setAllFindings(findingsResult.findings);
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

  const primaryFinding = allFindings[0] ?? overview?.findings.recent[0] ?? null;
  const secondaryFindings = allFindings.slice(1, 4);
  const affectedScope = useMemo(() => scopeLabel(primaryFinding), [primaryFinding]);

  return (
    <AppShell title="Overview" subtitle="What needs your attention first.">
      <div className="space-y-6">
        {error ? <ErrorBanner message={error} /> : null}

        <section className="rounded-xl border border-[var(--border)] bg-[var(--panel)] px-8 py-8">
          <div className="max-w-3xl">
            <div className="mb-4">
              <SeverityLabel severity={primaryFinding?.severity ?? 'info'} />
            </div>
            <h2 className="text-[28px] font-semibold tracking-[-0.03em] text-[var(--foreground)]">
              {primaryFinding?.problem_summary || primaryFinding?.title || 'No active diagnosis'}
            </h2>
            <p className="mt-4 text-[15px] leading-7 text-[var(--body)]">
              {primaryFinding
                ? primaryFinding.impact_summary || primaryFinding.message || 'Postgresome detected behavior that deserves investigation.'
                : 'Postgresome is still collecting enough evidence to explain what needs attention first.'}
            </p>
            <p className="mt-2 text-[14px] leading-6 text-[var(--muted-foreground)]">
              {primaryFinding ? supportingSummary(primaryFinding) : 'As soon as a reliable signal crosses its threshold, the first diagnosis will appear here.'}
            </p>
            <div className="mt-5 flex flex-wrap gap-3">
              <MetaPill label="Started" value={primaryFinding ? formatRelativeTimeShort(primaryFinding.first_seen_at) : 'Waiting'} />
              <MetaPill label="Impact" value={primaryFinding ? (primaryFinding.severity === 'critical' ? 'High' : 'Medium') : 'Unknown'} />
              <MetaPill label="Scope" value={affectedScope} />
              <MetaPill label="Confidence" value={primaryFinding?.confidence_label || 'High'} />
            </div>
            <div className="mt-6">
              <Link
                to={primaryFinding ? `/findings/${primaryFinding.id}` : '/findings'}
                className="inline-flex h-10 items-center gap-2 rounded-[10px] bg-[var(--primary)] px-4 text-[14px] font-medium text-white no-underline transition-colors hover:bg-[var(--primary-strong)]"
              >
                Investigate issue
                <ArrowRight size={14} />
              </Link>
            </div>
          </div>
        </section>

        <section className="space-y-4">
          <SectionHeading title="Other findings" body="Only the next few things worth your attention." />
          <div className="overflow-hidden rounded-xl border border-[var(--border)] bg-[var(--panel)]">
            {secondaryFindings.length === 0 ? (
              <div className="px-6 py-8 text-[14px] text-[var(--muted-foreground)]">No active findings yet. Postgresome is still collecting enough evidence to form a diagnosis.</div>
            ) : (
              secondaryFindings.map((finding, index) => (
                <Link
                  key={finding.id}
                  to={`/findings/${finding.id}`}
                  className={`flex w-full items-start gap-4 px-6 py-5 text-left no-underline transition-colors hover:bg-[var(--muted)] ${index !== 0 ? 'border-t border-[var(--border)]' : ''}`}
                >
                  <SeverityDot severity={finding.severity} />
                  <div className="min-w-0 flex-1">
                    <div className="flex flex-wrap items-center gap-2">
                      <h3 className="text-[15px] font-semibold text-[var(--foreground)]">{finding.problem_summary || finding.title}</h3>
                    </div>
                    <p className="mt-1 text-[14px] leading-6 text-[var(--muted-foreground)]">
                      {finding.evidence_summary || finding.impact_summary || 'Evidence suggests degraded database behavior that needs investigation.'}
                    </p>
                  </div>
                  <div className="hidden min-w-[120px] md:block">
                    <div className="text-[12px] text-[var(--muted-foreground)]">Started</div>
                    <div className="mt-1 text-[14px] font-medium text-[var(--body)]">{formatRelativeTimeShort(finding.first_seen_at)}</div>
                  </div>
                  <ChevronRight size={16} className="mt-1 shrink-0 text-[var(--muted-foreground)]" />
                </Link>
              ))
            )}
          </div>
        </section>
      </div>
    </AppShell>
  );
}

function ErrorBanner({ message }: { message: string }) {
  return <div className="rounded-xl border border-[var(--danger)] bg-[var(--danger-soft)] px-4 py-3 text-[13px] text-[var(--danger)]">{message}</div>;
}

function SectionHeading({ title, body }: { title: string; body: string }) {
  return (
    <div>
      <h2 className="text-[16px] font-semibold tracking-[-0.01em] text-[var(--foreground)]">{title}</h2>
      {body ? <p className="mt-1 text-[14px] leading-6 text-[var(--muted-foreground)]">{body}</p> : null}
    </div>
  );
}

function SeverityLabel({ severity }: { severity: string }) {
  const key = severity.toLowerCase();
  const className =
    key === 'critical'
      ? 'bg-[var(--danger-soft)] text-[var(--danger)]'
      : key === 'warning'
        ? 'bg-[var(--warning-soft)] text-[var(--warning)]'
        : key === 'healthy'
          ? 'bg-[var(--success-soft)] text-[var(--success)]'
          : 'bg-[var(--info-soft)] text-[var(--info)]';

  return <span className={`inline-flex rounded-full px-3 py-1 text-[12px] font-semibold ${className}`}>{severity}</span>;
}

function SeverityDot({ severity }: { severity: string }) {
  const color = severity === 'critical' ? 'bg-[var(--danger)]' : severity === 'warning' ? 'bg-[var(--warning)]' : 'bg-[var(--info)]';
  return <div className={`mt-1 h-2.5 w-2.5 shrink-0 rounded-full ${color}`} />;
}

function MetaPill({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-full border border-[var(--border)] bg-[var(--muted)] px-3 py-2 text-[13px]">
      <span className="text-[var(--muted-foreground)]">{label}: </span>
      <span className="font-medium text-[var(--foreground)]">{value}</span>
    </div>
  );
}

function supportingSummary(finding: IssueQueueItem) {
  if (finding.category.toLowerCase().includes('rollback')) {
    return 'Likely related to application errors, constraints, or timeouts.';
  }

  if (finding.category.toLowerCase().includes('cache')) {
    return 'Likely related to read amplification, cold working sets, or query patterns scanning more blocks than usual.';
  }

  return finding.suggested_action || 'Review the related workload and evidence trail to understand what changed first.';
}

function scopeLabel(finding: IssueQueueItem | null) {
  if (!finding) return 'Production database';
  if (finding.resource_type === 'query') return 'Query workload';
  if (finding.resource_type === 'table' || finding.resource_type === 'index') return finding.resource_name;
  return 'Production database';
}
