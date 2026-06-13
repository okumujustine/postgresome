import type { ReactNode } from 'react';

interface MetricCardProps {
  label: string;
  value: string;
  unit?: string;
  trendPercent?: number | null;
  invertTrend?: boolean;
  icon?: ReactNode;
  hideFooter?: boolean;
}

export function MetricCard({ label, value, unit, trendPercent, invertTrend = false, icon, hideFooter = false }: MetricCardProps) {
  const hasTrend = trendPercent != null && !Number.isNaN(trendPercent) && trendPercent !== 0;
  const up = hasTrend && trendPercent! >= 0;
  const good = !hasTrend ? null : invertTrend ? !up : up;
  const deltaColor = good == null ? 'var(--text-muted)' : good ? 'var(--success)' : 'var(--danger)';

  return (
    <div
      className="flex min-w-0 flex-col gap-3 rounded-[var(--radius-lg)] border p-5"
      style={{ background: 'var(--surface-card)', borderColor: 'var(--border-subtle)', boxShadow: 'var(--shadow-xs)' }}
    >
      <div className="flex items-center gap-2">
        {icon && (
          <span className="inline-flex" style={{ color: 'var(--text-muted)' }}>
            {icon}
          </span>
        )}
        <span
          className="truncate text-[12.5px] font-medium"
          style={{ color: 'var(--text-muted)', letterSpacing: 'var(--ls-snug)' }}
        >
          {label}
        </span>
      </div>

      <div className="flex min-w-0 items-baseline gap-1">
        <span
          className="tabular"
          style={{
            fontFamily: 'var(--font-sans)',
            fontSize: 'var(--fs-metric)',
            fontWeight: 'var(--fw-semibold)',
            color: 'var(--text-primary)',
            letterSpacing: 'var(--ls-tight)',
            lineHeight: 1,
          }}
        >
          {value}
        </span>
        {unit && (
          <span className="text-sm font-medium" style={{ color: 'var(--text-muted)' }}>
            {unit}
          </span>
        )}
      </div>

      {hideFooter ? null : hasTrend ? (
        <div className="flex items-center gap-[5px] text-[12.5px] font-medium" style={{ color: deltaColor }}>
          <svg
            width="12"
            height="12"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2.6"
            strokeLinecap="round"
            strokeLinejoin="round"
            style={{ transform: up ? 'none' : 'rotate(180deg)' }}
          >
            <line x1="12" y1="19" x2="12" y2="5" />
            <polyline points="5 12 12 5 19 12" />
          </svg>
          <span className="tabular">{Math.abs(trendPercent!).toFixed(1)}%</span>
          <span style={{ color: 'var(--text-muted)', fontWeight: 'var(--fw-regular)' }}>vs previous period</span>
        </div>
      ) : (
        <div className="text-[12.5px] font-medium" style={{ color: 'var(--text-muted)' }}>
          No change vs previous period
        </div>
      )}
    </div>
  );
}
