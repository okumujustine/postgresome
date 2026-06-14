import { Link } from 'react-router-dom';
import type { DashboardFinding } from '../types/dashboard';
import { formatRelativeTime } from '../lib/format';
import { SeverityPill } from './SeverityPill';

export function HealthIssueCard({ finding }: { finding: DashboardFinding }) {
  return (
    <div
      className="flex flex-col gap-3 rounded-[var(--radius-lg)] border px-4 py-[15px]"
      style={{ borderColor: 'var(--border-subtle)', background: 'var(--surface-card)' }}
    >
      <div className="flex items-start justify-between gap-3">
        <div className="flex min-w-0 flex-col gap-2">
          <div className="flex items-center gap-2">
            <SeverityPill severity={finding.severity} />
            <span className="text-[12px]" style={{ color: 'var(--text-muted)', fontFamily: 'var(--font-mono)' }}>
              {finding.resource_type}: {finding.resource_name}
            </span>
          </div>
          <div className="text-[15px] font-semibold leading-[1.35]" style={{ color: 'var(--text-primary)' }}>
            {finding.title}
          </div>
        </div>
        <Link
          to={`/issues/${finding.id}`}
          className="inline-flex h-[var(--control-h-sm)] shrink-0 items-center rounded-[var(--radius-md)] border px-3 text-[13px] font-medium no-underline"
          style={{ background: 'var(--surface-raised)', color: 'var(--text-secondary)', borderColor: 'var(--border-default)' }}
        >
          Investigate
        </Link>
      </div>

      <div className="grid gap-3 md:grid-cols-[minmax(0,1fr)_minmax(0,1fr)]">
        <div className="min-w-0">
          <div className="mb-1 text-[12px] font-medium" style={{ color: 'var(--text-muted)' }}>
            Problem
          </div>
          <div className="text-[13px] leading-[1.5]" style={{ color: 'var(--text-secondary)' }}>
            {finding.message}
          </div>
        </div>

        <div className="min-w-0">
          <div className="mb-1 text-[12px] font-medium" style={{ color: 'var(--text-muted)' }}>
            Evidence
          </div>
          <div className="text-[13px] leading-[1.5]" style={{ color: 'var(--text-secondary)' }}>
            {finding.resource_name} crossed the {finding.rule_key} threshold with a current value of{' '}
            {finding.current_value.toLocaleString()} against {finding.threshold_value.toLocaleString()}.
          </div>
        </div>
      </div>

      {finding.recommendation && (
        <div>
          <div className="mb-1 text-[12px] font-medium" style={{ color: 'var(--text-muted)' }}>
            Recommended fix
          </div>
          <div className="text-[13px] leading-[1.5]" style={{ color: 'var(--text-secondary)' }}>
            {finding.recommendation}
          </div>
        </div>
      )}

      <div className="flex flex-wrap items-center gap-x-4 gap-y-1 text-[12px]" style={{ color: 'var(--text-faint)' }}>
        <span>First seen {formatRelativeTime(finding.first_seen_at)}</span>
        <span>Last seen {formatRelativeTime(finding.last_seen_at)}</span>
        <span>Seen {finding.occurrence_count} times</span>
      </div>
    </div>
  );
}
