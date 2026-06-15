import { BookText, TriangleAlert, Wrench } from 'lucide-react';
import type { ReactNode } from 'react';
import type { IssueDetailResponse } from '../types/issues';
import { formatRelativeTime } from '../lib/format';
import { SeverityBadge, VerificationBadge } from './status-badges';
import { Badge } from './ui/badge';
import { DetailCard } from './ui/card';

export function FindingDetail({ detail }: { detail: IssueDetailResponse }) {
  const finding = detail.finding;
  const likelyCause = buildLikelyCause(detail);
  const evidenceItems = detail.evidence ?? [];
  const timeline = buildTimeline(detail);

  return (
    <div className="space-y-5">
      <DetailCard
        title={finding.problem_summary || finding.title}
        description={finding.impact_summary || 'Postgresome thinks this issue is affecting database behavior right now.'}
        actions={
          <div className="flex flex-wrap items-center gap-2">
            <SeverityBadge severity={finding.severity} />
            <VerificationBadge status={finding.verification_status} />
          </div>
        }
      >
        <div className="space-y-4">
          <div className="flex flex-wrap gap-2">
            <Badge variant="neutral">{finding.category}</Badge>
            <Badge variant="neutral">
              {finding.resource_type}: {finding.resource_name}
            </Badge>
            <Badge variant="neutral">Seen {finding.occurrence_count} times</Badge>
          </div>

          <div className="grid gap-3 md:grid-cols-4">
            <EvidenceStat label="Started" value={formatRelativeTime(finding.first_seen_at)} />
            <EvidenceStat label="Impact" value={finding.severity === 'critical' ? 'High' : finding.severity === 'warning' ? 'Medium' : 'Watch'} />
            <EvidenceStat label="Scope" value={friendlyScope(finding.resource_type, finding.resource_name)} />
            <EvidenceStat label="Confidence" value={finding.confidence_label || 'High'} />
          </div>
        </div>
      </DetailCard>

      <DetailCard title="Diagnosis" description="A single reading path: issue, explanation, evidence, and next step.">
        <div className="space-y-4">
          <InlinePanel icon={<BookText size={15} />} title="What happened" body={describeWhatHappened(detail)} />

          {evidenceItems.length > 0 ? (
            <div className="space-y-3">
              {evidenceItems.map((item) => (
                <EvidenceRow
                  key={item.id}
                  label={item.label}
                  role={item.role}
                  summary={item.summary}
                  currentValue={formatEvidenceValue(item.current_value)}
                  baselineValue={item.baseline_value ? formatEvidenceValue(item.baseline_value) : undefined}
                  changePercent={item.change_percent}
                  referenceId={item.reference_id}
                />
              ))}
            </div>
          ) : (
            <div className="grid gap-3 md:grid-cols-3">
              {buildEvidenceRows(detail).map((item) => (
                <EvidenceStat key={item.label} label={item.label} value={item.value} />
              ))}
            </div>
          )}

          <InlinePanel icon={<TriangleAlert size={15} />} title="Likely cause" body={likelyCause} />

          <InlinePanel
            icon={<Wrench size={15} />}
            title="Recommended next step"
            body={finding.suggested_action || finding.recommendation || 'Inspect the related workload and verify whether the triggering signal is still rising.'}
          />

          <div className="rounded-xl border border-[var(--border)] bg-[var(--muted)] px-4 py-4">
            <div className="text-[12px] font-semibold uppercase tracking-[0.08em] text-[var(--muted-foreground)]">Verify after the change</div>
            <div className="mt-2 text-[13px] leading-6 text-[var(--body)]">
              {finding.verification_hint || 'Confirm the evidence trends back toward baseline and the issue stops recurring.'}
            </div>
          </div>
        </div>
      </DetailCard>

      <DetailCard title="Evidence trail" description="Only the extra context you may want if you need one step more depth.">
        <div className="space-y-3">
          {timeline.map((item) => (
            <div key={item.label} className="flex gap-4 rounded-xl border border-[var(--border)] bg-[var(--panel)] px-4 py-4">
              <div className="min-w-[88px] text-[12px] font-medium text-[var(--muted-foreground)]">{item.time}</div>
              <div>
                <div className="text-[14px] font-medium text-[var(--foreground)]">{item.label}</div>
                <div className="mt-1 text-[13px] leading-6 text-[var(--muted-foreground)]">{item.note}</div>
              </div>
            </div>
          ))}
        </div>
      </DetailCard>
    </div>
  );
}

function EvidenceStat({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-xl border border-[var(--border)] bg-[var(--muted)] p-4">
      <div className="text-[12px] font-semibold uppercase tracking-[0.08em] text-[var(--muted-foreground)]">{label}</div>
      <div className="mt-2 text-[14px] font-medium text-[var(--foreground)]">{value}</div>
    </div>
  );
}

