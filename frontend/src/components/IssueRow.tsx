import { Link } from 'react-router-dom';
import type { DashboardFinding } from '../types/dashboard';
import { formatRelativeTime } from '../lib/format';
import { SeverityPill } from './SeverityPill';

export function IssueRow({ finding }: { finding: DashboardFinding }) {
  return (
    <div className="flex flex-col gap-2 py-3">
      <div className="flex items-start justify-between gap-3">
        <div className="flex min-w-0 items-center gap-2">
          <SeverityPill severity={finding.severity} />
          <Link
            to={`/issues/${finding.id}`}
            className="truncate text-sm font-semibold no-underline"
            style={{ color: 'var(--text-primary)' }}
          >
            {finding.title}
          </Link>
          {finding.status === 'resolved' && (
            <span
              className="inline-flex h-[22px] shrink-0 items-center rounded-[var(--radius-pill)] px-[9px] text-xs font-medium capitalize"
              style={{ background: 'var(--success-tint)', color: 'var(--success)', border: '1px solid rgba(26,127,55,0.25)', letterSpacing: 'var(--ls-snug)' }}
            >
              Resolved
            </span>
          )}
        </div>
        <Link
          to={`/issues/${finding.id}`}
          className="inline-flex h-[var(--control-h-sm)] shrink-0 items-center rounded-[var(--radius-md)] border px-3 text-[13px] font-medium no-underline"
          style={{ background: 'var(--surface-raised)', color: 'var(--text-secondary)', borderColor: 'var(--border-default)' }}
        >
          Open
        </Link>
      </div>

      <div className="flex min-w-0 items-center gap-2 text-[13px]" style={{ color: 'var(--text-secondary)' }}>
        <span className="shrink-0" style={{ fontFamily: 'var(--font-mono)', color: 'var(--text-muted)' }}>
          {finding.resource_type}: {finding.resource_name}
        </span>
        <span className="truncate">{finding.message}</span>
      </div>

      <div className="flex flex-wrap items-center gap-x-4 gap-y-1 text-xs" style={{ color: 'var(--text-faint)' }}>
        <span>First seen {formatRelativeTime(finding.first_seen_at)}</span>
        <span>Last seen {formatRelativeTime(finding.last_seen_at)}</span>
        <span>Seen {finding.occurrence_count}x</span>
      </div>
    </div>
  );
}
