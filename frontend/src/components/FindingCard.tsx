import type { DashboardFinding } from '../types/dashboard';
import { formatRelativeTime } from '../lib/format';
import { SeverityPill } from './SeverityPill';

export function FindingCard({ finding }: { finding: DashboardFinding }) {
  return (
    <div
      className="flex flex-col gap-2 rounded-[var(--radius-md)] border p-3"
      style={{ borderColor: 'var(--border-subtle)', background: 'var(--surface-raised)' }}
    >
      <div className="flex items-center justify-between gap-2">
        <div className="flex items-center gap-2">
          <SeverityPill severity={finding.severity} />
          <span
            className="text-xs font-medium"
            style={{ color: 'var(--text-muted)', fontFamily: 'var(--font-mono)' }}
          >
            {finding.category}
          </span>
        </div>
        <span className="shrink-0 text-xs" style={{ color: 'var(--text-faint)' }}>
          {formatRelativeTime(finding.detected_at)}
        </span>
      </div>
      <div className="text-sm font-semibold" style={{ color: 'var(--text-primary)' }}>
        {finding.title}
      </div>
      <div className="text-[13px]" style={{ color: 'var(--text-secondary)' }}>
        {finding.message}
      </div>
      {finding.recommendation && (
        <div
          className="rounded-[var(--radius-md)] border px-3 py-2 text-[13px]"
          style={{ background: 'var(--success-tint)', borderColor: 'rgba(26,127,55,0.18)', color: 'var(--text-secondary)' }}
        >
          <span className="font-semibold" style={{ color: 'var(--success)' }}>
            Suggested fix{' '}
          </span>
          {finding.recommendation}
        </div>
      )}
    </div>
  );
}
