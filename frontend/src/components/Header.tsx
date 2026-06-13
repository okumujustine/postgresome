import { ChevronDown, Database, RefreshCw } from 'lucide-react';
import type { MetricRange } from '../types/dashboard';
import { StatusBadge } from './StatusBadge';
import { useDatabaseInstance } from '../lib/databaseInstance';

const RANGE_OPTIONS: { value: MetricRange; label: string }[] = [
  { value: '15m', label: 'Last 15m' },
  { value: '1h', label: 'Last 1h' },
  { value: '6h', label: 'Last 6h' },
  { value: '24h', label: 'Last 24h' },
  { value: '7d', label: 'Last 7d' },
];

export interface HeaderProps {
  title: string;
  range?: MetricRange;
  onRangeChange?: (range: MetricRange) => void;
  onRefresh: () => void;
  refreshing: boolean;
}

export function Header({ title, range, onRangeChange, onRefresh, refreshing }: HeaderProps) {
  const { instances, selectedId, setSelectedId, loading } = useDatabaseInstance();
  const selectedInstance = instances.find((instance) => instance.id === selectedId);

  return (
    <header
      className="psm-header sticky top-0 z-20 flex shrink-0 items-center gap-[14px] border-b px-5 backdrop-blur-sm"
      style={{ height: 'var(--header-h)', borderColor: 'var(--border-subtle)', background: 'color-mix(in srgb, var(--bg-base) 82%, transparent)' }}
    >
      <div className="flex min-w-0 items-center gap-[11px]">
        <h1 className="m-0 truncate text-base font-semibold" style={{ color: 'var(--text-primary)', letterSpacing: 'var(--ls-snug)' }}>
          {title}
        </h1>
        {!loading && instances.length > 0 && (
          <>
            <span className="h-[18px] w-px" style={{ background: 'var(--border-default)' }} />
            <div className="relative flex items-center gap-2">
              <Database size={13} className="pointer-events-none absolute left-[9px]" style={{ color: 'var(--text-muted)' }} />
              <select
                value={selectedId}
                onChange={(event) => setSelectedId(event.target.value)}
                className="h-[28px] cursor-pointer appearance-none rounded-[var(--radius-pill)] border pr-7 pl-[28px] text-[12.5px] outline-none"
                style={{
                  background: 'var(--surface-card)',
                  borderColor: 'var(--border-subtle)',
                  color: 'var(--text-secondary)',
                  fontFamily: 'var(--font-mono)',
                }}
              >
                {instances.map((instance) => (
                  <option key={instance.id} value={instance.id} style={{ background: 'var(--surface-card)', color: 'var(--text-primary)' }}>
                    {instance.database_name}
                  </option>
                ))}
              </select>
              <ChevronDown size={12} className="pointer-events-none absolute right-[8px]" style={{ color: 'var(--text-muted)' }} />
              {selectedInstance && <StatusBadge status={selectedInstance.status} size="sm" />}
            </div>
          </>
        )}
      </div>

      <div className="flex-1" />

      {range && onRangeChange && (
        <div className="relative inline-block">
          <select
            value={range}
            onChange={(event) => onRangeChange(event.target.value as MetricRange)}
            className="h-[var(--control-h-sm)] cursor-pointer appearance-none rounded-[var(--radius-md)] border pr-8 pl-3 text-[13px] outline-none"
            style={{ background: 'var(--surface-raised)', color: 'var(--text-primary)', borderColor: 'var(--border-default)', fontFamily: 'var(--font-sans)' }}
          >
            {RANGE_OPTIONS.map((option) => (
              <option key={option.value} value={option.value} style={{ background: 'var(--surface-card)', color: 'var(--text-primary)' }}>
                {option.label}
              </option>
            ))}
          </select>
          <span className="pointer-events-none absolute right-[11px] top-1/2 -translate-y-1/2" style={{ color: 'var(--text-muted)' }}>
            <ChevronDown size={14} />
          </span>
        </div>
      )}

      <button
        onClick={onRefresh}
        aria-label="Refresh"
        disabled={refreshing}
        className="inline-flex h-[var(--control-h-sm)] w-[var(--control-h-sm)] shrink-0 cursor-pointer items-center justify-center rounded-[var(--radius-sm)] border"
        style={{ background: 'var(--surface-raised)', color: 'var(--text-secondary)', borderColor: 'var(--border-default)' }}
      >
        <RefreshCw size={15} className={refreshing ? 'animate-spin' : ''} />
      </button>
    </header>
  );
}
