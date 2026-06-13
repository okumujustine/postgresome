import { Link } from 'react-router-dom';
import type { DashboardFinding } from '../types/dashboard';
import { severityEmoji } from '../lib/format';

export function HealthIssueCard({ finding }: { finding: DashboardFinding }) {
  return (
    <div
      className="flex flex-col gap-2 rounded-[var(--radius-md)] border p-4"
      style={{ borderColor: 'var(--border-subtle)', background: 'var(--surface-raised)' }}
    >
      <div className="flex items-start justify-between gap-3">
        <div className="text-sm font-semibold" style={{ color: 'var(--text-primary)' }}>
          {severityEmoji(finding.severity)} {finding.title}
        </div>
        <Link
          to={`/issues/${finding.id}`}
          className="inline-flex h-[var(--control-h-sm)] shrink-0 items-center rounded-[var(--radius-md)] border px-3 text-[13px] font-medium no-underline"
          style={{ background: 'var(--surface-card)', color: 'var(--text-secondary)', borderColor: 'var(--border-default)' }}
        >
          Investigate
        </Link>
      </div>

      <div className="text-[13px]" style={{ color: 'var(--text-secondary)' }}>
        <span className="font-medium" style={{ color: 'var(--text-primary)' }}>
          Impact:{' '}
        </span>
        {finding.message}
      </div>

      <div className="text-[13px]" style={{ color: 'var(--text-secondary)' }}>
        <span className="font-medium" style={{ color: 'var(--text-primary)' }}>
          Evidence:{' '}
        </span>
        <span style={{ fontFamily: 'var(--font-mono)' }}>
          {finding.resource_type}: {finding.resource_name}
        </span>
        {' · '}
        Current: {finding.current_value.toLocaleString()} · Threshold: {finding.threshold_value.toLocaleString()}
      </div>

      {finding.recommendation && (
        <div className="text-[13px]" style={{ color: 'var(--text-secondary)' }}>
          <span className="font-medium" style={{ color: 'var(--text-primary)' }}>
            Suggested action:{' '}
          </span>
          {finding.recommendation}
        </div>
      )}
    </div>
  );
}
