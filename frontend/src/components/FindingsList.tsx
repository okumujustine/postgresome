import type { DashboardFindings } from '../types/dashboard';
import { Card } from './Card';

interface FindingsListProps {
  findings: DashboardFindings;
}

const SEVERITY_STYLES: Record<string, { fg: string; bg: string; border: string }> = {
  critical: { fg: '#FF9892', bg: 'var(--danger-tint)', border: 'rgba(248,81,73,0.32)' },
  warning: { fg: '#F2D17C', bg: 'var(--warning-tint)', border: 'rgba(227,179,65,0.32)' },
  info: { fg: 'var(--blue-300)', bg: 'var(--blue-tint)', border: 'rgba(77,141,255,0.3)' },
};

function SeverityPill({ severity }: { severity: string }) {
  const style = SEVERITY_STYLES[severity.toLowerCase()] ?? {
    fg: 'var(--text-secondary)',
    bg: 'var(--surface-active)',
    border: 'var(--border-default)',
  };

  return (
    <span
      className="inline-flex h-[22px] items-center rounded-[var(--radius-pill)] px-[9px] text-xs font-medium capitalize"
      style={{ background: style.bg, color: style.fg, border: `1px solid ${style.border}`, letterSpacing: 'var(--ls-snug)' }}
    >
      {severity}
    </span>
  );
}

function formatRelativeTime(iso: string): string {
  const diffMs = Date.now() - new Date(iso).getTime();
  const diffMin = Math.round(diffMs / 60000);
  if (diffMin < 1) return 'just now';
  if (diffMin < 60) return `${diffMin}m ago`;
  const diffHr = Math.round(diffMin / 60);
  if (diffHr < 24) return `${diffHr}h ago`;
  return `${Math.round(diffHr / 24)}d ago`;
}

export function FindingsList({ findings }: FindingsListProps) {
  return (
    <Card
      title="Findings"
      subtitle={`${findings.critical} critical · ${findings.warning} warning · ${findings.info} info`}
    >
      {findings.recent.length === 0 ? (
        <div className="text-sm" style={{ color: 'var(--text-muted)' }}>
          No findings in this time range.
        </div>
      ) : (
        <div className="flex flex-col gap-3">
          {findings.recent.map((finding) => (
            <div
              key={finding.id}
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
                <div className="text-xs" style={{ color: 'var(--text-muted)' }}>
                  <span className="font-medium" style={{ color: 'var(--text-secondary)' }}>
                    Recommendation:{' '}
                  </span>
                  {finding.recommendation}
                </div>
              )}
            </div>
          ))}
        </div>
      )}
    </Card>
  );
}