function InlinePanel({ icon, title, body }: { icon: ReactNode; title: string; body: string }) {
  return (
    <div className="rounded-xl border border-[var(--border)] bg-[var(--muted)] p-4">
      <div className="mb-2 flex items-center gap-2 text-[14px] font-medium text-[var(--foreground)]">
        {icon}
        {title}
      </div>
      <p className="text-[13px] leading-6 text-[var(--body)]">{body}</p>
    </div>
  );
}

function EvidenceRow({
  label,
  role,
  summary,
  currentValue,
  baselineValue,
  changePercent,
  referenceId,
}: {
  label: string;
  role: string;
  summary: string;
  currentValue: string;
  baselineValue?: string;
  changePercent: number;
  referenceId: string;
}) {
  return (
    <div className="rounded-xl border border-[var(--border)] bg-[var(--panel)] px-4 py-4">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <div className="flex flex-wrap items-center gap-2">
            <div className="text-[14px] font-medium text-[var(--foreground)]">{label}</div>
            <Badge variant="neutral">{role}</Badge>
          </div>
          <p className="mt-2 text-[13px] leading-6 text-[var(--body)]">{summary}</p>
        </div>
        <div className="min-w-[144px] text-right">
          <div className="font-mono text-[14px] font-medium text-[var(--foreground)]">{currentValue}</div>
          <div className="mt-1 font-mono text-[12px] text-[var(--muted-foreground)]">
            {baselineValue ? `baseline ${baselineValue}` : `${changePercent >= 0 ? '+' : ''}${changePercent.toFixed(0)}%`}
          </div>
        </div>
      </div>
      {referenceId ? <div className="mt-3 font-mono text-[12px] text-[var(--muted-foreground)]">{referenceId}</div> : null}
    </div>
  );
}

function describeWhatHappened(detail: IssueDetailResponse) {
  const finding = detail.finding;
  if (detail.historical_context) {
    const { previous_value, current_value, change_percent } = detail.historical_context;
    return `${finding.problem_summary || finding.title} shifted from ${previous_value.toLocaleString()} to ${current_value.toLocaleString()} over the recent observation window. This is a ${changePercentLabel(change_percent)} from the previous baseline and suggests database behavior changed materially rather than normal variance.`;
  }

  return finding.evidence_summary || 'Postgresome detected a meaningful shift in PostgreSQL behavior and raised this diagnosis because the signal moved beyond its normal range.';
}

function buildEvidenceRows(detail: IssueDetailResponse) {
  const finding = detail.finding;
  const rows = [
    { label: 'Current signal', value: finding.current_value.toLocaleString() },
    { label: 'Threshold', value: finding.threshold_value.toLocaleString() },
    { label: 'Occurrences', value: finding.occurrence_count.toLocaleString() },
  ];

  if (detail.historical_context) {
    rows.push({ label: detail.historical_context.baseline_label || 'Baseline', value: detail.historical_context.baseline_value.toLocaleString() });
  }

  return rows;
}

function buildTimeline(detail: IssueDetailResponse) {
  const finding = detail.finding;
  const items = [
    { time: formatRelativeTime(finding.first_seen_at), label: 'Issue started appearing', note: finding.problem_summary || finding.title },
    { time: formatRelativeTime(finding.detected_at), label: 'Threshold exceeded', note: finding.evidence_summary || 'The signal crossed the detection threshold.' },
    { time: formatRelativeTime(finding.last_seen_at), label: 'Latest evidence recorded', note: finding.verification_summary || 'Postgresome most recently observed this diagnosis in the latest collection cycle.' },
  ];

  for (const item of detail.evidence ?? []) {
    items.push({
      time: formatRelativeTime(item.observed_at),
      label: item.label,
      note: item.summary,
    });
  }

  if (finding.verified_fixed_at) {
    items.push({ time: formatRelativeTime(finding.verified_fixed_at), label: 'Fix verified', note: 'The evidence trended back toward baseline after a corrective change.' });
  }

  return items;
}

function buildLikelyCause(detail: IssueDetailResponse) {
  const category = detail.finding.category.toLowerCase();

  if (category.includes('query')) {
    return 'Execution time likely changed because the workload shape shifted, the plan regressed, or the query started reading more blocks than usual.';
  }

  if (category.includes('vacuum') || category.includes('bloat')) {
    return 'Maintenance may be falling behind, allowing dead tuples or stale statistics to accumulate faster than normal.';
  }

  if (category.includes('lock')) {
    return 'A blocking transaction or a write hotspot is probably holding resources long enough to affect other work.';
  }

  return 'Related application errors, workload changes, or resource pressure are likely contributing to the database behavior change.';
}

function friendlyScope(resourceType: string, resourceName: string) {
  if (!resourceType) return 'Production database';
  if (resourceType === 'query') return 'Query workload';
  if (resourceType === 'table') return resourceName;
  if (resourceType === 'index') return resourceName;
  return 'Production database';
}

function changePercentLabel(value: number) {
  return `${value >= 0 ? '+' : ''}${value.toFixed(0)}% change`;
}

function formatEvidenceValue(value: number) {
  if (Number.isInteger(value)) return value.toLocaleString();
  return value.toLocaleString(undefined, { maximumFractionDigits: 1 });
}
